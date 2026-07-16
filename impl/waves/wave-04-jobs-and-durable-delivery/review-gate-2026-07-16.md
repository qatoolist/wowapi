---
id: REVIEW-GATE-W04-2026-07-16
type: review-gate-record
parent_wave: W04
status: complete
created_at: 2026-07-16
updated_at: 2026-07-16
---

# Wave 04 — Independent review gate (2026-07-16, post-remediation)

## Reviewer identity

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

## Scope

Wave 04 (jobs and durable delivery) problem stories, post-remediation, per
`/private/tmp/claude-502/-Users-qatoolist-go-home-src-github-com-qatoolist-wowapi/97aeaae9-840e-4c51-bf72-b17540116e23/scratchpad/autopsy/verification/wave-04-jobs-and-durable-delivery.json`
(the autopsy's adversarial per-story verdicts) and
`.../scratchpad/autopsy/evidence/quality-rerun-postfix.log` (post-remediation full gate run: lint 0
issues, tests pass, coverage 84.5/84.0). Commands were re-run, not blindly trusted, for every story
in scope; only the decisive command(s) per story were re-executed (targeted `go test -count=1`
against the live local Postgres instance).

Commit basis: `HEAD 43b6e12` + remediation working tree changes present 2026-07-16 (notably
`foundation/webhook/service.go`, `foundation/webhook/tx_boundary_test.go`,
`foundation/webhook/tamper_matrix_test.go`, `kernel/auth/auth.go`/`auth_test.go`,
`kernel/database/isolation_test.go`, and the tracking/status-register/closure-report corrections),
all currently uncommitted.

## Per-story results

### W04-E01-S001 — Shared lease/fencing primitive
**Recommendation: accept-with-conditions.** `kernel/lease` builds and its unit tests pass
(`go test ./kernel/lease/... -count=1` → `ok`, 0.479s). Both `foundation/webhook` and
`foundation/notify` genuinely reuse the shared primitive (not a parallel bespoke implementation).
**Condition:** `closure.md`'s "Final status" section is still unfilled governance-template text
despite `story.md` claiming `status: accepted` — must be filled before this story is treated as
formally closed. Record: `epics/epic-001-lease-fencing-primitive-and-jobs/stories/story-001-shared-primitive/tasks/task-003-independent-review.md`.

### W04-E01-S002 — Jobs lease columns, fenced finalize, and fenced reclaim
**Recommendation: accept-with-conditions.** Migration `00038_jobs_lease_columns.sql` (31 lines)
confirmed present. AC-02 (fenced finalize rejects a stale worker) is proven end-to-end by the sibling
story's chaos test (`kernel/jobs/chaos/duplicate_worker_lease_expiry_test.go`,
`TestDuplicateWorkerLeaseExpiry`), re-run and PASSING with "stale finalize rejected" and effect
count == 1. **Condition:** same closure.md paperwork gap as W04-E01-S001.
Record: `.../story-002-jobs-lease-and-finalize/tasks/task-004-independent-review.md`.

### W04-E01-S003 — Worker idempotency contract and shared duplicate-worker chaos harness
**Recommendation: accept-with-conditions.** Re-ran `TestDuplicateWorkerLeaseExpiry` — PASS. This is
the shared harness underlying W04-E01-S002 and W04-E03-S002's own chaos tests, genuinely reused
(confirmed by direct import), not duplicated. **Condition:** same closure.md paperwork gap.
Record: `.../story-003-idempotency-and-chaos/tasks/task-005-independent-review.md`.

### W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol
**Recommendation: accept.** This was the C-1 defect: the webhook outbound leg previously ran
`secrets.Resolve` and the real HTTP `POST` inside an open `plat.WithTenant` database transaction,
directly contradicting AC-02/AC-03. The 2026-07-16 remediation genuinely restructures
`foundation/webhook/service.go` into claim (tx, assigns a `kernel/lease` lease) / effect (no tx,
`secrets.Resolve` + `sender.Post`) / finalize (short tx, lease-fenced `WHERE lease_token = $n AND
lease_generation = $n AND lease_expires_at > $n`) stages, mirroring `foundation/notify`'s existing
staging. The secret-resolution short-circuit is preserved: a `secretErr` releases the lease and
leaves the row untouched, matching pre-staging behavior — no POST is attempted when the secret
cannot be resolved. Verified by re-running the new regression suite:
```
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  go test ./foundation/webhook/... -run \
  'TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO|TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO|TestIntegrationHandleInbound_TamperedKeyID|TestIntegrationHandleInbound_TamperedSignatureVersion' \
  -count=1 -v
```
All 4 PASS. Full-package retest (`./foundation/webhook/... ./foundation/notify/...`) — both `ok`, no
regressions. Minor observation (non-blocking): `leaseTTL` (5m) is a fixed constant sized to cover
`OutboundTimeout` + the finalize round-trip but not derived from `OutboundTimeout` — a future
`OutboundTimeout` increase could silently shrink the fencing margin; recommend a follow-up, not a
blocker. Record: `.../story-001-notify-and-webhook-three-stage/tasks/task-004-independent-review.md`;
evidence updated in the story's `evidence/index.md` (EV-002/EV-003 now `retested`/`resolved`,
superseding the pre-remediation `not yet produced`/`implemented-incorrectly` state).

