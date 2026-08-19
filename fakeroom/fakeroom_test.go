package fakeroom

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/m-this/tf2-archipelago/gamedata"
)

// client drives the room's websocket directly, the way the bridge does, so the
// dedup and the seed can be checked without a whole apclient session.
type client struct {
	t    *testing.T
	conn *websocket.Conn
}

func dial(t *testing.T, address string) *client {
	t.Helper()
	conn, handshake, err := websocket.Dial(t.Context(), address, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if handshake != nil && handshake.Body != nil {
		_ = handshake.Body.Close()
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return &client{t: t, conn: conn}
}

func (c *client) send(message any) {
	c.t.Helper()
	body, err := json.Marshal([]any{message})
	if err != nil {
		c.t.Fatal(err)
	}
	if err := c.conn.Write(c.t.Context(), websocket.MessageText, body); err != nil {
		c.t.Fatalf("write: %v", err)
	}
}

// await reads messages until one with the wanted cmd shows up.
func (c *client) await(want string) map[string]json.RawMessage {
	c.t.Helper()
	ctx, cancel := context.WithTimeout(c.t.Context(), 3*time.Second)
	defer cancel()
	for {
		_, body, err := c.conn.Read(ctx)
		if err != nil {
			c.t.Fatalf("waiting for %s: %v", want, err)
		}
		var messages []map[string]json.RawMessage
		if err := json.Unmarshal(body, &messages); err != nil {
			c.t.Fatal(err)
		}
		for _, message := range messages {
			var cmd string
			if err := json.Unmarshal(message["cmd"], &cmd); err != nil {
				c.t.Fatal(err)
			}
			if cmd == want {
				return message
			}
		}
	}
}

func startRoom(t *testing.T) string {
	t.Helper()
	room, address, err := Start(t.Context(), Options{
		SlotName:     "tester",
		MissionCount: 2,
		Log:          func(string) {},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = room.Close(context.Background()) })
	return address
}

// The starting inventory is what a generated seed precollects. Without it the
// plugin has no class, no weapon slot and no mission it may play, so it
// enforces nothing: a play-test spent four waves able to pick any class.
func TestTheRunStartsWithSomethingToPlay(t *testing.T) {
	// Not connect(): Connected and the starting inventory ride in one frame,
	// and await returns the first match in it, dropping the rest.
	c := dial(t, startRoom(t))
	c.await("RoomInfo")
	c.send(map[string]any{"cmd": "Connect"})
	received := c.await("ReceivedItems")

	var payload struct {
		Index int `json:"index"`
		Items []struct {
			Item int64 `json:"item"`
		} `json:"items"`
	}
	if err := json.Unmarshal(mustRaw(received), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Index != 0 {
		t.Errorf("the starting inventory came at index %d", payload.Index)
	}

	kinds := map[gamedata.ItemKind]int{}
	for _, held := range payload.Items {
		item, known := gamedata.ItemByID(held.Item)
		if !known {
			t.Fatalf("the room handed out item %d, which the tables do not know", held.Item)
		}
		kinds[item.Kind]++
	}
	for _, kind := range []gamedata.ItemKind{
		gamedata.ItemMissionTicket, gamedata.ItemClass, gamedata.ItemWeaponSlot,
	} {
		if kinds[kind] != 1 {
			t.Errorf("the starting inventory holds %d of %s, want 1", kinds[kind], kind.Key())
		}
	}
}

// A gift is filler. The room used to hand out the run's own unlocks once a
// minute, which emptied the pool by the clock rather than by play.
func TestGiftsAreFillerRatherThanProgression(t *testing.T) {
	item, known := gamedata.ItemByID(fillerItem())
	if !known || item.Classification != gamedata.Filler {
		t.Fatalf("a gift carries %+v, want filler", item)
	}
}

// What the starting inventory holds is off the list the checks draw from, so
// nothing is handed out twice.
func TestTheUnlockOrderDropsWhatTheRunAlreadyHolds(t *testing.T) {
	start := startingInventory("mvm_decoy")
	order := unlockOrder(start)

	slots := 0
	for _, id := range order {
		item, _ := gamedata.ItemByID(id)
		if item.Kind == gamedata.ItemWeaponSlot {
			// The progressive slot has several copies and the run holds one.
			slots++
			continue
		}
		if item.Classification == gamedata.Filler {
			t.Error("filler is in the unlock pool")
		}
		if slices.Contains(start, id) {
			t.Errorf("%s is in the starting inventory and still in the pool", item.Name)
		}
	}
	held := 0
	for _, id := range start {
		if item, _ := gamedata.ItemByID(id); item.Kind == gamedata.ItemWeaponSlot {
			held++
		}
	}
	for _, item := range gamedata.Items {
		if item.Kind != gamedata.ItemWeaponSlot {
			continue
		}
		if want := int(item.Count) - held; slots != want {
			t.Errorf("the pool holds %d weapon slots, want %d", slots, want)
		}
	}
}

func TestDefaultMissionsSkipTheExcluded(t *testing.T) {
	got := defaultMissions(2, []string{"mvm_decoy"}, "", "")
	if len(got) != 2 || got[0] != "mvm_decoy_intermediate" {
		t.Errorf("missions = %v", got)
	}
	if got := defaultMissions(1, []string{"mvm_decoy"}, "", ""); len(got) != 1 || got[0] == "mvm_decoy" {
		t.Errorf("missions = %v", got)
	}
}

// A test run is what a player shapes an evening with before generating a real
// seed. It drew the first missions of the table whatever tier they picked, so
// the one setting that decides how hard the evening is did nothing.
func TestDefaultMissionsRespectTheTier(t *testing.T) {
	for _, key := range []string{"intermediate", "advanced", "expert"} {
		floor, known := gamedata.DifficultyByKey(key)
		if !known {
			t.Fatalf("%s is not a tier", key)
		}
		for _, popFile := range defaultMissions(8, nil, key, "") {
			mission, ok := gamedata.MissionByPopFile(popFile)
			if !ok {
				t.Fatalf("drew %s, which is not a mission", popFile)
			}
			// A floor, not a filter: the tier and everything harder.
			if mission.Difficulty < floor {
				t.Errorf("%s drew %s, which is %s", key, mission.Name, mission.Difficulty.Key())
			}
		}
	}
	// An unknown key is a typo in a settings file, and draws the whole pool
	// rather than nothing.
	if got := defaultMissions(3, nil, "nonsense", ""); len(got) != 3 {
		t.Errorf("an unknown tier drew %v", got)
	}
}

// The run begins on the first mission drawn, so a named start mission has to
// come first even when the tier would not have drawn it at all.
func TestDefaultMissionsStartWhereAsked(t *testing.T) {
	got := defaultMissions(4, nil, "normal", "mvm_coaltown_advanced")
	if len(got) == 0 || got[0] != "mvm_coaltown_advanced" {
		t.Fatalf("missions = %v", got)
	}
	if len(got) != 4 {
		t.Errorf("drew %d missions, want 4: %v", len(got), got)
	}
	// Named but outside the tier: the player asked for it by name, which is
	// more specific than the tier they asked for by key.
	got = defaultMissions(2, nil, "expert", "mvm_decoy")
	if len(got) == 0 || got[0] != "mvm_decoy" {
		t.Errorf("missions = %v", got)
	}
	// Not a mission at all: ignored rather than served as one.
	got = defaultMissions(2, nil, "normal", "mvm_nowhere")
	if slices.Contains(got, "mvm_nowhere") {
		t.Errorf("served a mission that does not exist: %v", got)
	}
}

func connect(t *testing.T, address string) *client {
	t.Helper()
	c := dial(t, address)
	c.await("RoomInfo")
	c.send(map[string]any{"cmd": "Connect"})
	c.await("Connected")
	return c
}

func itemCount(t *testing.T, message map[string]json.RawMessage) int {
	t.Helper()
	var items []json.RawMessage
	if err := json.Unmarshal(message["items"], &items); err != nil {
		t.Fatal(err)
	}
	return len(items)
}

// This is the bug: the bridge resends its whole check list on every report,
// which is correct against a real server because a repeat check there hands
// out nothing new. The room has to make that true itself, or a resend floods
// unlocks nobody earned.
func TestASecondReportOfTheSameChecksGrantsNothingMore(t *testing.T) {
	address := startRoom(t)
	c := connect(t, address)

	c.send(map[string]any{"cmd": "LocationChecks", "locations": []int64{111, 222}})
	first := c.await("ReceivedItems")
	if got := itemCount(t, first); got != 2 {
		t.Fatalf("first report granted %d item(s), want 2", got)
	}

	// The same two locations again, the way a reconnect or a ping-triggered
	// report resends them. A third, genuinely new one is mixed in so a
	// wrongly-silent room cannot be mistaken for one that dropped the message.
	c.send(map[string]any{"cmd": "LocationChecks", "locations": []int64{111, 222, 333}})
	second := c.await("ReceivedItems")
	if got := itemCount(t, second); got != 1 {
		t.Fatalf("the resend granted %d item(s), want 1 for the one new location", got)
	}
}

// Test mode's state on the bridge side is durable and keyed by seed name
// (BindSeed). A constant seed name would have every restart treat the
// bridge's already-recorded checks as still due a reward, and this
// memory-only room has no record of ever having paid them.
func TestEachStartUsesADifferentSeed(t *testing.T) {
	addressA := startRoom(t)
	addressB := startRoom(t)

	seedA := seedOf(t, addressA)
	seedB := seedOf(t, addressB)
	if seedA == seedB {
		t.Fatalf("two rooms shared the seed %q", seedA)
	}
}

func seedOf(t *testing.T, address string) string {
	t.Helper()
	c := dial(t, address)
	roomInfo := c.await("RoomInfo")
	var seed string
	if err := json.Unmarshal(roomInfo["seed_name"], &seed); err != nil {
		t.Fatal(err)
	}
	return seed
}
