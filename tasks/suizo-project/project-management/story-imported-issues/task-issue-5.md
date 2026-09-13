---
type: task
status: done
id: task-issue-5
title: "issue #5: Tie-breakers"
assignee: Arggon
parent: story-imported-issues
labels: [enhancement]
created: "2026-09-13"
updated: "2026-09-13"
issue: 5
---
## What
Real tie-break systems for standings, in the order the chess world expects them:

- **Buchholz** (sum of opponents' scores) — primary.
- **Buchholz cut 1** (drop lowest opponent) — secondary.
- **Direct encounter** — tertiary.
- **Sonneborn–Berger** — optional flag.

## Acceptance
- [ ] `suizo standings` shows tie-break columns.
- [ ] Documented formula per tie-break in docs/FORMAT.md or the standings docs.
- [ ] Table-driven tests with known small tournaments and hand-computed values.
> imported from issue #5

### 2026-09-13 @Arggon
Implemented via swiss-pairing-engine/standings-table/rounds/ui/tie-breakers stories (PRs #8-#12 merged).
