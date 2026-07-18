# Fourth adversarial closure audit (2026-07-17) — register and remediation

Verdict received: NO-PUSH at 9a28c53. The audit confirmed the third-round
closures (UploadEvent shape, hook atomicity, seed cloning, readiness seed
capture, I18n capture, RuntimeMigrations snapshots, fail-loud accessors,
health/provider validation, content-derived harness versions, F-04/F-05) but
found the recurring pattern in two more places: validation and ownership
enforced on the App.Boot path while equivalent public or legacy paths stayed
open.

## Findings

1. High — workflow validation optional through the public API:
   RegisterDefinition stored definitions before validation (deferred to
   Err()); NewRuntime accepted an unvalidated registry.
2. High — the Kernel aggregate outside the ownership boundary: Booted.Kernel
   replaceable and authoritative for StartWorker and the generated API; Boot
   retained the caller-owned *kernel.Kernel whose fields the composition root
   could mutate post-boot.
3. Medium — Booted's unexported runtime field breaks stable-v1 positional
   literals and changes hand-constructed behavior; needs an explicit API
   decision.
4. Medium — gateway comparison used fmt.Sprint equality: "1" matched 1,
   "true" matched true, Stringer implementations were invoked, and pre- vs
   post-reload routing could diverge (initial step executed over the
   caller-owned map).
5. Medium — the public Migrations mirror was populated from the original
   module filesystems, not the materialized snapshots.
6. Medium — typed-nil fs.FS values passed the Migrations/Seeds nil checks.
7. Low — no real goose/PostgreSQL regression through snapshotFS; template
   lint bypassable by aliasing `booted`; no negative lint fixture; trailing
   whitespace made the claimed clean diff-check false.

---

## Remediation addendum (2026-07-17, same branch, commit 5)

Scope statement: this addendum claims exactly what is listed below.

### Finding 1 — FIXED

RegisterDefinition validates gateway condition values SYNCHRONOUSLY before
storage: a non-scalar Equals returns an error and is never stored (the alias
cannot exist). Registry.Err() records completed validation; every executing
Runtime method (StartIn, Decide, CompleteTask, Delegate, Override, SweepSLA)
refuses an unvalidated or invalid registry with a named error. Regressions
(public API only, no App.Boot): TestRegisterDefinitionRejectsMutableEqualsSynchronously
(six shapes, asserts nothing stored) and TestRuntimeRefusesUnvalidatedRegistry
(never-validated, validation-failed, and validated-passes cases). Two
pre-existing fallback tests that executed without validation were updated to
complete validation first, as the contract now requires.

### Finding 2 — FIXED

Boot captures a STRUCT COPY of the kernel aggregate in the runtime view;
RuntimeKernel() is the accessor. StartWorker, the readiness detail providers
(rule/model hashes), and the generated api template consume it; the template
lint forbids booted.Kernel. Booted.Kernel remains an informational mirror.
Regression (TestKernelReplacementAndMutationCannotAlterRuntimeDependencies):
sets booted.Kernel = nil AND guts the caller-owned kernel's fields post-boot,
asserts RuntimeKernel() serves the captured dependencies, and runs StartWorker
with the nil field — the pre-fix field-reading code nil-panicked.

### Finding 3 — DECIDED and documented (D-0091)

The break is accepted deliberately: a hand-constructed Booted never passed
boot validation, and silently operating on one was itself the F-10 defect.
Recorded in docs/implementation/decisions.md (D-0091), on the Booted doc
comment, and in the CHANGELOG as a breaking change with migration guidance
(obtain Booted only from App.Boot). The stable-v1 positional freeze
intentionally covers app.Hook and document.UploadEvent, not app.Booted.

### Finding 4 — FIXED

Gateway comparison is type-preserving and canonical: values reduce by KIND to
string/bool/float64 (named scalar types included) and compare under Go
interface equality — a string can never match a number or bool, and no method
on any value is ever invoked. The instance context is CANONICAL JSON
everywhere: StartIn executes over the decoded persisted JSON (never the
caller-owned map), and task/auto output merges are JSON round-tripped before
merging, so routing is identical before and after reload. Regressions:
TestGatewayComparisonIsCanonicalAndTypeSafe (string-vs-number, string-vs-bool,
named types, panicking Stringer never invoked, cross-numeric match, and
before/after-reload equivalence for every scalar kind).

### Findings 5 & 6 — FIXED

The public Migrations mirror is populated from the SAME materialized snapshots
as the runtime view (independent outer map). Migrations/Seeds registration
rejects typed-nil fs.FS values via a shared nil-like validator; regression
cases added for both. The goose regression below covers the mirror path
implicitly (same snapshot values).

### Finding 7 — FIXED

- TestIntegrationMaterializedSnapshotMigratesRerunsAndResets drives the REAL
  goose engine over PostgreSQL through the materialized snapshot — apply,
  idempotent rerun, and reset — while the module-owned MapFS is corrupted and
  extended post-boot (the addition must be invisible; the corruption inert).
- The template lint moved to scripts/lint_templates.sh with alias detection
  (`x := booted` is a violation) and is exercised by NEGATIVE fixtures
  (internal/compat/template_lint_test.go): a field read fails, an alias
  fails, an accessor-only template passes.
- The trailing whitespace in the audit-3 record is fixed; git diff --check is
  clean across the branch range.

### Residual, recorded not claimed

Behavior-oriented accessors replacing Booted's informational fields entirely,
and pinned prior-release consumer fixture builds, remain future architecture
work (see the invariant ledger). The informational mirrors are
non-authoritative everywhere in framework and generated code, enforced by
consumer-path lint.
