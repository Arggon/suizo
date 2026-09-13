---
spec_id: swiss-pairing-001
title: Swiss pairing engine
status: implemented
created: 2026-09-13
---

# Spec: Swiss pairing engine (swiss-pairing-001)

## Purpose

Amateur tournaments run on the Swiss system: a fixed number of rounds, no
elimination, and pairings that match players with similar scores. This spec
defines the pairing algorithm for suizo — the function that, given the current
tournament state, produces the next round's matches.

**Invariants:**

- **No rematches.** Two players never meet twice when any rematch-free pairing
  of the score groups exists; floating to adjacent score groups is preferred
  over a rematch.
- **Nobody is dropped.** Every unwithdrawn player is paired or receives a bye,
  exactly once per round.
- **Byes are scarce.** A bye goes to the lowest-ranked player without a
  previous bye; nobody gets two byes.
- **Pure computation.** Pairing never mutates the tournament; the caller
  decides to persist the produced round. Deterministic for equal input: same
  state → same pairings (stable ordering by score then id).

## Synopsis

```go
// Pairings returns the matches for round N+1 (not yet persisted).
func (t *Tournament) Pairings() ([]Match, error)
```

Rules, in priority order (classic Dutch system, simplified for amateur play):

1. Rank players by score (desc), tie-broken by player id (asc) — real
   tie-breakers arrive with issue #5.
2. Walk score groups top-down; pair within a group in rank order
   (1 vs 2, 3 vs 4, …).
3. A group with an odd count floats its last player to the next group.
4. Avoid rematches: before accepting a pairing, check the players' meeting
   history; on collision, try swapping within the group, then float a
   different player.
5. The single unpaired survivor after all groups gets a bye if eligible
   (no previous bye, lowest rank first), else pairing fails with an error.
6. Colors: alternate where possible — a player who had white last round gets
   black (best effort; not a hard constraint in v1).

Error cases: fewer than 2 players; a bye is required but every candidate
already had one.

## Acceptance

- [ ] `Pairings()` on a fresh tournament with an even player count produces
      1v2, 3v4, … by id order, with no errors.
- [ ] After round 1 results, round 2 pairs winners against winners (score
      groups respected).
- [ ] No rematch is produced while a rematch-free assignment exists (property
      test over random small tournaments).
- [ ] Odd player count → exactly one bye, assigned to the lowest-ranked
      player without a previous bye; a second round never repeats the bye.
- [ ] Fewer than 2 players → error, not a panic or an empty round.
- [ ] All cases covered by table-driven tests in `pairing_test.go`; the
      existing suite stays green.
