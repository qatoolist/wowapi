---
id: W06
type: wave
title: Contracts, compatibility, and release gating
status: in-progress
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
included_epics:
  - W06-E01
  - W06-E02
  - W06-E03
  - W06-E04
depends_on:
  - W05
blocks:
  - W07
source_requirements:
  - DX-03
  - DX-04
  - DX-06
  - AR-03
  - REL-01
  - REL-02
  - REL-03
  - AR-05
  - CS-15
  - CS-22
  - CS-23
---

# W06 — Contracts, compatibility, and release gating

## Objective

Give the framework a state-of-the-art module DSL design record (DX-03), a framework-repo-owned golden
consumer that proves the CLI/generator surface end-to-end and across an upgrade (DX-04), an OpenAPI
merge that either captures every 3.1 field or fails loudly instead of silently dropping them (DX-06,
also resolving the AR-03 T2 duplicate per `impl/analysis/conflict-resolution.md` CONFLICT-01), the full
buildable-now slice of the Go/config/migration/container compatibility-gate programme (REL-03a) plus
its still-blocked legs recorded honestly as blocked rather than silently dropped (REL-03b), a release
pipeline that is gated on the exact commit being published (REL-01) with its final admin-only activation
step tracked as a distinct, human-gated story rather than conflated with the ~85% of the work that is
buildable and testable today, security scanning that actually blocks instead of soft-failing (REL-02),
and documentation gates that make a doc example's correctness a CI-enforced property instead of a prose
promise (AR-05 T3/T4/T5, per MATRIX CS-22's full mechanics spec).

## Rationale

`impl/index.md`'s wave map assigns W06 "DX-03 design + DX-04 golden consumer; DX-06 merge +
REL-03a/b diff gates; REL-01/REL-02 release gating (DEC-Q10 activation); doc-example gates (CS-22/AR-05)"
depending on W05, with the explicit rationale "W05 (AR-03 unblocks REL-03b legs)." This wave's four
epics are grouped because they share a single theme — the framework's *external contract with a
consumer* (a golden consumer, an OpenAPI document, a published release, a documentation example) must
be provably correct, not asserted correct. `impl/analysis/wave-allocation-detail.md` §W06 is the
canonical per-epic/story split this wave's tree follows exactly; `impl/analysis/requirement-inventory.md`
supplies the Target column each DX/REL/AR-05 row traces to. Two duplicate-scope resolutions land
entirely inside this wave: CONFLICT-01 (AR-03 T2 / DX-06 T1 — DX-06 is the single owner, AR-03's own
target story W05-E03-S001..S002 proceeds without T2) and the REL-03a/REL-03b split itself, which PLAN's
own text recommends verbatim (`premier-framework-implementation-plan.md:694`): "do not schedule as one
monolithic P1 item, or 5 of 9 sub-tasks silently block the other 4."

## Framework capabilities delivered

- A documented, externally-reviewable design for the state-of-the-art module DSL (`port`,
  `Manifest[T]`, `Operation[Request,Response]`) — explicitly labeled "target, not implemented" per
  AR-05 T5's labeling discipline, with no code produced by this wave (PLAN DX-03-T0's own framing:
  "Design-only").
- A framework-repo-owned golden-consumer fixture, installed via `go install` (not a repo-internal
  import), exercising resource/rule/workflow/event-handler/recurring-job/document-flow/notification/
  webhook generation across two modules, booted against real Postgres/MinIO/Mailpit/OTel, replayed
  through an upgrade-from-previous-version cycle, and wired into CI as a required gate.
- An OpenAPI merge that captures every 3.1 top-level field and every `components.*` field (not just
  `paths`/`components.schemas`), validates the merged document against 3.1.1/2020-12, and gates a
  breaking semantic diff — closing both DX-06's own defect and the identical AR-03 T2 closure contract
  by single ownership.
- The buildable-now half of the compatibility-gate programme (Go public API diff, module compile
  matrix, config-schema compatibility, migration upgrade-from-oldest-supported drill, container
  architecture smoke on every published architecture, SBOM/provenance/signature verification folded
  in from REL-01 T8/T9) plus the still-blocked half recorded with explicit per-leg unblocking criteria,
  not silently deferred.
