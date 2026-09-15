package composeenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestWriteChangesOnlyOwnedValuesAndPreservesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	body := "# keep this comment\nUNKNOWN=untouched\nSRCDS_HOSTNAME=old\nSRCDS_PW=old\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	before := settings.Defaults()
	before.SrcdsHostname = "old"
	before.SrcdsPw = "old"
	after := before
	after.SrcdsHostname = "Mann's $erver"
	after.SrcdsJoinHost = "tf2.example.com"
	if err := Write(path, before, after); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{
		"# keep this comment", "UNKNOWN=untouched", "SRCDS_PW=old",
		`SRCDS_HOSTNAME='Mann\'s $erver'`, "TF2AP_JOIN_HOST='tf2.example.com'",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("written .env does not contain %q:\n%s", want, text)
		}
	}
}

func TestWriteDoesNotReplaceTheMountedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("SRCDS_PORT=27015\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	before := settings.Defaults()
	after := before
	after.SrcdsPort = 27115
	infoBefore, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Write(path, before, after); err != nil {
		t.Fatal(err)
	}
	infoAfter, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(infoBefore, infoAfter) {
		t.Fatal("Write replaced the inode instead of updating the bind-mounted file")
	}
}
