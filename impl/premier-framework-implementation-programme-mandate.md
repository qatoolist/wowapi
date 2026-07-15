We have completed substantial planning and architectural review work across several documents. However, the implementation guidance currently exists in multiple documents, sections, findings, recommendations, matrices, and backlog items.

Your goal is to consolidate this material into a complete, practical, sequenced, and independently verifiable implementation programme for the framework.

## Primary source documents

At a minimum, thoroughly review and reconcile:

* `fable5-closure-depth-matrix-2026-07-11.md`
* `fable5-final-architecture-review-2026-07-11.md`
* `premier-framework-implementation-plan.md`

Also inspect all related documents, reviews, plans, roadmaps, closure reports, gap analyses, requirement documents, architecture documents, verification reports, and implementation notes that are referenced by these files or materially affect the implementation.

Do not limit the analysis to only the three primary files when related documents contain requirements, unresolved findings, dependencies, acceptance conditions, or implementation constraints.

## Objective

Prepare a doable implementation plan that can be executed story by story and task by task until the framework reaches the intended production-ready state.

The resulting plan must:

* consolidate all valid requirements and recommendations;
* remove duplication without losing meaning;
* identify conflicts and resolve or explicitly document them;
* convert findings into executable backlog items;
* organise the work into implementation waves;
* identify dependencies and sequencing;
* define measurable acceptance criteria;
* specify expected implementation artifacts;
* specify required verification evidence;
* support progress tracking at wave, epic, story, and task level;
* support independent review and audit;
* make it possible to determine whether every source requirement was implemented, deferred, rejected, or superseded.

This is not merely a summary or reformatting exercise. The output must become the authoritative implementation blueprint and execution-tracking structure.

---

# 1. Required analysis before creating the plan

Before generating implementation files, perform a complete analysis of the source material.

## 1.1 Build a source requirement inventory

Extract all actionable items, including:

* architectural findings;
* unresolved gaps;
* closure-depth requirements;
* mandatory framework capabilities;
* security findings;
* reliability requirements;
* testing requirements;
* observability requirements;
* configuration requirements;
* API requirements;
* data and persistence requirements;
* migration requirements;
* tenancy and isolation requirements;
* extensibility requirements;
* plugin or module requirements;
* developer-experience requirements;
* documentation requirements;
* CI/CD requirements;
* release requirements;
* operational-readiness requirements;
* backward-compatibility requirements;
* performance requirements;
* static-analysis requirements;
* quality-gate requirements;
* review findings;
* verification conditions;
* technical debt;
* deferred decisions;
* known risks.

Every extracted requirement or recommendation must receive a stable source identifier or retain its existing identifier where one already exists.

## 1.2 Reconcile overlapping documents

Where multiple documents discuss the same subject:

* identify the authoritative or latest requirement;
* merge compatible requirements;
* preserve stricter requirements unless there is a documented reason not to;
* identify obsolete or superseded guidance;
* document conflicts;
* do not silently discard any finding;
* do not duplicate the same work across multiple waves or stories.

## 1.3 Classify each source item

Each source item must be classified as one of:

* implementation requirement;
* architecture decision;
* design constraint;
* verification requirement;
* documentation requirement;
* operational requirement;
* quality gate;
* risk;
* technical debt;
* future enhancement;
* rejected recommendation;
* superseded recommendation;
* informational context.

## 1.4 Determine disposition

Every actionable source item must have a disposition:

* planned;
* already implemented but requires verification;
* partially implemented;
* blocked;
* deferred;
* rejected with rationale;
* superseded;
* duplicate of another item;
* not applicable with rationale.

No material source item may disappear merely because it does not fit neatly into the new plan.

---

# 2. Planning principles

The implementation programme must follow these principles.

## 2.1 Doability over theoretical completeness

The plan must be comprehensive but executable.

Break large ambitions into deliverable increments. Avoid stories that require broad, undefined, or unbounded implementation.

Each story should produce a meaningful, testable improvement to the framework.

## 2.2 Dependency-aware sequencing

Order work based on technical dependencies rather than document order.

Examples:

* foundational contracts before adapters;
* configuration foundations before configuration-dependent modules;
* observability primitives before module-specific instrumentation;
* tenancy contracts before tenant-aware persistence;
* lifecycle primitives before advanced extension systems;
* test infrastructure before coverage enforcement;
* implementation before final verification;
* baseline capture before behavioural changes.

