# EV-W01-E03-S001-005 — CSRF oversized-form-body rejection (fail-first pair)

- **Evidence ID**: EV-W01-E03-S001-005
- **Evidence type**: unit-test report (fail-first pair)
- **Story / task**: W01-E03-S001 / W01-E03-S001-T003
- **Acceptance criteria proven**: AC-W01-E03-S001-04 (functional half)
- **Execution command**: `go test ./kernel/httpx/ -run 'TestCSRF' -count=1 -v`
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: pre-fix, `TestCSRFOversizedFormBodyRejected` FAILED — a >1 MiB form body was fully buffered and its token (placed beyond the bound) was found, proving the unbounded read. Post-fix the full CSRF suite passes, including the new bound test, the MaxFormBytes override test, and the pre-existing in-bound form-field test (regression guarantee: in-bound requests unchanged).

## Pre-fix run (status: failed → resolved) — captured at 0a31186cada5c275a588c74081cf977adf346e61 before the csrf.go change

```
=== RUN   TestCSRFOversizedFormBodyRejected
    csrf_test.go:236: an oversized form body must never reach the handler via the CSRF form fallback
--- FAIL: TestCSRFOversizedFormBodyRejected (0.00s)
FAIL
FAIL	github.com/qatoolist/wowapi/kernel/httpx	0.333s
FAIL
```

## Post-fix run (status: passed, full CSRF suite)

```
=== RUN   TestCSRFSafeMethodIssuesCookie
--- PASS: TestCSRFSafeMethodIssuesCookie (0.00s)
=== RUN   TestCSRFSafeMethodReusesExistingCookie
--- PASS: TestCSRFSafeMethodReusesExistingCookie (0.00s)
=== RUN   TestCSRFUnsafeMethodRejectsMissingToken
--- PASS: TestCSRFUnsafeMethodRejectsMissingToken (0.00s)
=== RUN   TestCSRFUnsafeMethodRejectsMismatchedToken
--- PASS: TestCSRFUnsafeMethodRejectsMismatchedToken (0.00s)
=== RUN   TestCSRFUnsafeMethodPassesWithValidToken
--- PASS: TestCSRFUnsafeMethodPassesWithValidToken (0.00s)
=== RUN   TestCSRFUnsafeMethodRejectsEmptyCookieValue
--- PASS: TestCSRFUnsafeMethodRejectsEmptyCookieValue (0.00s)
=== RUN   TestCSRFUnsafeMethodNoFormFallbackConfigured
--- PASS: TestCSRFUnsafeMethodNoFormFallbackConfigured (0.00s)
=== RUN   TestCSRFUnsafeMethodAcceptsFormField
--- PASS: TestCSRFUnsafeMethodAcceptsFormField (0.00s)
=== RUN   TestCSRFOversizedFormBodyRejected
--- PASS: TestCSRFOversizedFormBodyRejected (0.00s)
=== RUN   TestCSRFCustomMaxFormBytesOverridesDefault
--- PASS: TestCSRFCustomMaxFormBytesOverridesDefault (0.00s)
=== RUN   TestCSRFSafeMethodsExemptFromTokenCheck
--- PASS: TestCSRFSafeMethodsExemptFromTokenCheck (0.00s)
=== RUN   TestCSRFCookieSameSiteVariants
--- PASS: TestCSRFCookieSameSiteVariants (0.00s)
ok  	github.com/qatoolist/wowapi/kernel/httpx	0.243s
```

## Reviewer completion addendum — 2026-07-16

**Reviewer**: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3).
**Review date**: 2026-07-16.
**Commit revision reviewed against**: HEAD 43b6e12 + remediation working tree 2026-07-16.
**Disposition**: Verified (existence + autopsy corroboration). Same disposition as ev-001 in this story; MaxBytesReader enforcement is this record's specific claim and is directly supported by the csrf.go grep evidence.

This addendum retroactively fills the evidence-policy-mandated "reviewer" field. The original
record above (including any "Pending — conductor acceptance gate" line) is left unmodified per
the failed-evidence preservation convention — this is an appended addendum, not a rewrite.
