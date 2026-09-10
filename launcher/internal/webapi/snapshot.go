package webapi

import (
	"fmt"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botlive"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/saveplan"
	"github.com/m-this/tf2-archipelago/launcher/internal/session"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// Snapshot is everything the page needs for one draw. Passwords never cross
// the loopback boundary; form already represents them as replacement fields.
type Snapshot struct {
	Title         string           `json:"title"`
	Slot          string           `json:"slot,omitempty"`
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
	Map           string `json:"map"`
	Name          string `json:"name"`
	Waves         string `json:"waves"`
	Compatibility string `json:"compatibility"`
	Mods          string `json:"mods"`
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
	screen := a.screenLocked(running)
	result := Snapshot{
		Title:  assets.Title("Mann vs Archipelago"),
		Slot:   s.APSlotName,
		Status: status, Running: running, Busy: a.busy, Room: room,
		Join: a.joinLineLocked(), JoinURL: a.joinURLLocked(), Mission: playing,
		Logs: slices.Clone(a.logs), Session: sessionState,
		Bots: botlive.Team(s), DrawnBots: botlive.Drawn(s),
		Form: screen.Form, FormPage: screen.Page, Notice: a.notice, NoticeSeq: a.noticeSeq,
		ItemServer: a.itemServer, MissionPool: screen.MissionPool, RestartNeeded: screen.RestartNeeded,
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
		played, _ := gamedata.MapByID(mission.Map)
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
			Map:           played.Name,
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

// Screen is the settings part of one draw: the rows, the page they are on, the
// mission table beside them and whether saving would restart a running server.
// A closed settings screen has a nil Form, which is not the same as a Form with
// no tabs.
type Screen struct {
	Form          *form.Model
	Page          string
	MissionPool   []MissionPoolRow
	RestartNeeded bool
}

// Screen answers what the settings screen holds now. Every settings call
// answers with one, because a change can move another row's bounds, add and
// remove rows outright, and move the mission table under them.
func (a *App) Screen() Screen {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.screenLocked(a.supervisor.Running())
}

// screenLocked builds it. Passwords are blanked here rather than at the wire,
// because there is one Screen and every interface reads this one.
func (a *App) screenLocked(running bool) Screen {
	if a.draft == nil {
		return Screen{Page: a.formPage}
	}
	built := form.Build(*a.draft, a.formEnvLocked())
	for page := range built.Tabs {
		for field := range built.Tabs[page].Fields {
			if built.Tabs[page].Fields[field].Kind == form.Password {
				built.Tabs[page].Fields[field].Value = ""
			}
		}
	}
	return Screen{
		Form:          &built,
		Page:          a.formPage,
		MissionPool:   missionPoolRows(*a.draft, a.community, a.imported),
		RestartNeeded: restartNeeded(running, a.settings, a.draft),
	}
}
