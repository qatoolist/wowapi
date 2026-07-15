---
id: W04-E04-DEPS
type: epic-dependencies
epic: W04-E04
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W02-E01** (the DATA-09 online-migration protocol) — narrowly, for **W04-E04-S001 only**. Per
  `../../wave.md`'s entry criteria: "W02's exit gate satisfied for W02-E01 specifically... this is
  the one concrete predecessor capability this wave's own stories require, and only for
  W04-E04-S001 (DATA-08 W6-T1's audit-hash migration)." Confirmed at wave scope by
  `impl/waves/wave-02-data-safety-and-migration-tooling/dependencies.md`'s downstream table: "W04-
  E04-S001 (DATA-08 W6-T1 audit hash widening) | W02-E01 (DATA-09 protocol) | ... the audit hash-
  chain widening migration (a breaking format change touching wowsociety's live audit rows) is
  expected to ship via DATA-09's protocol, not ad hoc." This is a whole-epic reference (W02-E01, not
  a specific sub-story of it) because W02's own downstream table names the dependency at that grain.
- **W04-E01, W04-E02, W04-E03** (within this wave) — no dependency. This epic's stories share no
  task-level dependency with the wave's other three epics; grouped into the same wave by shared
  requirement-source (`DATA-08`, `DX-07`) and shared wave-level acceptance authority
  (data/reliability lead), not by a build-order dependency.
- **W00** (baseline) — this epic depends on W00's exit gate at wave scope, per the programme's
  strict W00→W07 entry ordering (`impl/index.md`).

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| PROD-05 (product-level, wowsociety staging audit re-verification) | W04-E04-S001 (DATA-08 W6-T1) | `requirement-inventory.md` §D: "DATA-08 W6 staging audit re-verification before version bump | Product compliance drill | hash_version branch verification (D-04)." wowsociety must run a dedicated staging verification pass, confirming its live audit rows still verify under the new version-branched scheme, before `FRAMEWORK_VERSION` is bumped past S001's commit. Product-level, tracked here as a downstream coordination item, not implemented by this epic. |
| PROD-03 (product-level, wowsociety readiness backport) | W04-E04-S003 (DX-07 T1) | `requirement-inventory.md` §D: "wowsociety readiness/timeout backports to committed main.go | DX-07 T1 + FBL-09 fix the templates." wowsociety's own already-generated `cmd/api/main.go:240-243` shows the identical two-check readiness gap; a manual backport of the migration-currency check is a recommended, non-blocking follow-up once T1's template pattern is established. |
| W05-E03-S002 (AR-04 T5 waiver mechanism) | — (inverse relationship) | DX-07 T4, explicitly out of this epic's S003 scope, depends on W05-E03-S002's not-yet-built waiver mechanism — recorded here for completeness, mirroring the wave-level inverse dependency in `../../dependencies.md`. This epic does not implement T4; it only records the forward reference. |

## Internal (within this epic)

S001 → S002 form a two-phase pipeline: S002 (DATA-08 W6-T2 through T5 — external anchoring,
encrypted DSR export, central legal-hold wrapper, explicit per-class status) depends on S001 (W6-T1's
widened, versioned hash chain) per `wave.md`'s own dependency framing ("DATA-08 W6-T2 through W6-T5
(S002) depend on W6-T1 (S001)"). S003 (DX-07 T1-T3) has **no** dependency on S001 or S002 — it is an
independent readiness/diagnostics concern grouped into this epic by MATRIX CS-21's shared closure-
spec framing with FBL-02 (W02 scope, not this epic), not by a task dependency, per
`../../dependencies.md` (wave-level): "DX-07 (S003) has no dependency on S001/S002."

## Cross-wave dependencies

The W02-E01 → W04-E04-S001 edge described above. No other cross-wave dependency.

## External dependencies

None new. S001's migration ships through W02-E01's existing online-migration tooling; S002's DSR
export artifact introduces a new encryption-key-management dependency (per PLAN DATA-08 W6-T3's own
risk note), tracked as an implementation-time design decision in S002's own `plan.md`, not a
programme-level external dependency addition.

## Repository dependencies

wowsociety impact is real and material for S001 specifically — not for S002 or S003's current code
surface. Per `wave.md`'s wowsociety-impact note (DATA-08): "wowsociety produces real, live audit
rows today (`identity/service.go`, `policy/service.go`, `impersonation.go`'s grant/revoke writes,
`cmd/api/main.go`'s API-key audit wiring)." No `kernel/attachment`/`kernel/notify`/`kernel/retention`
usage was found in wowsociety at the time PLAN's evidence was gathered, so S002's W6-T3/T4/T5 land on
wowsociety's future DSR roadmap, not its current code — non-blocking for this epic's own closure.
S003's DX-07 impact is confirmed but explicitly non-breaking (template-only fix; wowsociety's already
-committed `main.go` is untouched by T1 landing).

## Tooling dependencies

None beyond the already-available Go/PostgreSQL toolchain and W02-E01's existing online-migration
tooling, which S001 consumes rather than duplicates.

## Decision dependencies

D-04 (already ratified) — enacted by S001 only. No other story in this epic depends on a D-0N
decision. See `epic.md` "Required decisions."
