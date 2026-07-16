---
id: W04-CLOSURE
type: wave-closure-report
wave: W04
status: in-progress
created_at: 2026-07-12
updated_at: 2026-07-16
---

# W04 — Closure report (interim — wave not closed)

**Correction note (autopsy remediation R-1, 2026-07-16):** frontmatter previously read
`status: accepted` while this file's own body was the unexecuted pre-execution template
("Wave 04 has not begun execution") — a direct self-contradiction (autopsy finding **C-5**,
`impl/reports/implementation-autopsy-report-2026-07-16.md`). Wave 04 has, in fact, partially
executed since the template was written. Frontmatter is corrected to `in-progress`
(`governance/status-model.md` §7.1: "at least one contained epic/story is actively being worked").
This body replaces the unexecuted template with an honest interim state summary. **This wave is
NOT closed** — no acceptance-criteria table, epic-completion table, or reviewer conclusion below
should be read as a closure claim.

## Interim state summary (2026-07-16)

- **W04-E01 (lease/fencing primitive and jobs)** and **W04-E03 (bulk multi-worker safety)** —
  verified real by the autopsy: named chaos tests exist and pass against the real Postgres
  instance (`kernel/jobs/chaos/duplicate_worker_lease_expiry_test.go`,
  `foundation/bulk/chaos/duplicate_worker_test.go`); migrations 00038/00044 confirmed. E01-S001/
  S002 closure.md bodies still carried unfilled template text despite `accepted` story.md
  front matter (autopsy finding M-2) — corrected separately in this remediation pass.
- **W04-E02 (remote I/O outside tx)** — mixed:
  - S001 (notify/webhook three-stage protocol): the notify path (`SendPending`) genuinely
    implements claim/effect/finalize staging outside the transaction. The webhook path does not —
    **confirmed code defect (autopsy C-1)**: `foundation/webhook/service.go` performs secret
    resolution and the HTTP POST inside an open DB transaction on both the dispatch and retry
    paths, contradicting the story's own accepted acceptance criteria. This defect was remediated
    2026-07-16 in the working tree; independent re-review of the fix is being scheduled.
  - S002 (inbound two-phase verification, adapter contracts, 6-boundary chaos test): **false
    acceptance reverted** — `story.md`/`closure.md` claimed `accepted` while the closure body
    itself already said "not implemented, verified, or closed" (autopsy C-2). Both files corrected
    to `planned` in this remediation pass; none of T4/T5/T6/T8's scope has actually started.
  - S003 (retry adoption): verified real by the autopsy (cenkalti/backoff/v5 dependency,
    kernel/retry wrapper, foundation/notify/service.go integration).
- **W04-E04 (compliance and readiness)** — partial: S001 (audit hash_version discriminator)
  verified real; S002 (DSR-hold path) honestly labeled `closed-pending-review`
  (implemented-incomplete per the autopsy); S003 (readiness checks) not independently verified
  this session (unsupported-by-evidence).

## Acceptance-criteria completion (interim, not a closure claim)

| AC | Status | Notes |
|---|---|---|
| AC-W04-01 | in-progress | E01 verified; see summary above. |
| AC-W04-02 | in-progress | E02: S001 defect remediated 2026-07-16, re-review pending; S002 reverted to planned. |
| AC-W04-03 | in-progress | E03 verified real. |
| AC-W04-04 | in-progress | E04 partial — S001 verified, S002/S003 incomplete/unsupported. |
| AC-W04-05 | not started | Not independently confirmed this session. |

## Epic completion (interim, not a closure claim)

| Epic | Status |
|---|---|
| W04-E01 | implemented-incomplete (autopsy worst-of-stories roll-up) |
| W04-E02 | implemented-incorrectly (autopsy worst-of-stories roll-up; C-1 remediated 2026-07-16, S002 reverted to planned) |
| W04-E03 | verified (autopsy) |
| W04-E04 | unsupported-by-evidence (autopsy worst-of-stories roll-up) |

## Reviewer conclusion

Not reviewed. No independent review gate has been executed for Wave 04. This interim report is
not a substitute for that review.

## Acceptance authority

Data/reliability lead (role-based, not yet exercised).

## Closure date

Not closed.

## Final status

`in-progress` — interim state as of 2026-07-16, per autopsy remediation R-1. Wave 04 closure
(the actual `W07-E04`-style closure gate for this wave) has not been executed; do not treat any
status above as an acceptance claim.

Conductor note (2026-07-16): `review-gate-2026-07-16.md` has since run and accepted E01 (all 3
stories), E02-S001/S003, E03 (both stories), and E04 (all 3 stories); E02-S002 remains genuinely
`planned`. Wave remains `in-progress` pending E02-S002.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
