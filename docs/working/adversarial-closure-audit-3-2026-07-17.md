• # Independent closure verdict

  Reviewer verdict: NO-PUSH. Closure is rejected at 9bfa0dd.

  F-04 and the behavioral portion of F-05 are genuinely fixed. However, F-10 still has two High-severity bypasses, and F-05
  introduced a Medium compatibility regression. The statement that “every Rejection2.md finding is closed” is therefore not
  supported.

  ## Confirmed remaining findings

   Severity      Classification        Finding
  ━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   High          Confirmed Defect      Workflow conditions retain mutable nested aliases after boot
  ────────────  ────────────────────  ──────────────────────────────────────────────────────────────────────────
   High          Confirmed Defect      Seed declarations remain replaceable and mutable after boot
  ────────────  ────────────────────  ──────────────────────────────────────────────────────────────────────────
   Medium        Confirmed Defect      UploadEvent change breaks stable-v1 source compatibility
  ────────────  ────────────────────  ──────────────────────────────────────────────────────────────────────────
   Medium        Confirmed Defect      Generated API can consume a replaced post-boot I18n catalog
  ────────────  ────────────────────  ──────────────────────────────────────────────────────────────────────────
   Medium/Low    Missing Safeguards    Migration filesystem aliases and hand-constructed Booted bypasses remain
  ────────────  ────────────────────  ──────────────────────────────────────────────────────────────────────────
   Low           Missing Safeguards    Several registry declarations still lack nil/empty validation

  ### 1. High — workflow condition alias survives sealing

  The workflow clone copies Branch.When by value, but kernel/workflow/definition.go:171 is any. If it contains a map, slice,
  pointer, or another mutable object, the clone retains the original reference.

  The runtime later consumes that value during gateway selection at kernel/workflow/runtime.go:626.

  - Violated invariant: booted workflow definitions must be isolated from module-owned declaration memory.
  - Trigger: register a condition whose Equals contains a mutable map or slice, boot, then mutate the retained value.
  - Impact: gateway routing can change after validation; concurrent mutation can race runtime readers.
  - Why tests missed it: the alias regression covers steps, transitions, policies, and assignees, but not the nested
    Branch.When.Equals value.

  - Smallest fix: restrict Equals to validated immutable scalar types, or recursively clone an explicitly supported JSON-like
    value set.

  - Required regression: mutate a retained map/slice after boot and prove gateway routing remains unchanged; run it under
    -race.

  The audit addendum’s claim that workflow branches are completely deep-copied is presently inaccurate.

  ### 2. High — seed state remains live after boot

  app/boot.go:47 is not captured in the protected runtime view.

  It is consumed directly by:

  - The generated migrator at internal/cli/templates/init/cmd_migrate_main.go.tmpl:175.
  - The live readiness path at app/health.go:49.

  seeds.Bundle contains mutable slices and nested slices.

  - Violated invariant: generated products must operate on the validated, boot-sealed seed catalog.
  - Trigger: replace Booted.Seeds, or mutate retained nested seed slices after boot.
  - Impact: derived products can apply a different catalog from the one boot validated; readiness hashes can change or race
    concurrently.

  - Smallest fix: deep-clone seeds into runtimeView, add a copy-returning RuntimeSeeds, and use it in both migration and
    readiness consumers.

  - Required regressions: replace the public field and mutate every nested declaration collection after boot; prove migration
    input and readiness output remain unchanged under -race.

  This is a concrete continuation of the rejected F-10 bypass class.

  ### 3. Medium — UploadEvent breaks stable-v1 source compatibility

  foundation/document/hooks.go:28 gained DeliveryID and Tx.

  Existing external code using an unkeyed literal now fails compilation:

  document.UploadEvent{
      doc, class, version, key, mime, size, sensitivity,
  }

  That is the same compatibility class correctly addressed for app.Hook. Calling the new fields “additive” in the changelog is
  misleading for positional struct literals.

  - Violated invariant: stable-v1 public APIs must remain source compatible.
  - Trigger: compile an existing derived project containing the old seven-field unkeyed literal.
  - Impact: downstream build failure.
  - Smallest safe fix: preserve the seven-field event and introduce a separately versioned transactional event/hook interface,
    or provide transaction and delivery metadata through a compatibility-preserving delivery context.

  - Required regression: an external-consumer compile fixture using the original positional literal.

  The transaction-bound behavior itself is sound; this finding concerns its public API delivery mechanism.

  ### 4. Medium — I18n catalog replacement still reaches generated products

  app/boot.go:53 remains assignable and is not captured in runtimeView.

  The generated API passes the public field directly to middleware at internal/cli/templates/init/cmd_api_main.go.tmpl:276.

  - Violated invariant: generated runtime consumers must use boot-validated internal state.
  - Trigger: replace booted.I18n after boot but before middleware construction.
  - Impact: the product serves a catalog different from the validated catalog.
  - Smallest fix: capture the frozen catalog in runtimeView, add RuntimeI18n, and update the generated template.
  - Required regression: replace the public field after boot and prove a generated consumer still uses the captured catalog.

  ## Additional residual gaps

  These do not outweigh the High findings, but should be recorded:

  - RuntimeMigrations copies the outer map only. A mutable fs.FS, such as fstest.MapFS, may still alter migration content
    after boot.

  - Public StartWorker accepts a hand-constructed Booted; the runtime.set == false fallback can consume registries that never
    passed App.Boot.

  - Migration and seed declarations accept nil fs.FS values.
  - Health declarations accept empty names and nil checks.
  - integration.Registry.Register can dereference a nil or typed-nil provider instead of collecting a boot error.
  - internal/sealer provides an out-of-module boundary, not an unforgeable in-repository authority. The narrower guarantee
    should remain explicit.

  ## Confirmed closures

  The review does confirm substantial, real remediation:

  - F-04 is closed. Cancel, recordFailure, and ReclaimStalled now serialize aggregate state correctly. The commit-window tests
    discriminate the missing-lock behavior.

  - F-05’s state/transaction invariant is closed. Object I/O occurs before the document lock; active state is rechecked under
    lock; transaction-bound effects roll back with confirmation; retries retain a stable delivery identity.

  - Router, Events, Jobs, Health, Recurring, and the migration outer map are protected against public-field replacement when
    the Booted value came from App.Boot.

  - Recurring declaration validation and nil document-hook rejection are working.
  - The named authz, document, rules, and notification declaration clones cover their tested mutable fields.

  ## Verification results

  The authoritative make ci-container completed with exit 0, including:

  - go vet ./...
  - Boundary and lifecycle lint
  - Full unit suite
  - Full go test -race ./...
  - Benchmark budgets
  - Full build

  Focused F-04/F-05 tests also passed against real PostgreSQL/S3 and under -race.

  These green results validate the covered fixes. They do not exercise the remaining workflow alias, seed replacement, I18n
  replacement, or old positional UploadEvent compilation cases.

  ## Graphify status

  Graphify was refreshed from the current branch using Google Gemini semantic extraction:

  - 31 changed code files and 4 changed documentation files processed
  - 2,221 unchanged files reused
  - 5,275 nodes
  - 12,123 edges
  - 495 communities
  - Estimated semantic extraction cost: approximately $0.012
  - Seven inherited extraction warnings remain for edges missing source_file

  The graph is current for 9bfa0dd.

  ## Repository state

  No tracked files were modified by this review. The only visible worktree entries remain the user-owned untracked files:

  ?? Rejection.md
  ?? Rejection2.md

  Nothing was pushed and no PR was opened. The branch should remain local until the two High findings and the stable-v1
  compatibility regression are remediated and covered by discriminating tests.



  ===========================
    Advise
  ===========================

  Question -- based on failures, what pattern do you see in implementation thats causing these failures? what development recommendations would you give to ensure each one of these and similar issues are handled well in advance?

  Answer -
  • The dominant pattern is an incomplete ownership boundary.

  The implementation often validates and “seals” the obvious registry object, but does not establish exclusive ownership of
  the complete transitive object graph or force every runtime consumer through the sealed representation. This produces fixes
  that close the demonstrated path while leaving structurally equivalent paths open.

  In short:

  > Boot is treated as a validation event, but not consistently as an ownership-transfer and immutable-snapshot boundary.

  ## Recurring implementation patterns

  ### 1. Shallow immutability is being mistaken for deep immutability

  Several fixes copy a struct, slice, or outer map, but nested mutable objects remain shared:

  - Branch.When is copied, while Condition.Equals any can retain a map, slice, or pointer.
  - The migration map is copied, while its fs.FS implementations may remain mutable.
  - Seed bundles contain nested slices that can still be retained or replaced.

  This is a Go-specific trap: copying a struct does not copy the data referenced by its slices, maps, pointers, interfaces, or
  mutable interface implementations.

  The real invariant should be:

  > No runtime-reachable value may share mutable storage with module-owned declaration input.

  That is stronger than “we called clone()” or “the registry is sealed.”

  ### 2. There are two sources of truth after boot

  Booted contains public informational fields while an internal runtimeView contains protected state. Some consumers use the
  runtime view, but other consumers—including generated templates—still read the public fields.

  That creates split-brain state:

  Validated internal state ── used by some framework paths
  Public Booted fields     ── used by other framework/generated paths

  Examples include seeds and I18n. The field may have been correct at boot, but replacing it afterward changes what a
  downstream consumer sees.

  A public mirror is safe only if runtime code never treats it as authoritative.

  ### 3. Invariants are enforced by convention rather than representation

  Several APIs permit states the framework does not actually support:

  - Condition.Equals any permits arbitrary mutable and non-comparable values.
  - A public Booted value can be constructed without App.Boot.
  - A fallback exists when the internal boot marker is not set.
  - Registries accept nil callbacks, nil filesystems, empty names, or typed-nil implementations until something later
    dereferences them.

  - A public struct is expanded even though stable-v1 compatibility is expected.

  The code depends on developers remembering rules that types and constructors do not enforce.

  This is why fixes tend to be path-specific: the illegal state continues to be representable, so another caller eventually
  finds it.

  ### 4. Fixes are sometimes moving a defect across architectural boundaries

  The F-05 transactional hook work fixed the database invariant, but placing Tx and DeliveryID directly onto an established
  public event caused a compatibility defect.

  That is a “risk displacement” pattern:

  Atomicity defect fixed
          ↓
  Public event expanded
          ↓
  Downstream source compatibility broken

  The local workflow is now correct, but the framework contract is not.

  Framework fixes need two reviews:

  1. Does this close the runtime invariant?
  2. Does the mechanism preserve API, ownership, layering, and generated-project contracts?

  ### 5. Tests enumerate known fields instead of testing the invariant generically

  The mutation tests covered several nested workflow fields, but omitted Condition.Equals. That means the test encoded a list
  of previously identified aliases rather than the underlying rule:

  > Mutating any registration input after boot cannot alter runtime behavior.

  Similarly, current-repository compilation cannot detect that an older derived project’s positional struct literal no longer
  compiles.

  The tests are often strong for the exact rejected example but insufficiently systematic for the defect class.

  ### 6. Generated projects are treated as secondary consumers

  Framework code was updated to use protected runtime accessors, while generated templates still consumed public Booted
  fields.

  Generated code is part of the product’s public behavior. Every public runtime API effectively has at least three consumer
  classes:

  - Framework-internal consumers
  - Generated-project consumers
  - Independently maintained derived projects

  A fix is incomplete until all three are checked.

  ### 7. “Seal” currently means several different things

  The implementation uses sealing to describe a mixture of:

  - Preventing future registry mutations
  - Copying declarations
  - Protecting runtime state
  - Validating declaration combinations
  - Preventing unauthorized construction
  - Preventing post-boot field replacement

  Those are related but distinct properties. Implementing one does not imply the others.

  A registry that panics on Register after boot may still expose mutable previously registered values. A cloned registry may
  still contain mutable nested references. A protected internal view may still be bypassed by public consumers.

  ## Development recommendations

  ### 1. Make boot a formal ownership-transfer boundary

  Adopt an explicit rule:

  > Everything passed into module registration is borrowed only until registration returns. Booted runtime state owns an
  > isolated, validated, immutable representation.

  Every registry should follow the same lifecycle:

  Declaration input
      ↓ normalize
      ↓ validate
      ↓ recursively copy/materialize
      ↓ compile
  Immutable runtime representation

  Runtime code should not retain declaration-layer objects.

  For complex declarations, define separate types:

  type WorkflowDeclaration struct {
      // Flexible authoring representation.
  }

  type CompiledWorkflow struct {
      // Restricted, normalized runtime representation.
  }

  Runtime components should accept CompiledWorkflow, never WorkflowDeclaration.

  This avoids attempting to make an open-ended authoring object simultaneously safe as a runtime object.

  ### 2. Eliminate any from invariant-bearing declaration fields

  Condition.Equals any is too permissive for a compiled workflow condition. Replace it with a closed value model, for example:

  type ScalarKind uint8

  const (
      ScalarString ScalarKind = iota
      ScalarBool
      ScalarInteger
      ScalarDecimal
  )

  type Scalar struct {
      kind ScalarKind
      s    string
      b    bool
      i    int64
  }

  If JSON-like objects are genuinely required, define and recursively clone a supported value algebra. Reject pointers,
  functions, channels, mutable custom implementations, unsupported numeric forms, and cyclic values at registration.

  The principle is:

  > If the framework cannot define how to compare, serialize, clone, and validate a value, it should not accept that value.

  ### 3. Establish one authoritative runtime state

  Public Booted fields should not be usable as runtime authority.

  Prefer something like:

  type Booted struct {
      runtime *runtimeState
      info    BootInfo
  }

  Expose behavior-oriented accessors:

  func (b *Booted) StartWorker(ctx context.Context) error
  func (b *Booted) ApplySeeds(ctx context.Context, db DB) error
  func (b *Booted) ReadinessHandler() http.Handler
  func (b *Booted) LocaleMiddleware() func(http.Handler) http.Handler

  This is safer than exposing registry objects or mutable snapshots.

  If informational views are needed, return fresh copies:

  func (b *Booted) SeedInfo() SeedInfo

  Framework and generated code should be unable to access the underlying mutable representation.

  ### 4. Remove the unbooted fallback

  If StartWorker requires a successfully booted application, it should reject anything else:

  if b == nil || b.runtime == nil {
      return ErrNotBooted
  }

  Do not fall back to public fields when the internal runtime view is absent. A fallback converts construction misuse into
  apparently valid but unvalidated operation.

  Where possible, make successful construction impossible outside App.Boot by using unexported fields and constructors. Even
  when Go permits zero values, operations should fail loudly on zero-constructed values.

  ### 5. Materialize interface-backed content at boot

  Copying an interface does not copy its underlying state. For migration and seed filesystems, materialize the actual content
  during boot:

  fs.FS declaration
      ↓ enumerate files
      ↓ read bytes
      ↓ validate paths and contents
      ↓ compute digest
      ↓ store immutable byte snapshots

  The runtime should consume the captured bytes, not call a module-owned fs.FS later.

  This provides:

  - Deterministic migrations
  - Stable readiness hashes
  - No post-boot filesystem aliasing
  - Better diagnostics when declared files are invalid
  - A natural integrity check between boot and execution

  ### 6. Separate domain events from transactional execution contexts

  A domain event should describe what happened. A transaction handle is an execution capability. Combining them makes API
  evolution and retry semantics harder.

  A safer design is a versioned hook interface:

  type UploadEvent struct {
      // Preserve stable-v1 fields.
  }

  type UploadDelivery struct {
      Event      UploadEvent
      DeliveryID string
  }

  type UploadEffects interface {
      EnqueueOutbox(...)
      RecordAudit(...)
  }

  type TransactionalUploadHook interface {
      Uploaded(
          context.Context,
          UploadDelivery,
          UploadEffects,
      ) error
  }

  The framework can provide an effects implementation bound to the confirming transaction without exposing the raw database
  transaction broadly.

  If compatibility cannot be preserved, introduce an explicit v2 API and document the migration. Do not call exported struct
  field additions universally “additive”; they break unkeyed literals.

  ### 7. Standardize declaration validation

  Every registry should apply the same minimum validation policy at registration or boot:

  - Non-empty, normalized, unique names
  - Nil and typed-nil rejection
  - Positive durations and limits
  - Non-nil callbacks
  - Non-nil filesystem/provider implementations
  - Valid ownership and dependency relationships
  - Supported value types only
  - No ambiguous zero-value behavior
  - Errors collected with declaration owner and source context

  Create reusable helpers rather than reimplementing this inconsistently:

  func RejectTypedNil[T any](name string, value T) error
  func RequireName(kind, name string) error
  func RequirePositive(kind, field string, value time.Duration) error

  Typed nil deserves an explicit test because an interface holding a nil pointer is not itself nil.

  ## Testing recommendations

  ### 1. Build a generic ownership-isolation test contract

  Every declaration registry should pass the same mutation matrix:

  1. Register a declaration.
  2. Mutate the original input before boot.
  3. Boot.
  4. Mutate the original input after boot.
  5. Mutate values returned by every getter.
  6. Replace every public informational field.
  7. Execute the corresponding runtime behavior.
  8. Assert runtime results are unchanged.
  9. Repeat under -race.

  For nested structures, mutate:

  - Every slice element
  - Slice backing arrays
  - Maps and map values
  - Pointers
  - Interface-held maps/slices/pointers
  - Nested callback containers
  - Filesystem contents
  - Returned snapshot collections

  This should be a reusable conformance suite, not bespoke tests for each registry.

  ### 2. Add property-based tests for clone independence

  For types with clone or compile operations, test the actual invariant:

  compile(x) == compile(clone(x))

  mutate(x)
  compiled result remains unchanged

  mutate(clone output)
  original and compiled result remain unchanged

  Generate nested values up to a bounded depth. Include maps, slices, empty values, duplicate keys, large numbers, typed nils,
  and unsupported values.

  For workflow conditions, property tests should also assert deterministic comparison and serialization.

  ### 3. Maintain external compatibility fixtures

  Create small derived-project fixtures pinned to the public API of each supported release:

  compat/
    v1.0-consumer/
    v1.1-consumer/
    current-generated/

  They should compile against the current framework during CI and deliberately use sensitive constructs:

  - Positional exported-struct literals
  - Implemented public interfaces
  - Embedded public types
  - Generated templates
  - Extension registration APIs
  - Error matching and serialization
  - Zero-recurring and empty-catalog configurations

  Tools such as API-diff checks help, but compilation of real prior consumers is the decisive gate.

  ### 4. Treat generated products as first-class tests

  For each generated product, test both generation and operation:

  - Generate from a clean directory.
  - Compile it against the current framework.
  - Boot with zero and nonzero declarations.
  - Run migration, API, and worker binaries.
  - Mutate public boot mirrors before construction of downstream handlers.
  - Verify the runtime still uses the validated internal state.
  - Exercise occupied ports, shutdown, missing data, retries, and invalid declarations.

  Templates should use only supported public accessors. A lint can prohibit direct template references such as booted.Seeds or
  booted.I18n.

  ### 5. Require discriminating regressions

  For every material finding, demonstrate that the regression fails when the relevant predicate, lock, clone, or runtime
  accessor is removed.

  This can be done through:

  - Temporary revert verification during review
  - Mutation testing
  - A deliberately defective test fixture
  - A narrow test seam where concurrency order must be forced

  But discriminating power should apply to the invariant, not merely one instance. For example, proving one workflow slice is
  cloned does not prove every nested mutable value is isolated.

  ## Architecture and CI controls

  ### Add an invariant ledger

  Maintain a machine-reviewable table for framework-wide invariants:

   Invariant                        Enforcement                            Consumers               Regression
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Runtime state isolated from      compile/deep snapshot                  API, worker, migrate    mutation conformance
   declarations
  ───────────────────────────────  ─────────────────────────────────────  ──────────────────────  ────────────────────────────
   Only boot-produced values can    private runtime marker                 all boot operations     zero-constructed rejection
   run
  ───────────────────────────────  ─────────────────────────────────────  ──────────────────────  ────────────────────────────
   Stable-v1 consumer compiles      compatibility fixture                  derived projects        prior-version build
  ───────────────────────────────  ─────────────────────────────────────  ──────────────────────  ────────────────────────────
   Hook effects are atomic          transaction-bound effects interface    document hooks          rollback/retry test
  ───────────────────────────────  ─────────────────────────────────────  ──────────────────────  ────────────────────────────
   Generated consumers use          accessor-only template lint            generated products      golden runtime
   runtime state

  When a change touches an enforcement point or adds a consumer, CI should require the ledger entry to be updated.

  ### Add consumer-path linting

  Static checks should prevent known boundary violations:

  - Generated templates cannot access informational Booted fields.
  - Runtime packages cannot import declaration-layer representations where compiled types exist.
  - Runtime code cannot call registry getters after boot if a runtime snapshot exists.
  - Exported stable structs cannot gain fields without compatibility approval.
  - New registry methods must be included in seal and mutation conformance tables.
  - New any, map, slice, pointer, callback, or fs.FS fields in declarations require clone and validation implementations.

  ### Review changes by defect class

  When one alias is found, the review scope should automatically expand to:

  - Every field in the same object graph
  - Every getter returning that graph
  - Every runtime consumer
  - Every generated consumer
  - Every equivalent registry
  - Every post-boot mutation path

  When a public compatibility break is found, inspect every public type changed in the remediation, not just the originally
  reported type.

  ## Practical definition of “closed”

  A finding in this framework should be marked closed only when all five statements are true:

  1. The demonstrated instance no longer fails.
  2. The underlying invalid state is unrepresentable or rejected.
  3. Equivalent paths across the complete object graph are covered.
  4. Framework, generated, and external consumers preserve the invariant.
  5. A discriminating regression and an architectural guard prevent recurrence.

  The current remediation work has been effective at fixing concrete database and concurrency failures. Its repeated weakness
  is that closure is declared after fixing the visible path, while the framework still permits the broader invalid state.
  Moving from path-based patching to ownership-based runtime design, closed types, single-source boot state, and generic
  conformance testing would prevent most of these findings before adversarial review.
