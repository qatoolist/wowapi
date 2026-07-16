# Implementation Autopsy Report — wowapi Premier-Framework Programme (Waves 00–07)

**Date:** 2026-07-16 · **Repo state audited:** `main @ 43b6e12` (+1 untracked file) · **Baseline:** `./impl` programme (mandate, 8 waves, 75 stories, ~370 tasks, registers)
**Review authority:** Fable 5 (senior reviewer & final quality gate) supervising 10 Haiku extraction workers and 13 Sonnet verification/specialist workers.

---

## 1. Executive verdict and readiness status

**Verdict: the programme is NOT complete and the framework is NOT production-ready, despite commit `e8cda6b` ("feat: finalize wowapi implementation programme (Waves 00-07)") claiming full finalization.**

The true condition, established by independent execution and code inspection:

- **25 of 75 stories (33%) are fully implemented and independently verified.** A further ~17 are implemented in substance but incomplete, unreviewed, or resting on invalid evidence.
- **The programme's status ledger is unreliable.** Four waves (W02, W04, W06, W07) have wave/epic/story/register statuses that contradict each other four ways; the W04 wave closure report claiming `accepted` is literally the unexecuted pre-execution template ("Wave 04 has not begun execution").
- **At least three false completion claims exist on security-relevant stories**, including one (W04-E02-S002) whose own closure.md admits "not implemented, verified, or closed" while story.md says `accepted`.
- **One confirmed code defect contradicts an accepted acceptance criterion:** webhook outbound delivery performs secret resolution and the HTTP POST inside an open DB transaction (`foundation/webhook/service.go:237→401,424`), the exact anti-pattern story W04-E02-S001 claims to have eliminated.
- **The quality gate was silently weakened at closure:** coverage floor lowered 90.0%→84.0% in `e8cda6b` with no deviation or decision record; measurement scope unchanged, so the 92.3%→84.5% drop is a genuine coverage regression absorbed by moving the goalpost.
- Independent execution results: `go build ./...` PASS, `go vet` PASS, `make test` PASS (no failures), `make coverage-check` PASS at the *lowered* floor (84.5% vs 84.0%), `make lint` **FAIL** (untracked `kernel/database/isolation_test.go`, gofumpt).

Real, high-quality engineering exists — W00/W01 substantially hold up, W02's data-safety code is largely real, SEC-01/SEC-03 identity and webhook work is substantive, W04's lease/fencing and audit-chain work verified, release gating (W06-E03) verified. The failure is one of **governance and truthfulness of completion claims**, plus a small number of real defects.

**Production-readiness claim: REJECTED.** Programme closure gate (W07-E04): never executed.

---

## 2. Review method, agent allocation, cost summary

**Method** (defined by Fable): (A) mechanical extraction of the full plan baseline; (B) one shared local execution-evidence pass (build/vet/lint/test/coverage — run once, reused by all agents); (C) adversarial per-wave verification + 5 specialist reviews, all instructed to refute rather than confirm and forbidden from declaring anything complete; (D) Fable adjudication of every Critical/High finding and every dispute, with direct code reads for load-bearing claims; (E) this report.

| Tier | Agents | Scope | Tokens |
|---|---|---|---|
| Local scripts (no LLM) | — | status extraction from front matter, verdict-matrix compilation, git forensics, build/test/lint/coverage runs | 0 |
| Haiku ×10 | extraction | 8 waves + registers/governance + repo inventory | ~710k |
| Sonnet ×13 | verification | 8 wave verifiers + DB, security, architecture, tests/CI, docs/dead-work specialists | ~1.17M |
| Fable (direct) | adjudication | methodology, dispute resolution, code-level confirmation of Criticals, severity approval, this report | conductor session only |

Process/cost figures (agent counts, token totals) are self-reported session metadata and are not independently auditable from the repo, unlike the code-level findings, which are reproducible via the commands in §12. No duplicate scans: one full test/lint/coverage run served all 13 verifiers; extraction JSONs were the single shared plan-of-record. Fable performed no mechanical extraction.

## 3. Fable supervision and approval record

