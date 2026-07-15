---
id: GOV-TEMPLATE-EPIC
type: template
title: Epic document template
status: template
owner: <owner>
reviewer: <reviewer>
priority: <priority>
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for an epic-level `epic.md`. Copy into
`impl/waves/wave-<NN>-<name>/epics/epic-<NNN>-<descriptive-name>/epic.md` and replace every
placeholder. Front matter and body sections per mandate §8.3.
-->

---
id: <W NN-E NN>
type: epic
title: <Epic title>
status: draft
wave: <W NN>
owner: <owner>
reviewer: <reviewer>
priority: <priority>
created_at: <YYYY-MM-DD>
updated_at: <YYYY-MM-DD>
source_requirements: []
depends_on: []
stories: []
decisions: []
risks: []
---

# <W NN-E NN> — <Epic title>

## Epic objective

*State the objective in one or two sentences: what coherent capability does this epic deliver.*

## Problem being solved

*Describe the problem or gap this epic addresses, tracing back to the source finding or requirement.*

## Scope

*Bound what is included in this epic's capability.*

## Out of scope

*State explicitly what is adjacent but excluded, and where that excluded work is tracked instead.*

## Source requirements

*List the source requirement IDs this epic implements, verifies, or otherwise addresses.*

## Architectural context

*Describe the relevant architectural context — affected layers, packages, or contracts — needed to understand this epic's stories.*

## Included stories

*List the stories contained in this epic, by ID and title, with a one-line description of each.*

## Dependencies

*List cross-story, cross-epic, or cross-wave dependencies this epic has.*

## Risks

*List the risks specific to this epic, referencing risk register IDs where they exist.*

## Required decisions

*List architectural or implementation decisions this epic requires, referencing decision register IDs, including unresolved ones.*

## Epic acceptance criteria

*State the measurable, epic-level acceptance criteria that must hold for this epic to be accepted — distinct from individual story acceptance criteria.*

## Closure conditions

*State the conditions that must all be true for this epic to be closed — story acceptance, decision resolution, no unresolved deviations.*
