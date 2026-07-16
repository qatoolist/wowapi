---
id: W03-REVIEW-GATE-2026-07-16
type: review-gate-report
wave: W03
status: complete
created_at: 2026-07-16
updated_at: 2026-07-16
---

# Wave 03 — Independent review gate (2026-07-16)

Executed per autopsy remediation R-3 (conductor dispatch), closing the gap the autopsy identified
under C-3/H-5: **W03-E03-S001 (SEC-03) was closed with its independent review never executed, and
W03-E02-S001 (SEC-06) was accepted on a self-review rather than an independent one.** Statuses
across the wave were honestly reverted from `accepted` to `implemented`/`ready` ahead of this gate
(2026-07-16, prior remediation pass). This report executes the review those reverts were waiting
on.

**Reviewer identity for every record produced by this gate**: Independent review agent (Claude
Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3). This reviewer
did not implement any of the code under review.

**Commit basis**: HEAD `43b6e128672f0b0997adcebc92703884deba5684` + remediation working tree
2026-07-16 (uncommitted changes present per `git status`: C-1 webhook out-of-tx staged delivery
fix, H-9 tamper tests in `foundation/webhook/`, H-3 auth fail-closed in `kernel/auth/`, tracing
tests, ledger truth-reconciliation edits).

**Environment**: Local dev; Go per repo `go.mod`; Postgres 16 reachable at
`postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`
(`DATABASE_URL`/`WOWAPI_REQUIRE_DB=1` set for DB-backed suites).

Per-story review records are filed in each story's own `tasks/task-NNN-independent-review.md` and
`closure.md` (see "Where recorded" column below); this document is the wave-level roll-up and
final recommendation. This reviewer recommends only — the conductor sets accepted status.

## Per-story results

### W03-E01-S001 — Grant schema and unconditional membership enforcement (SEC-01 core)

**Recommendation: accept.**

This story already had a completed T004 independent review (`status: complete`,
`closure.md: accepted`) predating this dispatch. This dispatch's scope was to **re-verify the
specific remediation** the autopsy's `implemented-incorrectly` verdict identified: `Verifier.Actor`
silently skipped the tenant-membership check (`ActiveTenantAccess`) whenever the configured
`PrincipalStore` did not implement the additive `AssurancePrincipalStore` interface — a
CS-07-class regression (server trusts unverified tenant state), gated on interface satisfaction
rather than `CapacityID`.

Read `kernel/auth/auth.go:284-359` directly. The fix is present and correct: `Actor` now fails
closed with `KindForbidden` (`"tenant membership verification unavailable"`) whenever
`claims.TenantID != uuid.Nil` and the store does not implement `AssurancePrincipalStore`
(lines 295-300), before any tenant-bound work proceeds. The type's own doc comment
(`auth.go:240-244`) now states this explicitly. Ran the cited regression test plus the full
adversarial suite:

```
go test ./kernel/auth/... -run 'TestActor_BaseOnlyStoreFailsClosedOnTenantClaim|TestActor_ZeroTenantRejected|TestActor_GarbageTenantRejected|TestActor_PrivilegedSessionResolvedFromGrant|TestActor_ForgedGrantIDRejected|TestActor_NoCapacityMultipleCapacitiesRejected|TestActor_ExplicitCapacityValidatedServerSide|TestActor_ExpiredGrantRejected|TestActor_RevokedGrantRejected|TestActor_WrongTenantGrantRejected|TestActor_WrongActorGrantRejected|TestActor_UnauthorizedApproverGrantRejected' -count=1 -v
```
All 12 named tests PASS (`ok github.com/qatoolist/wowapi/kernel/auth 0.855s`), including
`TestActor_BaseOnlyStoreFailsClosedOnTenantClaim`, which did not exist before the remediation and
directly proves a base-only `PrincipalStore` + non-zero `TenantID` now rejects rather than silently
proceeding.

**Process finding (not blocking, recorded for the record)**: the story's original T004 review
(`closure.md`, "Reviewer conclusion: ... no open issues", dated 2026-07-13) did not catch this
gap — the autopsy did, subsequently. This is evidence the earlier review was not adversarial
enough on this specific axis, not evidence of a current defect; the gap it missed is now closed
and independently re-verified under this dispatch. No action required beyond this note, since
T004 already exists and is not in this dispatch's "pending/todo" remediation scope.

**Where recorded**: no new file (T004 already complete); this verification is recorded in this
wave-level report and its command output above.

