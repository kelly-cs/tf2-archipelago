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
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/tailscalefastdl"
	"github.com/m-this/tf2-archipelago/launcher/internal/webapi"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

//go:embed index.html
var page []byte

// app is the browser page's own layer: the mux, the SSE stream and the file
// downloads. Everything a request means is webapi.App's, so this holds no state
// and adds no behaviour beyond translating HTTP into a method call.
type app struct{ *webapi.App }

func newApp(s settings.Settings, logger *slog.Logger) *app {
	return &app{App: webapi.New(s, logger)}
}

// Run serves the UI, opens it in the desktop browser, and blocks until Quit or
// an operating-system signal. It is one process, not a remotely exposed daemon.
func Run(s settings.Settings, logger *slog.Logger) error {
	a := newApp(s, logger)
	if file, err := apruntime.CreateLogFile(s.InstallRoot); err == nil {
		a.LogTo(file)
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
			a.Say("interface: %v", err)
			a.Quit()
		}
	}()
	go a.WatchSession()

	address := "http://" + listener.Addr().String()
	if err := openOrPrompt(address, os.Stderr, winproc.OpenURL); err != nil {
		a.Say("browser did not open automatically: %v", err)
	}
	a.Say("interface: %s", address)
	if s.APPort != 0 || s.TestMode {
		a.Start()
	} else {
		a.OpenSettings("Archipelago room")
	}

	signals, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	select {
	case <-signals.Done():
	case <-a.Quitting():
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

func (a *app) localHandler() http.Handler {
	return a.handler("127.0.0.1")
}

func (a *app) handler(authority string) http.Handler {
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

func (a *app) servePage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

// Local address discovery is internally bounded and does not perform work on
// behalf of the browser request.
//
//nolint:contextcheck // Address discovery owns its timeout instead of the request.
func (a *app) serveSnapshot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(a.Snapshot()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveBrowserFile(w http.ResponseWriter, r *http.Request, path, disposition string) {
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": filepath.Base(path)}))
	http.ServeFile(w, r, path)
}

func (a *app) serveSettingsFile(w http.ResponseWriter, r *http.Request) {
	path, err := settings.Path()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveBrowserFile(w, r, path, "inline")
}

func (a *app) servePlayerFile(w http.ResponseWriter, r *http.Request) {
	s, err := a.DraftSettings()
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

func (a *app) serveInstallLocation(w http.ResponseWriter, _ *http.Request) {
	s, err := a.DraftSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "Server folder:\n\n%s\n", s.InstallRoot)
}

func (a *app) serveGeneratedSeed(w http.ResponseWriter, r *http.Request) {
	s, err := a.DraftSettings()
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
func (a *app) serveDebugBundle(w http.ResponseWriter, r *http.Request) {
	s, err := a.DraftSettings()
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

func (a *app) serveFunnelApproval(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	result, err := tailscalefastdl.Authorize(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if result.ApprovalURL != "" {
		http.Redirect(w, r, result.ApprovalURL, http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, "Tailscale Funnel is ready for this tailnet. You can close this tab.")
}

func (a *app) serveEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	listener, done := a.Subscribe()
	defer done()
	_, _ = fmt.Fprint(w, "event: state\ndata: {}\n\n")
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case message := <-listener.Events():
			data, err := json.Marshal(message.Data)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", message.Name, data)
			flusher.Flush()
		}
	}
}

func (a *app) importAssets(request *http.Request) ([]string, error) {
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
		name, err := a.ImportArchive(part.FileName(), part)
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
	if err := a.AcceptImportedPacks(imported); err != nil {
		return nil, err
	}
	return imported, nil
}

// Server lifecycle operations deliberately survive the browser request.
func (a *app) serveServerAction(w http.ResponseWriter, r *http.Request) {
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
func (a *app) serveRCON(w http.ResponseWriter, r *http.Request) {
	if body, ok := decode[struct {
		Command string `json:"command"`
	}](w, r); ok {
		a.SendRCON(body.Command)
	}
}

// Mission changes use the same application-lifetime RCON queue.
//
//nolint:contextcheck // Accepted mission changes survive browser navigation.
func (a *app) serveMission(w http.ResponseWriter, r *http.Request) {
	if body, ok := decode[struct {
		PopFile string `json:"popfile"`
	}](w, r); ok {
		a.SendRCON("sm_ap_mission " + body.PopFile)
	}
}

func (a *app) serveSettingsOpen(w http.ResponseWriter, r *http.Request) {
	body, _ := decode[struct {
		Page string `json:"page"`
	}](w, r)
	a.OpenSettings(body.Page)
}

func (a *app) serveSettingsChange(w http.ResponseWriter, r *http.Request) {
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
func (a *app) serveSettingsAction(w http.ResponseWriter, r *http.Request) {
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
func (a *app) serveSettingsImport(w http.ResponseWriter, r *http.Request) {
	imported, err := a.importAssets(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.Notify("imported " + strings.Join(imported, ", "))
}

// Saving can restart the server and check the room after persistence; those
// application-lifetime tasks deliberately survive the browser request.
//
//nolint:contextcheck // Post-save lifecycle work outlives the request.
func (a *app) serveSettingsSave(w http.ResponseWriter, r *http.Request) {
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
func (a *app) serveSettingsCancel(http.ResponseWriter, *http.Request) { a.CancelSettings() }
