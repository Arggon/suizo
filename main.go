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
  suizo standings            Print the score table

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
	case "standings":
		return runStandings(s, stdout, stderr)
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

// runStandings prints the score table: Rk, Name, Pts and matches played.
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
		fmt.Fprintf(stdout, "%d\t%s\t%s\t%d\n", r.Rank, r.Name, formatPoints(r.Points), r.Played)
	}
	return 0
}

// formatPoints renders a score without a trailing .0: 1, 1.5, 0.5.
func formatPoints(p float64) string {
	return strconv.FormatFloat(p, 'f', -1, 64)
}
