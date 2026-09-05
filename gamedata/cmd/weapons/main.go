// Command weapons writes gamedata/weapons_generated.go from the bot mod's
// weapon pools and the game's own schema. The pools are what the mod will hand
// a bot; the schema is what the weapon is called. Neither is written by hand
// here, so the catalogue cannot drift from either.
//
// Usage: go run ./gamedata/cmd/weapons -pools <loadouts.sp> -schema <items_game.txt> -english <tf_english.txt>
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata/internal/tfschema"
)

// poolLine is one generated pool: int WEAPONS_SCOUT_PRIMARY[31] = {...};
var poolLine = regexp.MustCompile(`^int WEAPONS_([A-Z]+)_([A-Z0-9]+)\[\d+\] = \{(.*)\};$`)

// reskin is a repaint of a gun already in the list. The 15xxx block is not the
// whole of them: the Botkillers and the Festives sit in the 6xx to 11xx range
// with the real weapons, and a Scout primary menu of twelve Scatterguns is the
// thing this catalogue exists to avoid.
var reskin = regexp.MustCompile(`(?i)\bBotkiller\b|^Festive |^Australium `)

var (
	slots = map[string]string{
		"PRIMARY": "primary", "SECONDARY": "secondary", "MELEE": "melee",
		"PDA2": "pda2", "BUILDING": "building",
	}
	classes = map[string]string{
		"SCOUT": "scout", "SOLDIER": "soldier", "PYRO": "pyro", "DEMOMAN": "demoman",
		"HEAVY": "heavyweapons", "ENGINEER": "engineer", "MEDIC": "medic",
		"SNIPER": "sniper", "SPY": "spy",
	}
)

// reskinIndexMin is where the schema's own repaint block starts. Nothing above
// it is a distinct gun.
const reskinIndexMin = 15000

type pool struct {
	class, slot string
	indexes     []int
}

type row struct {
	class, slot string
	index       int
	name        string
}

func main() {
	pools := flag.String("pools", "", "the mod's generated loadouts.sp")
	schema := flag.String("schema", "", "tf/scripts/items/items_game.txt")
	english := flag.String("english", "", "tf/resource/tf_english.txt")
	flag.Parse()
	if *pools == "" || *schema == "" || *english == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*pools, *schema, *english); err != nil {
		fmt.Fprintln(os.Stderr, "weapons:", err)
		os.Exit(1)
	}
}

func run(poolsPath, schemaPath, englishPath string) error {
	pools, err := readPools(poolsPath)
	if err != nil {
		return err
	}
	root, err := tfschema.ParseFile(schemaPath)
	if err != nil {
		return err
	}
	items := root.Block("items")
	if items == nil {
		return fmt.Errorf("%s has no items block", schemaPath)
	}
	names, err := tfschema.English(englishPath)
	if err != nil {
		return err
	}

	var rows []row
	var missing []row
	for _, p := range pools {
		for _, index := range p.indexes {
			item := items.Block(strconv.Itoa(index))
			internal := ""
			if item != nil {
				internal = item.Text("name")
			}
			// "Upgradeable TF_WEAPON_X" is the MvM stock weapon, which the
			// menu already offers as stock. Two entries for one gun is noise.
			if strings.HasPrefix(internal, "Upgradeable ") {
				continue
			}
			name := ""
			if item != nil {
				if token := item.Text("item_name"); token != "" {
					name = names[strings.ToLower(strings.TrimPrefix(token, "#"))]
				}
			}
			if name == "" && internal != "" && !strings.Contains(internal, "TF_") {
				name = internal
			}
			r := row{class: p.class, slot: p.slot, index: index, name: name}
			if name == "" {
				missing = append(missing, r)
				continue
			}
			if reskin.MatchString(name) {
				continue
			}
			rows = append(rows, r)
		}
	}
	for _, r := range missing {
		fmt.Fprintf(os.Stderr, "no name for %s %s %d\n", r.class, r.slot, r.index)
	}
	slices.SortStableFunc(rows, func(a, b row) int {
		return strings.Compare(a.class+"\x00"+a.slot+"\x00"+strings.ToLower(a.name),
			b.class+"\x00"+b.slot+"\x00"+strings.ToLower(b.name))
	})
	fmt.Fprintf(os.Stderr, "%d entries, %d dropped for want of a name\n", len(rows), len(missing))

	source, err := format.Source([]byte(render(rows)))
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(source)
	return err
}

func readPools(path string) ([]pool, error) {
	file, err := os.Open(path) //nolint:gosec // the path is a flag on a maintainer tool
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	var pools []pool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := poolLine.FindStringSubmatch(strings.TrimSpace(scanner.Text()))
		if m == nil {
			continue
		}
		class, slot := classes[m[1]], slots[m[2]]
		if class == "" || slot == "" {
			fmt.Fprintf(os.Stderr, "skipping WEAPONS_%s_%s\n", m[1], m[2])
			continue
		}
		p := pool{class: class, slot: slot}
		for field := range strings.SplitSeq(m[3], ",") {
			// The default index is a named constant, and the repaints
			// are the block above 15000.
			index, err := strconv.Atoi(strings.TrimSpace(field))
			if err != nil || index >= reskinIndexMin {
				continue
			}
			p.indexes = append(p.indexes, index)
		}
		pools = append(pools, p)
	}
	if len(pools) == 0 {
		return nil, fmt.Errorf("%s holds no WEAPONS_ pools", path)
	}
	return pools, scanner.Err()
}

func render(rows []row) string {
	var b strings.Builder
	b.WriteString(`// Code generated by gamedata/cmd/weapons. DO NOT EDIT.

package gamedata

// Weapons is every non-stock item the mod can give a bot, drawn from the
// WEAPONS_* pools the bot mod generates into loadouts.sp. Stock is not in
// here: the mod spells it as the default index and every menu offers it
// separately.
//
// The reskins are left out. The 15xxx and 30xxx blocks are Festive and
// Botkiller repaints of guns already listed, and a Scout primary menu of
// thirty entries where twenty are Scatterguns is not a menu.
var Weapons = []Weapon{
`)
	for _, r := range rows {
		fmt.Fprintf(&b, "\t{%d, %s, %s, %s},\n", r.index, strconv.Quote(r.name), strconv.Quote(r.class), strconv.Quote(r.slot))
	}
	b.WriteString("}\n")
	return b.String()
}
