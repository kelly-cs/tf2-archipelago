//go:build !windows

package generate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// candidateDirs lists a bounded set of places where the official Linux
// tarball or AppImage commonly ends up. It deliberately does not recursively
// search the home directory: generation should never become a slow disk crawl.
func candidateDirs() []string {
	var dirs []string
	if generator, err := exec.LookPath("ArchipelagoGenerate"); err == nil {
		dirs = append(dirs, filepath.Dir(generator))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, "Applications", "Archipelago"),
			filepath.Join(home, ".local", "opt", "Archipelago"),
			filepath.Join(home, "Archipelago"),
			filepath.Join(home, "Downloads", "Archipelago"),
		)
		if runtime.GOOS == "linux" {
			dirs = append(dirs, appImagesIn(
				filepath.Join(home, "Applications"),
				filepath.Join(home, ".local", "bin"),
				filepath.Join(home, "Downloads"),
			)...)
		}
	}
	dirs = append(dirs, "/opt/Archipelago", "/usr/local/lib/Archipelago", "/ap")
	if runtime.GOOS == "linux" {
		dirs = append(dirs, appImagesIn("/opt", "/usr/local/bin")...)
	}
	return uniquePaths(dirs)
}

func appImagesIn(dirs ...string) []string {
	var found []string
	for _, dir := range dirs {
		matches, _ := filepath.Glob(filepath.Join(dir, "Archipelago*.AppImage"))
		for _, match := range matches {
			info, err := os.Stat(match)
			if err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
				found = append(found, match)
			}
		}
	}
	// Versioned AppImages sort oldest first. Prefer the newest-looking name.
	sort.Sort(sort.Reverse(sort.StringSlice(found)))
	return found
}

func generatorPath(appDir string) string {
	if isAppImage(appDir) {
		info, err := os.Stat(appDir)
		if err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			return appDir
		}
		// Return a path that cannot exist under a non-runnable image, so
		// FindApp rejects it as cleanly as a directory with no generator.
		return filepath.Join(appDir, "ArchipelagoGenerate")
	}
	if p := filepath.Join(appDir, "ArchipelagoGenerate"); exists(p) {
		return p
	}
	return filepath.Join(appDir, "Generate.py")
}

// generatorCommand runs a frozen build as it is, and a source checkout through
// python: a .py is not executable on its own.
func generatorCommand(exe string) (string, []string) {
	if isAppImage(exe) {
		// The official AppImage runs ArchipelagoLauncher, which dispatches the
		// Generate component. -- separates its arguments from the launcher.
		return exe, []string{"Generate", "--"}
	}
	if strings.HasSuffix(exe, ".py") {
		return "python3", []string{exe}
	}
	return exe, nil
}

func isAppImage(path string) bool {
	return runtime.GOOS == "linux" && strings.EqualFold(filepath.Ext(path), ".AppImage")
}

func standaloneApp(path string) bool { return isAppImage(path) }

func apworldInstallDir(appLocation string) (string, error) {
	if !isAppImage(appLocation) {
		return filepath.Join(appLocation, "custom_worlds"), nil
	}
	data := os.Getenv("XDG_DATA_HOME") //nolint:forbidigo // this is the path Archipelago itself follows
	if data == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot locate Archipelago's user data: %w", err)
		}
		data = filepath.Join(home, ".local", "share")
	}
	// An AppImage is mounted read-only. Archipelago's Utils.user_path chooses
	// this writable directory and loads third-party worlds from worlds/.
	return filepath.Join(data, "Archipelago", "worlds"), nil
}
