// Package webui is the launcher's one graphical interface on every desktop.
// It serves embedded HTML on loopback; the game server and bridge remain
// headless processes owned by Supervisor.
package webui

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botlive"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
	"github.com/m-this/tf2-archipelago/launcher/internal/rcon"
	"github.com/m-this/tf2-archipelago/launcher/internal/roomcheck"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/saveplan"
	"github.com/m-this/tf2-archipelago/launcher/internal/session"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/srcdsconfig"
	"github.com/m-this/tf2-archipelago/launcher/internal/tailscalefastdl"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

const (
	linesMax     = 20000
	sessionEvery = 5 * time.Second
)

//go:embed index.html
var page []byte

type event struct {
	Name string
	Data any
}

// Snapshot is everything the page needs for one draw. Passwords never cross
// the loopback boundary; form already represents them as replacement fields.
type Snapshot struct {
	Title         string           `json:"title"`
	Status        string           `json:"status"`
	Running       bool             `json:"running"`
	Busy          bool             `json:"busy"`
	Room          string           `json:"room"`
	Join          string           `json:"join"`
	JoinURL       string           `json:"join_url"`
	Mission       string           `json:"mission"`
	Logs          []apruntime.Line `json:"logs"`
	Session       session.Snapshot `json:"session"`
	SessionError  string           `json:"session_error,omitempty"`
	Bots          []botlive.Seat   `json:"bots"`
	DrawnBots     string           `json:"drawn_bots,omitempty"`
	Form          *form.Model      `json:"form,omitempty"`
	FormPage      string           `json:"form_page,omitempty"`
	Notice        string           `json:"notice,omitempty"`
	NoticeSeq     uint64           `json:"notice_seq,omitempty"`
	ItemServer    string           `json:"item_server,omitempty"`
	MissionPool   []MissionPoolRow `json:"mission_pool,omitempty"`
	RestartNeeded bool             `json:"restart_needed,omitempty"`
}

// MissionPoolRow is the domain data behind one dense row in the settings
// table. The form field still owns selection and mutation; this only keeps the
// browser from reverse-engineering facts out of a human-readable label.
type MissionPoolRow struct {
	Field         string `json:"field"`
	Source        string `json:"source"`
	Name          string `json:"name"`
	Waves         string `json:"waves"`
	Compatibility string `json:"compatibility"`
	Mods          string `json:"mods"`
}

// App owns UI state and translates HTTP commands into launcher operations.
// It deliberately has no browser concepts beyond publishing plain events.
type App struct {
	mu sync.Mutex

	settings   settings.Settings
	supervisor *apruntime.Supervisor
	logs       []apruntime.Line
	busy       bool
	install    context.CancelFunc
	steamURL   string
	mission    string
	snapshot   session.Snapshot
	fetchErr   error
	notice     string
	noticeSeq  uint64
	draft      *form.State
	formPage   string
	community  []string
	imported   []string
	smAsked    bool
	smHeld     bool
	itemServer string
	logFile    *os.File

	subscribers map[chan event]struct{}
	quit        chan struct{}
	quitOnce    sync.Once
}

func New(s settings.Settings, logger *slog.Logger) *App {
	community := availableCommunityPackNames(s.CommunityContentDir)
	a := &App{
		settings:    s,
		community:   community,
		imported:    importedCommunityPackNames(s.CommunityContentDir, community),
		subscribers: make(map[chan event]struct{}),
		quit:        make(chan struct{}),
	}
	a.supervisor = apruntime.NewSupervisor(s, logger, a.append)
	return a
}

// Run serves the UI, opens it in the desktop browser, and blocks until Quit or
// an operating-system signal. It is one process, not a remotely exposed daemon.
func Run(s settings.Settings, logger *slog.Logger) error {
	a := New(s, logger)
	if file, err := apruntime.CreateLogFile(s.InstallRoot); err == nil {
		a.logFile = file
		defer func() { _ = file.Close() }()
	}
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp4", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("cannot start the launcher interface: %w", err)
	}
	server := &http.Server{Handler: a.handler(listener.Addr().String()), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.say("interface: %v", err)
			a.Quit()
		}
	}()
	go a.watchSession()

	address := "http://" + listener.Addr().String()
	if err := openOrPrompt(address, os.Stderr, winproc.OpenURL); err != nil {
		a.say("browser did not open automatically: %v", err)
	}
	a.say("interface: %s", address)
	if s.APPort != 0 || s.TestMode {
		a.Start()
	} else {
		a.OpenSettings("Archipelago room")
	}

	signals, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	select {
	case <-signals.Done():
	case <-a.quit:
	}
	a.Stop()
	ctx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	return server.Shutdown(ctx)
}

// openOrPrompt hands the interface to the desktop browser. A missing desktop
// opener is normal under WSL, SSH and minimal Linux installations, so it is
// not a reason to tear down an otherwise working interface. Do not wait for an
// answer here: stdin may be closed or owned by a service, while the URL is
// already ready to use.
func openOrPrompt(address string, output io.Writer, opener func(string) error) error {
	err := opener(address)
	if err == nil {
		return nil
	}
	if promptErr := writeManualURL(output, address, err); promptErr != nil {
		return errors.Join(err, promptErr)
	}
	return err
}

