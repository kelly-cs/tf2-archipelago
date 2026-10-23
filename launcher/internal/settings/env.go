package settings

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

/*
EnvVar is one variable ApplyEnv reads, and the Compose services that read it
from their environment.

Services is what deploy/compose.yml and deploy/compose.seed.yml must hand that
variable to, and TestComposePassesEveryVariableToItsReaders holds them to it.
The admin service is never named: it loads the whole of .env, because it is
what writes it. A variable no service names is one only the launcher and the
admin page act on.

The archipelago and seed services both run playeryaml, so both read whatever
the player file renders.
*/
type EnvVar struct {
	Name     string
	Services []string
}

var (
	playerFile       = []string{"archipelago", "seed"}
	playerFileBridge = []string{"archipelago", "seed", "bridge"}
	gameServer       = []string{"srcds"}
)

// EnvVars lists every variable ApplyEnv reads, in the order the help prints
// them. The names are the ones deploy/.env.example already uses, so a compose
// operator's file works here unchanged.
var EnvVars = []EnvVar{
	{"TF2AP_INSTALL_ROOT", nil},
	{"TF2AP_COMMUNITY_CONTENT_DIR", nil},
	{"TF2AP_COMMUNITY_PACKS", playerFile},
	{"TF2AP_TEST_MODE", []string{"srcds", "bridge"}},
	{"TF2AP_ARCHIPELAGO_DIR", nil},
	{"AP_ROOM", nil},
	{"AP_HOST", []string{"bridge"}},
	{"AP_PORT", []string{"archipelago", "bridge"}},
	{"AP_TLS", []string{"bridge"}},
	{"AP_ROOM_URL", nil},
	{"AP_SLOT_NAME", playerFileBridge},
	{"AP_PASSWORD", []string{"archipelago", "bridge"}},
	{"SRCDS_HOSTNAME", gameServer},
	{"SRCDS_RCONPW", gameServer},
	{"SRCDS_PW", gameServer},
	{"SRCDS_PORT", gameServer},
	{"TF2AP_JOIN_HOST", nil},
	{"SRCDS_MAXPLAYERS", gameServer},
	{"SRCDS_START_MISSION", gameServer},
	{"SRCDS_STARTMAP", gameServer},
	{"SRCDS_TOKEN", gameServer},
	{"SRCDS_LAN", gameServer},
	{"SRCDS_REACH", gameServer},
	{"SRCDS_ADMIN_STEAMIDS", gameServer},
	{"SRCDS_MODS", []string{"archipelago", "seed", "bridge", "srcds"}},
	{"SRCDS_MOD_LOADING", nil},
	{"FASTDL_PORT", gameServer},
	{"SRCDS_DOWNLOADURL", gameServer},
	{"TAILSCALE_FASTDL", []string{"srcds", "tailscale-fastdl"}},
	{"SRCDS_BOTS", gameServer},
	{"SRCDS_BOT_TEAM_SIZE", gameServer},
	{"SRCDS_BOT_CLASS_BLACKLIST", gameServer},
	{"SRCDS_BOT_LOADOUTS", gameServer},
	{"SRCDS_BOT_CUSTOM_LOADOUTS", gameServer},
	{"SRCDS_BOT_TEAM_PRESETS", nil},
	{"SRCDS_BOT_NAMES_EXCLUDED", gameServer},
	{"SRCDS_BOT_NAMES_ADDED", gameServer},
	{"SRCDS_BOT_SEAT_NAMES", gameServer},
	{"SRCDS_BLU_HEALTH_PCT", nil},
	{"SRCDS_BOT_TEAM_COMP", gameServer},
	{"SRCDS_BOT_SEAT_LOADOUTS", gameServer},
	{"SRCDS_BOT_CARD_FORMS", gameServer},
	{"SRCDS_BOT_HATS", gameServer},
	{"SRCDS_BOT_BUY_ANYWHERE", gameServer},
	{"SRCDS_BOT_HAT_EFFECTS", gameServer},
	{"TF2AP_BOT_UPGRADES_CHAT", gameServer},
	{"MVM_MISSION_COUNT", playerFileBridge},
	{"MVM_DIFFICULTY", playerFileBridge},
	{"MVM_GOAL", playerFileBridge},
	{"MVM_MISSIONSANITY_PERCENTAGE", playerFile},
	{"MVM_MISSION_MODIFIERS", playerFileBridge},
	{"MVM_MINIMUM_MISSION_MODIFIERS", playerFileBridge},
	{"MVM_MAXIMUM_MISSION_MODIFIERS", playerFileBridge},
	{"MVM_MEDAL_ON_CLEAR", playerFile},
	{"MVM_VICTORY_CACHES", playerFileBridge},
	{"MVM_MILESTONE_CHECKS", playerFileBridge},
	{"MVM_CLASS_WEAPON_SLOTS", playerFile},
	{"MVM_GIANTSANITY", playerFileBridge},
	{"MVM_TANKSANITY", playerFileBridge},
	{"MVM_SERVER_SETTINGS", playerFile},
	{"MVM_DEATH_LINK", playerFileBridge},
	{"MVM_EXCLUDED_MISSIONS", playerFileBridge},
	{"MVM_START_MISSION", playerFileBridge},
	{"MVM_START_CLASS", playerFileBridge},
	{"MVM_COMMUNITY_MISSIONS", playerFileBridge},
	{"MVM_MISSION_TICKET_IMPORTANCE", playerFile},
	{"MVM_CLASS_UNLOCK_IMPORTANCE", playerFile},
	{"MVM_WEAPON_SLOT_IMPORTANCE", playerFile},
	{"MVM_WEAPON_BUFF_IMPORTANCE", playerFile},
	{"MVM_CASH_REWARDS", playerFile},
	{"MVM_BOT_CARDS", playerFile},
	{"MVM_STARTING_BOT_CARDS", playerFile},
	{"MVM_STARTING_BOT_CARD_MODE", playerFile},
	{"SRCDS_BOT_CARD_ROLLS", nil},
	{"MVM_WEAPON_BUFF_PERCENTAGE", playerFile},
	{"MVM_WEAPON_BUFF_STACK_CHANCE", playerFile},
	{"MVM_TRAP_PERCENTAGE", playerFile},
	{"BRIDGE_METRICS_PORT", nil},
}

