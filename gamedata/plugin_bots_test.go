package gamedata

import (
	"os"
	"strings"
	"testing"
)

func pluginSourceFunction(t *testing.T, path, signature string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	start := strings.Index(text, signature)
	if start < 0 {
		t.Fatalf("%s has no %s", path, signature)
	}
	open := strings.IndexByte(text[start:], '{') + start
	depth := 0
	for index := open; index < len(text); index++ {
		switch text[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : index+1]
			}
		}
	}
	t.Fatalf("%s has an unterminated %s", path, signature)
	return ""
}

func TestRobotHealthScaleIsRelayedAfterConfigs(t *testing.T) {
	bots := "../plugin/scripting/tf2_archipelago/bots.inc"
	source, err := os.ReadFile(bots)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		`CreateConVar("tf2ap_robot_health_scale", "1.0"`,
		"FCVAR_DONTRECORD",
		`FindConVar("sm_redbots_manager_blu_health_scale")`,
		"g_CvarDefenderRobotHealthScale.FloatValue = wanted",
	} {
		if !strings.Contains(string(source), required) {
			t.Fatalf("robot-health config relay has no %s", required)
		}
	}

	plugin := pluginSourceFunction(t, "../plugin/scripting/tf2_archipelago.sp",
		"public void OnConfigsExecuted()")
	if !strings.Contains(plugin, "Bots_OnConfigsExecuted()") {
		t.Fatal("robot health is not reasserted after SourceMod configs execute")
	}
}
