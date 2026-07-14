---
id: W02-E05-RISKS
type: epic-risks
epic: W02-E05
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E05 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W02-004 originates
at wave scope and is reproduced/elaborated here because it lands entirely within this epic's single
story. One further epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W02-004 | The catalog-manifest-format design investigation (S001-T001) may conclude that the seed-sync path requires a new dependency, a new table, or a design pattern not anticipated by this wave's planning documents — MATRIX CS-21 fixes the acceptance bar but explicitly defers the design detail ("design detail to be ratified in Phase 5") | Medium | Low-medium — a larger-than-expected design surface would expand S001's scope beyond its currently-estimated 6-task count, but the design-investigation-before-implementation structure exists specifically to absorb this without corrupting the acceptance bar itself | Low-medium | W02-E05-S001 | T001 is sequenced first and its own task record requires a documented decision + rationale for every named design question before any implementation task begins; if the decisions change the implementation task breakdown, that change is recorded as a plan revision (not silently rewritten as if it were always the plan) | If the investigation surfaces a need for new infrastructure beyond this story's bounded scope (mandate §12), split the excess into a follow-up story rather than silently expanding S001 past a reasonably reviewable size; if a decision is D-0N-caliber, escalate for ADR treatment per `epic.md`'s process safeguard | unassigned | open | Low — the design-investigation-first structure is the mitigation, and its own existence is this wave's acknowledgment of the risk |
| RISK-W02-E05-002 | The "RLS-respecting" requirement is in tension with the bootstrap context: seed-sync populates the very catalogs (roles, permissions, policies) that RLS-governed access presupposes, on a database where those catalogs are empty. A naive resolution — run the sync as a superuser/platform role that bypasses RLS entirely — would satisfy the mechanics while violating the requirement's intent; an over-strict resolution — require full RLS enforcement against empty catalogs — may make bootstrap impossible. The correct role posture (e.g. a dedicated seed role with narrowly-scoped grants, RLS `FORCE` semantics preserved on tenant tables, catalog tables distinguished from tenant-scoped tables) is a genuine design question the source does not answer | Medium | Medium — a wrong posture either reopens the platform-role-bypass class of concern DATA-01 T3's own risk note flags for its mismatch audit ("Requires a platform-role connection to bypass RLS for the scan"), or blocks the feature | Medium | W02-E05-S001 (T001 primarily; T002's implementation inherits the decision) | The RLS posture is an explicitly-named question in S001's `plan.md` "Unresolved questions" and T001's detailed work — it must receive a documented decision + rationale, including which role the sync runs as and why that does not undermine tenancy controls, before T002 implements it | If no posture satisfies both bootstrap feasibility and the RLS-respecting intent, escalate to the acceptance authority (data/reliability lead) with the trade-off documented rather than silently picking one side | unassigned | open | Low-medium once T001's documented decision lands and independent review (T006) specifically checks it |

## Residual risk after mitigation

RISK-W02-004 is expected to reduce to low residual risk once T001's investigation completes within
anticipated scope (or its contingency is exercised with a recorded split). RISK-W02-E05-002 reduces
to low-medium once the RLS posture is decided with documented rationale and independently reviewed —
some residual risk remains inherent in any privileged bootstrap path, which is exactly why the audit
record and dry-run mode are part of CS-21's own fix sketch rather than optional extras.
