package gamedata

import (
	"slices"
	"testing"
)

func TestChamberLimitsInSourcePawn(t *testing.T) {
	got := driver{body: `
    printnum(Chamber_Quarter(0));
    printnum(Chamber_Quarter(-1));
    printnum(Chamber_Quarter(1));
    printnum(Chamber_Quarter(3));
    printnum(Chamber_Quarter(200));
    printnum(Chamber_Quarter(150));
    printnum(Chamber_ReserveLimit(200, 3));
    printnum(Chamber_ReserveLimit(200, 1));
    printnum(Chamber_ReserveLimit(16, 0));
    printnum(!Chamber_ShouldSupply(3.9, 4.0));
    printnum(Chamber_ShouldSupply(4.0, 4.0));
 `}.run(t)
	want := []int32{0, -1, 1, 1, 50, 37, 50, 50, 16, 1, 1}
	if !slices.Equal(got, want) {
		t.Fatalf("limits = %v, want %v", got, want)
	}
}

func TestMentalVisibilityWindowsInSourcePawn(t *testing.T) {
	got := driver{body: `
    printnum(Mental_CanRecloak(10.0, 10.0, false));
    printnum(!Mental_CanRecloak(9.9, 10.0, false));
    printnum(!Mental_CanRecloak(11.0, 10.0, true));
    printnum(Mental_CanRecloak(11.0, 10.0, false));
    printnum(Mental_EntityVisibility(1.0, 1.3, 0.0) == 0.15);
    printnum(Mental_EntityVisibility(1.3, 1.3, 0.0) == 0.0);
    printnum(Mental_EntityVisibility(1.0, 1.3, 6.0) == 1.0);
    printnum(Mental_EntityVisibility(5.9, 1.3, 6.0) == 1.0);
    printnum(Mental_RecloakProgress(6.0, 6.0) == 0.0);
    printnum(Mental_RecloakProgress(6.5, 6.0) == 0.5);
    printnum(Mental_RecloakProgress(7.0, 6.0) == 1.0);
    printnum(Mental_EntityVisibility(6.0, 1.3, 6.0) == 1.0);
    printnum(Mental_EntityVisibility(6.5, 1.3, 6.0) == 0.5);
    printnum(Mental_EntityVisibility(7.0, 1.3, 6.0) == 0.0);
 `}.run(t)
	for i, value := range got {
		if value != 1 {
			t.Errorf("fade/reveal case %d failed", i)
		}
	}
	if len(got) != 14 {
		t.Fatalf("got %d cases, want 14", len(got))
	}
}
