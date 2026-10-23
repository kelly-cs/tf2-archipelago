//go:build !windows

package tray

// Run has no icon to show here, so it is the launcher and nothing beside it.
func Run(_ []byte, launch Launch) error {
	return launch(func(string, func()) {})
}
