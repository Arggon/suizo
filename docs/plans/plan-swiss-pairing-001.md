---
plan_id: swiss-pairing-001
title: Plan for Swiss pairing engine
spec: docs/specs/spec-swiss-pairing-001.md
status: proposed
created: 2026-09-13
---

# Plan: Swiss pairing engine (swiss-pairing-001)

Derived from `docs/specs/spec-swiss-pairing-001.md`. Each task carries a
verifiable acceptance criterion and links back to the spec.

## Tasks

### T1: Score computation inside tournament.go

- Add `func (t *Tournament) scores() map[string]float64` (win 1, draw 0.5,
  loss 0, bye 1) scanning `t.Rounds`.
- **Acceptance:** table-driven cases in `pairing_test.go` covering byes,
  draws and multi-round accumulation.

### T2: Pairings() core

- Add `func (t *Tournament) Pairings() ([]Match, error)` implementing the
  spec's 6 rules in `tournament.go` (same file as the type — the domain stays
  together).
- **Acceptance:** fresh-tournament and second-round cases produce the exact
  pairings asserted in the spec.

### T3: Rematch avoidance + floating

- Swap-then-float logic inside group pairing, using the meeting history from
  `t.Rounds`.
- **Acceptance:** property-style loop over deterministic pseudo-random small
  tournaments (8-15 players, 3-5 rounds) asserts zero rematches; the forced
  rematch case (2 players, 2 rounds) returns an error.

### T4: Byes

- Bye assignment per spec rules 5.
- **Acceptance:** odd-count cases from the spec checklist pass; double-bye is
  impossible across rounds.

### T5: CLI exposure

- `suizo pair` subcommand in `main.go`: computes `Pairings()`, appends the
  round, saves, prints the boards.
- **Acceptance:** end-to-end `run()` test — add 4 players, pair, assert round
  1 persisted in the store file.
