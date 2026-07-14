---
id: GOV-TEMPLATE-EVIDENCE
type: template
title: Evidence record template
status: template
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for a single evidence record. Copy into the appropriate subdirectory under
`.../story-<NNN>-<name>/evidence/{baselines,tests,coverage,logs,screenshots,benchmarks,security,
static-analysis,compatibility,regression,reviews,acceptance}/` and reference it from
`evidence/index.md`. Fields per mandate §10.

Guidance: failed evidence must be preserved, not deleted, when a later run passes. Mark status
using only: failed, superseded, retested, resolved, accepted exception. Evidence that does not
identify the tested revision must not be treated as final proof.
-->

---
id: <EV-W NN-E NN-S NNN-NNN>
type: evidence
title: <Evidence title>
status: template
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# <EV-W NN-E NN-S NNN-NNN> — <Evidence title>

## Evidence ID

*State the stable evidence identifier.*

## Evidence type

*State the evidence type (e.g. unit-test report, coverage report, benchmark result, security scan, review report, acceptance approval).*

## Story and task

*State the story ID and, if applicable, task ID this evidence supports.*

## Acceptance criteria proven

*List the acceptance criteria IDs this evidence proves.*

## Execution command

*State the exact command executed to produce this evidence.*

## Code revision or commit SHA

*State the commit SHA the evidence was produced against. Evidence without this must not be treated as final proof.*

## Branch or tag

*State the branch or tag the commit belongs to.*

## Execution environment

*Describe the environment the evidence was produced in (local, CI, staging, etc.).*

## Relevant tool versions

*List relevant tool versions (Go version, linter version, database version, etc.).*

## Date and time

*Record the date and time the evidence was produced.*

## Result

*State the result — pass, fail, or other outcome relevant to the evidence type.*

## File or URI

*State the path or URI where the raw evidence artifact is stored.*

## Checksum where appropriate

*State a checksum for the evidence file, where appropriate.*

## Reviewer

*State who reviewed this evidence.*

## Superseded evidence where applicable

*Reference any earlier evidence record this supersedes. Status vocabulary for evidence records:
failed / superseded / retested / resolved / accepted exception. Do not delete earlier failed
verification merely because a later run passes — mark it superseded and keep it.*
