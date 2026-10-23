package botfiles

import (
	"os"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
A seat given a name reaches the file, without asking for weapons.

The file is what the mod reads a seat's name out of, and it used to be written
only when some seat or class had picked a weapon: naming seat one and changing
nothing else wrote nothing at all, so the bot drew a name as though nobody had
said anything. The convar stays off, because that one is about weapons.
*/
func TestSeatNameReachesTheFileWithoutTurningOnTheWeapons(t *testing.T) {
	root := t.TempDir()
	s := settings.Defaults()
	s.SrcdsBotTeamComp = []string{"pyro", "engineer"}
	s.SrcdsBotSeatNames = []string{"Gravel Pit Gary", ""}

	weapons, err := Install(root, s)
	if err != nil {
		t.Fatal(err)
	}
	if weapons {
		t.Error("naming a seat turned on the custom loadouts")
	}
	body, err := os.ReadFile(LoadoutPath(root))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"name"	"Gravel Pit Gary"`) {
		t.Fatalf("the seat name is not in the file:\n%s", body)
	}
}
