---
id: W03-E05-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E05-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E05-S001 — Artifacts index

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E05-S001-001 | Ratification interim-reject implementation + design-decision record | source-code change + decision record | implementation | `RatifyBy` field on `workflow.Definition`/`Step`; `Validate` rejects non-empty `ratify_by` with a clear interim-posture error; decision recorded in `story.md`/`plan.md` | SEC-02 | W03-E05-S001-T001 | `kernel/workflow/definition.go` | produced |
| ART-W03-E05-S001-002 | Durable override audit-record implementation | source-code change | implementation | `workflow.Runtime.Override` writes a complete `kernel/audit.Entry` in the same tx as the state jump; audit-write failure rolls back the override; `NewRuntime` requires `*audit.Writer` | SEC-02 | W03-E05-S001-T002 | `kernel/workflow/runtime.go` | produced |
