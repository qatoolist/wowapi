# Fable 5 Final Architecture Review, Framework Capability Assessment and Authoritative Implementation-Planning Mandate

You are the senior architecture reviewer, technical mentor, delivery lead and final quality gatekeeper for this work.

A lower-cost, less-capable model has already completed an initial architecture-review and implementation-planning pass. Its reported completion summary is:

> 10 questions and 2 blocker classes—GitHub repository-administration actions and an unowned reference-performance-environment prerequisite—were recorded in §7 of the plan as requiring human decisions. Nothing was changed in the `wowsociety` repository. All 17 changed or newly created files remain uncommitted.

Treat this statement only as an agent-reported status.

Do not assume that:

* there are exactly 17 changed or newly created files;
* the preliminary plan is complete or technically correct;
* all 10 questions genuinely require human intervention;
* either reported blocker class blocks the complete workstream;
* `wowsociety` is unaffected because its code was not modified;
* passing tests prove architectural correctness;
* generated evidence is valid merely because it exists;
* an existing package, interface or document proves that a framework capability is production-ready;
* the current design must be preserved to avoid changes in the dependent product.

Your responsibility is to independently validate the evidence, review the complete uncommitted work, assess the overall production readiness of `wowapi`, identify missing capabilities, correct the preliminary plan, reassess the alleged blockers, determine the impact on `wowsociety`, and produce one authoritative, dependency-aware and implementation-ready plan and backlog.

The result must reflect your own senior architectural judgement.

---

# 1. Non-negotiable product-priority principle

`wowapi` is the framework and must be designed as the durable, reusable and production-grade foundation.

`wowsociety` is currently in a pilot or early-development stage and is not yet so mature that its current implementation must constrain the framework architecture.

Therefore:

> **Do not compromise the correctness, quality, abstractions, public contracts, security, maintainability or long-term architecture of `wowapi` merely to avoid rework in the current pilot implementation of `wowsociety`.**

At this stage, reasonable rework in `wowsociety` is acceptable where it is required to align with a technically correct framework design.

The dependency direction must remain:

`wowsociety` adapts to stable and correct `wowapi` contracts.

`wowapi` must not inherit weak abstractions, application-specific behaviour, accidental coupling or premature compatibility constraints from the current `wowsociety` pilot.

This does not permit careless breaking changes. Every material change must still be:

* architecturally justified;
* deliberately designed;
* documented;
* tested;
* traceable;
* accompanied by a clear `wowsociety` migration or rework plan.

However, where the choice is between:

1. preserving an immature `wowsociety` implementation; and
2. correcting or strengthening `wowapi`;

choose the technically correct long-term design for `wowapi`.

Do not preserve an inferior public API solely because `wowsociety` already consumes it.

Do not label necessary `wowsociety` pilot rework as a reason to defer a required framework correction.

Do not introduce temporary framework hacks merely to make the current product code continue compiling.

---

# 2. Authoritative source material

Review the following original architecture-review artifacts directly:

* `docs/implementation/architecture-directive-2026-07-11.md`
* `docs/implementation/evidence/architecture-review-2026-07-11/command-log.md`
* `docs/implementation/evidence/architecture-review-2026-07-11/evidence.json`

Also review:

* the preliminary implementation plan produced by the lower-cost model;
* every file created or modified by that model;
* the full document describing the must-have capabilities of a production development framework;
* the present `wowapi` repository;
* the present `wowsociety` repository where available;
* relevant roadmap, product-plan, architecture and implementation documents;
* relevant public APIs, internal APIs, tests, examples, workflows and evidence.

Use repository state as evidence, including:

* `git status`
* `git diff`
* `git diff --stat`
* untracked-file inventory
* relevant commit history
* package structure
* module boundaries
* dependency direction
* public contracts
* internal contracts
* tests
* examples
* CI workflows
* configuration
* migrations
* benchmarks
* generated evidence
* references from the preliminary plan to code, tests, commands and dependent services

Establish the exact changed-file count independently.

The original review artifacts and actual repository state are authoritative.

The lower-cost model’s report and plan are only preliminary material to be challenged and verified.

---

# 3. Repository preservation and safety

During this review:

* Do not commit.
* Do not push.
* Do not open a pull request.
* Do not merge.
* Do not publish a release.
* Do not modify repository settings.
* Do not change branch-protection rules.
* Do not create or modify repository secrets.
* Do not perform repository-administration actions unless explicitly authorised.
* Preserve the current uncommitted state.
* Prefer read-only commands.
* Clearly identify any command that could regenerate, reformat or modify tracked files before executing it.

The goal of this pass is authoritative review and final planning, not unauthorised publication.

---

# 4. Your role

You are not expected to perform all mechanical work yourself.

You are the:

