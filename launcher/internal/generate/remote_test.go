package generate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteSendsThePlayerFileAndKeepsTheArchiveName(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/generate" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "name: tf2\n" {
			t.Errorf("player file = %q", body)
		}
		w.Header().Set("Content-Disposition", `attachment; filename="AP_test.zip"`)
		_, _ = w.Write([]byte("PK-archive"))
	}))
	defer server.Close()

	result, err := Remote(context.Background(), server.URL, []byte("name: tf2\n"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = result.Body.Close() }()
	archive, _ := io.ReadAll(result.Body)
	if result.Name != "AP_test.zip" || string(archive) != "PK-archive" {
		t.Fatalf("result = %q %q", result.Name, archive)
	}
}

func TestRemoteReportsTheGeneratorFailure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "the YAML is invalid", http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	_, err := Remote(context.Background(), server.URL, []byte("broken"))
	if err == nil || !strings.Contains(err.Error(), "the YAML is invalid") {
		t.Fatalf("Remote error = %v", err)
	}
}
