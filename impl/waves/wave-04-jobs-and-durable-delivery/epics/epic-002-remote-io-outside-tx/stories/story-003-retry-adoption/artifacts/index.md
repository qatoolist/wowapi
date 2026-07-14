---
id: W04-E02-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E02-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E02-S003-001 | `cenkalti/backoff/v5` integration (both call sites) | source-code package | implementation | Replaces both hand-rolled retry implementations with the approved library, configured for parity | FBL-04 | W04-E02-S003-T001 | TBD at implementation time | not yet produced |
| ART-W04-E02-S003-002 | Retry-schedule-parity and fault-injection test suites | test suite | implementation | Proves the new library's behavior matches or improves on each prior schedule and behaves correctly under induced failure | FBL-04 | W04-E02-S003-T002 | TBD at implementation time | not yet produced |
| ART-W04-E02-S003-003 | Retry-configuration documentation | documentation | post-implementation | Documents both call sites' new retry configuration | FBL-04 | W04-E02-S003-T001 | TBD at implementation time | not yet produced |
