package gamedata

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Lifting pieces of the plugin out by name, so a driver can be built from the
// source that ships rather than from numbers copied next to it.

// defineFrom returns the whole `#define <name> <value>` line. The line, not the
// value: the driver compiles it as written, so a define that changes shape is
// still the plugin's own text.
func defineFrom(t *testing.T, path, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?m)^#define\s+` + regexp.QuoteMeta(name) + `\s+(.+?)\s*$`)
	match := pattern.FindStringSubmatch(read(t, path))
	if match == nil {
		t.Fatalf("%s has no #define %s", path, name)
	}
	return match[0]
}

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

/*
	sourceFunctionWithSignature is the whole function, signature included

sourceFunction beside it returns the body for a test that reads. This one
returns text that compiles, because the driver has to declare the function
before it can call it.

The brace counting is the same, and it is honest for this file: SourcePawn has
no raw strings, and a brace inside a string literal or a comment inside one of
these functions would break it. If that ever happens it breaks loudly, at
compile time, in the driver.
*/
func sourceFunctionWithSignature(t *testing.T, path, signature string) string {
	t.Helper()
	text := read(t, path)
	start := strings.Index(text, signature)
	if start < 0 {
		t.Fatalf("%s has no %s", path, signature)
	}
	open := strings.IndexByte(text[start:], '{')
	if open < 0 {
		t.Fatalf("%s has no body for %s", path, signature)
	}
	open += start

	depth := 0
	for i := open; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	t.Fatalf("%s: %s is not closed", path, signature)
	return ""
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
