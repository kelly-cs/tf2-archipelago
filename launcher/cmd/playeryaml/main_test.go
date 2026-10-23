package main

import (
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestUnknownNamesAreRefused(t *testing.T) {
	cases := map[string]string{
		"MVM_EXCLUDED_MISSIONS": "mvm_nowhere",
		"MVM_START_MISSION":     "mvm_nowhere",
		"MVM_START_CLASS":       "Nobody",
		"SRCDS_MODS":            "nothing",
	}
	for name, value := range cases {
		t.Setenv(name, value)
		err := checkNames(settings.ApplyEnv(settings.Defaults()))
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s=%s: err = %v", name, value, err)
		}
		t.Setenv(name, "")
	}
	if err := checkNames(settings.ApplyEnv(settings.Defaults())); err != nil {
		t.Errorf("the defaults were refused: %v", err)
	}
}
