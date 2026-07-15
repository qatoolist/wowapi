---
id: GOV-TEMPLATE-TASK
type: template
title: Task document template
status: template
parent_story: <W NN-E NN-S NNN>
owner: <owner>
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for a task-level `task.md`. Copy into
`.../stories/story-<NNN>-<name>/tasks/task-<NNN>-<descriptive-name>/task.md` and replace every
placeholder. Per the repository's documented per-task adaptation, this single file carries the
task definition (§8.6), implementation record (§8.7), verification record (§8.8), and deviations
record (§8.9) as sections, rather than four separate files as used at story scope.
-->

---
id: <W NN-E NN-S NNN-T NNN>
type: task
title: <Task title>
status: todo
parent_story: <W NN-E NN-S NNN>
owner: <owner>
created_at: <YYYY-MM-DD>
updated_at: <YYYY-MM-DD>
depends_on: []
acceptance_criteria: []
artifacts: []
evidence: []
---

# <W NN-E NN-S NNN-T NNN> — <Task title>

## Task Definition

*Per mandate §8.6. This section defines the task before work begins.*

### Task objective

*State the objective in one or two sentences: what concrete implementation or verification activity does this task perform.*

### Parent story

*State the parent story ID and title.*

### Owner

*State the owner responsible for this task.*

### Status

*State the current status, drawn only from the task status vocabulary in `governance/status-model.md` §7.3.*

### Dependencies

*List other tasks, decisions, or external factors this task depends on.*

### Detailed work

*Describe the detailed work this task involves.*

### Expected files or components affected

*List the files or components expected to be affected, where determinable.*

### Expected output

*Describe the expected output of this task.*

### Required artifacts

*List the artifact types this task is expected to produce.*

### Required evidence

*List the evidence types this task is expected to produce.*

### Related acceptance criteria

*List the story-level acceptance criteria IDs this task contributes to.*

### Completion criteria

*State the criteria that determine when this task is complete.*

### Verification method

*State how this task's output will be verified.*

### Risks

*List risks specific to this task.*

### Rollback or recovery considerations

*State rollback or recovery considerations, where applicable.*

## Implementation Record

*Per mandate §8.7. Do not pre-populate implementation claims for work that has not yet occurred.*

### What was actually implemented

*Record what was actually implemented, once implementation has occurred.*

### Components changed

*List the components actually changed.*

### Files changed

*List the files actually changed.*

### Interfaces introduced or changed

*List interfaces introduced or changed.*

### Configuration changes

*List configuration changes actually made.*

### Schema or migration changes

*List schema or migration changes actually made.*

### Security changes

*List security changes actually made.*

### Observability changes

*List observability changes actually made.*

### Tests added or modified

*List tests added or modified.*

### Commits

*List the commit SHAs associated with this task's implementation.*

### Pull requests

*List the pull request(s) associated with this task's implementation.*

### Implementation dates

*Record the dates implementation work occurred.*

### Technical debt introduced

*Record any technical debt introduced, referencing the technical-debt register.*

### Known limitations

*Record known limitations of the implementation.*

### Follow-up items

*Record follow-up items arising from this task.*

### Relationship to the approved plan

*State whether the implementation matched the approved plan, and reference any deviation records if it did not.*

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| *<AC ID>* | *<method>* | *<environment>* | *<expected result>* | *<evidence type>* | *<reviewer>* |

### Actual result

*Record the actual result once verification has been executed.*

### Pass or fail

*Record pass or fail.*

### Evidence identifier

*Record the evidence ID(s) produced.*

### Execution date

*Record the date verification was executed.*

### Commit or revision

*Record the commit SHA or revision verified.*

### Environment

*Record the environment verification was executed in.*

### Reviewer

*Record who reviewed the verification.*

### Findings

*Record any findings from verification.*

### Retest status

*Record whether a retest was required and its status.*

### Final conclusion

*Record the final conclusion of verification.*

## Deviations Record

*Per mandate §8.9. Initially state that deviations are not yet known. The approved plan must not be silently altered to hide deviations.*

*No deviations recorded yet.*

### Deviation ID

*Assign a stable deviation ID (e.g. DEV-<task-id>-001) if a deviation occurs.*

### Approved plan

*State what the approved plan said.*

### Actual implementation

*State what was actually implemented.*

### Reason

*State the reason for the deviation.*

### Impact

*State the impact of the deviation.*

### Risks

*State risks introduced by the deviation.*

### Approval

*State who approved the deviation and when.*

### Compensating controls

*State any compensating controls put in place.*

### Follow-up work

*State any follow-up work arising from the deviation.*
