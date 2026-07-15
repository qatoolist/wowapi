---
id: DEV-W01-E04-S002
type: deviations-record
parent_story: W01-E04-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations — W01-E04-S002

Per mandate §2.6: differences between the approved plan and the actual implementation are recorded
here, not painted over. All entries recorded 2026-07-13 against revision
`0a31186cada5c275a588c74081cf977adf346e61` (HEAD advanced mid-story to `05dce5c8`, impl/-only
delta — carry-forward rationale in `evidence/index.md`).

## DEV-01 — T001 touched one summary sentence beyond the single §6 row

Task-001 described the correction as "a single-row correction"; its own "Detailed work" step 4 and
"Risks" section anticipated that the row's real structure might need more, prescribing "recorded as
a deviation, not silently expanded." At HEAD, §6 closes with a summary sentence (plan line 774)
that enumerates the executed findings ("8 findings (…) have a real, independently-reviewed partial
closure") and counts the rest ("The other 30 remain PLANNED only"). Correcting the DX-05 row alone
would have created a fresh §6-internal contradiction (a 9th EXECUTED row under an "8 findings"
banner). Resolution: the row was corrected as planned AND the summary sentence's enumeration/counts
were updated (8→9 incl. DX-05, 30→29). Mechanical consequence, not a judgment call; recorded here
and flagged to Main.

## DEV-02 — T002 corrected one CLI-behavior prose claim, not just command examples

Blueprint-11 §3 step 6 claimed `wowapi init` seeds `configs/{base,local,dev,stage,prod}.yaml`;
a fresh scaffold at HEAD emits only `{base,local}.yaml`. This is a CLI-reality claim in the same
document, verified by execution, and leaving it wrong while fixing the command lines around it
would contradict the story's desired state ("no example left silently wrong"). Corrected; recorded
because a strict reading of T3 ("CLI examples") might not cover prose claims. Decision-table row 19.

## DEV-03 — Out-of-scope generator defects observed and routed, not fixed

Verifying the corrected examples in a fresh scaffold exposed two code-level defects:
(a) scaffolded `configs/*.yaml` carry `i18n.default_locale`/`i18n.go_bundles`/`i18n.locales_dir`
keys the framework config schema rejects, so `wowapi config validate --env local` fails on a
pristine scaffold — re-verified against a binary built from a clean `git archive HEAD` extraction
(rules out working-tree contamination from sibling workers' in-flight edits); (b) the scaffolded
`go.mod`'s invalid `+dirty` pseudo-version (known DX-01/SF-7) breaks the config-checker delegation
and `lint boundaries` inside the scaffold. Routing outcome (IRC, 2026-07-13): (a) escalated to
Main, ratified, and scope-added by Main to W01-E04-S001 — since FIXED there per its owner
(commented-example i18n block in `base.yaml.tmpl`; pristine-scaffold `config validate --env local`
now passes the framework-only path, new `TestInitScaffoldConfigValidates`, EV-W01-E04-S001-005;
the tools/configcheck delegation leg still fails until DX-01/T001 lands); (b) is W01-E04-S001's
DX-01/T001 — planned there, still open. No production code touched by this story (its
constraint); no doc text asserts these flows work at HEAD.

## DEV-04 — deferred-items-register.md not edited (interpretation recorded)

Task-002 T5 says its output "feeds" `impl/tracking/deferred-items-register.md`. That register's
own charter is "items explicitly deferred OUTSIDE this implementation programme" and it is a
DERIVED VIEW (canonical source = requirement-inventory disposition rows). DX-05 T5 is deferred
*inside* the programme (to `W06-E02-S002..S003` via REL-03), so it does not meet the register's
inclusion rule and the register is not this story's file to derive into. The deferral is recorded
in `artifacts/dx05-t5-deferral-note.md` + `story.md` out-of-scope; if the conductor prefers a
register row anyway, the note contains every §11.10 field needed to derive one.

## DEV-05 — blueprint init example documents HEAD flags, not the planned version-pin flag

The stale example `wowapi init … --wowapi-version vX.Y.Z` was corrected to the HEAD flag surface
(`--module` only) rather than to a future flag: the S001 owner confirmed over IRC (2026-07-13)
that DX-01's version flags are NOT landing in S001's current W01 slice. The version-pin intent is
not lost — it is the subject of `artifacts/dx05-t4-version-gate-design-note.md` and DX-01's own
plan; when DX-01's flags land, the flag-bearing example belongs in that change, keeping blueprint
examples executable at every commit.

Resolved 2026-07-13: DX-01 flags landed same wave (W01GenDX01); conductor added flag documentation to blueprint-11 init example.

## Ratification

DEV-01, DEV-03, and DEV-04 were explicitly ratified by the conductor (Main) over IRC on
2026-07-13: DEV-01 approved as the non-contradiction-preserving read; DEV-04 approved (no register
row; T5 stays tracked at W06/REL-03 via the deferral-note artifact); DEV-03's finding accepted and
reassigned by Main to W01-E04-S001 as a scope-add deviation. DEV-02 and DEV-05 are
scope-boundary/truthfulness clarifications within this story's own authority, recorded for the
reviewer.
