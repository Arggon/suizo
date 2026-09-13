// Command suizo is a Swiss-system tournament manager for amateur events —
// chess, futbolito, anything pairable. State lives in a single JSON file.
package main

import (
	"fmt"
	"io"
	"os"
)

const usage = `suizo - Swiss-system tournament manager

Usage:
  suizo players add <name>   Register a player; prints its id
  suizo players list         List registered players

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
