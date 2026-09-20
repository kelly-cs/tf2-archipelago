// Command waveprobe drives the disposable SRCDS wave smoke test over RCON.
// It never advances a wave itself: the test plugin waits for the game's
// mvm_wave_complete event while defeating each spawn after a seeded delay.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/rcon"
)

type options struct {
	address     string
	mission     string
	mode        string
	seed        int
	speed       int
	timeout     time.Duration
	gameTimeout time.Duration
	loadWait    time.Duration
	shard       int
	shards      int
	startWave   int
	endWave     int
	failFast    bool
	includeSig  bool
	plan        bool
}

type probeStatus struct {
	State      string
	Map        string
	Pop        string
	Max        int
	GameWave   int
	Expected   int
	Observed   int
	Bots       int
	Tanks      int
	DefTeam    int
	PlayerTeam int
	EnemyTeam  int
	Elapsed    float64
}

type result struct {
	Mission     string  `json:"mission"`
	Map         string  `json:"map"`
	Mode        string  `json:"mode"`
	Wave        int     `json:"wave"`
	Seed        int     `json:"seed"`
	State       string  `json:"state"`
	Bots        int     `json:"bots"`
	Tanks       int     `json:"tanks"`
	Seconds     float64 `json:"wall_seconds"`
	GameSeconds float64 `json:"game_seconds,omitempty"`
	Error       string  `json:"error,omitempty"`
}

func main() {
	var opt options
	flag.StringVar(&opt.address, "rcon", "127.0.0.1:27035", "isolated server RCON address")
	flag.StringVar(&opt.mission, "mission", "all", "population file name or all")
	flag.StringVar(&opt.mode, "mode", "both", "normal, surge, or both")
	flag.IntVar(&opt.seed, "seed", 1, "seed for each spawn's 15-25 game-second lifetime")
	flag.IntVar(&opt.speed, "speed", 10, "isolated server host_timescale")
	flag.DurationVar(&opt.timeout, "timeout", 2*time.Minute, "wall-clock timeout per wave")
	flag.DurationVar(&opt.gameTimeout, "game-timeout", 15*time.Minute, "game-time limit per wave")
	flag.DurationVar(&opt.loadWait, "load-timeout", 90*time.Second, "wall-clock timeout for a map or mission load")
	flag.IntVar(&opt.shard, "shard", 0, "zero-based mission shard")
	flag.IntVar(&opt.shards, "shards", 1, "number of mission shards")
	flag.IntVar(&opt.startWave, "start-wave", 1, "first wave to test (single mission only)")
	flag.IntVar(&opt.endWave, "end-wave", 0, "last wave to test (0 means mission end)")
	flag.BoolVar(&opt.failFast, "fail-fast", false, "stop after the first failed wave")
	flag.BoolVar(&opt.includeSig, "include-sigmod", true, "test SigMod missions")
	flag.BoolVar(&opt.plan, "plan", false, "print planned wave tests without connecting to a server")
	flag.Parse()
	if err := run(opt); err != nil {
		fmt.Fprintln(os.Stderr, "waveprobe:", err)
		os.Exit(1)
	}
}

func run(opt options) error {
	if opt.shards < 1 || opt.shard < 0 || opt.shard >= opt.shards {
		return errors.New("shard must be in [0, shards)")
	}
	if opt.speed < 1 || opt.speed > 20 {
		return errors.New("speed must be between 1 and 20")
	}
	if opt.mission == "all" && (opt.startWave != 1 || opt.endWave != 0) {
		return errors.New("wave range requires one named mission")
	}
	modes := []string{opt.mode}
	if opt.mode == "both" {
		modes = []string{"normal", "surge"}
	} else if opt.mode != "normal" && opt.mode != "surge" {
		return errors.New("mode must be normal, surge, or both")
	}
	missions, err := selectMissions(opt)
	if err != nil {
		return err
	}
	if opt.plan {
		return writePlan(missions, modes)
	}
	if opt.mission != "all" && strings.Contains(opt.mission, "_rev_") {
		return errors.New("reverse MvM needs a BLU objective simulator; kill-only waveprobe cannot validate it")
	}
	password := os.Getenv("WAVEPROBE_RCONPW")
	if password == "" {
		return errors.New("WAVEPROBE_RCONPW is empty")
	}
	server := server{address: opt.address, password: password}
	if err := server.prepare(opt); err != nil {
		return err
	}
	return runMissions(server, opt, missions, modes)
}

