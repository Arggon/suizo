package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPlayersCLI(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		prep       func(*store) // optional setup against the same store
		wantOut    string
		wantErrSub string
		wantCode   int
	}{
		{
			name:     "add prints the new id",
			args:     []string{"players", "add", "Ana"},
			wantOut:  "p1\n",
			wantCode: 0,
		},
		{
			name:     "list on empty store",
			args:     []string{"players", "list"},
			wantOut:  "(no players yet)\n",
			wantCode: 0,
		},
		{
			name:     "list after adds",
			args:     []string{"players", "list"},
			prep:     func(s *store) { mustAdd(t, s, "Ana", "Beto") },
			wantOut:  "p1\tAna\np2\tBeto\n",
			wantCode: 0,
		},
		{
			name:       "add without a name fails",
			args:       []string{"players", "add"},
			wantErrSub: "usage: suizo players add",
			wantCode:   2,
		},
		{
			name:       "empty name is rejected",
			args:       []string{"players", "add", ""},
			wantErrSub: "player name must not be empty",
			wantCode:   1,
		},
		{
			name:       "unknown subcommand fails",
			args:       []string{"players", "rename", "x"},
			wantErrSub: "unknown players command",
			wantCode:   2,
		},
		{
			name:       "unknown command fails",
			args:       []string{"rounds"},
			wantErrSub: "unknown command",
			wantCode:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "suizo.json")
			s, err := openStore(path)
			if err != nil {
				t.Fatalf("openStore: %v", err)
			}
			if tt.prep != nil {
				tt.prep(s)
			}

			var out, errOut bytes.Buffer
			code := run(tt.args, path, &out, &errOut)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %q)", code, tt.wantCode, errOut.String())
			}
			if out.String() != tt.wantOut {
				t.Errorf("stdout = %q, want %q", out.String(), tt.wantOut)
			}
			if tt.wantErrSub != "" && !strings.Contains(errOut.String(), tt.wantErrSub) {
				t.Errorf("stderr = %q, want it to contain %q", errOut.String(), tt.wantErrSub)
			}
		})
	}
}

// TestRunAddPersistsAcrossInvocations exercises the end-to-end guarantee:
// separate run() calls against the same file see each other's writes.
func TestRunAddPersistsAcrossInvocations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suizo.json")
	var out bytes.Buffer
	for _, name := range []string{"Ana", "Beto"} {
		if code := run([]string{"players", "add", name}, path, &out, &out); code != 0 {
			t.Fatalf("add %q exited %d", name, code)
		}
	}
	out.Reset()
	if code := run([]string{"players", "list"}, path, &out, &out); code != 0 {
		t.Fatalf("list exited %d", code)
	}
	if got, want := out.String(), "p1\tAna\np2\tBeto\n"; got != want {
		t.Errorf("list = %q, want %q", got, want)
	}
}

func mustAdd(t *testing.T, s *store, names ...string) {
	t.Helper()
	trn, err := s.load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, n := range names {
		if _, err := trn.addPlayer(n); err != nil {
			t.Fatalf("addPlayer(%q): %v", n, err)
		}
	}
	if err := s.save(trn); err != nil {
		t.Fatalf("save: %v", err)
	}
}
