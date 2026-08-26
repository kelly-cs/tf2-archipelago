// Package debugbundle collects everything somebody would ask a play-tester for
// into one zip: the launcher's log and the one before it, SourceMod's error
// logs, the game server's console log, what the bridge says about the run, the
// player file, and the settings with the passwords taken out.
//
// It exists because the alternative is asking a player to find five files in
// three folders, and because a zip posted in a chat channel is the shortest
// path from "it broke" to a diagnosis.
package debugbundle

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/session"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/srcdsconfig"
)

// fileBytesMax caps one file inside the zip. A console log from a long evening
// runs to hundreds of megabytes, and the end of it is the part that matters.
const fileBytesMax = 8 << 20

// bridgeTimeout bounds the one request this package makes. The bridge is on
// loopback, and a bundle must not hang on a bridge that is not running.
const bridgeTimeout = 3 * time.Second

// Write builds the zip next to the game files and returns its path. stamp names
// it, so two bundles from one evening do not overwrite each other.
func Write(s settings.Settings, versions map[string]string, stamp time.Time) (string, error) {
	name := fmt.Sprintf("debug-logs-%s.zip", stamp.Format("2006-01-02-150405"))
	path := filepath.Join(s.InstallRoot, name)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("cannot create %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	archive := zip.NewWriter(file)
	game := filepath.Join(s.InstallRoot, "tf-dedicated", "tf")

	add(archive, "summary.txt", strings.NewReader(summary(s, versions, stamp)))
	add(archive, "config.json", strings.NewReader(redactedSettings(s)))
	add(archive, "bridge.json", strings.NewReader(bridgeState()))
	// The run before this one as well: a player who hits a bug restarts the
	// server and then goes looking for the button, so the run that broke is
	// usually not the run the bundle is made from.
	copyIn(archive, apruntime.LogFileName, filepath.Join(s.InstallRoot, apruntime.LogFileName))
	copyIn(archive, apruntime.LogPreviousName, filepath.Join(s.InstallRoot, apruntime.LogPreviousName))
	copyIn(archive, "tf2.yaml", filepath.Join(s.InstallRoot, settings.PlayerFileName))
	copyIn(archive, apruntime.ConsoleLogName, filepath.Join(game, apruntime.ConsoleLogName))
	copyIn(archive, apruntime.ConsolePreviousName, filepath.Join(game, apruntime.ConsolePreviousName))
	copyIn(archive, "server.cfg", filepath.Join(game, "cfg", "server.cfg"))
	// The plugin's own config, which is written once and then belongs to the
	// server. Two bundles were read without knowing that this file still said
	// tf2ap_debug 0, which is why they held nothing worth reading.
	copyIn(archive, "tf2_archipelago.cfg", filepath.Join(game, "cfg", "sourcemod", "tf2_archipelago.cfg"))

	// The crash dumps, newest last. srcds runs under Breakpad and writes one
	// per crash, and a crash that leaves no line in any log leaves one of
	// these: it is the only file that names the function the server died in.
	for _, dump := range newestCrashDumps(game, 3) {
		copyIn(archive, filepath.Join("crashes", filepath.Base(dump)), dump)
	}

	// Every SourceMod log, because the error log names the plugin and the plain
	// log holds what happened around it.
	logs := filepath.Join(game, "addons", "sourcemod", "logs")
	for _, entry := range newestLogs(logs, 6) {
		copyIn(archive, filepath.Join("sourcemod", entry), filepath.Join(logs, entry))
	}

	if err := archive.Close(); err != nil {
		return "", fmt.Errorf("cannot finish %s: %w", path, err)
	}
	return path, nil
}

// bridgeURL is a variable so a test can point it at a bridge of its own.
var bridgeURL = session.BridgeURL

// bridgeState is what the bridge says about the run: whether it is connected,
// what it has checked, and which missions it believes are unlocked.
//
// It is fetched rather than read off disk, so it only answers while the server
// is up. A bundle made after everything was stopped carries the refusal, which
// is itself worth reading: it says the state came from nowhere rather than
// leaving a reader to assume the run was empty.
func bridgeState() string {
	ctx, cancel := context.WithTimeout(context.Background(), bridgeTimeout)
	defer cancel()

	snapshot, err := session.Fetch(ctx, bridgeURL)
	if err != nil {
		return fmt.Sprintf("{\n  \"error\": %q\n}\n", err.Error())
	}
	body, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\n  \"error\": %q\n}\n", err.Error())
	}
	return string(body) + "\n"
}

