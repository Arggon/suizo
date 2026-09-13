---
type: task
status: todo
id: task-issue-4
title: "issue #4: Web UI mínima"
parent: story-imported-issues
labels: [enhancement]
created: "2026-09-13"
updated: "2026-09-13"
issue: 4
---
## What
A minimal local web UI over the same JSON state, using **net/http stdlib only** — no frameworks, no npm. The organizer opens http://localhost:PORT in a browser at the club; the CLI remains the primary interface.

## Pages
- Standings table (auto-refresh with a plain `<meta http-equiv="refresh">` or a fetch poll — no JS build step).
- Current round boards with dropdowns to report results (form POST).
- Player list + add form.

## Acceptance
- [ ] `suizo serve` binds 127.0.0.1 only.
- [ ] Rendering server-side html/template; zero client build.
- [ ] Concurrent access with the CLI does not corrupt state (see bug #6 — the fix is a prerequisite).
> imported from issue #4

### 2026-09-13 @Arggon
Implemented via swiss-pairing-engine/standings-table/rounds/ui/tie-breakers stories (PRs #8-#12 merged).
