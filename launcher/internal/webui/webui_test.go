package webui

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/webapi"
)

func localRequest(method, target string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	request.Host = "127.0.0.1"
	return request
}

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
	app := newApp(s, nil)
	app.OpenSettings("Missions")

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("assets", settings.CommunityPackMoonlight)
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
	if !slices.Equal(imported, []string{settings.CommunityPackMoonlight}) {
		t.Fatalf("imported = %v", imported)
	}

	snapshot := app.Snapshot()
	field, ok := snapshot.Form.Field("missions.moonlight")
	if !ok || field.Value != "true" {
		t.Fatal("the imported pack was not selected")
	}
	for popFile, name := range map[string]string{
		"mvm_coaltown_int_trouble_in_mann_town": "Trouble in Mann Town",
		"mvm_mannworks_adv_manntenance":         "Manntenance",
	} {
		if !slices.ContainsFunc(snapshot.MissionPool, func(row webapi.MissionPoolRow) bool {
			return row.Name == name && row.Source == "Imported"
		}) {
			t.Errorf("imported Moonlight assets did not add %q to the mission pool", name)
		}
		if pooled, ok := snapshot.Form.Field("missions.pool." + popFile); !ok || pooled.Value != "true" {
			t.Errorf("imported mission %q is still excluded from the pool", name)
		}
	}
}

func TestPageCarriesTheWholeOperationalInterface(t *testing.T) {
	request := localRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	newApp(settings.Defaults(), nil).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET / answered %d", response.Code)
	}
	for _, want := range []string{
		"EventSource", "Start", "Restart", "Join", "Settings", "rcon", "Send", "Quit",
		"Source", "Map", "Mission name", "Wave #s", "Compatibility status", "Mods",
		"missionPoolHeader", "aria-sort",
		"state.join_url",
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
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	app := newApp(s, nil)
	app.OpenSettings("Run")

	for _, test := range []struct {
		path string
		want string
	}{
		{"/api/files/settings", "install_root"},
		{"/api/files/player", "Team Fortress 2 Mann vs Machine"},
		{"/api/files/install", s.InstallRoot},
	} {
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, localRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("GET %s answered %d: %s", test.path, response.Code, response.Body.String())
		} else if !strings.Contains(response.Body.String(), test.want) {
			t.Errorf("GET %s did not contain %q", test.path, test.want)
		}
	}

	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, localRequest(http.MethodGet, "/api/files/install/tf/cfg/server.cfg", nil))
	if response.Code != http.StatusNotFound {
		t.Errorf("an unrequested install file answered %d, want 404", response.Code)
	}
}

func TestJoinRunsInTheBrowser(t *testing.T) {
	body := string(page)
	if !strings.Contains(body, "window.location.href") || !strings.Contains(body, "state.join_url") {
		t.Fatal("Join does not open the Steam URL in the browser")
	}
	if strings.Contains(body, "/api/server/join") {
		t.Fatal("Join still delegates Steam launching to the server OS")
	}
	for _, path := range []string{
		"/api/files/settings", "/api/files/player", "/api/files/install",
		"/api/files/seed", "/api/files/debug", "/api/open/funnel",
	} {
		if !strings.Contains(body, path) {
			t.Errorf("the browser does not own %s", path)
		}
	}
}

func TestPeriodicDrawDoesNotRebuildUnchangedSettings(t *testing.T) {
	if !bytes.Contains(page, []byte("function drawSettings")) ||
		!bytes.Contains(page, []byte("JSON.stringify([snapshot.form, snapshot.mission_pool])")) {
		t.Fatal("settings controls are rebuilt on every session refresh")
	}
}

func TestAnotherOriginCannotPressButtons(t *testing.T) {
	request := localRequest(http.MethodPost, "/api/server/start", strings.NewReader("{}"))
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()
	newApp(settings.Defaults(), nil).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Errorf("cross-origin POST answered %d, want 403", response.Code)
	}
}

func TestRebindingHostCannotReachTheLauncher(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/rcon", strings.NewReader(`{"command":"say hello"}`))
	request.Host = "evil.example"
	request.Header.Set("Origin", "http://evil.example")
	response := httptest.NewRecorder()
	newApp(settings.Defaults(), nil).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Errorf("rebound Host and Origin answered %d, want 403", response.Code)
	}
}

func TestForeignHostCannotReadFiles(t *testing.T) {
	s := settings.Defaults()
	s.APPassword = "archipelago-secret"
	s.SrcdsRconPw = "rcon-secret"
	s.SrcdsToken = "steam-secret"
	app := newApp(s, nil)
	app.OpenSettings("")

	for _, path := range []string{
		"/api/files/settings", "/api/files/player", "/api/files/install",
		"/api/files/seed", "/api/files/debug",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Host = "evil.example"
		request.Header.Set("Origin", "http://evil.example")
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Errorf("foreign GET %s answered %d, want 403", path, response.Code)
		}
		for _, secret := range []string{s.APPassword, s.SrcdsRconPw, s.SrcdsToken} {
			if strings.Contains(response.Body.String(), secret) {
				t.Errorf("foreign GET %s exposed %q", path, secret)
			}
		}
	}
}

func TestEveryFormRowKindHasAnExplicitBrowserControl(t *testing.T) {
	for kind := form.Text; kind <= form.Confirm; kind++ {
		marker := "case FieldKind." + kind.String() + ":"
		if !bytes.Contains(page, []byte(marker)) {
			t.Errorf("form kind %q has no explicit browser control", kind)
		}
	}
}