- A release pipeline gated on the exact commit being published: a versioned gate manifest, a
  `workflow_call`-based `required-gates.yml` used identically by PR/main CI and by release, a
  no-write-permission `build-candidate` job, a `verify_release.sh` with golden-failure tests per
  verified property, and — as a separate, explicitly human-gated story — the final branch/tag/
  protected-environment activation that only a repository administrator can perform.
- Security scanning that blocks instead of soft-failing: Trivy flipped to `exit-code: "1"` with a
  reviewed waiver mechanism, a regression meta-check confirming CodeQL/Scorecard/dependency-review
  actually ran whenever the repository is public, and a local-scanner fallback for "the repository
  goes private again."
- A CI-enforced documentation-example compile gate (`internal/tools/docexamples`, the
  `<!-- doc-example: compile -->` marker convention, `make docs-check`), generated reference docs that
  byte-match AR-03's authoritative model export, and a lint labeling any remaining future-state design
  prose as "target, not implemented."

## Included epics

- **W06-E01 — consumer-and-dsl**: the DX-03 module-DSL design-investigation story and the DX-04 golden
  consumer + upgrade matrix.
- **W06-E02 — api-contract-gates**: the DX-06 OpenAPI merge-complete-or-loud closure (owning AR-03 T2),
  the REL-03a buildable-now compatibility gates, and the REL-03b still-blocked legs with per-leg
  unblocking criteria.
- **W06-E03 — release-gating**: the REL-01 buildable-now release pipeline, the REL-01 remainder +
  DEC-Q10 human-gated protection-activation story, and the REL-02 blocking-security-scans story.
