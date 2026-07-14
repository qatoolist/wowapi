---
id: W00-E02-ACCEPTANCE
type: epic-acceptance
epic: W00-E02
wave: W00
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E02 — Epic-level acceptance

Epic-level acceptance criteria, distinct from any single story's own AC list (mandate §8.3). The
epic cannot be marked `accepted` merely because its three stories are `implemented` — each
condition below requires the full evidence chain per `impl/governance/definition-of-done.md`.

**Satisfaction record (2026-07-13):** all six criteria below are **satisfied**, proven by the
three stories' verification and evidence records (AC-W00-E02-02's pass-as-capture ratified by the
conductor per DEV-W00-E02-S001-001). Independent review gate passed 2026-07-13 (reviewer
W00ReviewGate; conductor concurs).

## AC-W00-E02-01 — Quality baselines captured and registered

Coverage %, full-tree lint state (25-analyzer hit counts per MATRIX CS-23's inventory), bench-
budget state (post-#25 recalibration, 43 entries confirmed), and CI wall-clock per leg are each
captured as a registered evidence record (evidence ID, exact command, commit SHA, environment,
tool versions, date, result) under W00-E02-S001's `evidence/index.md`. Traces to W00-E02-S001.

## AC-W00-E02-02 — Drift against the MATRIX CS-23 snapshot is explicitly flagged or ruled out

The freshly captured 25-analyzer hit counts are compared, analyzer by analyzer, against the MATRIX
CS-23 recorded counts (zero-hit set: sqlclosecheck, rowserrcheck, bodyclose, wastedassign,
makezero, musttag, testifylint; near-zero: noctx 2, copyloopvar 1, gocritic exitAfterDefer 1;
gosec 38 with named triage list). Any analyzer whose fresh count differs from the MATRIX snapshot
is explicitly called out in the evidence record as drift, not silently absorbed into a single
aggregate number. Traces to W00-E02-S001.

## AC-W00-E02-03 — Dependency and toolchain inventory captured with zero unexplained drift

`go.mod` direct/indirect dependency list and pinned tool versions (golangci-lint, GoReleaser,
Trivy, goose/v3) are captured; every direct dependency is cross-checked against REVIEW §L's
approved register (10 original + backoff/golang-lru/gobreaker) with a documented disposition for
every entry — approved, newly-approved, or flagged as undocumented drift requiring escalation.
Traces to W00-E02-S002.

## AC-W00-E02-04 — Nine ADRs exist, each traceable to its REVIEW §F/§U source line

ADR files `adr-001` through `adr-009` exist under W00-E02-S003's `decisions/` directory, one per
D-01..D-09, each stating recommendation, safe default (where the source states one), owner, and
citing the exact REVIEW section (§F row or §U sentence) it formalizes. `decisions/index.md`
registers all nine with D-0N ID, title, status, and owner. Traces to W00-E02-S003.

## AC-W00-E02-05 — No ADR silently adds design content beyond its source

Independent review confirms (per RISK-W00-004 / RISK-W00-E02 mitigation) that no ADR states a
recommendation, safe default, or consequence not already present in REVIEW §F/§U — any necessary
elaboration is explicitly flagged in the ADR as a Wave-00-added clarification, distinct from the
original decision text. Traces to W00-E02-S003.

## AC-W00-E02-06 — Every story passes independent review before acceptance

Each of W00-E02-S001, S002, S003 has passed the independent-review checklist
(`impl/governance/definition-of-done.md` "Independent-review checklist") before moving to
`accepted`: implementation matches plan or deviations documented, evidence references the correct
commit SHA, artifacts registered, no source requirement silently dropped. Traces to all three
stories collectively.

## Acceptance authority

Framework architecture lead (role-based; see `../../wave.md` "Acceptance authority" — no named
human DRI assigned yet).
