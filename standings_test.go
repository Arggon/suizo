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
				{Rank: 1, ID: "p1", Name: "Ana", Points: 2, Played: 2, Buchholz: 2, BuchholzCut1: 1},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 1, Played: 2, Buchholz: 2},
				{Rank: 3, ID: "p3", Name: "Carla", Points: 1, Played: 2, Buchholz: 2},
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
				{Rank: 1, ID: "p1", Name: "Ana", Points: 0.5, Played: 1, Buchholz: 0.5, Direct: 0.5},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 0.5, Played: 1, Buchholz: 0.5, Direct: 0.5},
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

func TestStandingsTieBreakers(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Tournament
		want  []Standing
	}{
		{
			name: "hand-computed Buchholz and Cut 1",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
					{ID: "p3", Name: "Carla"},
					{ID: "p4", Name: "Diego"},
				}}
				t.Rounds = []Round{
					{Number: 1, Matches: []Match{
						{White: "p1", Black: "p2", Result: "1-0"},
						{White: "p3", Black: "p4", Result: "0-1"},
					}},
					{Number: 2, Matches: []Match{
						{White: "p1", Black: "p3", Result: "1-0"},
						{White: "p2", Black: "p4", Result: "1-0"},
					}},
				}
				// Scores: p1=2, p2=1, p4=1, p3=0. Tie group at 1 is {p2,p4}.
				// p1: opps p2(1)+p3(0) -> BH 1, cut1 1
				// p2: opps p1(2)+p4(1) -> BH 3, cut1 2, direct 1 (beat p4)
				// p4: opps p3(0)+p2(1) -> BH 1, cut1 1, direct 0
				// p3: opps p4(1)+p1(2) -> BH 3, cut1 2
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 2, Played: 2, Buchholz: 1, BuchholzCut1: 1},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 1, Played: 2, Buchholz: 3, BuchholzCut1: 2, Direct: 1},
				{Rank: 3, ID: "p4", Name: "Diego", Points: 1, Played: 2, Buchholz: 1, BuchholzCut1: 1},
				{Rank: 4, ID: "p3", Name: "Carla", Points: 0, Played: 2, Buchholz: 3, BuchholzCut1: 2},
			},
		},
		{
			name: "Buchholz decides over numeric id fallback",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
					{ID: "p3", Name: "Carla"},
					{ID: "p4", Name: "Diego"},
					{ID: "p5", Name: "Elena"},
				}}
				t.Rounds = []Round{
					{Number: 1, Matches: []Match{
						{White: "p1", Black: "p2", Result: "1-0"},
						{White: "p3", Black: "p4", Result: "1-0"},
						{White: "p5", IsBye: true},
					}},
					{Number: 2, Matches: []Match{
						{White: "p1", Black: "p3", Result: "1-0"},
						{White: "p2", Black: "p4", Result: "1-0"},
						{White: "p5", IsBye: true},
					}},
				}
				// Scores: p1=2, p5=2 (two byes), p2=1, p3=1, p4=0.
				// p1 and p5 tie on 2 points; id fallback alone would put p5
				// first, but Buchholz decides: p1's opponents p2(1)+p3(1)
				// give BH 2, while p5's byes give no opponents at all (BH 0).
				// p2 and p3 tie at 1 with equal BH/cut1 and never met, so
				// their order falls through to the numeric id.
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 2, Played: 2, Buchholz: 2, BuchholzCut1: 1},
				{Rank: 2, ID: "p5", Name: "Elena", Points: 2, Played: 2},
				{Rank: 3, ID: "p2", Name: "Beto", Points: 1, Played: 2, Buchholz: 2, BuchholzCut1: 2},
				{Rank: 4, ID: "p3", Name: "Carla", Points: 1, Played: 2, Buchholz: 2, BuchholzCut1: 2},
				{Rank: 5, ID: "p4", Name: "Diego", Points: 0, Played: 2, Buchholz: 2, BuchholzCut1: 1},
			},
		},
		{
			name: "direct encounter decides between two tied players",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p1", Name: "Ana"},
					{ID: "p4", Name: "Diego"},
					{ID: "p5", Name: "Elena"},
					{ID: "p6", Name: "Fede"},
				}}
				t.Rounds = []Round{
					{Number: 1, Matches: []Match{
						{White: "p1", Black: "p6", Result: "1-0"},
						{White: "p5", Black: "p4", Result: "1-0"},
					}},
					{Number: 2, Matches: []Match{
						{White: "p1", Black: "p5", Result: "1-0"},
						{White: "p4", Black: "p6", Result: "1-0"},
					}},
					{Number: 3, Matches: []Match{
						{White: "p6", Black: "p5", Result: "1-0"},
						{White: "p4", Black: "p1", Result: "1-0"},
					}},
				}
				// Scores: p1=2, p4=2, p5=1, p6=1. Groups: {p1,p4} and {p5,p6}.
				// At 2 points BH is 4 (cut1 3) for both, so the direct
				// encounter decides: p4 beat p1 -> p4 first. At 1 point both
				// have BH 5 (cut1 4), and p6 beat p5 -> direct 1 vs 0 puts
				// p6 first.
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p4", Name: "Diego", Points: 2, Played: 3, Buchholz: 4, BuchholzCut1: 3, Direct: 1},
				{Rank: 2, ID: "p1", Name: "Ana", Points: 2, Played: 3, Buchholz: 4, BuchholzCut1: 3},
				{Rank: 3, ID: "p6", Name: "Fede", Points: 1, Played: 3, Buchholz: 5, BuchholzCut1: 4, Direct: 1},
				{Rank: 4, ID: "p5", Name: "Elena", Points: 1, Played: 3, Buchholz: 5, BuchholzCut1: 4},
			},
		},
		{
			name: "bye boards contribute no opponent to Buchholz",
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
				// p1: opps p2(1)+p3(1) -> BH 2, cut1 1. p2 and p3 each got a
				// bye, but byes add no phantom opponent: their only opponent
				// is p1(2) -> BH 2,
				// cut1 0 (fewer than 2 opponents drops nothing... it is 0);
				// they never met, so direct is 0 for both and the numeric
				// id decides: p2 before p3.
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 2, Played: 2, Buchholz: 2, BuchholzCut1: 1},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 1, Played: 2, Buchholz: 2},
				{Rank: 3, ID: "p3", Name: "Carla", Points: 1, Played: 2, Buchholz: 2},
			},
		},
		{
			name: "all metrics equal order by numeric id asc",
			setup: func() *Tournament {
				t := &Tournament{Players: []Player{
					{ID: "p3", Name: "Carla"},
					{ID: "p10", Name: "Diez"},
					{ID: "p1", Name: "Ana"},
					{ID: "p2", Name: "Beto"},
				}}
				t.Rounds = []Round{
					{Number: 1, Matches: []Match{
						{White: "p1", Black: "p2", Result: "1-0"},
						{White: "p3", Black: "p10", Result: "1-0"},
					}},
					{Number: 2, Matches: []Match{
						{White: "p2", Black: "p3", Result: "1-0"},
						{White: "p10", Black: "p1", Result: "1-0"},
					}},
				}
				// Every player: 1 point, BH 2, cut1 1, direct 1 -> ids
				// p1, p2, p3, p10 (numeric order).
				return t
			},
			want: []Standing{
				{Rank: 1, ID: "p1", Name: "Ana", Points: 1, Played: 2, Buchholz: 2, BuchholzCut1: 1, Direct: 1},
				{Rank: 2, ID: "p2", Name: "Beto", Points: 1, Played: 2, Buchholz: 2, BuchholzCut1: 1, Direct: 1},
				{Rank: 3, ID: "p3", Name: "Carla", Points: 1, Played: 2, Buchholz: 2, BuchholzCut1: 1, Direct: 1},
				{Rank: 4, ID: "p10", Name: "Diez", Points: 1, Played: 2, Buchholz: 2, BuchholzCut1: 1, Direct: 1},
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
	want := "1\tAna\t1\t1\t0\t0\t0\n2\tBeto\t0\t1\t1\t0\t0\n"
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
