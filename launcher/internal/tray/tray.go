/*
Package tray is the launcher's icon in the notification area.

The launcher has no window: it serves a page and opens a browser on it. Close
the tab and the launcher is still there, running the server, with nothing on
the desktop to say so and no way back to it but the address in a log. The icon
is that way back: a click opens the page again, and the menu has Quit.

Windows only. On Linux the launcher is started from a terminal, and that
terminal is where it lives.
*/
package tray

// Serving is what the launcher tells the tray once its page answers: where the
// page is, and what to call to quit. It is the shape of webapi.Options.Serving.
type Serving = func(url string, quit func())

// Launch runs the launcher and returns when it has quit. It is handed the
// Serving hook to pass on.
type Launch = func(serving Serving) error