func writeManualURL(output io.Writer, address string, openErr error) error {
	if _, err := fmt.Fprintf(output, "\nThe launcher could not open a browser automatically: %v\n\n", openErr); err != nil {
		return fmt.Errorf("write browser fallback: %w", err)
	}
	if _, err := fmt.Fprintf(output, "Open this URL manually:\n\n    %s\n\n", address); err != nil {
		return fmt.Errorf("write browser fallback: %w", err)
	}
	if _, err := fmt.Fprintln(output, "The launcher will keep running. Press Ctrl+C here or Quit in the browser to stop it."); err != nil {
		return fmt.Errorf("write browser fallback: %w", err)
	}
	if _, err := fmt.Fprintln(output); err != nil {
		return fmt.Errorf("write browser fallback: %w", err)
	}
	return nil
}

func (a *App) Handler() http.Handler {
	return a.handler("127.0.0.1")
}

func (a *App) handler(authority string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.servePage)
	mux.HandleFunc("GET /api/snapshot", a.serveSnapshot)
	mux.HandleFunc("GET /api/events", a.serveEvents)
	mux.HandleFunc("GET /api/files/settings", a.serveSettingsFile)
	mux.HandleFunc("GET /api/files/player", a.servePlayerFile)
	mux.HandleFunc("GET /api/files/install", a.serveInstallLocation)
	mux.HandleFunc("GET /api/files/seed", a.serveGeneratedSeed)
	mux.HandleFunc("GET /api/files/debug", a.serveDebugBundle)
	mux.HandleFunc("GET /api/open/funnel", a.serveFunnelApproval)
	mux.HandleFunc("POST /api/server/{action}", a.serveServerAction)
	mux.HandleFunc("POST /api/rcon", a.serveRCON)
	mux.HandleFunc("POST /api/mission", a.serveMission)
	mux.HandleFunc("POST /api/settings/open", a.serveSettingsOpen)
	mux.HandleFunc("POST /api/settings/change", a.serveSettingsChange)
	mux.HandleFunc("POST /api/settings/action", a.serveSettingsAction)
	mux.HandleFunc("POST /api/settings/import", a.serveSettingsImport)
	mux.HandleFunc("POST /api/settings/save", a.serveSettingsSave)
	mux.HandleFunc("POST /api/settings/cancel", a.serveSettingsCancel)
	mux.HandleFunc("POST /api/quit", func(http.ResponseWriter, *http.Request) { a.Quit() })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Host is pinned to the address this listener actually bound. Comparing
		// Origin with Host is not enough: a DNS-rebinding attacker controls both.
		if r.Host != authority {
			http.Error(w, "wrong host", http.StatusForbidden)
			return
		}
		// Browsers send Origin for script requests. An empty Origin stays allowed
		// so trusted local tools can use the loopback API on this single-user app.
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+authority {
			http.Error(w, "wrong origin", http.StatusForbidden)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'")
		mux.ServeHTTP(w, r)
	})
}

func (a *App) servePage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

// Local address discovery is internally bounded and does not perform work on
// behalf of the browser request.
//
//nolint:contextcheck // Address discovery owns its timeout instead of the request.
func (a *App) serveSnapshot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(a.Snapshot()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) draftSettings() (settings.Settings, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return settings.Settings{}, errors.New("settings are not open")
	}
	return a.draft.Settings, nil
}

func serveBrowserFile(w http.ResponseWriter, r *http.Request, path, disposition string) {
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": filepath.Base(path)}))
	http.ServeFile(w, r, path)
}

func (a *App) serveSettingsFile(w http.ResponseWriter, r *http.Request) {
	path, err := settings.Path()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveBrowserFile(w, r, path, "inline")
}

func (a *App) servePlayerFile(w http.ResponseWriter, r *http.Request) {
	s, err := a.draftSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	path, err := settings.WritePlayerFile(s, assets.ArchipelagoVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveBrowserFile(w, r, path, "inline")
}

func (a *App) serveInstallLocation(w http.ResponseWriter, _ *http.Request) {
	s, err := a.draftSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "Server folder:\n\n%s\n", s.InstallRoot)
}

