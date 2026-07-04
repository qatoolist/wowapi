# wowapi — Review & Learning Register

Every real issue a review (third-party or internal) found in "finished" work, why the workflow missed it,
and the rule now enforced to prevent recurrence. This is the in-repo counterpart of the AI-agent
`review-learnings` memory. **Recurring classes are promoted to mandatory checklist items** in
[quality-gate-checklist.md](quality-gate-checklist.md).

## Why this exists
Across three review passes of goals reported "complete", reviewers found real defects each time. The
common thread was never obscure — it was wiring, coverage, and honesty gaps a competent reviewer sees in
one pass. The register turns those into durable prevention.

## Recurring patterns (each is now a mandatory checklist item)

| Pattern | What it looks like | Prevention rule |
|---|---|---|
| **Built-but-not-wired** | a primitive/artifact exists but nothing on the real path calls it | Trace entry→effect; prove runtime invocation; `check_unwired.sh` |
| **Deferred-claimed-as-done** | "follow-up / future orchestration" presented as complete | Enumerate sub-requirements; Partial ≠ Done |
| **Green-but-hollow tests** | suite passes because meaningful tests SKIP | Force infra (`WOWAPI_REQUIRE_DB=1`); `check_test_skips.sh` |
| **Artifact-doesn't-actually-work** | generated/rendered output never round-tripped | Parse/run/boot every emitted artifact |
| **Missing-required-infra** | feature needs a container/config/migration/grant not provided | Deliver the infra with the feature |
| **Local-not-production** | "works in the test DB" treated as done | Check prod config path, roles, secrets |

## Review pass 1 — post-Goal-2 findings (6)
1. **[High] Runtime authz not enforced** — `RouteMeta` was boot-validated but not consumed per request.
   *Missed because* we checked "the gate exists", not "the gate runs". → **Rule: enforced-at-runtime**
   (checklist B). Fix: `httpx.SecureHandler`/`gateRoute`, `DenyAllAuthenticator`.
2. **[High] `deploy render` emitted un-bootable manifests** (invalid env default; `${VAR}` where a
   secretref was required). *Missed because* the rendered output was never validated. → **Rule:
   artifacts-actually-work** (checklist D).
3. **[Med] Documented config scaffolding missing** (`internal/appcfg`, `tools/configcheck`). *Missed
   because* the decision-log deliverable wasn't cross-checked against what shipped. → **Rule: verify
   documented deliverables exist.**
4. **[Med] Pagination cursor off-by-one** (cursor from the dropped lookahead row). *Missed because* page
   boundaries weren't tested. → **Rule: test boundaries** (checklist E). Proven by revert.
5. **[Med] Green host CI hid skipped DB/E2E tests.** *Missed because* a green suite was taken as proof.
   → **Rule: green-but-hollow / force DB** (`WOWAPI_REQUIRE_DB=1`, checklist F).
6. **[Low] "Complete" claims were deferrals.** → **Rule: no overclaim** (checklist H).

## Review pass 2 — post-hardening findings (6)
- **F1 [High] durable audit vs logging sink not closed** — the durable `audit_logs` writer existed but
  denials only WARN-logged (built-but-not-wired). Fix: `kernel.durableAudit` writes in its own tenant tx
  (Evaluate is read-only). → reinforces **built-but-not-wired**.
- **F2 [High] E2 disposition/DSR was "future orchestration"** (deferred-claimed-as-done). Fix:
  `retention.Registry`+`Engine` + kernel/module/scheduler wiring.
- **F3 [High] O1 was a tracing seam, no adapter/sampling/propagation + no collector container**
  (missing-required-infra + deferred). Fix: `adapters/tracing/otel` + `Inject`/`Extract` + **Jaeger in
  compose**.
- **F4 [Med] R5 partial** — receipts done, channel prefs missing (deferred). Fix: `SetChannelPref` +
  migration 00022.
- **F5 [Med] config CLI never used the generated `tools/configcheck` shim** (built-but-not-wired). Fix:
  delegation + `config diff`.
- **F6 [Med] worker reused runtime DSN + SET ROLE** instead of a dedicated platform login
  (local-not-production). Fix: `db.platform_dsn`.

## Review pass 3 — internal (Independent Review Gate applied to the workflow update)
- **[Low] stray `
` tokens** in generated memory files. *Missed because* file tails weren't
  verified after write. → **Rule: verify artifact well-formedness after generation** (added to
  `review_gate.sh`: scan for stray tags).

## How to add a learning
Append: *what was found · why it happened · why the workflow missed it · prevention · which checklist/
test/script/gate to update.* If the class already appears above, add the instance under it; if a new
class recurs (≥2), add a row to the recurring-patterns table and a checklist item. Mirror significant
learnings into the AI-agent `review-learnings` memory.