### W04-E02-S002 — Inbound two-phase verification, adapter contracts, and 6-boundary chaos test
**Recommendation: not-ready.** Confirmed the 2026-07-16 status-honesty revert (autopsy R-1) is
accurate: `story.md` now reads `status: planned` (was falsely `accepted`), `closure.md` carries an
explicit correction note and its body is unchanged ("has not been implemented, verified, or
closed"), and `impl/tracking/status-register.md` correctly lists it `planned`. Nothing in the repo
partially claims completion. Re-confirmed the underlying gap is real, not just a status label: no
chaos test directory exists for notify or webhook anywhere in the repo (`find . -type d -iname
chaos` → only `kernel/jobs/chaos`, `foundation/bulk/chaos`), and `foundation/webhook/service.go`'s
`HandleInbound` still runs signature verification and secret resolution entirely inside the caller's
single open tenant transaction, by design and per its own doc comment — the opposite of the claimed
two-phase (snapshot / verify outside tx / re-check) protocol. This is genuinely future work; the
`planned` status should be preserved. Record: `.../story-002-inbound-two-phase-and-contracts/tasks/task-006-independent-review.md`.

### W04-E02-S003 — Adopt cenkalti/backoff/v5
**Recommendation: accept.** `cenkalti/backoff/v5 v5.0.3` confirmed in `go.mod`; `kernel/retry/retry.go`
wraps it; both `foundation/notify` and `foundation/webhook` reference `kernel/retry` directly (no
hand-rolled duplicate). `closure.md` is self-consistently filled in. Record:
`.../story-003-retry-adoption/tasks/task-003-lightweight-review.md`.

### W04-E03-S001 — Bulk multi-worker stopgap (false-safety-claim correction)
**Recommendation: accept.** `migrations/00016*.sql`'s header comment now correctly reads "single
processor per operation; multi-worker fan-out is added later," replacing the previously false "safe
across replicas" claim. `closure.md` is filled in as `accepted`, consistent with `story.md`.

### W04-E03-S002 — Leased claims, finalize fencing, lifecycle controls, and multi-worker chaos test
**Recommendation: accept.** Re-ran the named chaos test:
```
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  go test ./foundation/bulk/chaos/... -run TestIntegrationBulkDuplicateWorkerChaos -count=1 -v
```
PASS — 4 concurrent workers, random pause/resume/cancel, effect-ledger dedup confirms ledger
successes == done items with no sequence processed more than once. This task's own review was
self-performed by the implementing agent (explicitly disclosed in its own Deviations Record, not a
concealment) rather than independently reviewed at implementation time; this gate now supplies the
genuine independent confirmation (see the task file's new "Independent re-confirmation" section).
Minor, non-blocking: evidence docs cite the pseudo-path `DATA-04/chaos/duplicate_worker_test.go`;
the real path is `foundation/bulk/chaos/duplicate_worker_test.go` — cosmetic only.
Record: `.../story-002-leased-claims-and-lifecycle/tasks/task-006-independent-review.md`.

### W04-E04-S001 — Audit hash-chain widening with hash_version discriminator
**Recommendation: accept**, with a status-vocabulary normalization note. Re-ran
`TestIntegrationAuditChainDetectsPerFieldTampering` — PASS, 10 subtests, one per persisted field,
each independently confirming the widened hash catches tampering of that field; `hash_version`
branching fails closed on an unrecognized version. `closure.md`'s status token
`closed-pending-review` is not a value in this programme's documented status vocabulary
(`planned`/`ready`/`in-progress`/`accepted`/etc.) — recommend the conductor normalize it.
Record: `.../story-001-audit-hash-widening/tasks/task-002-independent-review.md`.

### W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, explicit per-class status
**Recommendation: accept**, with the same status-vocabulary normalization note. Performed the full
pending review (task T005 was `todo`). All four ACs verified with genuinely discriminating tests, not
duplicative or self-masking ones:
- AC-01: `TestIntegrationExternalAnchorTamperDetection` explicitly asserts the pre-existing local
  `Verify` guard *still passes* after the tamper (chain is internally self-consistent post-tamper)
  while the external-anchor `Verify` *fails* — proves the anchor catches what the local guard misses.
- AC-02: `TestIntegrationDSRExportArtifactWriteFailure` injects a failing artifact writer and asserts
  the DSR request stays `pending` (export completion genuinely gated on write success); the checksum
  (`SHA256(ciphertext)`) verifies the bytes actually written, per `deviations.md`.
- AC-03: `TestIntegrationCentralLegalHoldBlocksDisposeErase` uses a callback with *no internal hold
  check of its own* (would flip a `deleted` flag if the wrapper failed) and confirms the central
  wrapper blocks it — a genuine negative test, not one a redundant internal check could mask.
- AC-04: the `RecordClass` enumeration in `deviations.md` is honestly recorded as "zero classes
  registered in either wowapi or wowsociety today," predating the wrapper implementation, and
  explicit per-class status tests cover both callback-bearing and callback-absent classes.
All cited tests re-run and PASS. Record: `.../story-002-anchor-dsr-hold/tasks/task-005-independent-review.md`.

### W04-E04-S003 — Readiness and configuration diagnostics truthfulness
**Recommendation: accept.** This reverses the autopsy's prior `unsupported-by-evidence` verdict
(a time-budget limitation of that pass, not a defect finding) — the implementation is real. Located
and re-ran all three decisive test sets:
```
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  go test ./app/... -run 'TestIntegrationMigrationCurrencyCheckPassesWhenCurrent|TestIntegrationMigrationCurrencyCheckFailsWhenStale|TestIntegrationReadinessReportsSeedAndRuleHashes' -count=1 -v
go test ./internal/cli/... -run 'TestConfigDoctorDiscoversProductRootFromNestedSubdir|TestConfigDoctorDiscoversProductRootFromOutsideRepo|TestConfigDoctorReportsSkippedProductValidation' -count=1 -v
```
All PASS. `app/health.go`'s `MigrationCurrencyCheck` genuinely fails readiness (503) on a stale
migration version; `ReadinessWithCatalogs` unconditionally reports `migration_version`,
`seed_catalog_hash`, `rule_hash`, with `model_hash` honestly recorded as pending AR-01 in
`deviations.md` (`DEV-W04-E04-S003-001`) rather than silently claimed. `internal/cli/config_delegate.go`'s
`resolveProductRoot` genuinely shells out to `go env GOMOD` (not a CWD-relative fallback) with
`--project` as an explicit override, and reports whether product validation ran in both cases. No
DX-07 T4 (capacity/backpressure) scope drift found. **Fixed a real evidence-policy violation while
reviewing:** the story's `evidence/index.md` cited commit `HEAD` — a moving target, not a pinned SHA,
which evidence-policy.md explicitly forbids ("Evidence that does not identify the tested revision
must not be treated as final proof"; "never 'current HEAD'"). Updated to the pinned
`HEAD 43b6e12 + remediation working tree 2026-07-16` and corrected one execution-command name
mismatch (`TestConfigDoctorDiscoversProductRoot` did not match any real test; the actual names are
the two `...FromNestedSubdir`/`...FromOutsideRepo` funcs). Record:
`.../story-003-readiness-truthfulness/tasks/task-004-independent-review.md`; evidence corrected in
the story's `evidence/index.md`.

## Wave-level findings

1. **Systemic paperwork gap, partially remediated.** Five stories previously had `closure.md`
   "Final status" sections left as unfilled governance-template text despite `story.md` claiming
   `status: accepted`. This review found that **three of the five remain unfixed**
   (W04-E01-S001, W04-E01-S002, W04-E01-S003) — flagged as accept-with-conditions above. Two were
   resolved by this review's action: W04-E02-S001 now has real evidence recorded (was blocked on the
   C-1 code defect, now fixed); W04-E02-S002's closure.md was already honestly corrected by the prior
   remediation pass (status reverted to `planned`).
2. **Wave-level status inconsistency, not fully resolved.** `wave.md` line 5 still reads
   `status: accepted`, while `closure-report.md` (correctly, per the prior remediation) reads
   `status: in-progress` with an explicit note this is an interim state as of 2026-07-16 pending the
   closure of open conditions. **`wave.md`'s `accepted` status is not accurate** while W04-E01-S001/
   S002/S003's closure paperwork remains unfilled and W04-E02-S002 is `planned` (not accepted) — the
   conductor should not treat `wave.md`'s `accepted` token as authoritative until these are resolved.
3. **Status-vocabulary token `closed-pending-review`** (used by W04-E04-S001 and W04-E04-S002) is not
   part of the documented status vocabulary. Both stories PASS their review under this gate;
   recommend normalizing the token to `accepted` (given the PASS verdicts) or introducing a defined
   `in-review` value, rather than leaving an ad hoc token in the tracking system.
4. **Evidence-policy violation found and fixed**: W04-E04-S003's evidence record cited `HEAD` (a
   moving target) instead of a pinned commit SHA, violating evidence-policy.md's revision-pinning
   rule verbatim ("never 'current HEAD'"). Corrected during this review.

## Overall wave recommendation

**Accept-with-conditions**, pending: (a) W04-E01-S001/S002/S003's `closure.md` Final-status sections
being filled in (code is real and tested; this is a documentation-completeness gate only), (b)
`wave.md`'s status being reconciled with `closure-report.md`'s more accurate `in-progress` framing
until (a) is done and W04-E02-S002 is either implemented or permanently excluded from the wave's
acceptance scope, and (c) the `closed-pending-review` status token being normalized. No story in
scope has a surviving code-level AC failure except W04-E02-S002, which is correctly and honestly
labeled `planned`/not-ready and should not block the rest of the wave's closure once (a)/(b)/(c) are
addressed.

This is a recommendation only; the conductor adjudicates final status changes.
