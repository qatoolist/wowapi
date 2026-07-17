# Adversarial Framework Review Report

Date: 2026-07-17
Reviewed revision: `5f4ffcea03d2854cfc9693995a95f72d68c51d5b`
Scope: correctness, reliability, operability, compatibility, and misuse resistance. Security assessment was explicitly excluded.

## Executive verdict

**Not approved as production-correct or sufficiently reliable in the reviewed state.**

The review confirmed five High-severity implementation defects, one High-severity architectural misuse weakness, and four Medium-severity defects or missing safeguards. The highest-risk failures can:

- crash a worker process through concurrent use of a supposedly stateless retry schedule;
- leave a worker or generated API process alive after its critical child loop or listener has died;
- duplicate, lose, regress, or repeatedly restart resumable migration work because checkpoint identity and lease fencing are ineffective;
- mark a bulk operation complete while live work exists, then permanently strand a retryable item;
- attach a valid upload session's content to the wrong document or confirm an expired session;
- allow derived modules to overwrite or consume extension ports without the ownership, type, or dependency checks described by the public API.

The full Go suite passed. That result is materially incomplete evidence: the suite does not cover concurrent singleton use, child failure before parent cancellation, empty claims while peer workers own live items, cross-document session substitution, generated-project missing-row behavior, or the mismatch between the typed application compiler and the runtime boot path.

Release or premier-framework claims should be gated on fixes and adversarial regression tests for findings F-01 through F-05 and F-10, plus the generated-project supervision and CRUD cases in F-02 and F-06.

## Review coverage and method

The refreshed Graphify map was used only to route inspection. No graph edge or inferred relationship was accepted as defect evidence without direct source inspection or an executable reproduction.

Covered areas:

- production code under `app`, `kernel`, `foundation`, and public `module` APIs;
- database queries, transaction boundaries, migrations, checkpointing, leases, bulk state, outbox recovery, and document versioning;
- runtime supervision, cancellation, shutdown, retries, recurring jobs, diagnostics, and error propagation;
- pagination contracts and zero-value behavior;
- generated API, worker, and CRUD templates, including derived-project impact;
- repository-wide tests, generated-consumer tests, compatibility tests, and relevant documented contracts.

Evidence commands and results:

- `go test ./...` — passed across all packages, including `internal/cli`, `internal/compat`, `internal/e2e`, migration, app, outbox, retry, pagination, bulk, and document packages.
- Concurrent retry reproduction via `go run -race /tmp/wowapi_retry_race_repro.go` — failed with multiple data-race reports and an actual `panic: runtime error: index out of range [3] with length 3` at `kernel/retry/retry.go:65`.
- Worker supervision reproduction via `go run /tmp/wowapi_worker_error_repro.go` with a non-nil pool whose child loops immediately failed — remained blocked for two seconds and printed `HUNG after child loops failed`; `StartWorker` returned the stored relay error only after the parent context was cancelled.
- Pagination boundary reproduction via `go run /tmp/wowapi_pagination_repro.go` — `Defaults{}` produced `Limit=0, err=nil`; `Defaults{PerPage:-5}` produced `Limit=-5, err=nil`.

The two `/tmp` programs were isolated review harnesses and did not alter repository production or test code.

## Finding summary

| ID | Severity | Classification | Finding |
|---|---|---|---|
| F-01 | High | Confirmed Defect | Shared retry schedules race and can panic |
| F-02 | High | Confirmed Defect | Critical child failures are not supervised |
| F-03 | High | Confirmed Defect | Backfill checkpoint identity and fencing are broken |
| F-04 | High | Confirmed Defect | Bulk completion and lifecycle transitions violate aggregate state |
| F-05 | High | Confirmed Defect | Upload confirmation is not bound to document or expiry |
| F-06 | Medium | Confirmed Defect | Generated CRUD returns incorrect missing-resource outcomes |
| F-07 | Medium | Confirmed Defect / Reproducible Risk | Outbox retry recovery is hidden and can starve |
| F-08 | Medium | Missing Safeguard | Pagination accepts invalid default configuration |
| F-09 | Medium | Confirmed Defect | Recurring-job failures are reported as success |
| F-10 | High | Architectural Weakness | Runtime extensions bypass the ownership-bound compiler |

