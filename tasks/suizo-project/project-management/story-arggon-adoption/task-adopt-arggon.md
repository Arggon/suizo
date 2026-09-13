---
type: task
status: in_progress
id: task-adopt-arggon
title: Adopt ArggonManager in this repo
assignee: Arggon
parent: story-arggon-adoption
labels: []
created: "2026-09-13"
updated: "2026-09-13"
claimed_at: "2026-09-13T22:36:31.206Z"
---
## Context

This repo is adopting ArggonManager over an existing documentation set: `arggon init` generated the governing docs (marked `<!-- arggon:generated ... -->`, with TODO placeholders), while any pre-existing docs were left untouched on disk. Your job as the executing agent: extract the valuable content from the adopter docs into the generated ones, archive what you replace, and report back on this task.

Current inventory (paths, sizes, managed vs adopter-owned, stack manifests):

```
arggon adopt --dry-run --json
```

## Checklist

- [ ] 1. Read the arggon-generated governing docs first: AGENTS.md, docs/convention.md, docs/engineering.md, docs/playbooks/ (if present). Follow them for the rest of this migration.
- [ ] 2. Sweep the existing repo docs (list them from the inventory above): extract the project description, conventions, workflows, and stack info. Extract, don't wholesale-copy — rewrite into the target doc's structure and drop duplicated or outdated material.
- [ ] 3. Complete the arggon-generated docs with the extracted content — fill the TODO placeholders: project description in AGENTS.md; CONTRIBUTING.md specifics (environment setup, build/test commands); ARCHITECTURE.md problem statement. The SECURITY.md contact is human input — leave it flagged for a human, never invent it.
- [ ] 4. Archive replaced originals to backup/<YYYY-MM-DD>/ preserving their relative paths (use today's date). Only docs you REPLACED get archived; never archive README.md — merge into it instead.
- [ ] 5. Detect the stack from the manifests (package.json / requirements.txt / go.mod / Cargo.toml / pom.xml); for each technology create a playbook (`arggon playbook new <tech>`), research current versions and best practices with dated sources, then record them with `arggon playbook refresh <tech> --version <v>`.
- [ ] 6. Baseline the sanctioned edits: run `arggon adopt --ack` so the generated docs you completed in step 3 become the new x-generated baseline (their checksums are refreshed and they stop reporting as modified). Hand edits made AFTER this ack still report modified — the protection stays intact.
- [ ] 7. Verify: `arggon validate` + `arggon spec validate` (if specs exist) + `arggon playbook status`.
- [ ] 8. Report: comment on this task (`arggon comment task-adopt-arggon`) listing the extracted content, archived files, and created playbooks; flip this task done when the human reviews.

### 2026-09-13 @Arggon
Adoption checklist executed (2026-09-13).

**Sweep + extraction (steps 2-3):** pre-existing docs = README.md, docs/DECISIONS.md, docs/FORMAT.md. Extracted (rewritten, not copied): problem statement + constraints + big-picture + code map + layering rules into ARCHITECTURE.md; one-sentence description into AGENTS.md; Go toolchain setup + build/test/lint commands into CONTRIBUTING.md; testing expectations (unit vs CLI-integration split, merge gates) into docs/engineering.md. DECISIONS.md and FORMAT.md stay as project-owned docs (no overlap with generated docs; FORMAT.md referenced from ARCHITECTURE.md as schema source of truth).

**Backup (step 4):** nothing archived to backup/2026-09-13/ — no original was replaced; all extractions went INTO the generated placeholders. README.md untouched per checklist rule.

**Playbooks (step 5):** docs/playbooks/go.md (v1.27.1, researched 2026-09-13; sources: go.dev/doc/devel/release + go.dev/doc/go1.27, accessed 2026-09-13) and docs/playbooks/golangci-lint.md (v2.13.2, researched 2026-09-13; sources: github.com/golangci/golangci-lint/releases, accessed 2026-09-13).

**SECURITY.md:** contact arggondev@gmail.com recorded as owner-confirmed input (owner instruction at adoption time); the REVIEW flag for pre-release re-confirmation is kept in place.

**Ack (step 6):** adopt --ack baselined 16 generated docs (new checksums for the 6 edited ones).

**Verify (step 7):** arggon validate ok:true; spec validate ok:true; playbook status staleCount 0.

Note: arggon adopt initially failed with ADOPT_FAILED (no epic found) on a freshly-initialized tree; worked after creating initiative suizo-project + epic project-management. Recorded as a rough edge in the final experiment report.
