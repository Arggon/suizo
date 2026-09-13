package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeStateFile seeds a store file with the given JSON and returns its path.
func writeStateFile(t *testing.T, json string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "suizo.json")
	if err := os.WriteFile(path, []byte(json), 0o600); err != nil {
		t.Fatalf("write state file: %v", err)
	}
	return path
}

func TestStandings(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Tournament
		want  []Standing
	}{
		{
			name:  "empty tournament",
			setup: func() *Tournament { return &Tournament{} },
			want:  nil,
		},
		{
			name: "multi round accumulation",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
					{ID: "p3", Name: "Carla"},
				}}
				t.Rounds = []Round{
					{Number: 1, Matches: []Match{
						{White: "p1", Black: "p2", Result: "1-0"},
						{White: "p3", IsBye: true},
					}},
					{Number: 2, Matches: []Match{
						{White: "p1", Black: "p3", Result: "1-0"},
						{White: "p2", IsBye: true},
					}},
				}
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 2, Played: 2},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 1, Played: 2},
				{Rank: 3, ID: "p3", Name: "Carla", Points: 1, Played: 2},
			},
		},
		{
			name: "bye counts as a full point",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
				}}
				t.Rounds = []Round{{Number: 1, Matches: []Match{
					{White: "p1", IsBye: true},
				}}}
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 1, Played: 1},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 0, Played: 0},
			},
		},
		{
			name: "draws give half a point to both",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
				}}
				t.Rounds = []Round{{Number: 1, Matches: []Match{
					{White: "p1", Black: "p2", Result: "0.5-0.5"},
				}}}
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 0.5, Played: 1},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 0.5, Played: 1},
			},
		},
		{
			name: "identical scores order by numeric id asc",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p10", Name: "Diez"},
					{ID: "p9", Name: "Nueve"},
					{ID: "p2", Name: "Dos"},
				}}
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p2", Name: "Dos", Points: 0, Played: 0},
				{Rank: 2, ID: "p9", Name: "Nueve", Points: 0, Played: 0},
				{Rank: 3, ID: "p10", Name: "Diez", Points: 0, Played: 0},
			},
		},
		{
			name: "unreported matches score zero and do not count as played",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
				}}
				t.Rounds = []Round{{Number: 1, Matches: []Match{
					{White: "p1", Black: "p2"},
				}}}
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 0, Played: 0},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 0, Played: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.setup().Standings()
			if len(got) != len(tt.want) {
				t.Fatalf("Standings() = %d rows, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("row %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestStandingsPureRead(t *testing.T) {
	tour := &Tournament{Players: []Player{{ID: "p1", Name: "Ana"}, {ID: "p2", Name: "Beto"}}}
	tour.Rounds = []Round{{Number: 1, Matches: []Match{
		{White: "p1", Black: "p2", Result: "1-0"},
	}}}
	before := len(tour.Players)
	_ = tour.Standings()
	if len(tour.Players) != before || len(tour.Rounds) != 1 {
		t.Error("Standings mutated the tournament")
	}
}

func TestRunStandingsOutput(t *testing.T) {
	path := writeStateFile(t, `{
	  "name": "T", "players": [
	    {"id": "p1", "name": "Ana"},
	    {"id": "p2", "name": "Beto"}
	  ],
	  "rounds": [{"number": 1, "matches": [
	    {"white": "p1", "black": "p2", "result": "1-0"},
	    {"white": "p3", "isBye": true}
	  ]}]
	}`)
	var out, errBuf strings.Builder
	code := run([]string{"standings"}, path, &out, &errBuf)
	if code != 0 {
		t.Fatalf("run standings exit = %d, stderr = %q", code, errBuf.String())
	}
	want := "1\tAna\t1\t1\n2\tBeto\t0\t1\n"
	if out.String() != want {
		t.Errorf("output = %q, want %q", out.String(), want)
	}
}

func TestRunStandingsEmptyOutput(t *testing.T) {
	path := writeStateFile(t, `{"name": "T"}`)
	var out, errBuf strings.Builder
	code := run([]string{"standings"}, path, &out, &errBuf)
	if code != 0 {
		t.Fatalf("run standings exit = %d, stderr = %q", code, errBuf.String())
	}
	if out.String() != "(no players yet)\n" {
		t.Errorf("output = %q, want empty-tournament message", out.String())
	}
}
