---
id: GOV-TEMPLATE-DECISION
type: template
title: Architectural/implementation decision record template
status: template
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for a single decision record (ADR), per mandate §11.8. Copy into
`.../story-<NNN>-<name>/decisions/adr-<NNN>-<descriptive-name>.md` and reference it from
`decisions/index.md` and from `tracking/decision-register.md`.

Mandate §11.8, verbatim: "Record architectural and implementation decisions, including unresolved
decisions. Do not bury decisions only in prose."

Status vocabulary aligns with `decision-register.md`: proposed / ratified-pending-adr / ratified /
superseded / rejected.
-->

---
id: <ADR-W NN-E NN-S NNN-NNN>
type: decision
title: <Decision title>
status: proposed
context: <one-line pointer to the problem forcing this decision>
date: <YYYY-MM-DD>
deciders: []
related_source_items: []
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# <ADR-W NN-E NN-S NNN-NNN> — <Decision title>

## Decision ID

*State the stable decision identifier, following the §5 pattern ADR-W<NN>-E<NN>-S<NNN>-<NNN>.*

## Title

*State the decision title.*

## Status

*State the current status — one of: proposed, ratified-pending-adr, ratified, superseded, rejected.*

## Context

*Describe the problem forcing this decision — what question must be answered before work can proceed, and why it cannot be deferred.*

## Options considered

*List the options considered, with a brief note on the trade-offs of each. Include the option of doing nothing where relevant.*

## Decision

*State the decision made. If still unresolved, state that explicitly rather than omitting this section — unresolved decisions must still be recorded, not buried in prose.*

## Rationale

*Explain why this option was chosen over the alternatives.*

## Consequences

*State the consequences of this decision — what becomes easier, what becomes harder, what constraints it imposes on future work.*

## Related source items

*List related source requirement, finding, story, or epic IDs.*

## Date

*State the date this decision was made or last revisited.*

## Deciders

*List who made or is responsible for making this decision.*