No Critical issue was established. The retry panic is High rather than Critical because process termination is demonstrated, but fleet-wide failure or irreversible data loss was not.

## Confirmed findings

### F-01 — Shared retry schedules race and can panic

**Severity/classification:** High — Confirmed Defect
**Files and symbols:**

- `kernel/retry/retry.go:13-38`, `Schedule` and `Schedule.Next`
- `kernel/retry/retry.go:41-67`, `SequenceBackOff`
- `foundation/notify/service.go:38-47,694`, global `notifyBackoff`
- `foundation/webhook/service.go:26-32,201,726`, global `webhookBackoff`

**Violated invariant:** An attempt-number-to-duration function described as stateless must return the correct duration and be safe when delivery workers call it concurrently.

**Trigger and evidence:** Two goroutines call `Next` on the same schedule. `Next` resets and advances the same mutable `BackOff`; `SequenceBackOff.Reset` and `NextBackOff` mutate `idx` without synchronization. The isolated eight-goroutine race reproduction reported concurrent reads/writes at lines 30, 33, 55, and 66, then panicked when an interleaving moved `idx` beyond the slice.

This is a behavioral defect, not only race-detector hygiene. One goroutine can consume another's schedule positions, returning a retry that is too early or too late, and the observed panic can terminate the process.

**Affected workflows:** Notification finalization, inbound and outbound webhook retries, any derived project sharing a `retry.Schedule`, worker availability, retry exhaustion, and recovery latency.

**Root cause:** Mutable singleton state is exposed through an API whose comment promises stateless attempt lookup.

**Smallest safe fix:** Protect the complete reset-and-iterate operation with a mutex. A preferable longer-term design is an immutable indexed schedule or a factory that creates independent `BackOff` state per call. Reject a nil `BackOff` while changing the constructor.

**Required regression:** Concurrently request every supported attempt from many goroutines, assert exact durations, and run under `go test -race`. Include an orchestrated blocking fake to force reset/advance interleavings and assert no panic.

**Why tests missed it:** `kernel/retry/retry_test.go` and notify/webhook schedule tests are sequential.

### F-02 — Critical child failures are not supervised

**Severity/classification:** High — Confirmed Defect
**Files and symbols:**

- `app/worker.go:120-146`, `StartWorker`
- `kernel/outbox/relay.go:356-379`, `Relay.Run`
- `kernel/jobs/runner.go:587-619`, `Runner.Run`
- `kernel/jobs/scheduler.go:57-82`, `Scheduler.Run`
- `internal/cli/templates/init/cmd_api_main.go.tmpl:330-344`, generated HTTP hook
- `app/run.go:43-78`, `RunHooks`

**Violated invariant:** A long-lived process must remain alive only while its critical serving and processing loops are alive. An unexpected child failure must cancel siblings and make the process return non-zero promptly.

**Worker trigger and reproduction:** A database outage makes relay, runner, or scheduler return a non-cancellation error. `StartWorker` stores each result and closes `drained`, but line 130 waits only on `<-ctx.Done()`. The isolated broken-pool reproduction observed that the children failed, waited two seconds without a return, and returned the relay error only after explicit cancellation.

**Generated API trigger and proof:** An occupied or unavailable address makes `ListenAndServe` fail. The generated hook starts it in an unobserved goroutine, logs the error, and returns nil. `RunHooks` then waits on the parent context. The process can remain alive while serving nothing.

**Affected workflows and derived-project impact:** All generated workers, all custom consumers using `StartWorker`, generated API binaries, outbox delivery, jobs, scheduled maintenance, and orchestration restart behavior. The generated worker metrics listener uses the same unobserved pattern, although metrics-listener failure alone may be degradable rather than process-fatal.

