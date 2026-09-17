package webapi

import (
	"strings"
	"testing"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
)

func TestQhullWarningBurst(t *testing.T) {
	var n logNoise
	var got []apruntime.Line
	for copy := 0; copy < 303; copy++ {
		for i, text := range qhullWarning {
			text += "\x1b[38;2;255;255;255m"
			if copy == 302 && i == 3 {
				text += "L: Mapchange to mvm_rottenburg"
			}
			got = append(got, n.filter(apruntime.Line{Source: "srcds", Text: text})...)
		}
	}
	if len(got) != 6 || !strings.Contains(got[4].Text, "302 repeated") || got[5].Text != "L: Mapchange to mvm_rottenburg" {
		t.Fatalf("lost warning or mapchange: %+v", got)
	}
	for _, text := range qhullWarning {
		got = append(got, n.filter(apruntime.Line{Source: "srcds", Text: text})...)
	}
	if len(got) != 10 {
		t.Fatal("a new burst must show its first warning")
	}
}

func TestQhullPreservesUnexpectedOutput(t *testing.T) {
	var n logNoise
	first := apruntime.Line{Source: "srcds", Text: qhullWarning[0]}
	n.filter(first)
	other := apruntime.Line{Source: "bridge", Text: "connected"}
	if got := n.filter(other); len(got) != 1 || got[0] != other {
		t.Fatal(got)
	}
	failure := apruntime.Line{Source: "srcds", Text: "a different error"}
	got := n.filter(failure)
	if len(got) != 2 || got[0] != first || got[1] != failure {
		t.Fatal(got)
	}
}
