// Command playeryaml writes the Archipelago player file from the MVM_*
// environment. It runs inside the Archipelago image, before generation.
//
// The compose stack sets popfile names, because that is what every other part
// of this project calls a mission. The apworld's options take the mission's
// display name. This does the translation, from the same export the apworld
// reads, so the two cannot drift. A popfile this does not know is an error here
// rather than a surprise three hours into an evening.
//
// Usage: playeryaml <data directory> <archipelago version>
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: playeryaml <data directory> <archipelago version>")
		os.Exit(2)
	}
	out, err := build(os.Args[1], os.Args[2], settingsFromEnv())
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot write the player file:", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprint(os.Stdout, out)
}

// settingsFromEnv reads the run's shape. Empty and unset are the same thing:
// compose passes every variable through with no default, and the fallbacks
// live in build.
func settingsFromEnv() func(name string) string {
	return func(name string) string { return strings.TrimSpace(os.Getenv(name)) }
}
