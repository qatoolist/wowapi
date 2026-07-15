---
id: W01-E04-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E04-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S002 — Artifacts index

Per mandate §9.2. All five artifacts produced 2026-07-13, working tree at revision
`0a31186cada5c275a588c74081cf977adf346e61` (HEAD later advanced to `05dce5c8` with an impl/-only
delta — carry-forward rationale in `../evidence/index.md`; conductor commits at wave close).

## Implementation

| Artifact ID | Title | Type | Description | Source requirement | Producing task | Path | Version | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E04-S002-001 | Plan document §6 DX-05 row correction | design document (doc diff) | Corrects the plan document's §6 traceability-matrix row for DX-05 to match §9's execution record | T-DOC-01 | W01-E04-S002-T001 | `docs/implementation/premier-framework-implementation-plan.md` (§6, external to `impl/`); diff at `../evidence/reviews/ev-001-t001-plan-doc.diff` | n/a | produced |
| ART-W01-E04-S002-002 | DX-05 T3 blueprint-11 example decision table | design document | Per-example implement-or-delete decisions reconciling blueprint-11's CLI examples against `internal/cli/cli.go`; applied edits in `docs/blueprint/11-framework-distribution-and-consumption.md`, diff at `../evidence/reviews/ev-002-t002-blueprint11.diff` | DX-05 | W01-E04-S002-T002 | `dx05-t3-cli-example-decision-table.md` | n/a | produced |
| ART-W01-E04-S002-003 | DX-05 T4 version-gate design note | design document | Describes the version-compatibility gate for mutating generator commands, reusing S001's version-verification plumbing | DX-05 | W01-E04-S002-T002 | `dx05-t4-version-gate-design-note.md` | n/a | produced |
| ART-W01-E04-S002-004 | DX-05 T5 deferral note | design document | Records DX-05 T5's deferral to W06/REL-03 with cross-reference | DX-05 | W01-E04-S002-T002 | `dx05-t5-deferral-note.md` | n/a | produced |
| ART-W01-E04-S002-005 | FBL-03 wowsociety upstream register coordination recommendation | design document (coordination note) | PROD-level recommendation for PF-2 (contingent on S001), PF-6, RFF-001 register-entry corrections | FBL-03 | W01-E04-S002-T003 | `fbl03-wowsociety-register-coordination-recommendation.md` | n/a | produced |

## Retention

All artifacts above are retained for the lifetime of the programme's traceability record per
`governance/artifact-policy.md`; none are large generated outputs requiring checksum-only registration.