func writePlan(missions []gamedata.Mission, modes []string) error {
	for _, mission := range missions {
		played, ok := gamedata.MapByID(mission.Map)
		if !ok {
			return fmt.Errorf("unknown map for %s", mission.PopFile)
		}
		for _, mode := range modes {
			for wave := 1; wave <= int(mission.Waves); wave++ {
				state := "planned"
				if strings.Contains(mission.PopFile, "_rev_") {
					state = "unsupported_reverse"
				}
				writeResult(result{Mission: mission.PopFile, Map: played.Name,
					Mode: mode, Wave: wave, State: state})
			}
		}
	}
	return nil
}

func (s server) prepare(opt options) error {
	if err := s.await(3*time.Minute, func(probeStatus) bool { return true }); err != nil {
		return fmt.Errorf("test plugin is unavailable at %s: %w", opt.address, err)
	}
	if _, err := s.exec("sv_cheats 1"); err != nil {
		return err
	}
	// A bot reaching the hatch would test an undefended loss, not population
	// progression. Valve provides this cvar specifically for bot testing.
	if _, err := s.exec("tf_bot_flag_kill_on_touch 1"); err != nil {
		return err
	}
	if _, err := s.exec(fmt.Sprintf("host_timescale %d", opt.speed)); err != nil {
		return err
	}
	if reply, err := s.exec("host_timescale"); err != nil {
		return fmt.Errorf("read host_timescale: %w", err)
	} else if !timescaleMatches(reply, opt.speed) {
		return fmt.Errorf("host_timescale %d did not stick: %q", opt.speed, reply)
	}
	return nil
}

