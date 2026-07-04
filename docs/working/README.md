# wowapi — Working Capability Layer

A reusable operating system for developing, testing, reviewing, remediating, and regressing this project.
Grounded in the actual codebase, roadmaps, workflows, and three review passes' findings — not generic
advice. **Read this first, then keep these artifacts current as the project evolves.**

## The artifacts

| # | Artifact | Use it to… |
|---|---|---|
| 1 | [skills-and-knowledge-map.md](skills-and-knowledge-map.md) | Learn what you must know to work here (architecture, RLS, authz, async, compliance primitives, testing) with the real packages/traps |
| 2 | [best-practices.md](best-practices.md) | Know how to do the work (understand→inspect→reuse→test→gate→document→review) |
| 3 | [working-persona.md](working-persona.md) | Adopt the mindset (7 fused roles) before every task |
| 4 | [internal-scripts-policy.md](internal-scripts-policy.md) | Know where/how to put reusable check scripts (`miscellaneous/`) |
| 5 | [quality-gate-checklist.md](quality-gate-checklist.md) | Pass the mandatory review gate before declaring done |
| 6 | [review-learning-register.md](review-learning-register.md) | See what past reviews found and the rules that prevent recurrence |
| — | [`miscellaneous/`](../../miscellaneous/) | Run the mechanical checks (`review_gate.sh` + the individual audits) |

Related: the AI-agent memories `goal-completion-gate` + `review-learnings` (loaded each session), the
`independent-review-gate` skill, and the deeper references in `docs/blueprint/`, `docs/implementation/`
(decisions + evidence), and `docs/operations/`.

## How to use them (every task)

1. **Before implementing** — adopt the [persona](working-persona.md); read the relevant
   [skills map](skills-and-knowledge-map.md) section + blueprint; inspect existing code; enumerate every
   sub-requirement. Record load-bearing decisions in `docs/implementation/decisions.md` first.
2. **While implementing** — follow [best-practices.md](best-practices.md): reuse conventions, extend the
   right package, wire end to end (Kernel → Context → boot), provide required infra, TDD with real
   Postgres tests.
3. **Before declaring done** — run `miscellaneous/review_gate.sh` (mechanical) then the
   [quality-gate-checklist.md](quality-gate-checklist.md) with a fresh reviewer (the
   `independent-review-gate` skill). Fix→re-test→re-review until no third-party-review-level issue remains.
4. **When a review finds something** — fix per conventions + add a test, then log it in the
   [review-learning-register.md](review-learning-register.md); promote recurring classes to checklist
   rules and, if mechanically detectable, a script.

## For AI agents specifically
The `independent-review-gate` skill is mandatory before any `/goal` is marked complete (enforced by the
global CLAUDE.md rule + the `goal-completion-gate` memory). Load this directory when starting substantive
work; prefer `mcp__lumen__semantic_search` for discovery; never invent APIs/config/columns; never
duplicate existing tests/implementations.

## Maintenance
These are living documents. When the architecture, conventions, or risks change, update the affected
artifact in the same change — stale guidance is worse than none. Keep everything grounded in real project
facts (name the packages, files, and commands).
