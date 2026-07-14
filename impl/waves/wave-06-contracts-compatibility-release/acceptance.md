---
id: W06-ACCEPTANCE
type: wave-acceptance
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06 — Wave-level acceptance

## AC-W06-01 — Module DSL design recorded, not implemented

A design doc and an ADR-style decision record exist for the state-of-the-art module DSL (`port`,
`Manifest[T]`, `Operation[Request,Response]`), explicitly labeled "target, not implemented" per AR-05's
labeling discipline; no DX-03 implementation code is produced. Traces to W06-E01-S001.

## AC-W06-02 — Golden consumer proves the generator/CLI surface end-to-end and across an upgrade

The framework-repo-owned golden-consumer fixture installs via `go install`; exercises resource, rule,
workflow, event handler, recurring job, document flow, notification, and webhook generation across at
least two modules; boots against real Postgres/MinIO/Mailpit/OTel with authenticated CRUD, async
delivery, restart/retry, and RLS isolation all passing; replays an upgrade-from-previous-version cycle
with contracts re-passing; and is wired into CI as a required gate. Traces to W06-E01-S002.

## AC-W06-03 — OpenAPI merge complete-or-loud; AR-03 T2 duplicate closed by single ownership

The merge struct covers every OpenAPI 3.1 top-level field and every `components.*` field with an
explicit per-field merge policy; the merged document validates against 3.1.1/2020-12; an intentional
breaking-change fixture fails the semantic-diff gate. AR-03's own target story proceeds without its T2
task, per CONFLICT-01's resolution. Traces to W06-E02-S001.

## AC-W06-04 — Compatibility gates: buildable-now set complete, blocked set honestly recorded

REL-03a's six tasks (Go API diff, module compile matrix, config-schema compatibility, migration
upgrade-drill, container architecture smoke, SBOM/provenance-verify fold-in) are complete and evidenced.
REL-03b's three blocked legs are recorded with explicit per-leg unblocking criteria; no leg is silently
dropped or falsely marked complete. Traces to W06-E02-S002, W06-E02-S003.

## AC-W06-05 — Release pipeline gated on the exact published commit; activation tracked separately

The buildable-now release-pipeline work (manifest schema, `required-gates.yml`, `build-candidate`
split, `verify_release.sh` with golden-failure tests, SLSA-guarantee documentation) is complete,
evidenced, and tested against a scratch/throwaway repository. The final branch/tag/protected-environment
activation is tracked as its own explicitly human-gated story, not silently folded into a false "REL-01
done" claim. Traces to W06-E03-S001, W06-E03-S002.

## AC-W06-06 — Security scanning blocks instead of soft-failing

Trivy fails on CRITICAL/HIGH findings with an available fix (`exit-code: "1"`); a reviewed waiver
mechanism exists with owner/rationale/expiry/remediation-link per entry; a regression meta-check
confirms CodeQL/Scorecard/dependency-review actually ran whenever the repository is public; a
local-scanner fallback activates if the repository goes private again. Traces to W06-E03-S003.

## AC-W06-07 — Documentation examples are CI-proven, not merely asserted correct

The doc-example-compile-gate (`internal/tools/docexamples`, the `<!-- doc-example: compile -->` marker
convention) runs in CI via `make docs-check`; a deliberately staled example fails the gate; generated
reference docs byte-match AR-03's authoritative model export; remaining future-state design prose is
labeled "target, not implemented," not silently presented as shipped. Traces to W06-E04-S001,
W06-E04-S002.

## AC-W06-08 — Independent review passed

Every W06 story with a P0/critical priority has passed independent review per mandate §14. W06-E03-S002
is specifically checked for its blocked-entry criteria being honestly stated, not silently bypassed;
W06-E02-S003 is specifically checked for its three per-leg blocked-entry criteria being honestly
recorded, not silently resolved with an invented unblocking date.

## Acceptance authority

Release/security-engineering lead for W06-E02/E03; developer-experience lead for W06-E01/E04 — see
`wave.md` "Acceptance authority" for the full rationale.