**Root cause:** Spawned critical goroutines have no owned error channel or common fail-fast supervisor.

**Smallest safe fix:** Run critical children under an error group or equivalent result channel and child context. On the first unexpected return, cancel siblings, drain within `ShutdownDrain`, and return joined component errors. Bind the HTTP listener synchronously so bind failure is a `Start` error; also supervise unexpected post-start `Serve` termination.

**Required regression:**

- Start a worker over a controllably failing loop or broken pool with a live parent context; assert prompt error return, sibling cancellation, and bounded drain.
- Occupy the generated API address; assert `run` returns a bind error without an external signal.
- Start a fake service successfully and make it return a sentinel error later; assert process-level propagation.

**Why tests missed it:** Existing app tests cover synchronous hook start errors, normal cancellation, and drain timeout. They do not cover a child that dies before cancellation or a listener that fails asynchronously.

### F-03 — Backfill checkpoint identity and lease fencing are broken

**Severity/classification:** High — Confirmed Defect
**Files and symbols:**

- `migrations/00042_backfill_checkpoint_lease_columns.sql:21-34`
- `kernel/migration/backfill.go:85-87,93-140`, `Backfill.Run`
- `kernel/migration/backfill.go:188-228`, checkpoint selection
- `kernel/migration/backfill.go:231-272`, claim and checkpoint writes

**Violated invariants:**

1. A tenant-scoped job has an independent checkpoint for each tenant.
2. Only the current, unexpired lease owner may advance a checkpoint.
3. Lease generation must increase across ownership epochs so stale writers are rejected.

**Trigger and source proof:**

- The table uses `job_id` alone as the primary key while runtime reads optionally filter on `(job_id, tenant_id)`.
- A second tenant using the same stable `JobID` cannot create an independent row. Conflict handling does not update or predicate on `tenant_id`, so its reads repeatedly miss while writes collide with the first tenant's checkpoint.
- `claimCheckpoint` unconditionally replaces an existing live lease and resets the new in-memory generation to the primitive's initial value.
- `writeCheckpoint` and `writeCheckpointTx` update on `job_id` without a token, generation, expiry, or affected-row predicate.

Two runners can therefore read the same `last_key`, both claim the row, both process the same batch, and a stale runner can overwrite a newer checkpoint. Two tenants can corrupt or repeatedly restart each other's progress.

**Affected modules and data:** Migration/backfill helpers, DATA-09 style online migrations, all derived projects using stable job IDs across tenants, checkpoint progress, duplicate side effects, and rollout/recovery duration.

**Root cause:** The schema identity does not match the runtime identity, and the lease is recorded as metadata rather than used as a compare-and-swap fence.

**Smallest safe fix:** Introduce a deliberate normalized identity for global and tenant jobs, such as a composite key with an explicit global sentinel or appropriate partial unique indexes. Claim only an absent or expired lease, increment stored generation, return the acquired epoch, and fence every checkpoint update on job identity, token, generation, and unexpired ownership. Require exactly one affected row.

**Required regression:**

- Run the same `JobID` sequentially for two tenants and prove independent resume.
- Race two connections at a callback barrier and prove only one owns/processes a batch.
- Reclaim an expired lease and prove the stale owner cannot write.
- Prove a checkpoint never moves backward.

**Why tests missed it:** Existing tests establish only that lease fields become non-empty/non-zero and exercise a single tenant. They do not establish mutual exclusion, stale-owner rejection, or two-tenant identity.

### F-04 — Bulk completion and lifecycle transitions violate aggregate state

**Severity/classification:** High — Confirmed Defect
**Files and symbols:**

- `foundation/bulk/bulk.go:174-245`, `Service.Process`
- `foundation/bulk/bulk.go:305-321`, `claimSQL`
- `foundation/bulk/bulk.go:247-273`, `Pause`, `Resume`, and `Cancel`
- `foundation/bulk/bulk.go:503-511`, `mark`

