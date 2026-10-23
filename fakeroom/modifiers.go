package fakeroom

import (
	"math/rand/v2"

	"github.com/m-this/tf2-archipelago/gamedata"
)

type MissionModifier = gamedata.MissionModifier

// DrawMissionModifiers gives every mission a stable assignment for this fake
// room's lifetime. A real seed uses its own RNG; test mode only needs the same
// bounds and compatibility rules so it exercises the same game path.
func DrawMissionModifiers(missions []string, minimum, maximum int) map[string][]MissionModifier {
	minimum = max(0, min(minimum, 3))
	maximum = max(minimum, min(maximum, 3))
	assignments := make(map[string][]MissionModifier, len(missions))
	for _, mission := range missions {
		wanted := minimum
		if maximum > minimum {
			wanted += rand.IntN(maximum - minimum + 1)
		}
		if wanted == 0 {
			assignments[mission] = []MissionModifier{}
			continue
		}
		candidates := append([]MissionModifier(nil), gamedata.MissionModifiers...)
		rand.Shuffle(len(candidates), func(left, right int) {
			candidates[left], candidates[right] = candidates[right], candidates[left]
		})
		groups := make(map[string]bool)
		for _, modifier := range candidates {
			if modifier.ExclusiveGroup != "" && groups[modifier.ExclusiveGroup] {
				continue
			}
			groups[modifier.ExclusiveGroup] = modifier.ExclusiveGroup != ""
			assignments[mission] = append(assignments[mission], modifier)
			if len(assignments[mission]) == wanted {
				break
			}
		}
	}
	return assignments
}
