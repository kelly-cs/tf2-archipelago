package apclient

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/m-this/tf2-archipelago/fakeroom"
	"github.com/m-this/tf2-archipelago/gamedata/availability"
)

const binaryBroadcast = "mvm_transmission_rc7a_int_binary_broadcast"

// JustBacon's run drew Binary Broadcast on a server with no community pack,
// and nobody heard about it until the run switched to it. The seed is known
// the moment the room accepts the slot, so that is where it has to be said.
func TestUnplayableNamesWhatTheSeedDrawsAndTheFix(t *testing.T) {
	potato := []string{"archive-assets.zip"}
	tests := []struct {
		name     string
		server   *availability.Server
		missions []string
		refusal  []string
	}{
		{
			name:     "community mission whose pack is not on disk",
			server:   &availability.Server{Platform: "linux"},
			missions: []string{"mvm_decoy", binaryBroadcast},
			refusal:  []string{"Binary Broadcast", "Potato Archive is not on this server", "Settings > Missions"},
		},
		{
			name:     "community mission whose pack is on disk and not ticked",
			server:   &availability.Server{Platform: "linux", PacksPresent: potato},
			missions: []string{binaryBroadcast},
			refusal:  []string{"Binary Broadcast", "tick Potato Archive", "Settings > Missions"},
		},
		{
			name: "SigMod mission with SigMod not set up",
			server: &availability.Server{
				Platform: "linux", PacksSelected: potato, PacksPresent: potato,
				ModsSelected: []string{"sigsegv-mvm"},
			},
			missions: []string{"mvm_bronx_rc2_adv_point_of_impact"},
			refusal:  []string{"Point Of Impact", "SigMod is selected but its installation is missing"},
		},
		{
			name:     "community mission whose pack is installed",
			server:   &availability.Server{Platform: "linux", PacksSelected: potato, PacksPresent: potato},
			missions: []string{"mvm_decoy", binaryBroadcast},
		},
		{
			name:     "Valve missions only, on a server with no pack",
			server:   &availability.Server{Platform: "linux"},
			missions: []string{"mvm_decoy", "mvm_coaltown"},
		},
		{
			name:     "a mission the tables do not know is the plugin's to refuse",
			server:   &availability.Server{Platform: "linux"},
			missions: []string{"mvm_nowhere"},
		},
		{
			name:     "a bridge that was told nothing about its server",
			missions: []string{"mvm_decoy", binaryBroadcast},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := unplayable(test.server, test.missions)
			if test.refusal == nil {
				if err != nil {
					t.Fatalf("refused a seed this server can play: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("kept a seed this server cannot play")
			}
			for _, want := range test.refusal {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v", want, err)
				}
			}
		})
	}
}

// The refusal ends the session rather than retrying into it, and the bridge
// does not report itself connected to the seed it refused.
func TestASeedThisServerCannotPlayStopsTheBridge(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	room, address, err := fakeroom.Start(ctx, fakeroom.Options{
		SlotName: "tester",
		Missions: []string{"mvm_decoy", binaryBroadcast},
		Goal:     "final_boss",
		Log:      func(string) {},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = room.Close(context.Background()) })

	client := New(Options{
		URL:      address,
		SlotName: "tester",
		Store:    newStore(t),
		Logger:   slog.New(slog.DiscardHandler),
		Missions: &availability.Server{Platform: "linux"},
	})
	ended := make(chan error, 1)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		ended <- client.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-stopped
	})
	select {
	case err := <-ended:
		if err == nil || !strings.Contains(err.Error(), "Binary Broadcast") {
			t.Fatalf("the bridge stopped without naming the mission: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("the bridge kept a seed this server cannot play")
	}
	if client.Health().Connected {
		t.Error("the bridge reported itself connected to a seed it refused")
	}
}
