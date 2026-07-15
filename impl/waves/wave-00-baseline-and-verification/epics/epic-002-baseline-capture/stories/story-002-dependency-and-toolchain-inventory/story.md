---
id: W00-E02-S002
type: story
title: Dependency and toolchain inventory
status: accepted
wave: W00
epic: W00-E02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-04
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W00-E02-S002-01
  - AC-W00-E02-S002-02
  - AC-W00-E02-S002-03
artifacts:
  - ART-W00-E02-S002-001
  - ART-W00-E02-S002-002
evidence:
  - EV-W00-E02-S002-001
  - EV-W00-E02-S002-002
  - EV-W00-E02-S002-003
decisions: []
risks:
  - RISK-W00-E02-001
---

# W00-E02-S002 — Dependency and toolchain inventory

## Story ID

W00-E02-S002.

## Title

Dependency and toolchain inventory.

## Objective

Capture, at the current repository HEAD, the complete `go.mod` direct/indirect dependency list and
the pinned versions of the toolchain (`golangci-lint`, GoReleaser, Trivy, `goose/v3`), and
cross-check the dependency list against
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` ("REVIEW") §L's approved
register and §M's rejected register, with zero unexplained drift.

## Value to the framework

This story produces the current-HEAD, registered-evidence baseline that every later wave's
dependency-related work must cite rather than re-deriving from prose. Without it: FBL-04
(W04-E02-S003, adopting `cenkalti/backoff/v5` for duplicated retry logic) has no confirmed premise
that the package is already approved and present-but-unused in the module graph; any future new
dependency proposal has no current registered baseline to diff against; and a license or
maintenance-status regression in an existing dependency (e.g. the `yaml.v3` watch item REVIEW §L
flags) would go unnoticed until it caused a downstream problem. This is inventory and cross-check
work only — it is a framework-generic supply-chain hygiene activity, not tied to any downstream
product's domain.

## Problem statement

REVIEW §L/§M already state an approved-dependency register (all 10 original direct dependencies
approved, three new approvals for reuse work: `cenkalti/backoff/v5`, `hashicorp/golang-lru/v2`,
`sony/gobreaker`) and a rejected-dependency register (viper/envconfig, a new kernel message bus,
password-hashing libraries, custom crypto). Those registers were written against the SHA the
review was authored at. Per `impl/waves/wave-00-baseline-and-verification/wave.md` §"Exit
criteria": "go.mod direct/indirect dependency inventory and pinned tool versions are captured and
cross-checked against REVIEW §L's approved register (all 10 original + backoff/golang-lru/gobreaker
already approved — confirm no drift)." No such cross-check has yet been captured and registered
against the current `impl/` traceability structure — REVIEW §L/§M's claims exist only as prose in
a review document, not as a dated, commit-pinned evidence record. Separately, several toolchain
versions used by later-wave release/quality-gate work (GoReleaser, Trivy) are not yet confirmed
against what this repository's own CI/tooling configuration actually pins, and must be measured
rather than assumed.

## Source requirements

**Front-matter note on ID namespaces:** `naming-conventions.md` treats source-requirement IDs
(e.g. `PERF-01`, `D-01`, `FBL-04`) and acceptance-criteria IDs (e.g. `AC-W00-04`) as distinct,
non-interchangeable namespaces — `traceability-policy.md` generates the requirement-traceability
matrix by walking `source_requirements` back to rows in `impl/analysis/requirement-inventory.md`,
so mixing an `AC-...` ID into that field would corrupt that walk. This story is therefore allocated
by wave-level `AC-W00-04` ("Dependency and toolchain inventory captured") and epic-level
`AC-W00-E02-03` ("Dependency and toolchain inventory captured with zero unexplained drift") — see
`impl/waves/wave-00-baseline-and-verification/acceptance.md` and
`impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/acceptance.md`
respectively — but those AC IDs are recorded here in prose, not in the `source_requirements`
front-matter array.

`impl/analysis/requirement-inventory.md` does not carry a single clean requirement ID whose sole
subject is "dependency inventory" as such. The closest related row is **FBL-04** ("Adopt
`cenkalti/backoff/v5` for duplicated retry," §B, target `W04-E02-S003`) — that story *adopts* the
already-approved dependency into code; this story only *inventories and cross-checks* the register
FBL-04 depends on as a premise. FBL-04 is listed in front matter as a related-but-not-owned source
requirement for traceability; this story does not implement FBL-04 and is explicitly out of scope
for adoption work (see "Out of scope" below).

The primary source for this story's content is **REVIEW §L (Approved dependency register)** and
**REVIEW §M (Rejected dependency register)**, cited directly by name and section per the
`requirement-inventory.md` header's own statement that its sources include the REVIEW document
directly, not only requirement-inventory rows. No fabricated requirement ID has been invented to
cover this gap.

## Current-state assessment

Confirmed by direct inspection of `go.mod` at the time this story was authored (2026-07-12):

- 13 direct dependencies are declared under the top `require` block: `go-playground/validator/v10`,
  `golang-jwt/jwt/v5`, `google/uuid`, `jackc/pgx/v5`, `minio/minio-go/v7`, `pressly/goose/v3`,
  `prometheus/client_golang`, `shopspring/decimal`, `go.opentelemetry.io/otel` (+ 3 more
  `go.opentelemetry.io/otel/...` submodules: `exporters/otlp/otlptrace/otlptracehttp`, `sdk`,
  `trace`), `gopkg.in/yaml.v3`. This is a larger count than REVIEW §L's "10 current direct deps"
  figure because REVIEW §L appears to count the four `go.opentelemetry.io/otel*` entries as a
  single logical "otel×4" dependency rather than four separate `go.mod` require lines — this
  arithmetic reconciliation must be confirmed explicitly during task execution (see `plan.md`
  "Unresolved questions"), not assumed.
- `github.com/cenkalti/backoff/v5 v5.0.3` is present as an **indirect** dependency (confirmed at
  `go.mod` line 25), matching REVIEW §L's statement "`cenkalti/backoff/v5` (MIT, already
  transitive)".
- `github.com/sethvargo/go-retry v0.3.0` is present as an **indirect** dependency (confirmed at
  `go.mod` line 52) and is unused in application code. REVIEW's Stage-7 adjudication corrected an
  earlier auditor claim that this package was absent — this story's fresh `go list -m all` capture
  will re-confirm that fact at current HEAD as part of the inventory, not merely cite the prior
  correction.
- `hashicorp/golang-lru/v2` and `sony/gobreaker` (REVIEW §L's other two new approvals) do not
  appear anywhere in the current `go.mod` — neither as direct nor indirect. This is expected per
  REVIEW §L ("new approvals for reuse work" describes packages approved *for future adoption*, not
  necessarily already present); the inventory task must record their absence as a confirmed fact,
  not silently ignore it.
- `go.yaml.in/yaml/v3 v3.0.4` is present as an indirect dependency, matching REVIEW §L's "watch"
  item noting the community fork of `yaml.v3` is "already indirect."
- Pinned tool versions are **not yet confirmed** by this story's authoring pass: `golangci-lint`
  version is referenced as `v2.11.4` in sibling wave/epic documents (`wave.md`, `dependencies.md`,
  citing `Makefile:16`) but this story's own task must independently re-confirm that pin directly
  from the `Makefile`/CI config rather than trusting the citation secondhand. GoReleaser's pinned
  version and Trivy's pinned version + scanner configuration are **unconfirmed and TBD** as of this
  story's authoring — see "Assumptions" and `plan.md` "Unresolved questions."

## Desired state

A dated, commit-pinned evidence record exists showing: (a) the full `go list -m all` / `go mod
graph` output for the current HEAD; (b) a line-by-line disposition (approved / newly-approved /
undocumented drift requiring escalation) for every direct dependency against REVIEW §L; (c) an
explicit statement that no rejected dependency (REVIEW §M) has entered the module graph; (d) the
confirmed, currently-pinned version of `golangci-lint`, GoReleaser, Trivy, and `goose/v3`, sourced
from this repository's own configuration files rather than assumed or copied from another document.

## Scope

- Running `go list -m all`, `go mod graph`, and `go list -m -json all` (or equivalent) against the
  current repository HEAD and capturing the raw output.
- Cross-checking every **direct** `go.mod` dependency against REVIEW §L's approved register,
  producing an explicit disposition for each.
- Confirming the presence/absence and status of the three "new approvals for reuse work"
  (`cenkalti/backoff/v5`, `hashicorp/golang-lru/v2`, `sony/gobreaker`) and the `yaml.v3`/
  `go.yaml.in/yaml` watch item.
- Confirming no dependency in REVIEW §M's rejected register (viper, envconfig, a message-bus
  client, a password-hashing library) has entered `go.mod`, direct or indirect.
- Inspecting `Makefile`, `.github/workflows/*.yml`, and `.golangci.yml` (or equivalent) to confirm
  the pinned versions of `golangci-lint`, GoReleaser, Trivy, and `goose/v3`.
- Producing two documents: a dependency-inventory document and a tool-version-inventory document
  (registered as artifacts).

## Out of scope

- Adopting `cenkalti/backoff/v5`, `hashicorp/golang-lru/v2`, or `sony/gobreaker` into application
  code. Adoption of `cenkalti/backoff/v5` specifically is FBL-04, targeted at `W04-E02-S003`, a
  later epic. This story inventories and cross-checks only.
- Evaluating `sony/gobreaker` or `lestrrat-go/jwx` as P2 design decisions (REVIEW §K-P2) — those
  are deferred items (`DEF-02`/`DEF-03` per `requirement-inventory.md` §C), out of this story's
  scope.
- Determining or changing Trivy's scanner configuration policy — this story only records what is
  currently pinned/configured, it does not evaluate or change scan policy.
- Re-verifying any of the 8 executed finding-slices (SEC-02, PERF-01, PERF-06, DATA-08 W0, AR-04
  T1, AR-05 T1/T2, AR-06 T1, REL-04 T1-T4) — that is `W00-E01`'s scope entirely, not this story's.
- Capturing coverage/lint/bench/CI-wall-clock baselines — that is `W00-E02-S001`'s scope.
- Writing ADR files for D-01..D-09 — that is `W00-E02-S003`'s scope.

## Assumptions

- The `golangci-lint` v2.11.4 figure cited in `wave.md`/`dependencies.md` (sourced from
  `Makefile:16`) is treated as *unconfirmed by this story* until Task 002 independently re-reads
  the `Makefile` at execution time; it is not asserted as fact in this document.
- GoReleaser's pinned version is **unknown and TBD** as of story authoring. `W06-E03-S001` (REL-01,
  T6) is the later epic/story that needs a verified GoReleaser version for its own release-gating
  work; this story's Task 002 will determine and record whatever version is actually pinned in this
  repository's tooling today, without inventing a number in advance.
- Trivy's pinned version and scanner configuration are likewise **unknown and TBD** as of story
  authoring, to be determined from the repository's actual CI/tool configuration during Task 002
  execution.
- `goose/v3`'s version is assumed determinable directly from `go.mod` (`v3.27.2` per the current
  file) without needing a separate configuration lookup, since it is a Go module dependency rather
  than a standalone pinned binary tool.
- Network access to resolve/list modules (`go mod download`, `go list -m all`) is assumed available
  in the execution environment; this mirrors the epic-level "External dependencies" note in
  `../../dependencies.md`.

## Dependencies

- **None upstream within W00-E02** — per `../../dependencies.md` "Internal sequencing
  recommendation," S003 (ADR-ification) is recommended to follow S001 (quality baselines), but this
  story (S002) is independent of both and may run in parallel with either.
- **Downstream**: `W04-E02-S003` (FBL-04, adopting `cenkalti/backoff/v5`) cites this story's
  dependency inventory as confirming the package is already approved and present-but-unused, per
  `../../dependencies.md` "Downstream" section.

## Affected packages or components

None — this story reads and records repository/tooling state (`go.mod`, `go.sum`, `Makefile`, CI
workflow files, lint configuration). It makes no code or configuration changes.

## Compatibility considerations

Not applicable — this story produces documentation and evidence artifacts only; it does not modify
any dependency version, build configuration, or code path. No compatibility impact.

## Security considerations

This story's cross-check is itself a security-hygiene control: confirming no rejected dependency
(REVIEW §M — a new message bus, a password-hashing library, custom crypto, viper/envconfig) has
entered the module graph, and confirming REVIEW §L's stated license/maintenance rationale (MIT/
BSD/Apache permissive licenses, no unmitigated advisories, `jwt/v5`'s `WithValidMethods` alg-
confusion mitigation) still holds at current HEAD. No new security control is introduced; this
story verifies existing ones remain true.

## Performance considerations

Not applicable. This story performs read-only inventory commands (`go list`, `go mod graph`) and
file inspection; it has no runtime performance impact.

## Observability considerations

Not applicable. This story does not touch logging, metrics, or tracing code or configuration.

## Migration considerations

Not applicable. No data, schema, or configuration migration is involved.

## Documentation requirements

Two documents must be produced and registered per `artifacts/index.md`: a dependency-inventory
document (go.mod cross-check against REVIEW §L/§M) and a tool-version-inventory document
(golangci-lint/GoReleaser/Trivy/goose versions). Both are new documents produced by this story's
tasks, not updates to existing documentation.

## Acceptance criteria

- **AC-W00-E02-S002-01** — `go list -m all` and `go mod graph` output is captured at a named commit
  SHA, and every **direct** dependency in `go.mod` has an explicit, individually stated disposition
  (approved / newly-approved / undocumented drift) against REVIEW §L's approved register, with zero
  entries left unaddressed.
- **AC-W00-E02-S002-02** — Presence/absence of REVIEW §M's rejected dependencies (viper, envconfig,
  a message-bus client, a password-hashing library, custom crypto) is explicitly confirmed as
  "absent, as required" or flagged as drift requiring escalation; presence/absence of the three
  REVIEW §L "new approvals for reuse work" (`cenkalti/backoff/v5`, `hashicorp/golang-lru/v2`,
  `sony/gobreaker`) is explicitly recorded.
- **AC-W00-E02-S002-03** — Pinned versions of `golangci-lint`, GoReleaser, Trivy, and `goose/v3` are
  captured from this repository's own configuration/tooling files (not assumed or copied
  secondhand) and registered as an evidence-backed tool-version-inventory document; any version
  that cannot be confirmed (e.g. if GoReleaser or Trivy turns out to have no pin at all in this
  repository) is explicitly recorded as such rather than silently omitted.

## Required artifacts

- A dependency-inventory document (post-implementation-stage artifact per `artifact-policy.md`,
  since it records the result of investigation work rather than being consumed as an input) —
  registered in `artifacts/index.md`.
- A tool-version-inventory document — registered in `artifacts/index.md`.

Both artifacts are declared as "not yet produced" in `artifacts/index.md` at story-creation time
per the repository's Adaptation 2 (`impl/governance/naming-conventions.md`).

## Required evidence

- Raw `go list -m all` / `go mod graph` / `go list -m -json all` command output, commit-pinned.
- A cross-check diff or table showing each direct dependency's disposition against REVIEW §L/§M.
- Command output confirming the pinned versions of `golangci-lint`, GoReleaser, Trivy, and
  `goose/v3` (e.g. `golangci-lint --version`, the relevant `Makefile`/CI-workflow grep output).

All declared as "not yet produced" in `evidence/index.md` at story-creation time.

## Definition of ready

This story satisfies `impl/governance/definition-of-ready.md`'s story checklist: it targets one
coherent capability (dependency/toolchain inventory, not adoption or evaluation); scope and
out-of-scope are both stated above; a concrete approach exists (`plan.md`); it is independently
reviewable and verifiable without depending on another story's completion first; it traces to
`AC-W00-04`/`AC-W00-E02-03` and cites REVIEW §L/§M directly since no single clean
`requirement-inventory.md` ID exists for this exact subject (flagged explicitly above, not
silently resolved); acceptance criteria are numbered and measurable; dependencies and assumptions
are recorded; required artifacts/evidence are anticipated; compatibility/security/performance/
observability/migration considerations are each addressed (mostly as explicit not-applicable).

## Definition of done

This story will satisfy `impl/governance/definition-of-done.md` when: both required artifacts are
registered with full `artifact-policy.md` fields; all required evidence is registered with full
`evidence-policy.md` fields including a pinned commit SHA; every acceptance criterion has a `pass`
verification result; `deviations.md` is finalized (either "no deviations" or every deviation fully
recorded); `closure.md` is complete; and the independent-review checklist
(`definition-of-done.md` "Independent-review checklist") has passed clean, in particular confirming
this story did not silently drop the fact that no single clean requirement-inventory ID exists for
"dependency inventory" and did not invent one.

## Risks

- **RISK-W00-E02-001** (epic-level register, `../../risks.md`) — the approved-dependency cross-check
  misses drift because `go list -m all` output is only spot-checked rather than diffed line-by-line
  against REVIEW §L's named list. Mitigation: Task 001 must perform a full enumeration of every
  direct dependency, not a sample.

## Residual-risk expectations

Even after full enumeration, a residual risk remains that a *transitive* (indirect) dependency not
named in REVIEW §L/§M could carry a license or security concern REVIEW's authors did not evaluate,
since REVIEW §L's approval statement is scoped to "10 current direct deps" — this story's scope is
correspondingly a full inventory of indirect dependencies (for the record) but a disposition
judgment only against direct dependencies, matching REVIEW §L's own scope. This residual risk is
carried forward, not resolved, by this story; it is not a defect in this story's completion.
