---
id: GOV-TEMPLATE-VERIFICATION
type: template
title: Story-level verification record template
status: template
parent_story: <W NN-E NN-S NNN>
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Standalone version of the mandate §8.8 verification skeleton, for use as
`story-<NNN>-<name>/verification.md` — aggregated across the story's acceptance criteria. This
file must initially define the planned verification procedure only; post-execution fields are
filled in after verification actually runs.
-->

---
id: <VER-W NN-E NN-S NNN>
type: verification-record
parent_story: <W NN-E NN-S NNN>
status: draft
created_at: <YYYY-MM-DD>
updated_at: <YYYY-MM-DD>
---

# Verification record — <W NN-E NN-S NNN>

## Planned verification procedure

*Per mandate §8.8. One row per acceptance criterion for this story.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| *<AC-W NN-E NN-S NNN-01>* | *<method>* | *<environment>* | *<expected result>* | *<evidence type>* | *<reviewer>* |
| *<AC-W NN-E NN-S NNN-02>* | *<method>* | *<environment>* | *<expected result>* | *<evidence type>* | *<reviewer>* |

## Post-execution record

*Fill in after verification is actually executed. Do not record results that were not actually observed.*

### Actual result

*Record the actual result per acceptance criterion.*

### Pass or fail

*Record pass or fail per acceptance criterion.*

### Evidence identifier

*Record the evidence ID(s) produced, referencing `evidence/index.md`.*

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

*Record the final conclusion of verification for this story.*
