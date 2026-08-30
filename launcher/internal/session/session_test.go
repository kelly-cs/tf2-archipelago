package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchReadsTheBridge(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"api_version":3,"connected":true,"slot":"tf2","checks":4,"items":2}`))
	})
	mux.HandleFunc("GET /missions", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"missions":[{"popfile":"mvm_decoy","name":"Doe's Drill","map":"mvm_decoy","waves":8,"loadout":"medieval","unlocked":true,"cleared":false}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	snapshot, err := Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !snapshot.Health.Connected || snapshot.Health.Checks != 4 {
		t.Errorf("health = %+v", snapshot.Health)
	}
	if len(snapshot.Missions) != 1 || snapshot.Missions[0].Name != "Doe's Drill" || !snapshot.Missions[0].Unlocked {
		t.Errorf("missions = %+v", snapshot.Missions)
	}
	if got := snapshot.Missions[0].Loadout; got != "medieval" {
		t.Errorf("loadout = %q", got)
	}
	if got := snapshot.Health.Summary(); got != "connected as tf2, 4 checks, 2 items" {
		t.Errorf("summary = %q", got)
	}
}

func TestFetchSaysWhenTheBridgeIsGone(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	if _, err := Fetch(context.Background(), server.URL); err == nil {
		t.Fatal("a closed bridge did not fail")
	}
}