## 2.3 Framework-first scope

The plan must preserve the generic framework boundary.

Do not introduce housing-society-specific, legal-domain-specific, product-specific, or application-specific concepts into the framework unless the source architecture explicitly requires a generic abstraction supporting such use cases.

Where a requirement belongs to a downstream product rather than the framework:

* record it;
* classify it as product-level;
* exclude it from framework implementation;
* provide the rationale;
* identify any generic framework capability that must exist to support it.

## 2.4 Traceability

The plan must support this complete traceability chain:

```text
Source document
→ Source requirement or finding
→ Wave
→ Epic
→ Story
→ Acceptance criterion
→ Task
→ Implementation change
→ Artifact
→ Verification evidence
→ Review
→ Acceptance
```

## 2.5 Evidence-driven completion

No story should be considered complete merely because code was written.

Completion must require:

* implementation;
* required artifacts;
* tests;
* evidence;
* acceptance-criteria verification;
* review;
* documentation;
* resolution or acknowledgement of deviations and residual risks.

## 2.6 Preserve plan-versus-actual history

The approved implementation plan must not be rewritten after implementation to make it appear that the final implementation always matched the plan.

Differences must be recorded in a separate deviation record.

---

# 3. Required directory structure

Create the implementation programme under:

```text
impl/
```

Use the following structure as the baseline:

```text
impl/
├── index.md
│
├── governance/
│   ├── lifecycle.md
│   ├── status-model.md
│   ├── definition-of-ready.md
│   ├── definition-of-done.md
│   ├── evidence-policy.md
│   ├── artifact-policy.md
│   ├── traceability-policy.md
│   ├── naming-conventions.md
│   └── templates/
│       ├── wave-template.md
│       ├── epic-template.md
│       ├── story-template.md
│       ├── task-template.md
│       ├── implementation-template.md
│       ├── verification-template.md
│       ├── evidence-template.md
│       ├── artifact-template.md
│       ├── deviation-template.md
│       └── decision-template.md
│
├── analysis/
│   ├── source-inventory.md
│   ├── requirement-inventory.md
│   ├── findings-disposition.md
│   ├── conflict-resolution.md
│   ├── duplicate-analysis.md
│   ├── scope-boundary.md
│   └── planning-assumptions.md
│
├── tracking/
│   ├── index.md
│   ├── status-register.md
│   ├── source-traceability-matrix.md
│   ├── requirement-traceability-matrix.md
│   ├── dependency-register.md
│   ├── artifact-register.md
│   ├── evidence-register.md
│   ├── decision-register.md
│   ├── risk-register.md
│   ├── technical-debt-register.md
│   ├── deviation-register.md
│   ├── deferred-items-register.md
│   └── change-log.md
│
└── waves/
    ├── index.md
    │
    ├── wave-00-<descriptive-name>/
    │   ├── wave.md
    │   ├── progress.md
    │   ├── dependencies.md
    │   ├── risks.md
    │   ├── acceptance.md
    │   ├── closure-report.md
    │   │
    │   └── epics/
    │       ├── index.md
    │       │
    │       └── epic-001-<descriptive-name>/
    │           ├── epic.md
    │           ├── progress.md
    │           ├── dependencies.md
    │           ├── risks.md
    │           ├── acceptance.md
    │           ├── closure-report.md
    │           │
    │           └── stories/
    │               ├── index.md
    │               │
    │               └── story-001-<descriptive-name>/
    │                   ├── story.md
    │                   ├── plan.md
    │                   ├── implementation.md
    │                   ├── verification.md
    │                   ├── deviations.md
    │                   ├── closure.md
    │                   │
    │                   ├── tasks/
    │                   │   ├── index.md
    │                   │   │
    │                   │   └── task-001-<descriptive-name>/
    │                   │       ├── task.md
    │                   │       ├── implementation.md
    │                   │       ├── verification.md
    │                   │       └── deviations.md
    │                   │
    │                   ├── decisions/
    │                   │   ├── index.md
    │                   │   └── adr-001-<descriptive-name>.md
    │                   │
    │                   ├── artifacts/
    │                   │   ├── index.md
    │                   │   ├── pre-implementation/
    │                   │   ├── implementation/
    │                   │   └── post-implementation/
    │                   │
    │                   └── evidence/
    │                       ├── index.md
    │                       ├── baselines/
    │                       ├── tests/
    │                       ├── coverage/
    │                       ├── logs/
    │                       ├── screenshots/
    │                       ├── benchmarks/
    │                       ├── security/
    │                       ├── static-analysis/
    │                       ├── compatibility/
    │                       ├── regression/
    │                       ├── reviews/
    │                       └── acceptance/
    │
    ├── wave-01-<descriptive-name>/
    │   └── ...
    │
    └── wave-NN-<descriptive-name>/
        └── ...
```

