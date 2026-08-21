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
	SrcdsReach         Reach  `json:"srcds_reach"`
	SrcdsAdminSteamIDs string `json:"srcds_admin_steamids,omitempty"`

	// SrcdsStartMission is the popfile the server loads first. The map comes
	// with it: gamedata knows which map a mission runs on.
	SrcdsStartMission string `json:"srcds_start_mission"`

	// SrcdsStartMap is what older files hold. Read for the migration to
	// SrcdsStartMission and never written again.
	SrcdsStartMap string `json:"srcds_start_map,omitempty"`

	// SrcdsLanLegacy is the srcds_lan boolean SrcdsReach replaced. It is read
	// once, to pick the reach a config file written before 1.3 meant, and
	// never written back. Drop it when no such file is left.
	SrcdsLanLegacy *bool `json:"srcds_lan,omitempty"`

	// Defender bots. RED is filled to SrcdsBotTeamSize when a wave begins.
	SrcdsBots        bool `json:"srcds_bots"`
	SrcdsBotTeamSize int  `json:"srcds_bot_team_size"`

	// SrcdsBotClassBlacklist names the classes the bots never play, in the
	// mod's spelling (heavyweapons, not heavy). SrcdsBotLoadouts picks a
	// loadout preset per class, keyed the same way; a class not in it plays
	// stock.
	SrcdsBotClassBlacklist []string          `json:"srcds_bot_class_blacklist,omitempty"`
	SrcdsBotLoadouts       map[string]string `json:"srcds_bot_loadouts,omitempty"`

	// SrcdsBotTeamComp names the classes the bots fill RED with, in order,
	// keyed the same way. Empty leaves the mod to draw its own team, which is
	// what gave a play-test three Spies and two Scouts on an Advanced mission.
	// A team named here beats the blacklist.
	SrcdsBotTeamComp []string `json:"srcds_bot_team_comp,omitempty"`

	// BotUpgradesChat writes what the bots buy at the upgrade station to the
	// chat. Off by default: it is a line per purchase.
	BotUpgradesChat bool `json:"bot_upgrades_chat"`

	// What the bots look like, which changes nothing about how they play. A
	// hat each is on, so is the unusual effect on it, and so are the war paints
	// on the weapons: six mercenaries in the same stock hat is what a bot team
	// looks like otherwise, and this is the part of the mod nobody has to think
	// about.
	//
	// A config file written before these fields existed does not mention them,
	// and a bool nobody wrote reads back as false, so the whole lot arrived
	// switched off on every install that had ever saved a setting. Load tells
	// "the file said no" from "the file said nothing", which is what
	// SrcdsLanLegacy does a few fields down for the same reason.
	SrcdsBotHats        bool `json:"srcds_bot_hats"`
	SrcdsBotHatEffects  bool `json:"srcds_bot_hat_effects"`
	SrcdsBotWeaponSkins bool `json:"srcds_bot_weapon_skins"`

	// Run shape, for seed generation guidance (the launcher does not generate
	// seeds itself, but it can write a starter YAML for the Archipelago app).
	MvmMissionCount     int      `json:"mvm_mission_count"`
	MvmDifficulty       string   `json:"mvm_difficulty"`
	MvmGoal             string   `json:"mvm_goal"`
	MvmMissionsanityPct int      `json:"mvm_missionsanity_percentage"`
	MvmDeathLink        bool     `json:"mvm_death_link"`
	MvmExcludedMissions []string `json:"mvm_excluded_missions,omitempty"`

	// MvmStartMission is the popfile the run starts on and MvmStartClass the
	// mercenary it starts with, both empty for the seed's own random draw.
	// Popfile and not display name, the same as MvmExcludedMissions: every
	// other part of this project names a mission by its popfile, and yaml.go
	// does the one translation the apworld's options need.
	MvmStartMission string `json:"mvm_start_mission,omitempty"`
	MvmStartClass   string `json:"mvm_start_class,omitempty"`

	// Whether to enable the metrics listener and on what port.
	MetricsPort int `json:"metrics_port"`
}

// Defaults returns the factory settings, matching deploy/.env.example.
func Defaults() Settings {
	return Settings{
		InstallRoot:       defaultInstallRoot(),
		APHost:            "archipelago.gg",
		APPort:            0,
		APTls:             true,
		APSlotName:        "tf2",
		SrcdsHostname:     "Mann vs Archipelago",
		SrcdsPort:         27015,
		SrcdsMaxPlayers:   32,
		SrcdsStartMission: "mvm_decoy",
		SrcdsToken:        "0",
		// Lan, to match SrcdsToken above. Port reach with no token is the one
		// combination that cannot work: the server never logs in to Steam, so
		// it answers the query and then refuses the join.
		SrcdsReach:          ReachLan,
		SrcdsBots:           true,
		SrcdsBotTeamSize:    6,
		SrcdsBotHats:        true,
		SrcdsBotHatEffects:  true,
		SrcdsBotWeaponSkins: true,
		MvmMissionCount:     8,
		MvmDifficulty:       "intermediate",
		MvmGoal:             "final_boss",
		MvmMissionsanityPct: 80,
		MetricsPort:         24681,
	}
}

// EffectiveReach is where the players can come from with the token this server
// has, which is not always the reach that was asked for. See Effective.
func (s Settings) EffectiveReach() Reach { return Effective(s.SrcdsReach, s.SrcdsToken) }

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
	s, err := parse(data)
	if err != nil {
		return Settings{}, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	return s, nil
}

// parse is the config file's bytes as settings, defaults filled in. Separate
// from Load because the order matters and there is only one right one: the
// file, then the defaults for what it left out, then the three switches that
// have to tell "said no" from "said nothing".
func parse(data []byte) (Settings, error) {
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, err
	}
	return s.withDefaults().withAppearanceDefaults(data), nil
}

// withAppearanceDefaults fills the three appearance switches a file that
// predates them does not mention. Called by parse, after the file's own values
// are in.
//
// It reads the same bytes a second time rather than making the fields
// pointers: a *bool that every caller has to dereference is a worse trade than
// one function that knows why these three are different from the rest.
func (s Settings) withAppearanceDefaults(data []byte) Settings {
	var said struct {
		Hats    *bool `json:"srcds_bot_hats"`
		Effects *bool `json:"srcds_bot_hat_effects"`
		Skins   *bool `json:"srcds_bot_weapon_skins"`
	}
	// A file that parsed once parses again; anything else has already been
	// reported by the caller.
	if err := json.Unmarshal(data, &said); err != nil {
		return s
	}
	d := Defaults()
	if said.Hats == nil {
		s.SrcdsBotHats = d.SrcdsBotHats
	}
	if said.Effects == nil {
		s.SrcdsBotHatEffects = d.SrcdsBotHatEffects
	}
	if said.Skins == nil {
		s.SrcdsBotWeaponSkins = d.SrcdsBotWeaponSkins
	}
	return s
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
	if !s.SrcdsReach.Valid() {
		// A file that predates SrcdsReach says only whether sv_lan was on, and
		// that answer still stands: on meant the local network. A file that
		// says neither, or says something nobody recognizes, takes the default,
		// which EffectiveReach keeps local until there is a token to leave with.
		s.SrcdsReach = d.SrcdsReach
		if s.SrcdsLanLegacy != nil {
			if *s.SrcdsLanLegacy {
				s.SrcdsReach = ReachLan
			} else {
				s.SrcdsReach = ReachPort
			}
		}
	}
	s.SrcdsLanLegacy = nil
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