**Violated invariants:** An operation is complete only when no pending or running item exists; terminal states cannot be reopened; unknown operations cannot report successful transitions.

**Deterministic concurrent trigger:**

1. Worker A claims the operation's sole item and blocks in its callback.
2. Worker B sees no claimable row because A owns a live `running` item.
3. B treats the empty claim as global completion and unconditionally marks the operation `completed`.
4. A returns a retryable business error, so the item returns to `pending`.
5. Every later `Process` exits immediately because the aggregate status is already `completed`; the item is permanently stranded.

Lifecycle methods compound the defect: `Pause`, `Resume`, and `Cancel` issue unconditional state writes. They do not enforce allowed source states or inspect `RowsAffected`, so completed/cancelled operations can be relabelled and nonexistent IDs return nil.

**Affected workflows:** Concurrent bulk workers, retryable per-item failures, progress reporting, pause/resume/cancel APIs, derived batch imports/exports, and operator recovery.

**Root cause:** Aggregate state transitions are unconditional labels rather than compare-and-swap operations that encode item and source-state invariants.

**Smallest safe fix:** Use one conditional completion update with `NOT EXISTS` over all nonterminal items, including live `running` rows. An empty claim with peer-owned live work should return without completing. Encode legal source states in lifecycle updates, inspect affected rows, and distinguish not-found from invalid transition.

**Required regression:** A barrier-controlled two-worker test where A fails after B's empty claim; expired-lease variants; every legal and illegal lifecycle transition; completed-to-resume/cancel; and unknown IDs.

**Why tests missed it:** Current concurrency tests give each worker immediate successful work. Lifecycle tests cover only the happy path.

### F-05 — Upload confirmation is not bound to document or expiry

**Severity/classification:** High — Confirmed Defect
**Files and symbols:**

- `foundation/document/service.go:247-300`, caller-supplied document and storage validation
- `foundation/document/service.go:301-327`, session confirmation CAS
- `foundation/document/service.go:329-350`, version insertion and document update
- `migrations/00032_version_counters_and_upload_sessions.sql:39-87`, session schema

**Violated invariant:** A pending and unexpired upload session may confirm only the document, version, and storage object it reserved.

**Trigger and proof:** Create documents A and B, initiate a session for A, upload the object, then call `ConfirmUpload` with A's session ID, version, key, and checksum but B as `DocumentID`. The CAS predicates only on session ID, pending status, and checksum. It never checks or returns `document_id`, and it never checks `expires_at > now()`. The later insert and document counter update use caller-supplied B. If B has no conflicting version, the transaction can confirm A's session while attaching the version to B. A session whose timestamp has expired also remains confirmable until the sweeper changes its status.

**Affected data and callers:** Document-to-version association, document version counters, upload-session auditability, storage reconciliation, all modules using the document service, and derived projects exposing confirm endpoints.

**Root cause:** The CAS is scoped only to session settlement, not to the complete reserved identity or validity window.

**Smallest safe fix:** Predicate the CAS on `document_id`, `version_no`, `storage_key`, checksum, and `expires_at > now()`, or return the authoritative document identity and validate it before effects. Use only authoritative returned values for the version insert.

**Required regression:** Cross-document substitution must roll back; expired-but-unswept confirmation must conflict; wrong key/version must not settle the session; and the correct session must still confirm once.

**Evidence boundary:** Under the normal tenant transaction, a post-CAS mismatch error rolls the CAS back. The defect is missing document and time binding, not permanent settlement from a simple wrong key/version.

### F-06 — Generated CRUD returns incorrect missing-resource outcomes

**Severity/classification:** Medium — Confirmed Defect
**Files and symbols:**

