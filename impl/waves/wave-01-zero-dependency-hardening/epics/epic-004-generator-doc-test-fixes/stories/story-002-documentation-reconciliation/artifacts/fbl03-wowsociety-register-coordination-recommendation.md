---
id: ART-W01-E04-S002-005
type: artifact
title: FBL-03 — PROD-level coordination recommendation, wowsociety upstream finding register
parent_story: W01-E04-S002
producing_task: W01-E04-S002-T003
source_requirement: FBL-03
status: produced
created_at: 2026-07-13
---

# FBL-03 — wowsociety upstream register coordination recommendation

**PROD-level coordination note** (pattern: `impl/analysis/requirement-inventory.md` §D PROD-01..05).
This document recommends edits to files in the **wowsociety repository**, which this repository
does not own (mandate §2.3 framework/product boundary); no wowsociety file is edited by wowapi's
programme. A wowsociety maintainer (or a future cross-repository task) can apply it verbatim.

**Target register (confirmed by read-only inspection, 2026-07-13):**
`wowsociety/docs/upstream/` — index in `README.md` (per-finding table, lines 10–24), one file per
finding. All findings were recorded against wowapi `v1.0.0 / 287abc3`.

| Finding | Register file | Current register status | Recommended change | Contingency |
|---|---|---|---|---|
| PF-2 | `06-pf-2-gen-crud-emits-out-of-set-verb.md` (+ README index row 06) | Open — no `**Status:**` line | Add `**Status:** RESOLVED upstream in wowapi <commit>` naming the wowapi commit that ships the DX-02 fix (generator emits the in-set `deactivate` verb + generator-output-boots test), and annotate the README index row accordingly | **Do NOT apply until wowapi W01-E04-S001's DX-02 task has landed in a commit and been accepted.** Status snapshot 2026-07-13 (per the S001 owner): the fix is verified fail-first in S001's working tree (EV-W01-E04-S001-003/-004) but not yet committed — the conductor commits at wave close, so no shipping SHA exists yet; the register entry accurately describes wowapi HEAD `05dce5c8` behavior. The `<commit>` placeholder is filled with the wave-close commit SHA. |
| PF-6 | `01-pf-6-step-up-seedability.md` (+ README index row 01) | Entry body already carries `**Status:** RESOLVED upstream in wowapi d2a4164` — but the README index row 01 and the "PF-6 is prioritized" posting instruction (README lines 33–36) still present it as an open, prioritized item | Reconcile the index/posting-instructions with the entry's own RESOLVED status: mark index row 01 resolved and drop (or past-tense) the "PF-6 is prioritized" posting note | None — per REVIEW Answer 18 ("no active workarounds remain… mark the 2 stale upstream docs resolved"), taken as given per this epic's governing instructions; the entry's own RESOLVED header corroborates it. |
| RFF-001 | `03-rff-001-production-object-storage-adapter.md` (+ README index row 03) | Open — no `**Status:**` line; still described as "High (prod-blocking)" in the README ("PF-9 and RFF-001 are prod-blocking") | Add `**Status:** RESOLVED upstream` — wowapi now ships `adapters/storage/s3` implementing `storage.Adapter` (present at wowapi HEAD `05dce5c8`; exercised by 20 S3 tests, CI-wired per plan §9 REL-04 record). Mark index row 03 resolved and remove RFF-001 from the "prod-blocking" posting note (PF-9's own status is out of FBL-03's named scope) | None — second of REVIEW Answer 18's "2 stale upstream docs". |

## Explicit non-scope

- Other register entries (PF-1, PF-3, PF-4, PF-7..PF-11, SF-7, R5b) are NOT covered by FBL-03's
  named scope ("PF-2/PF-6/RFF-001"); the inventory note's "etc." is resolved conservatively — no
  recommendation is made for entries whose fixes have not been verified landed.
- This story's closure does not include verifying that wowsociety applies these edits
  (RISK-W01-E04-003, permanently-accepted residual; tracked at programme level if never applied).
