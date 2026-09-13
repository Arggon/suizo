package main

import (
	"fmt"
	"sort"
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

// score returns the tournament score of one player: win 1, draw 0.5, loss 0,
// bye 1. Unreported matches contribute 0.
func (t *Tournament) score(id string) float64 {
	var s float64
	for _, r := range t.Rounds {
		for _, m := range r.Matches {
			if m.IsBye {
				if m.White == id {
					s += 1
				}
				continue
			}
			switch m.Result {
			case "1-0":
				if m.White == id {
					s += 1
				}
			case "0-1":
				if m.Black == id {
					s += 1
				}
			case "0.5-0.5":
				if m.White == id || m.Black == id {
					s += 0.5
				}
			}
		}
	}
	return s
}

// rankedIDs returns all player ids ordered by score (desc), then id (asc).
// Ids sort numerically when both carry the p<N> shape, so p10 ranks after p9.
func (t *Tournament) rankedIDs() []string {
	ids := make([]string, 0, len(t.Players))
	for _, p := range t.Players {
		ids = append(ids, p.ID)
	}
	sort.Slice(ids, func(i, j int) bool {
		si, sj := t.score(ids[i]), t.score(ids[j])
		if si != sj {
			return si > sj
		}
		return playerIDLess(ids[i], ids[j])
	})
	return ids
}

// playerIDLess orders player ids numerically when both match p<N>, falling
// back to plain string order otherwise.
func playerIDLess(a, b string) bool {
	var na, nb int
	_, ea := fmt.Sscanf(a, playerIDPrefix+"%d", &na)
	_, eb := fmt.Sscanf(b, playerIDPrefix+"%d", &nb)
	if ea == nil && eb == nil && na != nb {
		return na < nb
	}
	return a < b
}

// hadBye reports whether the player ever received a bye.
func (t *Tournament) hadBye(id string) bool {
	for _, r := range t.Rounds {
		for _, m := range r.Matches {
			if m.IsBye && m.White == id {
				return true
			}
		}
	}
	return false
}

// met reports whether the two players were already paired in any round.
func (t *Tournament) met(a, b string) bool {
	for _, r := range t.Rounds {
		for _, m := range r.Matches {
			if m.IsBye {
				continue
			}
			if (m.White == a && m.Black == b) || (m.White == b && m.Black == a) {
				return true
			}
		}
	}
	return false
}

// lastColors maps player id to the color they had in the most recent round
// they played: true for white, false for black. Absent means no color yet.
func (t *Tournament) lastColors() map[string]bool {
	colors := make(map[string]bool)
	for i := len(t.Rounds) - 1; i >= 0; i-- {
		for _, m := range t.Rounds[i].Matches {
			if m.IsBye {
				continue
			}
			if _, ok := colors[m.White]; !ok {
				colors[m.White] = true
			}
			if _, ok := colors[m.Black]; !ok {
				colors[m.Black] = false
			}
		}
	}
	return colors
}

// balanceColors picks the white player for a new pairing, best effort: a
// player who had white last round yields black to one who did not.
func (t *Tournament) balanceColors(a, b string, colors map[string]bool) (string, string) {
	if colors[a] && !colors[b] {
		return b, a
	}
	return a, b
}

// Pairings computes the matches for round N+1 per ADR-0001's simplified Dutch
// system: rank by score desc then id asc, pair within score groups in rank
// order, float odd leftovers down, avoid rematches (swap, then float), give
// the single survivor a bye if it never had one, and balance colors best
// effort. It never mutates the tournament; the caller decides to persist.
func (t *Tournament) Pairings() ([]Match, error) {
	if len(t.Players) < 2 {
		return nil, errTooFewPlayers
	}
	ranked := t.rankedIDs()
	colors := t.lastColors()
	rank := make(map[string]int, len(ranked))
	for i, id := range ranked {
		rank[id] = i
	}

	// scoreGroups splits ranked into contiguous groups of equal score.
	var groups [][]string
	start := 0
	for i := 1; i <= len(ranked); i++ {
		if i == len(ranked) || t.score(ranked[i]) != t.score(ranked[start]) {
			groups = append(groups, ranked[start:i])
			start = i
		}
	}

	var matches []Match
	var floating []string
	for _, g := range groups {
		pool := make([]string, 0, len(floating)+len(g))
		pool = append(pool, floating...)
		pool = append(pool, g...)
		sort.Slice(pool, func(i, j int) bool { return rank[pool[i]] < rank[pool[j]] })
		floating = floating[:0]
		used := make(map[int]bool, len(pool))
		for i := 0; i < len(pool); i++ {
			if used[i] {
				continue
			}
			matched := false
			for j := i + 1; j < len(pool); j++ {
				if used[j] || t.met(pool[i], pool[j]) {
					continue
				}
				w, b := t.balanceColors(pool[i], pool[j], colors)
				matches = append(matches, Match{White: w, Black: b})
				used[i], used[j] = true, true
				matched = true
				break
			}
			if !matched {
				floating = append(floating, pool[i])
			}
		}
	}

	switch len(floating) {
	case 0:
		// nothing to do
	case 1:
		if t.hadBye(floating[0]) {
			return nil, fmt.Errorf("%w: %s", errNoByeCandidate, floating[0])
		}
		matches = append(matches, Match{White: floating[0], IsBye: true})
	default:
		return nil, fmt.Errorf("%w: %d players unpairable", errNoRematchFreePairing, len(floating))
	}
	return matches, nil
}
