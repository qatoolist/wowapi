---
id: CLOSURE-W01-E04-S001
type: closure-record
parent_story: W01-E04-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Closure — W01-E04-S001

*Closure record, 2026-07-13: all five tasks are implemented and verified with preserved fail-first
evidence — the DX-02 slice (T003/T004) and scope addition (T005) by W01Gen, the DX-01 slice
(T001/T002) by follow-up worker W01GenDX01 (DEV-W01-E04-S001-01 closed). The story stands at
`verified`; movement to `accepted` awaits the wave-level independent review gate (mandate §14) and
acceptance authority per this epic's `acceptance.md`.*

## Acceptance-criteria completion

- AC-W01-E04-S001-01 — **verified** (EV-W01-E04-S001-001; fail-first pairs across all three
  resolution paths, both live defect shapes — `v0.0.0` and the SF-7 `+dirty` stamp — captured pre-fix;
  every failure path proven fail-closed pre-write; `v0.0.0` fallback deleted).
- AC-W01-E04-S001-02 — **verified** (EV-W01-E04-S001-002; the T002 harness ran init → tidy → download
  → build → boot-smoke to success for BOTH the source-built and released CLI paths, hermetically, with
  per-step diagnostics; reuse clause satisfied via the shared scaffold primitive).
- AC-W01-E04-S001-03 — **verified** (EV-W01-E04-S001-003, fail-before/pass-after).
- AC-W01-E04-S001-04 — **verified** (EV-W01-E04-S001-004; pre-T003 failure matched the
  `kernel/authz/registry.go:88-90` rejection verbatim).
- AC-W01-E04-S001-05 (scope addition) — **verified** (EV-W01-E04-S001-005, fail-before/pass-after).

## Task completion

T001, T002, T003, T004, T005: done + verified (see `tasks/index.md`).

## Artifact completeness

ART-W01-E04-S001-001 through -006: produced (uncommitted working-tree changes at HEAD
`05dce5c8a548f7dce3222637ab2c82024236a2a0`; the wave conductor owns commits). See `artifacts/index.md`.

## Evidence completeness

EV-W01-E04-S001-001 through -005 complete per `governance/evidence-policy.md` (SHA, command,
environment, result, checksummed logs, preserved failed runs). See `evidence/index.md`.

## Unresolved findings

None in-story. Informational hand-offs: (a) W01-E04-S002's FBL-03 can recommend closing wowsociety's
SF-7 upstream finding (`12-sf-7-init-gomod-invalid-and-gitignored-local-overlay.md`) once this tree is
committed; (b) the former T005 residual (in-scaffold `config validate` delegation leg broken by the
go.mod defect) is cleared by T001 — scaffolds now carry a resolvable framework requirement on every
path.

## Accepted risks

- RISK-W01-005: **closed** — the template and `TestGenCRUDPermissionKeys` were corrected together,
  with the test-lock proven by a preserved pre-fix run before removal.
- Task-002's environment-sensitivity risk (network/`go mod download` flakiness): **closed by
  construction** — the harness is fully offline (file:// proxies, GOPRIVATE/GONOPROXY neutralized).
- Task-001's reachability-check risk (accepting a bad commit / rejecting a good one): **mitigated** —
  reachability is delegated to the go tool itself (`go list -m <module>@<revision>`), with both accept
  and reject cases regression-tested.

## Deferred work

DX-02's full P1/Wave-4 rewrite (disable-vs-minimal-slice, status column, real handlers) and DX-04
(golden consumer + upgrade matrix, W06) remain deferred exactly as `story.md` "Out of scope" records —
nothing silently dropped.

## Reviewer conclusion

Pending — independent review (mandate §14) not yet performed. The reviewer MUST specifically confirm:
(1) `TestGenCRUDPermissionKeys` was updated (RISK-W01-005), not only the template; (2) no code path in
`init_cmd.go`/`init_version.go` writes an unverified version (the only remaining `v0.0.0-` literal is
the inert placeholder written exclusively with a replace directive); (3) the harness's released path
exercises real proxy resolution (file:// GOPROXY), not a replace shortcut.

## Acceptance authority

Pending per this epic's `acceptance.md`.

## Closure date

2026-07-13 (story-level work complete; awaiting review gate).

## Final status

`verified` — all five ACs verified with evidence. Must not move to `accepted` before the wave-level
independent review and acceptance authority sign-off.
