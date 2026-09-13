---
adr_id: 0001
title: Swiss pairing algorithm — simplified Dutch system
status: Accepted
created: 2026-09-13
---

# ADR-0001: Swiss pairing algorithm — simplified Dutch system

- **Status:** Accepted (2026-09-13)
- **Exploration:** [docs/explorations/exploration-swiss-pairing-001.md](../explorations/exploration-swiss-pairing-001.md)
- **Spec:** [docs/specs/spec-swiss-pairing-001.md](../specs/spec-swiss-pairing-001.md)

## Context

suizo needs to produce next-round pairings for amateur Swiss tournaments.
FIDE's full Dutch Rules target rated play: hard color constraints, defined
bracket exchanges with backtracking. Club events (chess, futbolito) need the
recognizable core — score groups, no rematches, fair byes — and an
explanation a volunteer organizer can defend at the table. No maintained
pure-Go pairing library exists (checked 2026-09-13).

## Decision

Implement the **simplified Dutch system** in `tournament.go`:

1. Rank by score desc, id asc.
2. Pair within score groups in rank order; odd groups float the last player
   down.
3. Rematch avoidance by in-group swap, then alternative float — preferred
   over any rematch.
4. Bye: lowest-ranked player without a previous bye; exactly one per round.
5. Colors best-effort alternation only (soft constraint in v1).

## Consequences

- Deterministic, explainable pairings; ~150 lines + table-driven tests, zero
  dependencies.
- Not FIDE-conformant: suizo must NOT claim rated-event compatibility until
  the full bracket/exchange rules are implemented — the README and docs must
  not suggest otherwise.
- The score-group skeleton keeps the FIDE rules (hard colors, bracket
  exchanges) an additive change behind `Pairings()`; revisiting this decision
  would supersede this ADR, not rewrite it.
