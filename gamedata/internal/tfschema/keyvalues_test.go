package tfschema

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"unicode/utf16"
)

func TestParseReadsBlocksAndKeepsTheLastDuplicate(t *testing.T) {
	root, err := Parse(`"items_game"
{
	// a comment, which is skipped
	"items"
	{
		"200"
		{
			"name"	"The Force-a-Nature"
			"prefab"	"weapon_scattergun"
			"item_name"	"#TF_Unique_Achievement_Scattergun_Double"
		}
	}
	"twice"	"first"
	"twice"	"second"
	"escaped"	"say \"hi\""
}`)
	if err != nil {
		t.Fatal(err)
	}
	item := root.Block("items").Block("200")
	if item == nil || item.Text("name") != "The Force-a-Nature" {
		t.Fatalf("item 200 not read: %+v", item)
	}
	if got := root.Text("twice"); got != "second" {
		t.Errorf("duplicate key gave %q, want the last", got)
	}
	if got := root.Text("escaped"); got != `say "hi"` {
		t.Errorf("escape gave %q", got)
	}
	if got := root.Keys(); !slices.Equal(got, []string{"items", "twice", "escaped"}) {
		t.Errorf("keys %v", got)
	}
}

func TestEnglishDecodesUTF16AndLowersTheToken(t *testing.T) {
	text := "\"lang\"\n{\n\"Tokens\"\n{\n\"TF_Weapon_Bat\"\t\"Bat\"\n\"TF_Quote\"\t\"say \\\"hi\\\"\"\n}\n}\n"
	units := utf16.Encode([]rune(text))
	body := []byte{0xFF, 0xFE}
	for _, u := range units {
		body = binary.LittleEndian.AppendUint16(body, u)
	}
	path := filepath.Join(t.TempDir(), "tf_english.txt")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	names, err := English(path)
	if err != nil {
		t.Fatal(err)
	}
	if names["tf_weapon_bat"] != "Bat" || names["tf_quote"] != `say "hi"` {
		t.Errorf("names %v", names)
	}
}
