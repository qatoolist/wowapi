---
id: W02-E03-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E03-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E03-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E03-S001-001 | Locked-counter/sequence-row version-allocation mechanism | source-code package | implementation | Replaces inline `MAX(version)+1` in both `kernel/artifact.Generate` and `kernel/document.InitiateUpload` | DATA-05 | W02-E03-S001-T001 | TBD at implementation time | not yet produced |
| ART-W02-E03-S001-002 | Upload-session schema and table | schema | implementation | Durable `kernel/document` upload-session record: expiry, checksum/size, storage key, status, cleanup ownership | DATA-05 | W02-E03-S001-T002 | TBD at implementation time | not yet produced |
| ART-W02-E03-S001-003 | Atomic CAS confirmation logic | source-code package | implementation | CASes upload-session status and version allocation together | DATA-05 | W02-E03-S001-T003 | TBD at implementation time | not yet produced |
| ART-W02-E03-S001-004 | Scheduled GC sweep mechanism | source-code package | implementation | Removes expired/unreferenced upload objects with metrics/audit | DATA-05 | W02-E03-S001-T004 | TBD at implementation time | not yet produced |
| ART-W02-E03-S001-005 | Version-allocation, session-lifecycle, and GC documentation | documentation | post-implementation | Documents the counter mechanism, upload-session lifecycle, and GC sweep grace window/scheduling | DATA-05 | W02-E03-S001-T001, W02-E03-S001-T002, W02-E03-S001-T004 | TBD at implementation time | not yet produced |
