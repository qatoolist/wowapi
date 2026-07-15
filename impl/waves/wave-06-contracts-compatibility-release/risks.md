---
id: W06-RISKS
type: wave-risks
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W06-001 | W06-E03-S002 (protection-activation) cannot enter `ready`/`in-progress` until DEC-Q10 (branch protection, protected release Environment, tag protection ruleset) is resolved by a human with repo-admin access — a coding agent structurally cannot perform this action | High (confirmed today: `gh api repos/qatoolist/wowapi/branches/main/protection` → 404, `gh api .../environments` → `{"total_count":0}`) | Medium — does not block REL-01's own buildable-now work (S001, ~85% per REVIEW §G), but the release pipeline's final trust boundary (protected publish) cannot be end-to-end proven until this resolves | Medium | W06-E03-S002 | S001's own scope is deliberately split from S002 so the ~85% buildable/testable work is not held hostage to the admin-only remainder — REVIEW §G's own layer-by-layer table is the mitigation design | Track DEC-Q10 as an explicit, separately-ticketed item (REVIEW's own recommendation: "PF-REL-ADMIN-01") so agent-completable work is never silently gated on an unstaffed admin task | unassigned | open | Cannot be reduced further within this wave's scope — genuine human-administration dependency |
| RISK-W06-002 | DX-06 T2's OpenAPI 3.1 validator dependency (`pb33f/libopenapi` candidate) is chosen at implementation time without this wave's planning documents having pre-vetted its license or security posture | Medium | Medium — an unreviewed third-party validator dependency in a release-adjacent tool could introduce a supply-chain or licensing issue | Medium | W06-E02-S001 | MATRIX CS-15's own risk note requires "security-review licence" as part of the evaluation, not an afterthought — this wave's S001 records the decision as its own task with that requirement stated explicitly | If the evaluated candidate fails security/licence review, record the rejection and the alternative chosen as a deviation, not a silent substitution | unassigned | open | Low once the evaluation task's own review step is honored |
| RISK-W06-003 | REL-03b's three blocked legs (T3 on DX-06, T5 on DX-03+AR-03, T7 on DX-04) remain blocked past this wave's own closure if their unblocking stories land later than W06-E02-S003 attempts to close | Medium | Medium — a REL-03b leg landing late does not block this wave's other stories, but leaves REL-03 as a whole (a single MATRIX CS-15 closure spec) only partially closed at this wave's exit | Medium | W06-E02-S003 | The story's own `story.md` records explicit per-leg blocked-entry criteria rather than a single opaque "blocked" status, so a partial-acceptance disposition is possible and honestly recorded (per `governance/definition-of-done.md`'s "partially-accepted" status) | If a leg remains blocked at this wave's closure, record it in `closure-report.md` as explicitly deferred with its unblocking condition restated, not silently dropped | unassigned | open | Low-medium — the risk is schedule, not design; each leg's unblocking condition is already fully specified |
| RISK-W06-004 | ADR-005's own unresolved caveat ("verify against the pinned GoReleaser version at implementation time... not yet independently confirmed") surfaces a real incompatibility between the ratified split-mode decision and this repository's actual pinned GoReleaser version, only discovered during W06-E03-S001's T6 implementation | Low-medium | Medium — if GoReleaser's `--skip=publish` does not behave as REVIEW §F row 6 assumed, T6's implementation strategy would need to change mid-story | Low-medium | W06-E03-S001 | T6's own task record requires the version-confirmation step to happen before implementation is trusted as correct, per ADR-005's own "Consequences" section | If the pinned version does not support the assumed split-mode behavior, record this as a deviation from ADR-005 and escalate — do not silently hand-roll a substitute pipeline without recording why the ratified decision no longer holds | unassigned | open | Low once the version-confirmation step is executed as planned |

## Residual risk after mitigation

RISK-W06-002 and RISK-W06-004 are expected to reduce to low residual risk once their respective
implementation-time review/confirmation steps are executed as planned. RISK-W06-003 is a scheduling
risk, not a design gap, and is expected to resolve naturally as the unblocking stories land, tracked
honestly via partial-acceptance status if it does not. RISK-W06-001 cannot be resolved within this
wave's own execution capacity — it is a genuine, irreducible-within-this-wave human-administration
dependency, tracked but not eliminated.
