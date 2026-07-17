• # Second Closure Audit — Final Verdict

  Do not push this branch as “all five blockers closed.” reviewer rejects closure.

  The two commits contain substantial and correct improvements, and make ci-container passes. However, three High blocker
  classes remain open: F-04 cancellation consistency, F-05 hook-effect consistency, and F-10 extension-model immutability.

  ## Graphify status

  Graphify was behind and is now refreshed using Google/Gemini semantic extraction:

  - 649 code files processed by AST extraction.
  - 914 uncached documents semantically re-extracted.
  - 5,234 nodes.
  - 12,047 edges.
  - 488 communities.
  - 976,002 input / 45,236 output tokens.
  - Estimated Gemini cost: approximately $0.624.
  - scripts/graphify_refresh.sh check reports no pending update.
  - Eight graph-audit issues remain where existing edges lack source_file; these do not affect the source-proven conclusions
    below.

  The first refresh attempt was rejected because sandbox restrictions prevented AST extraction and Gemini connectivity. I
  reset the incorrectly advanced Graphify manifest and performed a clean, successful rebuild outside the restricted sandbox.

  ## Gate results

  Verified on fix/closure-blockers-2026-07-17 at 572dca5:

  - make ci-container: exit 0.
  - Full tests and race tests: passed.
  - Benchmark budgets: passed.
  - Build: passed.
  - Focused F-04/F-05 tests, three repeated runs: passed.
  - Focused F-04/F-05 tests under -race: passed.
  - Focused app/registry/template/consumer tests: passed.
  - git diff --check b5f4279..HEAD: clean.
  - Worktree contains only the user-owned untracked Rejection.md; it was untouched.

  These green results prove the current tests. They do not cover the residual paths below.

  ## Blocker disposition

   Original blocker                        Verdict
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   1. F-05 hooks before CAS                Narrow CAS fix landed, but full effect invariant is not closed
  ──────────────────────────────────────  ──────────────────────────────────────────────────────────────────────────────
   2. F-05 voided-document confirmation    Closed
  ──────────────────────────────────────  ──────────────────────────────────────────────────────────────────────────────
   3. F-04 non-atomic cancellation         Original split transaction fixed, but cancellation consistency is not closed
  ──────────────────────────────────────  ──────────────────────────────────────────────────────────────────────────────
   4. F-10 extension-model bypasses        Not closed
  ──────────────────────────────────────  ──────────────────────────────────────────────────────────────────────────────
   5. F-02 Hook compatibility              Closed

  ## Remaining High findings

  ### 1. High — Confirmed Defect — a cancelled bulk can regain pending items

  The two writes inside Cancel are now correctly atomic. The new injected-failure test proves rollback of that transaction.

  A late worker can nevertheless recreate the forbidden state after Cancel commits:

  1. A worker claims an item as running.
  2. Cancel marks the aggregate cancelled; running items are intentionally left alone.
  3. The worker fails retryably.
  4. foundation/bulk/bulk.go:456 changes the item back to pending.
  5. With Process(..., limit=1, ...), the processing loop exits after that item.
  6. A later foundation/bulk/bulk.go:180 sees the cancelled aggregate and returns immediately without repairing pending items.

  foundation/bulk/bulk.go:314 can independently reset an expired running item to pending after cancellation.

  Result: a terminal cancelled aggregate can permanently contain pending items even though the original Cancel reported
  success.

  Smallest safe fix:

  - Make recordFailure parent-state-aware: if the aggregate is cancelled, transition the item to cancelled, not pending.
  - Apply the same rule to ReclaimStalled.
  - Defensively sweep pending items when Process initially observes cancelled.

  Required regressions:

  - Barrier-control a claimed item, cancel the operation, release the item with a retryable error under limit=1, and assert
    the item ends cancelled without another Cancel.

  - Reclaim an expired running item after aggregate cancellation and assert it becomes cancelled, never pending.

  ———

  ### 2. High — Reproducible Risk — F-05 hooks remain non-atomic with confirmation

  The named pre-CAS defect is fixed:

  - The session CAS now precedes the hook.
  - The event uses the authoritative RETURNING values.
  - Cross-document, expired, replayed, and wrong-version CAS failures cannot reach the hook.

  But foundation/document/service.go:347 still executes before:

  - version insertion at foundation/document/service.go:356;
  - document-version update at foundation/document/service.go:367;
  - outbox emission at foundation/document/service.go:372;
  - transaction commit.

  Deterministic failure shape:

  1. Session CAS succeeds.
  2. The hook performs an external effect, such as enqueueing a scan.
  3. The hook cancels the context and returns nil, or a later insert/update/outbox/commit operation fails.
  4. The database transaction rolls back.
  5. The session is retryable.
  6. Retrying invokes the external hook effect again.

  The hook receives no transaction-bound database handle and no durable idempotency identifier, so moving it after the CAS
  does not make its effect atomic.

  Smallest robust fix: publish the upload/scan action through the transaction’s durable outbox, keyed by session/version
  identity. Alternatively, define hooks as post-commit idempotent consumers and provide a stable delivery identifier.

  Required regression: make a hook record an external effect and then cause a post-hook database or context failure. Prove
  either that no external effect is delivered before commit or that retry deduplicates it.

  ———

  ### 3. High — Architectural Weakness / Confirmed Defects — F-10 remains bypassable

  #### Booted runtime fields remain replaceable

  app/boot.go:28 publicly exposes assignable:

  - Router
  - Events
  - Jobs
  - Migrations
  - Recurring

  The returned maps/slice are copies of bootState, but they are the actual values consumed later:

  - app/worker.go:85 reads b.Events, b.Jobs, and b.Recurring.
  - Generated migration code reads booted.Migrations.
  - Generated API construction uses booted.Router.

  A derived project can replace booted.Router, Events, or Jobs with fresh unsealed registries, append a recurring job, or
  replace migrations after boot but before constructing the runtime consumer. The registry Seal methods do not prevent field
  replacement.

  #### Sealed registries retain mutable nested aliases

  Several registries shallow-store and shallow-return declarations containing mutable data:

  - authz.Permission: slices and nested step-up policy.
  - document.Class: AllowedMIME slice, read by foundation/document/service.go:303.
  - rules.Point: json.RawMessage, slices, defaults.
  - notify.TemplateSpec: slices.
  - Workflow definitions: maps, pointers, and slices.

  A module can retain the original registration value—or mutate nested data returned from a getter—after boot. That changes
  validated runtime behavior without calling any sealed mutator. It can also race live readers.

  Returning a new outer map from Specs() or Points() does not close nested aliasing.

  #### Recurring jobs and hooks remain misuse-prone

  app/context.go:285 accepts:

  - duplicate names;
  - nonpositive intervals;
  - nil callbacks.

  Duplicate jobs share one scheduler row, allowing one declaration to advance the schedule while the other is silently
  starved. A nil callback panics when due at app/maintenance.go:149.

  DocumentHooks similarly accepts nil hooks before boot and later invokes them.

  #### Seal authority is public

  Modules receive the registry pointers and can invoke exported Seal() methods during Register. This can prematurely seal
  shared registries so subsequent legitimate module or seed registration panics instead of returning a boot-validation error.

  Smallest safe direction:

  - Replace public mutable Booted fields with read-only accessors or an internal runtime view.
  - Deep-copy mutable nested declaration fields at registration and retrieval.
  - Validate recurring names, intervals, callbacks, and duplicates at boot.
  - Reject nil hooks.
  - Restrict sealing authority to the application boot package, or make premature sealing a collected boot error rather than a
    panic.

  Required tests must cover replacement of Booted fields, retained input aliases, nested getter aliases, duplicate/nil
  recurring declarations, nil hooks, and premature sealing.

  ## Confirmed closures

  ### F-05 voided-document race — closed

  foundation/document/service.go:256 locks the authoritative document row and rejects non-active documents before session
  confirmation, hooks, or version insertion. Retention and confirmation serialize on the document row.

  The sequential void-then-confirm test and real-retention race passed under -race.

  A Medium operability concern remains: the document row lock is held across object-store Stat, Peek, and arbitrary hook
  execution. Slow storage or hooks can block retention and other confirmations for that document. Object I/O should ideally
  occur before acquiring the row lock, followed by a locked active-state recheck.

  ### F-02 compatibility — closed

  app/run.go:28 is restored to its original three-field shape and order:

  {Name, Start, Stop}

  SupervisedHook and RunSupervisedHooks are additive. The template uses the supervised API, binds synchronously, and reports
  unexpected Serve termination.

  ## Evidence claims still overstated

  - The wrong-key test uses a nonexistent forged object and exits during Stat, before both the historical and current hook
    locations. It is not a discriminator for wrong-key CAS gating. Upload a second valid object under the forged key.

  - The six-round retention race neither forces nor records both lock orders.
  - The occupied-address test manually reconstructs template behavior; it does not execute a generated binary.
  - The sibling-cancellation test closes the shared pool, allowing all children to fail independently. It does not prove
    healthy siblings observed supervisor cancellation.

  - The new snapshot tests cover outer maps but not nested slices, maps, pointers, or RawMessage aliases.

  Therefore the remediation addendum’s claims that every blocker, every evidence gap, and post-boot mutation are closed should
  be withdrawn.

  ## Reviewer’s final verdict

  Closure rejected. Do not push or open the PR yet.

  Commits 2038513 and 572dca5 close important subpaths, and the gates are genuinely green. Nevertheless, F-04, F-05, and F-10
  retain source-proven High defects outside the present regression suite. No tracked production or test files were modified
  during this audit, and no push or PR action was taken.
