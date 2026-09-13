package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// roundsFixture builds a tournament with the given players and loads it into
// a temp-dir store, so tests exercise the real persistence path.
func roundsFixture(t *testing.T, names ...string) (*store, *Tournament) {
	t.Helper()
	s, err := openStore(t.TempDir() + "/suizo.json")
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	tm := &Tournament{Name: "Suizo tournament"}
	for _, n := range names {
		if _, err := tm.addPlayer(n); err != nil {
			t.Fatalf("addPlayer(%q): %v", n, err)
		}
	}
	if err := s.save(tm); err != nil {
		t.Fatalf("save: %v", err)
	}
	return s, tm
}

func TestStartRoundLifecycle(t *testing.T) {
	tests := []struct {
		name       string
		results    []string // "round board result" triples to report after round 1 opens
		wantErr    error    // error from the second StartRound attempt (nil = allowed)
		wantRounds int      // tournament round count after the full sequence
	}{
		{
			name:       "fresh tournament blocks the second start with no results",
			results:    nil,
			wantErr:    errRoundPending,
			wantRounds: 1,
		},
		{
			name:       "pending boards block the next round",
			results:    nil,
			wantErr:    errRoundPending,
			wantRounds: 1,
		},
		{
			name:       "partial results still block the next round",
			results:    []string{"1 1 1-0"},
			wantErr:    errRoundPending,
			wantRounds: 1,
		},
		{
			name:       "complete results allow the next round",
			results:    []string{"1 1 1-0", "1 2 0.5-0.5", "1 3 0-1", "1 4 1-0"},
			wantErr:    nil,
			wantRounds: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, tm := roundsFixture(t, "Ana", "Beto", "Carla", "Dan", "Eva", "Fede", "Gina", "Hugo")
			n, ms, err := tm.StartRound()
			if err != nil {
				t.Fatalf("first StartRound: %v", err)
			}
			if n != 1 || len(ms) == 0 {
				t.Fatalf("first StartRound got round %d with %d boards", n, len(ms))
			}
			for _, r := range tt.results {
				parts := strings.Fields(r)
				round, board := atoi(t, parts[0]), atoi(t, parts[1])
				if err := tm.SetResult(round, board, parts[2]); err != nil {
					t.Fatalf("SetResult(%d,%d,%s): %v", round, board, parts[2], err)
				}
			}
			// The blocked/allowed second attempt uses a copy so the final
			// persisted state below always comes from the first attempt path.
			probe := *tm
			probe.Rounds = append([]Round(nil), tm.Rounds...)
			_, _, err = probe.StartRound()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("second StartRound error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if _, _, err := tm.StartRound(); err != nil {
					t.Fatalf("second StartRound: %v", err)
				}
			}
			if err := s.save(tm); err != nil {
				t.Fatalf("save: %v", err)
			}
			got, err := s.load()
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if len(got.Rounds) != tt.wantRounds {
				t.Fatalf("rounds persisted = %d, want %d", len(got.Rounds), tt.wantRounds)
			}
		})
	}
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Fatalf("atoi(%q): not a number", s)
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func TestSetResultValidation(t *testing.T) {
	fresh := func(t *testing.T) *Tournament {
		t.Helper()
		_, tm := roundsFixture(t, "Ana", "Beto", "Carla", "Dan")
		if _, _, err := tm.StartRound(); err != nil { // round 1: 2 boards
			t.Fatalf("StartRound: %v", err)
		}
		return tm
	}
	tests := []struct {
		name    string
		round   int
		board   int
		result  string
		wantErr string // empty = success; "any" = any non-nil error
	}{
		{name: "win for white", round: 1, board: 1, result: "1-0"},
		{name: "win for black", round: 1, board: 1, result: "0-1"},
		{name: "draw", round: 1, board: 2, result: "0.5-0.5"},
		{name: "illegal value", round: 1, board: 1, result: "2-0", wantErr: "result must be one of"},
		{name: "empty value", round: 1, board: 1, result: "", wantErr: "result must be one of"},
		{name: "unknown round", round: 9, board: 1, result: "1-0", wantErr: "any"},
		{name: "unknown board", round: 1, board: 9, result: "1-0", wantErr: "any"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := fresh(t)
			err := tm.SetResult(tt.round, tt.board, tt.result)
			switch tt.wantErr {
			case "":
				if err != nil {
					t.Fatalf("SetResult(%d,%d,%q): %v", tt.round, tt.board, tt.result, err)
				}
			case "any":
				if err == nil {
					t.Fatalf("SetResult(%d,%d,%q): want error, got nil", tt.round, tt.board, tt.result)
				}
			default:
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("SetResult(%d,%d,%q) error = %v, want containing %q", tt.round, tt.board, tt.result, err, tt.wantErr)
				}
			}
		})
	}
}

func TestSetResultUnknownBoard(t *testing.T) {
	_, tm := roundsFixture(t, "Ana", "Beto", "Carla", "Dan")
	if _, _, err := tm.StartRound(); err != nil {
		t.Fatalf("StartRound: %v", err)
	}
	for _, board := range []int{0, 3, -1} {
		if err := tm.SetResult(1, board, "1-0"); err == nil {
			t.Fatalf("SetResult board %d: want error, got nil", board)
		}
	}
}

