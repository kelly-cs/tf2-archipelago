package apclient

import "strconv"

/*
	nameBook turns the ids in a server message back into names.

Archipelago sends a chat line as typed parts, and a part that names an item, a
location or a player carries the id in its Text with the name held elsewhere.
Reading Text alone therefore prints "Also it didn't print names of items or
locations, just IDs", which is what a player reported.

Player names arrive with Connected. Item and location names come from the data
package, which is a separate request and is keyed by game, because two games
may both number an item 1. The owning slot decides which game's names to read.

Every lookup falls back to the id. A part this cannot resolve is still readable,
and a chat line with one raw number in it beats no chat line at all.
*/
type nameBook struct {
	players map[int]string
	games   map[int]string              // slot -> game
	items   map[string]map[int64]string // game -> id -> name
	places  map[string]map[int64]string // game -> id -> name
}

func newNameBook() *nameBook {
	return &nameBook{
		players: map[int]string{},
		games:   map[int]string{},
		items:   map[string]map[int64]string{},
		places:  map[string]map[int64]string{},
	}
}

func (n *nameBook) player(slot int) string {
	if name := n.players[slot]; name != "" {
		return name
	}
	return "player " + strconv.Itoa(slot)
}

func (n *nameBook) item(id int64, slot int) string {
	return n.lookup(n.items, id, slot)
}

func (n *nameBook) location(id int64, slot int) string {
	return n.lookup(n.places, id, slot)
}

func (n *nameBook) lookup(from map[string]map[int64]string, id int64, slot int) string {
	if name := from[n.games[slot]][id]; name != "" {
		return name
	}
	// The owning slot is not always the one the id belongs to, and a part can
	// arrive with no player at all. One unambiguous match across every game is
	// better than a bare number; two are not, so those stay numbers.
	var found string
	for _, byID := range from {
		switch name := byID[id]; {
		case name == "":
		case found == "":
			found = name
		case found != name:
			return strconv.FormatInt(id, 10)
		}
	}
	if found != "" {
		return found
	}
	return strconv.FormatInt(id, 10)
}