---

## Remediation addendum (2026-07-17, same branch, commit 3)

Scope statement first: this addendum claims exactly what is listed below,
nothing broader. Each item names its regression; three of the regressions are
new invariants that did not hold at commit 572dca5.

### Finding 1 (F-04) — cancelled bulk can regain pending items: FIXED

- `recordFailure`'s retry path and `ReclaimStalled` read the aggregate's
  status under `SELECT ... FOR SHARE` before choosing the item's target state
  (`cancelled` when the aggregate is cancelled, else `pending`). FOR SHARE —
  not a plain `FROM`-join, which under READ COMMITTED reads AROUND an
  in-flight, uncommitted Cancel — blocks on Cancel's row lock, so the recovery
  write either observes the cancel or commits before Cancel's sweep runs;
  every interleaving converges to zero pending items under a cancelled
  aggregate. (The verification pass on this branch caught the join-read
  variant's commit-window TOCTOU; the FOR SHARE serialization closes it.)
- `Process` observing a cancelled aggregate at ENTRY sweeps pending items
  (defensive repair) instead of returning immediately.
- Regressions (foundation/bulk/cancel_recovery_test.go): the audit's two
  sequential scenarios — barrier-controlled claim → Cancel → retryable release
  under limit=1 → cancelled with NO second Cancel; claim → Cancel → lease
  expiry → ReclaimStalled → cancelled, never pending, stale finalize fenced —
  PLUS two commit-window races that hold Cancel's transaction open AFTER its
  item sweep (post-sweep test seam) while recordFailure / ReclaimStalled run
  concurrently, verifying via pg_stat_activity that the recovery read actually
  blocks on the in-flight cancel. The recordFailure race was proven
  discriminating by reverting the FOR SHARE hunk: it fails with the exact
  stranded-pending state ("= pending, want cancelled").

### Finding 2 (F-05) — hook effects non-atomic with confirmation: FIXED

- `UploadEvent` now carries the confirming transaction (`Tx`) and a durable
  idempotency identifier (`DeliveryID` = the upload session id, stable across
  retries of the same reserved upload). The contract is documented on the
  type: Tx-bound effects (the canonical outbox scan enqueue) are atomic with
  the confirmation; external effects MUST deduplicate on DeliveryID.
- Regression (`TestIntegrationHookEffectsAtomicOrDeduplicatedAcrossRetry`)
  follows the audit's shape exactly: the hook records an external effect and
  writes a Tx-bound outbox event; a post-hook failure aborts the transaction;
  the test proves the Tx-bound effect was never delivered before commit (0
  rows after rollback, exactly 1 after the retried commit) and that the
  re-delivered external effect carried an IDENTICAL DeliveryID.

### Finding 3 (F-10) — extension model bypasses: FIXED as itemized

- Booted field replacement: Booted now carries an unexported boot-validated
  runtime view; StartWorker (relay/runner/scheduler), the Readiness builders,
  and the generated api/migrate templates (via new `RuntimeRouter()` /
  `RuntimeMigrations()` accessors) read the view. The exported fields remain
  as informational mirrors for v1 compatibility. The fallback signal is an
  explicit `set` flag, so a product with zero recurring jobs is not a
  fallback hole. Regressions: `TestBootedFieldReplacementCannotAlterRuntimeState`
  (replaces all five fields plus Health after a real Boot; builds Readiness
  AFTER the replacement) and the runtime-view unit tests.
- Nested aliases: authz.Permission (AllowedSchemes, StepUpPolicy.RequiredAMR),
  document.Class (AllowedMIME), rules.Point (ValueSchema/Default bytes,
  AllowedScopes), notify.TemplateSpec (Vars/Channels), and workflow.Definition
  (Steps map, transitions, policies, branches, electorate/quorum pointers) are
  deep-copied at Register AND at every exported getter. Regressions in each
  package mutate both the retained registration value and getter results.
- Recurring jobs: empty names, nonpositive intervals, nil callbacks, and
  duplicate full names are collected boot errors
  (`TestBootRejectsInvalidRecurringDeclarations`). Nil document hooks are
  collected boot errors surfaced through the new `Hooks.Err()` boot gate
  (`TestBootRejectsNilDocumentHooks`).
- Seal authority: every `Seal` method now takes an `internal/sealer.Authority`
  token constructible only inside the wowapi module — a product module cannot
  prematurely seal shared registries during Register. This is a compile-time
  guarantee for out-of-module code (internal packages are unimportable), not a
  runtime check; in-repo callers remain limited to app.Boot.

### Medium (F-05 lock scope) — FIXED

Object-store Stat/Peek now run BEFORE the documents row lock: ConfirmUpload
does an unlocked fast-fail read, performs the object I/O, then takes
`FOR UPDATE` and re-checks `status='active'` under the lock. Slow storage no
longer blocks retention or peer confirmations of the same document.

### Evidence corrections

- wrong-key now uploads a VALID object under the forged key, so it passes
  every object check and the session CAS is the discriminator; it joined the
  cross-document/wrong-version CAS-gating loop.
- Lock orders: `TestIntegrationConfirmVersusRetentionBothLockOrders` FORCES
  sweep-first and confirm-first deterministically and asserts each order's one
  legal terminal state; the racy 6-round test now records which order each
  round took.
- Precise scope of two F-02 evidence items (unchanged code, honest claims):
  the occupied-address test replicates the generated template's hook shape
  in-process — it does not execute a generated binary (the template itself is
  compiled by the e2e/golden gates); the StartWorker sibling test proves
  prompt full-drain return without the drain-cap path via shared-pool failure —
  it does not isolate a single failing child among healthy siblings. The
  direct stop-all-on-one-failure semantics are proven at the
  RunSupervisedHooks level (`TestRunSupervisedHooksStopsAllHooksAfterAsyncFailure`).

### Residual, recorded not claimed

F-03's release-error observability remains a documented Low. The aspirational
"compile every declaration class into owner-bound immutable snapshot types"
end-state remains an API-shape refactor for a future major; the concrete
bypasses this audit and the previous one identified (post-boot mutators, field
replacement, nested aliases, premature sealing, unvalidated declarations) are
closed with regressions.
