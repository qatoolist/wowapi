---
id: DEV-W01-E01-S002
type: deviations-record
parent_story: W01-E01-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W01-E01-S002

## DEV-W01-E01-S002-001 — fresh-run counts drift vs the cited snapshot

gosec 111/26-non-test (not "38"); errorlint 3→4 non-test; exhaustive 4 (2 new `reflect.Kind`
switches in kernel/config); forcetypeassert 3 (+httpclient); usestdlibvars 5→9;
wrapcheck/revive 464/231 (not "~50 each"). All enumerated in
`evidence/static-analysis/{judged-set-enumeration.txt,phase2-fail-before-enumeration.txt}` and every
non-test hit dispositioned in `implementation.md`. Direction of every adjudication unchanged.

## DEV-W01-E01-S002-002 — G120 no longer reproduces (fixed by W01-E03, not this story)

The cited `kernel/httpx/csrf.go:118` G120 hit does not appear in the Phase-2 definitive gosec re-run
(routed to this story by W01Http): W01Http's W01-E03 http-hardening work fixed the unbounded form
parse mid-wave, before this story's enablement run. Recorded per the story's own out-of-scope rule
("a cited hit that no longer reproduces … is recorded as a finding in deviations.md, not silently
treated as this story's own completed work").

## DEV-W01-E01-S002-003 — nilerr was never actually enabled

`story.md` describes nilerr as an "already-enabled analyzer"; the committed `.golangci.yml` at HEAD
never enabled it (default: standard + depguard/misspell/unconvert/unparam). The story's operative
scope (annotate the policy.go non-finding; do not enable) was executed as written; the annotation
carries a `//nolint:nilerr` marker so it also survives any future enablement. Current-state claim
drift recorded here, not edited into story.md.

## DEV-W01-E01-S002-004 — test-file exclusions for gosec/forcetypeassert/errorlint

The plan implied per-hit triage of everything; 84–85 gosec, 9 forcetypeassert, and 4 errorlint hits
in `_test.go` files were dispositioned as a documented, config-level exclusion class (consistent
with the committed config's existing errcheck/unparam test rule) instead of per-site annotations.
Rationale inline in `.golangci.yml`; non-test code including shipped `testkit` remains fully
covered (its G101/noctx hits were individually dispositioned).

## Note (below deviation threshold)

G703 (path-traversal taint) surfaced at `benchbudget/main.go:92` only after its G304 twin was
suppressed — both rules cover the same tool-only file read; the annotation names both
(`#nosec G304 G703`).
