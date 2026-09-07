package settings

import (
	"fmt"

	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
)

// CheckRunSelection validates the mission pool before the launcher writes or
// generates a player file. The official generator repeats this check and
// remains the final authority.
func CheckRunSelection(s Settings) (runshape.Preflight, error) {
	report, err := runshape.CheckSelection(runshape.Selection{
		Pool:         MissionPool(s),
		Difficulty:   s.MvmDifficulty,
		MissionCount: s.MvmMissionCount,
		StartMission: s.MvmStartMission,
	})
	if err != nil {
		return report, fmt.Errorf("archipelago run selection: %w", err)
	}
	return report, nil
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
