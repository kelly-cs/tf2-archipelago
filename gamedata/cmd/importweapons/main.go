// Command importweapons reads TF2's item schema and prints candidate functional
// weapons as JSON. A maintainer tool, not part of a build: Archipelago item ids
// are permanent, so a newly discovered weapon is reviewed and appended to the
// catalogue by hand rather than applied.
//
// Usage: go run ./gamedata/cmd/importweapons <items_game.txt> <tf_english.txt>
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata/internal/tfschema"
)

var weaponSlots = map[string]bool{
	"primary": true, "secondary": true, "melee": true, "building": true, "pda": true, "pda2": true,
}

var variantNames = []string{"botkiller", "festive"}

// candidate is one item definition that sits in a weapon slot.
type candidate struct {
	DefIndex  int
	Name      string
	Slot      string
	ItemClass string
	// VariantOf is the definition this one is a paint of, or zero.
	VariantOf int
}

// Weapon is what is printed: one name, and every definition index that is it.
type Weapon struct {
	Name       string `json:"name"`
	Slot       string `json:"slot"`
	ItemClass  string `json:"item_class"`
	DefIndexes []int  `json:"def_indexes"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: importweapons <items_game.txt> <tf_english.txt>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "importweapons:", err)
		os.Exit(1)
	}
}

func run(schemaPath, englishPath string) error {
	root, err := tfschema.ParseFile(schemaPath)
	if err != nil {
		return err
	}
	names, err := tfschema.English(englishPath)
	if err != nil {
		return err
	}
	items, prefabs := root.Block("items"), root.Block("prefabs")
	if items == nil || prefabs == nil {
		return fmt.Errorf("%s has no items or no prefabs block", schemaPath)
	}
	found := candidates(items, prefabs, names)
	grouped := group(found)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(grouped)
}

func candidates(items, prefabs *tfschema.Object, names map[string]string) []candidate {
	var found []candidate
	for _, key := range items.Keys() {
		index, err := strconv.Atoi(key)
		item := items.Block(key)
		if err != nil || item == nil {
			continue
		}
		merged := merge(item, prefabs)
		// enabled is not a usability flag here: stock and promotional
		// weapon prefabs use zero even though their concrete item
		// definitions work.
		slot := merged["item_slot"]
		if !weaponSlots[slot] {
			continue
		}
		token := strings.ToLower(strings.TrimPrefix(merged["item_name"], "#"))
		name, ok := names[token]
		if !ok {
			name = strings.TrimPrefix(item.Text("name"), "The ")
		}
		if name == "" {
			continue
		}
		variantOf := 0
		if static := item.Block("static_attrs"); static != nil {
			variantOf, _ = strconv.Atoi(static.Text("paintkit_proto_def_index"))
		}
		found = append(found, candidate{
			DefIndex: index, Name: name, Slot: slot,
			ItemClass: merged["item_class"], VariantOf: variantOf,
		})
	}
	slices.SortStableFunc(found, func(a, b candidate) int {
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return a.DefIndex - b.DefIndex
	})
	return found
}

// merge flattens an item over its prefabs, parents first, so the item's own
// text fields win. Only the string fields are read, which is all the tools ask.
func merge(item *tfschema.Object, prefabs *tfschema.Object) map[string]string {
	merged := make(map[string]string)
	seen := make(map[string]bool)
	var add func(name string)
	add = func(name string) {
		prefab := prefabs.Block(name)
		if seen[name] || prefab == nil {
			return
		}
		seen[name] = true
		for parent := range strings.FieldsSeq(prefab.Text("prefab")) {
			add(parent)
		}
		copyText(merged, prefab)
	}
	for name := range strings.FieldsSeq(item.Text("prefab")) {
		add(name)
	}
	copyText(merged, item)
	return merged
}

func copyText(into map[string]string, from *tfschema.Object) {
	for _, key := range from.Keys() {
		if v, _ := from.Get(key); v.Block == nil {
			into[key] = v.Text
		}
	}
}

// group folds the candidates into weapons: one per name, with the paints,
// the Festives and the Botkillers attached to the gun they are a paint of.
func group(entries []candidate) []*Weapon {
	grouped := make(map[string]*Weapon)
	var order []*Weapon
	for _, e := range entries {
		lower := strings.ToLower(e.Name)
		if e.ItemClass == "slot_token" || e.VariantOf != 0 || slices.ContainsFunc(variantNames, func(m string) bool { return strings.Contains(lower, m) }) {
			continue
		}
		w, ok := grouped[e.Name]
		if !ok {
			w = &Weapon{Name: e.Name, Slot: e.Slot, ItemClass: e.ItemClass}
			grouped[e.Name] = w
			order = append(order, w)
		}
		w.DefIndexes = append(w.DefIndexes, e.DefIndex)
	}

	byDefinition := make(map[int]*Weapon)
	byName := make(map[string]*Weapon)
	for name, w := range grouped {
		byName[strings.ToLower(name)] = w
		for _, d := range w.DefIndexes {
			byDefinition[d] = w
		}
	}
	for _, e := range entries {
		var w *Weapon
		if e.VariantOf != 0 {
			w = byDefinition[e.VariantOf]
		}
		lower := strings.ToLower(e.Name)
		if w == nil && strings.HasPrefix(lower, "festive ") {
			w = byName[lower[len("festive "):]]
		}
		if w == nil && strings.Contains(lower, "botkiller") {
			for key, c := range byName {
				if strings.Contains(lower, key) && (w == nil || len(c.Name) > len(w.Name)) {
					w = c
				}
			}
		}
		if w != nil && !slices.Contains(w.DefIndexes, e.DefIndex) {
			w.DefIndexes = append(w.DefIndexes, e.DefIndex)
			slices.Sort(w.DefIndexes)
		}
	}
	slices.SortStableFunc(order, func(a, b *Weapon) int {
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return a.DefIndexes[0] - b.DefIndexes[0]
	})
	return order
}
