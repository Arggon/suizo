---
type: task
status: todo
id: task-issue-2
title: "issue #2: Standings / tabla de posiciones"
parent: story-imported-issues
labels: [enhancement]
created: "2026-09-13"
updated: "2026-09-13"
issue: 2
---
## What
Compute and display the standings table: score per player (win = 1, draw = 0.5, loss = 0, bye = 1), sorted.

## Output shape
```
Rk  Name           Pts  played
1   Ana García     2.0  3
2   Beto Ruiz      1.5  3
```

## Acceptance
- [ ] `suizo standings` prints the table above.
- [ ] Byes count as a full point.
- [ ] Players with identical scores order deterministically (stable tie order for now; real tie-breakers are issue #5).
- [ ] Table-driven tests for score accumulation across several rounds.
> imported from issue #2
