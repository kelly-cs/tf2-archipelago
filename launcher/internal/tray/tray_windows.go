//go:build windows

package tray

import (
	"fyne.io/systray"

	"github.com/m-this/tf2-archipelago/launcher/internal/browser"
)

/*
Run puts the icon up and runs the launcher beside it, and returns what the
launcher returned once both are gone.

The message loop wants the thread that started the program, which is why this
is called from main rather than started from anywhere: systray locks the main
goroutine to it at init. The launcher runs on a goroutine of its own, and
pulling the icon down when it quits is what ends the loop.
*/
func Run(icon []byte, launch Launch) error {
	quit := make(chan error, 1)
	systray.Run(func() {
		systray.SetIcon(icon)
		systray.SetTooltip("TF2 Archipelago")
		open := systray.AddMenuItem("Open the launcher", "Open the launcher in your browser")
		systray.AddSeparator()
		stop := systray.AddMenuItem("Quit", "Stop the server and close the launcher")

		done := make(chan struct{})
		go func() {
			quit <- launch(func(url string, quitLauncher func()) {
				systray.SetOnTapped(func() { _ = browser.Open(url) })
				go answer(url, quitLauncher, open, stop, done)
			})
			close(done)
			systray.Quit()
		}()
	}, nil)
	return <-quit
}

// answer is the menu, until the launcher is gone. There is no bound on how
// many times a player opens the page in an evening; done is what ends it.
func answer(url string, quit func(), open, stop *systray.MenuItem, done <-chan struct{}) {
	for {
		select {
		case <-open.ClickedCh:
			_ = browser.Open(url)
		case <-stop.ClickedCh:
			quit()
		case <-done:
			return
		}
	}
}
