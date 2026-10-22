package apclient

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata"
)

/*
	uninstalledPacks names the community missions the seed draws that this
	server has no population file for, and the packs that carry them.

The plugin refuses such a mission when the run switches to it, which can be an
hour in. The seed is known as soon as the room accepts the slot, so that is when
the player hears it, with the pack to tick. Valve's missions are not looked
for: their files live in the game's own archives, not in this directory.
*/
func uninstalledPacks(populationDir string, missions []string) error {
	if populationDir == "" {
		return nil
	}
	var names, packs []string
	for _, popFile := range missions {
		mission, known := gamedata.MissionByPopFile(popFile)
		if !known || !gamedata.IsCommunityMission(mission.ID) {
			continue
		}
		_, err := os.Stat(filepath.Join(populationDir, popFile+".pop"))
		if err == nil {
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("cannot look for %s.pop: %w", popFile, err)
		}
		names = append(names, mission.Name)
		pack := gamedata.MissionPack(mission.ID)
		if label := gamedata.PackLabel(pack); label != "" {
			pack = label
		}
		if !slices.Contains(packs, pack) {
			packs = append(packs, pack)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return fmt.Errorf(
		"this seed draws %s, and the community pack it comes from is not installed on this server. "+
			"Tick %s in Settings > Missions and press Start again, or play a seed generated without community missions",
		strings.Join(names, ", "), strings.Join(packs, " and "))
}
