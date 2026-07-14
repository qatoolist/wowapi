---
id: PLAN-W00-E02-S002
type: plan
parent_story: W00-E02-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W00-E02-S002 — Dependency and toolchain inventory

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide
enough information. Clearly distinguish confirmed facts, planned changes, and implementation
assumptions." This story produces no code changes; the plan below distinguishes **confirmed
facts** (already true at story-authoring time, verifiable by direct inspection), **what must be
freshly measured** during task execution, and **unresolved questions** that task execution must
answer rather than assume.

## Proposed architecture

Not applicable in the code-architecture sense — this story is a read-only inventory and
cross-check activity producing two documentation artifacts. No production code, package, or
interface is touched.

## Implementation strategy

Two sequential tasks, each independently completable:

1. **Task 001** — run the module-inventory commands, capture raw output, and produce the
   line-by-line cross-check against REVIEW §L (approved) and §M (rejected).
2. **Task 002** — inspect the repository's own tooling configuration (`Makefile`, CI workflow
   files, lint configuration) to confirm pinned versions of `golangci-lint`, GoReleaser, Trivy, and
   `goose/v3`, and produce the tool-version-inventory document.

Both tasks are independent of each other (neither's output is an input to the other) and may run
in parallel if resourcing allows; they are sequenced here only for narrative clarity.

## Expected package or module changes

None. No `go.mod`, `go.sum`, or source package is modified by this story.

## Expected file changes where determinable

**Confirmed — new files this story creates** (documentation, not code):

- A dependency-inventory document (exact path to be finalized when Task 001 executes; expected
  under this story's `artifacts/` tree per `artifact-policy.md`, e.g.
  `artifacts/post-implementation/dependency-inventory.md`, created on first real content per
  Adaptation 2).
- A tool-version-inventory document (same convention, e.g.
  `artifacts/post-implementation/tool-version-inventory.md`).
- Corresponding entries added to this story's `artifacts/index.md` and `evidence/index.md` (already
  scaffolded at story-creation time with "not yet produced" status; task execution updates them to
  point at the real files once produced).

**No existing repository file is modified.** `go.mod` itself is read, not written.

## Contracts and interfaces

Not applicable.

## Data structures

Not applicable.

## APIs

Not applicable.

## Configuration changes

None. This story reads existing configuration (`Makefile`, `.golangci.yml` or equivalent, CI
workflow YAML) but does not change it.

## Persistence changes

Not applicable.

## Migration strategy

Not applicable.

## Concurrency implications

None — this is a single-actor documentation/inventory activity.

## Error-handling strategy

Not applicable to production code. Procedurally: if a command in the task breakdown below fails to
run (e.g. `go mod download` cannot reach the network), the failure itself is recorded as evidence
(per `evidence-policy.md`, failed evidence is preserved, not silently retried until it disappears)
and the story's `deviations.md` records the impact if it changes what could be confirmed.

## Security controls

Not applicable — no new control is introduced. Confirming REVIEW §L/§M's dispositions still hold
is itself a supply-chain hygiene check, but it is a verification activity, not a new control.

## Observability changes

Not applicable.

## Testing strategy

Not applicable in the unit/integration-test sense — this story produces no testable code. Its
"testing" is the cross-check enumeration itself: every direct dependency must appear in the
disposition table with an explicit outcome (no sampling), per `RISK-W00-E02-001`'s mitigation.

## Regression strategy

Not applicable — no code path exists to regress. The regression risk this story guards against is
informational: a stale or incomplete inventory being treated as current.

## Compatibility strategy

Not applicable.

## Rollout strategy

Not applicable — this story's output is documentation merged via the normal PR/review process for
this repository (no runtime rollout).

## Rollback strategy

Not applicable — if the produced inventory documents are found to be wrong, they are corrected via
a follow-up task or a new evidence record (superseding the earlier one per `evidence-policy.md`),
not "rolled back" in a deployment sense.

## Implementation sequence

1. Confirm the exact commit SHA the story's tasks will execute against (recorded in every evidence
   record produced).
2. Task 001: run `go list -m all`, `go mod graph`, and `go list -m -json all` (or equivalent);
   capture raw output; build the direct-dependency disposition table against REVIEW §L; confirm
   REVIEW §M's rejected dependencies are absent; confirm presence/absence of the three "new
   approvals for reuse work" and the `yaml.v3`/`go.yaml.in/yaml` watch item; write the
   dependency-inventory document.
3. Task 002: inspect `Makefile`, `.github/workflows/*.yml`, and the lint configuration file(s) for
   pinned tool versions; run version-check commands where a binary is available in the execution
   environment (e.g. `golangci-lint --version`); write the tool-version-inventory document,
   explicitly marking any version that cannot be confirmed as TBD/unconfirmed rather than
   inventing one.
4. Register both documents in `artifacts/index.md`, all raw command output and the cross-check
   table in `evidence/index.md`, update `story.md` front matter (`artifacts`, `evidence` lists) and
   move status along the lifecycle per `governance/lifecycle.md` as work actually completes.

## Task breakdown

- **W00-E02-S002-T001** — go.mod inventory and approved-register cross-check.
- **W00-E02-S002-T002** — pinned tool-version inventory.

## Expected artifacts

- Dependency-inventory document (ART-W00-E02-S002-001, ID to be finalized at production time).
- Tool-version-inventory document (ART-W00-E02-S002-002).

## Expected evidence

- `go list -m all` / `go mod graph` / `go list -m -json all` raw output (EV-W00-E02-S002-001).
- Direct-dependency cross-check table/diff against REVIEW §L/§M (EV-W00-E02-S002-002).
- Tool version-check command outputs for `golangci-lint`, GoReleaser, Trivy, `goose/v3`
  (EV-W00-E02-S002-003).

## Unresolved questions

- **REVIEW §L's "10 current direct deps" vs. `go.mod`'s 13 top-level `require` lines.** REVIEW §L
  appears to count the four `go.opentelemetry.io/otel*` require lines as one logical dependency
  ("otel×4"). Task 001 must state explicitly whether this reconciles cleanly (13 lines = 10 logical
  dependencies, with otel counted as 4-lines-for-1) or whether an actual new direct dependency has
  been added since REVIEW was authored. This is not assumed in `story.md`'s current-state
  assessment — it is flagged there as something requiring confirmation, and this is the task that
  must resolve it.
- **GoReleaser's pinned version.** Unknown as of story authoring. Task 002 must determine it from
  this repository's actual configuration (e.g. `.goreleaser.yml`, a pinned version in
  `Makefile`/CI, or a `go install` version pin) — or explicitly record that no pin currently exists,
  if that is what is found. `W06-E03-S001` (REL-01 T6) is the later story that needs this value
  confirmed; this story does not block on that need, it simply produces the fact.
- **Trivy's pinned version and scanner configuration.** Unknown as of story authoring, same
  treatment as GoReleaser above.
- **Whether `golangci-lint` v2.11.4 (cited secondhand in `wave.md`/`../../dependencies.md` from
  `Makefile:16`) is still the pin at the commit this story actually executes against.** Task 002
  re-reads the `Makefile` directly rather than trusting the secondhand citation.

## Approval conditions

This plan is approved for execution once: (a) the two tasks below are accepted as the correct and
sufficient decomposition (no third task needed, per mandate §12's anti-fragmentation guidance —
inventory and tool-versions are two independently ownable, independently verifiable outputs, which
is the threshold for a task split); and (b) the unresolved questions above are acknowledged as
things the tasks will resolve during execution, not blockers to starting the tasks.