* architecture authority;
* technical mentor;
* delivery lead;
* framework-quality authority;
* task-allocation authority;
* blocker adjudicator;
* public-contract reviewer;
* compatibility and migration reviewer;
* reviewer of subordinate-agent work;
* final approval authority.

You must:

* assign appropriate work to subordinate agents;
* use the least expensive capable model;
* provide clear task boundaries;
* monitor outputs;
* inspect evidence;
* identify misinterpretations;
* detect superficial completion;
* issue corrective guidance;
* require revisions;
* reject unsupported conclusions;
* resolve important technical ambiguity;
* approve the final plan personally.

Delegation does not reduce your responsibility.

Do not blindly merge subordinate-agent outputs.

---

# 5. Cost-effective operating model

Use premium model capability only where it creates material architectural value.

## 5.1 Delegate mechanical work to lower-cost agents

Lower-cost agents should normally perform:

* document extraction;
* review-finding extraction;
* directive inventory;
* command-log extraction;
* conversion of `evidence.json` into structured tables;
* changed-file inventory;
* untracked-file inventory;
* repository file mapping;
* package inventory;
* public-API inventory;
* preliminary dependency searches;
* test inventory;
* documentation inventory;
* capability-checklist extraction;
* first-pass source-to-plan mapping;
* extraction of the 10 questions;
* extraction of the two blocker classes;
* first-pass `wowsociety` usage mapping;
* candidate-library research;
* dependency metadata collection;
* vulnerability-scan collection;
* licence collection;
* safe test and analysis command execution;
* coverage and race-output collection;
* first-pass matrices;
* first-pass backlog drafting;
* formatting;
* completeness checks;
* repetitive evidence compilation.

These agents may produce preliminary synthesis, but their conclusions must not be treated as final architecture decisions.

## 5.2 Fable 5 must personally perform

Fable 5 must personally handle or adjudicate:

* architectural interpretation;
* public API decisions;
* framework-versus-product boundaries;
* kernel-versus-adapter boundaries;
* compatibility strategy;
* deprecation policy;
* migration design;
* security-sensitive decisions;
* transaction and consistency design;
* lifecycle design;
* concurrency-sensitive issues;
* module-boundary decisions;
* adapter design;
* context-propagation design;
* dependency-selection approval;
* build-versus-reuse decisions;
* blocker classification;
* production-readiness judgement;
* capability prioritisation;
* sequencing;
* validation of tests;
* validation of evidence;
* correction of agent errors;
* final `wowsociety` rework strategy;
* final plan approval.

## 5.3 Escalation conditions

Escalate a lower-cost task to Fable 5 only where:

* architecture judgement is required;
* evidence is contradictory;
* ambiguity affects public contracts;
* a lower-cost agent repeatedly fails;
* security is affected;
* data consistency is affected;
* compatibility or migration is affected;
* concurrency is affected;
* the issue has broad framework impact.

Do not use Fable 5 merely because a task is lengthy.

Differentiate between intellectual complexity and mechanical volume.

---

# 6. Claude rate-limit controls

Claude-side rate limits are a material constraint.

Follow these rules:

* Use a maximum of two subordinate agents concurrently.
* Prefer one subordinate agent at a time.
* Never run multiple premium Claude agents concurrently.
* Fable 5 must remain the only flagship architecture authority.
* Do not run agents against overlapping files or substantially overlapping concerns.
* Do not duplicate complete reviews across agents.
* Preserve intermediate results.
* Review completed work before launching dependent work.
* Reduce execution to one subordinate agent immediately if rate-limit symptoms appear.
* Do not repeatedly retry failed parallel calls.

Two-agent concurrency may be used only for clearly independent tasks, such as:

* `wowapi` inventory and separate `wowsociety` usage inventory;
* documentation inventory and test execution;
* non-overlapping source extraction;
* changed-file inventory and library metadata collection.

Do not parallelise:

* architecture decisions;
* API design;
* compatibility strategy;
* migration strategy;
* blocker adjudication;
* final prioritisation;
* final sequencing;
* final sign-off.

For every implementation phase, state:

* allowed concurrency;
* required agent capability;
* parallel-safe tasks;
* sequential tasks;
* review checkpoint.

---

# 7. Delegation-first workflow

## Stage 1: Evidence and repository inventory

Assign a lower-cost evidence analyst to extract:

* every architecture finding;
* every directive;
* every command and result;
* every evidence entry;
* every referenced package, API and file;
* every unresolved assumption;
* every preliminary-plan task;
* every changed or new file;
* the 10 unresolved questions;
* the two blocker classes.

Require precise source references.

## Stage 2: Preliminary-plan comparison

Assign a lower-cost agent to compare the preliminary plan against:

* the architecture directive;
* the command log;
* `evidence.json`;
* repository state;
* current tests;
* current documentation.

The comparison must identify:

