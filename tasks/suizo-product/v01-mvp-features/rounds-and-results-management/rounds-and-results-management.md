---
type: story
status: in_progress
id: rounds-and-results-management
title: Rounds and results management
assignee: Arggon
branch: feat/rounds-and-results-management
parent: v01-mvp-features
labels: []
created: "2026-09-13"
updated: "2026-09-13"
claimed_at: "2026-09-13T23:00:24.916Z"
depends_on: [swiss-pairing-engine]
worktree_path: /home/arggon/Projects/suizo-rounds-and-results-management
---
<!--
  Placement (v0): tasks/suizo-product/v01-mvp-features/rounds-and-results-management/rounds-and-results-management.md (story index; required).
  parent MUST be the epic id. Optional style prefixes (e.g. story-) are not type discriminators.
-->

# Rounds and results management

Implements GitHub issue #3. Follow `docs/playbooks/go.md`.

## Context

Drive a tournament through its rounds: start the next round (pair + persist),
report results per board, show pending/done status.

**File placement constraint (coordinator order):** add round lifecycle methods
on `*Tournament` IN `tournament.go` (same file pairing and standings extend —
the coordinator resolves the resulting merge). CLI verbs in `main.go`, tests
in `rounds_test.go`.

Commands: `suizo rounds start` (pair + open next round), `suizo results
<round> <board> <1-0|0-1|0.5-0.5>` (board is 1-based), `suizo rounds status`.

## Acceptance

- [x] `rounds start` refuses while the last round has pending matches.
- [x] `results` validates: round exists, board exists, result is one of the three legal values.
- [x] `rounds status` prints per-board pending/done.
- [x] Table-driven tests for the lifecycle: start → partial results → completion blocked → completion allowed.
- [x] `go test ./...`, `go vet ./...`, `golangci-lint run` all clean.

## Notes

- `rounds start` reuses `Pairings()` from the pairing story — if that story
  has not landed yet, build against its documented signature and note the
  integration point here (the coordinator merges both).

### 2026-09-13 @Arggon
Implemented in PR #10 (merged). The merge into main after PR #9 conflicted in tournament.go + main.go (coordinator resolution: both feature blocks kept, documented in merge commit 6af344b — stress experiment 2 evidence).