The precise number of waves, epics, stories, and tasks must be determined from the actual analysis. Do not artificially force the work into three waves merely because the example shows `wave0`, `wave1`, and `wave2`.

Use as many waves as are necessary, but keep them meaningful and manageable.

---

# 4. Required planning hierarchy

Use the following hierarchy:

```text
Implementation programme
    └── Wave
        └── Epic
            └── Story
                └── Task
```

## 4.1 Wave

A wave represents a major implementation stage with a coherent outcome.

A wave must:

* have a clear objective;
* provide a meaningful framework milestone;
* define entry criteria;
* define exit criteria;
* identify included epics;
* identify dependencies on earlier waves;
* identify risks;
* identify required evidence;
* define wave-level acceptance;
* be independently closable.

Examples of possible wave themes may include:

* baseline and implementation readiness;
* foundational contracts and lifecycle;
* configuration and runtime foundations;
* persistence and transaction infrastructure;
* tenancy and isolation;
* transport and API foundations;
* observability and operational controls;
* security hardening;
* extensibility and module systems;
* testing and reliability hardening;
* performance and compatibility;
* documentation and developer experience;
* production-readiness verification.

These are examples only. Derive the actual wave structure from the source material.

## 4.2 Epic

An epic represents a substantial capability or workstream within a wave.

An epic must:

* deliver a coherent capability;
* contain multiple related stories where appropriate;
* avoid mixing unrelated concerns;
* define epic-level acceptance criteria;
* identify cross-story dependencies;
* identify architectural decisions;
* define epic closure conditions.

## 4.3 Story

A story represents a testable unit of framework value.

Each story must be:

* specific;
* bounded;
* implementable;
* independently reviewable;
* independently verifiable;
* traceable to source requirements;
* associated with measurable acceptance criteria.

Avoid vague stories such as:

* improve architecture;
* fix framework;
* enhance quality;
* add production readiness;
* improve testing.

Convert such items into precise, testable stories.

## 4.4 Task

A task represents a concrete implementation or verification activity.

Tasks may include:

* code implementation;
* refactoring;
* test creation;
* migration creation;
* configuration changes;
* documentation;
* benchmarking;
* static analysis;
* security validation;
* compatibility validation;
* artifact generation;
* evidence collection;
* independent review;
* closure activities.

Tasks must not be used as substitutes for acceptance criteria.

---

# 5. Identifier and naming rules

Use stable, immutable identifiers.

Examples:

```text
W00
W00-E01
W00-E01-S001
W00-E01-S001-T001

REQ-ARCH-001
AC-W00-E01-S001-01
ART-W00-E01-S001-001
EV-W00-E01-S001-001
ADR-W00-E01-S001-001
RISK-W00-E01-S001-001
DEV-W00-E01-S001-001
TD-W00-E01-S001-001
```

Rules:

* never reuse an identifier;
* never renumber existing identifiers merely to improve ordering;
* use zero-padded numbers;
* use descriptive directory names;
* keep filenames stable and generic within an identified directory;
* avoid ambiguous names such as `story_0.md`, `impl_0.md`, `related artifacts.md`, or `misc.md`.

Preferred example:

```text
story-001-runtime-lifecycle/
├── story.md
├── plan.md
├── implementation.md
├── verification.md
└── deviations.md
```

---

# 6. Required metadata

Every wave, epic, story, and task must include structured front matter.

Example story metadata:

```yaml
---
id: W01-E02-S003
type: story
title: Introduce deterministic runtime lifecycle management
status: planned
wave: W01
epic: W01-E02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - REQ-ARCH-014
  - REQ-CLOSURE-021
depends_on:
  - W01-E01-S002
blocks:
  - W02-E01-S001
acceptance_criteria:
  - AC-W01-E02-S003-01
  - AC-W01-E02-S003-02
artifacts: []
evidence: []
decisions: []
risks: []
---
```