- `internal/cli/templates/crud/resource.go.tmpl:133-149`, generated GET
- `internal/cli/templates/crud/resource.go.tmpl:227-283`, generated UPDATE and DELETE
- `kernel/httpx/errors.go:55-68`, unknown-error mapping

**Violated invariant:** Valid requests for absent or inactive resources must return the framework's not-found contract, not internal error or false success.

**Trigger and proof:**

- GET of a valid nonexistent UUID returns raw `pgx.ErrNoRows`. `httpx.WriteError` treats non-framework errors as opaque 500.
- UPDATE and DELETE ignore the command tag. Zero affected rows for an absent or inactive UUID still produce 200 and 204 respectively.

**Affected derived projects:** Every project using `wowapi gen crud`; client behavior, retries, caches, monitoring, and API contract compatibility.

**Root cause:** Generated happy-path handlers do not translate database absence into a domain error or verify write effects.

**Smallest safe fix:** Translate `pgx.ErrNoRows` and zero affected rows to `KindNotFound`; retain consistent inactive-row semantics.

**Required regression:** Generate an independent project and exercise GET, PUT, and DELETE for valid nonexistent and inactive UUIDs against real Postgres. Assert RFC 9457 not-found responses and no false success.

**Why tests missed it:** Generated-consumer tests compile and boot generated code but do not exercise missing-row runtime contracts.

### F-07 — Outbox retry recovery is hidden and can starve

**Severity/classification:** Medium — Confirmed Defect for swallowed errors; Reproducible Risk for sustained-traffic starvation
**Files and symbols:** `kernel/outbox/relay.go:356-397`, `Relay.Run` and `RequeueFailed`

**Violated invariant:** Retryable failed events must be returned to the dispatch queue on schedule, and recovery failure must be observable.

**Trigger and proof:** `Run` calls `RequeueFailed` only from the idle ticker branch and explicitly discards its error. Whenever `DispatchOnce` returns `n > 0`, the loop immediately continues, bypassing the ticker. A persistent update/grant/schema error therefore leaves failed events stuck with no returned error, log, or metric. Under sustained unrelated pending traffic, the idle branch may never run. A failed predecessor can block later events for the same aggregate while unrelated traffic keeps the drain loop busy.

**Affected workflows:** Outbox retry/DLQ progression, per-aggregate ordering, event-driven integrations, operators relying on process success, and derived event consumers.

**Smallest safe fix:** Run failed-event maintenance on an independent due schedule even while draining. Propagate its error or implement an explicit bounded retry with error logging and a failure metric.

**Required regression:** Inject a requeue-only database error and require observable failure. Add a failed aggregate predecessor plus continuous unrelated pending producer and prove the failed row becomes pending within a bounded interval.

**Evidence boundary:** The ignored error is source-proven. Sustained-traffic starvation is classified as Reproducible Risk until the dedicated producer test is executed.

### F-08 — Pagination accepts invalid default configuration

**Severity/classification:** Medium — Missing Safeguard
**File and symbols:** `kernel/pagination/pagination.go:43-105`, `Defaults`, `Request`, and `Parse`

**Violated invariant:** The documented parsed request limit is positive and clamped to `[1, MaxPerPage]` when an upper bound exists.

**Trigger and reproduction:** Omit `per_page` or pass `0` while using a zero or negative `Defaults.PerPage`. `Parse` returns that value unchanged. The isolated reproduction returned `Limit=0` and `Limit=-5` without error.

**Affected callers:** Public extension consumers constructing `Defaults`, custom list handlers, SQL `LIMIT` behavior, response consistency, and amplification risk when a caller interprets zero as unlimited. Generated CRUD currently supplies valid values, so this is mainly a misuse-resistance defect.

**Root cause:** Raw request values are validated, but configuration defaults bypass the same positive-range invariant.

**Smallest safe fix:** Validate defaults and return a configuration/validation error, or apply a documented positive framework default. Do not silently pass negative limits.

**Required regression:** Zero, negative, one, max, max-plus-one, empty, and explicit-zero table tests, including invalid combinations such as `PerPage > MaxPerPage`.

