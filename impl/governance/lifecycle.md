---
id: GOV-LIFECYCLE
type: governance
title: Lifecycle — state transitions for wave, epic, story, and task
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Lifecycle

Mandate §7. This file defines **transition rules**: which state moves to which, what must be
true before a transition is allowed, and who is authorized to make it. For the meaning of each
status word, see `status-model.md` — it is not repeated here.

## Roles referenced below

- **Owner** — the individual (or `unassigned`) named in an item's `owner` front-matter field;
  responsible for implementation-side transitions.
- **Reviewer** — the individual named in an item's `reviewer` front-matter field; responsible for
  verification-side transitions.
- **Acceptance authority** — the role/person with authority to move an item to `accepted` /
  `partially-accepted` (per `wave.md`/`epic.md`/`story.md` §8.2–8.4 "acceptance authority" /
  "definition of done" fields). May be the reviewer, a designated maintainer, or (per mandate
  §8.1) the programme owner for wave-level acceptance.

## Story lifecycle

### State diagram

```text
draft ─────▶ planned ─────▶ ready ─────▶ in-progress ─────▶ implemented ─────▶ verification ─────▶ verified ─────▶ accepted
  │             │              │               │                  │                  │                 │
  │             │              │               │                  │                  │                 │
  └────────────────────────────┴───────────────┴──────────────────┴──────────────────┴─────────────────┘
                                              (side-branches, reachable from any active state above)
                                                        │
                        ┌───────────────┬───────────────┼───────────────┐
                        ▼               ▼               ▼               ▼
                    blocked         deferred        cancelled      (return to prior
                        │               │                            active state
                        └───────────────┴──── resume ────────────────► once unblocked)
```

`blocked`, `deferred`, and `cancelled` are not steps of the main chain — they are reachable from
any of `planned`/`ready`/`in-progress`/`implemented`/`verification` when the corresponding
condition applies, and (except `cancelled`) return to the state the item was in when the
condition is resolved.

### Transition rules

| Transition | Entry criteria | Who moves it |
|---|---|---|
| `draft` → `planned` | `story.md` required content (mandate §8.4) is complete: objective, source requirements, scope/out-of-scope, numbered measurable acceptance criteria, dependencies, DoR/DoD sections drafted. | Owner |
| `planned` → `ready` | Definition of Ready satisfied in full (see `definition-of-ready.md`). | Owner, confirmed by reviewer |
| `ready` → `in-progress` | Owner assigned and has started implementation work; `plan.md` (mandate §8.5) exists. | Owner |
| `in-progress` → `implemented` | Implementation work is claimed complete: `implementation.md` populated per mandate §8.7 (what was implemented, files/components changed, tests added). This is a **claim**, not proof — see verification below. | Owner |
| `implemented` → `verification` | Verification procedure in `verification.md` (mandate §8.8 AC → method → environment → expected result → evidence type → reviewer table) is ready to execute. | Owner hands off to reviewer |
| `verification` → `verified` | Every acceptance criterion has valid evidence per `evidence-policy.md`, recorded with revision, environment, result, and evidence ID in `verification.md`. | Reviewer |
| `verified` → `accepted` | Independent-review checklist (mandate §14, reproduced in `definition-of-done.md`) passes; `closure.md` completed. | Acceptance authority |
| any active state → `blocked` | A dependency, decision, or resource is unresolved and work cannot proceed. | Owner or reviewer |
| `blocked` → prior active state | Blocking condition resolved. | Owner |
| any active state → `deferred` | Explicit approval to postpone, with a target milestone recorded. | Acceptance authority |
| any active state → `cancelled` | Rationale recorded; item will not be completed. | Acceptance authority |

Mandate §7, quoted verbatim, is the binding rule on the final transition above:

> A story must not be accepted solely because all tasks are marked complete.

Acceptance requires the full evidence-driven completion set in `definition-of-done.md` (mandate
§2.5) — task completion is necessary but never sufficient.

## Task lifecycle

### State diagram

```text
todo ─────▶ ready ─────▶ in-progress ─────▶ implemented ─────▶ verified ─────▶ done
  │            │               │                  │                │
  │            │               │                  │                │
  └────────────┴───────────────┴──────────────────┴────────────────┘
                                        │
                      ┌─────────────────┼─────────────────┐
                      ▼                 ▼                 ▼
                  blocked          cancelled        (return to prior
                      │                               active state
                      └──────── resume ───────────────► once unblocked)
```

### Transition rules

| Transition | Entry criteria | Who moves it |
|---|---|---|
| `todo` → `ready` | Task Definition of Ready satisfied (see `definition-of-ready.md`: parent story `ready`/`in-progress`, dependencies resolved or waived, owner assignable, mapped acceptance criteria identified). | Owner |
| `ready` → `in-progress` | Owner has started work. | Owner |
| `in-progress` → `implemented` | `implementation.md` §-section within the task file (Adaptation 1, see `naming-conventions.md`) records what was actually done. | Owner |
| `implemented` → `verified` | Task's verification method (mandate §8.6 "verification method") executed; evidence recorded in the task's `verification.md` section. | Owner or reviewer, per task risk |
| `verified` → `done` | Output registered: any produced artifact is in `artifacts/index.md`, any evidence is in `evidence/index.md`. | Owner |
| any active state → `blocked` | Dependency, decision, or resource unresolved. | Owner |
| `blocked` → prior active state | Blocking condition resolved. | Owner |
| any active state → `cancelled` | Rationale recorded. | Owner, confirmed by story owner |