---

## Remediation addendum (2026-07-17, same branch, commit 4)

Adopted lens (from this audit's Advise section): boot is an OWNERSHIP-TRANSFER
and immutable-snapshot boundary, not just a validation event. Each item below
names its enforcement and regression; claims are scoped to exactly what is
listed. The new docs/reference/invariant-ledger.md is the standing record.

### Finding 1 (High) — workflow condition alias: FIXED by unrepresentability

Condition.Equals stays `any` for YAML authoring, but validation now REJECTS
every non-scalar value (map, slice, pointer, func, struct, nil) with a
collected boot error — the framework only accepts condition values it can
compare, clone, and serialize deterministically, so the aliasable state is
unrepresentable rather than cloned. Regressions:
TestGatewayConditionRejectsMutableEqualsValues (seven value shapes) and
TestGatewayRoutingImmuneToRetainedDefinitionMutation (mutates everything
reachable in the retained declaration and proves gateway target selection
unchanged; runs under the -race gate).

### Finding 2 (High) — seed state: FIXED

seeds.Bundle.Clone() deep-copies every outer and nested slice
(PermissionSeed.StepUpAMR, RoleSeed.Permissions). Boot captures a clone in the
runtime view; Booted.Seeds is a second, independent clone (informational
mirror). RuntimeSeeds() returns a fresh deep copy; the readiness seed check
(app/health.go) and the generated migrate template consume the validated
bundle, never the field. Regression:
TestSeedsI18nAndMigrationContentAreBootCaptured (field replacement + nested
mutation of getter results and mirrors).

### Finding 3 (Medium) — UploadEvent compatibility: FIXED

UploadEvent is restored to its frozen seven-field v1 shape (positional-literal
compile fixtures in foundation/document and internal/compat). The
transactional contract is delivered through a compatibility-preserving
delivery context: document.UploadDeliveryFromContext(ctx) returns
{DeliveryID, Tx} inside a hook invocation. The atomicity regression was
migrated and still proves both properties. CHANGELOG corrected — the earlier
"additive" characterization was wrong for positional literals.

### Finding 4 (Medium) — I18n replacement: FIXED

The boot-frozen catalog is captured in the runtime view; RuntimeI18n() is the
accessor; the generated api template passes booted.RuntimeI18n() to
httpx.Locale. Regression asserts field replacement cannot change what the
accessor returns.

### Residual gaps — FIXED

- Migration filesystems are MATERIALIZED at boot: every declared file's bytes
  are read into an immutable snapshot (unexported snapshotFS); the runtime
  never calls a module-owned fs.FS again, and unreadable declarations fail
  boot. Regression mutates the module's MapFS post-boot and proves the
  snapshot serves the validated bytes (and no post-boot file appears).
- The unbooted fallback is REMOVED: Runtime accessors panic and StartWorker
  returns ErrNotBooted for any Booted value App.Boot did not produce
  (TestUnbootedBootedFailsLoudly). The one in-repo test that hand-constructed
  Booted was rewritten to boot for real.
- Nil fs.FS (Migrations/Seeds), empty-name/nil health checks, and nil or
  typed-nil integration providers are collected boot errors
  (TestBootRejectsNilAndEmptyDeclarations,
  TestRegisterRejectsNilAndTypedNilProviders).
- internal/sealer's doc now states the narrower guarantee explicitly: an
  out-of-module boundary; in-repository construction remains possible and is
  restricted to app.Boot by review.

### Systemic guards adopted (from the Advise recommendations)

- Consumer-path lint: scripts/lint_boundaries.sh fails when a generated
  template reads an informational Booted field instead of a Runtime* accessor.
- Stable-v1 compatibility fixture: internal/compat/stable_v1_consumer_test.go
  compiles the sensitive positional-literal constructs exactly as an external
  consumer writes them (app.Hook, document.UploadEvent).
- Invariant ledger: docs/reference/invariant-ledger.md maps each framework
  invariant to its enforcement, consumer classes, and guarding regression,
  with the rule that changes touching an enforcement point update the row.

### Residual, recorded not claimed

The Advise section's fuller programme — behavior-oriented Booted accessors
replacing the informational fields entirely, a generic property-based
ownership-conformance suite across all registries, and pinned prior-release
consumer fixtures (compat/v1.0-consumer builds) — is architecture work beyond
this remediation; the ledger records today's enforcement honestly. The
informational mirror fields remain exported for v1 compatibility but are
non-authoritative everywhere in framework and generated code.
