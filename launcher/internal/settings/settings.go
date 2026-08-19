// Package settings is the launcher's persisted user configuration. One JSON
// file under the OS's user config dir (on Windows, %APPDATA%\tf2ap\config.json),
// written on every successful start, restored on the next launch. It is the
// "last used parameters" memory the prompt asked for.
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/m-this/tf2-archipelago/gamedata"
)

// Settings is everything the launcher asks for and remembers. Fields are the
// .env values from the compose stack, renamed to the launcher's vocabulary.
type Settings struct {
	// InstallRoot is where the 14 GB of game files, SourceMod and steamcmd
	// live. Default is <home>/tf2-archipelago.
	InstallRoot string `json:"install_root"`

	// ArchipelagoDir is where the Archipelago app is, when it is not where its
	// installer puts it. Empty means look in the usual places.
	ArchipelagoDir string `json:"archipelago_dir,omitempty"`

	// TestMode plays without Archipelago: the launcher serves a multiworld of
	// one on loopback, makes up a seed, and hands out unlocks as waves are
	// cleared. For trying the server out, and for play-testing.
	TestMode bool `json:"test_mode"`

	// Archipelago room.
	APHost     string `json:"ap_host"`
	APPort     int    `json:"ap_port"`
	APTls      bool   `json:"ap_tls"`
	APSlotName string `json:"ap_slot_name"`
	APPassword string `json:"ap_password,omitempty"`

	// Game server.
	SrcdsHostname      string `json:"srcds_hostname"`
	SrcdsRconPw        string `json:"srcds_rcon_pw,omitempty"`
	SrcdsPw            string `json:"srcds_pw,omitempty"`
	SrcdsPort          int    `json:"srcds_port"`
	SrcdsMaxPlayers    int    `json:"srcds_max_players"`
	SrcdsToken         string `json:"srcds_token"`
	SrcdsLan           bool   `json:"srcds_lan"`
	SrcdsAdminSteamIDs string `json:"srcds_admin_steamids,omitempty"`

	// SrcdsStartMission is the popfile the server loads first. The map comes
	// with it: gamedata knows which map a mission runs on.
	SrcdsStartMission string `json:"srcds_start_mission"`

	// SrcdsStartMap is what older files hold. Read for the migration to
	// SrcdsStartMission and never written again.
	SrcdsStartMap string `json:"srcds_start_map,omitempty"`

	// SrcdsSteamNetworking relays the traffic over Steam Datagram Relay when
	// the server is not LAN: no port to forward, and Valve hands out the
	// address. Needs the login token like any public server.
	SrcdsSteamNetworking bool `json:"srcds_steam_networking"`

	// Defender bots. RED is filled to SrcdsBotTeamSize when a wave begins.
	SrcdsBots        bool `json:"srcds_bots"`
	SrcdsBotTeamSize int  `json:"srcds_bot_team_size"`

	// SrcdsBotClassBlacklist names the classes the bots never play, in the
	// mod's spelling (heavyweapons, not heavy). SrcdsBotLoadouts picks a
	// loadout preset per class, keyed the same way; a class not in it plays
	// stock.
	SrcdsBotClassBlacklist []string          `json:"srcds_bot_class_blacklist,omitempty"`
	SrcdsBotLoadouts       map[string]string `json:"srcds_bot_loadouts,omitempty"`

	// BotUpgradesChat writes what the bots buy at the upgrade station to the
	// chat. Off by default: it is a line per purchase.
	BotUpgradesChat bool `json:"bot_upgrades_chat"`

	// Run shape, for seed generation guidance (the launcher does not generate
	// seeds itself, but it can write a starter YAML for the Archipelago app).
	MvmMissionCount     int      `json:"mvm_mission_count"`
	MvmDifficulty       string   `json:"mvm_difficulty"`
	MvmGoal             string   `json:"mvm_goal"`
	MvmMissionsanityPct int      `json:"mvm_missionsanity_percentage"`
	MvmDeathLink        bool     `json:"mvm_death_link"`
	MvmExcludedMissions []string `json:"mvm_excluded_missions,omitempty"`

	// Whether to enable the metrics listener and on what port.
	MetricsPort int `json:"metrics_port"`
}

// Defaults returns the factory settings, matching deploy/.env.example.
func Defaults() Settings {
	return Settings{
		InstallRoot:         defaultInstallRoot(),
		APHost:              "archipelago.gg",
		APPort:              0,
		APTls:               true,
		APSlotName:          "tf2",
		SrcdsHostname:       "Mann vs Archipelago",
		SrcdsPort:           27015,
		SrcdsMaxPlayers:     32,
		SrcdsStartMission:   "mvm_decoy",
		SrcdsToken:          "0",
		SrcdsLan:            true,
		SrcdsBots:           true,
		SrcdsBotTeamSize:    6,
		MvmMissionCount:     8,
		MvmDifficulty:       "intermediate",
		MvmGoal:             "final_boss",
		MvmMissionsanityPct: 80,
		MetricsPort:         24681,
	}
}

// Path returns the config file location. On Windows this is
// %APPDATA%\tf2ap\config.json; on Linux ~/.config/tf2ap/config.json.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot find the user config directory: %w", err)
	}
	return filepath.Join(dir, "tf2ap", "config.json"), nil
}

