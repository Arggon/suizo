---
exploration_id: swiss-pairing-001
title: swiss-pairing
status: closed
created: 2026-09-13
---

# Exploration: swiss-pairing (swiss-pairing-001)

Spike record: compare pairing algorithm candidates for amateur Swiss
tournaments. Decision recorded in [ADR-0001](../adr/0001-swiss-pairing-algorithm.md).

## Candidates

1. **Dutch system (simplified)** — score groups, rank-order pairing within
   group, float odd players down, rematch avoidance by swap/float, best-effort
   color alternation.
2. **Full FIDE Dutch Rules** — the complete FIDE C.04 regulation set: B.1/B.2
   pairing brackets, B.6/B.7 color rules as hard constraints, B.12/B.13
   exchange procedures with backtracking.
3. **Cumulative/monrad variants** — local club rules of varying rigor
   (monrad: ranking by points then internal rating; often allows rematches in
   later rounds).
4. **Off-the-shelf pairing library** — import a Go (or shelled-out) tournament
   library instead of implementing.

## Criteria

- Correctness for amateur play (no rematches, fair byes) — weighted highest.
- Determinism / explainability to a club organizer ("why do I play Beto?").
- Implementation and test cost in stdlib-only Go.
- Regulatory conformance (only matters for FIDE-rated events — out of scope).

## Findings

- The Dutch system is the de-facto standard for non-rated amateur chess events;
  FIDE's full rules exist for rated events and are a superset with hard color
  constraints and defined backtracking (source:
  https://handbook.fide.com/chapter/C04StatutoryStandards, accessed
  2026-09-13).
- FIDE explicitly publishes simplified pairings guidance for amateur events;
  score-group pairing with down-floating is the recognizable core (source:
  https://handbook.fide.com/chapter/C04, accessed 2026-09-13).
- No maintained pure-Go Swiss pairing library surfaced on pkg.go.dev top
  results; the existing ones are unmaintained chess-GUI helper forks (source:
  https://pkg.go.dev/search?q=swiss+tournament, accessed 2026-09-13).
- Monrad-style variants tolerate rematches by design — they exist to keep
  later rounds "interesting", which conflicts with our no-rematch invariant
  (source: https://en.wikipedia.org/wiki/Monrad_system, accessed 2026-09-13).

## Recommendation

Simplified Dutch (candidate 1) for v1, with the score-group + float skeleton
shaped so FIDE bracket rules (candidate 2) can slot in later without changing
callers. No external library (candidate 4): nothing maintained exists, and the
core algorithm is ~150 lines plus table-driven tests. Monrad rejected: violates
the no-rematch invariant.

**Decision:** ADR-0001 accepts the simplified Dutch system.
