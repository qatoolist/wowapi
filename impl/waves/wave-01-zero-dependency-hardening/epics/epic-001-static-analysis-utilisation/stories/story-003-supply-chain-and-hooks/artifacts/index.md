---
id: W01-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E01-S003
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2": lifecycle
subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are created on first
real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E01-S003-001 | Updated `ci.yml` (`go mod verify` step) | configuration example | implementation | `go mod verify` step in the `unit` job, after the tidy check | FBL-07 | W01-E01-S003-T001 | `.github/workflows/ci.yml` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S003-002 | Updated `security-scan.yml` (license-scanning signal) | configuration example | implementation | `license` added to the `trivy` job's `scanners:` list (planned choice carried through unchanged) | FBL-07 | W01-E01-S003-T002 | `.github/workflows/security-scan.yml` | produced 2026-07-13 |
| ART-W01-E01-S003-003 | Updated `.githooks/pre-push` (DB-skip fix) | source-code change (shell script) | implementation | Requires the DB (WOWAPI_REQUIRE_DB=1 + compose-default DSN fallback) with a loud, actionable failure; explicit loud opt-out `WOWAPI_PREPUSH_SKIP_DB=1` | FBL-07 | W01-E01-S003-T003 | `.githooks/pre-push` | produced 2026-07-13 |
| ART-W01-E01-S003-004 | Nightly fuzz-schedule confirmation note | audit note | post-implementation | Inspection chain + observed scheduled run 29229288699; `-fuzz=` gap restated as W07 (REL-04 T8 / PERF-06 T3/T4) scope | FBL-07 | W01-E01-S003-T004 | `artifacts/nightly-fuzz-confirmation.md` | produced 2026-07-13 |
