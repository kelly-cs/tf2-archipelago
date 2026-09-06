//go:build windows

package debugbundle

import (
	"os"
	"path/filepath"
)

// systemCrashDumpDir returns where Windows Error Reporting stores user dumps.
// UserCacheDir is LOCALAPPDATA on Windows, without making this package parse
// process configuration itself.
func systemCrashDumpDir() string {
	local, err := os.UserCacheDir()
	if err != nil || local == "" {
		return ""
	}
	return filepath.Join(local, "CrashDumps")
}
