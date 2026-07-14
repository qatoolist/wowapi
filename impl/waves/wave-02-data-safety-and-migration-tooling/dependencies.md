---
id: W02-DEPS
type: wave-dependencies
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02 — Dependencies

## Upstream (waves this wave depends on)

- **W00** — full wave dependency per the strict W00→W07 entry ordering and per `impl/index.md`'s
  wave map row for W02 ("Depends on: W00"). W02 consumes W00's baseline lint/coverage/CI state as
  its starting point. No W02 story depends on a specific D-0N ADR from W00-E02-S003 — confirmed by
  scanning `requirement-inventory.md` §B for any D-0N row targeting DATA-09, DATA-01, DATA-05,
  DATA-06, or FBL-02; none exists (see `wave.md` "Assumptions" for the full negative-confirmation
  list).

## Downstream (waves that depend on this wave)

| Downstream item | Depends on (from W02) | Why |
|---|---|---|
| W03-E01-S001 (SEC-01 grant-table migration) | W02-E01 (DATA-09 protocol) | `impl/index.md`'s wave map states W03 depends on "W02 (grant-table migration uses DATA-09)" — SEC-01's new `identity_grant` table migration is expected to be built and rolled out using the expand/backfill/validate/contract protocol this wave establishes, not a bespoke one-off migration. |
| W04-E01-S001 (DATA-02 shared lease/fencing primitive) | W02-E01-S002 (minimal checkpoint lease) | W04-E01-S001's own scope note (`impl/analysis/wave-allocation-detail.md`) states it "replaces W02-E01-S002's minimal checkpoint lease; migration note" — W04 does not merely depend on W02 existing, it is scoped to specifically retire W02's interim lease implementation. This is the wave-level mirror of the E01-S002 deviation-risk recorded in `risks.md`. |
| W04-E04-S001 (DATA-08 W6-T1 audit hash widening) | W02-E01 (DATA-09 protocol) | `impl/analysis/wave-allocation-detail.md`'s W04-E04-S001 row states "dep W02-E01 protocol" — the audit hash-chain widening migration (a breaking format change touching wowsociety's live audit rows) is expected to ship via DATA-09's protocol, not ad hoc. |

## Cross-wave dependencies

None beyond the W00→W02 entry dependency and the two downstream consumers (W03, W04) stated above.
W02 does not depend on W01, W03, W04, W05, W06, or W07.

## Internal (within this wave, between epics)

- **W02-E02 depends on W02-E01.** DATA-01's riskiest steps (T4: add composite FK `NOT VALID`; T5:
  `VALIDATE CONSTRAINT`) are explicitly sequenced after DATA-09's protocol exists, per PLAN's own
  PF-DATA cross-cutting note (6): "sequence DATA-09 T1-T5 ahead of DATA-01 T4/T5 ... in the real
  release plan." `impl/analysis/wave-allocation-detail.md` states this exactly at story grain: "S002
  audit-fk-validate-negatives (T3, T4, T5, T7, T8 — T4/T5 gated on E01 S001/S002 acceptance)." This
  wave records that gate as a `depends_on` entry on W02-E02-S002's `story.md` front matter
  (`W02-E01-S001`, `W02-E01-S002`), not merely as prose — see
  `epics/epic-002-tenant-fk-integrity/dependencies.md` for the epic-scoped statement and
  `epics/epic-002-tenant-fk-integrity/stories/story-002-audit-fk-validate-negatives/story.md` for
  the story-scoped one.
- **W02-E03, W02-E04, W02-E05 are independent of W02-E01, W02-E02, and of each other.** DATA-05
  (version allocation, blob GC), DATA-06 (aggregate write contract), and FBL-02 (seed-sync) each
  target disjoint packages (`kernel/artifact`/`kernel/document`; `kernel/resource`; a new seed-sync
  command path respectively) and have no task-level dependency on DATA-09's protocol or DATA-01's
  FK work in the source tables (`requirement-inventory.md` rows DATA-05/DATA-06/FBL-02 cite no
  DATA-09/DATA-01 dependency). They may execute in any order or in parallel with E01/E02 and with
  each other.

## External dependencies

None new. DATA-09's tooling is built on the existing pgx/PostgreSQL toolchain already in use
elsewhere in the framework (no new external service). FBL-02's seed-sync path consumes the existing
catalog/seed data model; no new external dependency is introduced by the design-investigation task
as currently scoped — if the investigation concludes a new dependency is required, that is recorded
as a deviation, not silently added.

## Repository dependencies

None cross-repo for the epics' core framework work. wowsociety impact is real but non-blocking for
this wave's own closure:

- **DATA-01** — PLAN's own wowsociety-impact note: `policy_override.rule_version_id`
  (`internal/modules/policy/migrations/00002_override.sql:16`) is "a genuine independent instance of
  the DATA-01 pattern," requiring wowapi's `UNIQUE (tenant_id, id)` on `rule_versions` first (this
  wave's T1), then wowsociety's own migration via wowapi's DATA-09 protocol once it exists. This is
  tracked as `PROD-01` in `requirement-inventory.md` §D — product-level, excluded from this wave's
  framework-side closure per mandate §2.3.
- **DATA-09** — wowsociety's `docs/DEPLOY.md` documents a single-shot "migrate fully, then deploy
  everyone" model with no canary/soak window; PLAN's evidence states wowsociety should "adopt
  whatever manifest schema wowapi's tooling consumes" once it exists, not build its own. No wowapi-
  side blocking dependency; tracked for wowsociety's own future adoption, not required for this
  wave's closure.
- **DATA-08 W6-T1's product exposure** (audit hash widening, W04 scope, dependent on this wave's
  DATA-09 protocol per the downstream table above) is real and material for wowsociety's live audit
  rows, per PLAN's own note — tracked as `PROD-05` in `requirement-inventory.md` §D, staging-drill
  coordination, not this wave's concern directly.

## Tooling dependencies

None new. DATA-09's CI drill pipeline (T9) extends the existing CI infrastructure (`.github/
workflows/ci.yml` and related) rather than introducing a new CI system.

## Decision dependencies

None. Confirmed: no D-0N ADR targets any W02 finding (DATA-09, DATA-01, DATA-05, DATA-06, FBL-02).
See `wave.md` "Assumptions" for the full confirmation.
