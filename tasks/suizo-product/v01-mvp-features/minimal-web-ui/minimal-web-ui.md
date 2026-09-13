---
type: story
status: in_progress
id: minimal-web-ui
title: Minimal web UI
assignee: Arggon
branch: feat/minimal-web-ui
parent: v01-mvp-features
labels: []
created: "2026-09-13"
updated: "2026-09-13"
claimed_at: "2026-09-13T23:10:56.510Z"
depends_on: [rounds-and-results-management]
worktree_path: /home/arggon/Projects/suizo-minimal-web-ui
---
<!--
  Placement (v0): tasks/suizo-product/v01-mvp-features/minimal-web-ui/minimal-web-ui.md (story index; required).
  parent MUST be the epic id. Optional style prefixes (e.g. story-) are not type discriminators.
-->

# Minimal web UI

Implements GitHub issue #4. Follow `docs/playbooks/go.md`.

## Context

Local, read-mostly web view over the same JSON state: standings, current round
boards, players. `net/http` stdlib + `html/template` only — no frameworks, no
client build step, bound to 127.0.0.1. The CLI stays the primary interface;
result reporting via form POST is enough.

Handler functions live in `web.go`; they may read through `store` and domain
methods already defined in `tournament.go` (do not duplicate domain logic).

## Acceptance

- [x] `suizo serve` binds 127.0.0.1 only and serves: standings page, current-round page with result dropdowns (form POST → domain method), players page.
- [x] Server-side `html/template` rendering; zero JS build; auto-refresh via `<meta http-equiv="refresh">`.
- [x] `httptest`-based tests (1.27 `NewTestServer` or classic) covering GET pages + a POST result report round-trip.
- [x] `go test ./...`, `go vet ./...`, `golangci-lint run` all clean.

## Notes

- Depends on rounds/results methods landing (`rounds-and-results-management`).
