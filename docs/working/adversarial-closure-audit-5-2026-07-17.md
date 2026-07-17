# Fifth adversarial closure audit (2026-07-17) — register and remediation

> SUPERSEDED FRAMING (2026-07-17): this record describes the interim `/v2`
> module cutover. The product decision is now a CLEAN V1 RESET on the root
> module `github.com/qatoolist/wowapi` (next unused version v1.2.0; v1.0.0/
> v1.1.0 abandoned). The correctness/opacity fixes below stand unchanged; only
> the module path reverted from `/v2` to root. See
> docs/working/comprehensive-review-2026-07-17.md.


Verdict received: NO-PUSH at 51701f8, revised by the product owner's directive
that V1 is discarded (no compatibility obligations) and the reviewer's
follow-up: evaluate strictly as V2 — high bar for correctness, immutable
ownership, deterministic behavior, generated-project independence, and
production operation.

## Findings

1. High — RuntimeKernel() returned the authoritative mutable kernel pointer;
   the boot-time struct copy was shallow (nested Cfg maps/slices shared).
2. High — workflow validation became stale: registration after a clean Err()
   left `validated` true; concurrent registration/execution could race.
3. High — persisted workflow definitions bypassed semantic validation
   (ParseDefinition only) in instance execution and the SLA sweep.
4. High — auto actions and assignee resolvers received the live framework-
   owned context map.
5. Medium — float64 numeric canonicalization lost precision (2^53, int64/
   uint64 boundaries, float32 decimals, NaN/Inf representable).
6. Medium — CompleteTask serialized the caller's output twice (task row vs
   context merge could diverge under a stateful marshaler).
7. Disposition — V1/V2: resolved by decree; ship as V2, remove compatibility
   surfaces rather than protect them.
8. Low — template lint not alias-proof; no negative fixture; recommended
   structural (unexported-field) guarantees over grep.

---

## Remediation addendum (2026-07-17, same branch, commit 6 — the V2 cutover)

### Finding 1 — FIXED by removal

RuntimeKernel() no longer exists. The captured kernel view (struct copy with
config.Framework.Clone() — a reflective deep copy; JSON round-tripping is
impossible because secret fields redact on marshal) is reachable only through
the unexported runtime view; external consumers get narrow interface
accessors (RuntimeAuthz, RuntimeTx). Regressions: gutting the caller-owned
aggregate post-boot leaves the accessors and a running StartWorker unaffected
(TestKernelReplacementAndMutationCannotAlterRuntimeDependencies); mutating
the original kernel's nested config slices post-boot leaves the captured view
unchanged (TestCapturedKernelConfigIsDeeplyIsolated, via a test-only seam —
the aggregate stays unexported).

### Finding 2 — FIXED

The workflow registry is RWMutex-guarded and its validation state is
GENERATION-KEYED: every mutation attempt (all three registration methods)
bumps the generation and clears `validated` at entry; Err() records the
generation it validated; the runtime executes only when
validated && validatedGen == gen. Regressions:
TestValidationInvalidatedByEveryMutation (each method, rejection until a new
clean pass) and TestConcurrentRegistrationAndExecutionIsSafe under the -race
gate.

### Finding 3 — FIXED

parseAndValidateDefinition is the single loader for persisted definitions
(instance execution AND SweepSLA): parse, then Definition.Validate against
the registry's registered auto-action/resolver sets — the same semantic rules
as registered definitions, including the fail-closed gates (vote steps,
min_approvals > 1, self_approval:false, ratify_by), graph integrity, and
condition scalars — rejecting BEFORE any callback or state transition.
Regressions: TestIntegrationPersistedVoteDefinitionRejected (execution path
asserts the task stays open; SLA-sweep path errors) and the updated
corrupt-defense suite (dangling transitions, unknown assignee kinds now fail
at load).

### Finding 4 — FIXED

