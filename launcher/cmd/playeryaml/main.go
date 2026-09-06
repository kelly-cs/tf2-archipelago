// Command playeryaml writes the Archipelago player file from the MVM_*
// environment, with the renderer the launcher uses, so the Docker server and
// the exe generate from the same file for the same settings. It runs inside
// the Archipelago image, before generation.
//
// The compose stack names popfiles, because that is what every other part of
// this project calls a mission; the apworld's options take display names. A
// popfile, class or mod this does not know is an error here rather than a
// surprise three hours into an evening.
//
// Usage: playeryaml <archipelago version>
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: playeryaml <archipelago version>")
		os.Exit(2)
	}
	s := settings.ApplyEnv(settings.Defaults())
	if err := checkNames(s); err != nil {
		fmt.Fprintln(os.Stderr, "cannot write the player file:", err)
		os.Exit(1)
	}
	if _, err := settings.CheckRunSelection(s); err != nil {
		fmt.Fprintln(os.Stderr, "cannot write the player file:", err)
		os.Exit(1)
	}
	fmt.Print(settings.PlayerYAML(s, os.Args[1]))
}

// checkNames refuses what the renderer would silently drop: the exe has a
// screen to say so, and a container has only this line.
func checkNames(s settings.Settings) error {
	known := gamedata.ServerModKeys()
	for _, key := range s.SrcdsMods {
		if !slices.Contains(known, key) {
			return fmt.Errorf("SRCDS_MODS: %q is not a server mod of this game; name one of %s", key, strings.Join(known, ", "))
		}
	}
	// Every mission the tables know, whatever mod it needs: a mission
	// excluded on a server without its mod is out of the pool either way.
	popfiles := map[string]bool{}
	for _, mission := range gamedata.Missions {
		popfiles[mission.PopFile] = true
	}
	for _, popfile := range s.MvmExcludedMissions {
		if !popfiles[popfile] {
			return fmt.Errorf("MVM_EXCLUDED_MISSIONS: %q is not a mission of this game", popfile)
		}
	}
	if s.MvmStartMission != "" && s.MvmStartMission != "random" && !popfiles[s.MvmStartMission] {
		return fmt.Errorf("MVM_START_MISSION: %q is not a mission of this game", s.MvmStartMission)
	}
	if s.MvmStartClass != "" && s.MvmStartClass != "random" {
		names := make([]string, 0, len(gamedata.Classes))
		for _, class := range gamedata.Classes {
			names = append(names, class.Name)
		}
		if !slices.ContainsFunc(names, func(n string) bool { return strings.EqualFold(n, s.MvmStartClass) }) {
			return fmt.Errorf("MVM_START_CLASS: %q is not a class; name one of %s", s.MvmStartClass, strings.Join(names, ", "))
		}
	}
	return nil
}
