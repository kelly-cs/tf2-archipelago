package botnames

import (
	"slices"
	"strings"
	"testing"
)

func TestTheShippedPoolIsTheFileTheModReads(t *testing.T) {
	shipped := Shipped()
	if len(shipped) < 20 {
		t.Fatalf("%d shipped names, want the whole list", len(shipped))
	}
	if !slices.Contains(shipped, "Archimedes!") {
		t.Errorf("the shipped names are not the ones in bot_names.txt: %v", shipped[:5])
	}
	for _, name := range shipped {
		if name != strings.TrimSpace(name) || name == "" {
			t.Errorf("%q is not a name the mod can use", name)
		}
		if len(name) > NameMax {
			t.Errorf("%q is %d bytes, which the game would cut at %d", name, len(name), NameMax)
		}
	}
}

// What a player changed is kept as the difference, so a name added upstream
// reaches a settings file written before it existed.
func TestThePoolIsTheShippedListLessWhatWasTakenOutPlusWhatWasAdded(t *testing.T) {
	pool := Pool([]string{"Archimedes!"}, []string{"Gravel Pit Gary"})
	if slices.Contains(pool, "Archimedes!") {
		t.Error("a name taken out is still drawn")
	}
	if !slices.Contains(pool, "Gravel Pit Gary") {
		t.Error("a name added is not drawn")
	}
	if got := pool[len(pool)-1]; got != "Gravel Pit Gary" {
		t.Errorf("the added name is at %q, want the end of the list", got)
	}
	if len(pool) != len(Shipped()) {
		t.Errorf("%d names, want one out and one in on %d", len(pool), len(Shipped()))
	}
}

// The mod draws by index, so a name in the list twice comes up twice as often.
func TestANameAddedThatIsAlreadyShippedIsDrawnOnce(t *testing.T) {
	pool := Pool(nil, []string{"Archimedes!", "Archimedes!"})
	count := 0
	for _, name := range pool {
		if name == "Archimedes!" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("Archimedes! is in the pool %d times", count)
	}
}

// Taking out every shipped name and adding none is a pool of nothing, which is
// a thing a player can ask for and the mod has an answer to.
func TestAnEmptyPoolIsWrittenAsAnEmptyList(t *testing.T) {
	body := Render(Shipped(), nil)
	if strings.Count(body, "\n") != 1 || !strings.HasPrefix(body, "// Managed by tf2ap") {
		t.Fatalf("an empty pool rendered as:\n%s", body)
	}
}

func TestRenderIsOneNamePerLineUnderAHeader(t *testing.T) {
	body := Render([]string{"Archimedes!"}, []string{"Gravel Pit Gary"})
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	if !strings.HasPrefix(lines[0], "// Managed by tf2ap") {
		t.Fatalf("no header: %q", lines[0])
	}
	if slices.Contains(lines, "Archimedes!") {
		t.Error("a name taken out was written")
	}
	if lines[len(lines)-1] != "Gravel Pit Gary" {
		t.Errorf("last line is %q", lines[len(lines)-1])
	}
}
