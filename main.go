// Command suizo is a Swiss-system tournament manager for amateur events —
// chess, futbolito, anything pairable. State lives in a single JSON file.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

const usage = `suizo - Swiss-system tournament manager

Usage:
  suizo players add <name>   Register a player; prints its id
  suizo players list         List registered players
  suizo pair                 Pair the next round, persist it and print boards
  suizo rounds start         Open the next round: refuse while the last round has pending matches
  suizo rounds status [N]    Show per-board pending/done for round N (default: latest)
  suizo results <round> <board> <1-0|0-1|0.5-0.5>
                             Report a result; board is 1-based within the round
  suizo standings            Print the score table
  suizo serve [--addr host:port]
                             Serve the local web UI (default 127.0.0.1:8080,
                             override with --addr or $SUIZO_ADDR); loopback only

State file: $SUIZO_FILE (default ./suizo.json)
`

func main() {
	os.Exit(run(os.Args[1:], storePath(), os.Stdout, os.Stderr))
}

func storePath() string {
	if p := os.Getenv("SUIZO_FILE"); p != "" {
		return p
	}
	return "suizo.json"
}

func run(args []string, path string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	s, err := openStore(path)
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	switch args[0] {
	case "players":
		return runPlayers(args[1:], s, stdout, stderr)
	case "pair":
		return runPair(s, stdout, stderr)
	case "rounds":
		return runRounds(args[1:], s, stdout, stderr)
	case "results":
		return runResults(args[1:], s, stderr)
	case "standings":
		return runStandings(s, stdout, stderr)
	case "serve":
		return runServe(args[1:], s, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "suizo: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func runPlayers(args []string, s *store, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "add":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: suizo players add <name>")
			return 2
		}
		var id string
		err := s.update(func(t *Tournament) error {
			p, err := t.addPlayer(args[1])
			if err != nil {
				return err
			}
			id = p.ID
			return nil
		})
		if err != nil {
			fmt.Fprintf(stderr, "suizo: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "%s\n", id)
		return 0
	case "list":
		t, err := s.load()
		if err != nil {
			fmt.Fprintf(stderr, "suizo: %v\n", err)
			return 1
		}
		if len(t.Players) == 0 {
			fmt.Fprintln(stdout, "(no players yet)")
			return 0
		}
		for _, p := range t.Players {
			fmt.Fprintf(stdout, "%s\t%s\n", p.ID, p.Name)
		}
		return 0
	default:
		fmt.Fprintf(stderr, "suizo: unknown players command %q\n\n%s", args[0], usage)
		return 2
	}
}

// runPair computes the next round's pairings, appends them as round N+1 and
// saves through the locked store, then prints one board per line.
func runPair(s *store, stdout, stderr io.Writer) int {
	var matches []Match
	var round int
	err := s.update(func(t *Tournament) error {
		ms, err := t.Pairings()
		if err != nil {
			return err
		}
		matches = ms
		round = len(t.Rounds) + 1
		t.Rounds = append(t.Rounds, Round{Number: round, Matches: ms})
		return nil
	})
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Round %d\n", round)
	for _, m := range matches {
		if m.IsBye {
			fmt.Fprintf(stdout, "Board: %s gets a bye\n", m.White)
			continue
		}
		fmt.Fprintf(stdout, "Board: %s vs %s\n", m.White, m.Black)
	}
	return 0
}

func runRounds(args []string, s *store, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "suizo: unknown command %q (want \"rounds start\" or \"rounds status\")\n\n%s", "rounds", usage)
		return 2
	}
	switch args[0] {
	case "start":
		return runRoundsStart(s, stdout, stderr)
	case "status":
		return runRoundsStatus(args[1:], s, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "suizo: unknown rounds command %q\n\n%s", args[0], usage)
		return 2
	}
}

// runRoundsStart opens the next round and prints its boards, refusing while
// the last round still has pending matches.
func runRoundsStart(s *store, stdout, stderr io.Writer) int {
	var round int
	var matches []Match
	err := s.update(func(t *Tournament) error {
		n, ms, err := t.StartRound()
		if err != nil {
			return err
		}
		round, matches = n, ms
		return nil
	})
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	printBoards(stdout, round, matches)
	return 0
}

// runRoundsStatus prints per-board pending/done for the latest round, or the
// given round number.
func runRoundsStatus(args []string, s *store, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintln(stderr, "usage: suizo rounds status [round]")
		return 2
	}
	round := 0
	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "suizo: invalid round %q\n", args[0])
			return 2
		}
		round = n
	}
	t, err := s.load()
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	status, err := t.RoundStatus(round)
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Round %d\n", roundNumber(t.Rounds, round))
	for _, b := range status {
		state := "pending"
		if b.Done {
			state = "done"
		}
		switch {
		case b.Match.IsBye:
			fmt.Fprintf(stdout, "Board %d: %s bye\t%s\n", b.Board, b.Match.White, state)
		default:
			fmt.Fprintf(stdout, "Board %d: %s vs %s\t%s\n", b.Board, b.Match.White, b.Match.Black, state)
		}
	}
	return 0
}

// runStandings prints the score table: Rk, Name, Pts, matches played and the
// tie-breaker columns (Buchholz, Buchholz Cut 1, direct encounter).
func runStandings(s *store, stdout, stderr io.Writer) int {
	t, err := s.load()
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	rows := t.Standings()
	if len(rows) == 0 {
		fmt.Fprintln(stdout, "(no players yet)")
		return 0
	}
	for _, r := range rows {
		fmt.Fprintf(stdout, "%d\t%s\t%s\t%d\t%s\t%s\t%s\n",
			r.Rank, r.Name, formatPoints(r.Points), r.Played,
			formatPoints(r.Buchholz), formatPoints(r.BuchholzCut1), formatPoints(r.Direct))
	}
	return 0
}

// runResults validates its arguments and persists one board result.
func runResults(args []string, s *store, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: suizo results <round> <board> <1-0|0-1|0.5-0.5>")
		return 2
	}
	round, err1 := strconv.Atoi(args[0])
	board, err2 := strconv.Atoi(args[1])
	if err1 != nil || err2 != nil {
		fmt.Fprintln(stderr, "usage: suizo results <round> <board> <1-0|0-1|0.5-0.5>")
		return 2
	}
	err := s.update(func(t *Tournament) error {
		return t.SetResult(round, board, args[2])
	})
	if err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	return 0
}

// printBoards writes one board per line for the given round.
func printBoards(stdout io.Writer, round int, matches []Match) {
	fmt.Fprintf(stdout, "Round %d\n", round)
	for _, m := range matches {
		if m.IsBye {
			fmt.Fprintf(stdout, "Board: %s gets a bye\n", m.White)
			continue
		}
		fmt.Fprintf(stdout, "Board: %s vs %s\n", m.White, m.Black)
	}
}

// roundNumber resolves the display number for a status listing: the requested
// round, or the latest one when round is 0. Callers pass only validated input.
func roundNumber(rounds []Round, requested int) int {
	if requested > 0 {
		return requested
	}
	return rounds[len(rounds)-1].Number
}

// formatPoints renders a score without a trailing .0: 1, 1.5, 0.5.
func formatPoints(p float64) string {
	return strconv.FormatFloat(p, 'f', -1, 64)
}
