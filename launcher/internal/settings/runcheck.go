package settings

import (
	"fmt"
	"slices"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
)

// CheckRunSelection validates the mission pool before the launcher writes or
// generates a player file. The official generator repeats this check and
// remains the final authority.
func CheckRunSelection(s Settings) (runshape.Preflight, error) {
	// "random" is the apworld's own word for "draw one", the shipped default,
	// and not a popfile. The check has to see it as no start named, or every
	// stack that keeps the default is refused a seed.
	start := s.MvmStartMission
	if start == randomOption {
		start = ""
	}
	report, err := runshape.CheckSelection(runshape.Selection{
		Pool:         MissionPool(s),
		Difficulty:   s.MvmDifficulty,
		MissionCount: s.MvmMissionCount,
		StartMission: start,
		Giantsanity:  s.MvmGiantsanity,
		Tanksanity:   s.MvmTanksanity,
	})
	if err != nil {
		return report, fmt.Errorf("archipelago run selection: %w", err)
	}
	if s.MvmMissionModifiers &&
		(s.MvmModifierMin < 0 || s.MvmModifierMax > 3 || s.MvmModifierMin > s.MvmModifierMax) {
		return report, fmt.Errorf(
			"archipelago run selection: mission modifier bounds %d..%d must be ordered and within 0..3",
			s.MvmModifierMin, s.MvmModifierMax,
		)
	}
	return report, nil
}

// RequiredServerMods is the mod set needed by community missions currently in
// the pool (or explicitly named as the start). It deliberately ignores a mod
// checkbox on its own: selecting SigMod before downloading it is a valid draft
// as long as no SigMod mission has been selected yet.
func RequiredServerMods(s Settings) []string {
	needed := make(map[string]bool)
	if s.MvmCommunityMissions {
		for _, mission := range gamedata.Missions {
			if slices.Contains(s.MvmExcludedMissions, mission.PopFile) {
				continue
			}
			if key := gamedata.MissionServerMod(mission.ID); key != "" {
				needed[key] = true
			}
		}
	}
	if mission, known := gamedata.MissionByPopFile(s.MvmStartMission); known {
		if key := gamedata.MissionServerMod(mission.ID); key != "" {
			needed[key] = true
		}
	}
	var keys []string
	for _, key := range gamedata.ServerModKeys() {
		if needed[key] {
			keys = append(keys, key)
		}
	}
	return keys
}

// CheckServerModsReady prevents a seed from naming modded missions merely
// because a settings file claims the mod is enabled. The launcher supplies a
// set found by inspecting the managed installation on disk.
func CheckServerModsReady(s Settings, ready []string) error {
	for _, key := range RequiredServerMods(s) {
		if slices.Contains(ServerModKeys(s), key) && slices.Contains(ready, key) {
			continue
		}
		mod, _ := gamedata.ServerModByKey(key)
		if !slices.Contains(ServerModKeys(s), key) {
			return fmt.Errorf("%s mission selected: turn on %s on the Missions page, then press Download / set up selected server mods", mod.Name, mod.Name)
		}
		return fmt.Errorf("%s mission selected, but its managed installation is missing or incomplete: press Download / set up selected server mods", mod.Name)
	}
	return nil
}

// MissionPool is the part of the settings that decides which missions the
// generator may draw at all. One place builds it, so the ceiling the launcher
// offers, the tier labels and the preflight all count the pool the apworld
// counts rather than three different ones.
func MissionPool(s Settings) runshape.Pool {
	return runshape.Pool{
		Mods:      ServerModKeys(s),
		Community: s.MvmCommunityMissions,
		Excluded:  s.MvmExcludedMissions,
	}
}
