---
type: task
status: todo
id: task-issue-1
title: "issue #1: Swiss pairing engine"
parent: story-imported-issues
labels: [enhancement]
created: "2026-09-13"
updated: "2026-09-13"
issue: 1
---
## What
Implement the Swiss pairing algorithm: given the current tournament state (players + completed rounds), produce the pairings for the next round.

## Rules
- Sort players by score (descending), pair 1v2, 3v4, ... within score groups.
- **Nobody plays the same opponent twice.** If a rematch is unavoidable inside a score group, float players to the adjacent group.
- Odd number of players: the lowest-ranked player who has not yet had a bye receives one (match with `isBye: true`, scores 1 point).
- Colors (white/black) alternate where possible; a player must not get white twice in a row when avoidable.

## Acceptance
- [ ] `suizo pair` emits the next round and persists it.
- [ ] Table-driven tests covering: even/odd player counts, repeat-opponent avoidance with floating, byes never repeat, color alternation.
- [ ] docs/FORMAT.md unchanged (pairing writes only existing fields).
> imported from issue #1
