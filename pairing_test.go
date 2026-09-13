package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pairTestTournament builds a tournament with n players and the given rounds.
func pairTestTournament(t *testing.T, n int, rounds ...Round) *Tournament {
	t.Helper()
	tr := &Tournament{Name: "test"}
	for i := 0; i < n; i++ {
		if _, err := tr.addPlayer(fmt.Sprintf("P%d", i+1)); err != nil {
			t.Fatalf("addPlayer: %v", err)
		}
	}
	tr.Rounds = rounds
	return tr
}

// game is shorthand for a played match.
func game(white, black, result string) Match {
	return Match{White: white, Black: black, Result: result}
}

func TestPairingsFreshEvenCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{"2 players", 2, "p1-p2"},
		{"4 players", 4, "p1-p2 p3-p4"},
		{"6 players", 6, "p1-p2 p3-p4 p5-p6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := pairTestTournament(t, tt.count)
			ms, err := tr.Pairings()
			if err != nil {
				t.Fatalf("Pairings: %v", err)
			}
			if got := pairKey(ms); got != tt.want {
				t.Fatalf("pairings = %q, want %q", got, tt.want)
			}
		})
	}
}

// pairKey renders matches as sorted "w-b" pairs joined by spaces, so tests
// compare boards independent of color balancing.
func pairKey(ms []Match) string {
	parts := make([]string, 0, len(ms))
	for _, m := range ms {
		if m.IsBye {
			parts = append(parts, m.White+"-bye")
			continue
		}
		a, b := m.White, m.Black
		if playerIDLess(b, a) {
			a, b = b, a
		}
		parts = append(parts, a+"-"+b)
	}
	return strings.Join(parts, " ")
}

func TestPairingsScoreGroupsSecondRound(t *testing.T) {
	// Round 1: p1 beats p2, p3 beats p4. Winners (p1, p3) must meet in round 2.
	tr := pairTestTournament(t, 4, Round{Number: 1, Matches: []Match{
		game("p1", "p2", "1-0"),
		game("p3", "p4", "1-0"),
	}})
	ms, err := tr.Pairings()
	if err != nil {
		t.Fatalf("Pairings: %v", err)
	}
	if got := pairKey(ms); got != "p1-p3 p2-p4" {
		t.Fatalf("round 2 pairings = %q, want %q", got, "p1-p3 p2-p4")
	}
}

func TestPairingsDrawsSplitGroups(t *testing.T) {
	// p1 draws p2 -> both on 0.5; p3 beats p4 -> p3 alone at 1, p4 at 0.
	// Score groups: {p3} floats down into {p1, p2}, then {p4} follows.
	tr := pairTestTournament(t, 4, Round{Number: 1, Matches: []Match{
		game("p1", "p2", "0.5-0.5"),
		game("p3", "p4", "1-0"),
	}})
	ms, err := tr.Pairings()
	if err != nil {
		t.Fatalf("Pairings: %v", err)
	}
	if got := pairKey(ms); got != "p1-p3 p2-p4" {
		t.Fatalf("pairings = %q, want %q", got, "p1-p3 p2-p4")
	}
}

