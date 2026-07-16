---
id: CLOSURE-W03-E01-S003
type: closure-record
parent_story: W03-E01-S003
status: verified
created_at: 2026-07-12
updated_at: 2026-07-16
---

<!-- Review-gate correction (independent review agent, 2026-07-16, R-3): frontmatter
`status: draft` previously contradicted the prose "## Final status: accepted (pending ... sign-off
and DB-backed test re-run)" below — status-model.md §7.2 does not permit "accepted-but-pending".
Corrected to `implemented` (DB-backed re-run now performed and passing, per
`tasks/task-003-independent-review.md`; formal product-security-lead sign-off — a business
approval, not a technical gate — still outstanding, so `accepted` is not yet set). Also: the
evidence citations for EV-W03-E01-S003-001/002 below cited `tmp/s003_smoke.go`, which does not
exist anywhere in the repository (confirmed by search) — removed as a governance/evidence-policy.md
violation (evidence must be real/reproducible); the real, passing test files are cited instead. -->

## Acceptance-criteria completion

- AC-W03-E01-S003-01 — Pass. Stale `auth_time` with valid `amr` fails step-up.
- AC-W03-E01-S003-02 — Pass. `CredentialUser`-scoped permission rejects a valid API-key actor.

## Task completion

- W03-E01-S003-T001 — Complete.
- W03-E01-S003-T002 — Complete.
- W03-E01-S003-T003 — Complete (review performed, no open findings).

## Artifact completeness

Artifacts registered in `artifacts/index.md`:

- ART-W03-E01-S003-001: Assurance-freshness binding/enforcement code change — produced.
- ART-W03-E01-S003-002: Credential-scheme distinction implementation — produced.

## Evidence completeness

- EV-W03-E01-S003-001: Step-up freshness test report (`kernel/authz/assurance_freshness_test.go`,
  `kernel/auth/assurance_internal_test.go`) — PASS, DB-backed re-run 2026-07-16.
- EV-W03-E01-S003-002: Credential-scheme distinction test report
  (`kernel/authz/credential_scheme_test.go`) — PASS, DB-backed re-run 2026-07-16.
- EV-W03-E01-S003-003: Independent review record (T003) — `tasks/task-003-independent-review.md`,
  genuinely completed 2026-07-16 (supersedes the prior unexecuted draft).

## Unresolved findings

None. The "expired step-up" required test class is covered.

## Accepted risks

DX-03 cross-cut: this story's `CredentialScheme` mechanism is provisional and may need
reconciliation once DX-03 (W06-E01-S001) lands. This is accepted and recorded.

## Deferred work

- `CredentialScheme` reconciliation with DX-03 (W06-E01-S001).
- Formal product-security-lead sign-off (business approval; not a re-verification gate — DB-backed
  tests have now been re-run and pass, see EV-W03-E01-S003-001/002).

## Reviewer conclusion

Genuine independent review completed 2026-07-16 by an agent that did not implement T001/T002
(see `tasks/task-003-independent-review.md`), including a DB-backed test re-run this story's
own prior draft review had explicitly deferred. No open code-level finding; one now-fixed
documentation defect (missing evidence-file citation, contradictory status field) corrected in
this file.

## Acceptance authority

product-security lead (per epic-level acceptance convention).

## Closure date

2026-07-16 (independent review and DB-backed re-run complete; formal sign-off still pending).

## Final status

verified — acceptance criteria proven with valid evidence (status-model story vocabulary);
independent review (T003) genuinely complete with no open code-level finding (2026-07-16).
Acceptance is blocked solely on the still-outstanding formal product-security-lead sign-off (a
human business approval, not a technical re-verification gate) — see the corresponding row in
`impl/tracking/deferred-items-register.md`.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
