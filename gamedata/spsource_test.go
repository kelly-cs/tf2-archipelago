package gamedata

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

/*
Reading the plugin's own numbers, so a driver never carries a copy of one.

There used to be more here: a function that cut a whole function body out by
counting braces, so a driver could paste it. weapon_buffs_math.inc replaced
that. The driver includes the file the plugin includes, so the only thing still
lifted out of the source is a constant's value, which a test needs in Go to work
out what the answer should be.

Nothing here asserts. Reading a #define and checking the behaviour it produces
is a different thing from reading a #define and checking it is still spelled the
same way, and it is the first one these serve.
*/

// defineValue is the right-hand side of a define, for a test that needs the
// number rather than the line.
func defineValue(t *testing.T, path, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?m)^#define\s+` + regexp.QuoteMeta(name) + `\s+(.+?)\s*$`)
	match := pattern.FindStringSubmatch(read(t, path))
	if match == nil {
		t.Fatalf("%s has no #define %s", path, name)
	}
	return match[1]
}

func intDefine(t *testing.T, path, name string) int {
	t.Helper()
	value := defineValue(t, path, name)
	n, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("%s is #define %s %s, which is not a whole number", path, name, value)
	}
	return n
}

func floatDefine(t *testing.T, path, name string) float32 {
	t.Helper()
	value := defineValue(t, path, name)
	f, err := strconv.ParseFloat(value, 32)
	if err != nil {
		t.Fatalf("%s is #define %s %s, which is not a number", path, name, value)
	}
	return float32(f)
}

// bitsToFloat reads a cell back as the float the driver printed. printnum takes
// an int, so a float is printed as view_as<int> and reinterpreted here: the
// exact bits, never a rounded decimal that would hide a difference in the last
// place.
func bitsToFloat(cell int32) float32 {
	return math.Float32frombits(uint32(cell)) //nolint:gosec // a cell is 32 bits either way
}

// read is the file, or a failure naming it. The plugin sources are in the
// repository, so a missing one is a moved file rather than a machine without
// something installed.
func read(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// weaponDefinitions are the item definition indexes the generated table holds,
// read out of the table itself so a weapon added to gamedata is asked about
// here without anybody editing this file.
func weaponDefinitions(t *testing.T) []int {
	t.Helper()
	text := read(t, dataSource)
	start := strings.Index(text, "int g_WeaponByDefinition[][2] = {")
	if start < 0 {
		t.Fatalf("%s has no g_WeaponByDefinition", dataSource)
	}
	end := strings.Index(text[start:], "\n};")
	if end < 0 {
		t.Fatalf("%s: g_WeaponByDefinition is not closed", dataSource)
	}

	rows := regexp.MustCompile(`\{\s*(-?\d+),\s*(-?\d+)\s*\}`).FindAllStringSubmatch(text[start:start+end], -1)
	out := make([]int, 0, len(rows))
	for _, row := range rows {
		definition, err := strconv.Atoi(row[1])
		if err != nil {
			t.Fatalf("g_WeaponByDefinition holds %q", row[1])
		}
		out = append(out, definition)
	}
	return out
}