### F-09 — Recurring-job failures are reported as success

**Severity/classification:** Medium — Confirmed Defect
**Files and symbols:**

- `app/maintenance.go:161-183`, `registerModuleRecurring`
- `kernel/jobs/scheduler.go:90-107`, `Scheduler.Tick`
- `app/worker.go:101-112`, scheduler observer and metrics

**Violated invariant:** If one or more tenant executions fail, the task observer and error metric must not report a successful run.

**Trigger and proof:** A recurring callback fails for one or every active tenant. The fan-out logs each error but always returns nil. `Scheduler.Tick` passes nil to `onRun`; `StartWorker` records `ok=true` and does not increment `scheduler_task_errors_total`. The schedule is advanced even when every tenant failed.

**Affected workflows:** Module recurring jobs, operator alerts, dashboards, retry expectations, tenant-specific maintenance, and derived project callbacks.

**Root cause:** Partial-failure isolation was implemented by discarding the aggregate error rather than separating continued fan-out from final outcome reporting.

**Smallest safe fix:** Continue processing all tenants, collect failures with tenant context, and return a joined error after fan-out so the observer and metrics reflect non-success. Define whether failed tenants retry immediately or at the next interval.

**Required regression:** Mixed success/failure and all-fail tenant sets; assert every tenant is attempted, observer receives non-nil error, error metric increments, and diagnostics identify failed tenants.

### F-10 — Runtime extensions bypass the ownership-bound compiler

**Severity/classification:** High — Architectural Weakness
**Files and symbols:**

- `module/module.go:66-125`, public `Context` extension contract
- `app/context.go:322-351`, direct maps and `ProvidePort`/`Port`
- `app/boot.go:87-150`, actual module registration path
- `kernel/appmodel/appmodel.go` and `kernel/port`, typed compiler/port implementation not integrated into production `App.Boot`

**Violated invariant:** Public extension points documented as boot-checked must enforce owner, dependency, uniqueness, type, and post-boot immutability constraints.

**Trigger and proof:** A module calls `ProvidePort` with another module's prefix, provides the same name twice, provides a nil or wrong-typed implementation, or resolves a port from a nondependency. The runtime path performs an unconditional map assignment and lookup. No owner, type, declared-dependency, duplicate, or sealed-state check exists. `Migrations`, `Seeds`, `OpenAPI`, and health/recurring collectors similarly expose direct mutable registration behavior with uneven duplicate validation.

The repository contains a richer `kernel/appmodel` compiler and typed `kernel/port` facilities, but production `App.Boot` does not use them. Their tests therefore do not prove runtime module safety. Existing architecture documents disclose this gap, so it is an architectural weakness rather than a hidden implementation regression.

**Affected derived projects:** Every third-party/product module, module ordering and coupling, extension replacement, boot determinism, runtime type assertions, and compatibility across framework upgrades.

**Root cause:** Two extension models coexist: a validated typed compiler and the older mutable runtime registration maps. Only the latter drives boot.

**Smallest safe fix:** This is not a safe one-line patch. Stage the ownership-bound compiler into `App.Boot`, preserve compatibility with an adapter, and make registration immutable after boot. Reject duplicates, wrong owners/types, undeclared dependency access, missing providers, and retained-context mutation.

**Required regression:** Adversarial modules for duplicate and foreign-prefix providers, type mismatch, dependency bypass, missing port, cyclic requirements, retained context after boot, and deterministic boot errors. Generated consumers must run through the same compiler path.

## Missing safeguards and misleading tests

Cross-cutting missing safeguards:

- No common supervisor owns critical goroutines and their errors.
- Several database state transitions are unconditional writes instead of invariant-bearing compare-and-swap operations.
- Retry schedule APIs expose mutable state behind stateless-looking methods.
- Recovery and fan-out paths prefer continued execution but lose the aggregate failure signal.
- Generated-project tests emphasize build/boot success and omit missing-row and occupied-listener behavior.
- Runtime module boot does not use the typed ownership compiler whose tests give an appearance of stronger enforcement.