Auto actions and assignee resolvers receive a deep canonical COPY of the
instance context (canonicalizeContext); returned output is the only mutation
channel. Regression: TestIntegrationCallbacksReceiveIsolatedContext — a
mutating+retaining auto action cannot steer the downstream gateway or the
persisted context; a mutating resolver cannot alter the persisted context;
post-hoc mutation of the retained map is inert (-race gate).

### Finding 5 — FIXED

Exact numeric model: every context decode point uses json.Number
(decodeCanonicalContext — StartIn, canonicalizeContext, decodeJSONMap);
comparison reduces by KIND and compares numbers as exact rationals
(math/big.Rat); condition floats compare by their shortest round-trip decimal
at their own precision (float32(0.1) equals JSON 0.1); NaN/Inf are rejected
at declaration validation and unrepresentable in JSON context. Regression:
TestNumericComparisonExactBoundaries (2^53 vs 2^53+1, int64/uint64 maxima,
float32(0.1), NaN/Inf, scientific notation, and canonical round-trip
precision).

### Finding 6 — FIXED

CompleteTask canonicalizes the output EXACTLY ONCE and uses that value for
task persistence and the context merge. Regression:
TestIntegrationCompleteTaskSerializesOutputOnce (a stateful marshaler is
invoked exactly once; task row and merged context carry the same bytes).

### Finding 7 — EXECUTED (the V2 cutover)

- Module path github.com/qatoolist/wowapi across the repo, templates,
  release tooling, and the golden/e2e harness (synthetic versions v2.0.0-*;
  the local-replace pseudo-version is v2-family).
- app.Booted is OPAQUE: all informational mirror fields deleted; capabilities
  flow through Runtime* accessors (incl. new RuntimeOpenAPI); the former
  field-replacement regressions are structurally obsolete and were removed.
- The v1.1.0→candidate upgrade-replay gate was removed (cross-major module
  identity makes it meaningless); the golden consumer builds and boots a pure
  V2 product.
- Docs reframed: CHANGELOG (V2 header + V2 cutover section), README status,
  upgrade/deprecation policy (v2 rules), D-0091 rewritten as the V2 opacity
  decision, internal/compat/v2_contract_test.go is the V2 contract fixture.

### Finding 8 — FIXED structurally + guard retained

Opacity makes forbidden template reads impossible for Booted state; the
alias-aware lint and its negative fixtures remain as a standing guard for the
accessor discipline.

### Residual, recorded not claimed

The reviewer's fuller behavior-oriented surface (e.g. Booted.APIHandler()
composing the middleware chain internally) remains open architecture work;
today's accessors are narrow and validated but still hand out live sealed
collaborators (Router for mounting, Events/Jobs for the worker). The ledger
records enforcement honestly.

### Independent-gate findings on this round (fixed before commit)

The fresh adversarial reviewer verifying this addendum found and we fixed:

- `config.Framework.Clone()` silently ZEROED `config.Secret` fields (DB.DSN,
  MigrateDSN, PlatformDSN): reflection cannot Set unexported fields, and the
  struct case built a fresh zero value. Fixed with a whole-value copy before
  exported-field deep-copies (Secret's unexported strings are immutable, so
  the value copy is exactly right); the false "no unexported fields exist"
  comment corrected; discriminating regression added
  (TestClonePreservesSecretsAndIsolatesMutables — fails on the zeroing
  variant).
- Two `go mod edit -replace` paths missed the /v2 cutover
  (scripts/smoke_reference_stack.sh — CI-wired — and scripts/devbox/wow-link):
  the reference-stack smoke would no longer build against the local checkout.
- README quick-start still installed `cmd/wowapi@v1.1.0` on the old path;
  the upgrade policy doc opened with "stable v1 line" and used v1 minors as
  v2 examples; D-0091 cited the pre-rename fixture filename; two user-guide
  fences and one blueprint fence referenced deleted Booted fields; the
  template-lint fixture used the removed RuntimeKernel symbol.
- The json.Number contract for AutoInput.Context / ResolveInput.Context was
  undocumented for module authors — now documented on both types and in the
  module SDK blueprint (numbers are json.Number; assert accordingly).
