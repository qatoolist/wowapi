---
id: ART-W01-E01-S003-004
type: artifact
title: Nightly fuzz-schedule confirmation note
parent_story: W01-E01-S003
producing_task: W01-E01-S003-T004
status: produced
created_at: 2026-07-13
---

# Nightly fuzz-schedule confirmation (audit note)

Confirmed at commit `0a31186cada5c275a588c74081cf977adf346e61` (HEAD, 2026-07-13), by direct
inspection of the workflow/Makefile chain **and** an observed scheduled run — not by trusting the
`ci.yml` header comments.

## What was inspected (file chain at HEAD)

1. **Trigger exists and is nightly**: `.github/workflows/ci.yml:42-46` — `schedule:` with
   `cron: "17 3 * * *"` (daily, 03:17 UTC).
2. **Schedule reaches the fuzz step (no silent gating)**: on a `schedule` event the `changes` job's
   classifier has no usable diff base (`EVENT_NAME` is neither `pull_request` nor `merge_group`;
   `github.event.before` is empty), so it takes the explicit fail-safe branch (`ci.yml:97-104`)
   emitting `code=true` / `bench=true`. The `gate` job runs on `needs.changes.outputs.code == 'true'`
   (`ci.yml:187`), matrix leg `test` runs `make ci-container-test` (`ci.yml:227`).
3. **The step actually invokes fuzz targets in seed-replay mode**: `Makefile:324-326` —
   `ci-container-test` runs `sh -c 'make test-unit && go test ./kernel/filtering/ ./kernel/pagination/
   -run "^Fuzz" -count=1'` inside the toolbox with `WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1`.
   `-run "^Fuzz"` with **no `-fuzz=` flag** = seed-corpus replay only (in-code `f.Add` seeds:
   `kernel/filtering/fuzz_test.go` `FuzzFilterParse`/`FuzzParseSort`,
   `kernel/pagination/fuzz_test.go` `FuzzDecodeCursor`).

## Observed run (not just inspection)

`gh run list --workflow=ci.yml --event=schedule` shows consecutive daily scheduled runs on `main`,
both `success`: run **29229288699** (started 2026-07-13T06:32:19Z) and **29182363356**
(2026-07-12T06:14:45Z). The 2026-07-13 run's `authoritative gate — test` job log contains the exact
seed-replay invocation and the passing `ok … kernel/filtering` / `ok … kernel/pagination` results.
Full capture: `../evidence/logs/nightly-fuzz-observed-run.log`.

Observation: scheduled runs start ~3 h after the 03:17 UTC cron mark (06:14Z / 06:32Z) — GitHub's
documented best-effort scheduling delay, not a wiring defect; the cadence is genuinely nightly.

## Scope boundary (explicitly restated)

The coverage-guided **`-fuzz=` generation gap remains open** and is **W07 scope**
(REL-04 T8 / PERF-06 T3/T4, owner "PF-REL" per `premier-framework-implementation-plan.md`). This
story confirms the seed-replay schedule only; it neither closes nor duplicates the W07 work. No
`ci.yml` change was needed for this task (the wiring is correct as found).