// Load reads the config file, applying defaults for anything unset. A missing
// file is not an error: it returns Defaults.
func Load() (Settings, error) {
	path, err := Path()
	if err != nil {
		return Settings{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Defaults(), nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	s = s.withDefaults()
	return s, nil
}

// Render returns the settings as the JSON that Save writes. The debug bundle
// uses it to ship a copy with the passwords taken out.
func Render(s Settings) (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// Save writes the config file, creating the directory.
func Save(s Settings) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cannot create %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	return nil
}

// withDefaults fills zero values with the defaults, so an old config file that
// predates a field still works.
func (s Settings) withDefaults() Settings {
	d := Defaults()
	if s.InstallRoot == "" {
		s.InstallRoot = d.InstallRoot
	}
	if s.APSlotName == "" {
		s.APSlotName = d.APSlotName
	}
	if s.SrcdsHostname == "" {
		s.SrcdsHostname = d.SrcdsHostname
	}
	if s.SrcdsPort == 0 {
		s.SrcdsPort = d.SrcdsPort
	}
	if s.SrcdsMaxPlayers == 0 {
		s.SrcdsMaxPlayers = d.SrcdsMaxPlayers
	}
	if s.SrcdsBotTeamSize == 0 {
		s.SrcdsBotTeamSize = d.SrcdsBotTeamSize
	}
	if s.SrcdsStartMission == "" {
		s.SrcdsStartMission = startMissionFor(s.SrcdsStartMap, d.SrcdsStartMission)
	}
	s.SrcdsStartMap = ""
	if s.SrcdsToken == "" {
		s.SrcdsToken = d.SrcdsToken
	}
	if s.MvmMissionCount == 0 {
		s.MvmMissionCount = d.MvmMissionCount
	}
	if s.MvmDifficulty == "" {
		s.MvmDifficulty = d.MvmDifficulty
	}
	if s.MvmGoal == "" {
		s.MvmGoal = d.MvmGoal
	}
	if s.MvmMissionsanityPct == 0 {
		s.MvmMissionsanityPct = d.MvmMissionsanityPct
	}
	if s.MetricsPort == 0 {
		s.MetricsPort = d.MetricsPort
	}
	return s
}

func defaultInstallRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "tf2-archipelago"
	}
	return filepath.Join(home, "tf2-archipelago")
}

// startMissionFor is the migration from a start map to a start mission: the
// first mission the tables list on that map, which is its normal tier.
func startMissionFor(mapName, fallback string) string {
	for _, mission := range gamedata.Missions {
		if played, ok := gamedata.MapByID(mission.Map); ok && played.Name == mapName {
			return mission.PopFile
		}
	}
	return fallback
}

// Reach is how players get to the game server: the local network only,
// through Steam's relay, or straight to a forwarded port. Two flags in the
// file, one answer for the code that reads them.
type Reach string

const (
	ReachLan   Reach = "lan"
	ReachSteam Reach = "steam"
	ReachPort  Reach = "port"
)

// Reaches lists the choices, private first.
func Reaches() []Reach { return []Reach{ReachLan, ReachSteam, ReachPort} }

func (s Settings) Reach() Reach {
	switch {
	case s.SrcdsLan:
		return ReachLan
	case s.SrcdsSteamNetworking:
		return ReachSteam
	default:
		return ReachPort
	}
}

// WithReach returns the settings with the two flags a reach stands for.
func (s Settings) WithReach(reach Reach) Settings {
	s.SrcdsLan = reach == ReachLan
	s.SrcdsSteamNetworking = reach == ReachSteam
	return s
}

// HasToken reports whether a Game Server Login Token is set. "0" is the
// spelling for none, from the compose stack.
func (s Settings) HasToken() bool {
	return s.SrcdsToken != "" && s.SrcdsToken != "0"
}

// Label names the choice in one line, for a menu.
func (r Reach) Label() string {
	switch r {
	case ReachSteam:
		return "Over Steam, no port to open"
	case ReachPort:
		return "Over a port forwarded on the router"
	default:
		return "This machine and the local network only"
	}
}

// Help is the sentence under the label: what the choice costs and what it gives.
func (r Reach) Help() string {
	switch r {
	case ReachSteam:
		return "Steam relays the traffic. Valve hands the server an address like " +
			"169.254.13.42:20232 and your friends connect to that from anywhere. Nothing to " +
			"forward, and your own address stays hidden. Needs a login token, and the address " +
			"is a new one every time the server starts. The launcher shows it once the server is up."
	case ReachPort:
		return "Your friends connect to your public address, so the game port has to be forwarded " +
			"to this machine on your router and open in its firewall. Needs a login token."
	default:
		return "Nobody outside your network can join, whatever the router does. No Steam login, " +
			"no token, never on the public list."
	}
}

// ParseReach reads a saved or typed value. Anything it does not know is LAN:
// a typo must not open a server to the internet.
func ParseReach(value string) Reach {
	switch Reach(value) {
	case ReachSteam:
		return ReachSteam
	case ReachPort:
		return ReachPort
	default:
		return ReachLan
	}
}
