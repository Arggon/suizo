---
type: story
status: in_progress
id: swiss-pairing-engine
title: Swiss pairing engine
assignee: Arggon
branch: feat/swiss-pairing-engine
parent: v01-mvp-features
labels: []
created: "2026-09-13"
updated: "2026-09-13"
claimed_at: "2026-09-13T22:53:05.568Z"
worktree_path: /home/arggon/Projects/suizo-swiss-pairing-engine
---
<!--
  Placement (v0): tasks/suizo-product/v01-mvp-features/swiss-pairing-engine/swiss-pairing-engine.md (story index; required).
  parent MUST be the epic id. Optional style prefixes (e.g. story-) are not type discriminators.
-->

# Swiss pairing engine

Implements GitHub issue #1 and spec `docs/specs/spec-swiss-pairing-001.md`
(plan: `docs/plans/plan-swiss-pairing-001.md`, ADR-0001 has the algorithm
decision). Follow `docs/playbooks/go.md`.

## Context

The core domain feature: produce the next round's pairings for an amateur
Swiss tournament. Simplified Dutch system per ADR-0001 — score groups, no
rematches, fair byes, best-effort colors.

**File placement constraint (coordinator order):** the domain lives in
`tournament.go`. Add `Pairings()` and its helpers as methods on `*Tournament`
IN `tournament.go` (not a new file); tests go in `pairing_test.go`. The CLI
verb (`suizo pair`) goes in `main.go`.

## Acceptance

- [ ] `Pairings()` on a fresh even-count tournament pairs by id order (1v2, 3v4…).
- [ ] After round-1 results, round-2 pairs within score groups (winners vs winners).
- [ ] No rematch while a rematch-free assignment exists (deterministic pseudo-random loop over small tournaments).
- [ ] Odd count → one bye, lowest-ranked player without a previous bye; never two byes for one player.
- [ ] Fewer than 2 players → error (no panic, no empty round).
- [ ] `suizo pair` computes, persists round N+1, prints boards; end-to-end `run()` test against a temp store.
- [ ] `go test ./...`, `go vet ./...`, `golangci-lint run` all clean.
- [ ] spec + plan status flipped to `implemented` in this PR.

## Notes

- Conflict course: this story and `standings-table` both extend the
  `Tournament` struct in `tournament.go` — that merge is resolved by the
  coordinator (experiment: stress scenario 2).
