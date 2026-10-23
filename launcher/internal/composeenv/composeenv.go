// Package composeenv persists launcher settings in the .env file Docker
// Compose reads. It preserves comments, ordering and variables the launcher
// does not own; only values changed in the browser are rewritten.
package composeenv

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// Write applies the settings that changed between before and after to path.
// The file is truncated in place rather than replaced: Compose bind-mounts the
// inode, so an atomic rename would update only the container's mount and leave
// the operator's host .env unchanged.
func Write(path string, before, after settings.Settings) error {
	oldValues, newValues := settings.EnvValues(before), settings.EnvValues(after)
	changes := make(map[string]string)
	for name, value := range newValues {
		if oldValues[name] != value {
			changes[name] = value
		}
	}
	if len(changes) == 0 {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read Compose settings %s: %w", path, err)
	}
	lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	written := make(map[string]bool)
	for i, line := range lines {
		name, _, found := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		value, changed := changes[name]
		if !found || !changed || strings.HasPrefix(name, "#") {
			continue
		}
		lines[i] = name + "=" + quote(value)
		written[name] = true
	}
	names := make([]string, 0, len(changes))
	for name := range changes {
		if !written[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > 0 && len(lines) > 0 && lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	for _, name := range names {
		lines = append(lines, name+"="+quote(changes[name]))
	}
	rendered := []byte(strings.Join(lines, "\n") + "\n")
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot write Compose settings %s: %w", path, err)
	}
	_, writeErr := file.WriteAt(rendered, 0)
	if writeErr == nil {
		writeErr = file.Truncate(int64(len(rendered)))
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("cannot write Compose settings %s: %w", path, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("cannot close Compose settings %s: %w", path, closeErr)
	}
	return nil
}

// Single quotes keep Compose from interpolating passwords containing dollar
// signs. Compose accepts backslash-escaped quotes inside a single-quoted value.
func quote(value string) string {
	value = strings.ReplaceAll(value, "'", `\'`)
	value = strings.ReplaceAll(value, "\r", `\r`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return "'" + value + "'"
}
