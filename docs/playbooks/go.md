---
playbook_id: go
version: v1.27.1
researched: 2026-09-13
status: current
---

# go playbook (suizo)

Pinned version: **Go 1.27.1** (patch release, 2026-09-01; the 1.27 line shipped
2026-08-19).
Sources: [go.dev release history](https://go.dev/doc/devel/release) (accessed
2026-09-13), [Go 1.27 release notes](https://go.dev/doc/go1.27) (accessed
2026-09-13). Go supports the two most recent major lines — 1.27.x and 1.26.x —
so 1.27.1 is the current stable choice.

## Setup

- Toolchain pinned per developer via mise: `mise use go@1.27.1` (CONTRIBUTING.md
  documents it). No system-wide install required.
- Build: `go build -o suizo .` — one static binary, no runtime deps.
- `go.mod` declares `go 1.27.1`; do not lower it, `go vet`'s `stdversion` check
  (default-on since 1.27) flags stdlib symbols newer than the module's declared
  version.
- Verify: `go version` → `go version go1.27.1 linux/amd64` (verified on this
  machine 2026-09-13).

## Conventions

- **Stdlib only** for this project's lifetime: `net/http` (serving, issue #4),
  `encoding/json` (persistence), `flag`/`os.Args` (CLI). No web frameworks.
- **Table-driven tests are the default** test shape (see Testing).
- Error handling: sentinel errors for expected domain outcomes
  (`errors.Is`-comparable), `fmt.Errorf("context: %w", err)` for wrapping.
- Naming/structure per [Effective Go](https://go.dev/doc/effective_go)
  (accessed 2026-09-13): mixedCaps, doc comments on exported identifiers,
  flat `package main` layout while the binary is small.
- Since 1.27, `go fix` modernizers (`embedlit`, `slicesbackward`,
  `atomictypes`, `unsafefuncs`) are the sanctioned way to modernize code — run
  `go fix ./...` before large refactors rather than hand-editing
  ([release notes](https://go.dev/doc/go1.27), accessed 2026-09-13).

## Testing

- `go test ./...` gates every push; table-driven tests: one struct slice per
  behavior, `t.Run` per case, `t.Helper()` in setup helpers.
- 1.27 additions worth adopting when needed:
  `testing/synctest.Sleep` for virtual-time tests, and
  `net/http/httptest.NewTestServer` on an in-memory fake network — ideal for
  the future `suizo serve` tests without real ports
  ([release notes](https://go.dev/doc/go1.27), accessed 2026-09-13).
- `go test -json` output now carries `OutputType` (`error`, `error-continue`,
  `frame`) — if we add CI parsing later, consume that field.
- Integration-with-CLI pattern used here: call `run(args, storePath, stdout,
  stderr)` directly against a `t.TempDir()` file; no process spawning needed.

## Security

- **encoding/json v1 is now backed by the new v2 implementation** (1.27):
  stricter defaults (rejects invalid UTF-8 and duplicate object names),
  unmarshal significantly faster, error message text may differ. suizo relies
  on v1 API compatibility — hand-edited state files with duplicate keys are now
  rejected at load ([release notes](https://go.dev/doc/go1.27), accessed
  2026-09-13).
- If a vulnerability in the 1.27 line surfaces, patch releases land on the
  same line — re-pin to the latest 1.27.x (check
  [release history](https://go.dev/doc/devel/release), accessed 2026-09-13).
- `go tool trace -http` now binds localhost only — safe default, keep it.

## Upgrade policy

- Re-research when `arggon playbook status` flags this playbook stale
  (> 90 days by default), or when a new minor (1.28) or security patch lands.
- Upgrade rule: patch bumps (1.27.x → 1.27.y) are drop-in; minor bumps wait at
  least one patch release after GA and require a full-suite run plus a check
  of the release notes for `encoding/json` and `net/http` behavior changes.
- Update this file's `version` + `researched` via
  `arggon playbook refresh go --version <v>` after re-research.
