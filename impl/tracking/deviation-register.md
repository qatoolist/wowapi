---
id: TRACK-DEVIATION-REGISTER
type: register
title: Deviation register — plan-versus-actual deviations
status: active
created_at: 2026-07-12
updated_at: 2026-07-13
derived: true
---

# Deviation register

DERIVED VIEW. Per mandate §2.6 and §8.9: the approved implementation plan must not be rewritten
after implementation to make it appear the final implementation always matched the plan —
differences go in this register instead. Rows are sourced from each story's own `deviations.md`
(see `governance/templates/deviation-template.md`).

## Record format

Future deviation entries use the following fields, mirroring
`governance/templates/deviation-template.md`:

| Deviation ID | Story/Task | Approved plan | Actual implementation | Reason | Impact | Risks | Approval | Compensating controls | Follow-up work |
|---|---|---|---|---|---|---|---|---|---|
| DEV-W00-E01-S001-001 | W00-E01-S001/T003 | `go test ./kernel/... -run TestKernelRules -race` | Equivalent `-run 'TestIntegrationRulesResolverOrgAncestry'` (planned pattern matches no tests at HEAD) | Test name drift vs plan-time citation | None — same assertion exercised | none | conductor 2026-07-13 | Exact test name recorded in evidence | none |
| DEV-W00-E01-S001-002 | W00-E01-S001/T004 (AC-04) | `grep -rn "RunAPI\|RunWorker\|RunMigrate" README.md docs/blueprint/` returns zero hits | 7 hits found in docs/blueprint/ (04, 06, 10, 12); README + blueprint 11 clean; Context-method diff empty | Hits are future-state prose pre-existing at fix commit 345e4ce — AR-05 T1's executed scope was README + blueprint 11 only; no regression | AC-04 adjudicated: executed-slice re-pin PASSES on its actual scope; failed evidence record preserved per policy | Phantom-API prose could mislead readers until W06 | conductor 2026-07-13 (adjudication: re-scope AC-04 to executed T1 scope) | Failed grep evidence preserved (EV-W00-E01-S001-04) | Route the 7 hits to AR-05 T5 at W06-E04-S002 (added to that story's dispatch brief) |
| DEV-W00-E01-S002-001 | W00-E01-S002 | Quiet-machine bench run implied | Bench run under concurrent sibling load; exit-code AC unaffected (10x headroom, 43/43 pass) | Parallel wave execution | ns/op figures not reusable as quiet baseline | none | conductor 2026-07-13 | Load noted in evidence env fields; serialized window vs W00-E02-S001 | none |
| DEV-W00-E01-S002-002 | W00-E01-S002/T002 | "Remove one budgeted entry + make bench-budget" fail-first check | Ghost-entry-in-scratch-file via the tool's stdin-pipe contract | Removing an entry would relax, not trigger, the gate; mutating tracked file prohibited in W00 | Fail-first property proven more strictly | none | conductor 2026-07-13 | Scratch file preserved as artifact | none |
| DEV-W00-E02-S001-001 | W00-E02-S001/T002 | Enable the "25 queried analyzers" named by MATRIX CS-23 | 18 verbatim-recoverable analyzer names enabled; gap flagged, nothing invented | MATRIX CS-23 claims 25 but names only 18 | Baseline covers the recoverable set; drift facts carried into W01-E01 briefs | Residual 7-analyzer ambiguity | conductor 2026-07-13 | Throwaway config preserved as artifact; per-analyzer drift table registered | W01-E01 (FBL-05/07) re-derives the full set from MATRIX text |
| DEV-W00-E02-S003-001 | W00-E02-S003 | AC text says ADR status literal "accepted" | ADRs use status "ratified" per decision-template.md vocabulary | Template/register vocabulary wins over AC prose | None — semantics identical | none | conductor 2026-07-13 | Deviation recorded in story deviations.md | none |
| DEV-W01-E04-S001-04 | W01-E04-S001/T001 | No-flags default derives a VCS pseudo-version; defect cited as `devel` → `v0.0.0` fallback | Default path extended to handle Go 1.24+ stamped-version shapes (stamped-shape scope extension) | Go 1.24+ buildinfo stamps versions the plan's framing did not anticipate | Fail-closed default covers stamped shapes too | none | approved — conductor + W01ReviewGate 2026-07-13 | Recorded in story deviations.md (DEV-04) with fail-first evidence | none |
| DEV-W01-E02-S002-CONDUCTOR-01 | W01-E02-S002 | Story compatibility note implied scaffold templates wire the query tracer | Reviewer found init templates did not wire `WithQueryTracer`; conductor wired `database.WithQueryTracer(tracer)` into both init templates (api runtime pool; worker tracer block moved above pools) | Template-wiring gap missed at story verification | Gap closed pre-acceptance; internal/cli tests ok 28.9s | none | conductor 2026-07-13 (review-gate fix) | Re-ran internal/cli tests; note in story verification.md | none |
| DEV-W01-E04-S003-ADDENDUM | W01-E04-S003 | Evidence pinned at SHA 0a31186 with unmodified e2e harness | internal/e2e/e2e_test.go modified post-evidence (--local-framework flag added to integrate DX-01 fail-closed init) | Harness wiring needed for DX-01 integration, not a timing/DB change | Diagnosis conclusions unaffected; fresh runs PASS (10.5s, 13.3s) + reviewer re-run PASS (11.4s) | none | conductor 2026-07-13 (carry-forward addendum) | Addendum recorded in story evidence/index.md + deviations.md | none |
| DEV-W01-E01-S002-004 | W01-E01-S002 | Per-hit triage of every gosec/forcetypeassert/errorlint hit implied | `_test.go` hits (84–85 gosec, 9 forcetypeassert, 4 errorlint) dispositioned as a documented config-level exclusion class | Test-file hit volume; consistent exclusion class judged safer than blanket annotations | Exclusion class documented, not silent | Residual: test-file hits unexamined per-site | conductor + W01ReviewGate 2026-07-13 (documented) | Exclusion rationale recorded in story deviations.md | none |
| DEV-W01-E01-S001-001/002 | W01-E01-S001/T002 | noctx fail-before/pass-after against the 2 named exec.Command sites | noctx v2.11.4 does not report exec.Command sites (flags net/http only; 146 hits, 145 in `_test.go`); fail-before evidence substituted via gosec G204; sites fixed to CommandContext | noctx detection gap vs plan-time citation | AC-02 proven by substituted mechanism (gosec G204 + code diff); noctx run exit 0 | none | conductor + W01ReviewGate 2026-07-13 | Substitution recorded in evidence index (EV-002) + deviations.md | none |
