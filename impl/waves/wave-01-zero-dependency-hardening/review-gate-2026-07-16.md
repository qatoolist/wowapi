---
id: W01-REVIEW-GATE-2026-07-16
type: review-report
wave: W01
status: done
created_at: 2026-07-16
derived: false
---

# Wave 01 (zero-dependency-hardening) — independent review gate re-run, 2026-07-16

**Reviewer**: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5
conductor (autopsy remediation R-3).
**Scope**: autopsy finding H-4 — every epic-002 evidence record's `Reviewer` field read
"Pending — conductor acceptance gate", and `epic-002/acceptance.md`'s AC status table read
"not started" for all four ACs, despite the wave's 2026-07-13 closure report asserting an
independent review ("W01ReviewGate") had passed and all 10 stories were accepted.
**Commit revision reviewed against**: `HEAD 43b6e128672f0b0997adcebc92703884deba5684` +
uncommitted remediation working tree, 2026-07-16 (repo not pushed; see caveats below).
**Inputs used**: `scratchpad/autopsy/verification/wave-01-zero-dependency-hardening.json` (prior
adversarial per-story verification, commands + reasoning recorded 2026-07-13/-16), spot-check
re-runs of decisive commands against the current tree (below), direct inspection of the affected
evidence files and `epic-002/acceptance.md`.

## What was found (confirming the autopsy)

1. **H-4 confirmed, wave-wide, not epic-002-only.** A systematic scan of every `evidence/` file
   under `impl/waves/wave-01-zero-dependency-hardening/` for a `reviewer` field containing
   "pending" turned up **24 evidence records across all four epics** (E01-S003, E02-S001, E02-S002,
   E03-S001, E03-S002, E04-S001, E04-S003), not only the 6 in epic-002 the autopsy sampled. The
   wave's evidence-bundle convention differs by epic (some use markdown `## Reviewer` sections,
   some use a JSON `"reviewer"` key), but the defect is the same everywhere it appears: the field
   was left as a template placeholder, never populated with an actual reviewer identity, date, or
   sign-off note, in violation of `impl/governance/evidence-policy.md`'s "all fields mandatory"
   rule.
2. **`epic-002/acceptance.md`'s AC table said "not started" for all 4 ACs** directly above a
   narrative "Acceptance record" claiming "Satisfied... independent review passed (W01ReviewGate)"
   with no linked artifact. This contradiction is now corrected (see Remediation below).
3. **E01-S003's CI-execution gap is still open**, not resolved by remediation. Re-checked
   2026-07-16: the wave (including `closure-report.md` itself) remains an uncommitted/unpushed
   working tree, so the `go mod verify` CI step and Trivy license scanner added by this story have
   **never actually executed in CI** — only a local run + actionlint syntax check exist as
   evidence (`gomodverify-and-actionlint.log`). This is the same gap the autopsy flagged on
   2026-07-13, correctly self-disclosed in the story's own `evidence/index.md` ("produced
   (superseded-by-CI-run planned as retested after conductor push)") and in
   `closure-report.md`'s Open items, so it is not a new defect — but it is not fixed either, and
   AC-W01-03's CI-execution leg remains unproven pending an actual push.
4. **E04-S002's claimed artifact (`internal/cli/cli.go`) is confirmed to be an extraction
   registration error**, not a real defect. Direct inspection of
   `epics/epic-004-generator-doc-test-fixes/stories/story-002-documentation-reconciliation/`
   shows the real artifacts are all doc files (`artifacts/dx05-t3-cli-example-decision-table.md`,
   `dx05-t4-version-gate-design-note.md`, `dx05-t5-deferral-note.md`,
   `fbl03-wowsociety-register-coordination-recommendation.md`) plus review evidence
   (`evidence/reviews/ev-001-t001-plan-doc.diff`, `ev-002-t002-blueprint11.diff`,
   `ev-002-command-log.md`), matching the story's documentation-reconciliation subject matter
   exactly. `internal/cli/cli.go` was never the real artifact; the extraction tool mis-registered
   it. This does not affect the story's acceptance.

## Decisive command re-runs (spot-check per instructions — DB up, targeted `go test -count=1`)

