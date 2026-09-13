---
playbook_id: golangci-lint
version: v2.13.2
researched: 2026-09-13
status: current
---

# golangci-lint playbook (suizo)

Pinned version: **golangci-lint v2.13.2** (2026-08-28, bug-fix release on the
v2 line; v2.13.0 shipped 2026-08-19).
Sources: [golangci-lint releases](https://github.com/golangci/golangci-lint/releases)
(accessed 2026-09-13),
[changelog](https://golangci-lint.run/docs/product/changelog/) (accessed
2026-09-13). Exists and is stable — hence this playbook.

## Setup

- Pinned per developer via mise: `mise use golangci-lint@2.13.2`. Verified on
  this machine 2026-09-13: `golangci-lint version` → `version 2.13.2 built
  with go1.27.0`.
- The v2 line is a major rework of configuration and integration (2025); all
  config lives in `.golangci.yml` with a mandatory top-level `version: "2"`
  key. v1-style configs fail on v2 binaries.
- Run: `golangci-lint run` at the repo root; `golangci-lint run --fix` for
  auto-fixable issues only after reading the diff.

## Conventions

- suizo config (`.golangci.yml`): default linter set + `errcheck` with an
  explicit `exclude-functions` allowlist for `fmt.Fprint/Fprintf/Fprintln` —
  CLI reporting writes must not turn into error paths; the exit code carries
  the real signal. Deliberate internal ignores use `_ =` with a comment.
- Keep the exclusion list short and commented; every entry is a policy
  decision a reviewer can see.
- errcheck fixes in v2.13.2 (cache entropy decrease; `iface` 1.5.1,
  `staticcheck` 0.8.x updates — see
  [releases](https://github.com/golangci/golangci-lint/releases), accessed
  2026-09-13) — no behavior change for this repo's config.

## Testing

- Lint gates locally: `golangci-lint run` must exit 0 before every push, same
  as `go vet ./...` and `go test ./...`.
- After changing `.golangci.yml`, run once on a clean tree and once after
  introducing a known-bad snippet in a scratch file to confirm the rule fires.
- CI (when added): the official
  [golangci-lint-action](https://github.com/golangci/golangci-lint-action)
  with `version: v2.13.2` pinned — do not use `latest`.

## Security

- Run only pinned releases from the official GitHub releases; the binary is
  Go-built and statically linked — no plugin system in use here, which keeps
  the supply-chain surface to the linter itself.
- v2.13.2's decreased cache entropy (2026-08-28) hardens cache-directory
  collisions on shared machines — relevant for CI runners.

## Upgrade policy

- Re-research when `arggon playbook status` flags stale, when a minor (v2.14)
  lands, or when Go upgrades (the linter must support the Go toolchain in use).
- Bump patch versions freely after a full `run` on the repo; minor bumps get
  the changelog read and a config re-check first (v2 renamed several setting
  keys versus v1 — expect the same rigor between minors).
- After re-research: `arggon playbook refresh golangci-lint --version <v>`.
