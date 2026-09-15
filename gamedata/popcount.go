package gamedata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
What a wave holds, counted out of its population file.

Every giant and every tank in a mission is a check when the sanity options are
on (gh-68), so the seed has to know how many each wave spawns before any
server runs. That count is read here, out of the stock population syntax:
Wave blocks in order, each WaveSpawn's TotalCount, and what its spawner is. A
Support spawn that runs until the wave ends is not counted, because it has no
count to give; Support Limited is.

A bot is a giant when it carries Attributes MiniBoss, on itself, on the
template it names, or on that template's own template, and the attribute may
sit inside an EventChangeAttributes block: that is where the gatebots keep it.
A Squad spawns its members in turn up to TotalCount, so a squad of five with
two giants and a count of ten holds four. A RandomChoice whose members
disagree is counted as none and reported, and Valve's missions have no such
wave; a mission that does gets no per-giant checks for it rather than a
guess.

Checked against Cowser's hand count of every Valve wave: 180 waves agree, and
the one that does not is a squad of regular gatebot heavies the sheet took
for giants.
*/

// WaveKills is what one wave spawns that a player can be paid for killing.
type WaveKills struct {
	Giants uint8
	Tanks  uint8
	// Uncertain names the WaveSpawns whose giant count depends on a random
	// draw, so the wave's giants are not to be trusted as a check count.
	Uncertain []string
}

// popNode is one key with either a value or a block. Keys repeat, so a block
// is a list rather than a map.
type popNode struct {
	key   string
	value string
	block []popNode
}

var popBaseLine = regexp.MustCompile(`(?m)^#base\s+(\S+)`)

// CountWaveKills reads a population file. resolve returns the body of a #base
// file by name, or nil when it cannot.
func CountWaveKills(body []byte, resolve func(name string) []byte) ([]WaveKills, error) {
	templates := map[string][]popNode{}
	root, err := popTree(body, resolve, templates, map[string]bool{})
	if err != nil {
		return nil, err
	}
	var waves []WaveKills
	for _, wave := range popChildren(root, "Wave") {
		var kills WaveKills
		for _, spawn := range popChildren(wave.block, "WaveSpawn") {
			if support := popValue(spawn.block, "Support"); support != "" && !strings.EqualFold(support, "limited") {
				continue
			}
			total := 1
			if text := popValue(spawn.block, "TotalCount"); text != "" {
				parsed, err := strconv.Atoi(text)
				if err != nil {
					return nil, fmt.Errorf("TotalCount %q is not a number", text)
				}
				total = parsed
			}
			for _, spawner := range spawn.block {
				if !popIsSpawner(spawner.key) {
					continue
				}
				giants, tanks, uncertain := popCountSpawner(spawner, templates, total, 0)
				kills.Giants += uint8(giants)
				kills.Tanks += uint8(tanks)
				if uncertain {
					kills.Uncertain = append(kills.Uncertain, popValue(spawn.block, "Name"))
				}
			}
		}
		waves = append(waves, kills)
	}
	return waves, nil
}