func TestScoreComputation(t *testing.T) {
	tests := []struct {
		name   string
		rounds []Round
		id     string
		want   float64
	}{
		{"win", []Round{{Matches: []Match{game("p1", "p2", "1-0")}}}, "p1", 1},
		{"loss", []Round{{Matches: []Match{game("p1", "p2", "1-0")}}}, "p2", 0},
		{"black win", []Round{{Matches: []Match{game("p1", "p2", "0-1")}}}, "p2", 1},
		{"draw", []Round{{Matches: []Match{game("p1", "p2", "0.5-0.5")}}}, "p1", 0.5},
		{"bye", []Round{{Matches: []Match{{White: "p1", IsBye: true}}}}, "p1", 1},
		{"accumulates", []Round{
			{Matches: []Match{game("p1", "p2", "1-0")}},
			{Matches: []Match{game("p1", "p3", "0.5-0.5")}},
		}, "p1", 1.5},
		{"unreported match counts 0", []Round{{Matches: []Match{{White: "p1", Black: "p2"}}}}, "p1", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := pairTestTournament(t, 4, tt.rounds...)
			if got := tr.score(tt.id); got != tt.want {
				t.Fatalf("score(%s) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestPairingsNoRematches(t *testing.T) {
	// Deterministic pseudo-random small tournaments: after each computed round
	// is folded back in, no pair may ever meet twice while space remains.
	for seed := int64(1); seed <= 20; seed++ {
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			n := 8 + rng.Intn(8) // 8-15 players
			tr := pairTestTournament(t, n)
			for round := 0; round < 3+rng.Intn(3); round++ { // 3-5 rounds
				ms, err := tr.Pairings()
				if errors.Is(err, errNoRematchFreePairing) || errors.Is(err, errNoByeCandidate) {
					// Spec-sanctioned outcome: a rematch is forced once every
					// candidate opponent has already been met.
					return
				}
				if err != nil {
					t.Fatalf("round %d: Pairings: %v", round+1, err)
				}
				assertNoRematch(t, tr, ms)
				// Fold results in: white wins decided by rng, byes score 1.
				fold := Round{Number: round + 1, Matches: ms}
				for i := range fold.Matches {
					m := &fold.Matches[i]
					if m.IsBye {
						continue
					}
					if rng.Intn(2) == 0 {
						m.Result = "1-0"
					} else {
						m.Result = "0-1"
					}
				}
				tr.Rounds = append(tr.Rounds, fold)
			}
		})
	}
}

// assertNoRematch fails if ms repeats any earlier pairing within tr, or
// internally duplicates a pair.
func assertNoRematch(t *testing.T, tr *Tournament, ms []Match) {
	t.Helper()
	seen := make(map[string]bool)
	for _, r := range tr.Rounds {
		for _, m := range r.Matches {
			if !m.IsBye {
				seen[unorderedPair(m.White, m.Black)] = true
			}
		}
	}
	for _, m := range ms {
		if m.IsBye {
			continue
		}
		key := unorderedPair(m.White, m.Black)
		if seen[key] {
			t.Fatalf("rematch produced: %s", key)
		}
		seen[key] = true
	}
}

func unorderedPair(a, b string) string {
	if playerIDLess(b, a) {
		a, b = b, a
	}
	return a + "|" + b
}

func TestPairingsForcedRematchErrors(t *testing.T) {
	// Two players who already met: no rematch-free assignment exists.
	tr := pairTestTournament(t, 2, Round{Number: 1, Matches: []Match{
		game("p1", "p2", "1-0"),
	}})
	if _, err := tr.Pairings(); err == nil {
		t.Fatal("Pairings with forced rematch: expected error, got nil")
	}
}

func TestPairingsBye(t *testing.T) {
	t.Run("odd count gives bye to lowest ranked without previous bye", func(t *testing.T) {
		tr := pairTestTournament(t, 3)
		ms, err := tr.Pairings()
		if err != nil {
			t.Fatalf("Pairings: %v", err)
		}
		if got := pairKey(ms); got != "p1-p2 p3-bye" {
			t.Fatalf("pairings = %q, want %q", got, "p1-p2 p3-bye")
		}
	})
	t.Run("second round never repeats the bye", func(t *testing.T) {
		tr := pairTestTournament(t, 3, Round{Number: 1, Matches: []Match{
			game("p1", "p2", "1-0"),
			{White: "p3", IsBye: true},
		}})
		ms, err := tr.Pairings()
		if err != nil {
			t.Fatalf("Pairings: %v", err)
		}
		byes := 0
		for _, m := range ms {
			if m.IsBye {
				byes++
				if m.White == "p3" {
					t.Fatal("p3 received a second bye")
				}
			}
		}
		if byes != 1 {
			t.Fatalf("got %d byes, want exactly 1", byes)
		}
	})
	t.Run("double bye for one player impossible", func(t *testing.T) {
		for n := 3; n <= 9; n += 2 {
			tr := pairTestTournament(t, n)
			for round := 1; round <= 2; round++ {
				ms, err := tr.Pairings()
				if err != nil {
					t.Fatalf("n=%d round=%d: %v", n, round, err)
				}
				byes := 0
				for _, m := range ms {
					if !m.IsBye {
						continue
					}
					byes++
					if tr.hadBye(m.White) {
						t.Fatalf("n=%d round=%d: %s got a second bye", n, round, m.White)
					}
				}
				if byes != 1 {
					t.Fatalf("n=%d round=%d: %d byes, want 1", n, round, byes)
				}
				tr.Rounds = append(tr.Rounds, Round{Number: round, Matches: ms})
			}
		}
	})
}

func TestPairingsTooFewPlayers(t *testing.T) {
	tests := []struct {
		name   string
		rounds []Round
	}{
		{"zero players", nil},
		{"one player", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := 0
			if tt.name == "one player" {
				n = 1
			}
			tr := pairTestTournament(t, n, tt.rounds...)
			ms, err := tr.Pairings()
			if err == nil {
				t.Fatalf("Pairings = %v, want error", ms)
			}
		})
	}
}

func TestPairingsDoesNotMutate(t *testing.T) {
	tr := pairTestTournament(t, 4)
	before := len(tr.Rounds)
	if _, err := tr.Pairings(); err != nil {
		t.Fatalf("Pairings: %v", err)
	}
	if len(tr.Rounds) != before {
		t.Fatalf("Pairings mutated Rounds: %d -> %d", before, len(tr.Rounds))
	}
}

func TestPairingsDeterministic(t *testing.T) {
	tr := pairTestTournament(t, 5)
	a, err := tr.Pairings()
	if err != nil {
		t.Fatalf("Pairings: %v", err)
	}
	b, err := tr.Pairings()
	if err != nil {
		t.Fatalf("Pairings: %v", err)
	}
	if pairKey(a) != pairKey(b) {
		t.Fatalf("Pairings not deterministic: %q vs %q", pairKey(a), pairKey(b))
	}
}

func TestPairingsColorAlternation(t *testing.T) {
	// p1 had white last round; paired again it should get black when the
	// opponent did not.
	tr := pairTestTournament(t, 4, Round{Number: 1, Matches: []Match{
		game("p1", "p3", "1-0"),
		game("p2", "p4", "0-1"),
	}})
	ms, err := tr.Pairings()
	if err != nil {
		t.Fatalf("Pairings: %v", err)
	}
	for _, m := range ms {
		if m.White == "p1" {
			t.Fatalf("p1 got white twice in a row (%s vs %s)", m.White, m.Black)
		}
	}
}

func TestRunPairEndToEnd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suizo.json")
	stdout, stderr := &strings.Builder{}, &strings.Builder{}
	if code := run([]string{"players", "add", "Ana"}, path, stdout, stderr); code != 0 {
		t.Fatalf("players add: code=%d stderr=%q", code, stderr.String())
	}
	for _, name := range []string{"Beto", "Carla", "Dana"} {
		if code := run([]string{"players", "add", name}, path, stdout, stderr); code != 0 {
			t.Fatalf("players add %s: code=%d", name, code)
		}
	}
	stdout.Reset()
	if code := run([]string{"pair"}, path, stdout, &strings.Builder{}); code != 0 {
		t.Fatalf("pair: code=%d stderr=%q", code, stderr.String())
	}
	want := "Round 1\nBoard: p1 vs p2\nBoard: p3 vs p4\n"
	if got := stdout.String(); got != want {
		t.Fatalf("pair output = %q, want %q", got, want)
	}
	// Round 1 must be persisted in the store file.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	if !strings.Contains(string(data), `"number": 1`) {
		t.Fatalf("store file missing round 1:\n%s", data)
	}
}