// summary is the first thing to read: what was installed, what the run is, and
// where it was pointed.
func summary(s settings.Settings, versions map[string]string, stamp time.Time) string {
	var b strings.Builder
	fmt.Fprintf(&b, "collected     %s\n", stamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "install root  %s\n", s.InstallRoot)
	fmt.Fprintf(&b, "room          %s (tls=%v), slot %s\n",
		settings.Room{Host: s.APHost, Port: s.APPort}, s.APTls, s.APSlotName)
	fmt.Fprintf(&b, "server        %q on port %d, reach=%s\n", s.SrcdsHostname, s.SrcdsPort, s.SrcdsReach)
	fmt.Fprintf(&b, "start mission %s\n", s.SrcdsStartMission)
	fmt.Fprintf(&b, "bots          on=%v, fill RED to %d\n", s.SrcdsBots, s.SrcdsBotTeamSize)
	fmt.Fprintf(&b, "run           %d missions, %s and harder, goal %s, missionsanity %d%%, death link %v\n",
		s.MvmMissionCount, s.MvmDifficulty, s.MvmGoal, s.MvmMissionsanityPct, s.MvmDeathLink)

	names := make([]string, 0, len(versions))
	for name := range versions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "%-13s %s\n", name, versions[name])
	}
	b.WriteString(configDrift(s))
	b.WriteString(scanLogs(
		filepath.Join(s.InstallRoot, apruntime.LogFileName),
		filepath.Join(s.InstallRoot, apruntime.LogPreviousName),
		filepath.Join(s.InstallRoot, "tf-dedicated", "tf", apruntime.ConsoleLogName),
	))

	b.WriteString("\nPasswords are not in this file, and not in config.json either.\n")
	return b.String()
}

/* configDrift says whether the server.cfg on disk is the one these settings
 * render.
 *
 * A bundle once held a config.json that blacklisted the Spy beside a server.cfg
 * that blacklisted nothing, and the disagreement was the whole bug. Reading it
 * meant knowing to compare the two files by eye. The summary should say it.
 */
func configDrift(s settings.Settings) string {
	target := filepath.Join(s.InstallRoot, "tf-dedicated", "tf", "cfg", "server.cfg")
	onDisk, err := os.ReadFile(target)
	if err != nil {
		return "\nserver.cfg     not readable, so the server is running on something unknown\n"
	}
	wanted, err := srcdsconfig.RenderServerCfg(s)
	if err != nil {
		return ""
	}
	if string(onDisk) == wanted {
		return "\nserver.cfg     matches these settings\n"
	}
	return "\nserver.cfg     DIFFERS from these settings: the running server is not\n" +
		"               playing what config.json in this bundle says. Compare the two.\n"
}

// redactedSettings renders the settings with every secret replaced.
//
// This zip is made to be posted in a chat channel. An RCON password in it is a
// stranger's admin console, and a room password is somebody else's multiworld.
func redactedSettings(s settings.Settings) string {
	const hidden = "(removed from the bundle)"
	if s.SrcdsRconPw != "" {
		s.SrcdsRconPw = hidden
	}
	if s.SrcdsPw != "" {
		s.SrcdsPw = hidden
	}
	if s.APPassword != "" {
		s.APPassword = hidden
	}
	if s.SrcdsToken != "" && s.SrcdsToken != "0" {
		s.SrcdsToken = hidden
	}
	body, err := settings.Render(s)
	if err != nil {
		return fmt.Sprintf("cannot render the settings: %v\n", err)
	}
	return body
}

/* newestCrashDumps finds the newest minidumps srcds left behind.
 *
 * The game writes them beside the executable or under tf/, depending on which
 * part of it crashed, and it never cleans them up. Newest by modification time
 * rather than by name: the names carry a crash id, not a date.
 */
func newestCrashDumps(gameDir string, limit int) []string {
	var found []string
	for _, dir := range []string{gameDir, filepath.Dir(gameDir)} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".mdmp") {
				continue
			}
			found = append(found, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Slice(found, func(i, j int) bool {
		return modTime(found[i]).Before(modTime(found[j]))
	})
	if len(found) > limit {
		found = found[len(found)-limit:]
	}
	return found
}

func modTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// newestLogs returns up to limit file names from dir, newest last.
func newestLogs(dir string, limit int) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) > limit {
		names = names[len(names)-limit:]
	}
	return names
}

func add(archive *zip.Writer, name string, body io.Reader) {
	writer, err := archive.Create(filepath.ToSlash(name))
	if err != nil {
		return
	}
	_, _ = io.Copy(writer, body)
}

// copyIn adds a file if it is there, keeping the last fileBytesMax of it. A
// missing file is normal: not every run has a console log or a SourceMod error.
func copyIn(archive *zip.Writer, name, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return
	}
	if info.Size() > fileBytesMax {
		if _, err := file.Seek(info.Size()-fileBytesMax, 0); err != nil {
			return
		}
		name = strings.TrimSuffix(name, filepath.Ext(name)) + "-tail" + filepath.Ext(name)
	}
	add(archive, name, file)
}