func (a *App) serveGeneratedSeed(w http.ResponseWriter, r *http.Request) {
	s, err := a.draftSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	result, err := generate.Run(r.Context(), generate.Options{
		Settings: s, AppDir: s.ArchipelagoDir,
		Apworld: assets.Apworld(), ArchipelagoVersion: assets.ArchipelagoVersion,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveBrowserFile(w, r, result.Archive, "attachment")
}

// Bundle collection owns a short timeout for its optional bridge snapshot and
// should finish producing the requested download if the browser disconnects.
//
//nolint:contextcheck // Bundle collection owns its bridge timeout.
func (a *App) serveDebugBundle(w http.ResponseWriter, r *http.Request) {
	s, err := a.draftSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	path, err := debugbundle.Write(s, assets.Versions(), time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveBrowserFile(w, r, path, "attachment")
}

func (a *App) serveFunnelApproval(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	result, err := authorizeFunnel(ctx)
	if err != nil {
		http.Error(w, funnelSetupAdvice(err), http.StatusBadGateway)
		return
	}
	if result.ApprovalURL != "" {
		http.Redirect(w, r, result.ApprovalURL, http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, "Tailscale Funnel is ready for this tailnet. You can close this tab.")
}

var authorizeFunnel = tailscalefastdl.Authorize

func funnelSetupAdvice(err error) string {
	executable, pathErr := os.Executable()
	if pathErr != nil {
		executable = "tf2ap-linux-amd64"
	}
	command := strconv.Quote(executable) + " -setup-funnel"
	if _, ok := errors.AsType[*tailscalefastdl.OperatorRequiredError](err); ok {
		return "Tailscale needs one-time permission for your user to manage Funnel.\n\n" +
			"Run this once in a terminal:\n\n    sudo tailscale set --operator=$USER\n\n" +
			"Then run the launcher normally; -setup-funnel performs the Funnel command automatically:\n\n    " + command
	}
	detail := strings.TrimSpace(strings.Split(err.Error(), "\n")[0])
	if detail == "" || strings.Contains(strings.ToLower(detail), "tailscale funnel") {
		detail = "Tailscale could not complete the setup check."
	}
	return "The browser could not finish Tailscale Funnel setup.\n\n" +
		"Run this launcher from a terminal instead; -setup-funnel performs the Funnel command automatically and prints any approval URL:\n\n    " +
		command + "\n\nTailscale said: " + detail
}

func (a *App) serveEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	ch := make(chan event, 128)
	a.mu.Lock()
	a.subscribers[ch] = struct{}{}
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.subscribers, ch)
		a.mu.Unlock()
	}()
	_, _ = fmt.Fprint(w, "event: state\ndata: {}\n\n")
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case message := <-ch:
			data, err := json.Marshal(message.Data)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", message.Name, data)
			flusher.Flush()
		}
	}
}

func (a *App) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	running := a.supervisor.Running()
	status := "stopped"
	if a.busy && !running {
		status = "starting"
	} else if running {
		status = "running"
		if settings.Effective(a.settings.SrcdsReach, a.settings.SrcdsToken) == settings.ReachSteam && a.steamURL == "" {
			status = "starting"
		}
	}
	s := a.settings
	playing := apruntime.StartMap(s)
	if running && a.mission != "" {
		playing = a.mission
	}
	room := "no room set"
	if s.TestMode {
		room = "test mode"
	} else if s.APPort != 0 {
		room = "room " + (settings.Room{Host: s.APHost, Port: s.APPort}).String()
	}
	if playing != "" {
		room += "   " + playing
	}
	sessionState := a.snapshot
	sessionState.Missions = slices.Clone(sessionState.Missions)
	for i := range sessionState.Missions {
		mission, known := gamedata.MissionByPopFile(sessionState.Missions[i].PopFile)
		if known && slices.Contains(a.imported, gamedata.MissionPack(mission.ID)) {
			sessionState.Missions[i].Source = "Imported"
		}
	}
	var model *form.Model
	var missionPool []MissionPoolRow
	restartOnSave := false
	if a.draft != nil {
		built := form.Build(*a.draft, a.formEnvLocked())
		for page := range built.Tabs {
			for field := range built.Tabs[page].Fields {
				if built.Tabs[page].Fields[field].Kind == form.Password {
					built.Tabs[page].Fields[field].Value = ""
				}
			}
		}
		model = &built
		missionPool = missionPoolRows(*a.draft, a.community, a.imported)
		restartOnSave = restartNeeded(running, s, a.draft)
	}
	result := Snapshot{
		Title:  assets.Title("Mann vs Archipelago"),
		Status: status, Running: running, Busy: a.busy, Room: room,
		Join: a.joinLineLocked(), JoinURL: a.joinURLLocked(), Mission: playing,
		Logs: slices.Clone(a.logs), Session: sessionState,
		Bots: botlive.Team(s), DrawnBots: botlive.Drawn(s),
		Form: model, FormPage: a.formPage, Notice: a.notice, NoticeSeq: a.noticeSeq,
		ItemServer: a.itemServer, MissionPool: missionPool, RestartNeeded: restartOnSave,
	}
	if a.fetchErr != nil {
		result.SessionError = a.fetchErr.Error()
	}
	return result
}

func restartNeeded(running bool, before settings.Settings, draft *form.State) bool {
	return running && draft != nil && saveplan.For(before, draft.Settings).Restart
}

