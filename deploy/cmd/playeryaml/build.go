package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const random = "random"

type meta struct {
	Game       string                  `json:"game"`
	Classes    []struct{ Name string } `json:"classes"`
	ServerMods []struct{ Key string }  `json:"server_mods"`
}

type missions struct {
	Missions []struct {
		Name    string `json:"name"`
		PopFile string `json:"pop_file"`
	} `json:"missions"`
}

func readJSON(path string, into any) error {
	body, err := os.ReadFile(path) //nolint:gosec // the data directory is an argument
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, into); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// shape is the run as the environment asked for it, translated and checked.
type shape struct {
	game, slotName, archipelagoVersion string
	startMission, startClass           string
	excluded, serverMods               []string
	read                               func(string) string
}

// build renders the player file. read answers a setting by name, empty when
// it is not set, which falls through to the same defaults the apworld has.
func build(data, archipelagoVersion string, read func(string) string) (string, error) {
	var m meta
	if err := readJSON(filepath.Join(data, "meta.json"), &m); err != nil {
		return "", err
	}
	var ms missions
	if err := readJSON(filepath.Join(data, "missions.json"), &ms); err != nil {
		return "", err
	}
	run, err := resolve(m, ms, archipelagoVersion, read)
	if err != nil {
		return "", err
	}
	return render(run), nil
}

func or(read func(string) string, name, fallback string) string {
	if v := read(name); v != "" {
		return v
	}
	return fallback
}

// listed splits a comma-separated setting, dropping the empties .env leaves.
func listed(read func(string) string, name string) []string {
	var out []string
	for item := range strings.SplitSeq(read(name), ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func resolve(m meta, ms missions, archipelagoVersion string, read func(string) string) (shape, error) {
	byPopfile := make(map[string]string, len(ms.Missions))
	for _, mission := range ms.Missions {
		byPopfile[mission.PopFile] = mission.Name
	}
	missionName := func(popfile, variable string) (string, error) {
		name, ok := byPopfile[popfile]
		if !ok {
			return "", fmt.Errorf("%s: %q is not a mission of this game", variable, popfile)
		}
		return name, nil
	}

	run := shape{game: m.Game, slotName: or(read, "AP_SLOT_NAME", "tf2"), archipelagoVersion: archipelagoVersion, read: read}
	for _, popfile := range listed(read, "MVM_EXCLUDED_MISSIONS") {
		name, err := missionName(popfile, "MVM_EXCLUDED_MISSIONS")
		if err != nil {
			return shape{}, err
		}
		run.excluded = append(run.excluded, name)
	}

	run.startMission = or(read, "MVM_START_MISSION", random)
	if run.startMission != random {
		name, err := missionName(run.startMission, "MVM_START_MISSION")
		if err != nil {
			return shape{}, err
		}
		run.startMission = name
	}

	run.startClass = or(read, "MVM_START_CLASS", random)
	if run.startClass != random {
		names := make([]string, 0, len(m.Classes))
		for _, c := range m.Classes {
			names = append(names, c.Name)
		}
		if !slices.Contains(names, run.startClass) {
			slices.Sort(names)
			return shape{}, fmt.Errorf("MVM_START_CLASS: %q is not a class; name one of %s", run.startClass, strings.Join(names, ", "))
		}
	}

	known := make([]string, 0, len(m.ServerMods))
	for _, mod := range m.ServerMods {
		known = append(known, mod.Key)
	}
	for _, key := range listed(read, "SRCDS_MODS") {
		if !slices.Contains(known, key) {
			slices.Sort(known)
			return shape{}, fmt.Errorf("SRCDS_MODS: %q is not a server mod of this game; name one of %s", key, strings.Join(known, ", "))
		}
		run.serverMods = append(run.serverMods, key)
	}
	return run, nil
}

func render(run shape) string {
	var b strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	or := func(name, fallback string) string { return or(run.read, name, fallback) }
	line("name: %s", yamlString(run.slotName))
	line("game: %s", yamlString(run.game))
	line("requires:")
	line("  version: %s", run.archipelagoVersion)
	line("%s:", yamlString(run.game))
	line("  mission_count: %s", or("MVM_MISSION_COUNT", "8"))
	line("  difficulty_pool: %s", or("MVM_DIFFICULTY", "intermediate"))
	line("  goal: %s", or("MVM_GOAL", "final_boss"))
	line("  missionsanity_percentage: %s", or("MVM_MISSIONSANITY_PERCENTAGE", "80"))
	line("  death_link: %s", or("MVM_DEATH_LINK", "false"))
	line("  trap_percentage: %s", or("MVM_TRAP_PERCENTAGE", "0"))
	line("  start_mission: %s", yamlString(run.startMission))
	line("  start_class: %s", yamlString(run.startClass))
	yamlList(line, "excluded_missions", run.excluded)
	line("  community_missions: %s", or("MVM_COMMUNITY_MISSIONS", "true"))
	yamlList(line, "server_mods", run.serverMods)
	return b.String()
}

func yamlList(line func(string, ...any), key string, values []string) {
	if len(values) == 0 {
		line("  %s: []", key)
		return
	}
	line("  %s:", key)
	for _, v := range values {
		line("    - %s", yamlString(v))
	}
}

func yamlString(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}
