---
type: task
status: todo
id: task-issue-3
title: "issue #3: Gestión de rondas y resultados"
parent: story-imported-issues
labels: [enhancement]
created: "2026-09-13"
updated: "2026-09-13"
issue: 3
---
## What
Commands to drive the tournament through its rounds:

- `suizo rounds start` — pair and open the next round.
- `suizo results <round> <board> <1-0|0-1|0.5-0.5>` — report a result.
- `suizo rounds status` — show which boards are still pending.

## Acceptance
- [ ] Cannot start round N+1 while round N has pending matches.
- [ ] Result validation: only the three legal results are accepted; board and player ids must exist.
- [ ] `rounds status` shows per-board pending/done at a glance.
- [ ] Table-driven tests for the round lifecycle (start → partial results → completion).
> imported from issue #3

### 2026-09-13 @Arggon
Implemented via swiss-pairing-engine/standings-table/rounds/ui/tie-breakers stories (PRs #8-#12 merged).
