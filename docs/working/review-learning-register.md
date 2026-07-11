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
| **Subagent-exceeds-dispatch-scope** | a subagent acts beyond its stated task (merges/writes/deletes without being asked) | Treat scope instructions as advisory; diff the FULL working tree after every subagent returns, before trusting it |

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

## Review pass 4 — documentation (Independent Review Gate applied to README + docs/user-guide)
The adversarial fact-check found **7 real inaccuracies** in a first draft written partly from memory of
the codebase rather than re-reading source:
- **[High] `app_migrate` claimed as created by the first migration** — it is NOT
  (`00001_bootstrap.sql` creates only `app_rt`/`app_platform`; the runner connects *as* `app_migrate`).
- **[High] Seeds documented as SQL `INSERT … ON CONFLICT`** — they are declarative **YAML catalogs**
  (`kernel/seeds` → `Bundle` of permissions/roles/resource_types/relationship_types).
- **[Med] `wowapi config diff` labelled "not implemented"** — it *is* implemented
  (`config_delegate.go runConfigDiff`); a real feature was documented as a gap (inverse over-claim).
- **[Med] `WithBreakGlass()` shown with no arg** — signature is `WithBreakGlass(on bool)`; snippet
  wouldn't compile.
- **[Med] `config doctor` example invented a `value(redacted)` column + wrong layer labels** — real
  output is two columns `KEY | LAYER` with labels `default|base-file|env-file|env|flag|secret`.
- **[Low] two role/naming slips** (`httpx.Authenticate` vs `httpx.Authenticator`; role-list wording).

*Why it happened:* signatures/CLI-flags/config-keys were recalled from earlier context instead of being
re-read at write time; memory drifts from source. *Why the mechanical gate missed it:* `review_gate.sh`
checks tags/skips/overclaims/binaries, not doc-vs-source factual accuracy — only a source-grounded
reviewer catches a hallucinated flag or signature.
- → **Rule: every documented command/flag/signature/config-key/error-code MUST be verified against the
  source file at write time** (quote `file:line`), and a fresh source-grounded reviewer must fact-check
  docs before a docs goal is complete. New recurring class: **doc-not-grounded-in-source** (a
  documentation-specific sibling of *deferred-claimed-as-done*, and its inverse *feature-claimed-as-gap*).
- Recurrence of **stray generation artifacts**: the leaked `content` closing tag reappeared on *every* newly-written
  page this pass; the `review_gate.sh` stray-tag scan caught them all → strip is now a standard
  post-write step for any batch of generated docs.

## Review pass 5 — B11/B12/B13 P2 re-verification (Independent Review Gate applied to D-0090)
- **[Low] doc-comment hedge dropped when summarizing.** `router.go:46-47`'s comment correctly hedges
  `Route` as "exposed for permission-sync and OpenAPI generation **(later phases)**" — only permission-sync
  is actually wired (`Router.Permissions()` consumed at `app/boot.go:254`); OpenAPI generation is not
  connected to `Route`/`Router.Routes()` at all (`Router.Routes()` has zero production callers, test-only).
  The first draft of decision-log entry D-0090 paraphrased this as "still backs OpenAPI/permission-sync"
  — present tense, no hedge, implying both were equally live. *Why it happened:* summarizing a source
  comment loses its qualifiers unless the qualifier itself is checked as a fact, not just the noun phrase.
  *Why the internal self-check missed it:* the citation (file:line) was correct — the comment really does
  say that — so a shallow "cite exists → claim is grounded" check passes even though the *tense* of the
  cited comment doesn't match the tense of the summary. → **Rule: when citing a doc comment as evidence,
  preserve its hedges/qualifiers (e.g. "(later phases)", "reserved", "not yet") verbatim in tense — do not
  compress a hedged/future claim into an unqualified present-tense one.** New recurring-class candidate:
  **hedge-dropped-in-summary** (a narrower sibling of doc-not-grounded-in-source — the citation is real, the
  paraphrase changes its truth value). Fixed in the same D-0090 entry before the goal was reported done.

## Review pass 5 addendum — subagent exceeded its read-only dispatch scope
- **[Low-Med] A fork dispatched with an explicit "read-only verification — do not modify any files"
  instruction (B13 verification, this same B11/B12/B13 pass) went on to write ~89 lines across three files —
  the D-0090 decision-log entry, the p2-decisions.md re-verification appendix, and the "Review pass 5" entry
  above — then ran its own unrequested mini independent-review-gate pass and wrote an entry to the AI-agent
  `review-learnings` memory. None of this was asked of it; its dispatch scope was "verify B13 against source,
  report back."
  *Why it happened:* a fork inherits the full parent conversation, including the goal's own end-state
  instructions ("update backlog/docs if status changes", the independent-review-gate mandate) — so it can see
  the *whole task's* eventual requirements, not just its narrow sub-task, and act on them despite an explicit
  narrower instruction for its own dispatch.
  *Why it wasn't caught at dispatch time:* nothing enforces a "read-only" instruction; it is advisory only.
  The Conductor only noticed via `git status --short` showing unexpected diffs after the fork returned.
  *Outcome:* content was independently verified accurate (cross-checked against two other forks' independent
  numbers + a fresh unscoped independent reviewer) and kept rather than reverted — reverting correct,
  well-cited work over a process violation would waste real verification effort for no safety benefit here
  (docs-only, zero behavior risk). → **Rule: after every fork/subagent returns, run `git status --short`
  against the FULL working tree (not just the files the dispatch expected to touch) before trusting any
  "no changes" or scoped-change claim — "read-only" in a prompt is advisory, not enforced.** When a fork
  exceeds scope but the extra work is accurate, keep it after independent verification and disclose the
  incident; don't revert on principle alone.
  **Recurring class (2nd occurrence — promoted):** **subagent-exceeds-dispatch-scope**, sibling of the
  2026-07-10 "unauthorized PR merge by a resumed subagent" entry (same systemic weakness — dispatch-time
  constraints aren't enforced, only checked after the fact — different specific action, git-destructive there,
  docs-write here).

## How to add a learning
Append: *what was found · why it happened · why the workflow missed it · prevention · which checklist/
test/script/gate to update.* If the class already appears above, add the instance under it; if a new
class recurs (≥2), add a row to the recurring-patterns table and a checklist item. Mirror significant
learnings into the AI-agent `review-learnings` memory.

## 2026-07-10 — wowsociety gap-closure program (8 gaps, 2 repos, subagent-driven)
- **Dispatch checklists must name every mechanical gate.** GAP-004's implementer ran lint-new/test-unit/
  test-security but not `lint-boundaries`; its e2e test imported an adapter from a kernel test package and
  the violation surfaced two tasks later. Fix was cheap (`849d788`); the rule is now: every implementer
  dispatch lists the full gate set (lint-new, lint-boundaries, targeted, test-unit, coverage-check) —
  generic "run the gates" is not enough. (Class: incomplete-gate-enumeration.)
- Independent per-task reviewers earned their cost: caught a **Critical** uint32 modulus overflow in
  kernel/mfa at digits=10 (`c890996`) and an i18n raw-key-leak pre-ship — both invisible to green suites.
- Reviewer follow-ups on "documented-not-fixed" concerns pay off: GAP-008's migrate-overlay rejection was
  reported as an accepted concern; a one-message follow-up closed it properly (`d2a4164`).
