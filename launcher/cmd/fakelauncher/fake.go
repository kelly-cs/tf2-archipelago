package main

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/m-this/tf2-archipelago/launcher/internal/botlive"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	launcherv1 "github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1"
	"github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1/launcherv1connect"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/session"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/webapi"
)

// The clock the log lines carry. Fixed, so a screenshot of the fake is the same
// picture on every run and a diff of two of them is a real difference.
var fakeClock = time.Date(2026, time.September, 10, 20, 15, 0, 0, time.UTC)

// fake holds one run's worth of made-up state. Every method is behind mu: the
// browser talks over several connections at once.
type fake struct {
	mu        sync.Mutex
	settings  settings.Settings
	draft     *form.State
	running   bool
	mission   string
	logs      []apruntime.Line
	notice    string
	noticeSeq uint64
	listeners map[chan *launcherv1.StreamMessage]struct{}
}

func newFake() *fake {
	base := settings.Defaults()
	base.APHost, base.APPort, base.APSlotName = "archipelago.gg", 38281, "Scout"
	base.InstallRoot = "/home/player/tf2-archipelago"

	f := &fake{
		settings:  base,
		listeners: make(map[chan *launcherv1.StreamMessage]struct{}),
	}
	for i, text := range []string{
		"launcher ready",
		"Server is hibernating",
		"Executing dedicated server config: server.cfg",
		"[SM] Loaded plugin tf2_archipelago.smx",
		"Warning: sv_pure is not set",
		"bridge connected to archipelago.gg:38281 as Scout",
	} {
		source := "srcds"
		if i == 0 || i == 5 {
			source = "launcher"
		}
		f.logs = append(f.logs, apruntime.Line{
			At: fakeClock.Add(time.Duration(i) * time.Second), Source: source, Text: text,
		})
	}
	return f
}

func (f *fake) register(mux *http.ServeMux, authority string) {
	options := connect.WithInterceptors(validate.NewInterceptor())
	mux.Handle(launcherv1connect.NewLauncherServiceHandler(launcherRPC{f}, options))
	mux.Handle(launcherv1connect.NewSettingsServiceHandler(settingsRPC{f}, options))
	mux.Handle(launcherv1connect.NewFilesServiceHandler(filesRPC{f}, options))
	mux.HandleFunc("GET /ws", f.serveStream(authority))
	mux.HandleFunc("POST /fake/reset", f.serveReset)
}

// serveReset puts the run back to how it started. It is the fake's own control
// surface and no part of the launcher's contract: the browser tests use it so
// each one begins from the same place, whatever the one before it pressed.
func (f *fake) serveReset(w http.ResponseWriter, _ *http.Request) {
	fresh := newFake()
	f.mu.Lock()
	f.settings, f.draft = fresh.settings, nil
	f.running, f.mission = false, ""
	f.logs = fresh.logs
	f.notice, f.noticeSeq = "", 0
	f.redrawLocked()
	f.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// snapshot builds the same webapi.Snapshot the real launcher would, so the
// browser sees the real encoding of a real form.Model.
func (f *fake) snapshotLocked() webapi.Snapshot {
	status := "stopped"
	if f.running {
		status = "running"
	}
	room := "room archipelago.gg:38281"
	if f.mission != "" {
		room += "   " + f.mission
	}
	snapshot := webapi.Snapshot{
		Title:      "Mann vs Archipelago (fake)",
		Status:     status,
		Running:    f.running,
		Room:       room,
		Join:       "connect 127.0.0.1:27015; password \"\"",
		JoinURL:    "steam://run/440//+connect%20127.0.0.1:27015",
		Mission:    f.mission,
		Logs:       slices.Clone(f.logs),
		Session:    fakeSession(f.running),
		Bots:       botlive.Team(f.settings),
		DrawnBots:  botlive.Drawn(f.settings),
		Notice:     f.notice,
		NoticeSeq:  f.noticeSeq,
		ItemServer: "item server: ready",
	}
	screen := f.screenLocked()
	snapshot.Form, snapshot.FormPage = screen.Form, screen.Page
	snapshot.MissionPool, snapshot.RestartNeeded = screen.MissionPool, screen.RestartNeeded
	return snapshot
}

func (f *fake) screenLocked() webapi.Screen {
	if f.draft == nil {
		return webapi.Screen{}
	}
	model := form.Build(*f.draft, form.Env{})
	for tab := range model.Tabs {
		for field := range model.Tabs[tab].Fields {
			if model.Tabs[tab].Fields[field].Kind == form.Password {
				model.Tabs[tab].Fields[field].Value = ""
			}
		}
	}
	return webapi.Screen{Form: &model, Page: "Archipelago room", MissionPool: fakePool(*f.draft)}
}

func (f *fake) say(text string) {
	f.mu.Lock()
	line := apruntime.Line{At: fakeClock.Add(time.Duration(len(f.logs)) * time.Second), Source: "launcher", Text: text}
	f.logs = append(f.logs, line)
	f.publishLocked(&launcherv1.StreamMessage{
		Body: &launcherv1.StreamMessage_Line{Line: &launcherv1.LogLine{
			Source: line.Source, Text: line.Text,
		}},
	})
	f.mu.Unlock()
}

func (f *fake) redraw() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.redrawLocked()
}