* covered findings;
* omitted findings;
* partially addressed findings;
* unsupported recommendations;
* vague tasks;
* missing dependencies;
* missing tests;
* weak evidence;
* incorrect priorities;
* duplicate tasks;
* false blockers;
* missing `wowsociety` impact.

## Stage 3: Whole-framework capability inventory

Assign a lower-cost agent to convert the complete framework-capability document into a structured checklist containing:

* all 30 capability areas;
* all sub-capabilities;
* the 20 minimum mandatory capabilities;
* the central design principle;
* the proposed four framework levels.

The agent should map likely repository evidence but must not declare final readiness.

## Stage 4: Reuse and dependency inventory

Assign a lower-cost agent to identify:

* existing custom implementations;
* mature Go standard-library alternatives;
* mature third-party candidates;
* current dependencies;
* overlapping dependencies;
* abandoned or unsafe dependencies;
* dependency licences;
* vulnerability information;
* public APIs exposing third-party types.

## Stage 5: Technical review by Fable 5

Fable 5 must personally review:

* critical and high-severity findings;
* all mandatory capability conclusions;
* all missing capability conclusions;
* public API implications;
* framework-layer placement;
* security implications;
* lifecycle implications;
* transaction semantics;
* concurrency;
* compatibility;
* migration;
* dependency choices;
* framework scope;
* `wowsociety` rework implications;
* disputed or unsupported agent conclusions.

Where subordinate-agent work is deficient:

1. explain the deficiency;
2. issue specific corrective instructions;
3. require revision;
4. review the revised output;
5. do not advance the work package until corrected.

## Stage 6: Plan construction

Use lower-cost agents to draft:

* findings register;
* capability matrix;
* changed-file table;
* task register;
* test matrix;
* evidence matrix;
* dependency register;
* build-versus-reuse matrix;
* `wowsociety` impact matrix;
* risk register;
* decision register;
* traceability matrix;
* phase plan.

Fable 5 must then personally correct and approve:

* architecture rationale;
* scope;
* layer placement;
* dependency direction;
* task dependencies;
* compatibility decisions;
* migration sequencing;
* acceptance criteria;
* test strategy;
* framework-versus-product allocation;
* final priorities.

## Stage 7: Independent completeness audit

Assign a separate cost-effective reviewer to identify:

* missing findings;
* broken traceability;
* missing capability rows;
* missing mandatory tasks;
* vague acceptance criteria;
* incomplete evidence;
* contradictory sequencing;
* missing `wowsociety` work;
* unsupported production-readiness claims.

Fable 5 must adjudicate the audit.

---

# 8. Review every changed or newly created file

Inspect the complete uncommitted change set.

For every file, determine:

* its purpose;
* originating finding;
* whether it is correct;
* whether it reflects repository reality;
* whether it duplicates another artifact;
* whether it contains unsupported assumptions;
* whether it belongs in the repository;
* whether it belongs at its current location;
* whether it should be retained, revised, moved, merged, split or removed;
* whether it introduces unnecessary scope;
* whether its completion claims are supported.

Produce a changed-file disposition table containing:

* path;
* status;
* purpose;
* originating source;
* quality assessment;
* issues;
* required correction;
* retain/revise/remove/move/merge decision;
* assigned agent capability;
* verification;
* resulting task ID.

Do not review only the main plan.

---

# 9. Critically review the preliminary plan

For every section and task, determine whether it is:

* correct and complete;
* correct but under-specified;
* misunderstood;
* improperly prioritised;
* missing dependencies;
* too broad;
* over-fragmented;
* missing acceptance criteria;
* missing fail-first tests;
* missing regression coverage;
* missing evidence;
* assigned to the wrong repository;
* assigned to the wrong framework layer;
* unsupported by evidence;
* duplicative;
* unnecessary;
* based on a false blocker;
* incorrectly constrained by the current `wowsociety` pilot.

Produce a preliminary-plan review table containing:

* plan item;
* originating evidence;
* assessment;
* issue;
* correction;
* final disposition;
* resulting task IDs.

Explicitly identify:

* accepted items;
* strengthened items;
* corrected items;
* rejected items;
* entirely missed items.

---

# 10. Reassess the 10 unresolved questions

Classify each question as:

1. Answerable from repository evidence.
2. Answerable through technical analysis.
3. Answerable through testing.
4. A decision Fable 5 should make.
5. A `wowsociety` product decision.
6. A repository-administration decision.
7. An infrastructure decision.
8. A genuine human or business decision.
9. An unnecessary question created by incomplete analysis.

For each question:

* restate it;
* explain why it exists;
* identify evidence;
* provide the recommended decision;
* provide a safe default;
* explain alternatives;
* explain consequences;
* identify the owner;
* state whether implementation can proceed;
* identify independent work;
* record final disposition.

Reduce human-dependent questions to the smallest genuine set.

Do not leave architecture decisions unresolved merely to transfer responsibility to the user.

---

