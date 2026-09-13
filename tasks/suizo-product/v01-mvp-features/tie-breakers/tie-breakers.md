---
type: story
status: in_progress
id: tie-breakers
title: Tie-breakers
assignee: Arggon
branch: feat/tie-breakers
parent: v01-mvp-features
labels: []
created: "2026-09-13"
updated: "2026-09-13"
claimed_at: "2026-09-13T23:20:23.468Z"
worktree_path: /home/arggon/Projects/suizo-tie-breakers
---
<!--
  Placement (v0): tasks/suizo-product/v01-mvp-features/tie-breakers/tie-breakers.md (story index; required).
  parent MUST be the epic id. Optional style prefixes (e.g. story-) are not type discriminators.
-->

# Tie-breakers

Implements GitHub issue #5. LATE story: filed after the parent epic
`v01-mvp-features` had cascade-completed. Reopening the done epic for this is
an **owner-authorized deviation** (documented in the experiment log).

## Context

Standings currently tie-break by numeric player id — arbitrary. Real Swiss
events use documented tie-break systems. This story adds them, in the order
chess practice expects.

**File placement:** extend `Standings()`/`Standing` in `tournament.go`
(same shared struct/file as the rest of the domain).

## Acceptance

- [x] `Standing` gains Buchholz (sum of opponents' scores), Buchholz Cut 1 (drop lowest opponent score) and direct-encounter columns.
- [x] `suizo standings` prints the new columns; formulas documented in docs/FORMAT.md.
- [x] Table-driven tests with known small tournaments and hand-computed Buchholz values, plus a case where direct encounter decides between two tied players.
- [x] `go test ./...`, `go vet ./...`, `golangci-lint run` all clean.

## Notes

- Byes contribute no opponent score to Buchholz (standard practice: a bye
  opponent does not exist; the bye point still counts for the player's own
  score).
- The imported `task-issue-5` mirrors this work on GitHub; it is completed
  together with this story.
