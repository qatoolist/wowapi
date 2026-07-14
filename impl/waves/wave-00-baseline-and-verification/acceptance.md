---
id: W00-ACCEPTANCE
type: wave-acceptance
wave: W00
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00 — Wave-level acceptance

Wave 00 is accepted when **all** of the following hold. Each condition traces to a story; the wave
cannot be marked `accepted` on the strength of "all tasks complete" alone (mandate §7) — every
condition below requires actual evidence, reviewed.

**Satisfaction record (2026-07-13):** all seven conditions below are **satisfied** — see
`closure-report.md` for the per-AC evidence mapping. Independent review gate passed 2026-07-13
(reviewer W00ReviewGate; conductor concurs). AC-W00-01's AR-05 clause satisfied on the executed
T1/T2 scope per conductor adjudication DEV-W00-E01-S001-002.

## AC-W00-01 — Executed finding-slices re-verified

For each of SEC-02 (T1-T3), PERF-01, PERF-06 (T1), DATA-08 (W0-T1/W0-T2), AR-04 (T1), AR-05
(T1/T2), AR-06 (T1), REL-04 (T1-T4): the exact test file(s)/command(s) named in PLAN §5/§8/§9 have
been re-run at the wave's closing commit SHA, the result is `pass`, and an evidence record exists
per mandate §10 (evidence ID, command, commit SHA, environment, date, result, reviewer). Traces to
W00-E01-S001, W00-E01-S002, W00-E01-S003.

## AC-W00-02 — Verify-outcome rows re-pinned

CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze + key-echo fallback), CS-24 (SSRF
dial-time guard) each have an evidence pointer confirming the citations in their MATRIX spec still
hold at the wave's closing commit. Traces to W00-E01-S003.

## AC-W00-03 — Quality baselines captured

Coverage %, full-tree lint state (25-analyzer hit counts per MATRIX CS-23's inventory), bench-budget
state (post-#25 recalibration, 43 entries), and CI wall-clock per leg are captured as registered
evidence artifacts, dated and commit-pinned. Traces to W00-E02-S001.

## AC-W00-04 — Dependency and toolchain inventory captured

go.mod direct/indirect dependency list and pinned tool versions (golangci-lint, GoReleaser, Trivy,
etc.) are captured and cross-checked against REVIEW §L's approved-dependency register with zero
unexplained drift. Traces to W00-E02-S002.

## AC-W00-05 — D-01..D-09 ratified as ADRs

Nine ADR files exist, one per decision D-01 through D-09, each stating recommendation, safe default
(where applicable), and owner, sourced from REVIEW §F/§U, registered in the decision register.
Traces to W00-E02-S003.

## AC-W00-06 — No unresolved regression

If AC-W00-01 or AC-W00-02 surfaces a regression (a slice that no longer passes at current HEAD), a
follow-up task exists to fix it and the affected story does not move to `accepted` until resolved or
the regression is explicitly accepted as a residual risk by the acceptance authority.

## AC-W00-07 — Independent review passed

Every W00 story has passed independent review per mandate §14: implementation (verification, in
this wave's case) matches the approved plan or deviations are documented; evidence references the
correct code revision; no unsupported completion claims are made; no source requirement was silently
dropped.

## Acceptance authority

Framework architecture lead (role-based; see `wave.md`).
