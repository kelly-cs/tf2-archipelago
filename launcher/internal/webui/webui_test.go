package webui

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestOpenFailurePromptsWithManualURL(t *testing.T) {
	const address = "http://localhost:35337"
	var output bytes.Buffer
	wantErr := errors.New("xdg-open is missing")
	err := openOrPrompt(address, &output, func(got string) error {
		if got != address {
			t.Errorf("opened %q, want %q", got, address)
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("openOrPrompt error = %v, want %v", err, wantErr)
	}
	got := output.String()
	for _, value := range []string{address, wantErr.Error(), "keep running"} {
		if !strings.Contains(got, value) {
			t.Errorf("manual prompt does not contain %q:\\n%s", value, got)
		}
	}
}

func TestOpenSuccessDoesNotPrompt(t *testing.T) {
	var output bytes.Buffer
	err := openOrPrompt("http://127.0.0.1:1", &output, func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 {
		t.Fatalf("successful open printed %q", output.String())
	}
}

func TestMissionPoolRowsCarryTableMetadata(t *testing.T) {
	rows := missionPoolRows(form.NewState(settings.Defaults()), nil, nil)
	if len(rows) == 0 {
		t.Fatal("the mission pool table is empty")
	}
	row := rows[0]
	if row.Field == "" || row.Source == "" || row.Name == "" || row.Waves == "" ||
		row.Compatibility == "" || row.Mods == "" {
		t.Fatalf("mission pool row has an empty column: %+v", row)
	}
	if !strings.HasPrefix(row.Field, "missions.pool.") {
		t.Errorf("mission pool field = %q", row.Field)
	}
}

func TestMissionPoolRowsExplainDifficultyFloor(t *testing.T) {
	state := form.NewState(settings.Defaults())
	state.Settings.MvmDifficulty = "advanced"
	rows := missionPoolRows(state, nil, nil)
	if !slices.ContainsFunc(rows, func(row MissionPoolRow) bool {
		return strings.HasPrefix(row.Compatibility, "Below Advanced")
	}) {
		t.Fatal("the table does not explain why lower-tier missions are ineligible")
	}
}

func assetZIP(t *testing.T) []byte {
	t.Helper()
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("tf/download/maps/mvm_example.bsp")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("map")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
}

func TestImportSelectsPackAndMissions(t *testing.T) {
	s := settings.Defaults()
	s.CommunityContentDir = t.TempDir()
	app := New(s, nil)
	app.OpenSettings("Missions")

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("assets", settings.CommunityPackPotato)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(assetZIP(t)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/import", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	imported, err := app.importAssets(request)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(imported, []string{settings.CommunityPackPotato}) {
		t.Fatalf("imported = %v", imported)
	}
	if !slices.Contains(app.draft.Settings.CommunityPacks, settings.CommunityPackPotato) {
		t.Fatal("the imported pack was not selected")
	}
	rows := missionPoolRows(*app.draft, app.community, app.imported)
	if !slices.ContainsFunc(rows, func(row MissionPoolRow) bool { return row.Source == "Imported" }) {
		t.Fatal("the imported pack added no Imported mission rows")
	}
}

func TestPoolNoneClearsTheNamedStartMission(t *testing.T) {
	app := New(settings.Defaults(), nil)
	app.OpenSettings("Missions")
	app.draft.Settings.MvmStartMission = "mvm_decoy"
	app.setPool(false)
	if got := app.draft.Settings.MvmStartMission; got != "" {
		t.Errorf("pool none kept start mission %q", got)
	}
}

func TestPageCarriesTheWholeOperationalInterface(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	New(settings.Defaults(), nil).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET / answered %d", response.Code)
	}
	for _, want := range []string{
		"EventSource", "Start", "Restart", "Join", "Settings", "rcon", "Send", "Quit",
		"Source", "Mission name", "Wave #s", "Compatibility status", "Mods",
		"location.href=state.join_url",
	} {
		if !strings.Contains(response.Body.String(), want) {
			t.Errorf("the page does not contain %q", want)
		}
	}
	if strings.Contains(response.Body.String(), "/api/server/join") {
		t.Error("Join still asks the launcher host to open Steam")
	}
}

func TestFilesOpenThroughTheBrowser(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	s := settings.Defaults()
	s.InstallRoot = t.TempDir()
	if err := os.WriteFile(filepath.Join(s.InstallRoot, "visible.txt"), []byte("install file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	app := New(s, nil)
	app.OpenSettings("Run")

	for _, test := range []struct {
		path string
		want string
	}{
		{"/api/files/settings", "install_root"},
		{"/api/files/player", "Team Fortress 2 Mann vs Machine"},
		{"/api/files/install/visible.txt", "install file"},
	} {
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("GET %s answered %d: %s", test.path, response.Code, response.Body.String())
		} else if !strings.Contains(response.Body.String(), test.want) {
			t.Errorf("GET %s did not contain %q", test.path, test.want)
		}
	}
}

func TestSnapshotProvidesSteamURLToTheBrowser(t *testing.T) {
	snapshot := New(settings.Defaults(), nil).Snapshot()
	if !strings.HasPrefix(snapshot.JoinURL, "steam://run/440//+connect%20") {
		t.Errorf("join URL = %q", snapshot.JoinURL)
	}
}

func TestJoinRunsInTheBrowser(t *testing.T) {
	body := string(page)
	if !strings.Contains(body, "location.href=state.join_url") {
		t.Fatal("Join does not open the Steam URL in the browser")
	}
	if strings.Contains(body, "/api/server/join") {
		t.Fatal("Join still delegates Steam launching to the server OS")
	}
	for _, path := range []string{
		"/api/files/settings", "/api/files/player", "/api/files/install/",
		"/api/files/seed", "/api/files/debug", "/api/open/funnel",
	} {
		if !strings.Contains(body, path) {
			t.Errorf("the browser does not own %s", path)
		}
	}
}

func TestSteamJoinWaitsForThePublishedAddress(t *testing.T) {
	s := settings.Defaults()
	s.SrcdsToken = "real-token"
	s.SrcdsReach = settings.ReachSteam
	app := New(s, nil)
	if got := app.Snapshot().JoinURL; got != "" {
		t.Fatalf("Join used a local address while Steam was still starting: %q", got)
	}
	app.append(apruntime.Line{Source: "srcds", Text: "FakeIP allocation succeeded: 169.254.13.42:20232, 20233"})
	if got := app.Snapshot().JoinURL; !strings.Contains(got, "169.254.13.42:20232") {
		t.Fatalf("Join did not use Steam's published address: %q", got)
	}
}

func TestPeriodicDrawDoesNotRebuildUnchangedSettings(t *testing.T) {
	if !bytes.Contains(page, []byte("key!==formKey")) {
		t.Fatal("settings controls are rebuilt on every session refresh")
	}
}

func TestSettingsAreTheFormModelAndChangesGoThroughApply(t *testing.T) {
	app := New(settings.Defaults(), nil)
	app.OpenSettings("Rewards")
	if err := app.Change(form.Change{Field: "rewards.traps", Value: "42"}); err != nil {
		t.Fatal(err)
	}

	snapshot := app.Snapshot()
	if snapshot.Form == nil {
		t.Fatal("opening settings produced no form model")
	}
	field, ok := snapshot.Form.Field("rewards.traps")
	if !ok {
		t.Fatal("the form omitted rewards.traps")
	}
	if field.Value != "42" {
		t.Errorf("the trap share is %q, want 42", field.Value)
	}
	if app.settings.MvmTrapPct == 42 {
		t.Error("editing the draft changed the running settings before Save")
	}
}

func TestSnapshotDoesNotExposePasswords(t *testing.T) {
	s := settings.Defaults()
	s.APPassword = "archipelago-secret"
	s.SrcdsRconPw = "rcon-secret"
	s.SrcdsPw = "game-secret"
	app := New(s, nil)
	app.OpenSettings("")

	data, err := json.Marshal(app.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, secret := range []string{"archipelago-secret", "rcon-secret"} {
		if strings.Contains(text, secret) {
			t.Errorf("the browser snapshot exposes %q", secret)
		}
	}
	// The join line intentionally contains the ordinary game password: it is
	// what the player copies to friends and what the old window showed.
	if !strings.Contains(text, "game-secret") {
		t.Error("the snapshot lost the join password")
	}
}

func TestAnotherOriginCannotPressButtons(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/server/start", strings.NewReader("{}"))
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()
	New(settings.Defaults(), nil).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Errorf("cross-origin POST answered %d, want 403", response.Code)
	}
}

func TestEveryFormButtonIsWiredAndNothingStaleIsWired(t *testing.T) {
	state := form.NewState(settings.Defaults())
	declared := map[string]bool{}
	for _, spec := range form.Specs(state, form.Env{}) {
		if spec.Kind == form.Action || spec.Kind == form.Confirm {
			declared[spec.ID] = true
			if !slices.Contains(wiredActions, spec.ID) {
				t.Errorf("form declares %q and the browser does not wire it", spec.ID)
			}
		}
	}
	for _, id := range wiredActions {
		if !declared[id] {
			t.Errorf("the browser wires %q and form does not declare it", id)
		}
	}
}