func runMissions(s server, opt options, missions []gamedata.Mission, modes []string) error {
	failures := 0
	for _, mission := range missions {
		if strings.Contains(mission.PopFile, "_rev_") {
			continue
		}
		played, ok := gamedata.MapByID(mission.Map)
		if !ok {
			return fmt.Errorf("unknown map for %s", mission.PopFile)
		}
		for _, mode := range modes {
			if err := s.load(played.Name, mission, mode, opt.loadWait); err != nil {
				writeResult(result{Mission: mission.PopFile, Map: played.Name, Mode: mode,
					State: "load_failed", Error: err.Error()})
				failures++
				if opt.failFast {
					return fmt.Errorf("%d wave tests failed", failures)
				}
				continue
			}
			count, err := s.runWaves(opt, played.Name, mission, mode)
			failures += count
			if err != nil {
				return err
			}
			if failures > 0 && opt.failFast {
				return fmt.Errorf("%d wave tests failed", failures)
			}
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d wave tests failed", failures)
	}
	return nil
}

func (s server) runWaves(opt options, mapName string, mission gamedata.Mission, mode string) (int, error) {
	first := opt.startWave
	last := int(mission.Waves)
	if opt.endWave > 0 {
		last = opt.endWave
	}
	if first > 1 {
		if _, err := s.exec(fmt.Sprintf("tf_mvm_jump_to_wave %d 1", first)); err != nil {
			return 0, err
		}
	}
	failures := 0
	for wave := first; wave <= last; wave++ {
		if err := s.recordWave(opt, mapName, mission, mode, wave); err == nil {
			continue
		}
		failures++
		if opt.failFast || wave == last {
			break
		}
		// The failed wave may still be running. Reload the mission and jump
		// ahead so later waves are tested independently too.
		if err := s.load(mapName, mission, mode, opt.loadWait); err != nil {
			writeResult(result{Mission: mission.PopFile, Map: mapName, Mode: mode,
				State: "load_failed", Error: err.Error()})
			failures++
			break
		}
		if _, err := s.exec(fmt.Sprintf("tf_mvm_jump_to_wave %d 1", wave+1)); err != nil {
			writeResult(result{Mission: mission.PopFile, Map: mapName, Mode: mode,
				State: "load_failed", Error: err.Error()})
			failures++
			break
		}
	}
	return failures, nil
}

func (s server) recordWave(opt options, mapName string, mission gamedata.Mission, mode string, wave int) error {
	started := time.Now()
	status, err := s.testWave(mission, wave, opt.seed, opt.timeout, opt.gameTimeout)
	row := result{Mission: mission.PopFile, Map: mapName, Mode: mode,
		Wave: wave, Seed: opt.seed, State: status.State, Bots: status.Bots,
		Tanks: status.Tanks, Seconds: time.Since(started).Seconds(), GameSeconds: status.Elapsed}
	if err != nil {
		row.State = "failed"
		if errors.Is(err, errWallTimeout) {
			row.State = "inconclusive"
		}
		row.Error = err.Error()
	}
	writeResult(row)
	return err
}

func selectMissions(opt options) ([]gamedata.Mission, error) {
	if opt.mission != "all" {
		mission, ok := gamedata.MissionByPopFile(opt.mission)
		if !ok {
			return nil, fmt.Errorf("unknown mission %q", opt.mission)
		}
		if opt.startWave < 1 || opt.startWave > int(mission.Waves) ||
			(opt.endWave > 0 && (opt.endWave < opt.startWave || opt.endWave > int(mission.Waves))) {
			return nil, errors.New("wave range is outside the mission")
		}
		return []gamedata.Mission{mission}, nil
	}
	var selected []gamedata.Mission
	for _, mission := range gamedata.Missions {
		requirement := gamedata.MissionRequirement(mission.ID)
		if requirement == "no_nav" || (requirement == "sigsegv-mvm" && !opt.includeSig) {
			continue
		}
		if int(crc32.ChecksumIEEE([]byte(mission.PopFile)))%opt.shards != opt.shard {
			continue
		}
		selected = append(selected, mission)
	}
	sort.Slice(selected, func(i, j int) bool {
		left, _ := gamedata.MapByID(selected[i].Map)
		right, _ := gamedata.MapByID(selected[j].Map)
		if left.Name == right.Name {
			return selected[i].PopFile < selected[j].PopFile
		}
		return left.Name < right.Name
	})
	return selected, nil
}

func writeResult(row result) {
	body, err := json.Marshal(row)
	if err != nil {
		fmt.Fprintf(os.Stderr, "waveprobe: encode result: %v\n", err)
		return
	}
	fmt.Println(string(body))
}

type server struct {
	address  string
	password string
}

func (s server) exec(command string) (string, error) {
	client, err := rcon.Dial(s.address, s.password)
	if err != nil {
		return "", err
	}
	defer func() { _ = client.Close() }()
	return client.Exec(command)
}

func (s server) status() (probeStatus, error) {
	reply, err := s.exec("sm_waveprobe_status")
	if err != nil {
		return probeStatus{}, err
	}
	return parseStatus(reply)
}

func parseStatus(reply string) (probeStatus, error) {
	position := strings.Index(reply, "WAVEPROBE state=")
	if position < 0 {
		return probeStatus{}, fmt.Errorf("unexpected probe reply: %q", reply)
	}
	fields := map[string]string{}
	for token := range strings.FieldsSeq(reply[position+len("WAVEPROBE "):]) {
		key, value, ok := strings.Cut(token, "=")
		if ok {
			fields[key] = value
		}
	}
	var status probeStatus
	status.State = fields["state"]
	status.Map = fields["map"]
	status.Pop = fields["pop"]
	var err error
	for _, field := range []struct {
		name string
		dest *int
	}{{"max", &status.Max}, {"gamewave", &status.GameWave},
		{"expected", &status.Expected}, {"observed", &status.Observed},
		{"bots", &status.Bots}, {"tanks", &status.Tanks},
		{"defteam", &status.DefTeam}, {"playerteam", &status.PlayerTeam},
		{"enemyteam", &status.EnemyTeam}} {
		*field.dest, err = strconv.Atoi(fields[field.name])
		if err != nil {
			return probeStatus{}, fmt.Errorf("invalid %s in probe reply %q: %w", field.name, reply, err)
		}
	}
	status.Elapsed, _ = strconv.ParseFloat(fields["elapsed"], 64)
	if status.State == "" || status.Map == "" || status.Pop == "" {
		return probeStatus{}, fmt.Errorf("incomplete probe reply: %q", reply)
	}
	return status, nil
}

func timescaleMatches(reply string, speed int) bool {
	equals := strings.IndexByte(reply, '=')
	if equals < 0 {
		return false
	}
	start := strings.IndexByte(reply[equals+1:], '"')
	if start < 0 {
		return false
	}
	start += equals + 1
	end := strings.IndexByte(reply[start+1:], '"')
	if end < 0 {
		return false
	}
	value, err := strconv.ParseFloat(reply[start+1:start+1+end], 64)
	return err == nil && value == float64(speed)
}

func (s server) await(timeout time.Duration, predicate func(probeStatus) bool) error {
	deadline := time.Now().Add(timeout)
	var last probeStatus
	var lastErr error
	for time.Now().Before(deadline) {
		last, lastErr = s.status()
		if lastErr == nil && predicate(last) {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Errorf("timed out after %s (last status %+v): %w", timeout, last, lastErr)
	}
	return fmt.Errorf("timed out after %s (last status %+v)", timeout, last)
}

func (s server) load(mapName string, mission gamedata.Mission, mode string, timeout time.Duration) error {
	status, err := s.status()
	if err != nil {
		return err
	}
	if status.Map != mapName {
		// The server may drop the RCON connection as changelevel runs. The
		// status poll, rather than that connection, determines success.
		_, _ = s.exec("changelevel " + mapName)
		if err := s.await(timeout, func(st probeStatus) bool { return st.Map == mapName }); err != nil {
			return err
		}
		// The map name becomes visible before the population manager finishes
		// initializing. A popfile lookup in that window can reject a valid file.
		time.Sleep(time.Second)
	}
	teamName, playerTeam, err := s.configureMode(mission, mode)
	if err != nil {
		return err
	}
	// The modifier command can queue a popfile reload. Let that command run,
	// then explicitly load the full name so the observed wave uses this mode.
	time.Sleep(500 * time.Millisecond)
	var popReply string
	popDeadline := time.Now().Add(min(timeout, 15*time.Second))
	for {
		popReply, err = s.exec("tf_mvm_popfile " + mission.PopFile)
		if err == nil && !strings.Contains(popReply, "Could not find a valid population file") {
			break
		}
		if time.Now().After(popDeadline) {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	if strings.Contains(popReply, "Could not find a valid population file") {
		return fmt.Errorf("population file %s was rejected: %s", mission.PopFile, strings.TrimSpace(popReply))
	}
	loaded := func(st probeStatus) bool {
		return st.Map == mapName && st.Pop == mission.PopFile &&
			st.Max == int(mission.Waves) && st.GameWave == 1 &&
			st.DefTeam == playerTeam && st.PlayerTeam == playerTeam &&
			st.EnemyTeam == 5-playerTeam
	}
	if err := s.await(5*time.Second, loaded); err != nil {
		// A popfile reload can drop the fake player after the first wake.
		// Recreate it once the population manager is initialized.
		if _, wakeErr := s.exec("sm_waveprobe_wake " + teamName); wakeErr != nil {
			return wakeErr
		}
		if retryErr := s.await(timeout, loaded); retryErr != nil {
			return fmt.Errorf("mission %s did not load: %w", mission.PopFile, retryErr)
		}
	}
	return nil
}

func (s server) configureMode(mission gamedata.Mission, mode string) (string, int, error) {
	if _, err := s.exec("sm_waveprobe_reset"); err != nil {
		return "", 0, err
	}
	// No human is connected to these disposable servers. server.cfg resets
	// this value on map changes.
	if _, err := s.exec("tf_mvm_min_players_to_start 0"); err != nil {
		return "", 0, err
	}
	// Bot Surge prints a line per rewritten spawner at debug level 1. That
	// overflows one Source RCON packet on some missions and drowns the probe.
	if _, err := s.exec("tf2ap_debug 0"); err != nil {
		return "", 0, err
	}
	if _, err := s.exec("tf2ap_bots_wait_for_players 0"); err != nil {
		return "", 0, err
	}
	teamName := "red"
	playerTeam := 2
	if strings.Contains(mission.PopFile, "_rev_") {
		teamName = "blue"
		playerTeam = 3
	}
	if _, err := s.exec("sm_waveprobe_wake " + teamName); err != nil {
		return "", 0, err
	}
	if _, err := s.exec("sm_ap_modifier clear"); err != nil {
		return "", 0, err
	}
	if mode == "surge" {
		if _, err := s.exec("sm_ap_modifier on bot_surge"); err != nil {
			return "", 0, err
		}
	}
	modifiers, err := s.exec("sm_ap_modifiers")
	if err != nil {
		return "", 0, err
	}
	active := strings.Contains(modifiers, "Mission modifiers: Bot Surge")
	if active != (mode == "surge") {
		return "", 0, fmt.Errorf("expected mode %s, got %q", mode, strings.TrimSpace(modifiers))
	}
	return teamName, playerTeam, nil
}

var errWallTimeout = errors.New("wall-clock timeout before game-time limit")

func (s server) testWave(mission gamedata.Mission, wave, seed int, timeout, gameTimeout time.Duration) (probeStatus, error) {
	if err := s.await(timeout, func(st probeStatus) bool {
		return st.Pop == mission.PopFile && st.GameWave == wave
	}); err != nil {
		return probeStatus{}, fmt.Errorf("wave %d was not initialized: %w", wave, err)
	}
	if _, err := s.exec(fmt.Sprintf("sm_waveprobe_arm %d %d", wave, seed)); err != nil {
		return probeStatus{}, err
	}
	if _, err := s.exec("mp_restartgame 1"); err != nil {
		return probeStatus{}, err
	}
	deadline := time.Now().Add(timeout)
	var last probeStatus
	for time.Now().Before(deadline) {
		status, err := s.status()
		if err == nil {
			last = status
			if status.State == "passed" && status.Expected == wave && status.Observed == wave {
				return status, nil
			}
			if status.State == "failed" {
				return status, fmt.Errorf("game or probe failed wave %d: %+v", wave, status)
			}
			if status.State == "running" && status.Elapsed >= gameTimeout.Seconds() {
				s.stopWave(mission)
				return status, fmt.Errorf("wave %d remained active for %.1f game seconds: %+v", wave, status.Elapsed, status)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	s.stopWave(mission)
	return last, fmt.Errorf("%w: wave %d did not complete within %s, %.1f game seconds: %+v",
		errWallTimeout, wave, timeout, last.Elapsed, last)
}

func (s server) stopWave(mission gamedata.Mission) {
	// An active Bot Surge wave can keep creating bots and projectiles during
	// a queued map change. Reset the population manager before leaving it.
	_, _ = s.exec("tf_mvm_popfile " + mission.PopFile)
	_, _ = s.exec("sm_waveprobe_reset")
}