# 11. Reassess GitHub repository-administration blockers

Separate:

* implementation;
* local validation;
* workflow authoring;
* CI validation;
* branch protection;
* status-check configuration;
* repository permissions;
* secrets;
* environments;
* merge controls;
* release controls;
* rollout enforcement.

For every item, identify:

* work possible locally;
* work possible without admin access;
* exact administrator action;
* required permissions;
* owner;
* whether it blocks implementation, verification, enforcement or rollout;
* operator instructions;
* post-action evidence;
* interim default.

Do not label complete workstreams blocked where only final activation requires repository ownership.

---

# 12. Reassess the reference-performance-environment blocker

Determine:

* which benchmarks require stable hardware;
* which can run locally;
* which can run in normal CI;
* which can use relative comparison;
* which can use containerised workloads;
* which can begin as advisory;
* which require dedicated scheduled execution.

Produce a reference-environment specification covering:

* owner;
* environment type;
* CPU;
* memory;
* storage;
* OS;
* architecture;
* Go and toolchain versions;
* dataset;
* workload;
* warm-up;
* repetitions;
* aggregation method;
* tolerance;
* baseline storage;
* advisory threshold;
* failure threshold;
* CI or scheduled strategy;
* evidence retention;
* calibration process.

Where ownership requires a human decision, recommend the owner and define a provisional approach.

Do not allow this issue to block unrelated correctness, testing, profiling or benchmark-development work.

---

# 13. Mandatory whole-framework capability assessment

Assess `wowapi` against the entire capability baseline, not only the central design-principle section.

Review all 30 capability areas:

1. Clear application structure
2. Dependency management and inversion of control
3. Configuration management
4. Application lifecycle management
5. Error handling
6. Logging
7. Observability
8. Security foundations
9. Validation
10. Transport abstraction
11. Routing and middleware
12. Data-access integration
13. Transaction and consistency support
14. Background processing
15. Resilience
16. Testing support
17. Developer tooling
18. API contract management
19. Extensibility
20. Modularity
21. Compatibility and upgradeability
22. Performance controls
23. Caching
24. Events and messaging
25. Multi-tenancy and contextual execution
26. Internationalisation and time handling
27. File and object-storage integration
28. Auditability and compliance
29. Deployment and operational readiness
30. Documentation and governance

Review all listed sub-capabilities.

Do not treat the headings as sufficient coverage.

---

# 14. Minimum mandatory capability assessment

Independently assess:

1. Clear architecture and module structure
2. Dependency injection or equivalent decoupling
3. Typed configuration with validation
4. Structured error handling
5. Structured logging
6. Metrics, tracing and health checks
7. Authentication and authorization integration
8. Input validation and secure defaults
9. Graceful startup and shutdown
10. Database and transaction integration
11. Timeouts, cancellation and resilience controls
12. Unit, integration and functional testing support
13. API contract and versioning support
14. Background job and messaging integration
15. Extension and plugin mechanisms
16. Static analysis, linting and developer tooling
17. Deployment readiness
18. Upgrade, compatibility and deprecation policies
19. Auditability
20. Complete documentation

For each, classify `wowapi` as:

* ready;
* conditionally ready;
* not ready;
* not applicable, with justification.

A capability is not ready merely because:

* a package exists;
* an interface exists;
* documentation mentions it;
* a roadmap item exists;
* one happy-path test passes;
* a third-party library could theoretically provide it.

---

# 15. Capability classification model

Classify capabilities and sub-capabilities as:

## A. Production-ready

Requires:

* stable abstraction;
* correct implementation or adapter contract;
* lifecycle integration;
* secure defaults;
* configuration validation;
* meaningful tests;
* negative-path coverage;
* documentation;
* examples;
* compatibility expectations;
* operational evidence.

## B. Implemented but incomplete

A material aspect is missing.

## C. Partial or experimental

Implementation exists but must not be presented as stable.

## D. Planned but not implemented

Exists only in documents or backlog.

## E. Missing and required

Required for the intended production framework but absent.

## F. Adapter responsibility

The framework should provide a stable integration boundary rather than the complete vendor implementation.

## G. Optional or use-case dependent

Must integrate safely without becoming a mandatory kernel dependency.

## H. Explicit non-goal

Outside scope with a documented rationale and recommended integration path.

Do not use “optional” or “non-goal” to hide a missing production foundation.

---

# 16. Cross-capability review

Review whether capabilities work together correctly.

Examples:

* lifecycle across servers, workers, queues and databases;
* cancellation across transports and adapters;
* retries with idempotency and transactions;
* logging, tracing, errors and audit correlation;
* security across all supported transports;
* tenant context in background jobs;
* configuration validation for optional modules;
* outbox integration with transactions and workers;
* graceful shutdown with in-flight jobs;
* API compatibility with configuration and migration compatibility;
* testing of deterministic clocks, IDs, retries and concurrency;
* readiness probes during dependency failure;
* error translation across transport boundaries.