func missionPoolRows(s form.State, availablePacks, importedPacks []string) []MissionPoolRow {
	floor, hasFloor := gamedata.DifficultyByKey(s.Settings.MvmDifficulty)
	missions := runshape.VisibleMissions(availablePacks)
	rows := make([]MissionPoolRow, 0, len(missions))
	for _, mission := range missions {
		compatibility := gamedata.RequirementLabel(gamedata.MissionRequirement(mission.ID))
		if gamedata.IsPlayableMission(mission.ID) {
			compatibility = "Ready"
			if hasFloor && mission.Difficulty < floor {
				compatibility = "Below " + floor.String() + " floor"
			}
		}
		mods := "—"
		if key := gamedata.MissionServerMod(mission.ID); key != "" {
			if mod, ok := gamedata.ServerModByKey(key); ok {
				mods = mod.Name
			} else {
				mods = key
			}
		}
		rows = append(rows, MissionPoolRow{
			Field:         "missions.pool." + mission.PopFile,
			Source:        missionSource(mission, importedPacks),
			Name:          mission.Name,
			Waves:         fmt.Sprintf("1–%d", mission.Waves),
			Compatibility: compatibility,
			Mods:          mods,
		})
	}
	return rows
}

func missionSource(mission gamedata.Mission, importedPacks []string) string {
	if slices.Contains(importedPacks, gamedata.MissionPack(mission.ID)) {
		return "Imported"
	}
	switch gamedata.MissionPack(mission.ID) {
	case settings.CommunityPackPotato:
		return "Potato Archive"
	case settings.CommunityPackMoonlight:
		return "Moonlight Archive"
	default:
		return "Valve"
	}
}

func (a *App) append(line apruntime.Line) {
	a.mu.Lock()
	a.logs = append(a.logs, line)
	if len(a.logs) > linesMax {
		a.logs = a.logs[len(a.logs)-linesMax:]
	}
	restart := false
	addressChanged := false
	if a.logFile != nil {
		_, _ = fmt.Fprintf(a.logFile, "%s  %-8s %s\n", line.At.Format("15:04:05"), line.Source, line.Text)
	}
	if line.Source == "srcds" {
		if address := apruntime.FakeIPAddress(strings.TrimSpace(line.Text)); address != "" {
			addressChanged = address != a.steamURL
			a.steamURL = address
		}
		if mission := apruntime.LoadedMission(line.Text); mission != "" {
			a.mission = mission
		}
		if note := apruntime.ItemServerLine(line.Text); note != "" {
			a.itemServer = note
		}
		if apruntime.SourceModWasUpdated(line.Text) && !a.smAsked {
			a.smAsked = true
			a.smHeld = a.draft != nil
			restart = !a.smHeld
		}
	}
	a.publishLocked(event{Name: "log", Data: line})
	if addressChanged {
		a.publishLocked(event{Name: "state", Data: struct{}{}})
	}
	a.mu.Unlock()
	if restart {
		a.say("SourceMod updated its gamedata. Restarting the server to load it.")
		a.Restart()
	}
}

func (a *App) say(format string, args ...any) {
	a.append(apruntime.Line{At: time.Now(), Source: "launcher", Text: fmt.Sprintf(format, args...)})
}

func (a *App) notify(text string) {
	a.mu.Lock()
	a.notice = text
	a.noticeSeq++
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
	a.say("%s", text)
}

func (a *App) publishState() {
	a.mu.Lock()
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
}

func (a *App) publishLocked(message event) {
	for subscriber := range a.subscribers {
		select {
		case subscriber <- message:
		default:
		}
	}
}

// Starting the competing server processes is application-lifetime work. It
// deliberately survives the HTTP request that triggered it.
//
//nolint:contextcheck // The server lifecycle is independent of browser requests.
func (a *App) Start() {
	a.mu.Lock()
	if a.busy || a.supervisor.Running() {
		a.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.busy, a.install, a.steamURL, a.mission = true, cancel, "", ""
	s := a.settings
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
	go apruntime.Guard("the browser interface", a.sayLine, func() {
		defer func() {
			cancel()
			a.mu.Lock()
			a.busy, a.install = false, nil
			a.publishLocked(event{Name: "state", Data: struct{}{}})
			a.mu.Unlock()
		}()
		if _, err := installer.Ensure(ctx, s.InstallRoot, settings.CommunityArchives(s), func(f string, args ...any) {
			a.append(apruntime.Line{At: time.Now(), Source: "install", Text: fmt.Sprintf(f, args...)})
		}); err != nil {
			if ctx.Err() == nil {
				a.say("install failed: %v", err)
				a.say("%s.", installer.RepairAdvice)
			}
			return
		}
		for _, line := range apruntime.ConnectLines(s) {
			a.say("%s", line)
		}
		if err := a.supervisor.Start(func(err error) {
			if err != nil {
				a.say("%v", err)
			}
			a.publishState()
		}); err != nil {
			a.say("%v", err)
			var funnel *apruntime.TailscaleFastDLStartError
			if errors.As(err, &funnel) && funnel.ApprovalURL != "" {
				a.notify("Tailscale Funnel approval required: " + funnel.ApprovalURL)
			}
		}
	})
}

func (a *App) Stop() {
	a.mu.Lock()
	cancel := a.install
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	a.supervisor.Stop()
	a.publishState()
}

func (a *App) Restart() {
	go func() {
		a.Stop()
		a.Start()
	}()
}

func (a *App) Quit() { a.quitOnce.Do(func() { close(a.quit) }) }

func (a *App) SendRCON(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}
	a.append(apruntime.Line{At: time.Now(), Source: "rcon", Text: "> " + command})
	go apruntime.Guard("an RCON command", a.sayLine, func() {
		client, err := dialRCON(a.supervisor.Settings())
		if err != nil {
			a.say("rcon: %v", err)
			return
		}
		defer func() { _ = client.Close() }()
		reply, err := client.Exec(command)
		if err != nil {
			a.say("rcon: %v", err)
			return
		}
		for line := range strings.SplitSeq(reply, "\n") {
			if strings.TrimSpace(line) != "" {
				a.append(apruntime.Line{At: time.Now(), Source: "rcon", Text: line})
			}
		}
	})
}

