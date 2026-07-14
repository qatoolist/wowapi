---
id: GOV-TEMPLATE-WAVE
type: template
title: Wave document template
status: template
owner: <owner>
reviewer: <acceptance-authority>
priority: <priority>
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for a wave-level `wave.md`. Copy into `impl/waves/wave-<NN>-<descriptive-name>/wave.md`
and replace every placeholder. Front matter fields per mandate §6 pattern adapted to wave level;
body sections per mandate §8.2. Do not invent content — leave the guidance line as a prompt for
the future author until the wave is actually planned.
-->

---
id: <W NN>
type: wave
title: <Wave title>
status: draft
owner: <owner>
reviewer: <acceptance-authority>
priority: <priority>
created_at: <YYYY-MM-DD>
updated_at: <YYYY-MM-DD>
source_requirements: []
depends_on: []
epics: []
risks: []
entry_criteria: []
exit_criteria: []
---

# <W NN> — <Wave title>

## Wave ID and title

*State the wave's stable identifier and its title exactly as they appear in the front matter.*

## Objective

*State the objective in one or two sentences: what capability does this wave deliver and why now.*

## Rationale

*Explain why this wave exists at this point in the sequence — what dependency or risk makes it necessary here rather than earlier or later.*

## Framework capabilities delivered

*List the concrete, generic framework capabilities this wave adds or hardens — not product features.*

## Included epics

*List the epics contained in this wave, by ID and title, with a one-line description of each.*

## Entry criteria

*State the conditions that must hold before this wave may be marked `ready` — predecessor wave acceptance, required decisions, required baselines.*

## Exit criteria

*State the measurable conditions that must hold before this wave may be marked `accepted`.*

## Dependencies

*List dependencies on earlier waves, external decisions, or repository state, each with its dependency type.*

## Assumptions

*State any assumptions this wave's plan relies on, distinguished from confirmed facts.*

## Risks

*List the risks specific to this wave, referencing risk register IDs where they exist.*

## Quality gates

*State the quality gates this wave must pass — coverage floor, lint, static analysis, security scan, etc.*

## Required artifacts

*List the artifact types this wave is expected to produce, referencing the artifact index.*

## Required evidence

*List the evidence types this wave is expected to produce, referencing the evidence index.*

## Expected implementation outcome

*Describe, in outcome terms, what the framework will be able to do once this wave is accepted.*

## Acceptance authority

*Name the role or person with authority to accept this wave's closure.*

## Closure conditions

*State the conditions that must all be true for this wave to be closed — epic acceptance, evidence completeness, no unresolved deviations.*