Tasks must not be used as substitutes for acceptance criteria (mandate §4.4) — a task reaching
`done` proves that task's own completion criteria, not the story's acceptance criteria as a
whole. Per the quoted §7 rule above, story acceptance is judged against the story's own AC and
evidence, never inferred from "all tasks done."

## Epic lifecycle

Epics share the wave/epic vocabulary (`status-model.md` §7.1).

### State diagram

```text
proposed ─────▶ planned ─────▶ ready ─────▶ in-progress ─────▶ verification ─────▶ accepted
   │               │              │               │                   │
   │               │              │               │                   │
   └───────────────┴──────────────┴───────────────┴───────────────────┘
                                          │
                ┌─────────────┬──────────┼──────────┬─────────────┐
                ▼             ▼          ▼          ▼             ▼
            blocked      deferred   cancelled  partially-      (return to
                │             │          │      accepted        prior state
                └─────────────┴────resume─┴──────────┴───────────► once unblocked)
```

### Transition rules

| Transition | Entry criteria | Who moves it |
|---|---|---|
| `proposed` → `planned` | `epic.md` required content complete (mandate §8.3): objective, scope/out-of-scope, source requirements, included stories, dependencies, epic acceptance criteria. | Epic owner |
| `planned` → `ready` | All entry criteria in `epic.md` satisfied; predecessor epics this epic depends on are `accepted` or an approved exception is documented. | Epic owner |
| `ready` → `in-progress` | At least one contained story has moved to `in-progress`. | Epic owner |
| `in-progress` → `verification` | All mandatory stories in the epic are at least `implemented`; epic-level verification (cross-story integration checks, if any) begins. | Epic owner hands off to reviewer |
| `verification` → `accepted` | All mandatory stories `accepted`; epic closure conditions (mandate §8.3) met; `closure-report.md` complete. | Acceptance authority |
| `verification` → `partially-accepted` | Some but not all mandatory stories are `accepted`; remainder explicitly `deferred` with approval or `blocked` with an open risk acknowledged. | Acceptance authority |
| any active state → `blocked` | Cross-story or cross-epic dependency unresolved. | Epic owner or reviewer |
| any active state → `deferred` | Explicit approval, target milestone recorded. | Acceptance authority |
| any active state → `cancelled` | Rationale recorded. | Acceptance authority |

An epic must not reach `accepted` merely because its stories are `implemented` — the same §7
rule applies one level up: story-level `accepted` status (achieved through the full evidence
chain) is the unit epics roll up from, not task or implementation completion claims.

## Wave lifecycle

### State diagram

```text
proposed ─────▶ planned ─────▶ ready ─────▶ in-progress ─────▶ verification ─────▶ accepted
   │               │              │               │                   │
   │               │              │               │                   │
   └───────────────┴──────────────┴───────────────┴───────────────────┘
                                          │
                ┌─────────────┬──────────┼──────────┬─────────────┐
                ▼             ▼          ▼          ▼             ▼
            blocked      deferred   cancelled  partially-      (return to
                │             │          │      accepted        prior state
                └─────────────┴────resume─┴──────────┴───────────► once unblocked)
```

### Transition rules

| Transition | Entry criteria | Who moves it |
|---|---|---|
| `proposed` → `planned` | `wave.md` required content complete (mandate §8.2): objective, rationale, included epics, entry/exit criteria, dependencies, risks, quality gates. | Wave owner (programme owner) |
| `planned` → `ready` | Entry criteria satisfied: per mandate §15, "A later wave must not be marked ready when mandatory predecessor capabilities remain unaccepted, unless an approved exception is documented." | Wave owner |
| `ready` → `in-progress` | At least one contained epic has moved to `in-progress`. | Wave owner |
| `in-progress` → `verification` | All mandatory epics in the wave are `accepted` or `partially-accepted` with documented rationale; wave exit criteria checks begin. | Wave owner hands off to reviewer |
| `verification` → `accepted` | Wave exit criteria (mandate §8.2) fully met; `closure-report.md` complete per `impl/index.md` "Programme acceptance." | Acceptance authority |
| `verification` → `partially-accepted` | Some mandatory epics accepted, remainder deferred-with-approval or blocked with acknowledged risk. | Acceptance authority |
| any active state → `blocked` | Cross-wave dependency or a tracked human decision (e.g. DEC-Q1/Q9/Q10) unresolved. | Wave owner |
| any active state → `deferred` | Explicit approval, target milestone recorded. | Acceptance authority |
| any active state → `cancelled` | Rationale recorded. | Acceptance authority |

Execution order for wave entry is strictly sequential (`impl/index.md`: "strictly W00→W07 for
wave entry"); independent epics/stories inside a `ready`/`in-progress` wave may proceed in
parallel per their own dependencies.

## Cross-reference

Status word definitions: `status-model.md`. Entry-condition detail for `ready`:
`definition-of-ready.md`. Completion requirements for `implemented`/`verified`/`accepted`:
`definition-of-done.md`. Evidence backing every `verified`/`accepted` transition:
`evidence-policy.md`.