func dialRCON(s settings.Settings) (*rcon.Client, error) {
	var last error
	for _, address := range apruntime.RconAddresses(s) {
		client, err := rcon.Dial(address, s.SrcdsRconPw)
		if err == nil {
			return client, nil
		}
		last = err
	}
	return nil, last
}

func (a *App) watchSession() {
	ticker := time.NewTicker(sessionEvery)
	defer ticker.Stop()
	for {
		select {
		case <-a.quit:
			return
		case <-ticker.C:
			if !a.supervisor.Running() {
				continue
			}
			snapshot, err := session.Fetch(context.Background(), session.BridgeURL)
			a.mu.Lock()
			a.snapshot, a.fetchErr = snapshot, err
			a.publishLocked(event{Name: "state", Data: struct{}{}})
			a.mu.Unlock()
		}
	}
}

func (a *App) OpenSettings(page string) {
	a.mu.Lock()
	state := form.NewState(a.settings)
	a.draft, a.formPage = &state, page
	a.community = availableCommunityPackNames(a.settings.CommunityContentDir)
	a.imported = importedCommunityPackNames(a.settings.CommunityContentDir, a.community)
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
}

func (a *App) Change(c form.Change) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return errors.New("settings are not open")
	}
	next, err := form.Apply(*a.draft, a.formEnvLocked(), c)
	if err != nil {
		return err
	}
	*a.draft = next
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	return nil
}

func (a *App) SaveSettings(restart bool) error {
	a.mu.Lock()
	if a.draft == nil {
		a.mu.Unlock()
		return errors.New("settings are not open")
	}
	draft := *a.draft
	a.mu.Unlock()
	room, roomErr := settings.ParseRoom(draft.Draft.Room)
	if roomErr == nil {
		draft.Settings.APHost, draft.Settings.APPort, draft.Settings.APTls = room.Host, room.Port, room.TLS
	} else if strings.TrimSpace(draft.Draft.Room) == "" {
		draft.Settings.APHost, draft.Settings.APPort = "", 0
	}
	written, err := settings.Persist(draft.Settings)
	if err != nil {
		return err
	}
	before := a.supervisor.Settings()
	a.mu.Lock()
	a.settings, a.draft = written, nil
	a.notice = "settings saved"
	a.noticeSeq++
	heldRestart := a.smHeld
	a.smHeld = false
	a.mu.Unlock()
	a.supervisor.SetSettings(written)
	if _, err := settings.WritePlayerFile(written, assets.ArchipelagoVersion); err != nil {
		a.say("%v", err)
	}
	if a.supervisor.Running() {
		plan := saveplan.For(before, written)
		if plan.Team {
			if err := srcdsconfig.Install(written); err != nil {
				a.say("cannot write the bot files: %v", err)
			} else {
				for _, command := range botlive.Commands(before, written) {
					a.SendRCON(command)
				}
			}
		}
		switch {
		case plan.Restart && restart:
			a.say("settings saved. Restarting the server to apply them.")
			a.Restart()
		case heldRestart:
			a.say("SourceMod updated its gamedata. Restarting the server to load it.")
			a.Restart()
		case plan.Restart:
			a.say("settings saved. The server is still playing on what it started with: press Restart to apply them.")
		case plan.Quiet():
			a.say("settings saved. The server keeps playing: nothing here changes a run it is already in.")
		}
	}
	go a.reportRoom(written, draft.Draft.Room, roomErr)
	a.publishState()
	return nil
}

func (a *App) reportRoom(s settings.Settings, typed string, parseErr error) {
	if parseErr != nil && strings.TrimSpace(typed) != "" {
		a.notify("the room address was not saved: " + parseErr.Error() + ". " + roomcheck.NotConfigured.Advice())
		return
	}
	result, err := roomcheck.Check(context.Background(), s)
	if err != nil {
		a.notify("settings saved, but the room did not answer: " + err.Error() + ". " + result.Advice())
		return
	}
	a.notify("settings saved. " + result.Advice())
}

