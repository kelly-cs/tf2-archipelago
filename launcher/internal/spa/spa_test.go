package spa

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func built() fstest.MapFS {
	return fstest.MapFS{
		"index.html":          {Data: []byte("<app-root></app-root>")},
		"main-AMM3UY6G.js":    {Data: []byte("bootstrapApplication")},
		"styles-QAEZWQS6.css": {Data: []byte("body{}")},
	}
}

func get(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

// A refresh on a client-side route is the one thing a file server gets wrong on
// its own: /session is not a file, and a 404 there loses the player's page.
func TestClientRouteServesTheApp(t *testing.T) {
	handler := handlerFor(built())
	for _, path := range []string{"/", "/session", "/settings/missions", "/log"} {
		response := get(t, handler, path)
		if response.Code != http.StatusOK {
			t.Errorf("%s answered %d, want 200", path, response.Code)
		}
		if body := response.Body.String(); body != "<app-root></app-root>" {
			t.Errorf("%s served %q, want index.html", path, body)
		}
	}
}

func TestFilesAreServedAsThemselves(t *testing.T) {
	response := get(t, handlerFor(built()), "/main-AMM3UY6G.js")
	if body := response.Body.String(); body != "bootstrapApplication" {
		t.Errorf("served %q, want the bundle", body)
	}
}

// A hashed name is safe to keep forever and index.html never is: getting this
// the wrong way round leaves a player looking at yesterday's launcher after an
// update, with no way to know it.
func TestOnlyHashedFilesAreCachedForever(t *testing.T) {
	handler := handlerFor(built())
	if got := get(t, handler, "/main-AMM3UY6G.js").Header().Get("Cache-Control"); got != immutableMax {
		t.Errorf("the bundle got %q, want %q", got, immutableMax)
	}
	if got := get(t, handler, "/index.html").Header().Get("Cache-Control"); got == immutableMax {
		t.Error("index.html must not be cached forever: it names the bundle")
	}
	if got := get(t, handler, "/session").Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("a client route got %q, want no-cache", got)
	}
}

func TestHashedNamesTheBuildWrites(t *testing.T) {
	for name, want := range map[string]bool{
		"main-AMM3UY6G.js":        true,
		"styles-QAEZWQS6.css":     true,
		"media/logo-4KXQ2P9F.svg": true,
		"index.html":              false,
		"favicon.ico":             false,
		"3rdpartylicenses.txt":    false,
		"prerendered-routes.json": false,
		"a-b.js":                  false,
	} {
		if got := hashed(name); got != want {
			t.Errorf("hashed(%q) = %v, want %v", name, got, want)
		}
	}
}

// The real embed, so a build that stopped writing index.html fails here rather
// than serving a blank page to the player.
func TestTheEmbeddedBuildHasAPage(t *testing.T) {
	response := get(t, Handler(), "/")
	if response.Code != http.StatusOK {
		t.Fatalf("the embedded app answered %d, want 200", response.Code)
	}
	if response.Body.Len() == 0 {
		t.Error("the embedded index.html is empty")
	}
}
