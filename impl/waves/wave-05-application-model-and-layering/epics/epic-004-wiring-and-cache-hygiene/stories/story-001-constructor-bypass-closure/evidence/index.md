---
id: W05-E04-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W05-E04-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W05-E04-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2".
Both entries were produced on 2026-07-13 against baseline `733ef3e` plus the W05 working-tree changes.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W05-E04-S001-001 | adversarial-lint report (`AR-06/constructor_boundary_lint_test.txt`) | W05-E04-S001-T001 | AC-W05-E04-S001-01 | `go test -v ./internal/tools/constructorlint`; `go test -race ./internal/tools/constructorlint`; `make lint-boundaries` | `733ef3e` + W05 diff | aliased `authz.NewStore` and `authz.NewSQLStore` bypasses rejected; composition root + FBL-01 forwarding shim accepted; focused, race, full-tree, and CI-wired gates passed | produced |
| EV-W05-E04-S001-002 | audit report (`AR-06/kernel_constructor_audit.md`) | W05-E04-S001-T002 | AC-W05-E04-S001-02 | source audit of all constructor calls and function literals in `kernel/kernel.go` | `733ef3e` + W05 diff | 23 executable constructor calls confined to composition; all 3 closures reuse composed dependencies | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state.