// popTree parses a file into the WaveSchedule block, collecting the templates
// it and its #base files declare.
func popTree(body []byte, resolve func(string) []byte, templates map[string][]popNode, seen map[string]bool) ([]popNode, error) {
	for _, match := range popBaseLine.FindAllSubmatch(body, -1) {
		name := string(match[1])
		if seen[name] {
			continue
		}
		seen[name] = true
		base := resolve(name)
		if base == nil {
			continue
		}
		if _, err := popTree(base, resolve, templates, seen); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	stripped := popBaseLine.ReplaceAll(body, nil)
	tokens := populationTokens(stripped)
	nodes, rest := popParse(tokens)
	if len(rest) != 0 {
		return nil, fmt.Errorf("unbalanced braces")
	}
	root := nodes
	if schedule := popChildren(nodes, "WaveSchedule"); len(schedule) > 0 {
		root = schedule[0].block
	}
	for _, block := range popChildren(root, "Templates") {
		for _, template := range block.block {
			templates[strings.ToLower(template.key)] = template.block
		}
	}
	return root, nil
}

func popParse(tokens []string) ([]popNode, []string) {
	var nodes []popNode
	for len(tokens) > 0 {
		switch tokens[0] {
		case "}":
			return nodes, tokens[1:]
		case "{":
			// A block with no key: skip it whole.
			_, tokens = popParse(tokens[1:])
			continue
		}
		key := tokens[0]
		tokens = tokens[1:]
		if len(tokens) == 0 {
			return nodes, nil
		}
		if tokens[0] == "{" {
			var block []popNode
			block, tokens = popParse(tokens[1:])
			nodes = append(nodes, popNode{key: key, block: block})
			continue
		}
		nodes = append(nodes, popNode{key: key, value: tokens[0]})
		tokens = tokens[1:]
	}
	return nodes, nil
}

func popChildren(nodes []popNode, key string) []popNode {
	var out []popNode
	for _, node := range nodes {
		if strings.EqualFold(node.key, key) && node.block != nil {
			out = append(out, node)
		}
	}
	return out
}

func popValue(nodes []popNode, key string) string {
	for _, node := range nodes {
		if strings.EqualFold(node.key, key) && node.block == nil {
			return node.value
		}
	}
	return ""
}

func popIsSpawner(key string) bool {
	switch strings.ToLower(key) {
	case "tfbot", "squad", "randomchoice", "tank", "mob":
		return true
	}
	return false
}

// popCountSpawner is what one spawner yields over total spawns: giants, tanks,
// and whether a random draw decides the giants.
func popCountSpawner(spawner popNode, templates map[string][]popNode, total, depth int) (int, int, bool) {
	if depth > 8 {
		return 0, 0, false
	}
	switch strings.ToLower(spawner.key) {
	case "tank":
		return 0, total, false
	case "tfbot":
		if popIsGiant(spawner.block, templates, 0) {
			return total, 0, false
		}
		return 0, 0, false
	case "mob":
		for _, node := range spawner.block {
			if strings.EqualFold(node.key, "TFBot") && popIsGiant(node.block, templates, 0) {
				return total, 0, false
			}
		}
		return 0, 0, false
	case "squad":
		var members []popNode
		for _, node := range spawner.block {
			if popIsSpawner(node.key) {
				members = append(members, node)
			}
		}
		if len(members) == 0 {
			return 0, 0, false
		}
		giants, tanks := 0, 0
		for i := range total {
			g, t, _ := popCountSpawner(members[i%len(members)], templates, 1, depth+1)
			giants += g
			tanks += t
		}
		return giants, tanks, false
	case "randomchoice":
		all, any := true, false
		for _, node := range spawner.block {
			if !popIsSpawner(node.key) {
				continue
			}
			g, _, _ := popCountSpawner(node, templates, 1, depth+1)
			all = all && g > 0
			any = any || g > 0
		}
		switch {
		case all && any:
			return total, 0, false
		case any:
			return 0, 0, true
		}
	}
	return 0, 0, false
}

func popIsGiant(bot []popNode, templates map[string][]popNode, depth int) bool {
	if depth > 10 {
		return false
	}
	if popHasMiniBoss(bot, 0) {
		return true
	}
	for _, node := range bot {
		if strings.EqualFold(node.key, "Template") && node.block == nil {
			if template, ok := templates[strings.ToLower(node.value)]; ok && popIsGiant(template, templates, depth+1) {
				return true
			}
		}
	}
	return false
}

// popHasMiniBoss looks through the bot and its EventChangeAttributes blocks,
// and not into an item's own attributes.
func popHasMiniBoss(nodes []popNode, depth int) bool {
	if depth > 4 {
		return false
	}
	for _, node := range nodes {
		if strings.EqualFold(node.key, "Attributes") && strings.EqualFold(node.value, "MiniBoss") {
			return true
		}
		if node.block == nil {
			continue
		}
		switch strings.ToLower(node.key) {
		case "characterattributes", "itemattributes", "item":
			continue
		}
		if popHasMiniBoss(node.block, depth+1) {
			return true
		}
	}
	return false
}
