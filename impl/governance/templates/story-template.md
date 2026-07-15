---
id: GOV-TEMPLATE-STORY
type: template
title: Story document template
status: template
owner: unassigned
reviewer: unassigned
priority: <priority>
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for a story-level `story.md`. Copy into
`impl/waves/wave-<NN>-<name>/epics/epic-<NNN>-<name>/stories/story-<NNN>-<descriptive-name>/story.md`
and replace every placeholder. Front matter follows the full example given in mandate §6. Body
sections per mandate §8.4, plus a "## Plan" section reproducing the §8.5 plan skeleton — see the
guidance note under that heading for why it also belongs in a sibling `plan.md` file.
-->

---
id: <W NN-E NN-S NNN>
type: story
title: <Story title>
status: draft
wave: <W NN>
epic: <W NN-E NN>
owner: unassigned
reviewer: unassigned
priority: <priority>
created_at: <YYYY-MM-DD>
updated_at: <YYYY-MM-DD>
source_requirements: []
depends_on: []
blocks: []
acceptance_criteria: []
artifacts: []
evidence: []
decisions: []
risks: []
---

# <W NN-E NN-S NNN> — <Story title>

## Story ID

*Restate the story's stable identifier exactly as it appears in the front matter.*

## Title

*Restate the story title exactly as it appears in the front matter.*

## Objective

*State the objective in one or two sentences: what testable unit of framework value does this story deliver.*

## Value to the framework

*Explain why this story matters to the framework as a generic platform kernel, not to any specific downstream product.*

## Problem statement

*Describe the specific problem, gap, or finding this story resolves, tracing back to its source.*

## Source requirements

*List the source requirement IDs this story implements, verifies, or otherwise addresses.*

## Current-state assessment

*Describe the current, verified state of the affected code or capability — distinguish what is confirmed from what is assumed.*

## Desired state

*Describe the state the framework will be in once this story is accepted.*

## Scope

*Bound precisely what this story covers. Avoid vague scope statements per mandate §4.3.*

## Out of scope

*State explicitly what is adjacent but excluded, and where that excluded work is tracked instead.*

## Assumptions

*State any assumptions this story's plan relies on, distinguished from confirmed facts.*

## Dependencies

*List dependencies on other stories, epics, decisions, or external factors, each with its dependency type.*

## Affected packages or components

*List the packages, modules, or components this story is expected to touch.*

## Compatibility considerations

*State backward-compatibility implications and how they will be preserved or explicitly broken with rationale.*

## Security considerations

*State security implications and required controls.*

## Performance considerations

*State performance implications and any budgets or benchmarks that apply.*

## Observability considerations

*State logging, metrics, or tracing implications.*

## Migration considerations

*State data, schema, or configuration migration implications, if any.*

## Documentation requirements

*State what documentation must be created or updated as part of this story.*

## Acceptance criteria

*List numbered, measurable acceptance criteria (e.g. AC-<story-id>-01, AC-<story-id>-02, ...). Each criterion must be objectively verifiable — not aspirational.*

## Required artifacts

*List the artifact types this story is expected to produce, referencing `artifacts/index.md`.*

## Required evidence

*List the evidence types this story is expected to produce, referencing `evidence/index.md`.*

## Definition of ready

*Confirm this story satisfies `governance/definition-of-ready.md` before it may move to `ready`.*

## Definition of done

*Confirm this story will satisfy `governance/definition-of-done.md` before it may move to `accepted`.*

## Risks

*List the risks specific to this story, referencing risk register IDs where they exist.*

## Residual-risk expectations

*State what residual risk is expected to remain even after acceptance, and how it will be tracked.*

## Plan

*Guidance: the mandate's `plan.md` (§8.5) is a sibling file at `story-<NNN>-<name>/plan.md`, using
the same section list given below. This section reproduces that skeleton here so the story
template is self-contained; when instantiating a story, copy this section list into the actual
`plan.md` file rather than leaving the plan embedded in `story.md`.*

*Per mandate §8.5, verbatim: "Do not invent precise code changes where the repository does not yet
provide enough information. Clearly distinguish confirmed facts, planned changes, and
implementation assumptions."*

### Proposed architecture

*Describe the proposed architecture for this story's change.*

### Implementation strategy

*Describe the overall implementation strategy and approach.*

### Expected package or module changes

*List the packages or modules expected to change.*

### Expected file changes where determinable

*List specific files expected to change, only where this can be determined in advance.*

### Contracts and interfaces

*Describe new or changed contracts and interfaces.*

### Data structures

*Describe new or changed data structures.*

### APIs

*Describe new or changed APIs.*

### Configuration changes

*Describe configuration changes required.*

### Persistence changes

*Describe persistence-layer changes required.*

### Migration strategy

*Describe the migration strategy, if data or schema migration is involved.*

### Concurrency implications

*Describe concurrency implications and how they will be handled.*

### Error-handling strategy

*Describe the error-handling strategy for this change.*

### Security controls

*Describe security controls introduced or affected.*

### Observability changes

*Describe observability changes introduced.*

### Testing strategy

*Describe the testing strategy — unit, integration, negative, concurrency, etc.*

### Regression strategy

*Describe how regression risk will be identified and mitigated.*

### Compatibility strategy

*Describe how backward compatibility will be preserved or intentionally broken.*

### Rollout strategy

*Describe how this change will be rolled out.*

### Rollback strategy

*Describe how this change can be rolled back if necessary.*

### Implementation sequence

*Describe the intended sequence of implementation steps.*

### Task breakdown

*List the tasks this story decomposes into, by ID and title.*

### Expected artifacts

*List artifacts this plan expects to produce.*

### Expected evidence

*List evidence this plan expects to produce.*

### Unresolved questions

*List open questions that must be resolved before or during implementation.*

### Approval conditions

*State the conditions under which this plan is considered approved and ready for implementation.*