Metadata fields should be consistent enough to permit automated parsing and derived reporting.

Do not create multiple manually maintained sources of truth for the same status.

The canonical status belongs in the item’s primary document. Roll-up indexes and registers should either be generated from metadata or clearly marked as derived views.

---

# 7. Required status model

Define and consistently use a controlled status vocabulary.

## 7.1 Wave and epic statuses

```text
proposed
planned
ready
in-progress
blocked
verification
accepted
partially-accepted
deferred
cancelled
```

## 7.2 Story statuses

```text
draft
planned
ready
in-progress
implemented
verification
verified
accepted
blocked
deferred
cancelled
```

## 7.3 Task statuses

```text
todo
ready
in-progress
blocked
implemented
verified
done
cancelled
```

The expected story lifecycle is:

```text
draft
→ planned
→ ready
→ in-progress
→ implemented
→ verification
→ verified
→ accepted
```

Definitions:

* `implemented`: the implementation work is claimed to be complete;
* `verification`: implementation is undergoing formal verification;
* `verified`: acceptance criteria have been proven with valid evidence;
* `accepted`: the designated reviewer or authority has accepted the result.

A story must not be accepted solely because all tasks are marked complete.

---

# 8. Required content for each planning document

## 8.1 `impl/index.md`

This must provide:

* purpose of the implementation programme;
* source documents;
* scope;
* non-goals;
* planning principles;
* overall wave map;
* execution order;
* programme-level risks;
* programme-level acceptance conditions;
* links to governance, analysis, tracking, and waves;
* explanation of how implementation progress will be maintained.

## 8.2 `wave.md`

Each wave document must include:

* wave ID and title;
* objective;
* rationale;
* framework capabilities delivered;
* included epics;
* entry criteria;
* exit criteria;
* dependencies;
* assumptions;
* risks;
* quality gates;
* required artifacts;
* required evidence;
* expected implementation outcome;
* acceptance authority;
* closure conditions.

## 8.3 `epic.md`

Each epic document must include:

* epic objective;
* problem being solved;
* scope;
* out of scope;
* source requirements;
* architectural context;
* included stories;
* dependencies;
* risks;
* required decisions;
* epic acceptance criteria;
* closure conditions.

## 8.4 `story.md`

Each story document must include:

* story ID;
* title;
* objective;
* value to the framework;
* problem statement;
* source requirements;
* current-state assessment;
* desired state;
* scope;
* out of scope;
* assumptions;
* dependencies;
* affected packages or components;
* compatibility considerations;
* security considerations;
* performance considerations;
* observability considerations;
* migration considerations;
* documentation requirements;
* acceptance criteria;
* required artifacts;
* required evidence;
* definition of ready;
* definition of done;
* risks;
* residual-risk expectations.

Acceptance criteria must be numbered and measurable.

## 8.5 `plan.md`

The story plan must describe the proposed approach before implementation.

Include:

* proposed architecture;
* implementation strategy;
* expected package or module changes;
* expected file changes where determinable;
* contracts and interfaces;
* data structures;
* APIs;
* configuration changes;
* persistence changes;
* migration strategy;
* concurrency implications;
* error-handling strategy;
* security controls;
* observability changes;
* testing strategy;
* regression strategy;
* compatibility strategy;
* rollout strategy;
* rollback strategy;
* implementation sequence;
* task breakdown;
* expected artifacts;
* expected evidence;
* unresolved questions;
* approval conditions.

Do not invent precise code changes where the repository does not yet provide enough information. Clearly distinguish confirmed facts, planned changes, and implementation assumptions.

## 8.6 `task.md`

Each task must include:

* task objective;
* parent story;
* owner;
* status;
* dependencies;
* detailed work;
* expected files or components affected;
* expected output;
* required artifacts;
* required evidence;
* related acceptance criteria;
* completion criteria;
* verification method;
* risks;
* rollback or recovery considerations where applicable.

## 8.7 `implementation.md`

This file must initially provide a structured implementation-record template.

Once the work is implemented, it must record:

* what was actually implemented;
* components changed;
* files changed;
* interfaces introduced or changed;
* configuration changes;
* schema or migration changes;
* security changes;
* observability changes;
* tests added or modified;
* commits;
* pull requests;
* implementation dates;
* technical debt introduced;
* known limitations;
* follow-up items;
* relationship to the approved plan.

Do not pre-populate implementation claims for work that has not yet occurred.

## 8.8 `verification.md`

This file must initially define the planned verification procedure.

It must provide a table connecting:

```text
Acceptance criterion
→ Verification method
→ Required environment
→ Expected result
→ Evidence type
→ Reviewer
```

After execution, it must record:

* actual result;
* pass or fail;
* evidence identifier;
* execution date;
* commit or revision;
* environment;
* reviewer;
* findings;
* retest status;
* final conclusion.

## 8.9 `deviations.md`

Initially state that deviations are not yet known.

During implementation, record:

* deviation ID;
* approved plan;
* actual implementation;
* reason;
* impact;
* risks;
* approval;
* compensating controls;
* follow-up work.

The approved plan must not be silently altered to hide deviations.

## 8.10 `closure.md`

Define and later record:

* acceptance-criteria completion;
* task completion;
* artifact completeness;
* evidence completeness;
* unresolved findings;
* accepted risks;
* deferred work;
* reviewer conclusion;
* acceptance authority;
* closure date;
* final status.

---

# 9. Artifact-management requirements

Artifacts and evidence must be tracked separately.

## 9.1 Artifact definition

An artifact is something consumed, produced, modified, or delivered as part of implementation.

Examples:

* source-code packages;
* interfaces;
* schemas;
* migrations;
* API specifications;
* architecture diagrams;
* generated code;
* configuration examples;
* deployment manifests;
* runbooks;
* compatibility matrices;
* design documents;
* release notes;
* binaries;
* benchmark definitions;
* migration utilities.

## 9.2 Artifact index

Every story must contain an `artifacts/index.md` with entries containing:

* artifact ID;
* title;
* type;
* lifecycle stage;
* description;
* source requirement;
* producing task;
* repository path or storage location;
* version;
* checksum where appropriate;
* status;
* reviewer;
* retention requirement.

## 9.3 Artifact lifecycle

Organise artifacts into:

```text
pre-implementation/
implementation/
post-implementation/
```

Examples:

### Pre-implementation

* current-state baseline;
* existing behaviour report;
* architecture baseline;
* compatibility baseline;
* performance baseline;
* source inventory;
* risk assessment;
* design inputs.

### Implementation

* schemas;
* interfaces;
* migrations;
* generated outputs;
* code-generation definitions;
* configuration changes;
* architecture decisions;
* implementation notes.

### Post-implementation

* migration results;
* release notes;
* runbooks;
* upgrade guidance;
* compatibility reports;
* final architecture diagrams;
* acceptance packages;
* closure reports.

Do not duplicate large generated artifacts unnecessarily. Register their authoritative path, version, size, checksum, and generation command.

---

# 10. Evidence-management requirements

Evidence proves that implementation or acceptance criteria are satisfied.

Examples:

* unit-test reports;
* functional-test reports;
* integration-test reports;
* race-detector reports;
* coverage reports;
* benchmark results;
* static-analysis reports;
* security scans;
* dependency scans;
* compatibility results;
* CI execution records;
* screenshots;
* execution logs;
* migration logs;
* review reports;
* acceptance approvals;
* regression reports.

Every evidence record must identify:

* evidence ID;
* evidence type;
* story and task;
* acceptance criteria proven;
* execution command;
* code revision or commit SHA;
* branch or tag;
* execution environment;
* relevant tool versions;
* date and time;
* result;
* file or URI;
* checksum where appropriate;
* reviewer;
* superseded evidence where applicable.

Evidence that does not identify the tested revision must not be treated as final proof.

Failed evidence must be preserved and marked appropriately:

* failed;
* superseded;
* retested;
* resolved;
* accepted exception.

Do not delete earlier failed verification merely because a later run passes.

---

# 11. Required registers and matrices

## 11.1 Source inventory

Map every source document and relevant section.

Include:

* document;
* version or date;
* status;
* authority;
* relevant sections;
* relationships to other documents;
* whether it is active, superseded, or historical.

## 11.2 Requirement inventory

For every requirement:

* requirement ID;
* source;
* description;
* classification;
* priority;
* severity where relevant;
* status;
* disposition;
* target wave;
* target epic;
* target story;
* notes.

## 11.3 Findings disposition

For every review finding:

* finding ID;
* source;
* severity;
* description;
* current state;
* proposed resolution;
* disposition;
* implementation location;
* verification requirement;
* closure status.

## 11.4 Source traceability matrix

Map:

```text
Source document and section
→ Requirement or finding
→ Disposition
→ Planned implementation item
```

## 11.5 Requirement traceability matrix

Map:

```text
Requirement
→ Wave
→ Epic
→ Story
→ Acceptance criterion
→ Task
→ Artifact
→ Evidence
→ Final result
```

## 11.6 Dependency register

Track:

* dependency ID;
* source item;
* target item;
* dependency type;
* description;
* blocking status;
* resolution.

Identify:

* story dependencies;
* epic dependencies;
* cross-wave dependencies;
* external dependencies;
* repository dependencies;
* tooling dependencies;
* decision dependencies.

## 11.7 Risk register

Include:

* risk ID;
* description;
* likelihood;
* impact;
* severity;
* affected items;
* mitigation;
* contingency;
* owner;
* status;
* residual risk.

## 11.8 Decision register

Record architectural and implementation decisions, including unresolved decisions.

Do not bury decisions only in prose.

## 11.9 Technical-debt register

Track:

* debt ID;
* origin;
* description;
* reason accepted;
* impact;
* target resolution wave;
* related stories;
* acceptance authority.

## 11.10 Deferred-items register

Any item deferred outside the current implementation programme must include:

* source;
* rationale;
* prerequisites;
* risk of deferral;
* intended future milestone;
* approval.

---

# 12. Story sizing and decomposition rules

Stories must be decomposed when they:

* affect several unrelated framework capabilities;
* contain multiple independent acceptance outcomes;
* require different reviewers;
* span several waves;
* cannot be verified independently;
* mix foundational implementation and broad migration;
* combine implementation with unrelated documentation;
* are too broad to reasonably complete and verify as one unit.

Tasks must be decomposed when they:

* produce multiple unrelated outputs;
* need separate ownership;
* need separate evidence;
* can block independently;
* have materially different risks.

However, avoid excessive fragmentation into trivial tasks that provide no tracking value.

---

# 13. Testing and quality expectations

Derive exact requirements from the source documents, but ensure every relevant story considers:

* unit tests;
* negative tests;
* integration tests;
* functional tests;
* concurrency tests;
* race detection;
* regression tests;
* compatibility tests;
* security tests;
* performance tests;
* migration tests;
* failure-recovery tests;
* static analysis;
* linting;
* dependency analysis;
* coverage impact.

Where the existing programme requires a coverage floor, preserve it explicitly and define:

* scope of measurement;
* exclusions;
* baseline;
* minimum threshold;
* enforcement mechanism;
* CI behaviour;
* evidence format.

Do not create tests merely to increase numerical coverage. Tests must validate meaningful behaviour and failure conditions.

Avoid duplicating existing test coverage unless the new test closes an identified behavioural gap.

---

# 14. Independent review requirements

For critical stories, define an independent review step.

The reviewer should verify:

* implementation matches the approved plan or deviations are documented;
* acceptance criteria are complete;
* tests meaningfully prove behaviour;
* evidence references the correct code revision;
* artifacts are registered;
* regression risk is addressed;
* architecture boundaries are preserved;
* security implications are handled;
* no unsupported completion claims are made;
* no source requirement has been silently dropped.

Review findings must be recorded and resolved or explicitly accepted before closure.

---

# 15. Wave construction guidance

Construct waves according to dependency and risk.

Each wave should ideally:

* establish a stable baseline;
* deliver a coherent capability;
* reduce uncertainty for later work;
* contain its own verification and closure;
* avoid depending on unfinished work from a later wave;
* have an explicit exit gate.

A later wave must not be marked ready when mandatory predecessor capabilities remain unaccepted, unless an approved exception is documented.

Consider whether a dedicated Wave 00 is necessary for:

* baseline capture;
* repository health assessment;
* source reconciliation;
* test-infrastructure validation;
* CI validation;
* current coverage measurement;
* current static-analysis state;
* dependency inventory;
* architectural decision inventory;
* generation of implementation prerequisites.

