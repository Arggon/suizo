package main

import (
	"fmt"
	"strconv"
)

const playerIDPrefix = "p"

// Player is a tournament participant.
type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Match is a single pairing. A bye is a Match with IsBye set and Black empty.
// Result uses chess notation: "1-0", "0-1" or "0.5-0.5"; empty until reported.
type Match struct {
	White  string `json:"white"`
	Black  string `json:"black,omitempty"`
	IsBye  bool   `json:"isBye,omitempty"`
	Result string `json:"result,omitempty"`
}

// Round is one round of pairings.
type Round struct {
	Number  int     `json:"number"`
	Matches []Match `json:"matches"`
}

// Tournament is the full state of one tournament: players, rounds and results.
// It is persisted as a single JSON document by store.go; docs/FORMAT.md is the
// authoritative description of the on-disk shape.
type Tournament struct {
	Name    string   `json:"name"`
	Players []Player `json:"players"`
	Rounds  []Round  `json:"rounds"`
}

// player returns the player with the given id, or nil if absent.
func (t *Tournament) player(id string) *Player {
	for i := range t.Players {
		if t.Players[i].ID == id {
			return &t.Players[i]
		}
	}
	return nil
}

// addPlayer registers a new player and returns it. Names must be non-empty.
func (t *Tournament) addPlayer(name string) (Player, error) {
	if name == "" {
		return Player{}, errEmptyName
	}
	p := Player{ID: t.nextPlayerID(), Name: name}
	t.Players = append(t.Players, p)
	return p, nil
}

// nextPlayerID returns "p<N+1>" where N is the largest numeric suffix among
// existing ids, so ids stay stable even if players are ever removed.
func (t *Tournament) nextPlayerID() string {
	max := 0
	for _, p := range t.Players {
		var n int
		if _, err := fmt.Sscanf(p.ID, playerIDPrefix+"%d", &n); err == nil && n > max {
			max = n
		}
	}
	return playerIDPrefix + strconv.Itoa(max+1)
}