func TestRoundStatus(t *testing.T) {
	tests := []struct {
		name        string
		results     []string // triples reported on round 1
		query       int      // round passed to RoundStatus (0 = latest)
		wantDone    int
		wantPending int
		wantErr     bool
	}{
		{name: "all game boards pending, bye done by default", query: 0, wantDone: 1, wantPending: 2},
		{name: "partial results", results: []string{"1 1 1-0"}, query: 0, wantDone: 2, wantPending: 1},
		{name: "explicit round number", results: []string{"1 1 0-1", "1 2 0.5-0.5"}, query: 1, wantDone: 3, wantPending: 0},
		{name: "bye boards count as done", results: []string{"1 1 1-0", "1 2 1-0"}, query: 0, wantDone: 3, wantPending: 0},
		{name: "unknown round errors", query: 5, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, tm := roundsFixture(t, "Ana", "Beto", "Carla", "Dan", "Eva")
			// 5 players -> 2 boards + 1 bye in round 1.
			if _, _, err := tm.StartRound(); err != nil {
				t.Fatalf("StartRound: %v", err)
			}
			if err := s.save(tm); err != nil {
				t.Fatalf("save: %v", err)
			}
			for _, r := range tt.results {
				parts := strings.Fields(r)
				if err := tm.SetResult(atoi(t, parts[0]), atoi(t, parts[1]), parts[2]); err != nil {
					t.Fatalf("SetResult(%s): %v", r, err)
				}
			}
			status, err := tm.RoundStatus(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatal("RoundStatus: want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("RoundStatus: %v", err)
			}
			done, pend := 0, 0
			for _, b := range status {
				if b.Done {
					done++
				} else {
					pend++
				}
			}
			if done != tt.wantDone || pend != tt.wantPending {
				t.Fatalf("status = %d done / %d pending, want %d / %d", done, pend, tt.wantDone, tt.wantPending)
			}
		})
	}
}

func TestRoundStatusEmptyTournament(t *testing.T) {
	_, tm := roundsFixture(t)
	if _, err := tm.RoundStatus(0); !errors.Is(err, errNoRounds) {
		t.Fatalf("RoundStatus on empty tournament = %v, want %v", err, errNoRounds)
	}
}

func TestRoundsCLI(t *testing.T) {
	path := t.TempDir() + "/suizo.json"
	var out, errOut bytes.Buffer
	for _, name := range []string{"Ana", "Beto", "Carla", "Dan", "Eva"} {
		if code := run([]string{"players", "add", name}, path, &out, &out); code != 0 {
			t.Fatalf("players add %s: exit %d", name, code)
		}
	}
	if code := run([]string{"rounds", "start"}, path, &out, &errOut); code != 0 {
		t.Fatalf("rounds start: exit %d, stderr %q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Round 1") {
		t.Fatalf("rounds start output %q lacks round header", out.String())
	}
	// Status with no round argument targets the latest round.
	out.Reset()
	if code := run([]string{"rounds", "status"}, path, &out, &errOut); code != 0 {
		t.Fatalf("rounds status: exit %d, stderr %q", code, errOut.String())
	}
	if strings.Count(out.String(), "pending") != 2 {
		t.Fatalf("rounds status output %q: want 2 pending boards", out.String())
	}
	// Reporting board 1 lets a second start through only after all boards close.
	if code := run([]string{"results", "1", "1", "1-0"}, path, &out, &errOut); code != 0 {
		t.Fatalf("results: exit %d, stderr %q", code, errOut.String())
	}
	errOut.Reset()
	if code := run([]string{"rounds", "start"}, path, &out, &errOut); code != 1 {
		t.Fatalf("rounds start with pending boards: exit %d, want 1", code)
	}
	if !strings.Contains(errOut.String(), "pending") {
		t.Fatalf("blocked start stderr %q lacks pending message", errOut.String())
	}
	// Illegal result values are refused with exit 1.
	for _, res := range []string{"2-0", "draw", ""} {
		errOut.Reset()
		if code := run([]string{"results", "1", "2", res}, path, &out, &errOut); code != 1 {
			t.Fatalf("results %q: exit %d, want 1", res, code)
		}
	}
	// Unknown round and board are refused.
	if code := run([]string{"results", "7", "1", "1-0"}, path, &out, &errOut); code != 1 {
		t.Fatalf("results unknown round: exit %d, want 1", code)
	}
	if code := run([]string{"results", "1", "9", "1-0"}, path, &out, &errOut); code != 1 {
		t.Fatalf("results unknown board: exit %d, want 1", code)
	}
	// Finish the remaining game board (board 3 is the bye, already done), then
	// the next start succeeds.
	if code := run([]string{"results", "1", "2", "0.5-0.5"}, path, &out, &errOut); code != 0 {
		t.Fatalf("results board 2: exit %d, stderr %q", code, errOut.String())
	}
	out.Reset()
	if code := run([]string{"rounds", "start"}, path, &out, &errOut); code != 0 {
		t.Fatalf("rounds start after completion: exit %d, stderr %q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Round 2") {
		t.Fatalf("second start output %q lacks Round 2", out.String())
	}
	// rounds status with an explicit round number.
	out.Reset()
	if code := run([]string{"rounds", "status", "1"}, path, &out, &errOut); code != 0 {
		t.Fatalf("rounds status 1: exit %d", code)
	}
	if strings.Count(out.String(), "done") != 3 {
		t.Fatalf("rounds status 1 output %q: want 3 done boards", out.String())
	}
}