// EnvNames is the names of EnvVars, in the same order.
var EnvNames = envNames()

func envNames() []string {
	names := make([]string, len(EnvVars))
	for i, v := range EnvVars {
		names[i] = v.Name
	}
	return names
}

// ApplyEnv overlays the environment on top of the saved settings. A variable
// that is set wins over the file; one that is unset changes nothing.
//
// This is what makes the exe scriptable: a shortcut, a .bat or a CI job can
// set SRCDS_RCONPW and AP_PORT and start a server with no prompts and no
// config file. Values taken from the environment are not written back, so an
// override for one run does not become the saved answer.
// applyBotEnv is the bots' half: who fills RED, what they carry, and how far
// the robots are taken down for a short team. Split out because ApplyEnv reads
// as one list and a list has a length limit.
func applyBotEnv(s Settings) Settings {
	boolean(&s.SrcdsBots, "SRCDS_BOTS")
	num(&s.SrcdsBotTeamSize, "SRCDS_BOT_TEAM_SIZE")
	num(&s.SrcdsBluHealthPct, "SRCDS_BLU_HEALTH_PCT")
	list(&s.SrcdsBotClassBlacklist, "SRCDS_BOT_CLASS_BLACKLIST")
	pairs(&s.SrcdsBotLoadouts, "SRCDS_BOT_LOADOUTS")
	fromJSON(&s.SrcdsBotCustomLoadouts, "SRCDS_BOT_CUSTOM_LOADOUTS")
	fromJSON(&s.SrcdsBotTeamPresets, "SRCDS_BOT_TEAM_PRESETS")
	list(&s.SrcdsBotNamesExcluded, "SRCDS_BOT_NAMES_EXCLUDED")
	list(&s.SrcdsBotNamesAdded, "SRCDS_BOT_NAMES_ADDED")
	seatList(&s.SrcdsBotSeatNames, "SRCDS_BOT_SEAT_NAMES")
	seatList(&s.SrcdsBotTeamComp, "SRCDS_BOT_TEAM_COMP")
	seatList(&s.SrcdsBotSeatLoadouts, "SRCDS_BOT_SEAT_LOADOUTS")
	pairs(&s.SrcdsBotCardForms, "SRCDS_BOT_CARD_FORMS")
	fromJSON(&s.SrcdsBotCardRolls, "SRCDS_BOT_CARD_ROLLS")
	boolean(&s.SrcdsBotHats, "SRCDS_BOT_HATS")
	boolean(&s.SrcdsBotBuyAnywhere, "SRCDS_BOT_BUY_ANYWHERE")
	boolean(&s.SrcdsBotHatEffects, "SRCDS_BOT_HAT_EFFECTS")
	boolean(&s.BotUpgradesChat, "TF2AP_BOT_UPGRADES_CHAT")
	return s
}

