# AR-06 kernel constructor audit

**Scope:** `kernel/kernel.go` at baseline `733ef3e` plus the W05 working-tree changes.  
**Requirement:** AR-06 T3 / AC-W05-E04-S001-02.  
**Method:** inspect every production constructor call and every anonymous function literal in the file, then verify each closure uses the composed instance rather than constructing a replacement.

## Findings

The file contains 23 executable cross-package `New*` calls: one in `newArtifactWriter` and 22 in the `New` composition root. All are executed while composing the kernel. The text `authz.NewStore()` on the `orgAncestry` comment is not an executable call.

The file contains three anonymous function literals:

1. `orgAncestry` closes over the already-composed `authzStore` and calls `authzStore.OrgAncestors`; it does not construct a second store.
2. The allowlist-change callback records the supplied change through the composed logger; it constructs no infrastructure.
3. The durable-audit transaction callback records through the receiver's composed writer; it constructs no infrastructure.

No other closure captures a fresh infrastructure instance. The historical `orgAncestry` bypass is isolated to the already-fixed site, and the current file contains no remaining instance of the pattern.

## Regression guard

`internal/tools/constructorlint` rejects guarded framework infrastructure constructors outside approved composition packages. `make lint-boundaries` runs the analyzer before the existing import/vocabulary boundary checks, and `analyzer_test.go` includes an aliased `authz.NewStore` adversarial fixture plus the allowed `kernel` composition-root control.