- **W06-E04 — documentation-gates**: the AR-05 T3 doc-example-compile-gate (MATRIX CS-22's full spec)
  and the AR-05 T4/T5 generated-docs-and-labels story.

## Entry criteria

- W05's exit gate satisfied — per `impl/index.md`'s wave map, W06 "Depends on | W05." The specific
  unblocking dependency is AR-03: W06-E02-S003's REL-03b T5 leg and W06-E01-S001's DX-03 design both
  require AR-03's remainder (W05-E03-S001..S002) to have landed, per `impl/analysis/wave-allocation-
  detail.md`'s "Cross-wave sequencing notes" ("W06-E02-S003 legs unblock individually as DX-06 / W05-E03
  / W06-E01-S002 land").
- W00's ADR-ification story (W00-E02-S003) must have produced ADR-005 (`ADR-W00-E02-S003-005`,
  GoReleaser `--skip=publish` split-mode decision) before W06-E03-S001's T6 can implement against it —
  confirmed already produced; this wave's S001 consumes that ADR by reference, it does not mint a new
  one (see `epics/epic-003-release-gating/stories/story-001-exact-commit-release-pipeline/story.md`).

## Exit criteria

- DX-03's design doc and ADR-style decision record exist, labeled "target, not implemented"; no DX-03
  code is produced by this wave (deferred per `requirement-inventory.md`'s own DX-03 disposition:
  "deferred").
- DX-04's golden-consumer fixture installs via `go install`, exercises the named subsystem set across
  two modules, boots against real infrastructure, replays an upgrade-from-previous-version cycle, and
  is wired into CI as a required gate — PLAN DX-04 T1–T5 satisfied.
- DX-06's merge struct covers every OpenAPI 3.1 top-level field and `components.*` field with an
  explicit per-field merge policy, the merged document validates against 3.1.1/2020-12, and a semantic
  diff gate rejects an intentional breaking-change fixture — PLAN DX-06 T1–T3 satisfied, AR-03 T2's
  identical contract closed by the same work.
- REL-03a's six buildable-now tasks (T1, T2, T4, T6, T8, T9) are complete and evidenced; REL-03b's
  three blocked legs (T3, T5, T7) are recorded with explicit per-leg entry criteria naming the exact
  unblocking story, not silently dropped.
- REL-01's buildable-now task set (T1–T8) is complete, evidenced, and testable against a scratch/
  throwaway repository; the final admin-only activation (branch protection, protected release
  Environment, tag protection ruleset — DEC-Q10) is tracked as its own story, correctly recorded as
  blocked-on-human-action, not silently absorbed into a false "done" claim for REL-01 as a whole.
- REL-02's Trivy blocking flip, waiver schema, visibility-guard regression meta-check, and private-repo
  local-scanner fallback are complete and wired into REL-01's manifest.
- The doc-example-compile-gate (AR-05 T3) runs in CI, `make docs-check` exists, and a deliberately
  staled example fails the gate; generated reference docs byte-match AR-03's model export (AR-05 T4);
  remaining future-state design prose is labeled, not silently presented as implemented (AR-05 T5).

## Dependencies

Depends on W05 (full wave) per `impl/index.md`'s wave map. No dependency on W00–W04 beyond the
transitive W00→W05 entry chain. See `dependencies.md` for the full upstream/downstream detail,
including the specific AR-03-unblocks-REL-03b-T5 and AR-03-unblocks-DX-03 sequencing notes, and the
DX-01 T5 harness (W01-E04-S001) that DX-04 T1 reuses as a shared primitive rather than re-building.

## Assumptions

- DX-03 is explicitly a design-only, deferred item at this wave (`requirement-inventory.md`: "Design-
  investigation story only (Wave-4-class per plan)") — this wave produces no DX-03 implementation code,
  consistent with PLAN's own framing that DX-03-T1..Tn implementation is "Deferred — out of near-term
  scope per §12 Wave 4." This is confirmed from source, not an invented scope reduction.
- The OpenAPI-validator dependency for DX-06 T2 (`pb33f/libopenapi` or an equivalent) is not yet
  selected — MATRIX CS-15 states this explicitly: "an OpenAPI 3.1 validator dependency needed for DX-06
  T2 (evaluate `pb33f/libopenapi` — decision at implementation, security-review licence)." This wave's
  W06-E02-S001 records the decision as an implementation-time task, not a pre-made choice.
- REL-01's GoReleaser split-mode approach is governed by the already-ratified ADR-005
  (`ADR-W00-E02-S003-005`), which itself carries an unresolved caveat: "verify against the pinned
  GoReleaser version at implementation time (this is a caveat, not yet independently confirmed)." This
  wave's W06-E03-S001 inherits that caveat rather than re-deciding or silently resolving it.
- DEC-Q10 (repo-admin activation: branch protection, protected release Environment, tag protection
  ruleset) is confirmed `blocked (human)` in `requirement-inventory.md` §B — this wave's W06-E03-S002
  cannot enter `ready`/`in-progress` until a human with repo-admin access resolves it, consistent with
  how REVIEW §F row 10 and §G's layer-by-layer table describe the split between authorable-now work and
  admin-only activation.

## Risks

See `risks.md`. Headline risks: DEC-Q10's human-gated activation blocking W06-E03-S002's own closure
indefinitely if no repo administrator acts; the OpenAPI-validator dependency decision (DX-06 T2) being
made without adequate security/licence review if rushed; REL-03b's three legs remaining blocked past
this wave's own closure if their unblocking stories (E02-S001, W05-E03, E01-S002) are delayed relative
to E02-S003; the GoReleaser split-mode caveat in ADR-005 surfacing a real incompatibility only at
implementation time.

## Quality gates

- DX-04's fail-first evidence is its own T1–T5 acceptance-criteria columns: the golden-consumer fixture
  installs via `go install`, exercises the named subsystem set, boots against real infrastructure, and
  the upgrade-from-previous-version replay is a two-pass integration test, not a single-pass assertion.
- DX-06's fail-first evidence is MATRIX CS-15's own framing: "fixture fragment with a `security` block
  → merged output today lacks it (provable now); seeded breaking-API fixture fails the new diff gate."
- REL-01's machine-acceptance floor (PLAN REL-01's own framing) is required for every task: "a
  deliberately failing check prevents `build-candidate`; changing the tag target changes both manifest
  SHAs; tampering with gate results or candidate bytes is detected; publish rejects any artifact/digest
  absent from the manifest; post-publish verification succeeds from a clean runner with no build
  workspace."
- REL-02's fail-first evidence is a seeded-vulnerability fixture proving fail-then-pass-after-removal,
  per PLAN REL-02 T1's own "Tests" column.
- AR-05 T3's fail-first evidence is MATRIX CS-22's own framing: "run the extractor against a fixture doc
  referencing a phantom API... gate fails; current corrected docs pass."

## Required artifacts

- DX-03: the module-DSL design doc; the ADR-style decision record labeled "target, not implemented."
- DX-04: the golden-consumer fixture scaffold job; the CI gate wiring.
- DX-06: the expanded OpenAPI merge struct; the 3.1.1/2020-12 validator wiring; the semantic-diff gate.
- REL-03a/b: the Go API-diff CI job; the module compile matrix; the config-schema compatibility gate;
  the migration upgrade-drill extension; the container architecture smoke job; the SBOM/provenance-
  verify fold-in; the REL-03b unblocking-criteria record.
- REL-01: `ci/release-gates.yaml` manifest schema + validator; `required-gates.yml`; `build-candidate`
  split; `verify_release.sh` with golden-failure tests; the SLSA-guarantee documentation.
- REL-02: the Trivy blocking flip; the waiver-schema file format + validator; the visibility-guard
  regression meta-check; the private-repo local-scanner fallback.
- AR-05: the `internal/tools/docexamples` extractor tool; the `<!-- doc-example: compile -->` marker
  convention; the `make docs-check` target; the generated reference docs; the future-state-labeling
  lint.

## Required evidence

- DX-03: none beyond the design doc and decision record themselves (this is an investigation story, not
  an implementation story with test evidence).
- DX-04: fixture-installs-via-go-install evidence; per-subsystem coverage evidence; boot-and-exercise
  evidence against real infrastructure; two-pass upgrade-replay evidence; CI-gate-wiring evidence.
- DX-06: per-field-type fixture evidence; structural-validation evidence; seeded-breaking-fixture
  semantic-diff evidence.
- REL-03a/b: seeded-breaking-API-fixture evidence; compile-matrix evidence; seeded-breaking-config-
  fixture evidence; migration upgrade-drill evidence; architecture-smoke evidence; SBOM/provenance
  evidence (shared with REL-01 T8/T9); the REL-03b unblocking-criteria record itself.
- REL-01: manifest-schema fixture evidence; seeded-failure gate-results evidence; tamper-test evidence;
  golden-failure `verify_release.sh` evidence, one per verified property; end-to-end dry-run evidence
  against a disposable repo.
- REL-02: seeded-vulnerability fail-then-pass evidence; waiver-schema fixture evidence; forced-private
  regression-guard evidence; seeded-SAST-fixture fallback evidence.
- AR-05: the adversarial staled-example fixture's fail-then-pass evidence; the generated-docs golden-
  diff evidence; the future-state-labeling lint's fixture evidence.

## Expected implementation outcome

A framework whose external-facing contracts — its own future design intent, its golden consumer, its
published OpenAPI document, its released binaries, its security posture, and its documentation examples
— are each provably correct by a CI-enforced gate rather than asserted correct by prose, with every
still-genuinely-blocked leg (REL-03b's three tasks, REL-01's admin-only activation) recorded honestly as
blocked, not silently reclassified as done.

## Acceptance authority

Release/security-engineering lead for W06-E02/E03 (per PLAN §5.6's own "Accountable role:
release/security-engineering lead" for PF-REL, and MATRIX CS-15's shared DX-06/REL-03 scope);
developer-experience lead for W06-E01/E04 (per PLAN §5.4's "Accountable role: developer-experience
lead" for PF-DX, which DX-03/DX-04/AR-05 all trace back to).

## Closure conditions

All exit criteria satisfied; all four epics' `closure-report.md` accepted; `waves/index.md`'s W06 row
updated to reflect `accepted` status; REL-03b's three blocked legs and W06-E03-S002's DEC-Q10
human-gated activation are each explicitly recorded as either resolved or as an accepted, tracked,
non-silent open item — this wave must not be closed by silently reclassifying a genuinely blocked item
as done.

## Status update (2026-07-16)

`status: in-progress` — corrected from the stale `planned` front matter, which had not been
touched since 2026-07-12 despite substantial story-level activity. Honest summary per
`review-gate-2026-07-16.md`: 8 of 10 stories independently reviewed 2026-07-16. E02-S003 and
E03-S002 remain blocked (E02-S003 on W05 dependencies; E03-S002 on the human DEC-Q10 gate).
E01-S001 is verified-not-accepted, pending W05's AR-01/AR-02 stories reaching `accepted`.
E04-S002 is accepted scoped to T5 only — T4 remains blocked on W05-E03's manifest work reaching
`accepted`. No story's claim was found false or overstated once evidence was examined.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
