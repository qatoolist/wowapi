---
id: W04-E04-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E04-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.
No artifact for DX-07 T4 exists in this index — T4 is explicitly out of scope (see `story.md`).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E04-S003-001 | Migration-currency readiness check | source-code change | implementation | Generated readiness template gains a migration-currency check; `/readyz` fails on version lag | DX-07 T1 | W04-E04-S003-T001 | TBD at implementation time | not yet produced |
| ART-W04-E04-S003-002 | Seed/rule/model-hash readiness reporting | source-code change | implementation | Readiness payload reports migration version, seed/rule hash, and model hash | DX-07 T2 | W04-E04-S003-T002 | TBD at implementation time | not yet produced |
| ART-W04-E04-S003-003 | config doctor product-root discovery fix | source-code change | implementation | Replaces CWD-relative os.Stat discovery with go env GOMOD/--project; explicit product-validation-ran reporting | DX-07 T3 | W04-E04-S003-T003 | TBD at implementation time | not yet produced |
| ART-W04-E04-S003-004 | Readiness/config-diagnostics documentation | documentation | post-implementation | Documents all three changes and DX-07 T4's explicit exclusion | DX-07 T1, T2, T3 | W04-E04-S003-T001, T002, T003 | TBD at implementation time | not yet produced |