func (a *App) CancelSettings() {
	a.mu.Lock()
	a.draft = nil
	heldRestart := a.smHeld
	a.smHeld = false
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
	if heldRestart {
		a.say("SourceMod updated its gamedata. Restarting the server to load it.")
		a.Restart()
	}
}

func (a *App) formEnvLocked() form.Env {
	dirs := generate.SearchPath("")
	appDir := ""
	if len(dirs) > 0 {
		appDir = dirs[0]
	}
	return form.Env{CommunityAvailable: slices.Clone(a.community), AppDirDefault: appDir}
}

func availableCommunityPackNames(folder string) []string {
	paths := installer.AvailableCommunityArchives(settings.KnownCommunityArchives(strings.TrimSpace(folder)))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, filepath.Base(path))
	}
	return result
}

func importedCommunityPackNames(folder string, available []string) []string {
	var imported []string
	for _, name := range available {
		if _, err := os.Stat(filepath.Join(folder, name+".imported")); err == nil {
			imported = append(imported, name)
		}
	}
	return imported
}

func recognizedCommunityArchive(name string) (string, bool) {
	name = strings.ToLower(filepath.Base(name))
	switch name {
	case settings.CommunityPackPotato, settings.CommunityPackMoonlight:
		return name, true
	default:
		return "", false
	}
}

func importCommunityArchive(folder, name string, source io.Reader) (string, error) {
	originalName := name
	name, recognized := recognizedCommunityArchive(name)
	if !recognized {
		return "", fmt.Errorf("unrecognized asset pack %q; choose %s or %s", originalName, settings.CommunityPackPotato, settings.CommunityPackMoonlight)
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", fmt.Errorf("cannot create the asset cache: %w", err)
	}
	temporary, err := os.CreateTemp(folder, ".tf2ap-import-*.zip")
	if err != nil {
		return "", fmt.Errorf("cannot stage %s: %w", name, err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if _, err := io.Copy(temporary, source); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("cannot import %s: %w", name, err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("cannot finish %s: %w", name, err)
	}
	if err := installer.ValidateCommunityArchives([]string{temporaryName}, func(string, ...any) {}); err != nil {
		return "", err
	}
	target := filepath.Join(folder, name)
	if err := os.Rename(temporaryName, target); err != nil {
		if removeErr := os.Remove(target); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return "", fmt.Errorf("cannot replace %s: %w", name, removeErr)
		}
		if err := os.Rename(temporaryName, target); err != nil {
			return "", fmt.Errorf("cannot keep %s: %w", name, err)
		}
	}
	if err := os.WriteFile(target+".imported", nil, 0o644); err != nil {
		return "", fmt.Errorf("cannot mark %s as imported: %w", name, err)
	}
	return name, nil
}

func (a *App) importAssets(request *http.Request) ([]string, error) {
	a.mu.Lock()
	if a.draft == nil {
		a.mu.Unlock()
		return nil, errors.New("settings are not open")
	}
	folder := a.draft.Settings.CommunityContentDir
	a.mu.Unlock()

	reader, err := request.MultipartReader()
	if err != nil {
		return nil, fmt.Errorf("asset import needs ZIP files: %w", err)
	}
	var imported []string
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cannot read imported assets: %w", err)
		}
		if part.FileName() == "" {
			_ = part.Close()
			continue
		}
		name, err := importCommunityArchive(folder, part.FileName(), part)
		_ = part.Close()
		if err != nil {
			return nil, err
		}
		if !slices.Contains(imported, name) {
			imported = append(imported, name)
		}
	}
	if len(imported) == 0 {
		return nil, errors.New("choose archive-assets.zip or mlarchive-assets.zip")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return nil, errors.New("settings were closed while the assets were importing")
	}
	a.community = availableCommunityPackNames(folder)
	a.imported = importedCommunityPackNames(folder, a.community)
	a.draft.Settings.MvmCommunityMissions = true
	for _, name := range imported {
		if !slices.Contains(a.draft.Settings.CommunityPacks, name) {
			a.draft.Settings.CommunityPacks = append(a.draft.Settings.CommunityPacks, name)
		}
		for _, mission := range gamedata.PlayableMissions() {
			if gamedata.MissionPack(mission.ID) == name {
				a.draft.Settings.MvmExcludedMissions = slices.DeleteFunc(
					a.draft.Settings.MvmExcludedMissions,
					func(popFile string) bool { return popFile == mission.PopFile })
			}
		}
	}
	a.publishLocked(event{Name: "state", Data: struct{}{}})
	return imported, nil
}

