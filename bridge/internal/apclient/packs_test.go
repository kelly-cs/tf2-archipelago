package apclient

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/m-this/tf2-archipelago/fakeroom"
)

// JustBacon's run drew Binary Broadcast on a server with no community pack,
// and nobody heard about it until the run switched to it. The seed is known
// the moment the room accepts the slot, so that is where it has to be said.
func TestASeedWhoseCommunityMissionIsNotInstalledStopsTheBridge(t *testing.T) {
	const binaryBroadcast = "mvm_transmission_rc7a_int_binary_broadcast"
	tests := []struct {
		name      string
		missions  []string
		installed []string
		dir       bool
		refusal   []string
	}{
		{
			name:     "community mission without its pop file",
			missions: []string{"mvm_decoy", binaryBroadcast},
			dir:      true,
			refusal:  []string{"Binary Broadcast", "Potato Archive", "Settings > Missions"},
		},
		{
			name:      "community mission with its pop file",
			missions:  []string{"mvm_decoy", binaryBroadcast},
			installed: []string{binaryBroadcast},
			dir:       true,
		},
		{
			name:     "Valve missions only, whose pop files live in the game's own archives",
			missions: []string{"mvm_decoy", "mvm_coaltown"},
			dir:      true,
		},
		{
			name:     "no population directory to look in",
			missions: []string{"mvm_decoy", binaryBroadcast},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			room, address, err := fakeroom.Start(ctx, fakeroom.Options{
				SlotName: "tester",
				Missions: test.missions,
				Goal:     "final_boss",
				Log:      func(string) {},
			})
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			t.Cleanup(func() { _ = room.Close(context.Background()) })

			population := ""
			if test.dir {
				population = t.TempDir()
				for _, popFile := range test.installed {
					if err := os.WriteFile(filepath.Join(population, popFile+".pop"), nil, 0o644); err != nil {
						t.Fatal(err)
					}
				}
			}
			client := New(Options{
				URL:           address,
				SlotName:      "tester",
				Store:         newStore(t),
				Logger:        slog.New(slog.DiscardHandler),
				PopulationDir: population,
			})
			ended := make(chan error, 1)
			go func() { ended <- client.Run(ctx) }()

			if test.refusal == nil {
				waitFor(t, "the handshake", func() bool { return client.Health().Connected })
				return
			}
			select {
			case err := <-ended:
				if err == nil {
					t.Fatal("the bridge stopped without saying why")
				}
				for _, want := range test.refusal {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("the refusal does not say %q: %v", want, err)
					}
				}
			case <-ctx.Done():
				t.Fatal("the bridge kept a seed this server cannot play")
			}
			if client.Health().Connected {
				t.Error("the bridge reported itself connected to a seed it refused")
			}
		})
	}
}