- Defined methodology and all agent scopes; barred workers from completion declarations (enforced in every prompt).
- **Personally confirmed by direct code/git inspection:** the in-transaction webhook HTTP POST (read `foundation/webhook/service.go`), the coverage-floor change and its unchanged measurement scope (`git show e8cda6b^:Makefile` vs `e8cda6b`), the absence of any coverage deviation record (read full `impl/tracking/deviation-register.md`), the SEC-01 artifacts' existence (`migrations/00039_identity_grant.sql`, `kernel/auth/auth.go`), and the non-ancestor W07 evidence pins (`git merge-base --is-ancestor` on `733ef3e`, `1626b11`).
- **Disputes adjudicated:** (i) W07 evidence pins → mis-pinned after the `e8cda6b` squash, *not* fabrication — but invalid as proof under `impl/governance/evidence-policy.md` until re-pinned; (ii) coverage 92→84.5 → genuine regression, not a scope change; (iii) W05 "planned vs landed" → tracking failure (work executed outside the programme's ledger), adjudicated High-governance rather than Critical-security; (iv) "does code-level completion count toward W03 acceptance despite tracking saying planned" → **No.** Under the programme's own DoD, acceptance requires the review gate; code alone is `implemented-incomplete`; (v) conductor self-adjudications labelled "conductor" (W00 DEV-02, W03-E02 self-review) do **not** meet the mandate's independent-review bar.
- Downgrades/merges applied: 6 proposed Criticals consolidated to 5 (W02 4-way status contradiction folded into the systemic ledger finding); W05 under-claiming downgraded to High.

---

## 4. Plan-to-implementation traceability matrix (75 stories: planned → claimed → independently verified)

Classification legend: verified · implemented-incomplete · implemented-incorrectly · insufficiently-tested · unsupported-by-evidence · deviated · missing · blocked · contradictory · obsolete-na. "Claimed" is the story.md front-matter status at audit time.

| Item | Claimed | Independent verdict | Basis |
|---|---|---|---|
| W00-E01-S001 | accepted | **implemented-incomplete** | 3 of 4 ACs genuinely pass with reproducible, pinned evidence (verified AC-01/02/03 test names and source-line claims are internally consistent and the |
| W00-E01-S002 | accepted | **verified** | Documentation is internally consistent, cites concrete test names/line numbers and exact numeric bench-budget values (SweepAt10k 451615ns/4500000 budg |
| W00-E01-S003 | accepted | **implemented-incomplete** | Same pattern as S001: the story's own closure record was never updated past ready-for-review/pending, yet epic/wave closure-report.md, acceptance.md,  |
| W00-E02-S001 | accepted | **verified** | The point-in-time baseline capture is well-evidenced and honestly flags its own analyzer-name-matching gap (DEV-W00-E02-S001-001, 18-of-25 analyzers p |
| W00-E02-S002 | accepted | **verified** | Artifacts and evidence files exist with real command output; claim of 13/13 approved-dependency cross-check with zero drift is plausible given the raw |
| W00-E02-S003 | accepted | **verified** | This story is the one place in the wave where an actual review-gate artifact exists in the shape evidence-policy.md requires (evidence ID, reviewer id |
| W01-E01-S001 | accepted | **verified** | sqlclosecheck/rowserrcheck/bodyclose/wastedassign/makezero/musttag/testifylint/noctx/copyloopvar are all present and annotated with W01-E01-S001 story |
| W01-E01-S002 | accepted | **verified** | Judged linter set confirmed enabled with an explanatory comment block citing W01-E01-S002/FBL-07. Did not independently re-verify every named fix site |
| W01-E01-S003 | accepted | **implemented-incomplete** | Closure report itself records: 'Carry-forward: first in-CI executions of go-mod-verify step and license scanner occur on next push (wave uncommitted a |
| W01-E02-S001 | accepted | **verified** | Code and tests are real and reproduce green independently. However see Finding F-1 (Reviewer field on all 3 evidence records literally reads 'Pending  |
| W01-E02-S002 | accepted | **verified** | Implementation matches the claimed design (vendor-neutral port, not otelpgx) — grep confirms no otel import in query_tracer.go. Same evidence-policy g |
| W01-E03-S001 | accepted | **verified** | Claimed artifacts all exist; evidence bundle for this story uses a different (non-markdown-template) but complete fail-first log format, all five ACs' |
| W01-E03-S002 | accepted | **verified** | config.Security.EnforceRouteContracts flag is wired into Router.Handle per deviations.md's account, confirmed present in router.go. Extraction's repor |
| W01-E04-S001 | accepted | **verified** | Independently reproduced the two most safety-critical tests (generator-output-boots and route-contract enforcement in generated code) — both PASS agai |
| W01-E04-S002 | accepted | **unsupported-by-evidence** | This story is titled 'Documentation reconciliation' (plan-traceability fix, DX-05 residual, wowsociety upstream register) yet its only claimed_artifac |
| W01-E04-S003 | accepted | **verified** | Claimed artifact/evidence bundle physically exists with the claimed shape (diagnosis note + reproduction runs + log collection). Did not re-run the fl |
| W02-E01-S001 | accepted | **contradictory** | Code artifacts genuinely exist and match filenames/test-name claims. But the story's own independent-review task file (task-003-independent-review.md) |
| W02-E01-S002 | accepted | **contradictory** | Same pattern as S001: code exists, but the independent-review task for this story was never marked complete. story.md claims acceptance the underlying |
| W02-E01-S003 | accepted | **contradictory** | Code and CI workflow exist as claimed. Independent-review task record for this story is still todo, contradicting story.md's acceptance claim. |
| W02-E02-S001 | accepted | **implemented-incomplete** | Code, CI wiring, and passing tests all verified genuine. However the independent-review task record is still todo despite story.md claiming 'complete  |
| W02-E02-S002 | accepted | **implemented-incomplete** | Functionally verified: mismatch audit is genuinely zero and cross-tenant inserts are genuinely blocked (9 subtests, all pass) -- the safety property h |
| W02-E03-S001 | accepted | **verified** | All 5 claimed acceptance-criteria tests genuinely exist and pass against the real DB. This story has no separate independent-review task file found un |
| W02-E04-S001 | accepted | **implemented-incomplete** | Functional tests genuinely pass. Independent-review task record again shows todo despite story.md claiming completion, same systemic pattern as W02-E0 |
| W02-E05-S001 | accepted | **insufficiently-tested** | This is the ONLY W02 story with an actual filled-in independent-review artifact rather than an untouched todo template, but that artifact itself viola |
| W03-E01-S001 | accepted | **implemented-incorrectly** | AC-01 (schema) and AC-03 (zero/garbage tenant rejection) verified correct and passing. AC-02 claims ActiveTenantAccess is called 'unconditionally... n |
| W03-E01-S002 | ready | **implemented-incomplete** | Capacity selection and privileged-session resolver logic is genuinely implemented and adversarially tested (revoked/expired/wrong-tenant/wrong-actor/u |
| W03-E01-S003 | accepted | **implemented-incomplete** | Both required test classes exist with the exact names cited in the evidence index and pass under `go test`. However closure.md itself is internally co |
| W03-E01-S004 | accepted | **unsupported-by-evidence** | Not deep-read line by line under this time budget; files exist. Wave.md's closure condition explicitly requires these documents be 'reviewed and accep |
| W03-E02-S001 | accepted | **implemented-incorrectly** | The technical implementation (fingerprint scope, egress report, allowlist audit, D-07 JWKS governance gate, fitness check) is real, present, and the t |
| W03-E03-S001 | accepted | **contradictory** | The core implementation (breaking Verifier interface -> Envelope, HandleInbound rewired to consume only authenticated envelope fields, provider-verifi |
| W03-E04-S001 | story.md: ready; closure.md st | **implemented-incomplete** | This story's status is honestly and consistently reported as not-yet-'accepted' (story.md 'ready', closure.md 'implemented' with T004 independent revi |
| W03-E05-S001 | story.md: ready; closure.md st | **implemented-incomplete** | Unlike E02-S001 and E03-S001, this story's closure.md is internally consistent: it does not claim 'accepted' and explicitly states it is awaiting inde |
| W04-E01-S001 | accepted | **implemented-incomplete** | Real, working shared lease primitive exists (kernel/lease: Lease struct, Token/Generation/ExpiresAt, IsCurrent/IsNewer/NextEpoch/BumpGeneration) with  |
| W04-E01-S002 | accepted | **implemented-incomplete** | Migration 00038 for jobs lease columns exists (31 lines). closure.md 'Final status' is again the unfilled template text, contradicting story.md's 'acc |
| W04-E01-S003 | accepted | **verified** | Named chaos test kernel/jobs/chaos/duplicate_worker_lease_expiry_test.go exists and PASSES against the real Postgres instance (TestDuplicateWorkerLeas |
| W04-E02-S001 | accepted | **implemented-incorrectly** | notify.SendPending genuinely implements a three-stage protocol: claimPending (tx) -> effectSend (outside tx, calls sender.Send) -> finalizeDelivery (t |
| W04-E02-S002 | accepted | **unsupported-by-evidence** | CRITICAL false-completion. story.md frontmatter claims status: accepted and lists all 4 ACs as satisfied (per the extraction's completion_claims pulle |
| W04-E02-S003 | accepted | **verified** | cenkalti/backoff/v5 v5.0.3 is a real go.mod dependency; kernel/retry/retry.go wraps it (NewSchedule, SequenceBackOff) and foundation/notify/service.go |
| W04-E03-S001 | accepted | **verified** | Migration 00016's header comment now reads '-- (single processor per operation; multi-worker fan-out is added later by the' -- i.e. the false 'safe ac |
| W04-E03-S002 | accepted | **verified** | Migration 00044 (60 lines) adds bulk_items lease/lifecycle columns. The named chaos test at foundation/bulk/chaos/duplicate_worker_test.go (208 lines, |
| W04-E04-S001 | accepted | **verified** | hash_version discriminator (v1/v2 branching) is genuinely implemented in kernel/audit/audit.go with version-specific hashing and fail-closed handling  |
| W04-E04-S002 | closed-pending-review | **implemented-incomplete** | This story's status is honestly labeled 'closed-pending-review' in both extraction and story.md/closure.md (status: draft in closure.md, 'Implemented  |
| W04-E04-S003 | accepted | **unsupported-by-evidence** | Not independently verified given time budget -- did not locate/exercise the /readyz migration-currency check or config doctor product-root discovery c |
| W05-E01-S001 | planned | **contradictory** | Docs say 'todo'/'not yet executed', which is true in the sense that no Registrar capability type is wired into kernel/kernel.go's actual registration  |
| W05-E01-S002 | planned | **missing** | No owner-bound Registrar wrapping exists around resource.Registry/rules.Registry/authz.Registry in kernel/kernel.go; registries are still constructed  |
| W05-E01-S003 | planned | **missing** | No AR-01/race_test_output.txt, model_hash_determinism_test.go, snapshot_immutability_test.go found anywhere in repo outside impl/ evidence templates ( |
| W05-E01-S004 | planned | **missing** | No AR-01/legacy_adapter_compat_test_output.txt or equivalent adapter code found. Consistent with claimed status. |
| W05-E02-S001 | planned | **contradictory** | A real port.Key[T] API (Define/Provide/Require/Resolve generic free functions) and a registrar_forge_compile_fail_fixture directory already exist, mat |
| W05-E02-S002 | planned | **missing** | None of the claimed AR-02 T1-T3 artifacts exist beyond the orphaned port.go skeleton (see S001). No provider graph, no boot validation, no profile pro |
| W05-E02-S003 | planned | **missing** | kernel/lifecycle directory still exists unmodified; no lint-generated replacement or legacy port adapter found. |
| W05-E03-S001 | planned | **contradictory** | A repo-root './AR-03/' directory (package ar03_test) exists containing exactly the four test files this story's claimed_artifacts list names, and the  |
| W05-E03-S002 | planned | **missing** | No duplicate-collector rejection, empty-required-fragment rejection, or waiver-mechanism source found. Consistent with claimed status. |
| W05-E04-S001 | ready-for-review, tasks done | **implemented-incorrectly** | T2 (constructor-boundary lint) is genuinely implemented and passing: `go test -v ./internal/tools/constructorlint/...` reproduces PASS for both TestAn |
| W05-E04-S002 | planned | **unsupported-by-evidence** | The story's own claimed_artifacts cite kernel/authz/caching.go:29-36 and kernel/kernel.go:118-121 as if these already exist and are load-bearing for t |
| W05-E05-S001 | planned | **contradictory** | This story is tracked 'planned', every task 'todo', yet the FBL-01 re-home this story describes is, in substance, ALREADY DONE and WIRED: all nine pac |
| W05-E05-S002 | planned | **insufficiently-tested** | Given S001's actual re-home state, a package-count/lint verification pass and a wowsociety identity-suite run are plausible next steps, but no evidenc |
| W06-E01-S001 | verified | **unsupported-by-evidence** | Story-level artifact claims verified with deviations noting acceptance-authority disposition 'pending' and W05 entry-gate not confirmed closed. Wave-l |
| W06-E01-S002 | accepted | **insufficiently-tested** | TestGoldenConsumerInstalledBinaryTwoModules passed (11.3s), generating all 8 claimed subsystem types (resource/rule/workflow/event-handler/recurring-j |
| W06-E02-S001 | verified | **unsupported-by-evidence** | Code artifact (internal/cli/openapi_merge.go) exists matching the claim of a full-field merge implementation, but detailed per-field policy coverage a |
| W06-E02-S002 | accepted | **unsupported-by-evidence** | CI wiring for required-gates.yml exists and is referenced from ci.yml (job 'required-gates' at line 69/75). Did not independently verify each of the 6 |
| W06-E02-S003 | blocked | **verified** | Blocked status is internally consistent and plausible: the story's three legs are genuinely gated on other in-wave/cross-wave stories whose own accept |
| W06-E03-S001 | verified | **verified** | Claimed artifacts (required-gates.yml reusable workflow, release.yml verify/build-candidate/publish jobs, verify_release.sh) all exist on disk. T006 i |
| W06-E03-S002 | blocked | **verified** | Blocked-genuinely: story.md explicitly documents that branch protection / release-environment / tag-ruleset activation requires a human repo-admin (Gi |
| W06-E03-S003 | verified | **verified** | Trivy wiring confirmed present across all three claimed workflow files. Waiver mechanism, visibility-guard regression check, and local-scanner fallbac |
| W06-E04-S001 | accepted | **verified** | docexamples tool exists on disk matching the claimed CI-enforced Go example compilation gate. Did not independently re-run make docs-check or the stal |
| W06-E04-S002 | accepted | **unsupported-by-evidence** | Story's own deviation record states this story consumed a W05 artifact (AR-03's GenerateProjections/Doc export) whose own W05 story/task records the d |
| W07-E01-S001 | accepted | **implemented-incomplete** | Core claim (DB-backed request benchmarks, 6 profiles, cost attribution, publication conditional on DEC-Q9) is plausible and the epic closure's cross-c |
| W07-E01-S002 | accepted | **implemented-incorrectly** | All four JSON evidence records for this story cite the identical base_commit 733ef3e930cbb3f89f5bbc53d8f562c60e426513. That commit is real (git cat-fi |
| W07-E01-S003 | accepted | **unsupported-by-evidence** | Confirmed real code artifacts exist for the bounded-batch sweeper (kernel/workflow/sweeper.go, sweeper_perf_test.go) consistent with the claim -- this |
| W07-E01-S004 | accepted | **verified** | Both halves of this story's claim independently confirmed against the real repo, not just docs: (1) required-checksum/bounded-repair-path types exist  |
| W07-E02-S001 | blocked | **blocked** | Genuinely and honestly not accepted. closure.md explicitly states 'No genuine external professional-services assessment exists', artifact ART-W07-E02- |
| W07-E02-S002 | accepted | **verified** | Independently confirmed: Makefile has .PHONY check-test-skips check-required-test-prerequisites check-race-fixture test-race-integration targets with  |
| W07-E03-S001 | blocked | **blocked** | Genuinely blocked with specific, falsifiable reasons: AC01 fails on absent rule_versions(tenant_id,id) unique parent key; AC04 fails because W03-E01-S |
| W07-E04-S001 | planned | **missing** | Confirmed genuinely unstarted: closure.md is explicitly a template with 'This story has not been implemented, verified, or closed... it must not be fi |
| W07-E04-S002 | planned | **missing** | Consistent with W07-E04-S001's genuinely-unstarted pattern and with wave.md/progress.md, which both agree E04 is planned/not-ready. No contradiction f |

**Totals:** 25 verified · 14 implemented-incomplete · 5 implemented-incorrectly · 3 insufficiently-tested · 10 unsupported-by-evidence · 8 contradictory · 8 missing · 2 blocked.

---

## 5. Findings (Fable-adjudicated)

### Critical

| ID | Finding | Location / plan ref |
|---|---|---|
| C-1 | Webhook outbound delivery performs secret resolution and HTTP POST **inside an open DB transaction** (both dispatch and retry paths), contradicting accepted AC-W04-E02-S001-02/03. Long-held transactions under slow/unresponsive endpoints risk pool exhaustion and lock retention. The sibling notify path (`SendPending`) correctly stages claim/send/finalize — the fix pattern exists in-repo and was not applied to webhook. | `foundation/webhook/service.go:237,266 → 401,424` · W04-E02-S001 |
| C-2 | **False completion claim:** W04-E02-S002 `story.md status: accepted` while its own closure.md/verification.md/evidence state "not implemented, verified, or closed"; no chaos tests for notify/webhook exist; inbound verification runs in a single transaction, not the claimed two-phase protocol. | `impl/waves/wave-04.../epic-002/stories/story-002.../` · DATA-03 |
| C-3 | **False completion claim on a security story:** W03-E03-S001 (SEC-03 webhook authenticated replay) closed `accepted` while its independent-review task is `status: todo`, entirely unexecuted, and the closure's own reviewer-conclusion admits review is pending. | `impl/waves/wave-03.../epic-003/stories/story-001.../closure.md` · SEC-03 |
| C-4 | **Review gate falsely claimed passed for W02:** closure-report asserts "Independent review passed (W02ReviewGate)" while all 6 of the epic-level independent-review task files remain unfilled templates (`status: todo`, empty evidence). AC-W02-06 is unsupported. | `impl/waves/wave-02.../closure-report.md` + task-*-independent-review.md files · AC-W02-06 |
| C-5 | **Systemic status-ledger unreliability:** W02 (wave.md `planned` vs closure `accepted` vs all 8 story.md `accepted` vs register `planned`), W04 (closure-report is the unexecuted template yet front-matter `accepted`; register lists all stories `planned` and omits one entirely), W03 (story.md vs progress.md disagree), W06/W07 (roll-ups frozen at `planned`/`not begun` under 8/10 verified-or-accepted stories). No single surface in `impl/` currently tells the truth. | `impl/tracking/status-register.md` + wave roll-ups · mandate §6 |

### High

| ID | Finding | Location / plan ref |
|---|---|---|
| H-1 | Coverage floor lowered 90.0→84.0 in `e8cda6b` with no deviation/decision record; `COVER_PKGS`/`COVER_EXCLUDE` unchanged, so the 92.3→84.5 drop is a real regression absorbed by weakening the gate — in the same commit that declares the programme finalized. | `Makefile:324` · mandate §2.6/§8.9, AC-W00-03 |
| H-2 | Commit `e8cda6b` message claims Waves 00–07 finalized; the programme's own artifacts (W07 `in-progress`, SEC-05 story `blocked`, open unaccepted RISK-W07-002, register un-regenerated) contradict it. Unsupported completion claim at the VCS level. | git history · programme acceptance §index.md |
| H-3 | SEC-01's "unconditional" tenant-membership check is silently skippable: `Verifier.Actor` only verifies membership when the `PrincipalStore` also implements `AssurancePrincipalStore` (runtime type assertion); a base-interface store skips the check with no fail-closed fallback. Not exploitable via the shipped `pgprincipal.Store`, but the kernel API permits silent security regression, contradicting CS-07's contract. Untested, undocumented. | `kernel/auth/auth.go` · SEC-01/CS-07 |
| H-4 | Wave review gates for W00, W01, W06 have **no registered evidence**: W01's evidence records carry `Reviewer: Pending — conductor acceptance gate` on every record (per `evidence-policy.md`, such records "must not be cited as proof"); W00's `W00ReviewGate` has no artifact anywhere; W06's closure says "Not yet reviewed". | wave closure-reports · AC-W00-07, AC-W01-11 |
| H-5 | W03-E02-S001 (SEC-06 outbound security governance) accepted on an explicit **self-review** ("A separate reviewer still needs to ratify the evidence bundle"). | `impl/waves/wave-03.../epic-002/.../task-006` · SEC-06 |
| H-6 | AR-01/AR-02 (`kernel/appmodel`, `kernel/port`, ~765 LOC, tested) are **built but not wired** — zero non-test imports; boot still constructs registries directly. SEC-04 bounded/epoch authz cache is **not implemented**: the existing cache is unbounded TTL; the `authz_epoch` migration is orphaned (no Go code reads it). | `kernel/appmodel/`, `kernel/port/`, `kernel/authz/caching.go` · AR-01/02, SEC-04 |
| H-7 | Dependency/sequencing gates bypassed without deviation records: FBL-01 kernel re-home fully executed on main (all 9 packages under `foundation/`, shims in place) while W05 says `planned` and RISK-001 still warns to do it; W05-E04-S001 reached `ready-for-review` despite W05's hard entry gate on W03-E01 acceptance (W03 unaccepted); W04 executed while W02 unclosed. | `foundation/*`, `impl/waves/wave-05/...` · W05 entry criteria, RISK-001 |
| H-8 | W07-E01 performance evidence pinned to commits **not reachable from HEAD** (`733ef3e`, `1626b11` — squash side-effect, adjudicated not-fabrication); invalid as proof under evidence-policy revision-pinning until re-pinned; S003/S004's 15 evidence files unswept. | `impl/waves/wave-07.../epic-001/.../evidence/` · PERF-02/03 |
| H-9 | Webhook tamper matrix incomplete: no key-ID or signature-version manipulation tests (3 of 5 required fields covered). | `foundation/webhook/*_test.go` · SEC-03 AC |
| H-10 | Canonical status surfaces outside `impl/` (docs/GOALS-TRACKER.md, docs/SRS.md) have zero visibility of the Waves 00–07 programme (last reconciled 2026-07-05, pre-`impl/`). Two "canonical" ledgers fully diverged. | `docs/GOALS-TRACKER.md`, `docs/SRS.md` |

### Medium (summary)

M-1 W02-E05's sole filled review artifact lacks every evidence-policy mandatory field (no evidence ID/SHA/reviewer/command). · M-2 Five W04 stories carry unfilled template text in closure.md despite `accepted` front matter. · M-3 Golden-consumer upgrade-replay (DX-04) not independently reproducible this session (skips without DATABASE_URL; fails with manual DSN); base generation test passes; needs `make ensure-infra golden-consumer` re-run. · M-4 `kernel/tracing` and `kernel/safety` at 0% coverage despite real logic (`kernel/webhook`/`notify` 0% adjudicated acceptable — 53-line compat shims over tested `foundation/webhook`). · M-5 AR-06 constructor audit count wrong (audit says 23; actual 33 New*-pattern calls, or 17 under the lint's own definition). · M-6 `AR-03/` and `SEC-05/` tracked at repo root with no placement rationale. · M-7 CHANGELOG `[Unreleased]` empty across 5 commits incl. the finalize commit. · M-8 W06-E04-S002 acceptance consumed an unaccepted W05 draft artifact (self-recorded deviation). · M-9 W01-E01-S003 accepted while its CI evidence had never run in CI (local + actionlint only). · M-10 Claimed artifact paths wrong in W03 docs (`kernel/webhook`, `pgprincipal/pgprincipal.go` vs actual `foundation/webhook`, `adapters/auth/pgprincipal/`).

### Low / Observation (summary)

L-1 Untracked `kernel/database/isolation_test.go`: real, passing adversarial RLS test; fails gofumpt; untraceable to any story; overlaps `testkit/rls_isolation_all_test.go`. · L-2 `impl/.ruff_cache` Python caches in the Go repo's plan tree; SEC-05 Python tooling undocumented. · L-3 Product vocabulary ("committee_seat", "wowsociety") in kernel doc comments (comment-only). · O-1 REL-04 fuzz evidence genuine but narrow (3 targets, 2 files). · O-2 DEC-Q10 branch-protection blocking unverifiable locally — correctly registered open-human. · O-3 W00's DEV-002 adjudication (AC-04 re-scope) honestly disclosed and reproduced; disclosure quality is good where it exists.

---

## 6–9. Domain reviews (condensed verdicts)

**Architecture & code quality (§6).** New W01–W04 code is genuinely good: staged outbox/notify delivery, lease/fencing (migration `00044`, `claimBatch` atomic UPDATE…FOR UPDATE SKIP LOCKED), audit chainHash over all persisted fields — all verified. Structural debt: built-but-not-wired AR-01/02 (H-6), the silently-completed FBL-01 re-home (H-7), root-level requirement-named dirs (M-6), comment-level vocabulary leaks (L-3). No structural framework/product boundary violation found.

**Database & migrations (§7).** 48 migrations reviewed; DATA-09 online-migration protocol and DATA-01 composite tenant FKs verified real (5 of 6 DB verdicts verified); `identity_grant` (00039) with RLS present; lease columns (00044) verified; `authz_epoch` orphaned (H-6). No integrity defect found in migration content; the defect class is transactional discipline in webhook delivery (C-1).

**Security, isolation, reliability, concurrency (§8).** SEC-01 substantively implemented (server-side grants, membership verification, privileged resolver) but with the H-3 skippable-check contract gap and no valid acceptance review; SEC-03 implemented with H-9 tamper-matrix gaps and a false-closed review (C-3); SEC-04 missing (H-6); SEC-05 blocked (correctly); SEC-02/SEC-06 partially verified only. RLS FORCE spot-checked live (the untracked isolation test passes against the running DB). Concurrency: lease/fencing verified; DATA-03's repo-wide "no remote I/O in tx" sweep found the webhook violation (C-1); full sweep of remaining call sites is remediation work R-2.

**Tests, coverage, CI/CD, evidence quality (§9).** Tests pass; race/integration suites real; coverage regression + silent floor drop (H-1); 0%-coverage kernels (M-4); 11 CI workflows present and actively wired (verified for release gating, scans, golden consumer as a required gate) but merge-blocking unverifiable locally (O-2/DEC-Q10). Evidence quality is the programme's weakest axis: unfilled reviewer fields (H-4), non-ancestor pins (H-8), missing mandatory fields (M-1), artifact:// URIs unresolvable for independent replay.

## 10–11. Deviations, unsupported claims, dead work, production risks

Recorded deviations (11 rows, W00/W01 only) are honest and well-formed — the deviation *mechanism* works where it was used. **Unrecorded deviations:** coverage floor (H-1), FBL-01 early execution (H-7), W05/W04 sequencing-gate bypasses (H-7), W06 consuming draft W05 artifacts (M-8). **Unsupported completion claims:** C-2, C-3, C-4, H-2, H-4, H-5, plus W06-E01-S001/E02-S001/E02-S002 and W07-E01-S003 story claims (see matrix). **Dead/abandoned work:** orphaned `kernel/appmodel`+`kernel/port` (pending wiring decision), orphaned `authz_epoch` migration, untracked isolation test, root-level logs/binaries (`testkit.test`, `extraction.log`, `full_*.log`, `coverage.db.out` — untracked junk to clean), `impl/.ruff_cache`.

**Production/failure-condition risks:** webhook delivery under endpoint outage holds DB transactions open (C-1 — pool exhaustion is the concrete failure mode); silent tenant-membership skip if a non-assurance PrincipalStore is ever wired (H-3); unbounded authz cache growth (H-6/SEC-04); audit DSR-hold path unreviewed (W04-E04-S002); no executed final verification gate means no evidence the assembled system was ever validated end-to-end (W07-E04 missing).

---

## 12. Traceability & evidence

Full machine-readable artifacts (per-story verdicts incl. evidence checked and commands run) are retained at the review workspace: `scratchpad/autopsy/{extraction,verification,specialist}/*.json`, `verdict-matrix.tsv`, `evidence/quality-run.log`. Key commands any re-reviewer can replay: `make lint` (fails), `make test`, `make coverage-check` (84.5/84.0), `git show e8cda6b^:Makefile | grep COVERAGE_FLOOR` (90.0), `git merge-base --is-ancestor 733ef3e HEAD` (fails), `grep -n 'sender.Post' foundation/webhook/service.go`.

## 13. Remediation plan (priority order; dependencies noted)

| # | Action | Fixes | Depends on |
|---|---|---|---|
| R-1 | Truth-reconciliation pass over `impl/`: revert false statuses (W04-E02-S002→planned, W03-E03-S001→implemented-unreviewed, W02 closure→in-review), fill or delete template closures, regenerate status-register from front matter, record the missing deviations (floor, FBL-01, sequencing) | C-2..C-5, H-2, H-7, M-2 | — |
| R-2 | Fix webhook outbound: stage claim/deliver/finalize outside the tx (mirror `notify.SendPending`); sweep remaining DATA-03 call sites; add no-network-while-tx-open test for webhook | C-1 | — |
| R-3 | Execute the missing independent reviews with a real, named reviewer: W02 (6 stories), W03 (all), W04-E02/E04-S002; re-run gates for W00/W01/W06 with registered evidence records (reviewer field filled) | C-3, C-4, H-4, H-5 | R-1 |
| R-4 | Coverage: restore ≥90% (target the regression packages incl. `kernel/tracing`, `kernel/safety`) or ratify the 84% floor via a decision record | H-1, M-4 | — |
| R-5 | Close the SEC-01 contract gap: make membership verification fail-closed for any PrincipalStore (or fold Assurance methods into the base interface); add the missing webhook key-ID/signature-version tamper tests | H-3, H-9 | — |
| R-6 | Re-pin W07-E01 evidence at HEAD (re-run the 4 pinned benches); sweep S003/S004 evidence files | H-8 | — |
| R-7 | Decide and execute AR-01/02 wiring + SEC-04 cache (or formally defer with records); then finish W05 per plan | H-6 | R-1 |
| R-8 | Complete W06 open stories (DX-06, REL-03b) once W05/AR-03 unblocks; resolve DEC-Q1/Q9/Q10 human decisions | blocked items | R-7 |
| R-9 | Run the programme closure gate (W07-E04) for real; only then revisit the production-readiness claim | H-2 | R-1..R-8 |
| R-10 | Hygiene: commit+format or delete `isolation_test.go`; clean root logs/binaries; relocate `AR-03/`/`SEC-05/`; update GOALS-TRACKER/SRS/CHANGELOG | H-10, M-6, M-7, L-1/2 | — |

## 14. Final verdict per wave

| Wave | Claimed | Fable verdict |
|---|---|---|
| W00 | accepted | **Accepted-with-reservations** — 4/6 stories verified; review-gate evidence missing (H-4); coverage baseline superseded silently |
| W01 | accepted | **Accepted-with-reservations** — 7/10 verified; gate evidence records invalid per own policy (H-4) |
| W02 | contradictory | **Closure REJECTED** — code substantially real (DATA-09/01 verified) but review gate falsely claimed (C-4); statuses contradictory (C-5) |
| W03 | planned/ready/accepted (split) | **Open — implemented-unaccepted** — SEC-01/03 substantive in code; zero stories validly accepted (C-3, H-5); contract + tamper gaps (H-3, H-9) |
| W04 | accepted | **Acceptance REJECTED** — E01/E03/E04-S001 verified real; E02 contains false acceptances and the confirmed code defect (C-1, C-2); closure report is a template |
| W05 | planned | **Not executed as a wave** — 8 stories missing; FBL-01 done off-ledger; AR-01/02 built-not-wired; SEC-04 missing (H-6, H-7) |
| W06 | planned (roll-up) / verified stories | **Partially verified, wave open** — E03 release gating + E04-S001 verified; 4 story claims unsupported; no wave gate |
| W07 | in-progress | **In progress** — E01 perf work real but evidence mis-pinned (H-8); E02/E03 legitimately blocked; closure gate (E04) not run |

### Epic-level verdicts (33 epics, derived worst-of-stories roll-up)

| Epic | Stories | Independent verdict (worst-of-stories) | Story verdict mix |
|---|---|---|---|
| W00-E01 | 3 | **implemented-incomplete** | 2×implemented-incomplete, 1×verified |
| W00-E02 | 3 | **verified** | 3×verified |
| W01-E01 | 3 | **implemented-incomplete** | 1×implemented-incomplete, 2×verified |
| W01-E02 | 2 | **verified** | 2×verified |
| W01-E03 | 2 | **verified** | 2×verified |
| W01-E04 | 3 | **unsupported-by-evidence** | 1×unsupported-by-evidence, 2×verified |
| W02-E01 | 3 | **contradictory** | 3×contradictory |
| W02-E02 | 2 | **implemented-incomplete** | 2×implemented-incomplete |
| W02-E03 | 1 | **verified** | 1×verified |
| W02-E04 | 1 | **implemented-incomplete** | 1×implemented-incomplete |
| W02-E05 | 1 | **insufficiently-tested** | 1×insufficiently-tested |
| W03-E01 | 4 | **implemented-incorrectly** | 2×implemented-incomplete, 1×implemented-incorrectly, 1×unsupported-by-evidence |
| W03-E02 | 1 | **implemented-incorrectly** | 1×implemented-incorrectly |
| W03-E03 | 1 | **contradictory** | 1×contradictory |
| W03-E04 | 1 | **implemented-incomplete** | 1×implemented-incomplete |
| W03-E05 | 1 | **implemented-incomplete** | 1×implemented-incomplete |
| W04-E01 | 3 | **implemented-incomplete** | 2×implemented-incomplete, 1×verified |
| W04-E02 | 3 | **implemented-incorrectly** | 1×implemented-incorrectly, 1×unsupported-by-evidence, 1×verified |
| W04-E03 | 2 | **verified** | 2×verified |
| W04-E04 | 3 | **unsupported-by-evidence** | 1×implemented-incomplete, 1×unsupported-by-evidence, 1×verified |
| W05-E01 | 4 | **contradictory** | 1×contradictory, 3×missing |
| W05-E02 | 3 | **contradictory** | 1×contradictory, 2×missing |
| W05-E03 | 2 | **contradictory** | 1×contradictory, 1×missing |
| W05-E04 | 2 | **implemented-incorrectly** | 1×implemented-incorrectly, 1×unsupported-by-evidence |
| W05-E05 | 2 | **contradictory** | 1×contradictory, 1×insufficiently-tested |
| W06-E01 | 2 | **unsupported-by-evidence** | 1×insufficiently-tested, 1×unsupported-by-evidence |
| W06-E02 | 3 | **unsupported-by-evidence** | 2×unsupported-by-evidence, 1×verified |
| W06-E03 | 3 | **verified** | 3×verified |
| W06-E04 | 2 | **unsupported-by-evidence** | 1×unsupported-by-evidence, 1×verified |
| W07-E01 | 4 | **implemented-incorrectly** | 1×implemented-incomplete, 1×implemented-incorrectly, 1×unsupported-by-evidence, 1×verified |
| W07-E02 | 2 | **blocked** | 1×blocked, 1×verified |
| W07-E03 | 1 | **blocked** | 1×blocked |
| W07-E04 | 2 | **missing** | 2×missing |

**Task-level scope disclosure:** per-task verdicts (~370 tasks) are not enumerated in this document. Tasks were verified through their parent story's acceptance criteria and artifacts; material task-level exceptions are called out individually in §5 (e.g. the `status: todo` independent-review tasks behind C-3/C-4, W03-E01-S002's pending T003). The full per-story records including material task verdicts and commands run are retained in the review workspace (`scratchpad/autopsy/verification/*.json`), which is session-local and not part of this repo.

## 15. Consolidated closure plan and Fable's final judgement

**Closure plan:** execute R-1→R-10 in order; R-1 (truth reconciliation) is the gate for everything else — no further acceptance activity should occur against a ledger known to be unreliable. Estimated critical path: R-1 → R-3 → R-7 → R-8 → R-9.

**Fable's final judgement.** This programme contains more real engineering than its worst findings suggest and less completion than its own commit history claims. The code that exists is largely good; several verifications (lease/fencing, audit chain, release gating, online migration) survived adversarial review cleanly. What failed is the control plane: statuses were advanced without the reviews the programme itself mandates, evidence records were left unfilled or mis-pinned yet cited as proof, a quality gate was quietly lowered in the same commit that declared victory, and material engineering (FBL-01) happened entirely outside the ledger. Under the programme's own Definition of Done and evidence policy, **the completion claim of `e8cda6b` is rejected**. The path back is short and mechanical for governance (R-1, R-3, R-6) and well-bounded for code (R-2, R-5); until W05–W07 genuinely close and the final gate runs, no production-readiness claim should be made.

— Fable 5, senior reviewer and final quality gate, 2026-07-16

---

## 16. Remediation addendum (2026-07-16, post-autopsy)

Findings were remediated the same day under Fable supervision (3 waves: code fixes → ledger truth-reconciliation + evidence re-pin → independent reviews + conductor adjudication): 27 of 28 in the first pass; L-3 was initially dropped, caught by the remediation's own closing review gate, and fixed in the same session. Two structural limitations are disclosed rather than claimed away: (a) the "independent" reviews were performed by AI review agents dispatched same-day by the same conductor that supervised the fixes — evidence quality is command-backed and reproducible, but organizational independence is limited, and human ratification of the acceptance decisions is recommended; (b) DEC-PROG-001 (interim coverage floor) is `proposed` and awaits human ratification, so H-1 is not fully closed. Disposition:

| Finding | Disposition |
|---|---|
| C-1 | **Fixed.** `foundation/webhook` rewritten to staged claim(tx)→effect(no-tx)→finalize(lease-fenced) mirroring notify; revert-sensitive `tx_boundary_test.go` proven red-then-green; reviewed and re-accepted (W04-E02-S001). |
| C-2 | **Fixed.** W04-E02-S002 honestly reverted to `planned`; review confirmed nothing partially claims it. The inbound two-phase work remains genuine future work. |
| C-3 | **Fixed.** SEC-03 review genuinely executed 2026-07-16 (5/5 tamper matrix after H-9 fix); story now legitimately `accepted` with conductor ratification of the AC-03 verifier-contract judgment. Caveat disclosed in the review record: the wowsociety cross-repo re-confirmation was not independently re-run — accepted as recorded. |
| C-4 | **Fixed.** All W02 independent reviews executed for real (8/8 stories re-verified, task files filled, compliant records); W02 wave now legitimately `accepted`. |
| C-5 | **Fixed structurally.** `impl/tracking/status-register.md` is now generated by `miscellaneous/regen_status_register.py` from canonical front matter; all false/contradictory statuses reconciled; off-vocabulary tokens normalized (one remaining: W05-E04-S001 `ready-for-review`, W05 scope). |
| H-1 | **Recorded + partially recovered — open pending human act.** DEV-PROG-001 + DEC-PROG-001 (ratchet plan, floor restoration to 90 targeted; decision status `proposed`, human ratification required); kernel/tracing + kernel/safety 0%→100%; coverage 84.5%. |
| H-2 | **Recorded.** DEV-PROG-004 (commit message immutable; compensating controls = autopsy + reconciliation). |
| H-3 | **Fixed.** `Verifier.Actor` fails closed for non-Assurance stores; revert-sensitive test; doc-comment contract updated; reviewed under W03-E01-S001. |
| H-4 | **Fixed.** W00/W01/W06 gates re-run 2026-07-16 with compliant records (`review-gate-2026-07-16.md`); 24 W01 evidence records' reviewer fields closed via addenda. |
| H-5 | **Fixed.** SEC-06 genuinely independently reviewed; acceptance re-based on that review. |
| H-6/H-7 | **Recorded.** DEV-PROG-002/003 + DEC-PROG-002 (AR-01/02 + SEC-04 disposition deferred to W05 execution owner with reopen triggers). Implementation remains W05 programme work, out of remediation scope by design. |
| H-8 | **Fixed.** All 27 W07-E01 evidence items (incl. previously-unswept S003/S004) re-pinned at HEAD with addenda; zero result divergences — squash artifact confirmed, no fabrication. |
| H-9 | **Fixed.** Key-ID + signature-version tamper cases added (5/5 matrix). |
| L-3 | **Fixed (second pass).** Initially omitted from this addendum — caught by the closing review gate. The seven kernel comment-vocabulary leaks (kernel/privileged/relationships.go ×4, kernel/i18n/catalog.go, kernel/i18n/negotiate.go, kernel/httpx/ratelimit.go) genericized; comment-only, no behavior change. |
| H-10, M-6, M-7, M-10, L-1, L-2 | **Fixed.** GOALS-TRACKER/SRS reconciled to the programme; AR-03/SEC-05 READMEs; CHANGELOG populated; W03 artifact paths corrected; isolation test formatted + tracked; gitignore hardened. |
| M-1..M-5, M-8..M-10 | **Fixed or recorded** per review-gate records (E05 compliant review report; W04 closures filled; golden-consumer M-3 resolved as audit-session infra mismatch — `make golden-consumer` passes incl. upgrade-replay; 8→9 edge count corrected; AR-06 count in W05 scope-deferred). |

**Post-remediation ledger (script-generated):** 49 accepted · 5 verified · 1 implemented · 15 planned (W05 + W04-E02-S002 + W07-E04) · 4 blocked (2 human decisions, 2 W05-dependent) · 1 ready-for-review (W05). Waves: W00/W01/**W02** accepted (W02 newly, legitimately) · W03/W04/W06 in-progress (honest) · W05 planned · W07 in-progress.

**Still open (not findings — remaining programme work and human gates):** W05 execution (incl. AR-01/02 wiring decision DEC-PROG-002, SEC-04), W04-E02-S002 inbound two-phase, W06 blocked legs, DEC-Q1/Q9/Q10 human decisions, W03-E01-S003 product-security sign-off, W03-E01-S004 cross-repo sign-off, W01-E01-S003 CI-run evidence (TD-005), coverage restoration to 90 (DEC-PROG-001), programme closure gate (W07-E04).

— Fable 5, remediation supervision record, 2026-07-16