func applyServerEnv(s Settings) Settings {
	str(&s.SrcdsHostname, "SRCDS_HOSTNAME")
	str(&s.SrcdsRconPw, "SRCDS_RCONPW")
	str(&s.SrcdsPw, "SRCDS_PW")
	num(&s.SrcdsPort, "SRCDS_PORT")
	str(&s.SrcdsJoinHost, "TF2AP_JOIN_HOST")
	num(&s.SrcdsMaxPlayers, "SRCDS_MAXPLAYERS")
	// The compose stack names a map; the launcher names a mission. Both work,
	// the mission wins.
	if value, ok := os.LookupEnv("SRCDS_STARTMAP"); ok {
		s.SrcdsStartMission = startMissionFor(value, s.SrcdsStartMission)
	}
	str(&s.SrcdsStartMission, "SRCDS_START_MISSION")
	str(&s.SrcdsToken, "SRCDS_TOKEN")

	// SRCDS_LAN is the older spelling and covers two of the three reaches, so
	// SRCDS_REACH is read after it and wins when both are set.
	if lan, ok := lookupBool("SRCDS_LAN"); ok {
		s.SrcdsReach = ReachLan
		if !lan {
			s.SrcdsReach = ReachPort
		}
	}
	if value, ok := os.LookupEnv("SRCDS_REACH"); ok {
		if reach, ok := ParseReach(strings.ToLower(strings.TrimSpace(value))); ok {
			s.SrcdsReach = reach
		}
	}
	str(&s.SrcdsAdminSteamIDs, "SRCDS_ADMIN_STEAMIDS")
	list(&s.SrcdsMods, "SRCDS_MODS")
	if value, ok := os.LookupEnv("SRCDS_MOD_LOADING"); ok {
		if loading := ModLoading(strings.ToLower(strings.TrimSpace(value))); loading.Valid() {
			s.SrcdsModLoading = loading
		}
	}
	num(&s.FastDLPort, "FASTDL_PORT")
	str(&s.SrcdsDownloadURL, "SRCDS_DOWNLOADURL")
	boolean(&s.TailscaleFastDL, "TAILSCALE_FASTDL")
	return applyBotEnv(s)
}

func ApplyEnv(s Settings) Settings {
	str(&s.InstallRoot, "TF2AP_INSTALL_ROOT")
	str(&s.CommunityContentDir, "TF2AP_COMMUNITY_CONTENT_DIR")
	list(&s.CommunityPacks, "TF2AP_COMMUNITY_PACKS")
	boolean(&s.TestMode, "TF2AP_TEST_MODE")
	str(&s.ArchipelagoDir, "TF2AP_ARCHIPELAGO_DIR")

	// AP_ROOM is the whole address in one variable, which is how the room page
	// gives it. The three parts stay readable for a compose .env.
	if value, ok := os.LookupEnv("AP_ROOM"); ok {
		if room, err := ParseRoom(value); err == nil {
			s.APHost, s.APPort, s.APTls = room.Host, room.Port, room.TLS
		}
	}
	str(&s.APHost, "AP_HOST")
	num(&s.APPort, "AP_PORT")
	boolean(&s.APTls, "AP_TLS")
	str(&s.APRoomURL, "AP_ROOM_URL")
	str(&s.APSlotName, "AP_SLOT_NAME")
	str(&s.APPassword, "AP_PASSWORD")

	s = applyServerEnv(s)

	num(&s.MvmMissionCount, "MVM_MISSION_COUNT")
	str(&s.MvmDifficulty, "MVM_DIFFICULTY")
	str(&s.MvmGoal, "MVM_GOAL")
	num(&s.MvmMissionsanityPct, "MVM_MISSIONSANITY_PERCENTAGE")
	applyMissionModifierEnv(&s)
	boolean(&s.MvmMedalOnClear, "MVM_MEDAL_ON_CLEAR")
	boolean(&s.MvmVictoryCaches, "MVM_VICTORY_CACHES")
	boolean(&s.MvmMilestoneChecks, "MVM_MILESTONE_CHECKS")
	classSlots(&s.MvmClassWeaponSlots, "MVM_CLASS_WEAPON_SLOTS")
	boolean(&s.MvmGiantsanity, "MVM_GIANTSANITY")
	boolean(&s.MvmTanksanity, "MVM_TANKSANITY")
	boolean(&s.MvmServerSettings, "MVM_SERVER_SETTINGS")
	boolean(&s.MvmDeathLink, "MVM_DEATH_LINK")
	list(&s.MvmExcludedMissions, "MVM_EXCLUDED_MISSIONS")
	str(&s.MvmStartMission, "MVM_START_MISSION")
	str(&s.MvmStartClass, "MVM_START_CLASS")
	boolean(&s.MvmCommunityMissions, "MVM_COMMUNITY_MISSIONS")
	applyRewardEnv(&s)

	num(&s.MetricsPort, "BRIDGE_METRICS_PORT")
	return s
}

