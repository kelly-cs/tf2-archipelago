// Package composeenv persists launcher settings in the .env file Docker
// Compose reads. It preserves comments, ordering and variables the launcher
// does not own; only values changed in the browser are rewritten.
package composeenv

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// Write applies the settings that changed between before and after to path.
// The file is truncated in place rather than replaced: Compose bind-mounts the
// inode, so an atomic rename would update only the container's mount and leave
// the operator's host .env unchanged.
func Write(path string, before, after settings.Settings) error {
	oldValues, newValues := values(before), values(after)
	changes := make(map[string]string)
	for name, value := range newValues {
		if oldValues[name] != value {
			changes[name] = value
		}
	}
	if len(changes) == 0 {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read Compose settings %s: %w", path, err)
	}
	lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	written := make(map[string]bool)
	for i, line := range lines {
		name, _, found := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		value, changed := changes[name]
		if !found || !changed || strings.HasPrefix(name, "#") {
			continue
		}
		lines[i] = name + "=" + quote(value)
		written[name] = true
	}
	names := make([]string, 0, len(changes))
	for name := range changes {
		if !written[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > 0 && len(lines) > 0 && lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	for _, name := range names {
		lines = append(lines, name+"="+quote(changes[name]))
	}
	rendered := []byte(strings.Join(lines, "\n") + "\n")
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot write Compose settings %s: %w", path, err)
	}
	_, writeErr := file.WriteAt(rendered, 0)
	if writeErr == nil {
		writeErr = file.Truncate(int64(len(rendered)))
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("cannot write Compose settings %s: %w", path, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("cannot close Compose settings %s: %w", path, closeErr)
	}
	return nil
}

// Single quotes keep Compose from interpolating passwords containing dollar
// signs. Compose accepts backslash-escaped quotes inside a single-quoted value.
func quote(value string) string {
	value = strings.ReplaceAll(value, "'", `\'`)
	value = strings.ReplaceAll(value, "\r", `\r`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return "'" + value + "'"
}

func values(s settings.Settings) map[string]string {
	return map[string]string{
		"TF2AP_INSTALL_ROOT":            s.InstallRoot,
		"TF2AP_COMMUNITY_CONTENT_DIR":   s.CommunityContentDir,
		"TF2AP_COMMUNITY_PACKS":         strings.Join(s.CommunityPacks, ","),
		"TF2AP_TEST_MODE":               boolean(s.TestMode),
		"TF2AP_ARCHIPELAGO_DIR":         s.ArchipelagoDir,
		"AP_HOST":                       s.APHost,
		"AP_PORT":                       strconv.Itoa(s.APPort),
		"AP_TLS":                        boolean(s.APTls),
		"AP_SLOT_NAME":                  s.APSlotName,
		"AP_PASSWORD":                   s.APPassword,
		"SRCDS_HOSTNAME":                s.SrcdsHostname,
		"SRCDS_RCONPW":                  s.SrcdsRconPw,
		"SRCDS_PW":                      s.SrcdsPw,
		"SRCDS_PORT":                    strconv.Itoa(s.SrcdsPort),
		"TF2AP_JOIN_HOST":               s.SrcdsJoinHost,
		"SRCDS_MAXPLAYERS":              strconv.Itoa(s.SrcdsMaxPlayers),
		"SRCDS_START_MISSION":           s.SrcdsStartMission,
		"SRCDS_TOKEN":                   s.SrcdsToken,
		"SRCDS_REACH":                   string(s.SrcdsReach),
		"SRCDS_ADMIN_STEAMIDS":          s.SrcdsAdminSteamIDs,
		"SRCDS_MODS":                    strings.Join(s.SrcdsMods, ","),
		"FASTDL_PORT":                   strconv.Itoa(s.FastDLPort),
		"SRCDS_DOWNLOADURL":             s.SrcdsDownloadURL,
		"TAILSCALE_FASTDL":              boolean(s.TailscaleFastDL),
		"SRCDS_BOTS":                    boolean(s.SrcdsBots),
		"SRCDS_BOT_TEAM_SIZE":           strconv.Itoa(s.SrcdsBotTeamSize),
		"SRCDS_BOT_CLASS_BLACKLIST":     strings.Join(s.SrcdsBotClassBlacklist, ","),
		"SRCDS_BOT_LOADOUTS":            pairs(s.SrcdsBotLoadouts),
		"SRCDS_BLU_HEALTH_PCT":          strconv.Itoa(s.SrcdsBluHealthPct),
		"SRCDS_BOT_TEAM_COMP":           strings.Join(s.SrcdsBotTeamComp, ","),
		"SRCDS_BOT_SEAT_LOADOUTS":       strings.Join(s.SrcdsBotSeatLoadouts, ","),
		"SRCDS_BOT_HATS":                boolean(s.SrcdsBotHats),
		"SRCDS_BOT_HAT_EFFECTS":         boolean(s.SrcdsBotHatEffects),
		"TF2AP_BOT_UPGRADES_CHAT":       boolean(s.BotUpgradesChat),
		"MVM_MISSION_COUNT":             strconv.Itoa(s.MvmMissionCount),
		"MVM_DIFFICULTY":                s.MvmDifficulty,
		"MVM_GOAL":                      s.MvmGoal,
		"MVM_MISSIONSANITY_PERCENTAGE":  strconv.Itoa(s.MvmMissionsanityPct),
		"MVM_MISSION_MODIFIERS":         boolean(s.MvmMissionModifiers),
		"MVM_MINIMUM_MISSION_MODIFIERS": strconv.Itoa(s.MvmModifierMin),
		"MVM_MAXIMUM_MISSION_MODIFIERS": strconv.Itoa(s.MvmModifierMax),
		"MVM_MEDAL_ON_CLEAR":            boolean(s.MvmMedalOnClear),
		"MVM_VICTORY_CACHES":            boolean(s.MvmVictoryCaches),
		"MVM_MILESTONE_CHECKS":          boolean(s.MvmMilestoneChecks),
		"MVM_CLASS_WEAPON_SLOTS":        boolean(s.MvmClassWeaponSlots),
		"MVM_SERVER_SETTINGS":           boolean(s.MvmServerSettings),
		"MVM_DEATH_LINK":                boolean(s.MvmDeathLink),
		"MVM_EXCLUDED_MISSIONS":         strings.Join(s.MvmExcludedMissions, ","),
		"MVM_START_MISSION":             s.MvmStartMission,
		"MVM_START_CLASS":               s.MvmStartClass,
		"MVM_COMMUNITY_MISSIONS":        boolean(s.MvmCommunityMissions),
		"MVM_MISSION_TICKET_IMPORTANCE": s.MvmMissionTicketImportance,
		"MVM_CLASS_UNLOCK_IMPORTANCE":   s.MvmClassUnlockImportance,
		"MVM_WEAPON_SLOT_IMPORTANCE":    s.MvmWeaponSlotImportance,
		"MVM_WEAPON_BUFF_IMPORTANCE":    s.MvmWeaponBuffImportance,
		"MVM_CASH_REWARDS":              boolean(s.MvmCashRewards),
		"MVM_WEAPON_BUFF_PERCENTAGE":    strconv.Itoa(s.MvmWeaponBuffPct),
		"MVM_WEAPON_BUFF_STACK_CHANCE":  strconv.Itoa(s.MvmWeaponBuffStackChance),
		"MVM_TRAP_PERCENTAGE":           strconv.Itoa(s.MvmTrapPct),
		"BRIDGE_METRICS_PORT":           strconv.Itoa(s.MetricsPort),
	}
}

func boolean(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func pairs(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+values[key])
	}
	return strings.Join(out, ",")
}
