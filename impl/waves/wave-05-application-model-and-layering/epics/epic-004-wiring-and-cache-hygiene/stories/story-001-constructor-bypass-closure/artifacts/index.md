---
id: W05-E04-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W05-E04-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W05-E04-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2".
Both required artifacts were produced on 2026-07-13.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W05-E04-S001-001 | Constructor-boundary lint tool | source-code package | implementation | Fails CI on a reintroduced ad hoc infrastructure constructor outside composition packages | AR-06 | W05-E04-S001-T001 | `internal/tools/constructorlint`; `Makefile` `lint-constructors` target | produced |
| ART-W05-E04-S001-002 | kernel/kernel.go audit report | documentation | verification | Confirms the closure-captures-a-fresh-instance pattern is isolated to the already-fixed site | AR-06 | W05-E04-S001-T002 | `evidence/AR-06/kernel_constructor_audit.md` | produced |
