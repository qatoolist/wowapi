---
id: W00-E02-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W00-E02-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E02-S003 — Evidence index

Per mandate §10. The expected evidence for this story is the independent-review fidelity-check
record for each ADR — confirming (a) completeness against `decision-template.md`'s required
sections and (b) fidelity against the exact REVIEW §F/§U source line, per `verification.md`'s
planned procedure. Evidence type throughout: **review report**.

Per this repository's Adaptation 2 (`naming-conventions.md`), no `reviews/` subdirectory is
pre-created — it is created on first real evidence content. The reviewer may produce either nine
per-ADR review reports or one consolidated report covering all nine; either is acceptable as long
as every acceptance criterion's evidence identifier is traceable to which ADR(s) it covers. The
table below assumes nine per-ADR entries as the default expectation; if a consolidated report is
produced instead, this table is updated to reflect a single evidence ID covering all nine, cross-
referenced from each.

The independent review pass executed 2026-07-13 as ONE consolidated report covering all nine
ADRs (permitted above), so all nine per-ADR rows share the same File/URI; EV-010 is the scripted
structural/index cross-check log that complements it.

| Evidence ID | Evidence type | Story and task | Acceptance criteria proven | Execution command | Commit SHA | Branch/tag | Environment | Tool versions | Date/time | Result | File/URI | Reviewer | Status |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EV-W00-E02-S003-001 | review report | W00-E02-S003 / T001 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-001-framework-owns-grant-authority.md` vs REVIEW §F row 2 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-002 | review report | W00-E02-S003 / T001 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-002-single-registrar-typed-keys.md` vs REVIEW §F row 3 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-003 | review report | W00-E02-S003 / T001 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-003-post-seal-mutation-error-not-panic.md` vs REVIEW §F row 4 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-004 | review report | W00-E02-S003 / T002 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-004-audit-hash-version-column.md` vs REVIEW §F row 5 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-005 | review report | W00-E02-S003 / T002 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-005-goreleaser-skip-publish-split.md` vs REVIEW §F row 6 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-006 | review report | W00-E02-S003 / T002 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-006-authz-epoch-table-not-message-bus.md` vs REVIEW §F row 7 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-007 | review report | W00-E02-S003 / T002 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-007-jwks-trusted-issuer-config-gate.md` vs REVIEW §F row 8 (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-008 | review report | W00-E02-S003 / T003 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-008-pgx-query-tracer-not-otelpgx.md` vs REVIEW §U (+ MATRIX CS-05) (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-009 | review report | W00-E02-S003 / T003 | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | manual line-by-line review of `decisions/adr-009-secrets-boot-time-rotation-contract.md` vs REVIEW §U (+ MATRIX CS-25) (consolidated report) | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0); concurrent sibling W00 workers (non-timing review) | not applicable (manual review) | 2026-07-13 | pass | [reviews/adr-fidelity-review-2026-07-13.md](reviews/adr-fidelity-review-2026-07-13.md) | W00-E02-S003 execution worker + reviewer subagent (independent of 2026-07-12 authoring pass) | pass |
| EV-W00-E02-S003-010 | execution log (scripted structure + index cross-check) | W00-E02-S003 / T001–T003 | AC-W00-E02-S003-01, AC-W00-E02-S003-02 | python (eval kernel) scan of all nine ADRs vs `decision-template.md` sections/front matter + `decisions/index.md` row-by-row front-matter match — script description in the log header | 0a31186cada5c275a588c74081cf977adf346e61 | main (story files uncommitted working-tree additions) | local checkout, Darwin arm64 (macOS 25.5.0), python3 eval kernel | python 3 (eval kernel) | 2026-07-13 | pass | [logs/adr-structure-check-2026-07-13.log](logs/adr-structure-check-2026-07-13.log) | W00-E02-S003 execution worker | pass |

The AC-W00-E02-S003-02 cross-check (`decisions/index.md` vs each ADR's front matter) is covered
twice: scripted in EV-W00-E02-S003-010 and manually within the consolidated review report.

Superseded/failed evidence, if any review finds a fidelity gap, is preserved per
`impl/governance/evidence-policy.md` — never deleted, marked `failed` and retained alongside the
`retested` record once the gap is fixed.
