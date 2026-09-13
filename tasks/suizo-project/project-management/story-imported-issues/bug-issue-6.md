---
type: bug
status: todo
id: bug-issue-6
title: "issue #6: Bug: concurrent CLI invocations silently lose writes (no file locking)"
parent: story-imported-issues
labels: [bug]
created: "2026-09-13"
updated: "2026-09-13"
issue: 6
---
## Found while
Self-review + stress loop of the MVP store (post-atomic-write refactor).

## Description
`store.save` is atomic per-write (temp + rename), but the CLI's read-modify-write cycle has **no cross-process locking**. Two concurrent `suizo players add` invocations both `load()` the same document, each computes its own next id, and the second `rename` overwrites the first player entirely — data loss with exit code 0 on both processes.

This is realistic at a tournament desk: organizer and assistant both registering late arrivals at the same time.

## Reproduction
```bash
for i in $(seq 1 30); do
  rm -f suizo.json
  suizo players add "A$i" & suizo players add "B$i" & wait
  test "$(suizo players list | wc -l)" = 2 || echo "LOST WRITE in iteration $i"
done
```

Observed on this machine: **21 of 30 iterations lost a write.**

## Expected
Both players persist; ids never collide; the loser of the race retries or blocks, never overwrites.

## Direction
Advisory file lock (`syscall.Flock` on a sidecar lockfile, or O_EXCL lock file with retry+timeout) around the load→save cycle. Must stay stdlib-only and must not break the single-user fast path.
> imported from issue #6