Misleading or incomplete evidence:

- `go test ./...` is green but has no concurrent `Schedule.Next` test.
- Worker tests cover clean cancellation and drain limits, not pre-cancellation child failure.
- Bulk concurrency tests do not create an empty claim while a peer owns a live running item.
- Backfill tests check that lease fields are populated, not that they exclude a second owner or fence a stale writer.
- Document CAS tests use the correct document and test expiry only through the sweeper.
- Generated CRUD tests compile and boot output but do not execute absence semantics.
- Outbox tests call `RequeueFailed` directly or become idle; they do not test ignored recovery errors or sustained traffic.
- Recurring-job tests verify execution, not truthful failure metrics.
- `kernel/appmodel` tests validate a compiler that the actual boot path does not invoke.

## Recommended remediation order

1. Fix F-01 immediately: it is a demonstrated process-crash path with a small, contained remediation.
2. Introduce a shared lifecycle supervisor and use it for worker loops and generated HTTP serving (F-02).
3. Correct checkpoint schema identity and implement real fencing before any further online backfill rollout (F-03).
4. Repair bulk aggregate completion and lifecycle compare-and-swap transitions together (F-04).
5. Bind upload confirmation to authoritative session identity and expiry (F-05).
6. Correct generated CRUD outcomes and add real generated-project API contract tests (F-06).
7. Separate outbox recovery timing from drain activity and make recovery failures observable (F-07).
8. Enforce pagination defaults and truthful recurring-job failure aggregation (F-08, F-09).
9. Plan and execute the compiler-to-runtime integration as an explicit compatibility project (F-10).

Each fix should land with its adversarial test first or in the same change. Green happy-path tests are not sufficient closure evidence.

## Rejected false positives

- Graph degree for `NewDB` and `CreateTenant` is dominated by testkit call density; it is not proof of production coupling.
- `kernel/appmodel.Compiler` graph centrality does not prove runtime enforcement. The disconnection is F-10; compiler tests must not be cited as current boot protection.
- A shared jobs batch lease token is not independently defective because finalization also predicates on job ID and generation.
- Outbox external handler effects are documented as at-least-once. No duplicate-external-effect defect was reported without a stronger contract.
- Outbox inbox insertion is transactional with handler database effects, so handler failure rolls back both the effect and inbox claim.
- `ConfirmUpload` key/version mismatch after CAS normally rolls back under the tenant transaction. The confirmed defect is missing document and expiry binding.
- `Runner.ClaimOnce` uses `context.WithoutCancel`, but `execOne` applies a per-job timeout and the outer worker has a documented drain cap. That is distinct from F-02.
- Pagination trailing JSON is explicitly rejected in the complete decoder path.
- `SequenceBackOff` does clamp above its final entry in sequential use. Its defect is shared concurrent mutation.
- The generated-project independence claim is not wholly untested: isolated versioned consumers build and boot. The gaps are runtime listener/supervisor and missing-row behavior.
- The migration 00036 rollback sequence retains a foreign-key backstop through reverse ordering; it was not reported.
- The statutory sequence allocator and allocation ledger share the caller's transaction, so rollback restores the number as documented.

## Downgraded observations and residual risks

- `kernel/config/load.go:148-165` resolves secret-provider calls with `context.Background`, and the public options accept no caller context. A stalled network/KMS provider cannot observe process cancellation. This is a Low-to-Medium architectural weakness pending a provider-hang reproduction. Add a context-taking loader or context option and a finite startup budget.
- Generated API/worker tracing and worker metrics cleanup use `context.WithoutCancel` and discard shutdown errors. API HTTP shutdown itself is bounded by `RunHooks`, so this was not generalized to all cleanup. Worker metrics and OTLP cleanup can still exceed termination grace; add fresh cleanup deadlines and propagate or log errors.
- Generated worker metrics listener failure follows the unobserved goroutine pattern, but metrics availability was not established as process-critical. Treat it as an affected F-02 surface with a policy decision rather than a standalone High defect.
- Constructor APIs such as relay/runner accept dependencies whose nil misuse may panic later. This was not exhaustively audited or promoted where composition already guarantees non-nil values.

