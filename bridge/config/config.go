// Package config reads the bridge's configuration from the environment.
//
// Environment only, no config file: the bridge runs next to a compose file that
// already owns every other setting in this stack.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is everything the bridge needs to start. Nothing here changes while it runs.
type Config struct {
	// ArchipelagoURL is ws:// or wss://, host and port included.
	ArchipelagoURL string
	SlotName       string
	Password       string

	// Listen is the plugin-facing address. Loopback always: srcds and the
	// bridge share a network namespace and nothing else may reach it.
	Listen string

	// MetricsListen serves Prometheus metrics, and only those. It is separate
	// from Listen because a scraper is on another machine while the plugin's API
	// stays on loopback. Empty turns it off.
	MetricsListen string

	// GameQueryAddr is the game server's own UDP port, asked A2S_INFO on a scrape
	// to report how many people are connected. Its name on the docker network,
	// not loopback: srcds does not answer a query sent to 127.0.0.1. Empty leaves
	// the player metrics out.
	GameQueryAddr string

	StatePath string

	// PollTimeout is how long GET /grants is held open; the plugin's own timeout must be longer.
	PollTimeout time.Duration

	// TestMode plays without Archipelago: the bridge serves a multiworld of one
	// on loopback and dials that instead of ArchipelagoURL. For trying the
	// stack out, and for play-testing without a room and a seed.
	TestMode bool
	TestRun  TestRun
}

// TestRun is the seed shape requested by the Docker settings when test mode is on.
type TestRun struct {
	MissionCount, ModifierMin, ModifierMax                                               int
	Difficulty, Goal, StartMission, StartClass                                           string
	Excluded, ServerMods                                                                 []string
	CommunityMissions                                                                    bool
	MissionModifiers, VictoryCaches, MilestoneChecks, Giantsanity, Tanksanity, DeathLink bool
}

// Load reads the environment. Every value has a default that works inside the
// compose file, and bad ones are refused here rather than at first use.
func Load() (Config, error) {
	host := env("AP_HOST", "archipelago")
	port := env("AP_PORT", "38281")
	scheme := "ws"
	tls, err := boolEnv("AP_TLS")
	if err != nil {
		return Config{}, err
	}
	if tls {
		scheme = "wss"
	}

	timeout, err := durationEnv("BRIDGE_POLL_TIMEOUT", 25*time.Second)
	if err != nil {
		return Config{}, err
	}

	testMode, err := boolEnv("TF2AP_TEST_MODE")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ArchipelagoURL: (&url.URL{Scheme: scheme, Host: host + ":" + port}).String(),
		SlotName:       env("AP_SLOT_NAME", "tf2"),
		Password:       os.Getenv("AP_PASSWORD"),
		Listen:         env("BRIDGE_LISTEN", "127.0.0.1:24680"),
		MetricsListen:  os.Getenv("BRIDGE_METRICS_LISTEN"),
		GameQueryAddr:  env("BRIDGE_GAME_QUERY", "srcds:27015"),
		StatePath:      env("BRIDGE_STATE", "/data/bridge.json"),
		PollTimeout:    timeout,
		TestMode:       testMode,
	}
	if cfg.SlotName == "" {
		return Config{}, fmt.Errorf("AP_SLOT_NAME is empty")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("AP_PORT %q is not a number", port)
	}
	if cfg.TestMode {
		if cfg.TestRun, err = loadTestRun(); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

func loadTestRun() (run TestRun, err error) {
	if run.MissionCount, err = intEnv("MVM_MISSION_COUNT", 8); err != nil {
		return TestRun{}, err
	}
	if run.ModifierMin, err = intEnv("MVM_MINIMUM_MISSION_MODIFIERS", 1); err != nil {
		return TestRun{}, err
	}
	if run.ModifierMax, err = intEnv("MVM_MAXIMUM_MISSION_MODIFIERS", 2); err != nil {
		return TestRun{}, err
	}
	if run.MissionModifiers, err = boolEnv("MVM_MISSION_MODIFIERS"); err != nil {
		return TestRun{}, err
	}
	if run.VictoryCaches, err = boolEnv("MVM_VICTORY_CACHES"); err != nil {
		return TestRun{}, err
	}
	if run.MilestoneChecks, err = boolEnv("MVM_MILESTONE_CHECKS"); err != nil {
		return TestRun{}, err
	}
	if run.Giantsanity, err = boolEnv("MVM_GIANTSANITY"); err != nil {
		return TestRun{}, err
	}
	if run.Tanksanity, err = boolEnv("MVM_TANKSANITY"); err != nil {
		return TestRun{}, err
	}
	if run.DeathLink, err = boolEnv("MVM_DEATH_LINK"); err != nil {
		return TestRun{}, err
	}
	if run.CommunityMissions, err = boolEnvDefault("MVM_COMMUNITY_MISSIONS", true); err != nil {
		return TestRun{}, err
	}
	if run.MissionModifiers && (run.ModifierMin < 0 || run.ModifierMax > 3 || run.ModifierMin > run.ModifierMax) {
		return TestRun{}, fmt.Errorf("mission modifier bounds must be within 0..3 and minimum <= maximum")
	}
	run.Difficulty = env("MVM_DIFFICULTY", "intermediate")
	run.Goal = env("MVM_GOAL", "final_boss")
	run.StartMission = os.Getenv("MVM_START_MISSION")
	run.StartClass = os.Getenv("MVM_START_CLASS")
	run.Excluded = csvEnv("MVM_EXCLUDED_MISSIONS")
	run.ServerMods = csvEnv("SRCDS_MODS")
	return run, nil
}

func csvEnv(key string) []string {
	var values []string
	for name := range strings.SplitSeq(os.Getenv(key), ",") {
		if name = strings.TrimSpace(name); name != "" {
			values = append(values, name)
		}
	}
	return values
}

func intEnv(key string, fallback int) (int, error) {
	value := env(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s %q is not an integer", key, value)
	}
	return parsed, nil
}

func env(key, fallback string) string {
	if value, set := os.LookupEnv(key); set && value != "" {
		return value
	}
	return fallback
}

func boolEnv(key string) (bool, error) {
	value, set := os.LookupEnv(key)
	if !set || value == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s %q is not a boolean", key, value)
	}
	return parsed, nil
}

func boolEnvDefault(key string, fallback bool) (bool, error) {
	value, set := os.LookupEnv(key)
	if !set || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s %q is not a boolean", key, value)
	}
	return parsed, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value, set := os.LookupEnv(key)
	if !set || value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s %q is not a duration", key, value)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive, got %s", key, parsed)
	}
	return parsed, nil
}