### W03-E01-S002 — Capacity selection and privileged-session resolver

**Recommendation: accept.**

T003 (independent review) was `status: todo`; now `status: done`. Reviewed
`kernel/auth/auth.go:299-357` and re-ran all 9 named adversarial tests — PASS. Confirmed both ACs
satisfied, all six named grant-rejection conditions independently tested (not collapsed), `Actor`
struct shape preserved, capacity-selection mechanism documented. Noted (not a new finding) that
T4/T5's gates share the same `hasAssurance` type assertion reviewed under E01-S001 above, which is
now fail-closed — no separate exposure specific to this story.

**Where recorded**:
`impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-002-capacity-and-privileged-resolver/tasks/task-003-independent-review.md`.

### W03-E01-S003 — Assurance freshness and credential-scheme distinction

**Recommendation: accept-with-conditions** — fix `closure.md`'s documentation defects (now done by
this reviewer; see below) before the conductor sets `accepted`; formal product-security-lead
sign-off (a business approval, not a technical re-verification) remains genuinely outstanding.

T003 had review prose drafted but never formally executed (no filled Verification Record, no
reviewer identity, `status: todo`, and the draft's own text noted DB-backed tests were skipped).
This dispatch performed the actual review: re-ran all 11 named tests with a live DB — PASS
(`TestStepUpFreshness*` ×5, `TestCredentialScheme*` ×6 including
`TestActorInternal_AssuranceFieldsPropagate`). Both ACs genuinely satisfied.

Two pre-existing documentation defects, corrected in this pass:
1. `closure.md` frontmatter `status: draft` directly contradicted its own prose `## Final status:
   accepted (pending ...)` — an oxymoron per `governance/status-model.md` §7.2. Corrected to
   `implemented`.
2. `closure.md`'s evidence-completeness section cited `tmp/s003_smoke.go` as evidence for
   EV-W03-E01-S003-001/002; confirmed by search that this file does not exist anywhere in the
   repository — a referenced-but-missing evidence artifact (evidence-policy.md violation).
   Removed the citation; the evidence entries now cite only the real, passing test files (which
   `evidence/index.md` — as opposed to `closure.md` — already correctly cited).

**Where recorded**:
`.../story-003-assurance-and-credential-schemes/tasks/task-003-independent-review.md` and
`.../story-003-assurance-and-credential-schemes/closure.md` (both edited in place).

### W03-E01-S004 — Cross-repo cutover plan for the wowsociety impersonation-flow breaking change

**Recommendation: not-ready** (as currently marked `accepted` in `story.md`) —
**accept-with-conditions** if re-scoped to "wowapi-side documentation complete, cross-repo sign-off
pending."

In-repo verification: `sequencing-plan.md` (113 lines), `staging-validation-plan.md` (235 lines),
and `rollback-plan.md` (180 lines) all exist and are substantive (not template stubs) — confirmed
by direct inspection of line counts and structure. `story.md`'s own frontmatter, however, lists
`owner: unassigned` and `reviewer: unassigned` while simultaneously stating `status: accepted` —
an unreviewed/unowned document cannot honestly carry an `accepted` status per this programme's own
definition-of-done convention (every other story in this wave requires an independent-review task
before `accepted`; this story has none in its `tasks/` directory, only T001-T003 authoring tasks).

The wave's own closure condition (per `wave.md`, cited in the prior autopsy verification) requires
these documents be "reviewed and accepted by both a wowapi-side and a wowsociety-side reviewer."
**The wowsociety-side reviewer sign-off is out of this repo's scope to verify** — no wowsociety
repo state is visible from this dispatch, and no artifact in this wowapi repo can prove or disprove
that a wowsociety-side reviewer has signed off. This is recorded as an explicit scope boundary, not
resolved either way. What this reviewer can and does confirm: the wowapi-side content is real,
substantive, and internally consistent with the SEC-01/S001-S003 stories it depends on; what
remains unverifiable from this repo is the cross-repo sign-off the story's own closure condition
requires. Recommend the conductor either (a) obtain and record the wowsociety-side sign-off before
treating this story as `accepted`, or (b) explicitly narrow the story's acceptance criteria to
wowapi-side authorship only and record that narrowing as a deviation.

**Where recorded**: this wave-level report only (no in-repo evidence exists to confirm or deny the
wowsociety-side leg; no task file edited since none of this story's task files are in a
pending/todo review state — the defect is in `story.md`'s own status claim, which this reviewer
does not have authority to overwrite, only to flag for the conductor).

### W03-E02-S001 — Outbound-security escape-hatch governance (SEC-06)

**Recommendation: accept.**

T006 (independent review) was `status: done` but was a **self-review by the implementer**
("Self-review against the independent-review-gate checklist... A separate reviewer (T006) still
needs to ratify the evidence bundle") — the exact defect this dispatch exists to close (autopsy
H-5). A genuine independent review is now performed, superseding the self-review. Re-ran:

```
go test ./kernel/config/... -run 'TestFitnessCheck|TestRecordAllowlistChange|TestEgressExceptions' -count=1
go test ./kernel/auth/... -run TestNewJWKSKeySource -v
```
All pass. Specifically confirmed non-vacuously: `TestFitnessCheckDetectsKnownViolation` proves the
fitness check fires against a deliberately introduced violation (not merely a clean-tree pass), and
`TestNewJWKSKeySource_ProdCustomClientRequiresTrustedIssuers` proves the D-07 gate genuinely fails
closed (`prod` + custom client + no trusted issuers → rejected), not merely logged as a warning.
All 5 ACs satisfied.

**Where recorded**:
`.../story-001-outbound-security-governance/tasks/task-006-independent-review.md` and
`.../story-001-outbound-security-governance/closure.md` (both edited in place, self-review content
retained with a correction note, not deleted, per evidence-policy's failed-evidence preservation
principle).

### W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data (SEC-03)

**Recommendation: accept.**

This is the story the autopsy's C-3 finding centers on: T005 (independent review) was `status:
todo` with every Verification Record field literally "Not yet executed," while `closure.md`
claimed `Final status: accepted` — a direct false-completion claim on a security-critical fix. This
was already honestly reverted to `implemented` ahead of this dispatch. This dispatch performs the
actual T005 review.

Re-ran the full DB-backed suite (`go test ./foundation/webhook/... -run
'TestIntegrationHandleInbound' -count=1 -v`) — 13/13 PASS, including the two 2026-07-16 tamper-matrix
additions (`_TamperedKeyID`, `_TamperedSignatureVersion`) that close autopsy finding H-9 (previously
only 3 of 5 required tamper-matrix fields were independently tested; now 5/5: body, timestamp,
event-ID, key-ID, signature-version, each its own test case).

**Judgment call, made explicit per this dispatch's brief**: the shipped `HMACVerifier` is a
body-only HMAC scheme and does not itself bind key-ID or signature-version into its signature —
`provider-verifier-contract.md` documents this as conformant ("From authenticated data if the
scheme authenticates a key id; otherwise empty"). The two new tamper tests prove the contract using
a test-local `keyedVerifier` (the same technique already used elsewhere in this test file for the
timestamp case), which demonstrates the `Envelope`/`HandleInbound` plumbing correctly rejects a
tampered key-ID/sig-version whenever a `Verifier` implementation does bind them. **This reviewer's
judgment: AC-W03-E03-S001-03 ("No security decision in `HandleInbound` reads a raw `InboundIn`
field; the adversarial tamper matrix ... passes") is satisfied by this approach** — the AC is about
the downstream plumbing never trusting an unauthenticated field, not about mandating that the
specific shipped `HMACVerifier` itself authenticate every one of the five fields. Nothing in
`story.md`'s T2 scope description obligates `HMACVerifier` specifically to bind key-ID/sig-version.
If a stricter reading is intended (shipped verifier must bind all 5 fields itself), that would be a
scope question for the conductor to resolve with the story owner, not a defect in the tests as
written.

Also confirmed: breaking-interface-change documentation present in `story.md` "Compatibility
considerations"; `provider-verifier-contract.md` (T4) accurately reflects the as-built `Envelope`
synthesis and its stated limitation; `closure.md`'s recorded wowsociety-consumer re-confirmation
(RISK-W03-006) is not a bare restatement of PLAN's original snapshot but a specific finding
recorded at this story's own execution — accepted as recorded per the story's evidence trail (not
independently re-run against the wowsociety repo in this pass, out of this repo's scope).

**Non-blocking hygiene note**: several registers (including the wave's own extraction JSON) cite
`kernel/webhook/...` as the artifact path; the real package is `foundation/webhook`. Recommend
correcting on next edit; not a review blocker since the correct package was independently located,
built, and tested.

**Where recorded**:
`.../story-001-webhook-authenticated-replay/tasks/task-005-independent-review.md` and
`.../story-001-webhook-authenticated-replay/closure.md` (both edited in place).

### W03-E04-S001 — Relationship semantics (party-subject evaluation, subject-kind matrix, mutation governance)

**Recommendation: accept.**

T004 was `status: todo`; now `status: done`. Re-ran, with a live DB, all 6 named tests across
`kernel/relationship/...` — PASS, including the 3-subtest `TestIntegrationRelationshipSubjectKindMatrix`
and the fail-closed-default test `TestUnitResolveSubjectUnsupportedKind`. Confirmed: the W03-E01
gate was genuinely honored (code consumes the finalized `Actor` shape without referencing any
field removed/renamed by the S001 remediation); T3's actor-attribution wiring does not
re-implement DATA-06 T2's mechanism; the cache-invalidation sub-criterion is honestly
deferred-linked to W05-E04-S002, not silently dropped.

Also corrected a dispute the prior verification JSON raised: the extraction file had mis-marked
this story's T001-T003 as `todo` when `closure.md` and the code both correctly show them `done` —
confirmed this was an extraction artifact, not a story defect.

**Where recorded**:
`.../story-001-relationship-semantics/tasks/task-004-independent-review.md` and
`.../story-001-relationship-semantics/closure.md` (both edited in place).

### W03-E05-S001 — Workflow privileged completion (ratification decision + durable override audit)

**Recommendation: accept.**

T003 was `status: pending`; now `status: done`. This story's `closure.md` was already the most
honest in the wave — never claimed `accepted`, correctly stated "Awaiting independent review."
Re-ran, with a live DB, all 7 named tests/subtests across `kernel/workflow/...` — PASS. Specifically
verified by reading `kernel/workflow/override_audit_test.go:200-244`
(`TestOverrideAuditFailureRollsBack`) that the fault-injection test is genuinely adversarial: it
mutates `audit.Entry.Metadata` to include a `chan int` value, which is not JSON-serializable and
causes a real failure in `audit.Record`'s canonicalization path (not a mocked/stubbed error), then
asserts by direct DB query that the workflow instance, its open task, and the audit-log table are
all unaffected (zero-effect rollback) — not merely that `Override` returned a non-nil error. T1-T3's
fail-closed behavior was re-confirmed by execution (`TestIntegrationOverrideAuthzGate`,
`TestIntegrationOverrideFailsClosedWithoutPermission`), not assumed unchanged.

**Where recorded**:
`.../story-001-workflow-privileged-completion/tasks/task-003-independent-review.md` and
`.../story-001-workflow-privileged-completion/closure.md` (both edited in place).

## Wave-level recommendation

**Recommendation: accept-with-conditions.**

Seven of eight in-scope stories (E01-S001 through E01-S003, E02-S001, E03-S001, E04-S001, E05-S001)
are recommended `accept` with all cited acceptance criteria independently re-verified against a
live DB and current HEAD + working tree. One story (E01-S004) is recommended `not-ready` in its
current `accepted` form because its own closure condition (bilateral wowapi + wowsociety reviewer
sign-off) cannot be confirmed from this repo and its `story.md` carries `owner: unassigned`/
`reviewer: unassigned` alongside `status: accepted` — a contradiction this reviewer does not have
authority to resolve unilaterally.

Conditions for the conductor before treating Wave 03 itself as closed:
1. E01-S004: obtain/record the wowsociety-side reviewer sign-off, or explicitly narrow the story's
   scope to wowapi-side-only and record that as a deviation.
2. E01-S003: the technical defects this gate found are already corrected in `closure.md`
   (contradictory status, missing evidence file); the formal product-security-lead sign-off
   remains a genuine open item, separate from and not blocking the technical `accept`
   recommendation for the other 7 stories.
3. Wave/epic-level governance documents (`wave.md`, `closure-report.md`,
   `epics/epic-001-.../closure-report.md`) remain frozen at template/`planned` state and were
   **not** in this dispatch's scope to rewrite — the conductor should still reconcile those
   against the now-corrected story-level statuses (this is the Medium-severity finding the prior
   verification pass already flagged and this gate did not re-litigate).

No story in this wave's reviewed set was found to fail its acceptance criteria outright. The two
false-completion patterns the autopsy identified (SEC-03 closed with review never executed;
SEC-06 accepted on self-review) are both closed by this gate with genuine, independently-executed
reviews, each recorded in the affected story's own task and closure files per
`governance/evidence-policy.md`.