Do not assume Wave 00 contains feature implementation unless that is necessary to make later implementation safe and measurable.

---

# 16. Progress tracking

Every index and progress file must provide a useful roll-up.

## 16.1 `waves/index.md`

Include:

* wave ID;
* title;
* objective;
* status;
* priority;
* dependencies;
* epic count;
* story count;
* progress;
* evidence completeness;
* planned exit gate.

## 16.2 Wave progress

Include:

* epic status;
* story status;
* blocked items;
* critical dependencies;
* open decisions;
* open risks;
* artifact completeness;
* evidence completeness;
* review state;
* exit-gate readiness.

## 16.3 Epic progress

Include:

* story status;
* task completion;
* acceptance-criteria progress;
* unresolved blockers;
* required decisions;
* verification progress;
* closure readiness.

## 16.4 Story task index

Include:

* task ID;
* title;
* owner;
* status;
* dependencies;
* output;
* related acceptance criteria;
* implementation state;
* verification state.

Avoid manually copying status into many places unless those views are generated. Clearly identify derived documents.

---

# 17. Required validation before finalising the plan

Before declaring the implementation programme complete, verify:

1. Every primary source document was fully reviewed.
2. All material related documents were considered.
3. Every actionable source item has a disposition.
4. Every planned requirement maps to a wave, epic, and story.
5. Every story has measurable acceptance criteria.
6. Every story has implementation, verification, artifact, evidence, deviation, and closure structures.
7. Dependencies are explicitly recorded.
8. No story depends on a later wave without explanation.
9. Duplicated requirements are consolidated.
10. Conflicts are documented and resolved or escalated.
11. Product-specific concerns have not leaked into the generic framework.
12. Already-implemented items are separated from unimplemented work.
13. Verification-only stories are created where implementation exists but proof is missing.
14. Artifact and evidence requirements are clearly separated.
15. Failed evidence preservation is addressed.
16. Plan-versus-actual deviation handling is defined.
17. Status values are consistent.
18. Identifiers are stable and unique.
19. Indexes and registers are internally consistent.
20. The final plan is practical enough that an implementation agent can begin from Wave 00 without re-planning the entire programme.

---

# 18. Important execution constraints

* Do not implement framework code as part of this planning task unless explicitly instructed separately.
* Do not modify unrelated source code.
* Do not create fictional completion evidence.
* Do not claim that tests passed unless they were actually executed.
* Do not claim that a requirement is implemented merely because a document mentions it.
* Do not delete or overwrite the original planning and review documents.
* Do not silently resolve ambiguous architecture decisions.
* Record assumptions explicitly.
* Preserve source references.
* Prefer precise, bounded stories over aspirational descriptions.
* Do not create placeholder files containing only headings where meaningful planning content can be derived.
* Where implementation details cannot yet be known, state what must be determined during the story rather than inventing specifics.
* Do not create unnecessary documentation duplication.
* Do not use the new implementation plan to rewrite historical facts.
* Do not commit changes unless explicitly authorised.
* Do not archive or move existing documents unless explicitly authorised.
* Do not treat the directory example as a substitute for analysing the actual work.

---

# 19. Expected final deliverables

The completed task must produce:

1. A complete `impl/` directory.
2. Governance and lifecycle documentation.
3. Source and requirement inventories.
4. Conflict and duplicate analysis.
5. A multi-wave implementation roadmap.
6. Fully defined epics.
7. Executable stories.
8. Detailed task breakdowns.
9. Acceptance criteria.
10. Planned implementation approaches.
11. Planned verification procedures.
12. Artifact structures and registers.
13. Evidence structures and registers.
14. Traceability matrices.
15. Dependency, risk, decision, deviation, technical-debt, and deferred-item registers.
16. Wave, epic, and story closure criteria.
17. A final completeness and consistency review.

The result should provide a clear and authoritative picture of:

* what must be built;
* why it must be built;
* in what order it must be built;
* what each implementation unit changes;
* how each unit will be tested;
* what evidence must be preserved;
* what artifacts must be produced;
* how progress will be tracked;
* how independent reviewers will verify completion;
* how the framework will move from its current state to the intended high-quality, production-ready state.

The final implementation programme must be detailed enough that implementation work can proceed wave by wave without repeatedly returning to the original source documents to rediscover scope, dependencies, expectations, or closure conditions.
