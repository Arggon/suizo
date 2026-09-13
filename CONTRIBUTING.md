<!-- arggon:generated template="CONTRIBUTING.md" -->
# Contributing to suizo

Thanks for helping. Work is tracked in-tree under `tasks/` (Markdown work items managed by `arggon`) — GitHub is used for PRs only.

## Getting started

This is a Go project (Go 1.27.x; see `docs/playbooks/go.md` for the pinned
version and tooling). Install a toolchain via your package manager or
[mise](https://mise.jdx.dev): `mise use go@1.27.1 golangci-lint@2.13.2`.

1. Read [`AGENTS.md`](AGENTS.md) — the task workflow for humans and agents alike.
2. Find a claimable item: `arggon list --status todo --json`.
3. Claim it: `arggon update <id> --status in_progress --assignee <your-login>`. Never steal a claim.

## Build, test, lint

```console
$ go build -o suizo .      # the binary
$ go test ./...            # table-driven suite; must be green before every push
$ go vet ./...
$ golangci-lint run        # config in .golangci.yml
```

## Branches

One branch per work item, generated from the item id:

- `arggon branch <id>` — follows the configured patterns in `tasks/.convention.yml`.
- Defaults: `feat/<id>`, `fix/<id>`, `docs/<id>`, `chore/<id>`.
- Keep PRs small and focused; one concern per PR when possible.

## Commits

- Imperative mood, scoped prefix when useful: `feat: …`, `fix: …`, `docs: …`, `chore: …`, `test: …`.
- Reference the work item id in the commit body when it stands alone.

## Pull requests

- Reference the work item id in the PR title or body; move the item to `done` only when the PR fully finishes it.
- Update docs in the same PR as the change they describe.

### PR checklist

- [ ] Linked work item from `tasks/` (or a clear docs-only / chore reason)
- [ ] Tests pass locally (`go test ./...` plus `go vet ./...` and `golangci-lint run`)
- [ ] Docs updated in the same PR when behavior changed
- [ ] PR references the work item id

## Reporting bugs and filing work

File work items in the tree, not on GitHub: `arggon create bug "<title>" --parent <story-id>`. See [`docs/tracking.md`](docs/tracking.md) for how tracking works in this repo.

<!--
Copyright 2026 suizo contributors
-->