## Unreviewed areas

- Full migration up/down and N-1/N compatibility over large, populated schemas with lock contention and interrupted execution.
- Exhaustive production-scale query-plan and index validation.
- Storage-provider outage, partial object-write reconciliation, and orphan cleanup for every adapter.
- Database pool exhaustion, connection churn, cancellation, and failover behavior.
- All webhook/notification provider partial-failure combinations beyond the retry-schedule defect.
- Cache invalidation across multiple process replicas.
- Audit-chain anchoring and historical pagination under concurrent writes.
- Workflow transition and sweeper SQL beyond graph-guided spot checks.
- Rules history resolution across every temporal overlap and activation race.
- Every generated deployment/container workflow and every supported PostgreSQL version.
- Exhaustive OpenAPI compatibility and middleware-order permutations.

These are residual risks, not claims of correctness.

## Fable final quality-gate verdict

Fable independently challenged the candidates, merged shared root causes, rejected speculative graph conclusions, and confirmed the final severities.

**Final verdict: not approved as production-correct or sufficiently reliable.** Multiple independent High defects can crash a process, leave it falsely alive with critical work stopped, corrupt or strand resumable work, or associate uploaded content with the wrong document. The green suite is materially misleading because it omits the exact adversarial triggers that violate those invariants.

The final gate requires remediation and regression evidence for F-01 through F-05 and F-10, plus the generated supervision and missing-resource contracts, before a production-readiness claim should be reconsidered.

---

## Remediation addendum (2026-07-17, post-review)

All ten findings were remediated on branch `fix/adversarial-remediation-2026-07-17`
with adversarial regression tests that failed before each fix:

- F-01/F-08/F-09 + F-02/F-03/F-06/F-07: commit `6770a4b`.
- F-04/F-05: commit `9ae702b`.
- F-10: commit `3d750a3` — compiler-to-runtime integration (ownership,
  duplicate, nil, declared-dependency, and post-boot-seal enforcement wired
  into `App.Boot` via `kernel/appmodel`).
- Independent-gate fixes: commit `313b6ac` — migration 00049 ledger
  registration, skip-manifest entry SKIP-040, extension seal extended from
  ports to every boot collector (Migrations/Seeds/OpenAPI/Health/
  RecurringJob/I18n), harness module-cache self-invalidation, CHANGELOG
  compatibility notes.

Closure evidence, per-finding commands, and the final verdict live in the
Remediation and Closure Report delivered with the branch. This report's
"not approved" verdict describes revision `5f4ffce`, superseded by the
remediated branch.

**Correction (2026-07-17, post-closure-review):** the follow-up adversarial
closure review (`docs/working/adversarial-closure-review-2026-07-17.md`)
REJECTED the first remediation's "all ten closed / production-correctness
restored" claim: F-04 (non-atomic cancellation), F-05 (hooks before the CAS;
voided-document confirmation), and F-10 (ignored port errors, typed nils,
duplicate collectors, mutable Booted/registry exposure) retained material
bypasses, and F-02's `Hook.Failed` field was a source-incompatible API change.
Those blockers were remediated in the closure-blockers branch that carries
this correction; the closure-review record documents the final disposition.

Harness observation recorded during remediation: the e2e/golden module-proxy
fixtures reuse constant synthetic framework versions; the shared GOMODCACHE and
build cache can serve stale zips/objects that mask working-tree changes
(manifesting as bogus "could not import" errors). Purge the synthetic versions
or content-hash them — tracked as follow-up.
