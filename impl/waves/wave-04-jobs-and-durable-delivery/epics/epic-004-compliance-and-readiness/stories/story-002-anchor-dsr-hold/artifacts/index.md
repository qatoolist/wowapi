---
id: W04-E04-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E04-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E04-S002-001 | External anchor mechanism | source-code package | implementation | Periodically anchors the audit chain head externally; tamper detectable even if local head_hash compromised | DATA-08 W6-T2 | W04-E04-S002-T001 | `kernel/audit/external_anchor.go` | produced |
| ART-W04-E04-S002-002 | DSR export artifact writer | source-code package | implementation | Replaces retention/engine.go's in-memory map with an encrypted immutable artifact (manifest, per-class results, checksum, expiry, access policy, download audit) | DATA-08 W6-T3 | W04-E04-S002-T002 | `kernel/retention/artifact.go` | produced |
| ART-W04-E04-S002-003 | Central legal-hold enforcement wrapper | source-code package | implementation | Every Dispose/Erase callback passes through this wrapper | DATA-08 W6-T4 | W04-E04-S002-T003 | `kernel/retention/engine.go` | produced |
| ART-W04-E04-S002-004 | RecordClass callback enumeration record | design record | pre-implementation | Enumerates every registered RecordClass/callback in both wowapi and wowsociety, predating the legal-hold wrapper's implementation | DATA-08 W6-T4 | W04-E04-S002-T003 | `evidence/index.md` (this table) and `deviations.md` | produced |
| ART-W04-E04-S002-005 | Explicit per-class DSR status reporting mechanism | source-code package | implementation | Ensures every registered record class appears in the DSR result set with a status, never a silent omission | DATA-08 W6-T5 | W04-E04-S002-T004 | `kernel/retention/engine.go`, `kernel/retention/artifact.go` | produced |
| ART-W04-E04-S002-006 | Anchor/DSR-export/legal-hold/explicit-status documentation | documentation | post-implementation | Documents all four mechanisms and their contracts | DATA-08 W6-T2, T3, T4, T5 | W04-E04-S002-T001, T002, T003, T004 | `deviations.md`, `implementation.md` | produced |