func (f *fake) redrawLocked() {
	state := f.snapshotLocked().Proto()
	state.Logs = nil
	f.publishLocked(&launcherv1.StreamMessage{
		Body: &launcherv1.StreamMessage_Snapshot{Snapshot: state},
	})
}

func (f *fake) publishLocked(message *launcherv1.StreamMessage) {
	for listener := range f.listeners {
		select {
		case listener <- message:
		default:
		}
	}
}

func (f *fake) serveStream(authority string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		socket, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{authority}})
		if err != nil {
			return
		}
		defer func() { _ = socket.CloseNow() }()
		ctx := socket.CloseRead(r.Context())

		events := make(chan *launcherv1.StreamMessage, 128)
		f.mu.Lock()
		f.listeners[events] = struct{}{}
		first := &launcherv1.StreamMessage{
			Body: &launcherv1.StreamMessage_Snapshot{Snapshot: f.snapshotLocked().Proto()},
		}
		f.mu.Unlock()
		defer func() {
			f.mu.Lock()
			delete(f.listeners, events)
			f.mu.Unlock()
		}()

		if err := writeFrame(ctx, socket, first); err != nil {
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case message := <-events:
				if err := writeFrame(ctx, socket, message); err != nil {
					return
				}
			}
		}
	}
}

func writeFrame(ctx context.Context, socket *websocket.Conn, message *launcherv1.StreamMessage) error {
	frame, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return socket.Write(ctx, websocket.MessageBinary, frame)
}

func fakeSession(running bool) session.Snapshot {
	return session.Snapshot{
		Health: session.Health{
			Connected: running, Slot: "Scout", Seed: "PROOF-OF-A-SEED",
			Checks: 12, Items: 7, DeathLink: true,
		},
		Missions: []session.Mission{
			{PopFile: "mvm_decoy_advanced", Name: "Doe's Doom", Map: "Decoy", Waves: 7, Source: "Valve", Unlocked: true, Played: true},
			{PopFile: "mvm_coaltown_advanced", Name: "Caliginous Caper", Map: "Coal Town", Waves: 6, Source: "Valve", Unlocked: true},
			{PopFile: "mvm_mannworks_advanced", Name: "Mean Machines", Map: "Mannworks", Waves: 7, Source: "Valve", Cleared: true},
			{PopFile: "mvm_rottenburg_advanced", Name: "Bavarian Botbash", Map: "Rottenburg", Waves: 7, Source: "Valve"},
		},
		Unlocks: []session.Unlock{
			{Kind: "Class", Name: "Scout"},
			{Kind: "Class", Name: "Soldier"},
			{Kind: "Weapon slot", Name: "Soldier, secondary"},
			{Kind: "Weapon buff", Name: "Scattergun damage", Level: 2},
			{Kind: "Mission", Name: "Doe's Doom"},
		},
	}
}

// fakePool is the mission table with the compatibility words the real launcher
// uses. The tick still lives on the form row, so the table only carries what a
// human-readable label would otherwise have to be parsed for.
func fakePool(state form.State) []webapi.MissionPoolRow {
	rows := make([]webapi.MissionPoolRow, 0, 8)
	for i, mission := range []struct{ name, place, waves, compat, mods string }{
		{"Doe's Doom", "Decoy", "1-7", "Ready", "—"},
		{"Caliginous Caper", "Coal Town", "1-6", "Ready", "—"},
		{"Mean Machines", "Mannworks", "1-7", "Ready", "—"},
		{"Bavarian Botbash", "Rottenburg", "1-7", "Ready", "—"},
		{"Ghost Town", "Ghost Town", "1-6", "Below Advanced floor", "—"},
		{"Hamlet Hostility", "Hamlet", "1-6", "Medieval", "—"},
		{"Trouble in Mann Town", "Coal Town", "1-6", "Community missions are off", "SigMod"},
	} {
		rows = append(rows, webapi.MissionPoolRow{
			Field:         "missions.pool.fake_" + strconv.Itoa(i),
			Source:        "Valve",
			Map:           mission.place,
			Name:          mission.name,
			Waves:         mission.waves,
			Compatibility: mission.compat,
			Mods:          mission.mods,
		})
	}
	_ = state
	return rows
}