func (a *App) Dispatch(id string) error {
	a.mu.Lock()
	if a.draft == nil {
		a.mu.Unlock()
		return errors.New("settings are not open")
	}
	if !form.Dispatchable(*a.draft, a.formEnvLocked(), id) {
		a.mu.Unlock()
		return fmt.Errorf("no settings action %q", id)
	}
	s := *a.draft
	a.mu.Unlock()

	switch id {
	case "missions.pool_all", "missions.pool_none":
		a.setPool(id == "missions.pool_all")
	case "bots.save_team":
		a.saveTeam(s)
	case "bots.remove_team":
		a.removeTeam(s)
	case "loadout.save":
		a.saveLoadout(s)
	case "missions.check_selection":
		a.checkMissionSelection(s.Settings)
	case "missions.download_packs":
		go a.downloadPacks(s.Settings)
	case "missions.import_assets":
		a.notify("choose the local asset ZIPs in the browser")
	case "server.repair":
		go a.repair(s.Settings.InstallRoot)
	case "server.reset":
		return a.resetSettings()
	default:
		return fmt.Errorf("settings action %q is not wired", id)
	}
	return nil
}

func (a *App) saveTeam(s form.State) {
	name := strings.TrimSpace(s.Draft.TeamName)
	if name == "" {
		a.notify("name the team first")
		return
	}
	presets := maps.Clone(s.Settings.SrcdsBotTeamPresets)
	if presets == nil {
		presets = map[string]settings.BotTeam{}
	}
	presets[name] = settings.BotTeamOf(s.Settings)
	a.mutateDraft(func(state *form.State) {
		state.Settings.SrcdsBotTeamPresets, state.Draft.TeamName = presets, ""
	})
	a.notify("saved the team as " + name)
}

func (a *App) removeTeam(s form.State) {
	name := strings.TrimSpace(s.Draft.TeamName)
	if name == "" {
		a.notify("name the team to remove first")
		return
	}
	if _, ok := s.Settings.SrcdsBotTeamPresets[name]; !ok {
		a.notify("no team saved as " + name)
		return
	}
	presets := maps.Clone(s.Settings.SrcdsBotTeamPresets)
	delete(presets, name)
	if len(presets) == 0 {
		presets = nil
	}
	a.mutateDraft(func(state *form.State) {
		state.Settings.SrcdsBotTeamPresets, state.Draft.TeamName = presets, ""
	})
	a.notify("removed the team " + name)
}

func (a *App) saveLoadout(s form.State) {
	name := strings.TrimSpace(s.Draft.LoadoutName)
	if name == "" {
		a.notify("name the loadout first")
		return
	}
	built := maps.Clone(s.Settings.SrcdsBotCustomLoadouts)
	if built == nil {
		built = map[string]botloadout.Built{}
	}
	built[name] = s.Draft.Loadout
	a.mutateDraft(func(state *form.State) { state.Settings.SrcdsBotCustomLoadouts = built })
	a.notify("saved the loadout as " + name)
}

func (a *App) checkMissionSelection(s settings.Settings) {
	result, err := settings.CheckRunSelection(s)
	if err != nil {
		a.notify(err.Error())
		return
	}
	a.notify(result.Summary())
}

var wiredActions = []string{
	"run.generate", "run.open_player_file", "run.open_folder", "run.open_settings_file",
	"missions.download_packs", "missions.import_assets", "missions.check_selection",
	"missions.pool_all", "missions.pool_none",
	"server.debug_bundle", "server.repair", "server.reset",
	"net.check_funnel", "bots.save_team", "bots.remove_team", "loadout.save",
}

func (a *App) mutateDraft(change func(*form.State)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft != nil {
		change(a.draft)
		a.publishLocked(event{Name: "state", Data: struct{}{}})
	}
}

func (a *App) setPool(all bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return
	}
	var excluded []string
	if !all {
		for _, mission := range gamedata.PlayableMissions() {
			excluded = append(excluded, mission.PopFile)
		}
	} else {
		visible := runshape.VisibleMissions(a.community)
		for _, mission := range gamedata.PlayableMissions() {
			if gamedata.MissionPack(mission.ID) != "" && !slices.ContainsFunc(visible, func(candidate gamedata.Mission) bool { return candidate.ID == mission.ID }) {
				excluded = append(excluded, mission.PopFile)
			}
		}
	}
	a.draft.Settings.MvmExcludedMissions = excluded
	if slices.Contains(excluded, a.draft.Settings.MvmStartMission) {
		a.draft.Settings.MvmStartMission = ""
	}
	a.notice = map[bool]string{true: "every mission is in the pool", false: "every mission is left out"}[all]
	a.noticeSeq++
	a.publishLocked(event{Name: "state", Data: struct{}{}})
}

func (a *App) downloadPacks(s settings.Settings) {
	folder := strings.TrimSpace(s.CommunityContentDir)
	if folder == "" {
		a.notify("choose an asset pack folder first")
		return
	}
	archives := settings.CommunityArchives(s)
	if len(archives) == 0 {
		a.notify("select at least one community pack first")
		return
	}
	if err := installer.DownloadCommunityArchives(context.Background(), archives, func(f string, args ...any) { a.say(f, args...) }); err != nil {
		a.notify("community assets: " + err.Error())
		return
	}
	a.mu.Lock()
	a.community = availableCommunityPackNames(folder)
	a.mu.Unlock()
	a.notify("selected community packs are ready in " + folder)
}

