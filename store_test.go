package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Tournament)
	}{
		{
			name:   "fresh tournament",
			mutate: func(*Tournament) {},
		},
		{
			name: "with players",
			mutate: func(t *Tournament) {
				t.Players = []Player{{ID: "p1", Name: "Ana"}, {ID: "p2", Name: "Beto"}}
			},
		},
		{
			name: "with rounds and results",
			mutate: func(t *Tournament) {
				t.Players = []Player{{ID: "p1", Name: "Ana"}, {ID: "p2", Name: "Beto"}}
				t.Rounds = []Round{{
					Number: 1,
					Matches: []Match{
						{White: "p1", Black: "p2", Result: "1-0"},
						{White: "p3", IsBye: true},
					},
				}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "nested", "suizo.json")
			s, err := openStore(path)
			if err != nil {
				t.Fatalf("openStore: %v", err)
			}

			want := &Tournament{Name: "test"}
			tt.mutate(want)

			if err := s.save(want); err != nil {
				t.Fatalf("save: %v", err)
			}
			got, err := s.load()
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("round trip changed the tournament:\n got %#v\nwant %#v", got, want)
			}
		})
	}
}

func TestLoadMissingFileReturnsFreshTournament(t *testing.T) {
	s, err := openStore(filepath.Join(t.TempDir(), "suizo.json"))
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	got, err := s.load()
	if err != nil {
		t.Fatalf("load on missing file: %v", err)
	}
	if got.Players != nil || got.Rounds != nil {
		t.Errorf("fresh tournament has players/rounds: %+v", got)
	}
}

func TestLoadCorruptFileFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suizo.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	s, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	if _, err := s.load(); err == nil {
		t.Fatal("load on corrupt file succeeded, want error")
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	s, err := openStore(filepath.Join(dir, "suizo.json"))
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := s.save(&Tournament{Name: "test"}); err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			t.Errorf("stray file left behind: %s", e.Name())
		}
	}
}

// TestSaveDocumentIsIndentedAndTrailingNewline pins the human-readable on-disk
// shape promised by docs/FORMAT.md.
func TestSaveDocumentIsIndentedAndTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suizo.json")
	s, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	if err := s.save(&Tournament{Name: "test"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("document does not end with a newline")
	}
	if !bytes.Contains(data, []byte("\n  \"name\"")) {
		t.Error("document is not indented with two spaces")
	}
	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("document is not valid JSON: %v", err)
	}
}
