---
type: story
status: todo
id: standings-table
title: Standings table
parent: v01-mvp-features
labels: []
created: "2026-09-13"
updated: "2026-09-13"
depends_on: [swiss-pairing-engine]
---
<!--
  Placement (v0): tasks/suizo-product/v01-mvp-features/standings-table/standings-table.md (story index; required).
  parent MUST be the epic id. Optional style prefixes (e.g. story-) are not type discriminators.
-->

# Standings table

Implements GitHub issue #2. Follow `docs/playbooks/go.md`.

## Context

Score table per player (win 1, draw 0.5, loss 0, bye 1), sorted, deterministic
tie order (id asc) until real tie-breakers arrive (issue #5). Pure read: no
mutation, no persistence.

**File placement constraint (coordinator order):** add `scores()`/`Standings()`
as methods on `*Tournament` IN `tournament.go` (same file the pairing story is
extending — the merge conflict is expected and will be resolved by the
coordinator). CLI verb in `main.go`, tests in `standings_test.go`.

## Acceptance

- [ ] `suizo standings` prints `Rk  Name  Pts  played` with byes counted as a full point.
- [ ] Identical scores order deterministically (id asc).
- [ ] Table-driven tests: multi-round accumulation, byes, draws, empty tournament.
- [ ] `go test ./...`, `go vet ./...`, `golangci-lint run` all clean.

## Notes

- Depends on `swiss-pairing-engine` only for score-group semantics — the
  standings code itself is independent and may proceed in parallel.
