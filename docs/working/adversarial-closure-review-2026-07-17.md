• # Adversarial Framework Closure Review

  ## Executive verdict

  Closure rejected. The remediation is substantial, but the claim that all ten findings are closed and production-correctness
  is restored is not supported.

  Fable’s final quality gate found:

  - F-01, F-03, F-06, F-07, F-08, and F-09 are substantively closed.
  - F-02’s operational defect is fixed, but its public API change introduces a compatibility regression.
  - F-04 remains open because cancellation can partially commit and cannot reliably recover.
  - F-05 remains open through two material bypasses, including a High document-retention consistency defect.
  - F-10 remains materially bypassable; the claimed “whole extension model sealed” invariant is not enforced.

  This is not a superficial “tests are green but I feel uncertain” result. The remaining defects have exact source paths and
  deterministic or reasoned reproductions.

  ## Repository and gate position

  The remediation is no longer a local four-commit branch:

  - PR #33 (https://github.com/qatoolist/wowapi/pull/33) was squash-merged.
  - Workspace: clean.
  - Branch: main.
  - HEAD: b5f4279a30cafd87176f2a75ad7bc3253b123a3f.
  - HEAD == origin/main.
  - git diff --check 5f4ffce..HEAD: clean.
  - All four pre-squash commit objects referenced by the remediation addendum remain locally resolvable.

  Verification performed:

  - Fresh local make ci-container: exit 0, including vet, boundary/lifecycle lint, full tests, race tests, benchmark budgets,
    and build.

  - Fresh focused race execution for runtime remediations: exit 0.
  - PR exact-head CI passed authoritative DB+S3 test and race legs, compatibility gates, migration reversal, golden consumer,
    coverage, tenant-FK gate, and reference-stack smoke.

  - The post-merge main CI run for the exact squash SHA was still in progress at the audit cutoff, with no observed failure:
    main CI run (https://github.com/qatoolist/wowapi/actions/runs/29544828611).

  The green gates are genuine. They simply do not exercise the bypasses below.

  ## Confirmed closure blockers

  ### 1. High — Confirmed Defect — rejected upload confirmations execute hooks

  Affected finding: F-05.

  foundation/document/service.go:294 invokes OnFileUpload before the authoritative session CAS at foundation/document/
  service.go:314.

  The hook event is constructed from caller-supplied:

  - DocumentID
  - VersionNo
  - StorageKey

  Only after the hook succeeds does the code establish that the session is pending, unexpired, and bound to that document,
  version, key, and checksum.

  Trigger:

  1. Submit a cross-document, expired, replayed, wrong-key, or wrong-version confirmation.
  2. Pass the object metadata checks.
  3. The upload hook executes.
  4. The later CAS rejects the confirmation.

  Affected workflows include malware-scan enqueueing and any derived-project hook with an external or non-transactional side
  effect. The closure statement that “all effects use RETURNING values” is therefore false.

  Smallest safe fix:

  - Establish or claim the valid session first.
  - Obtain the authoritative identity from the database.
  - Build the hook event only from authoritative values.
  - Ensure rejected and replayed confirmation attempts cannot invoke hooks.

  Required regression: use a counting/enqueueing hook and prove zero calls for cross-document, expired, replayed, wrong-key,
  and wrong-version confirmations.

  ———

  ### 2. High — Confirmed Defect — confirmation can add an active version beneath a voided document

  Affected finding: F-05.

  The document lookup at foundation/document/service.go:247 loads class and sensitivity but does not require the document to
  remain active and does not lock the document against retention processing.

  Reproduction by reasoned proof:

  1. Initiate an upload session while the document is active.
  2. Upload the object.
  3. Run retention, which voids the document.
  4. Confirm the still-unexpired session.
  5. The session CAS succeeds and inserts an active version under the voided document.
  6. Future retention sweeps select active documents, so this late version is not revisited.

  Violated invariant: a terminal/voided document must never acquire a new active version.

  Impact:

  - Document/version state inconsistency.
  - Retention and blob-lifecycle correctness.
  - Derived upload APIs.
  - Products treating voiding as terminal.

  Smallest safe fix:

  - Lock the authoritative document row during confirmation.
  - Require documents.status = 'active'.
  - Serialize confirmation and retention through a compatible locking order.
  - Perform the check before hooks or version creation.

  Required regressions:

  - Initiate → void → confirm must conflict, create no version, and invoke no hook.
  - A barrier-controlled retention-versus-confirmation race must end with either a valid version that retention voids or a
    rejected confirmation—never an active version under a voided document.

  ———

  ### 3. High — Reproducible Risk — bulk cancellation is a non-recoverable partial-write workflow

  Affected finding: F-04.

  foundation/bulk/bulk.go:274 performs two separate transactions:

  1. Transition the aggregate to cancelled.
  2. Cancel pending items.

  If the second transaction fails, the aggregate is terminal while pending items remain. A retry cannot repair the state
  because cancelled → cancelled is rejected as an invalid transition.

  The new NOT EXISTS completion CAS and legal transition matrix are correct, but they do not close this cancellation path.

  Violated invariant: cancellation must atomically update aggregate and item state, or remain safely retryable.

  Smallest safe fix: execute the aggregate CAS and pending-item update in one tenant transaction. If backward-compatible
  idempotency is needed, an already-cancelled aggregate may perform pending-item cleanup without reopening the aggregate.

  Required regression: inject a failure after the aggregate update and before item cancellation; assert full rollback and
  successful retry.

  ———

  ### 4. High — Architectural Weakness / Confirmed Defects — the F-10 extension model remains bypassable

  Affected finding: F-10.

  Several independent paths contradict the claimed ownership-bound, immutable extension model.

  #### Port misuse does not necessarily fail boot

  app/context.go:421 returns errors for malformed, undeclared, or missing ports, but those errors are not accumulated into
  boot.portErrs.

  A module can ignore the error and return nil from Register; boot then succeeds. The dependency regression at app/
  ports_enforcement_test.go:85 explicitly discards the boot result and only checks the immediate resolution error. It
  therefore does not prove the documented “unsatisfied dependency fails boot” contract.

  Typed-nil implementations are also not rejected by impl == nil.

  #### Booted exposes mutable backing structures

  app/boot.go:27 publicly exposes mutable:

  - Router
  - Events
  - Jobs
  - OpenAPI map
  - Health map
  - Migrations map
  - Recurring slice

  app/boot.go:317 returns the original backing collectors rather than immutable snapshots or read-only interfaces. Derived
  projects can therefore mutate post-boot state directly, bypassing moduleContext.mustBeUnsealed, including creating
  concurrent-map hazards.

  #### Retained contexts expose mutable registries

  The seal protects direct methods such as Health, Migrations, and OpenAPI, but retained contexts still return raw shared
  registries for routes, permissions, resources, events, jobs, rules, workflows, providers, templates, and hooks.

  Some registries accept a caller-provided owner. Some inspection methods, including kernel/resource/resource.go:81 and
  kernel/rules/rules.go:155, return backing maps.

  Duplicate collector registrations can also overwrite existing values without a boot error.

  Smallest safe fix:

  - Compile every declaration class into owner-bound immutable snapshots.
  - Accumulate every port-resolution error into boot validation.
  - Reject typed nils using reflection or a typed port API.
  - Reject duplicate collector registrations.
  - Stop exporting mutable backing maps, slices, and registries.
  - Expose runtime-only read interfaces after boot.

  Required regressions:

  - Ignored Port error still fails boot.
  - Missing provider fails boot.
  - Typed-nil implementation fails boot.
  - Duplicate registrations fail boot.
  - Direct Booted map/slice mutation cannot alter runtime state.
  - Snapshot access cannot mutate registry backing maps.
  - Retained-context mutation tests cover every registry/collector class under -race.

  ———

  ### 5. Medium — API Compatibility Regression — Hook.Failed is not purely additive

  Affected finding: F-02.

  The operational supervision fix is substantively correct, but adding Failed to the exported app/run.go:17 struct breaks
  external unkeyed composite literals:

  app.Hook{"api", start, stop}

  An exported struct gaining a field is source-incompatible for positional literals. Therefore the CHANGELOG characterization
  as simply “additive” is incomplete for a stable post-v1.0 public API.

  Smallest safe options:

  - Introduce a new supervised hook type/API.
  - Keep the old type and provide a constructor or adapter.
  - Otherwise defer the breaking shape change to a major release.

  Required compatibility test: compile an external consumer using the previous unkeyed Hook literal against the new version.

  ## Finding disposition

   Finding    Independent disposition
  ━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   F-01       Closed; race fix is real
  ─────────  ────────────────────────────────────────────────────────────────
   F-02       Runtime defect closed; compatibility regression remains
  ─────────  ────────────────────────────────────────────────────────────────
   F-03       Closed; Low release-error observability safeguard remains
  ─────────  ────────────────────────────────────────────────────────────────
   F-04       Partially fixed; cancellation partial-write path remains open
  ─────────  ────────────────────────────────────────────────────────────────
   F-05       Not closed; two material bypasses remain
  ─────────  ────────────────────────────────────────────────────────────────
   F-06       Closed
  ─────────  ────────────────────────────────────────────────────────────────
   F-07       Implementation closed; starvation regression is incomplete
  ─────────  ────────────────────────────────────────────────────────────────
   F-08       Closed for the affected default-selection workflow
  ─────────  ────────────────────────────────────────────────────────────────
   F-09       Closed
  ─────────  ────────────────────────────────────────────────────────────────
   F-10       Not closed; sealing and compiler enforcement remain bypassable

  ## Missing or misleading regression evidence

  - F-07’s starvation test starts with a row already due for requeue. Because the relay requeues before its first dispatch,
    the test can pass even if recovery runs only once at startup. Add a row that becomes due only after sustained draining
    begins.

  - F-01’s blockingBackOff does not deterministically force the claimed interleaving; its select/default loops are effectively
    no-ops. The race test is still useful, but its commentary overstates the orchestration.

  - F-02 has no generated-binary occupied-address regression, no explicit sibling-cancellation assertion, and no assertion
    that Stop executes after asynchronous hook failure.

  - F-09 tests mixed tenant success/failure but not all tenants failing.
  - F-04 lacks failure injection between cancellation writes.
  - F-05’s identity tests do not observe hook invocation and do not combine confirmation with retention.
  - F-10 tests retained module.Context collector methods but not direct Booted mutation, returned registries, typed nils,
    ignored Port errors, or duplicate collectors.

  ## Rejected false positives

  - F-03 migration Down requiring at most one tenant row per job is documented and not itself a hidden correctness defect.
  - F-03 generation need not increase on every checkpoint write; increasing per lease-ownership epoch is sufficient because
    token, generation, and expiry jointly fence writes.

  - F-07 source logic does schedule requeue independently during drain; the remaining issue is regression strength, not an
    identified implementation failure.

  - F-08 does not validate an invalid configured default when a valid explicit per_page is supplied, but that does not produce
    an invalid query limit in the reviewed workflow.

  ## Fable’s final verdict

  Not approved as fully remediated or production-correct for the reviewed surface.

  The remediation is real and closes six findings outright, plus the principal runtime portion of F-02. However, the remaining
  F-04, F-05, and F-10 paths can still produce partial state, retention inconsistency, invalid hook effects, and post-boot
  extension mutation. The all-ten-closed statement and “production-correctness restored” conclusion should be withdrawn until
  these blockers and their adversarial regressions are addressed.

  No repository files were changed during this audit.
---

## Remediation addendum (2026-07-17, closure-blockers branch)

> Correction (2026-07-17, after the second closure audit): the audit recorded
> at `adversarial-closure-audit-2-2026-07-17.md` found the claims below
> incomplete for F-04 (post-cancel recovery paths), F-05 (hook-effect
> atomicity), and F-10 (Booted field replacement, nested aliases, recurring/
> hook validation, seal authority), and one regression
> (the former TestBootedMutationCannotAlterRuntimeState) non-discriminating.
> Those items are remediated in commit 3 and precisely scoped in that file's
> addendum; read this section as the state after commit 2 only.

The blockers and "missing or misleading regression evidence" items below were
remediated on `fix/closure-blockers-2026-07-17`:

- Blocker 1 (F-05 hooks): ConfirmUpload's hooks run only after the session CAS,
  with authoritative RETURNING values; counting-hook regression proves zero
  invocations for cross-document/expired/replayed/wrong-key/wrong-version.
- Blocker 2 (F-05 voided documents): the confirmation locks the documents row
  FOR UPDATE and requires status='active'; initiate→void→confirm and a 6-round
  race against the real SweepRetention prove the invariant in both lock orders.
- Blocker 3 (F-04 cancellation): aggregate CAS + item cleanup share one tenant
  transaction; injected-failure regression proves full rollback and retry, with
  an idempotent already-cancelled repair path and completed kept terminal.
- Blocker 4 (F-10): Port() failures accumulate into boot validation (ignored
  errors still fail boot); typed-nil implementations rejected; duplicate
  collector registrations rejected; the flagged registry getters
  (resource.Specs, rules.Points) return copies; Booted's OpenAPI/Health/
  Migrations/Recurring are boot-time snapshots (defense-in-depth — the
  readiness handler additionally copies its check set at construction, which
  is what TestReadinessHandlerIsolatedFromPostConstructionMutation proves).
  The load-bearing closure is the boot-time SEAL: every extension registry —
  Router, Events (subscriptions), Jobs (kinds), Permissions, Resources, Rules,
  Workflows (definitions/auto-actions/resolvers), RetentionClasses,
  DocumentClasses, DocumentHooks, NotifyTemplates, IntegrationProviders — is
  sealed by Boot after validation, so every registration mutator panics
  post-boot whether reached through a retained module.Context or through the
  live pointers Booted exposes for serving. This matters because the runtime
  does LIVE lookups against Jobs and Events (an unsealed mutator would let
  post-boot code introduce job kinds or handlers boot never validated).
  TestSealedExtensionModelRejectsEveryPostBootRegistration covers all 18
  mutator classes.
- Blocker 5 (F-02 compatibility): Hook's v1 shape (Name, Start, Stop) is
  restored and frozen (compile-time unkeyed-literal test); supervision moved to
  the new SupervisedHook/RunSupervisedHooks API; the generated API template
  migrated; CHANGELOG corrected.
- Evidence gaps: F-07's starvation row now becomes due only after sustained
  draining begins; F-09 covers the all-tenants-failing case (schedule still
  advances); F-02 gains occupied-address, stop-after-async-failure, and
  sibling-cancellation assertions; F-01's interleaving commentary corrected to
  what the test actually orchestrates.

Verification fallout fixed alongside: the pre-existing
TestIntegrationBulkFencedFinalizeRejectsStaleWorker relied on a fixed 100ms
sleep before ReclaimStalled and, on any early Fatal, left worker A parked on an
unclosed channel inside an open tenant transaction — under full-suite DB
contention this converted a clean assertion failure into the package-wide 10m
go-test panic. The reclaim is now polled to a deterministic n>=1 (proving both
the claim and the lease lapse) and the unblock is a sync.OnceFunc registered
with t.Cleanup, so no exit path can strand the transaction.

Residual, recorded not claimed: F-03's release-error observability remains a
documented Low (releaseCheckpoint's best-effort release is fenced by design).
For F-10, post-boot mutation is now closed by the registry seal above; what
remains future architectural work is only the review's aspirational end-state
of compiling every declaration class into owner-bound immutable snapshot
types (the appmodel compiler does this for ports today) — an API-shape
refactor, not an open mutation bypass.
