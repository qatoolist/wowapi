---
id: ART-W01-E04-S002-003
type: artifact
title: DX-05 T4 — version-compatibility gate design note (design only, not implemented)
parent_story: W01-E04-S002
producing_task: W01-E04-S002-T002
source_requirement: DX-05 (T4)
status: produced
created_at: 2026-07-13
---

# DX-05 T4 — version-compatibility gate design note

**Deliverable class:** design, not implementation (per `../plan.md` "Implementation strategy" and
`../tasks/task-002-dx05-residual-reconciliation.md`). No production code is changed by this note.

## Problem

`docs/implementation/premier-framework-implementation-plan.md` §5.4 DX-05 T4: "`wowapi version`
fails mutating generator commands on incompatible major/minor pairing — incompatible pairing exits
nonzero pre-write." Today `runVersion` (`internal/cli/cli.go:67-90`) only *prints* the CLI build
version and the product `go.mod` requirement; nothing prevents `wowapi gen crud`, `init`,
`new-module`, or `migrate create` from writing files whose generated code targets a framework
version incompatible with the one the product imports.

## Trigger point

The gate runs at the start of every **mutating generator command** — `init`, `new-module`,
`gen crud`, `migrate create` — before any file is written, mirroring DX-01's
fail-closed-before-any-file-write discipline (plan §5.4 DX-01). Read-only commands
(`version`, `config *`, `lint *`, `seed validate`, `openapi merge`, `deploy render`) are not
gated; `wowapi version` reports the comparison verdict but never fails the build (matching the
plan's own framing that the *mutating commands* are the enforcement point).

## Comparison logic

- Inputs: (a) the CLI's own build version (`internal/buildinfo.Version()`); (b) the product's
  `go.mod` requirement for `github.com/qatoolist/wowapi`, resolved via S001/DX-01's
  version-verification plumbing (`buildinfo.FindGoMod` + the `go list -m`-based resolution check
  DX-01 T1 builds), including `replace`-directive awareness.
- Verdict: **major/minor compatibility class, not exact-version equality.** Same major AND same
  minor ⇒ proceed. Same major, differing minor ⇒ proceed with warning only if the CLI is the
  *older* minor per the v1/N-1 policy (`docs/operations/upgrade-and-deprecation-policy.md`,
  DX-05 T2); otherwise fail. Differing major ⇒ always fail.
- Failure mode: exit nonzero **before any file write**, with a remediation message naming both
  versions and the two fixes (`go get` the matching framework version, or reinstall the CLI at
  the product's pinned version). Never a silent skip or partial generation.
- Unresolvable inputs (no go.mod, un-parseable requirement, DX-01's dirty-pseudo-version case)
  fail closed with the DX-01 remediation text — the gate must not degrade to a warning when it
  cannot determine compatibility.

## Explicit dependency on S001 (this is the load-bearing constraint)

This gate **reuses W01-E04-S001 T001's DX-01 version-verification plumbing** — the `go list -m`
resolution check and version-comparison logic — rather than building a parallel mechanism
(`../story.md` "Dependencies"; `../plan.md` "Implementation sequence"). Status at authoring time
(confirmed with the S001 owner over IRC, 2026-07-13): S001's current W01 slice ships the DX-02
verb fix only; the DX-01 flag/plumbing work is planned in S001's plan but **not landed**.
Therefore T4 implementation MUST NOT begin until that plumbing exists; only this design note is
in-scope for W01-E04-S002. Implementation timing is left open pending S001 per `../plan.md`
"Unresolved questions" (follow-on task, tracked via the story's follow-up items).

## Compatibility constraint on the design itself

The gate must not reject currently-valid pairings (plan.md "Compatibility strategy"): exact-match
and policy-permitted N/N-1 minor skew must continue to pass. The eventual implementation needs an
adversarial fixture — a mismatched major/minor pairing correctly rejected pre-write, plus a
matched pairing accepted — per `../plan.md` "Testing strategy".
