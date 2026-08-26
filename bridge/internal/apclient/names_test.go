package apclient

import (
	"encoding/json"
	"testing"
)

// The line a player sees, assembled from the typed parts the server sends.
func renderJSON(t *testing.T, raw string, names *nameBook) string {
	t.Helper()
	var printed printJSON
	if err := json.Unmarshal([]byte(raw), &printed); err != nil {
		t.Fatalf("cannot read the message: %v", err)
	}
	return printed.text(names)
}

func book() *nameBook {
	n := newNameBook()
	n.players[1] = "Cowser"
	n.players[2] = "Peppy"
	n.games[1] = "Team Fortress 2 Mann vs Machine"
	n.games[2] = "Ocarina of Time"
	n.items["Team Fortress 2 Mann vs Machine"] = map[int64]string{77: "Cash Bundle"}
	n.places["Team Fortress 2 Mann vs Machine"] = map[int64]string{5: "Doe's Drill Wave 1"}
	n.items["Ocarina of Time"] = map[int64]string{77: "Hookshot"}
	return n
}

/*
Ids become names.

Reported as "it didn't print names of items or locations, just IDs". The server
sends a chat line as typed parts and puts the id in Text, so a reader that takes
Text alone prints numbers at people.
*/
func TestAChatLineNamesItsItemsAndPlayers(t *testing.T) {
	raw := `{"type":"ItemSend","data":[
		{"type":"player_id","text":"1"},
		{"type":"text","text":" found "},
		{"type":"item_id","text":"77","player":1},
		{"type":"text","text":" at "},
		{"type":"location_id","text":"5","player":1}]}`
	if got, want := renderJSON(t, raw, book()), "Cowser found Cash Bundle at Doe's Drill Wave 1"; got != want {
		t.Errorf("line = %q, want %q", got, want)
	}
}

// The owning slot decides whose numbering an id belongs to, because two games
// may both number an item 77.
func TestTheOwningSlotPicksTheGame(t *testing.T) {
	raw := `{"type":"ItemSend","data":[{"type":"item_id","text":"77","player":2}]}`
	if got := renderJSON(t, raw, book()); got != "Hookshot" {
		t.Errorf("item = %q, want %q", got, "Hookshot")
	}
}

/*
An id that cannot be resolved stays an id.

A chat line with one raw number in it beats no chat line, and an id that two
games disagree about is worse as a guess than as a number.
*/
func TestAnUnresolvedIDIsLeftAlone(t *testing.T) {
	raw := `{"type":"ItemSend","data":[{"type":"item_id","text":"999","player":1}]}`
	if got := renderJSON(t, raw, book()); got != "999" {
		t.Errorf("unknown item = %q, want %q", got, "999")
	}
	ambiguous := `{"type":"ItemSend","data":[{"type":"item_id","text":"77","player":9}]}`
	if got := renderJSON(t, ambiguous, book()); got != "77" {
		t.Errorf("ambiguous item = %q, want %q", got, "77")
	}
}

// Before the handshake there is no book, and the line still has to render.
func TestNoBookLeavesTheTextAsItCame(t *testing.T) {
	raw := `{"type":"ItemSend","data":[{"type":"player_id","text":"1"},{"type":"text","text":" joined"}]}`
	if got := renderJSON(t, raw, nil); got != "1 joined" {
		t.Errorf("line = %q, want %q", got, "1 joined")
	}
}

// A slot nobody named still reads as something.
func TestAnUnnamedPlayerFallsBackToItsSlot(t *testing.T) {
	if got := book().player(7); got != "player 7" {
		t.Errorf("player = %q, want %q", got, "player 7")
	}
}