| Story | Command | Result |
|---|---|---|
| W01-E02-S001 | `go test ./kernel/logging/... -run 'TestLogRecordInsideActiveSpanCarriesTraceAndSpanIDs\|TestLogRecordWithoutSpanOmitsCorrelationKeys' -v -count=1` | PASS (both tests + subtests), `ok kernel/logging 0.615s` |
| W01-E02-S002 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/database/... -run TestIntegrationQueryTracerChildSpanInTraceTree -v -count=1` | PASS, `ok kernel/database 0.278s`. (Without `DATABASE_URL`/`WOWAPI_REQUIRE_DB=1` this test SKIPs — that is the autopsy's own noted tooling gap, now resolved.) |
| W01-E04-S001 | `go test ./internal/cli/... -run 'TestGenCRUDOutputBoots\|TestGenCRUDMutatingRoutesDeclareContractsAndUseValidatedHandler' -v -count=1` | PASS (1.86s, <0.01s) |
| W01-E01-S001, S002, E03-S001, E03-S002, E04-S003 | not re-run (autopsy's file-existence + config-grep checks were already decisive and were not disturbed by remediation; reused per instructions) | verified by reuse |

## Per-story recommendation

| Story | Autopsy verdict | This gate's recommendation | Basis |
|---|---|---|---|
| W01-E01-S001 | verified | **accept** | Linter config present, evidence bundle complete; reused autopsy verification (no story-level reviewer-field gap found in this epic's convention). |
| W01-E01-S002 | verified | **accept** | Judged linter set confirmed in config; reused autopsy verification. |
| W01-E01-S003 | implemented-incomplete | **accept-with-conditions** | Local run + actionlint evidence is real and sufficient for AC-…-01/02/04; AC-W01-03's CI-execution leg is still genuinely unproven as of 2026-07-16 (wave unpushed) — condition: register an actual CI-run evidence record (status `retested`, referencing `EV-W01-E01-S003-001`/`-002` as superseded) before this AC is called `met` rather than carried forward. |
| W01-E02-S001 | verified (reviewer-field gap) | **accept** | Reviewer-field gap now closed via addenda on `EV-W01-E02-S001-001/002/003.md` and `evidence/index.md`; decisive test re-run PASS against current tree. |
| W01-E02-S002 | verified (reviewer-field gap) | **accept** | Reviewer-field gap now closed via addenda on `EV-W01-E02-S002-001/002/003.md` and `evidence/index.md`; decisive real-DB integration test re-run PASS, resolving the autopsy's own verification-tooling gap. |
| W01-E03-S001 | verified | **accept** | Evidence bundle complete and reviewer-field addenda applied to all 5 records; reused autopsy file/config checks. |
| W01-E03-S002 | verified | **accept** | `EnforceRouteContracts` wiring confirmed; deviations.md "None" independently confirmed accurate; reviewer-field addenda applied to all 4 records. |
| W01-E04-S001 | verified | **accept** | Two safety-critical generator tests re-run PASS on current tree; reviewer-field addenda applied to all 5 DX-01/DX-02 records. |
| W01-E04-S002 | unsupported-by-evidence (artifact mismatch) | **accept** | Investigated and resolved: the flagged `internal/cli/cli.go` claimed-artifact is confirmed a registration error, not a real defect — actual artifacts (doc-reconciliation files) exist, match the story's subject matter, and are unaffected. No code-claim to substantiate; nothing to reject. |
| W01-E04-S003 | verified | **accept** | Diagnosis/decision evidence bundle confirmed present with claimed shape; reviewer-field addendum applied. Flaky e2e suite itself intentionally not re-run (diagnosis record, not a code change). |

## Wave-level recommendation

**Accept-with-conditions.** Nine of ten stories are unconditionally accepted on this re-review.
One story (W01-E01-S003) carries a still-open, self-disclosed condition (CI-execution evidence for
`go mod verify`/license scanning is real-but-local-only, pending an actual CI run once the wave is
pushed) that was already correctly flagged as a carry-forward item and does not, on its own,
justify rejecting the story — but it must not be silently dropped at the next wave gate.

The root defect this re-review was scoped to fix — H-4, unfilled `Reviewer` fields making every
"accepted" claim in this wave technically unsubstantiated per `evidence-policy.md` — is now
**closed**: 24 evidence records across all 4 epics received signed, dated reviewer-completion
addenda (append-only, originals preserved per the failed-evidence preservation rule), and
`epic-002/acceptance.md`'s AC table now reflects the real, re-verified status instead of
"not started".

## Files touched by this review

Reviewer-completion addenda appended (originals untouched) to 24 evidence records:
- `epics/epic-001-static-analysis-utilisation/stories/story-003-supply-chain-and-hooks/evidence/index.md`
- `epics/epic-002-observability-correlation/stories/story-001-trace-log-correlation/evidence/{index.md, tests/EV-W01-E02-S001-001.md, tests/EV-W01-E02-S001-002.md, benchmarks/EV-W01-E02-S001-003.md}`
- `epics/epic-002-observability-correlation/stories/story-002-pgx-query-tracer/evidence/{index.md, tests/EV-W01-E02-S002-001.md, tests/EV-W01-E02-S002-002.md, tests/EV-W01-E02-S002-003.md}`
- `epics/epic-003-http-hardening/stories/story-001-server-timeouts-and-body-bounds/evidence/{static-analysis/ev-004-gosec-scoped.md, tests/ev-001-template-render-fail-first.md, tests/ev-002-config-defaults.md, tests/ev-003-prod-zero-rejection-fail-first.md, tests/ev-005-csrf-oversized-body-fail-first.md}`
- `epics/epic-003-http-hardening/stories/story-002-central-validation-enforcement/evidence/tests/{ev-001-boot-rejection-fail-first.md, ev-002-adversarial-400-field-errors.md, ev-003-waiver-exemption.md, ev-004-crud-template-migration.md}`
- `epics/epic-004-generator-doc-test-fixes/stories/story-001-generator-correctness/evidence/{DX-01/t1-flag-verify.json, DX-01/t5-e2e-temp-dir.json, DX-02/w0-t2-boots-test.json, DX-02/w0-t2-verb-fix.json, DX-02/scaffold-config-validate-fix.json}`
- `epics/epic-004-generator-doc-test-fixes/stories/story-003-e2e-flake-diagnosis/evidence/premier/T-TEST-01/reproduction-runs.md`

`epic-002-observability-correlation/acceptance.md` updated: AC status table changed from
"not started" (×4) to `met`/`met (spot-checked, not re-run)` with citations to the re-run
commands and results; original 2026-07-13 acceptance narrative preserved verbatim with a
historical-accuracy note appended, plus a new 2026-07-16 acceptance record.

## Unresolved / carried forward

- W01-E01-S003 AC-W01-03's CI-execution leg (see conditions above) — track at next wave gate.
- AC-W01-E02-02 (allocation-neutrality benchmark) was spot-checked by code inspection, not
  re-executed; recommend a fresh benchmark run at the next full quality gate rather than treating
  this review's addendum as new benchmark evidence.
- Line-level re-verification of every named fix site in W01-E01-S002 (gosec/errorlint/exhaustive
  annotations) was not repeated in this pass — reused the autopsy's config-level verification, per
  the "reuse prior verification, spot-check only decisive commands" instruction for this gate.