Identify capabilities that exist independently but fail when combined.

---

# 17. Central design principle

Assess whether `wowapi` actually provides:

> **Strong defaults, stable abstractions, enforceable conventions and replaceable adapters.**

## Strong defaults

The safe, maintainable and observable option should be the easiest default.

## Stable abstractions

Public contracts must represent durable framework concepts rather than current vendor implementations.

## Enforceable conventions

Conventions should be enforced through:

* package boundaries;
* constructors;
* configuration validation;
* registration;
* static analysis;
* linting;
* code generation;
* CI;
* tests;
* restricted public APIs.

Documentation alone is insufficient.

## Replaceable adapters

Infrastructure implementations must be replaceable without rewriting business logic.

Identify leakage of:

* vendor types;
* database-driver types;
* concrete loggers;
* tracing SDK objects;
* HTTP framework request types;
* queue-client messages;
* cloud SDK objects;
* application-domain types;
* global state.

---

# 18. Validate the four-level architecture

Assess current packages and proposed work against:

## Kernel

* lifecycle;
* configuration;
* dependency management;
* errors;
* modules;
* stable context and contracts.

The kernel must remain small and stable.

## Application foundation

* validation;
* domain execution;
* transaction boundaries;
* security-policy integration;
* domain boundaries;
* application commands.

## Infrastructure adapters

* transports;
* databases;
* messaging;
* queues;
* caching;
* files;
* identity;
* external clients.

## Operational foundation

* testing;
* logging;
* metrics;
* tracing;
* resilience;
* diagnostics;
* deployment;
* performance governance;
* compatibility governance;
* documentation and release governance.

Identify:

* kernel bloat;
* vendor leakage;
* domain-specific behaviour in framework code;
* circular dependencies;
* unstable dependencies flowing inward;
* missing architectural homes;
* optional concerns made mandatory;
* operational capabilities treated as afterthoughts.

---

# 19. Mandatory principle: Do not reinvent the wheel

A foundational rule is:

> **Do not build custom framework functionality where the Go standard library or a mature, secure, actively maintained and widely adopted library already solves the problem well.**

`wowapi` should add value through:

* stable abstractions;
* lifecycle integration;
* enforceable conventions;
* configuration;
* observability;
* security policy;
* error translation;
* testing support;
* replaceable adapters;
* coherent developer experience.

It should not duplicate mature general-purpose engineering work.

## 19.1 Reuse-before-build decision order

Use this order:

1. Use the Go standard library where sufficient.
2. Use a mature and safe third-party library where needed.
3. Create a narrow framework integration or adapter where framework policy or stability is required.
4. Build custom functionality only where no suitable mature solution exists or existing solutions fundamentally conflict with framework requirements.

Custom implementation is the exception.

## 19.2 Areas requiring reuse evaluation

Evaluate existing solutions for:

* logging;
* configuration;
* validation;
* dependency injection;
* routing;
* middleware;
* authentication;
* authorization;
* metrics;
* tracing;
* OpenTelemetry;
* health checks;
* retry;
* backoff;
* circuit breakers;
* rate limiting;
* caching;
* database access;
* migrations;
* transactions;
* background jobs;
* messaging;
* schema validation;
* API generation;
* contract compatibility;
* test containers;
* mocking;
* property testing;
* static analysis;
* cryptography;
* secrets integration;
* object storage;
* internationalisation;
* vulnerability scanning;
* supply-chain verification.

## 19.3 Library evaluation

Evaluate:

* exact functional fit;
* maintenance activity;
* maintainers;
* production adoption;
* API stability;
* release quality;
* documentation;
* tests;
* security history;
* open advisories;
* transitive dependencies;
* supply-chain risk;
* licence;
* Go-version compatibility;
* performance;
* context and cancellation support;
* concurrency behaviour;
* observability;
* replaceability;
* abandonment risk.

Do not approve a library based only on GitHub stars.

## 19.4 Security gate

Reject dependencies with:

* unmitigated critical or high vulnerabilities;
* abandoned maintenance;
* unclear ownership;
* suspicious releases;
* incompatible licensing;
* unsafe defaults that cannot be controlled;
* excessive opaque dependencies;
* unnecessary install-time execution;
* unexplained network, filesystem or process privileges;
* unstable APIs that would leak into public contracts.

Never reimplement:

* cryptography;
* password hashing;
* token signing;
* certificate validation;
* OAuth;
* OpenID Connect;
* secure random generation;
* encryption protocols;
* low-level authorization engines;

without extraordinary justification and specialist review.

## 19.5 Avoid dependency sprawl

“Do not reinvent the wheel” does not mean adding a dependency for trivial work.

Prefer the standard library where adequate.

Avoid:

* multiple libraries for the same responsibility;
* dependencies larger than the problem;
* wrappers that add no value;
* uncontrolled transitive dependencies;
* optional dependencies forced into the kernel;
* provider-specific APIs exposed as permanent framework contracts.

## 19.6 Review existing custom code

Identify existing custom implementations that may duplicate mature solutions, including:

* retry logic;
* validation;
* configuration parsing;
* schedulers;
* caching;
* routing;
* middleware;
* dependency injection;
* observability;
* security utilities;
* migrations;
* infrastructure clients;
* copied or forked code.

For each, determine whether to:

* retain;
* simplify;
* replace with the standard library;
* replace with a mature dependency;
* place behind an adapter;
* deprecate;
* remove.

Do not replace custom code automatically. Consider compatibility, performance, migration cost, security and long-term maintenance.

---

# 20. `wowsociety` impact and permitted rework

Perform a complete dependency-impact review.

Identify:

* framework APIs consumed by `wowsociety`;
* behaviours relied upon;
* workarounds for framework limitations;
* duplicated framework concerns;
* direct infrastructure coupling;
* current public-type dependencies;
* configuration dependencies;
* error-behaviour dependencies;
* middleware-order dependencies;
* transaction assumptions;
* lifecycle assumptions;
* observability assumptions;
* testing assumptions.

For every framework task, classify `wowsociety` impact as:

* none;
* verification only;
* later adoption opportunity;
* test update;
* configuration update;
* code adaptation;
* refactor;
* adapter migration;
* data migration;
* breaking change;
* coordinated rollout;
* targeted investigation required.

## 20.1 Framework-first rule

When a framework correction requires `wowsociety` rework:

* preserve the correct `wowapi` design;
* create explicit `wowsociety` migration tasks;
* update product tests;
* update configuration;
* update adapters;
* remove obsolete product workarounds;
* validate end-to-end behaviour.

Do not weaken `wowapi` to avoid pilot-product rework.

## 20.2 Compatibility policy for the pilot stage

Because `wowsociety` remains in pilot:

* reasonable breaking changes may be accepted;
* immature APIs may be corrected;
* package boundaries may be improved;
* configuration may be redesigned;
* product workarounds may be removed;
* integration code may be rewritten.

However, every breaking change must still include:

* rationale;
* affected usage;
* migration instructions;
* product task IDs;
* validation;
* rollout sequence;
* rollback considerations.

Breaking changes must be deliberate, not accidental.

## 20.3 Do not move product code into the framework automatically

A concern should move into `wowapi` only where it is:

* domain-independent;
* reusable;
* stable;
* appropriate for a framework;
* consistent with the four-level architecture.

Do not generalise housing-society-specific concepts into the framework.

---

# 21. Testing requirements

Do not accept vague statements such as:

* “add tests”;
* “tests pass”;
* “improve coverage”;
* “verify compatibility.”

For every material task, define:

* the defect or risk;
* fail-first reproduction;
* expected pre-fix failure;
* expected post-fix behaviour;
* unit tests;
* integration tests;
* contract tests;
* functional tests;
* regression tests;
* negative-path tests;
* race tests;
* concurrency tests;
* migration tests;
* configuration tests;
* compatibility tests;
* fault-injection tests;
* security tests;
* benchmarks;
* `wowsociety` end-to-end tests;
* exact commands;
* expected results;
* retained evidence.

Existing tests may encode defective behaviour.

Review test logic independently.

Passing tests are necessary but not sufficient.

---

# 22. Evidence requirements

For every task, define evidence such as:

* changed-file list;
* patch;
* test output;
* coverage output;
* race-detector output;
* lint output;
* static-analysis output;
* benchmark output;
* dependency graph;
* vulnerability report;
* licence report;
* API compatibility report;
* migration dry-run;
* before-and-after reproduction;
* `wowsociety` validation;
* documentation;
* ADR;
* reviewer sign-off.

Validate that:

* commands genuinely ran;
* output matches the reviewed revision;
* tests verify the intended behaviour;
* evidence is complete;
* failures were not omitted;
* benchmarks are comparable.

---

# 23. Required capability matrix

Produce a matrix containing:

* capability ID;
* area;
* sub-capability;
* mandatory or optional;
* intended framework level;
* current status;
* current location;
* public contract;
* implementation evidence;
* lifecycle evidence;
* configuration evidence;
* test evidence;
* documentation evidence;
* security assessment;
* extensibility assessment;
* operational-readiness assessment;
* compatibility assessment;
* reuse-versus-build decision;
* `wowsociety` usage;
* gap;
* risk;
* disposition;
* backlog task IDs.

Cover the actual sub-capabilities, not only the headings.

---

# 24. Build-versus-reuse matrix

For every material capability gap, record:

* requirement;
* standard-library option;
* mature library candidates;
* existing internal implementation;
* functional fit;
* maintenance;
* security;
* licence;
* dependency cost;
* public-API leakage risk;
* integration effort;
* migration impact;
* `wowsociety` impact;
* decision:

  * standard library;
  * direct library reuse;
  * adapter around a library;
  * retain internal implementation;
  * replace internal implementation;
  * custom build;
* justification;
* approval authority.

No material custom subsystem may be approved without a reuse assessment.

---

# 25. Required backlog structure

Produce one consolidated backlog.

Do not maintain separate task lists for:

* original review findings;
* preliminary-plan corrections;
* capability gaps;
* blockers;
* dependency changes;
* `wowsociety` rework;
* changed-file corrections.

Deduplicate overlapping work while preserving traceability.

Every task must include:

* task ID;
* title;
* originating findings;
* originating capability;
* preliminary-plan reference;
* changed-file reference;
* problem statement;
* architectural rationale;
* target framework level;
* exact scope;
* explicit non-scope;
* affected components;
* dependencies;
* implementation order;
* assigned agent capability;
* allowed concurrency;
* implementation guidance;
* reuse-versus-build decision;
* selected dependency where applicable;
* public-contract impact;
* compatibility impact;
* migration impact;
* rollback considerations;
* `wowsociety` impact and rework;
* fail-first verification;
* tests;
* security verification;
* performance verification;
* documentation;
* evidence;
* acceptance criteria;
* reviewer;
* review gate;
* rollout;
* severity;
* priority;
* phase;
* status.

Tasks must be assignable and independently verifiable.

---

# 26. Priority model

## P0 — Correctness or safety blocker

Examples:

* data-loss risk;
* transaction corruption;
* security bypass;
* secret exposure;
* unsafe shutdown;
* resource leak;
* severe concurrency defect;
* consistency failure;
* broken public contract.

## P1 — Required production foundation

Examples:

* missing mandatory capability;
* unstable abstraction;
* missing configuration validation;
* missing lifecycle foundation;
* missing error foundation;
* missing testing foundation;
* missing observability;
* missing compatibility policy;
* missing deployment safety;
* missing dependency-direction enforcement.

## P2 — Important framework maturity

Examples:

* stronger tooling;
* improved adapters;
* enhanced resilience;
* performance governance;
* richer compliance support;
* stronger contract management.

## P3 — Optional ecosystem capability

Examples:

* optional transports;
* specialised caching;
* optional multi-tenancy;
* optional storage adapters;
* convenience tooling.

Priority must be based on architecture dependency and production risk.

Do not lower framework priorities merely to reduce `wowsociety` rework.

---

# 27. Required phases and gates

Define phases such as:

1. Evidence and repository validation
2. Preliminary-plan review
3. Mandatory capability assessment
4. Architecture and layer approval
5. Public-contract and migration design
6. Foundational implementation
7. Adapter implementation
8. `wowsociety` alignment and rework
9. Targeted tests
10. Full regression and race validation
11. Performance validation
12. Evidence audit
13. Documentation and migration completion
14. Final architecture sign-off

For each phase include:

* objectives;
* task IDs;
* dependencies;
* concurrency;
* agent capability;
* tests;
* evidence;
* `wowsociety` work;
* reviewer;
* entry criteria;
* exit criteria.

---

# 28. Required final deliverables

Produce:

## A. Executive architecture assessment

State:

* current condition;
* highest risks;
* missing mandatory capabilities;
* main preliminary-review defects;
* true blockers;
* whether `wowapi` is currently production-ready.

## B. Repository-state inventory

Account for all changed, new and untracked files.

## C. Changed-file disposition table

Review every file.

## D. Preliminary-plan quality review

Identify what was correct, weak, wrong, missing or unnecessary.

## E. Consolidated findings register

Include evidence, severity, implication and tasks.

## F. Resolution of the 10 questions

Reduce them to the smallest genuine decision set.

## G. Blocker-resolution plan

Separate implementation, validation, enforcement and rollout blockers.

## H. Complete capability matrix

Cover all 30 capability areas and sub-capabilities.

## I. Mandatory-capability readiness assessment

Classify all 20 mandatory capabilities.

## J. Four-level architecture map

Map current packages and proposed tasks to:

* Kernel;
* Application foundation;
* Infrastructure adapters;
* Operational foundation.

## K. Reuse-opportunity register

Identify custom work that may be replaced or simplified.

## L. Approved dependency register

Include:

* dependency;
* version;
* purpose;
* layer;
* licence;
* maintenance status;
* security status;
* alternatives;
* rationale;
* upgrade policy;
* replacement strategy.

## M. Rejected dependency register

Record material rejected candidates and reasons.

## N. Final phased implementation plan

Provide the dependency-aware programme.

## O. Detailed task register

Provide complete implementation-ready tasks.

## P. `wowapi`-to-`wowsociety` impact and rework matrix

Map every relevant framework change to:

* product impact;
* required rework;
* migration;
* tests;
* rollout sequence.

## Q. Test and evidence matrix

Map:

`Finding/capability → Task → Fail-first test → Command → Expected result → Evidence`

## R. Agent-allocation plan

Include:

* work package;
* agent capability;
* reason;
* Fable 5 involvement;
* concurrency;
* review authority;
* escalation condition.

## S. Traceability matrix

Map:

`Capability → Sub-capability → Evidence → Gap → Original finding → Preliminary item → Changed file → Final task → Test → Evidence → wowsociety rework → Review gate`

## T. Risk register

Include technical, security, compatibility, migration, performance, delivery, evidence and product risks.

## U. Decision register

Include recommendation, alternatives, owner and safe default.

## V. Backlog quality audit

Confirm that:

* every finding has a disposition;
* every changed file has a disposition;
* all questions were reassessed;
* blockers were constrained;
* all capability areas were reviewed;
* all mandatory gaps have tasks or justified dispositions;
* reuse assessments were performed;
* unsafe dependencies were rejected;
* tasks have objective acceptance criteria;
* tests and evidence are defined;
* dependencies are sequenced;
* `wowsociety` rework is represented;
* optional capabilities do not pollute the kernel;
* traceability is complete.

---

# 29. Required final answers

Explicitly answer:

1. Is `wowapi` currently production-capable?
2. Which architecture findings remain unresolved?
3. What did the lower-cost reviewer miss or misunderstand?
4. Which questions genuinely require human decisions?
5. Which GitHub actions genuinely require admin access?
6. What performance work can proceed immediately?
7. Which mandatory capabilities are missing?
8. Which capabilities are present but immature?
9. Which should be adapter contracts?
10. Which should remain optional?
11. Which proposals would bloat the framework?
12. Is the four-level architecture correct?
13. Which packages are in the wrong layer?
14. Which custom components should be replaced by mature libraries?
15. Which dependencies should be approved or rejected?
16. What is the minimum correctly sequenced production-readiness backlog?
17. What `wowsociety` rework is required?
18. Which `wowsociety` workarounds should be removed?
19. Which breaking changes are justified at the pilot stage?
20. How will completion be objectively demonstrated?
21. Which work genuinely requires Fable 5?
22. Which work must remain assigned to lower-cost agents?

---

# 30. Final approval gate

Do not approve the final plan unless:

* repository state is independently verified;
* every changed file is reviewed;
* every original finding is reconciled;
* the preliminary plan is critically assessed;
* the 10 questions are reduced appropriately;
* the blocker classes are accurately constrained;
* all 30 capability areas are assessed;
* all 20 mandatory capabilities are assessed;
* cross-capability behaviour is reviewed;
* the four-level architecture is validated;
* the reuse-before-build assessment is complete;
* unsafe dependencies are rejected;
* foundational work precedes dependent features;
* `wowsociety` rework is explicitly planned;
* `wowapi` has not been compromised to preserve pilot code;
* backlog tasks are implementation-ready;
* acceptance criteria are objective;
* meaningful tests are defined;
* evidence is credible;
* kernel scope remains controlled;
* vendor details do not leak into stable contracts;
* agent allocation is cost-effective;
* Claude concurrency is rate-limit safe;
* Fable 5 has personally approved every material architecture decision.

---

# 31. Final conduct rules

* Do not cosmetically rewrite the preliminary plan.
* Do not trust agent completion claims without evidence.
* Do not treat package names as proof of capability.
* Do not treat documentation as implementation.
* Do not treat passing existing tests as proof of correctness.
* Do not leave architecture questions unanswered unnecessarily.
* Do not classify administrator-only rollout work as an implementation blocker.
* Do not use Fable 5 for mechanical extraction.
* Do not duplicate expensive review effort.
* Do not reinvent mature solutions.
* Do not select libraries based only on popularity.
* Do not use abandoned or unsafe dependencies.
* Do not introduce dependencies for trivial standard-library functionality.
* Do not reimplement security-sensitive primitives.
* Do not create wrappers that add no value.
* Do not allow optional features to bloat the kernel.
* Do not preserve an inferior `wowapi` abstraction merely to avoid `wowsociety` pilot rework.
* Do not introduce temporary framework hacks for current product compatibility.
* Do not make breaking changes without explicit migration and validation tasks.
* Do not move application-domain logic into the framework.
* Do not claim production readiness without implementation, tests, documentation, compatibility and operational evidence.
* Do not commit, push or modify repository settings without explicit authorisation.

The intended outcome is:

> **A technically uncompromised, reusable and production-grade `wowapi` framework, built using mature and safe existing solutions wherever appropriate, with `wowsociety` deliberately reworked where necessary to adopt the corrected framework design.**

Produce one consolidated, authoritative implementation plan and backlog that another engineering team can execute without having to reinterpret vague recommendations.
