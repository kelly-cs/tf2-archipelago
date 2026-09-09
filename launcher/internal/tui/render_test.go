package tui

import (
	"testing"
	"time"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
)

// TestRenderForTheEye draws the screens so a person can look at them:
// go test ./launcher/internal/tui -run TestRenderForTheEye -v
//
// It asserts nothing the other tests do not. What it is for is the thing no
// assertion catches, which is a screen that reads badly.
func TestRenderForTheEye(t *testing.T) {
	m := screen(t)
	for _, text := range []string{
		"Mann vs Machine wave 1 of Doe's Drill started",
		"cleared wave 1: check sent",
		"received Scout from Roseburst",
		"the bots bought damage at the upgrade station",
	} {
		m.take(apruntime.Line{At: time.Now(), Source: "srcds", Text: text})
	}
	m.drain()

	t.Log("\n" + m.View())

	m.Update(key(","))
	t.Log("\n" + m.form.view(m.width, m.height))
}