func applyMissionModifierEnv(s *Settings) {
	boolean(&s.MvmMissionModifiers, "MVM_MISSION_MODIFIERS")
	num(&s.MvmModifierMin, "MVM_MINIMUM_MISSION_MODIFIERS")
	num(&s.MvmModifierMax, "MVM_MAXIMUM_MISSION_MODIFIERS")
}

func applyRewardEnv(s *Settings) {
	str(&s.MvmMissionTicketImportance, "MVM_MISSION_TICKET_IMPORTANCE")
	str(&s.MvmClassUnlockImportance, "MVM_CLASS_UNLOCK_IMPORTANCE")
	str(&s.MvmWeaponSlotImportance, "MVM_WEAPON_SLOT_IMPORTANCE")
	str(&s.MvmWeaponBuffImportance, "MVM_WEAPON_BUFF_IMPORTANCE")
	boolean(&s.MvmCashRewards, "MVM_CASH_REWARDS")
	boolean(&s.MvmBotCards, "MVM_BOT_CARDS")
	num(&s.MvmStartingBotCards, "MVM_STARTING_BOT_CARDS")
	str(&s.MvmStartingBotCardMode, "MVM_STARTING_BOT_CARD_MODE")
	num(&s.MvmWeaponBuffPct, "MVM_WEAPON_BUFF_PERCENTAGE")
	num(&s.MvmWeaponBuffStackChance, "MVM_WEAPON_BUFF_STACK_CHANCE")
	num(&s.MvmTrapPct, "MVM_TRAP_PERCENTAGE")
}

// FromEnv reports whether name is set, so the UI can skip a prompt whose
// answer the environment already gave.
func FromEnv(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

func str(target *string, name string) {
	if value, ok := os.LookupEnv(name); ok {
		*target = value
	}
}

func num(target *int, name string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return
	}
	*target = parsed
}

// boolean accepts the spellings an .env file and a shell both produce.
func boolean(target *bool, name string) {
	if value, ok := lookupBool(name); ok {
		*target = value
	}
}

// Truthy reads one of the spellings an .env file and a shell both produce.
// Anything else, including an empty string, is false: a flag nobody set stays
// off, and so does one somebody spelled wrong.
func Truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// lookupBool is boolean for a caller that does not have a *bool to write into.
// The second result is false both when the variable is unset and when it holds
// something that is not a boolean at all.
func lookupBool(name string) (bool, bool) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	}
	return false, false
}

// list reads a comma-separated value. An empty value empties the list, which
// is how a shortcut clears a saved one.
func list(target *[]string, name string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}
	*target = SplitList(value)
}

// seatList reads a list where entry n is seat n of RED. Drop the empties the
// way list does and every seat after a draw moves up one.
func seatList(target *[]string, name string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}
	*target = SplitSeats(value)
}

// pairs reads "key=value,key=value".
/*
fromJSON reads one of the tables somebody built in front of the page, which the
Compose stack carries as JSON because they are records rather than words.

A value that will not parse is left alone rather than emptied. The alternative
is a stack that loses every loadout or team somebody built to one bad edit of
.env, and the file is edited by hand.
*/
func fromJSON[T any](target *map[string]T, name string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}
	if strings.TrimSpace(value) == "" {
		*target = nil
		return
	}
	parsed := make(map[string]T)
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return
	}
	*target = parsed
}

func pairs(target *map[string]string, name string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}
	parsed := make(map[string]string)
	for _, entry := range SplitList(value) {
		key, item, found := strings.Cut(entry, "=")
		if found && key != "" {
			parsed[strings.TrimSpace(key)] = strings.TrimSpace(item)
		}
	}
	*target = parsed
}

// SplitSeats is SplitList that keeps the empties, because a seat nobody named
// is still a seat. It drops the trailing empties, which name no seat.
func SplitSeats(value string) []string {
	var out []string
	last := -1
	for entry := range strings.SplitSeq(value, ",") {
		out = append(out, strings.TrimSpace(entry))
		if out[len(out)-1] != "" {
			last = len(out) - 1
		}
	}
	return out[:last+1]
}

// SplitList splits on commas and drops blanks, the way every list in an .env
// file is written.
func SplitList(value string) []string {
	var out []string
	for entry := range strings.SplitSeq(value, ",") {
		if entry = strings.TrimSpace(entry); entry != "" {
			out = append(out, entry)
		}
	}
	return out
}

// classSlots reads the three-way one, which was a boolean before 1.16.1 and
// still reads as one: a compose file that says 1 or true gets the progressive
// slots it was asking for.
func classSlots(target *ClassSlots, name string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		*target = ClassSlotsProgressive
	case "0", "false", "no", "off", "":
		*target = ClassSlotsOff
	default:
		*target = ClassSlots(strings.ToLower(strings.TrimSpace(value))).OrOff()
	}
}
