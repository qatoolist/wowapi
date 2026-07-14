---
id: W03-E04-RISKS
type: epic-risks
epic: W03-E04
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E04 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W03-003
originates at wave scope and lands entirely within this epic's single story. One further
epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W03-003 | T4's cache-invalidation acceptance criterion depends on W05-E04-S002 (SEC-04 epoch table, D-06) landing — if W05 is delayed relative to W03, this epic either ships with an incomplete AC or must artificially wait on a later wave | Medium | Medium | Medium | W03-E04-S001 | T4's cache-invalidation AC is deferred-linked rather than blocking this epic's overall acceptance — T1/T2/T4's non-cache-invalidation portions can close independently | Close with the cache-invalidation AC explicitly marked deferred-linked (not silently dropped) if W05-E04-S002 has not landed by the time this epic is otherwise ready | unassigned | open | Low-medium once the deferred-link framing is honored |
| RISK-W03-E04-002 | This epic's implementation is started before W03-E01 reaches `accepted`, producing rework once SEC-01's actual (post-review) principal-model shape differs from what T1 assumed during premature implementation | Low (process risk, not a defect risk) | High — if the hard dependency is not honored, T1's actor-resolution logic could need a substantial rewrite | Medium | W03-E04-S001 | This epic's `progress.md` and `dependencies.md` both record the hard blocking dependency explicitly; the story's own Definition of Ready should include "W03-E01 accepted" as an entry condition, not merely "W03-E01 started" | If discovered mid-implementation that W03-E01 has not yet reached `accepted`, pause this epic's work rather than continuing against a moving target | unassigned | open | Low, given the dependency is explicitly and repeatedly documented across this epic's files |

## Residual risk after mitigation

RISK-W03-003 reduces to Low-medium once the deferred-link framing is honored. RISK-W03-E04-002
reduces to Low provided the documented hard-dependency gate is actually respected during execution,
which is a process discipline this epic's own files make explicit at every level (epic, progress,
dependencies).
