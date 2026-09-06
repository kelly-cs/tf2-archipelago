package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.12.0", "v1.11.3", true},
		{"v1.11.3", "v1.11.3", false},
		{"v1.11.2", "v1.11.3", false},
		{"v2.0.0", "v1.99.99", true},
		{"nightly", "v1.11.3", false},
		{"v1.12.0", "dev", false},
		{"v1.12.0-rc1", "v1.11.9", true},
		{"1.12", "v1.11.9", true},
	}
	for _, c := range cases {
		if got := Newer(c.latest, c.current); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestNoticeNamesTheReleaseAndStaysQuietOtherwise(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name": "v1.12.0"}`))
	}))
	defer server.Close()

	got := Check{Current: "v1.11.3", API: server.URL, Client: server.Client()}.Notice(context.Background())
	if !strings.Contains(got, "v1.12.0") || !strings.Contains(got, "v1.11.3") || !strings.Contains(got, DownloadPage) {
		t.Errorf("notice %q", got)
	}
	if got := (Check{Current: "v1.12.0", API: server.URL, Client: server.Client()}).Notice(context.Background()); got != "" {
		t.Errorf("a current build was told %q", got)
	}
	if got := (Check{Current: "", API: server.URL, Client: server.Client()}).Notice(context.Background()); got != "" {
		t.Errorf("a hand build was told %q", got)
	}

	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()
	if got := (Check{Current: "v1.11.3", API: down.URL, Client: down.Client()}).Notice(context.Background()); got != "" {
		t.Errorf("an unreachable GitHub produced %q", got)
	}
}