func (a *App) repair(root string) {
	a.Stop()
	_, _ = winproc.KillUnder(root)
	removed, err := installer.Clean(root)
	if err != nil {
		a.notify("repair: " + err.Error())
		return
	}
	if len(removed) == 0 {
		a.notify("repair: nothing to remove")
	} else {
		a.notify("repair removed " + strings.Join(removed, ", "))
	}
}

func (a *App) resetSettings() error {
	fresh := settings.Defaults()
	fresh.InstallRoot = a.supervisor.Settings().InstallRoot
	written, err := settings.Persist(fresh)
	if err != nil {
		return err
	}
	a.supervisor.SetSettings(written)
	a.mu.Lock()
	a.settings = written
	state := form.NewState(written)
	a.draft = &state
	a.mu.Unlock()
	a.notify("every setting is back to its default")
	return nil
}

func (a *App) joinLineLocked() string {
	port := fmt.Sprintf("%d", a.settings.SrcdsPort)
	var parts []string
	if settings.Effective(a.settings.SrcdsReach, a.settings.SrcdsToken) == settings.ReachSteam {
		if a.steamURL == "" {
			parts = append(parts, "Steam public IP: waiting")
		} else {
			parts = append(parts, "Steam public IP: "+a.steamURL)
		}
	}
	for _, address := range apruntime.LocalAddresses() {
		parts = append(parts, address+":"+port)
	}
	if len(parts) == 0 {
		parts = append(parts, "127.0.0.1:"+port)
	}
	line := strings.Join(parts, "   ")
	if a.settings.SrcdsPw != "" {
		line += "   (password " + a.settings.SrcdsPw + ")"
	}
	return line
}

func (a *App) joinURLLocked() string {
	if settings.Effective(a.settings.SrcdsReach, a.settings.SrcdsToken) == settings.ReachSteam && a.steamURL == "" {
		return ""
	}
	return apruntime.SteamConnectURL(a.settings, a.steamURL)
}

func (a *App) sayLine(text string) { a.say("%s", text) }

// Server lifecycle operations deliberately survive the browser request.
func (a *App) serveServerAction(w http.ResponseWriter, r *http.Request) {
	switch r.PathValue("action") {
	case "start":
		a.Start()
	case "stop":
		go a.Stop()
	case "restart":
		a.Restart()
	default:
		http.Error(w, "unknown action", http.StatusNotFound)
	}
}

func decode[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var value T
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return value, false
	}
	return value, true
}

// RCON is queued as application-lifetime work so a browser navigation cannot
// cancel a command after it was accepted.
//
//nolint:contextcheck // Accepted commands survive browser navigation.
func (a *App) serveRCON(w http.ResponseWriter, r *http.Request) {
	if body, ok := decode[struct {
		Command string `json:"command"`
	}](w, r); ok {
		a.SendRCON(body.Command)
	}
}

// Mission changes use the same application-lifetime RCON queue.
//
//nolint:contextcheck // Accepted mission changes survive browser navigation.
func (a *App) serveMission(w http.ResponseWriter, r *http.Request) {
	if body, ok := decode[struct {
		PopFile string `json:"popfile"`
	}](w, r); ok {
		a.SendRCON("sm_ap_mission " + body.PopFile)
	}
}

func (a *App) serveSettingsOpen(w http.ResponseWriter, r *http.Request) {
	body, _ := decode[struct {
		Page string `json:"page"`
	}](w, r)
	a.OpenSettings(body.Page)
}

func (a *App) serveSettingsChange(w http.ResponseWriter, r *http.Request) {
	if c, ok := decode[form.Change](w, r); ok {
		if err := a.Change(c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

// Settings actions can start downloads, repairs and server lifecycle work
// which deliberately survives the browser request.
//
//nolint:contextcheck // Dispatched actions can outlive the request.
func (a *App) serveSettingsAction(w http.ResponseWriter, r *http.Request) {
	if body, ok := decode[struct {
		ID string `json:"id"`
	}](w, r); ok {
		if err := a.Dispatch(body.ID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

// The accepted import is reported to every connected browser, independent of
// the upload request remaining open for the final notification.
func (a *App) serveSettingsImport(w http.ResponseWriter, r *http.Request) {
	imported, err := a.importAssets(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.notify("imported " + strings.Join(imported, ", "))
}

// Saving can restart the server and check the room after persistence; those
// application-lifetime tasks deliberately survive the browser request.
//
//nolint:contextcheck // Post-save lifecycle work outlives the request.
func (a *App) serveSettingsSave(w http.ResponseWriter, r *http.Request) {
	body, ok := decode[struct {
		Restart bool `json:"restart"`
	}](w, r)
	if !ok {
		return
	}
	if err := a.SaveSettings(body.Restart); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// Cancelling can release a deferred server restart, which must survive the
// browser request.
func (a *App) serveSettingsCancel(http.ResponseWriter, *http.Request) { a.CancelSettings() }
