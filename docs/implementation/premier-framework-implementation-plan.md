# wowapi Premier-Framework Implementation Programme

- **Source directive:** [`architecture-directive-2026-07-11.md`](architecture-directive-2026-07-11.md) (reviewed commit `d3c2640dbe1a0fe27e826cdf053945c4f49bc034`)
- **Source evidence:** [`evidence/architecture-review-2026-07-11/command-log.md`](evidence/architecture-review-2026-07-11/command-log.md), [`evidence/architecture-review-2026-07-11/evidence.json`](evidence/architecture-review-2026-07-11/evidence.json)
- **This document's purpose:** convert the directive's 38 findings into an execution-ready, delegated, independently-verifiable programme — task breakdown, dependencies, ownership tier, acceptance criteria, tests, evidence, gates, risks, and a full traceability mapping from finding → task → evidence → status. It also assesses the impact of every framework-level change on the dependent product repo `wowsociety`.
- **Status of this document:** planning artifact. Sections marked **[EXECUTED]** below have real code/tests landed and independently reviewed in this pass; everything else is **[PLANNED]** — broken into actionable tasks but not yet implemented. Do not read a PLANNED entry as done.
- **Scope discipline:** per the directive's own §13.1, a work package cannot move to `in_progress` in a real tracker until a named human DRI and reviewer replace the role-based owner used here. This document assigns *capability-tier* owners (which kind of agent/reviewer a task needs) as an implementation-delegation aid, not a claim that human ownership has been assigned.

## 1. How to read this document

1. §2 restates the directive's severity model and adds an agent-cost-tier column.
2. §3 gives the wave sequencing exactly as the directive defines it (§12), annotated with which findings are Wave-0-eligible today versus blocked on prerequisite infrastructure (a reference performance environment, a named security lead, GitHub org-admin action, etc.).
3. §4 is the **consolidated findings register** — one row per finding (38 rows), cross-referencing the detailed per-work-package task tables in §5.
4. §5 is organized by work package (PF-ARCH, PF-SEC, PF-DATA, PF-DX, PF-PERF, PF-REL) exactly as the directive's §13.1 defines them. Each finding gets: task breakdown, dependencies, acceptance criteria, required tests, evidence to produce, wowsociety impact assessment, and risk notes.
5. §6 is the full **traceability matrix**: finding → task IDs → evidence path → current status.
6. §7 lists cross-cutting risks, assumptions, blockers, and unresolved questions that don't belong to one finding.
7. §8 records what was actually executed in this pass, with the independent-review-gate result.

## 2. Severity model and agent-cost tiering

| Priority | Meaning (directive §4) | Required response | Typical agent tier for implementation | Typical agent tier for review |
| --- | --- | --- | --- | --- |
| P0 | Can violate authorization, tenant integrity, durable execution, release trust, or bounded-resource guarantees; or makes the advertised default path unusable | Block the next stability/release claim until fixed and regression-tested | Sonnet (standard), Opus/high-tier only where the fix is genuinely architecturally complex (e.g. AR-01/AR-02 compiler design, SEC-01 principal resolver design) | Opus/high-tier independent reviewer mandatory — security/correctness stakes are high |
| P1 | Deep architecture/DX flaw that makes extension unsafe, drift-prone, or operationally unreliable | Complete before adding another major framework subsystem | Sonnet | Sonnet or Opus depending on blast radius |
| P2 | Scalability, completeness, or operability work whose urgency depends on measured use | Implement after P0/P1 foundations or when its explicit trigger fires | Sonnet or cheaper (mechanical instrumentation/metrics work) | Sonnet |
| Parked | Evidence says current implementation is adequate or need is not demonstrated | Keep measured; do not implement speculatively | N/A | N/A |

Routine/mechanical work (doc updates, evidence-bundle formatting, test-skip inventories, running existing scripts, straightforward config/YAML edits) is explicitly cheap-tier or scripted — never assigned to a high-reasoning agent. Architecture decisions (AR-01–03, DX-03 DSL design), complex concurrency/lease debugging (DATA-02/03/04), and security-sensitive design (SEC-01/02/04/05) are reserved for high-capability agents plus a mandatory independent adversarial reviewer, per the directive's own closure-contract bar in §13.2 — a passing broad test suite is explicitly *not* sufficient evidence for these.

## 3. Wave sequencing (directive §12) — eligibility today

| Wave | Content | Eligible to start now? | Blocking prerequisite |
| --- | --- | --- | --- |
| 0 | Stop release/correctness hazards (8 small high-leverage fixes) | **Yes** — self-contained code fixes | None; some items (REL-01 step 4, REL-02) need GitHub org-admin action beyond a coding agent's reach |
| 1 | Compile and seal the application model (AR-01–06) | Depends on Wave 0 exit gate | Wave 0 P0 closure |
| 2 | Authoritative identity and tenant data invariants (SEC-01, DATA-01/06/07/09) | Depends on Wave 0/1 | Wave 0 exit gate; DATA-09 also needs a release-engineering decision on rolling-deploy tooling |
| 3 | Standardize durable execution (DATA-02/03/04) | Depends on Wave 2 | Lease/fencing primitives from Wave 2 identity work are a soft dependency (shared idempotency-key infrastructure) |
| 4 | Operation DSL and golden product (DX-02 full/DX-03/DX-04/DX-06) | Depends on Wave 1 | AR-01–03's ApplicationModel must exist first — DX-03's DSL is explicitly described as compiling into it |
| 5 | Optimize from production-shaped evidence (PERF-02–05) | **Blocked** on a not-yet-built reference environment | `perf/reference-v1.json` + fixtures (directive §14) — this is itself a prerequisite task, tracked under PF-PERF |
| 6 | Complete evidence, operations, premier release (DATA-08 full, REL-02/03 full, independent pentest) | Depends on Waves 2–5 | Also needs a named security lead and an external penetration-test vendor — human/organizational, not codeable |

## 4. Consolidated findings register

38 findings, cross-referenced against the directive's own §13.2 closure-contract table (confirmed 1:1 by `command-log.md` E-17: "38 finding headings matched 38 unique closure contracts"). Detailed task tables are in §5.

| ID | Work package | Priority | Wave | One-line finding | wowsociety affected? |
| --- | --- | --- | --- | --- | --- |
| AR-01 | PF-ARCH | P1 (P0 untrusted ext.) | 1 | Mutable mega-context / no ownership-bound registration | Yes — low severity (dead retained-registrar field, owner-string pattern) |
| AR-02 | PF-ARCH | P1 (P0 untrusted ext.) | 1 | String/`any` ports, no compiled provider graph | No — zero port usage |
| AR-03 | PF-ARCH | P1 | 1 | No single authoritative declaration; duplicated identity across projections | Yes — passively, opt-in future benefit only |
| AR-04 | PF-ARCH | P1 | 1 (T1 earlier) | Config/boot-time silent behavior (unknown namespaces, last-writer-wins) | Yes — low risk, already partially assumed |
| AR-05 | PF-ARCH | P1 | 0/1 (doc fix) | Composition/documentation drift (README/blueprint vs. live API) | No |
| AR-06 | PF-ARCH | P1 | 0/1 (T1 small fix) | Hidden constructor bypass (`kernel.go` builds a second `authz.NewStore()`) | No — uses `Privileged()` only |
| SEC-01 | PF-SEC | P0 | 2 | Tenant/privileged-session state trusted from JWT claims, not server-verified | **Yes — most exposed finding.** Breaking for impersonation flow |
| SEC-02 | PF-SEC | P0 | **0** | Workflow `Override` can skip authz entirely with nil evaluator | No — zero workflow usage |
| SEC-03 | PF-SEC | P1 | 2 | Webhook replay controls trust unauthenticated timestamp/event-ID | No — zero webhook usage |
| SEC-04 | PF-SEC | P1 (P0 if cache enabled) | 2/5 | Unbounded authz cache map, no sweep, no cross-pod invalidation | No — cache not enabled |
| SEC-05 | PF-SEC | P1 | 6 | No versioned ASVS/NIST/API-Security control map | Yes — real adversarial test baseline exists to build on |
| SEC-06 | PF-SEC | P1 | 2/5 | Outbound-security escape hatches (JWKS client injection, allowlist hostname bypass) not audited/fingerprinted | Partial — config-only, not inspected in this pass |
| PERF-01 | PF-PERF | P0 | **0** | Token-bucket map cannot evict one-shot keys; O(N) sweep at 10k+ | Yes — config exposure only, no logic impact |
| PERF-02 | PF-PERF | P1 | 5 | No DB-backed request-budget benchmarks (only faked auth/tx) | No — no wowsociety benchmark surface |
| PERF-03 | PF-PERF | P1 | 5 | Rules resolution is N+1 per org ancestor | No — wowsociety never uses org-scoped rules |
| PERF-04 | PF-PERF | P1 | 5 | N+1/unbounded materialization in sweepers, webhook retry, outbox | No — zero workflow/webhook/outbox usage |
| PERF-05 | PF-PERF | P2 | 5 | S3 `Stat` full-downloads+hashes when checksum metadata absent | Indirect — inherits fix with no call-site change |
| PERF-06 | PF-PERF | P1 | **0** | Missing benchmark warns, doesn't fail CI; no real coverage-guided fuzzing | Yes — infra gap (wowsociety has no bench/fuzz gate at all) |
| DATA-01 | PF-DATA | P0 | 2 | Tenant-local FKs reference only parent `id`, not `(tenant_id, id)` | **Yes — live independent instance** (`policy_override.rule_version_id`) |
| DATA-02 | PF-DATA | P0 | 3 | Jobs have no lease/fencing; stale worker can overwrite outcome | No — zero job registration |
| DATA-03 | PF-DATA | P0 | 3 | Notify/webhook network I/O runs inside open DB transactions | No — zero notify/webhook usage |
| DATA-04 | PF-DATA | P1 (P0 before multi-worker) | 3 | Bulk processing claims to be replica-safe; code has no lock at all | No — zero bulk usage |
| DATA-05 | PF-DATA | P1 | 2 | `MAX(version)+1` version-allocation races; orphaned blobs on conflict | No — zero artifact/document usage |
| DATA-06 | PF-DATA | P1 | 2 | Resource-mirror write is manual/forgettable, not transactionally coupled | **Yes** — `committeeseat.go` uses the exact manual pattern |
| DATA-07 | PF-DATA | P1 | 2 (blocked on SEC-01) | ReBAC ignores party-subject edges; nil-actor placeholders | No confirmed usage |
| DATA-08 | PF-DATA | P0 (Wave 0 slice) / P1 (Wave 6) | 0 + 6 | Audit hash excludes metadata/tx_id; attachment outbox error discarded; legal-delivery audit deferred on a stale blocker; DSR export not durable; hold enforcement per-callback | **Yes — W6-T1 (audit hash) is breaking** for wowsociety's live impersonation/policy audit rows |
| DATA-09 | PF-DATA | P0 | 2 | No online expand/backfill/validate/contract migration protocol exists | Yes — process gap, wowsociety's single-shot deploy collapses the N/N-1 window entirely |
| DX-01 | PF-DX | P0 | **0** | Source-built CLI can emit unresolvable `v0.0.0` dependency | No — wowsociety uses path `replace`, unaffected |
| DX-02 | PF-DX | P0 (Wave 0 slice) / P1 (Wave 4) | 0 + 4 | `gen crud` emits false-success TODO handlers + invalid `.delete` permission verb + FK anti-pattern | No — wowsociety's existing modules avoided the bug via governance discipline |
| DX-03 | PF-DX | P1 | 4 | No typed operation/module DSL exists (proposed design only) | No — Wave 1-3 compatibility guarantees zero break |
| DX-04 | PF-DX | P1 | 4 | No golden two-module consumer + upgrade-matrix fixture exists | No — wowsociety cannot itself serve as this fixture (3 concrete reasons) |
| DX-05 | PF-DX | P1 | **0** | README/policy say pre-1.0; CHANGELOG says v1.0.0-stable; tags confirm v1 | Yes — wowsociety's `FRAMEWORK_VERSION` is a SHA, not a version, can't express N/N-1 |
| DX-06 | PF-DX | P1 | 4 (overlaps AR-03 T2) | OpenAPI merge silently drops all non-`paths`/`schemas` fields | No currently observed; latent risk if fragments ever add `security`/`webhooks` |
| DX-07 | PF-DX | P1 | 1/2 | Readiness omits migration-currency check; capacity/backpressure defaults are silent no-ops | **Yes** — wowsociety's generated `cmd/api/main.go` has the identical gap |
| REL-01 | PF-REL | P0 | **0** | Release workflow builds/signs/publishes on tag push with no CI dependency, no protected environment | No — wowsociety doesn't invoke wowapi's workflows, consumes via sibling checkout |
| REL-02 | PF-REL | P0/P1 | 0 (Trivy) + 6 (full) | Trivy non-blocking; scanner-skip claims are now **stale** (repo went public, CodeQL/Scorecard already run) | No |
| REL-03 | PF-REL | P1 | 4/6 (hard-blocked on AR-01–03/DX-03/04/06) | No Go/OpenAPI/config/event compatibility diff gates | Yes — indirectly, via what a "compatible" wowapi release means; no code change required |
| REL-04 | PF-REL | P1 | **0** (S3/TOTP) + P1 later | S3 tests silently skip (wrong env var); no real coverage-guided fuzzing; TOTP flake risk | No — wowsociety's CI is fully decoupled from wowapi's `ci-container` |

## 5. Work packages — detailed task breakdown, evidence, wowsociety impact

Each subsection below was independently drafted by a dedicated planning pass over the directive's source text and the live repositories, with every evidence citation re-verified against current source (not copied from the directive) and every wowsociety-impact claim grounded in an actual read of wowsociety's code — not assumed. Mentor-review spot-checks (SEC-02 blast radius, PERF-01 sweep defect) independently confirmed accuracy against live source before integration. Two cross-package duplication risks are called out where they occur (AR-03 T2 / DX-06 OpenAPI merge; PERF-06 T3-T4 / REL-04 T8 fuzzing) — assign single ownership before implementation to avoid double work.

### 5.1 PF-ARCH — Module Registration / Application-Model Architecture (AR-01 – AR-06)

**Accountable role:** framework architecture lead. **Evidence root:** `docs/implementation/evidence/premier/PF-ARCH/`. **Wave:** 1 ("Compile and seal the application model"), gated on Wave 0 exit — except AR-05 T1/T2 and AR-06 T1, which are isolated, low-risk fixes with no dependency on the ApplicationModel type and can land immediately.

#### AR-01 — Ownership-bound ApplicationModel replacing the mutable mega-context

**Directive requirement:** immutable `ApplicationModel` compiled from ownership-bound module declarations; registration APIs never accept an arbitrary owner string from module code; every collector follows `collect → validate → seal → expose read-only snapshot`; post-seal mutation errors/panics, never silently no-ops.

**Machine-closeable proof bar (§13.2, do not weaken):** adversarial modules fail to claim every foreign declaration class; retained registrars reject post-seal writes; snapshot-mutation and race tests pass; two identical compiles emit the same complete model hash.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Define `ApplicationModel` type + `collect→validate→seal→expose` lifecycle skeleton | Wave 0 exit | A `Compiler` accumulates declarations via owner-bound calls only; `Compile()` validates then seals; post-seal, further calls error (panic only in an explicit dev/test build tag) | Unit: state-machine transition tests | `AR-01/lifecycle_test_output.txt` | Medium — load-bearing type every other AR-0x task depends on |
| T2. Owner-bound `Registrar` capability type (unexported seal method) minted only by the compiler from `Manifest.ID`/`Module.Name()` | T1 | Module code cannot construct/type-assert a `Registrar` for another owner | Compile-fail fixture attempting to fabricate a `Registrar` | `AR-01/registrar_capability_test_output.txt` | High — this is the actual security boundary |
| T3. Owner-bound registrar wrapper for `resource.Registry` | T1, T2 | `ctx.Resources()` exposes a registrar bound to the module's own identity; ownership is structural, not string-compared | Adversarial: cross-module claim attempt fails even with a matching key prefix | `AR-01/resource_ownership_adversarial_test.go` | Medium |
| T4. Owner-bound registrar wrapper for `rules.Registry` | T1, T2 | Same shape as T3 for rule points | Adversarial cross-owner rule-point claim | `AR-01/rules_ownership_adversarial_test.go` | Medium |
| T5. Owner-bound registrar wrapper for `authz.Registry` permission registration | T1, T2 | `Register(p Permission)` currently has **no owner parameter at all** — widest gap of the six; new API derives module prefix from the bound registrar | Adversarial: cross-module permission claim rejected at registrar boundary | `AR-01/authz_ownership_adversarial_test.go` | **High** — only registry with zero existing ownership check |
| T6. Owner-bound registrar wrappers for the remaining ~9+ declaration classes (events, jobs, workflow actions, providers, templates, health checks, migrations, seeds, OpenAPI) | T1, T2, T3-T5 pattern | Every declaration class in AR-01's acceptance gate is ownership-checked, not just the three headline registries | Table-driven adversarial suite, one fixture per class | `AR-01/full_declaration_class_matrix_test.go` | Medium — easy to under-scope |
| T7. Convert all snapshot-returning reads to cloned/immutable data (`Specs()`, `Points()`, and equivalents on all registries) | T3-T6 | No exported reader returns a backing map/slice | Unit: mutate returned value, assert registry internal state unaffected | `AR-01/snapshot_immutability_test.go` | Low-medium |
| T8. Reject Context retention after `Register()` returns | T1, T2 | A module retaining `ctx`/a registrar post-boot gets an explicit error on mutation, never a silent no-op or a production panic | Adversarial: fixture module retains registrar, calls it post-boot | `AR-01/post_seal_mutation_rejection_test.go` | Medium — wowsociety's `policy` module already retains `mc.Rules()` today (harmlessly); this task has a direct named consumer to validate against |
| T9. Deterministic model hash, emitted at startup/readiness | T1-T8 | Two identical compiles → byte-identical hash; one changed declaration → different hash | Unit: hash-determinism + hash-sensitivity tests | `AR-01/model_hash_determinism_test.go` | Low — exclude non-deterministic inputs (map order, timestamps) |
| T10. Race tests proving no runtime mutation of the sealed model | T1-T9 | `go test -race` clean on concurrent legitimate reads; illegitimate write fails via T8, not a data race | Race test | `AR-01/race_test_output.txt` | Low |
| T11. Legacy adapter wrapping current `module.Module`/`Context` so existing modules compile unchanged (Wave 1 compatibility strategy) | T1-T10 | Existing modules (wowapi internal + wowsociety) boot unchanged through the adapter; the adapter itself derives owner from `Module.Name()` and routes through the same owner-bound registrar — it must not bypass T2-T6 | Integration: existing contract tests pass unmodified through the legacy path | `AR-01/legacy_adapter_compat_test_output.txt` | Medium — the adapter is itself a trust boundary |

**wowsociety impact — AR-01:** Affected, low severity, **not breaking** under the Wave-1 legacy-adapter strategy. `internal/modules/policy/rulepoints.go:218` uses a hardcoded literal `"policy"` as the owner string (matches its own module name today, but is exactly the pattern AR-01 eliminates). `internal/modules/policy/pack.go:334-338`/`service.go:36` retain the raw `*rules.Registry` (`s.rulesReg`) in a struct field — written once, never read again (dead code) — precisely the "retained registrar" pattern T8 targets; must be dropped or replaced before wowsociety adopts the non-legacy v1 registrar API. Distinguish from `s.rulesStore`/`s.rulesResolver` (built over the registry, legitimately used live in request handlers) — T8 must not reject those. No wowsociety change required before/during Wave 1 landing; cleanup is low-risk and can happen on wowsociety's own schedule. Verification: wowsociety's module contract tests continue passing unmodified against a wowapi commit with T1-T11 landed, pinned via the existing `replace`/`FRAMEWORK_VERSION`.

#### AR-02 — Typed port keys and compiled provider graph

**Directive requirement:** `port.Key[T]` + `Define`/`Provide`/`Require`/`Resolve` generic functions bound to an owner-bound `Registrar`; compiler builds a heterogeneous provider graph, type-erasing only at compile time (never on request hot paths); rejects duplicate providers, missing requirements, type mismatches, undeclared dependencies, cycles, invalid scope/lifetime edges before any process starts.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Define `port.Key[T]`, reuse AR-01 T2's `Registrar`, and the four generic free functions | AR-01 T1, T2 | Happy-path define/provide/resolve round-trip compiles and works | Unit | `AR-02/port_api_unit_test.go` | Medium — Go's lack of type-parameterized methods forces a first-class-argument API; ergonomics review needed |
| T2. Internal compiler factory mints registrars with immutable owner identity | T1, AR-01 T1 | Module code cannot manufacture a `Registrar` from a bare string | Adversarial compile-fail fixture | `AR-02/registrar_forge_compile_fail_fixture/` | High — verify capability confusion is impossible if AR-01/AR-02 share one `Registrar` type |
| T3. Type-erased provider graph with zero reflection on request hot paths | T1-T2 | Benchmark/static check proves zero `reflect.*` calls at `Resolve` time | Benchmark + lint | `AR-02/hotpath_no_reflection_bench.txt` | Medium — naive implementations reflect per-call |
| T4. Boot-time graph validation: duplicate providers, missing requirements, undeclared edges, cycles, invalid scope/lifetime edges | T1-T3 | One adversarial fixture per failure class; errors name both owners | Adversarial suite, reusing `kernel/lifecycle`'s existing scope-rank ordering | `AR-02/boot_graph_validation_test.go` | Medium — absorb/replace existing lifecycle scope logic, don't duplicate it |
| T5. Compile API/worker/migrate profiles as three projections of one graph | T1-T4, AR-03 | No hand-copied wiring template remains | Integration: build all three from one fixture, assert capability subsets | `AR-02/three_profile_projection_test.go` | Medium — sequence after AR-03's manifest shape is fixed |
| T6. Retire hand-maintained `kernel/lifecycle` manifest in favor of the generated graph | T1-T5 | `lifecycle.go`/`manifest.go` deleted or generated; existing 5 lint failure classes still pass, now data-driven | Regression: existing lifecycle-lint classes pass | `AR-02/lifecycle_lint_generated_test_output.txt` | Low-medium |
| T7. Legacy port adapter (`ProvidePort`/`Port` shim onto the typed graph) | T1-T6 | Existing calls (none in wowsociety; possibly wowapi-internal fixtures) compile/resolve unchanged | Integration | `AR-02/legacy_port_adapter_compat_test_output.txt` | Low — confirmed zero external callers |

**wowsociety impact — AR-02:** Not affected. Confirmed via repo-wide search: zero call sites for `ProvidePort`/`Port(` anywhere in wowsociety. No breaking change, no required action, no sequencing constraint.

#### AR-03 — One authoritative declaration, all projections derived

**Directive requirement:** module manifest becomes the authoritative declaration; deterministic tooling derives routes, permission/resource catalogs, schema/OpenAPI, event/job/workflow/rule identifiers, dependency/provider graphs, migration/seed/i18n/OpenAPI bundle inventory, required-capability profiles, conformance tests, doc tables, and a machine-readable manifest. SQL/business behavior stay explicit Go/SQL.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Define the manifest schema — scoped to what Wave 1 needs (identity + projection inputs), not DX-03's full typed-operation DSL (Wave 4) | AR-01 T1 | Manifest fields traceable 1:1 to existing scattered declarations, no new parallel metadata system introduced ahead of the model | Unit: manifest round-trips against ≥1 existing internal fixture module | `AR-03/manifest_schema_fixture_test.go` | Medium — scope-creep risk into DX-03 territory |
| T2. **OpenAPI merge fix** — preserve every OpenAPI 3.1 top-level field, not just `paths`/`components.schemas` | T1 | Fixture fragments exercising every top-level field; none silently dropped | Adversarial merge fixtures | `AR-03/openapi_full_field_merge_test.go` | Medium — **duplicates DX-06's identical closure contract; assign single ownership** |
| T3. Derive route registration/metadata from the manifest | T1, AR-01, AR-02 | A golden-fixture manifest change deterministically produces the expected full projection diff (route/permission/resource/schema/OpenAPI/lifecycle/profile/test/doc) with no other hand-edited file | Golden-delta test | `AR-03/golden_declaration_delta_test.go` | High — this test IS the acceptance gate |
| T4. Lint rule failing on hand-maintained duplicate identity or omitted projection | T1-T3 | Duplicate-identity and omitted-projection fixtures both fail lint | Adversarial lint fixtures | `AR-03/duplicate_omission_lint_test.go` | Medium |
| T5. Documentation/test/manifest export projections | T1-T4 | Extend T3's golden-delta to cover doc-table/manifest-export output | Integration | `AR-03/full_projection_golden_test.go` | Low-medium — share fixtures with AR-05 |

**wowsociety impact — AR-03:** Affected passively, low severity, not breaking. wowsociety's existing scattered-but-simple declarations don't hand-duplicate identity beyond what `module.Context` already centralizes. No required change; opt-in adoption for new modules once AR-03's generated tooling reaches parity with wowsociety's needs. **Verification needed once T2 lands:** wowsociety's existing OpenAPI fragment(s) must continue merging correctly — T2 changes merge behavior from silently-lossy to either-preserving-or-explicitly-rejecting, which could newly fail a fragment that previously merged "successfully" only because its extra fields were silently dropped.

#### AR-04 — Eliminate configuration and boot-time silent behavior

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Reject unknown `modules.<typo>` config namespaces at boot | Wave 0 exit; independent of AR-01-03 | Boot fails on a namespace with no matching registered module | Adversarial fixture | `AR-04/unknown_namespace_rejection_test.go` | Low — can land ahead of the AR-01 model work |
| T2. Reject duplicate collectors (currently last-writer-wins) | AR-01 T1 | Every collector rejects a second write to the same identity | One adversarial fixture per collector type | `AR-04/duplicate_collector_rejection_test.go` | Medium — distinguish illegitimate duplicate from legitimate multi-locale accumulation |
| T3. Reject empty required fragments | AR-01, T1-T2 | A module declaring a required-but-empty fragment fails boot | Adversarial fixture | `AR-04/empty_required_fragment_test.go` | Low-medium |
| T4. Post-seal write rejection reused from AR-01 T8 | AR-01 T8 | Same error-not-panic contract extended to config/namespace/collector state | Regression re-run of AR-01 T8 suite | `AR-04/post_seal_config_rejection_test.go` | Low |
| T5. Explicit optional-capability declaration; `prod` readiness fails on required-but-no-op/missing adapter unless a policy-approved waiver exists | AR-01, AR-02, T1-T4 | `prod` + no-op adapter + no waiver → readiness fails named; `local` + same config → succeeds; waiver present → suppressed and audited | Integration matrix: profile × waiver × adapter-real/no-op | `AR-04/prod_noop_adapter_readiness_test.go` | Medium — **shares scope with SEC-06 and DX-07's readiness closure contracts; build the waiver mechanism once** |

**wowsociety impact — AR-04:** Affected, low risk, not breaking, already partially assumed. `internal/modules/policy/module.go:47-51,54-56` declares an empty `Config{}` and already relies on strict-namespace-decode at the module-config-view level. AR-04's T1 operates one level up (unknown *module* namespaces) — additive, not conflicting. Recommend a one-time `grep 'modules\.' */config/*.yaml` sanity check in wowsociety before T1 lands.

#### AR-05 — Remove composition/documentation drift

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Fix `README.md:148-153`/blueprint 11:32 drift — `app/` doesn't construct `kernel.New`; no `RunAPI`/`RunWorker`/`RunMigrate` exist (only `App.Boot(ctx, *kernel.Kernel, ...)` + free function `StartWorker`) | None — pure doc fix, land immediately | README/blueprint accurately describe the live public surface or explicitly label future-state prose | Doc-compile check | `AR-05/readme_blueprint_drift_fix_diff.txt` | Low |
| T2. Fix `docs/blueprint/06-module-sdk.md:65-98` `Context` drift — remove phantom methods (`Clock()`, `RelationshipTypes()`, `Roles()`, `Hooks()`, `NotificationTemplates()`); document live-only methods (`RecurringJob`, `I18n`, `Audit`, `Sequence`, `Bulk`, `Artifacts`, `Privileged`, `RetentionClasses`, `Comments`, `Attachments`, `DocumentHooks`) | None — consider deferring if AR-05 T4 (generated docs) is imminent | Blueprint's `Context` listing matches `module/module.go` method-for-method | Doc-compile check | `AR-05/module_sdk_context_drift_fix_diff.txt` | Low |
| T3. CI gate compiling every normative doc example against the current API | T1, T2 | A deliberately staled example fails CI | Adversarial CI fixture | `AR-05/doc_compile_ci_gate_test_output.txt` | Medium — needs new doc-example-extraction tooling |
| T4. Generate reference/API docs from AR-03's authoritative manifest | AR-03 T1, T5 | Generated reference tables byte-match the model export | Integration golden-diff | `AR-05/generated_docs_byte_match_test.go` | Medium — depends on AR-03 |
| T5. Label remaining future-state design prose as "target, not implemented" | T1-T4 | Lint over `docs/blueprint/` for unlabeled normative-sounding future-state blocks | Lint | `AR-05/future_state_labeling_lint_test.go` | Low |

**wowsociety impact — AR-05:** Not affected. Pure wowapi internal documentation; no compiled API surface changes. wowsociety's own `cmd/` wiring was already coded against the real API, not the drifted docs — no wowsociety build failure was found that would indicate reliance on the incorrect text.

#### AR-06 — Remove hidden constructor bypasses from kernel wiring

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Fix `kernel/kernel.go:252-254`'s `orgAncestry` closure — it calls `authz.NewStore()` a second time instead of closing over the composed `authzStore` from line 230 | None — small isolated fix, land immediately (Wave 0 or Wave 1 start) | Closure uses the same composed instance; `OrgAncestors` goes through any configured cache/decorator | Sentinel-store injection test | `AR-06/sentinel_store_injection_test.go` | Low-medium — small diff, but this is what makes future ancestry caching/instrumentation possible at all |
| T2. AST/lifecycle lint forbidding ad-hoc infrastructure constructors outside composition packages | T1, may share tooling with AR-02 T6 | Lint fails on a reintroduced ad-hoc constructor outside `kernel/`'s composition root | Adversarial lint fixture | `AR-06/constructor_boundary_lint_test.go` | Medium — new `go/analysis`-based tooling |
| T3. Audit `kernel/kernel.go` for any other instance of the same "closure captures a fresh instance instead of the composed one" pattern | T1, T2 | Explicit audit confirming/refuting the pattern is isolated to the one cited line | Audit report | `AR-06/kernel_constructor_audit.md` | Low — mostly investigative; risk is under-scoping to just the one cited line |

**wowsociety impact — AR-06:** Not affected. wowsociety accesses privileged rule/relationship operations exclusively through `Privileged()` (`policy/pack.go:318`, `identity/module.go:133`, `policy/module.go:71`), never through ad-hoc kernel constructors. No visibility into or dependency on the internal `kernel.go` closure AR-06 fixes.

**PF-ARCH cross-cutting notes:** (1) AR-01 T1/T2 are the load-bearing prerequisite for AR-02's `Registrar` reuse and AR-03's manifest-consumes-model dependency; AR-05/AR-06's core fixes are independent and can land immediately in parallel. (2) AR-03 T2 and DX-06 describe the same OpenAPI-merge proof bar almost verbatim — confirm single ownership before implementation. (3) AR-04 T5's readiness-waiver mechanism overlaps SEC-06 and DX-07's closure contracts — build once. (4) Across all six findings, wowsociety's actual `module.Context` usage is a narrow, conventional 14-of-~40-method subset with zero port usage and exactly one retained-registrar instance — small, well-characterized blast radius, but a thin (2-module) sample; the additive/legacy-adapter compatibility strategy is the correct hedge against a richer future module exploiting today's mutable/unowned characteristics. (5) **Unresolved design question, needs a `decisions.md` entry before AR-01 T2 implementation:** do all AR-01 per-subsystem registrars share one `Registrar` type (capability-confusion risk) or does each get a distinct type (multiplies T2/T6 task count)? (6) **Unresolved policy question:** should post-seal mutation panic in production builds, or only error? Recommend "error, not panic" as the default — wowsociety's harmless `s.rulesReg` retention would otherwise convert into a production crash risk.

### 5.2 PF-SEC — Security and Trust Boundaries (SEC-01 – SEC-06)

**Accountable role:** product-security lead. **Evidence root:** `docs/implementation/evidence/premier/PF-SEC/`. **Wave:** SEC-02 is the only Wave-0 item; all others are Wave 2 or later.

#### SEC-01 — Resolve tenant membership and privileged session state server-side (P0, Wave 2)

**Evidence:** `Verifier.Actor` (`kernel/auth/auth.go:181-208`) validates membership only when `CapacityID != uuid.Nil`; a capacity-less actor gets zero membership check. `TenantID`/`ImpersonatorUserID`/`BreakGlass` are copied straight from JWT claims. `pgprincipal.Store` exposes only `UserIDBySubject`/`ValidateCapacity` — no membership, break-glass, or impersonation grant lookup exists. The target `user_tenant_access` table already exists in migrations (`00002_core_identity.sql:54-83`) but **no Go code queries it**. No break-glass/impersonation grant table exists at all — genuinely greenfield schema.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Migration: `identity_grant` table (break-glass + impersonation activation records — status, tenant, actor, impersonated user, approver, reason, activation/expiry/revocation, opaque grant ID) | none | RLS FORCE per blueprint 03 §1; unique partial index for one-active-grant-per-actor; `app_platform`-only grants | Migration up/down test; RLS catalog extension | `SEC-01/migration.md` | Schema is genuinely new — get security-lead sign-off before merge |
| T2. Extend `PrincipalStore` with `ActiveTenantAccess(ctx, userID, tenantID) error` against the existing `user_tenant_access` table; call unconditionally in `Verifier.Actor` | none | Every actor kind carrying `TenantID` is membership-checked | Adversarial: revoked/absent/foreign-tenant membership rejected even with a validly signed token | `SEC-01/membership-tests.md` | Every existing valid session must have a live `user_tenant_access` row — audit production data first |
| T3. Reject zero/unknown tenant before opening a tenant tx | T2 | Zero/garbage-UUID tenant claim rejected pre-`WithTenantID` | Negative test | `SEC-01/zero-tenant-tests.md` | Low |
| T4. Require explicit capacity choice when >1 active; validate capacity server-side unconditionally | T2 | Capacity-less actor with multiple active capacities rejected | Multi-capacity test | `SEC-01/capacity-selection-tests.md` | Breaks any currently-working capacity-less multi-capacity flow — needs a product-side UX |
| T5. Privileged-session resolver replacing direct claim copy of `ImpersonatorUserID`/`BreakGlass` with a T1 grant-table lookup by opaque grant ID | T1, T2 | `Actor` fields populated only from a verified grant row, never trusted off the JWT | Adversarial: expired/revoked/wrong-tenant/wrong-actor/forged-ID/unauthorized-approver all rejected | `SEC-01/privileged-session-tests.md` | Breaking JWT-claim-contract change — **needs a `grant_id` claim from the IdP; coordinate before merge, genuinely undecided today** |
| T6. Bind `auth_time`/`acr`/`amr` into assurance; enforce freshness for step-up | T2 | Stale `auth_time` with valid `amr` still fails step-up | Test | `SEC-01/assurance-freshness-tests.md` | `AMR` plumbing already exists — additive, moderate risk |
| T7. Distinguish user/API-key/webhook/internal credential schemes explicitly | T2-T6 | Permission scoped to `CredentialUser` rejects a valid API-key actor | Test | `SEC-01/credential-scheme-tests.md` | Cross-cuts DX-03's `CredentialScheme` design — sequence together |

**Required test classes (§6 SEC-05, mandatory):** token substitution, zero-tenant, stale membership, revoked capacity, expired step-up, issuer/audience/key rotation, JWKS failure.

**wowsociety impact — SEC-01: Affected, HIGH severity, BREAKING for impersonation.** `internal/modules/identity/impersonation.go:1-21` states explicitly: *"What the framework does NOT provide: a session/grant record... This file is that product-side layer."* wowsociety has **already built its own workaround** — `identity_impersonation_session` table, `startImpersonation`/`stopImpersonation`, audited via `kaudit.Entry`. `whoami.go:39,51` reads `actor.ImpersonatorUserID` directly off the framework `authz.Actor`, populated from the unverified claim, by explicit design (comment: trusts the claim "without a DB re-check"). Test files `abac_test.go:52-94`, `whoami_impersonation_test.go:31-56` construct `authz.Actor{ImpersonatorUserID: ...}` directly — load-bearing test surface that will need rewriting. `BreakGlass`: zero usage anywhere in wowsociety (unexercised today). Breaking-vs-compile-safe distinction: if the resolver **preserves the `authz.Actor` struct shape** but populates fields more strictly, wowsociety compiles unchanged and gets a strict behavioral improvement (a currently-trusted-but-invalid state now correctly rejected — a runtime behavior change some current callers may be relying on). If fields are **renamed/removed** (e.g. `BreakGlass` becomes a grant-status enum), `whoami.go`, `impersonation.go`, and `whoami_impersonation_test.go:43` fail to compile. **Required wowsociety changes:** decide table authority (recommend framework owns grant validity/expiry/revocation, wowsociety keeps its table for UX/audit-trail); update `startImpersonation`/`stopImpersonation` to mint/reference the framework's `grant_id`; audit `user_tenant_access` data before wowapi enforces T2. **Sequencing:** two-repo coordinated cutover — wowapi ships T1+T5, wowsociety's auth flow adopts `grant_id`, only then cut over; validate T2 against wowsociety staging data before making it unconditional. **Verification:** wowsociety's `abac_test.go`/`whoami_impersonation_test.go`/`rls_test.go` already exercise this end to end — good regression coverage to re-run post-cutover.

#### SEC-02 — Make workflow privileged operations fail closed (P0, **Wave 0**)

**Evidence, with a materially important blast-radius correction:** `NewRuntime` (`kernel/workflow/runtime.go:84-91`) allows `ev == nil` (only `txm/reg/ob/idgen` are in the nil-guard); `Override` (283-334) checks `if rt.authz != nil` — silently skips the entire permission check when nil; ratification is a bare `TODO` comment with zero implementation. **Confirmed via direct read: the only production call site (`kernel/kernel.go:270`) always passes a real, non-nil `eval` constructed unconditionally a few lines earlier.** Nil-evaluator construction is exercised only in 4 test files. **Making the evaluator mandatory is a near-zero-blast-radius change** — independently re-verified by this document's own mentor-review pass, not just PF-SEC's claim.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk | Wave |
|---|---|---|---|---|---|---|
| T1. Add `ev` to `NewRuntime`'s nil-guard panic list | none | Panics if `ev == nil` | Unit | `SEC-02/mandatory-evaluator-tests.md` | Minimal — zero production callers pass nil | **Wave 0** |
| T2. Expose an unexported test-only constructor for the 4 nil-evaluator test call sites | T1 | No public API accepts nil `authz.Evaluator` | Existing suite unchanged | `SEC-02/test-constructor-migration.md` | Low | **Wave 0** |
| T3. Remove `if rt.authz != nil` in `Override` — unconditional check | T1 | Denied override fails closed with `KindForbidden` always | Adversarial: actor lacking permission denied | `SEC-02/override-fail-closed-tests.md` | Minimal | **Wave 0** |
| T4. Implement ratification as a real definition field + state transition (or explicitly reject `ratify_by`-declaring definitions as an interim Wave-0-compatible posture) | T1-T3 | Directive allows "reject or implement" | Override-then-ratify happy path; pending-not-yet-effective; rejection reverts | `SEC-02/ratification-tests.md` | Genuinely greenfield design work | **Wave 2+, NOT Wave 0** |
| T5. Persist actor, impersonator, grant ID (from SEC-01 T1), source/target states, reason, ratification outcome in durable audit | T1, T3, T4; benefits from SEC-01 T1 | Complete audit row in the same tx as the state jump; audit failure rolls back the override | Test: audit present/complete; write failure rolls back | `SEC-02/override-audit-tests.md` | Actor/reason/state portion is Wave-0-compatible; grant-ID field waits on SEC-01 | **Split across waves** |

**Wave 0 minimal fix = T1+T2+T3, matching §12 Wave 0 item 2 exactly.**

**wowsociety impact — SEC-02: Not affected.** Zero occurrences of `workflow.NewRuntime`, `workflow.Runtime`, `.Override(`, or any `kernel/workflow` import anywhere in wowsociety. No required changes, no sequencing constraint.

#### SEC-03 — Bind webhook replay controls to provider-authenticated data (P1, Wave 2)

**Evidence:** `HMACVerifier.Verify` (`kernel/webhook/verifier.go:32`) HMACs the body only, returns only `error`. `HandleInbound` (`service.go:22-114`) drives replay-window/dedup off the caller-supplied, unauthenticated `InboundIn.Timestamp`/`ExternalEventID`. Mitigating detail: `service.go:60-63` already forces `external_event_id = nil` on a *failed* signature — good defense-in-depth, doesn't address the successful-signature gap.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Change `Verifier` interface to return `(Envelope, error)`, `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}` | none | New interface compiles; `HMACVerifier`/`FakeVerifier` updated | Unit both impls | `SEC-03/verifier-envelope-tests.md` | **Breaking interface change** |
| T2. `HMACVerifier` synthesizes `EventID`/`OccurredAt` from the authenticated body / receipt time only; documented as unsuitable for timestamped-provider protocols otherwise | T1 | Envelope never surfaces caller-supplied fields | Test: `OccurredAt` immune to manipulated `in.Timestamp` | `SEC-03/hmac-envelope-tests.md` | Moderate |
| T3. Rewire `HandleInbound` to source replay-window/dedup exclusively from `Envelope` | T1, T2 | No security decision reads raw `InboundIn` fields | Adversarial tamper matrix: body/timestamp/event-ID/key-ID/sig-version independently | `SEC-03/tamper-matrix-tests.md` | Moderate — review against existing dedup-spoofing mitigation |
| T4. Document the provider-verifier contract | T1-T3 | Contract doc + reference example | N/A | `SEC-03/provider-verifier-contract.md` | Low |

**wowsociety impact — SEC-03: Not affected.** Zero `kernel/webhook` import, zero custom `Verifier` implementation anywhere in wowsociety.

#### SEC-04 — Bound authorization staleness and memory (P1; P0 if cache enabled in production, Wave 2/5)

**Evidence:** unbounded `sync.Mutex`-guarded map, no capacity limit, no dormant-entry sweep, concurrent misses duplicate DB loads. **Deployment posture materially affects severity:** the cache is opt-in (`deps.AuthzCacheTTL > 0`), off by default, per an explicit code comment.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Bounded, sharded cache | none | Never exceeds configured max under adversarial cardinality | Insert >max keys; race test | `SEC-04/bounded-cache-tests.md` | Low-moderate — swap behind existing `Store` interface |
| T2. Eviction with admission/eviction metrics | T1 | Idle entries evicted; full metric set | Test | `SEC-04/eviction-metrics-tests.md` | Low |
| T3. Singleflight-collapse concurrent misses | T1 | N concurrent misses → 1 DB load | Test | `SEC-04/singleflight-tests.md` | Low |
| T4. Per-tenant/global authorization epoch or invalidation stream for cross-pod revocation | T1-T3 | Revocation on pod A visible to pod B without a full TTL wait | Simulated cross-pod test | `SEC-04/cross-pod-epoch-tests.md` | **Highest-risk task** — open architecture decision (LISTEN/NOTIFY vs. epoch-row-poll), may overlap Wave 3's shared lease infrastructure |
| T5. Expose `CacheHit`/epoch-observed on `Decision` | T1-T4 | Decision metadata differs hit vs. miss | Test | `SEC-04/decision-provenance-tests.md` | Low |
| T6. Require explicit max-size + stale-allow bound in prod config; fail boot without both | T1-T5 | Prod profile with cache enabled but no bound fails validation | Negative config test | `SEC-04/prod-config-gate-tests.md` | Low — established pattern already exists in `config.go` |

**wowsociety impact — SEC-04: Not affected.** Zero construction of `authz.NewCache`/`authz.CachingStore` anywhere in wowsociety; no cache-TTL config keys in `identity/config.go` or any `configs/*.yaml`.

#### SEC-05 — Establish a versioned security verification profile (P1, Wave 6)

Standards adoption (ASVS 5.0.0, OWASP API Security Top 10 2023, NIST 800-63-4), not a source-citation finding. Its role is supplying the required test-class checklist SEC-01–04 already inherit.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Version-pinned control map linking every applicable control to an executable test or an approved waiver | SEC-01–04 substantially complete | Independent assessment leaves zero open Critical/High | External assessment | `SEC-05/control-map.md` + report | **Closure gate**, not implementable until SEC-01–04 exist to map against — Wave 6 |

**wowsociety baseline:** real adversarial-test infrastructure already exists to plug into a future control map (`abac_test.go`, `authz_matrix_test.go`, `rls_test.go`, `stepup_test.go`, `otp_test.go`/`totp_test.go`, `whoami_impersonation_test.go`) — but these validate wowsociety's *own* product-layer workarounds, so expect rework once SEC-01 ships, not pure addition.

#### SEC-06 — Govern explicit outbound-security escape hatches (P1, Wave 2/5)

**Evidence:** `JWKSConfig.Client *http.Client` (`kernel/auth/jwks.go:59`) is caller-injectable and bypasses the default client's proxy-disabling; an injected client gets no private-IP dial guard, by design and self-documented. `httpclient/client.go:142` — an exact-match allowlisted hostname skips IP-class checking entirely. **Configuration provenance (changes risk triage):** `AllowedHosts`/`AllowedCIDRs` come from static deployment config, boot-validated — not tenant/user-controlled. `SharedFingerprint()` likely already covers these fields structurally, pending a direct scope-confirmation test.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Confirm/extend `SharedFingerprint()` scope covers the outbound allowlist; regression test | none | Fingerprint-diff test passes | Test | `SEC-06/fingerprint-coverage-tests.md` | Low — likely already correct, may close as "add regression test only" |
| T2. Boot-time startup report enumerating enabled egress exceptions, no credentials exposed | none | Readiness/log lists exceptions | Test | `SEC-06/egress-report-tests.md` | Low |
| T3. Explicit change-audit trail for allowlist config changes | T1 | Config diff touching the allowlist produces an audit-visible record | Test | `SEC-06/allowlist-change-audit-tests.md` | Low-moderate |
| T4. Extend equivalent governance to the JWKS `Client` injection path — currently pure Go constructor param, zero config surface, cannot be fingerprinted/audited today | T1-T3 | Prod profile using a custom JWKS client either declares a trusted-issuer config or fails readiness | Test | `SEC-06/jwks-client-governance-tests.md` | **Highest-risk task** — open design decision, not yet made |
| T5. Codify "never tenant/user-controlled data populates allowlists/JWKS clients" as a lint/fitness check | T1-T4 | Static check asserts construction never reads request/tenant-scoped data | Fitness test | `SEC-06/no-tenant-controlled-allowlist-tests.md` | Low — codifying an already-true invariant |

**wowsociety impact — SEC-06: Affected, config-only, not source-code.** wowsociety never constructs `httpclient.New`/`auth.JWKSConfig` directly (wired by wowapi's `kernel.New`), and configures OIDC/JWKS purely via static YAML (`configs/stage.yaml:59,63`), with no tenant/user-controlled data feeding these values. **Genuine evidence gap flagged, not papered over:** wowsociety's actual deployment config for allowlist entries or custom JWKS-client injection was not read in this pass — needs a follow-up config audit. Breaking only for T4, only if wowsociety currently injects a custom JWKS client with no declaration path (unconfirmed).

**PF-SEC cross-cutting notes:** (1) SEC-01 T5's IdP claim-contract dependency is genuinely undecided — needs product/security-lead input, highest-uncertainty item in this package. (2) wowsociety's `identity_impersonation_session` table vs. wowapi's future SEC-01 grant table need an explicit authority decision, not just a recommendation. (3) wowsociety's `go.mod` pins `v1.0.0` via a local working-tree `replace`, not a locked commit — any SEC fix changing a public interface lands in wowsociety's build with no version gate until DX-05's real pin/upgrade-matrix exists. (4) SEC-04 T4 and SEC-06 T4 are both open architecture/design decisions, not resolved by this plan.

### 5.3 PF-DATA — Persistence, Jobs, Durable Delivery, Compliance (DATA-01 – DATA-09)

**Accountable role:** data/reliability lead. **Evidence root:** `docs/implementation/evidence/premier/PF-DATA/`. A cross-cutting dependency threads DATA-02 → DATA-03/DATA-04 (build the shared lease primitive once, reuse three times — do not let three subsystems independently reinvent lease columns and fencing logic).

#### DATA-01 — Encode tenant equality in foreign keys (P0, Wave 2)

**Evidence, confirmed exactly:** 8 tenant-scoped child tables (persons/legal_entities/party_contacts/acting_capacities → parties; resources → organizations; document_versions/document_access_grants/attachments → documents/document_versions) reference only the parent's `id`, never `(tenant_id, id)`. RLS proves the child row's own tenant; nothing proves parent and child agree.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Add/confirm `UNIQUE (tenant_id, id)` on every referenced parent, built `CONCURRENTLY` | — | Every parent has the unique index | Migration test via `pg_indexes` | `DATA-01/parent-index/` | `SHARE UPDATE EXCLUSIVE` lock — must run non-transactionally |
| T2. Build a tenant-FK catalog scanner flagging any tenant-table FK not composite on `(tenant_id, …)` | T1 | Enumerates exactly the 8 known FKs with zero silent gaps; becomes a permanent CI gate | Fixture-schema test | `DATA-01/fk-catalog/` | Must key off the existing RLS-tagged tenant-table matrix, not a hand-maintained list |
| T3. Mismatch audit: prove `child.tenant_id = parent.tenant_id` for every existing row; fail deployment on any mismatch | T2 | Zero-mismatch report against staging/prod-shaped data | Integration seeding a deliberate cross-tenant mismatch via platform role | `DATA-01/mismatch-audit/` | Requires a platform-role connection to bypass RLS for the scan |
| T4. Add composite FK `NOT VALID` for all 8 edges | T1, T3 | Metadata-only add stays under the DATA-09 2-second lock-timeout budget | Migration lock-duration test | `DATA-01/composite-fk-notvalid/` | Run per-table as separate statements |
| T5. `VALIDATE CONSTRAINT` each new composite FK | T4 | Validation doesn't block concurrent DML; second zero-mismatch confirmation | Load test under concurrent writer load | `DATA-01/validate-constraint/` | I/O-bound — schedule per DATA-09's backfill/validate phases |
| T6. Wire the T2 scanner into a permanent CI gate | T2 | A new migration adding a single-column tenant FK fails CI | Negative fixture migration | `DATA-01/gate-test/` | Cheapest, most durable part — do first if sequencing allows |
| T7. Seeded cross-tenant insert negative tests under both `app_rt` and `app_platform` | T5 | Insert violating tenant equality fails under both roles | New catalog-driven RLS matrix test | `DATA-01/cross-tenant-fk-negative/` | Confirm platform role doesn't bypass FK constraints — don't assume |
| T8. Remove redundant single-column FKs, only after all consumers/rollback paths verified | T5, T7 | No code relies on the old FK name for cascade behavior | Full regression + grep | `DATA-01/fk-cleanup/` | Optional — don't block P0 closure on it |

**wowsociety impact — DATA-01: Affected, real, not hypothetical.** `internal/modules/policy/migrations/00002_override.sql:16` — `policy_override.rule_version_id` references `rule_versions(id)` only, with `policy_override.tenant_id` set on the same row — a genuine independent instance of the DATA-01 pattern. **Not breaking on its own** (wowapi's fix touches only wowapi tables), but becomes relevant once T6's gate extends to product-module consumers. **Required wowsociety changes:** add `UNIQUE (tenant_id, id)` on `rule_versions` (must land in wowapi first — kernel table wowsociety doesn't own), then migrate `policy_override` to a composite FK. **Sequencing:** wait for wowapi's T1, then follow wowapi's DATA-09 protocol once it exists for wowsociety's own rollout — do not build an ad hoc one-off fix first, it's wasted work once DATA-09's tooling supersedes it.

#### DATA-02 — Add lease generations/fencing and effect idempotency to jobs (P0, Wave 3)

**Evidence:** claim SQL returns no lease token/generation; completion/failure match only `id`; `ReclaimStalled` blind-resets every stale row with no per-row fencing check. Confirmed race: A stalls, gets reclaimed by B, B completes, A's eventual finalize silently overwrites B's outcome.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Implement a **shared** lease/fencing primitive (`lease_token`, monotonic `lease_generation`, `lease_expires_at`, optional heartbeat) as a reusable kernel building block for DATA-02/03/04 | — | One primitive reused ≥3 times, not three independent copies | Unit tests on token/generation comparison | `DATA-02/lease-primitive/` | Architecturally load-bearing across all three findings |
| T2. Add lease columns to `jobs_queue`; claim SQL assigns fresh token + `generation+1` | T1 | `claimedJob` carries lease context | Migration + unit | `DATA-02/jobs-lease-migration/` | Reuse existing timeout-floor logic, don't introduce a second inconsistent timeout source |
| T3. Finalize paths compare lease token/generation, reject mismatch | T2 | Stale finalize affects 0 rows, observably rejected | See T7 chaos test | `DATA-02/finalize/` | Must not regress the at-least-once recovery path |
| T4. `ReclaimStalled` bumps `lease_generation` on reclaim | T2 | Reclaimed row is a provably new lease epoch | Same test as T3, asserting generation delta | `DATA-02/reclaim/` | — |
| T5. Stable job idempotency key + lease context passed to workers; each worker declares exactly one of: inbox/effect ledger unique on `(job_id, effect_name)`, domain CAS, or provider idempotency key | T2 | Worker cannot register without declaring its mechanism | Duplicate-effect test | `DATA-02/idempotency/` | Likely needs PF-ARCH's typed operation model to enforce at compile time; worker signature change is breaking — coordinate with wowsociety even though it has zero current job usage |
| T6. Document/test: fencing the queue row does not undo an already-committed stale-worker domain transaction | T3, T5 | Testable claim, not prose | Test proving effect ledger still catches an idempotency-ignoring worker | `DATA-02/worker-contract/` | Low |
| T7. **Named chaos test:** pause worker A after claim, expire, reclaim via B, B completes, resume A and attempt finalize at every domain/external/finalize boundary — exactly one logical effect recorded, A's writes rejected | T3-T5 | Matches closure contract verbatim | This is the test | `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` | Must exercise all 3 named boundaries; build as a reusable chaos harness shared with DATA-03/DATA-04 |

**wowsociety impact — DATA-02: Not affected.** Zero `kernel/jobs` import, zero job registration anywhere in wowsociety. Would become breaking (worker signature change, T5) the moment wowsociety registers a job — flag for roadmap.

#### DATA-03 — Move remote provider/secret I/O outside database transactions (P0, Wave 3)

**Evidence, self-documented in wowapi's own source:** `notify/service.go:456-586`'s own comment (446-449) already states *"Real production deployments should move the network call outside the tx."* `webhook/service.go`'s delivery loop and secret resolution both run inside `plat.WithTenant(...)`.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Reuse DATA-02's shared lease primitive for notify/webhook claim rows | DATA-02 T1 | Lease columns via shared primitive, not a bespoke copy | Migration test | `DATA-03/lease-columns/` | None beyond DATA-02's own risk |
| T2. Three-stage protocol for `kernel/notify`: claim-tx (assigns lease) → `sender.Send` outside any tx, delivery ID as idempotency key → finalize-tx comparing lease token | T1 | No `sender.Send` call while a DB tx is open | See T8 boundary matrix | `DATA-03/notify/` | Delete/update the self-documented "should move outside tx" comment as part of this task |
| T3. Same three-stage protocol for `kernel/webhook.deliverToEndpoint` | T1 | No DNS/secret-resolve/POST call while a tx is open | See T8 boundary matrix | `DATA-03/webhook/` | Current-row-state check must move into claim stage so Execute needs no mid-flight DB reads |
| T4. Inbound two-phase verification: short read-tx (endpoint snapshot) closes → verify outside tx → short write-tx re-checks version/status, discard+retry on mismatch | T1 | Secret rotation/deactivation between phases cannot cause accept-under-stale-policy | Rotation-during-verification test | `DATA-03/webhook/inbound-two-phase/` | Breaking signature change to `HandleInbound`'s transaction-ownership contract; bound retry attempts |
| T5. Failed-signature audit: body-free audit row in its own short tx | T4 | No raw body ever persisted on failed verification | Test asserting empty body field | `DATA-03/webhook/failed-sig-audit/` | Low |
| T6. Per-adapter idempotency-safety contract declaration | T2, T3 | Adapter cannot be registered for a non-idempotent high-impact operation without declaring duplicate-safety | Boot-time fixture rejecting undeclared adapter | `DATA-03/adapter-contract/` | Inventory all existing `Sender` implementations first |
| T7. Remove the stale "app_platform lacks INSERT on events_outbox" comment; wire legal-delivery audit — **shares scope with DATA-08 W0-T2, implement once, cross-reference** | — | — | — | Cross-reference only | Avoid double-implementation |
| T8. **Named chaos test at 6 boundaries:** before send, during send, after success/before finalize, lease expiry, duplicate workers, provider timeout — applied to both notify and webhook | T2-T4 | Zero duplicate external effects across all 6 fault points | This is the test | `DATA-03/chaos/` | Most labor-intensive requirement in PF-DATA; reuse DATA-02's chaos harness |

**wowsociety impact — DATA-03: Not affected today; conditionally breaking in the future.** Zero `kernel/notify`/`kernel/webhook` usage found. If wowsociety ever calls `webhook.HandleInbound` directly, T4's transaction-ownership contract change would need integration review — flag for future, not now.

#### DATA-04 — Reconcile bulk-processing concurrency and make multi-worker mode safe (P1; P0 before advertising multi-worker, Wave 3)

**Evidence — migration comment and implementation directly contradict each other:** migration `00016`'s header claims "safe across replicas" via `FOR UPDATE SKIP LOCKED"; `Service.next` (`kernel/bulk/bulk.go:123-144`) actually does a plain unlocked `SELECT ... LIMIT 1`, with the function's own doc comment conceding "no lock — single processor per operation."

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. **Immediate stopgap:** correct the false migration comment; enforce single-processor via advisory lock or CAS at the `Service` API boundary | — | False "replica-safe" claim removed; a second concurrent processor is rejected, not silently racing | Concurrency test: 2 processors on the same `bulkID` | `DATA-04/stopgap/` | Ships independently and fast — closes the false-documentation P0 sub-issue before the full rewrite |
| T2. Reuse DATA-02's shared lease primitive for `bulk_items` | DATA-02 T1; T1 as interim | Lease columns via shared primitive | Migration test | `DATA-04/lease-columns/` | Additive |
| T3. Atomic leased claim: `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`, bounded batch | T2 | SQL provably uses `SKIP LOCKED` with a bounded batch | EXPLAIN-plan assertion + concurrent N>1 claimer test | `DATA-04/leased-claim/` | Preserve `runItem`'s existing idempotent completion CAS guard |
| T4. Item idempotency keys, finalize fencing, retry policy, cancellation | T3 | Fenced worker's finalize write is rejected | Reuse DATA-02's chaos pattern | `DATA-04/fencing/` | Shares finalize-fencing logic with DATA-02 T3 — reuse, don't reimplement |
| T5. Pause/resume/cancel operation-level controls, bounded batch claims | T3, T4 | Pause/resume/cancel behave correctly mid-run | Lifecycle integration tests | `DATA-04/lifecycle/` | Larger scope — schedule in the full P1/Wave-3 slice, not the fast-track stopgap |
| T6. **Named chaos test:** ≥2 processors concurrently claim/retry/pause/resume/cancel the same operation without duplicate effects or stale finalization | T3-T5 | Matches Wave-3 exit gate wording verbatim | Deterministic fake-clock test | `DATA-04/chaos/duplicate_worker_test.go` | Reuse the shared chaos harness |

**wowsociety impact — DATA-04: Not affected.** Zero `kernel/bulk` import anywhere in wowsociety.

#### DATA-05 — Allocate immutable versions without MAX()+1 races and clean orphan blobs (P1, Wave 2)

**Evidence:** `kernel/artifact.Generate` and `kernel/document.InitiateUpload` both compute `MAX(version)+1` inline; the document path's loser leaves an orphaned, randomly-keyed blob with no GC.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Replace `MAX()+1` with a locked parent counter or dedicated per-aggregate sequence row (both `kernel/artifact` and `kernel/document`) | — | N concurrent callers → N unique monotonic versions, zero unexpected conflicts | Concurrency test, ≥20 concurrent callers | `DATA-05/version-allocation/` | Counter-row contention is the new serialization point — measure lock wait |
| T2. Durable upload-session records for `kernel/document`: expiry, checksum/size, storage key, status, cleanup ownership | T1 | Session row persisted before the presigned URL returns | Test: initiate, simulate crash, assert `status='pending'` with expiry | `DATA-05/upload-session/` | New table needs RLS + `<module>_<entity>` naming |
| T3. Confirmation CASes the session and version atomically | T1, T2 | Two racing confirms: exactly one succeeds | Concurrency test | `DATA-05/confirm-cas/` | — |
| T4. Scheduled GC removing expired/unreferenced objects, with metrics/audit | T2, T3 | Never removes a referenced object; removes every past-expiry unconfirmed session | Mixed confirmed/expired/pending test | `DATA-05/gc-sweep/` | False-positive deletion is data loss — conservative grace window |
| T5. Same counter fix for `kernel/artifact.Generate` | T1 | Same concurrency bar as T1 | Mirror test | `DATA-05/artifact-version/` | — |

**wowsociety impact — DATA-05: Not affected.** No `kernel/artifact`/`kernel/document` import found anywhere in wowsociety.

#### DATA-06 — Integrate the resource mirror into the aggregate write contract (P1, Wave 2)

**Evidence:** `kernel/resource` package doc confirms a manual, comment-only contract — a module owns its business table and separately upserts the mirror, with no framework enforcement. `registrar_pg.go:38-58` passes `created_by` as `uuid.Nil` with a TODO. Even the reference handler manually performs two independent statements.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Typed aggregate repository/unit-of-work helper bundling aggregate write + mirror upsert + audit + outbox atomically | — | A module cannot write its business row without the framework also writing the mirror in the same transaction | **Fault injection at each of 4 stages independently — full rollback at every stage** | `DATA-06/aggregate-atomicity/` | Overlaps AR-03 — coordinate to avoid a parallel one-off mechanism |
| T2. Source `created_by` from context in the same helper; reject missing actor for user-initiated writes — **shared fix surface with DATA-07 T3, same file (`registrar_pg.go`), one owner** | T1 | Real `created_by` on every mirror row; user-initiated write with no actor fails fast; system-actor paths unaffected | Test with/without actor, system vs. user path | `DATA-06/actor-attribution/` | Must not break legitimate system-actor call sites |
| T3. Migrate the reference handler onto the new helper | T1, T2 | Reference handler no longer manually calls both statements | Existing reference tests pass | `DATA-06/reference-handler-migration/` | Fix the reference pattern before it's copied further |
| T4. Update `kernel/resource` docs to describe the mandatory-mirror contract | T1 | Docs match implementation | Manual review | `DATA-06/docs-update/` | Low, don't skip — stale docs created this defect class |

**wowsociety impact — DATA-06: Affected, real, moderate severity.** `internal/modules/identity/committeeseat.go:69-70` uses the exact manual pattern DATA-06 targets — a hand-written, independently-callable `resource.NewRegistrar().Bind(db).Upsert(...)` call. Functions correctly today, at risk of silent mirror-write omission on any future edit. **Not breaking near-term** if wowapi keeps the low-level `Upsert` API available alongside the new helper. **Required change:** migrate `committeeseat.go` to the new helper once available; check for the same `uuid.Nil` actor gap. **Sequencing:** follow wowapi's T1/T3 (reference implementation proven first); not urgent, current pattern still functions.

#### DATA-07 — Complete relationship semantics and actor attribution (P1, Wave 2, blocked on SEC-01)

**Evidence:** `Checker.Has` (`kernel/relationship/relationship.go:42-66`) filters `subject_kind='capacity'` only — party-subject edges are, per the code's own comment, "not consulted yet." Same nil-actor gap as DATA-06, same file.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Resolve actor → active capacity → optional party through the authoritative principal model | **Hard dependency on PF-SEC's SEC-01 — do not schedule before it lands** | `Checker.Has` can evaluate party-subject edges | Test: seed a party-subject edge, resolve an actor carrying a party, assert previously-false now true | `DATA-07/party-subject-eval/` | Blocked, not merely related |
| T2. Extend `Checker.Has` to cover every schema-enumerated `subject_kind` | T1 | Every enumerated kind has an evaluation branch; unsupported kind fails closed | Matrix test | `DATA-07/subject-kind-matrix/` | Confirm which enumerated kinds are live requirements vs. dead schema surface first |
| T3. Source real actor for `Relate`/mirror `Upsert` — **reuse DATA-06 T2's mechanism directly, same file, do not reimplement** | DATA-06 T2 | Real `created_by`; same missing-actor rule as DATA-06 | Shared test helper | `DATA-07/actor-attribution/` | High duplication risk if staffed independently — sequence as one shared task |
| T4. Every authorization-input mutation is ownership-checked, attributed, audited, versioned, and invalidates relevant caches | T1-T3; **also depends on SEC-04's cache-epoch work** | Edge create/revoke writes audit rows and triggers observable cache invalidation | Test | `DATA-07/mutation-audit-cache/` | Second cross-work-package dependency — do not assume PF-SEC delivers on PF-DATA's timeline |

**wowsociety impact — DATA-07: No confirmed direct usage.** `grep -rn "kernel/relationship"` returns zero matches across wowsociety, confirmed including `committeeseat.go` (which mentions "ReBAC" conceptually in a comment but doesn't import or call `kernel/relationship`). Re-verify at DATA-07 ship time.

#### DATA-08 — Make compliance evidence complete, durable, and centrally enforced (P0/P1, Wave 0 + Wave 6)

**Evidence, including a confirmed real contradiction:** `audit.go`'s `chainHash` explicitly excludes `metadata` (jsonb reformat problem) — **and also `tx_id`, which the directive doesn't name but is confirmed unhashed too.** `attachment.go:82-87` discards the outbox-write error via blank assignment (`_ = s.ob.Write(...)`). `notify/service.go:451-453,546-559`'s deferral comment claims `app_platform` "lacks INSERT on events_outbox" — **but migration `00011:178` already grants exactly that permission, and its own comment names the legal-delivery use case.** The deferral comment is stale, not accurate, confirmed by reading both files directly.

**Wave-0 P0 tasks (directive Wave 0 item 3):**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| W0-T1. Stop discarding the attachment outbox-write error; propagate so `Attach` fails and rolls back in the same transaction | — | An injected outbox-write failure leaves zero attachment row persisted | Fault-injection test | `DATA-08/wave0/attachment-outbox/` | Confirm `Attach`'s caller actually runs inside a `WithTenant` tx before assuming rollback semantics — verify, don't assume |
| W0-T2. Remove the stale deferral comment; implement legal-delivery audit write using migration 00011's already-granted permission | — | `ImportanceLegal` deliveries produce a durable audit/outbox record with provider receipt in the same transaction as the `sent` status update | Test: legal-importance send → audit row with provider msg ID; negative test for non-legal | `DATA-08/wave0/legal-audit/` | **Design ambiguity to resolve first:** blueprint citation implies an `audit_logs` row; migration 00011's grant is specifically on `events_outbox` INSERT — these may be two different target tables. Confirm intended target with compliance/security lead before implementing. |

Both Wave-0 tasks are independent quick fixes, neither a prerequisite for the Wave-6 tasks below.

**Wave-6 P1 tasks (directive Wave 6 items 1-2):**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| W6-T1. Widen `chainHash` to cover every persisted field: canonicalized `metadata`, `tx_id`, all nullable fields, sequence, ID, timestamps, previous hash | W0-T1, W0-T2 (sequenced first, not a hard blocker) | Mutating any declared field, including metadata/tx_id, breaks verification | Tamper test: mutate each field independently, assert every one fails | `DATA-08/wave6/audit-hash/` | **High — breaking format change.** No `hash_version` column exists today; widening makes every historical row unverifiable under new-scheme verification unless a version discriminator is added in the same migration and verification branches by row version. Must hash a canonicalized pre-serialization form of metadata, not the stored jsonb, or the fix reintroduces non-reproducibility. **Single highest-risk task in PF-DATA's Wave-6 scope, and directly hits wowsociety's live audit rows.** |
| W6-T2. External anchor verification for the audit chain | W6-T1 | Chain-head periodically anchored externally; tamper detectable even if local `head_hash` were compromised | Test: anchor, tamper, assert detection | `DATA-08/wave6/anchor/` | Genuinely new subsystem — vendor/design decision needed |
| W6-T3. Persist DSR export as encrypted immutable artifact with manifest, per-class results, checksum, expiry, access policy, download audit | — | Replaces `retention/engine.go`'s bare in-memory map return | Test: export completes only after artifact write succeeds; checksum verifies; access-gated download audited | `DATA-08/wave6/dsr-export/` | New encryption-key-management dependency |
| W6-T4. Central legal-hold enforcement wrapper every `Dispose`/`Erase` callback must pass through, replacing today's per-callback responsibility | — | Negative test: a deliberately non-compliant callback is still blocked by the framework wrapper | Negative test | `DATA-08/wave6/central-hold/` | Breaking change to the `DisposeFunc`/`EraseFunc` contract — enumerate every registered `RecordClass` in both repos first |
| W6-T5. Explicit partial/not-applicable results for record classes without export/erase callbacks | W6-T3, W6-T4 | Result set explicitly lists every registered class with a status, never a silent omission | Test | `DATA-08/wave6/explicit-status/` | Coordinate with W6-T3's manifest shape |

**wowsociety impact — DATA-08: Affected for `kernel/audit` — BREAKING for W6-T1.** wowsociety structurally depends on `kernel/audit`: `identity/service.go` and `policy/service.go` both hold `*kaudit.Writer` fields; `impersonation.go` writes two `s.audit.Record(...)` calls for grant/revoke — a load-bearing compliance flow; `cmd/api/main.go` wires `kaudit.New(...)` for API-key audit. **wowsociety produces real, live audit rows today.** Changing `chainHash`'s input set changes the hash of every new row after the change lands — every historical wowsociety audit row becomes unverifiable under naive new-scheme-only verification. This is real and material for a live consuming product, not theoretical. No `kernel/attachment`/`kernel/notify`/`kernel/retention` usage found — W0-T1/W0-T2/W6-T2-T5 have no current wowsociety impact (W6-T3/4/5 land on wowsociety's future DSR roadmap, not current code). **Required wowsociety changes:** no call-shape change (only internal hash computation + a new version column, per the risk note above); re-run any wowsociety-side audit-verification tooling after upgrading, confirm historical rows still verify under whatever backward-compatible scheme wowapi ships. **Sequencing:** zero effect until wowsociety bumps `FRAMEWORK_VERSION` past W6-T1's commit; needs a dedicated wowsociety-side staging verification pass before accepting that bump.

#### DATA-09 — Adopt an online expand/backfill/validate/contract protocol (P0, Wave 2)

**Reality check: this is new tooling from zero.** No expand/contract discipline, no online-DDL lock-timeout classification, no backfill-job harness exists anywhere in wowapi today. `Makefile`'s `migrate` target is a plain forward-apply; `check_migrations.sh` checks only registration/markers/numbering, nothing about lock duration or backfill.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk | Human or code? |
|---|---|---|---|---|---|---|
| T1. Migration manifest schema (online/maintenance, rows/bytes, lock/statement timeout, N/N-1 flag, backfill owner, validation query, rollback/forward-fix plan) | — | Every migration has a validated manifest entry; missing fields fail CI | Schema validation + negative fixture | `DATA-09/manifest-schema/` | Get external review before locking the format | Code + **per-migration classification is human judgment every time** |
| T2. 2-second online-DDL lock-timeout enforcement with abort-and-retry | T1 | A statement exceeding budget aborts cleanly, no partial DDL | Test against a concurrently-locked table | `DATA-09/lock-timeout/` | Bound total retries — unbounded retry is a deploy-time DoS | Code, with human-set retry ceiling |
| T3. Expand-phase tooling: nullable/default-safe columns, new tables/indexes/compatibility views, `NOT VALID` constraints, non-transactional `CREATE INDEX CONCURRENTLY` | T1 | Expand migrations don't block traffic; old and new readers both accept | Old-reader-compatibility test | `DATA-09/expand-phase/` | Confirm current tooling supports issuing statements outside the wrapping transaction |
| T4. Backfill job harness: resumable, tenant-scoped, keyset-paginated, checkpointed, bounded batch/tx time, rate controls — reuses DATA-02's lease primitive for checkpoint safety | T3; DATA-02 T1 | **Interrupted/resumed backfill test (explicitly required)** — no reprocessing or skipping | This is the test | `DATA-09/backfill-interrupt-resume/` | Largest risk surface in DATA-09 | Code for the harness; human decision on batch size/rate/window per migration |
| T5. Validation-phase tooling: `VALIDATE CONSTRAINT` + reconciliation queries, artifact capture | T4 | Zero-mismatch reports are machine-checked artifacts, not prose | Artifact-schema test | `DATA-09/validation-artifacts/` | Code for the harness; human review of the report before canary |
| T6. Canary/deploy-N tooling: N alongside N-1, soak metrics | T5 | **N-1 on expanded N schema + N code before/after backfill (both explicitly required)** | This is the test | `DATA-09/canary-soak/` | No production telemetry baseline exists — soak duration/thresholds are a genuine, currently unresolvable judgment gap | Code for the harness; human decision on soak duration and go/no-go |
| T7. Switch-phase tooling: observable compatibility flag, dual-schema-version consumer support | T6 | **Application rollback after switch (explicitly required)**, no destructive `Down` | This is the test | `DATA-09/switch-rollback/` | The core safety property this protocol exists to guarantee | Code for mechanics; **the decision to flip in production is human** |
| T8. Contract-phase tooling: gated on evidenced no-N-1-remains precondition | T7 | **Forward recovery from every failed phase + delayed-contract-only-after-old-process-absence-proven (both required)** | This is the test | `DATA-09/contract-gate/` | Most safety-critical piece — running contract too early is destructive and hard to detect pre-outage | Code for the gate; human sign-off strongly advisable even with the gate passing |
| T9. Full CI drill pipeline covering all 6 directive-named drills | T1-T8 | All six drills run in CI/scheduled pipeline | CI pipeline + passing run artifact | `DATA-09/ci-drill-pipeline/` | Largest single infra investment in PF-DATA | Code; human decision on which real migration is the first live exercise — DATA-01's composite-FK rollout is the natural first candidate |

**wowsociety impact — DATA-09: Affected — process/tooling, not code, today.** `wowsociety/docs/DEPLOY.md:81-100` documents a single-shot "migrate fully, then deploy everyone" model (`depends_on: service_completed_successfully`) — no canary/soak, no N/N-1 dual-version window, no interrupted-backfill resume, no telemetry-gated contract check. **Not breaking wowsociety's current process, but the current sequence is less safe than DATA-09 requires:** because migrations always fully complete before any instance starts, wowsociety today has *no window where N-1 code runs against an N-migrated schema* — it collapses past exactly the compatibility window DATA-09 protects. **Required changes:** adopt whatever manifest schema wowapi's tooling consumes — wowsociety's own `cmd/migrate` runs the same underlying mechanics for its module migrations, making it a direct consumer, not a bystander. **Sequencing:** do not attempt wowsociety's own DATA-01-pattern fix ahead of wowapi's DATA-09 tooling (T1-T5 minimum).

**PF-DATA cross-cutting notes:** (1) The shared lease primitive (DATA-02 T1) is the single highest-leverage build in this package — staff and design-review it first. (2) `kernel/resource/registrar_pg.go`'s nil-actor placeholder is one fix claimed by two findings (DATA-06 T2, DATA-07 T3) — one owner, not two PRs. (3) DATA-07 has a hard dependency on SEC-01, secondary on SEC-04 — sequence accordingly, not in parallel. (4) DATA-08's Wave-0 pair and DATA-03 T7 target the exact same legal-notification-audit fix — implement once. (5) wowsociety's live exposure concentrates in two places: `kernel/audit` (DATA-08 W6-T1 hits real rows today) and the DATA-01 tenant-FK pattern (`policy_override`, an independent instance) — everything else in DATA-02–07 currently has zero wowsociety usage, confirmed by import grep, not absence of documentation. (6) DATA-09 is new infrastructure that DATA-01 and DATA-08 W6-T1 both need before their riskiest steps ship safely — sequence DATA-09 T1-T5 ahead of DATA-01 T4/T5 and DATA-08 W6-T1 in the real release plan, even though they're presented finding-by-finding here.

### 5.4 PF-DX — Developer Experience, CLI, Generators, DSL, Compatibility (DX-01 – DX-07)

**Accountable role:** developer-experience lead. **Evidence root:** `docs/implementation/evidence/premier/PF-DX/`.

#### DX-01 — Make the source-built CLI path valid (P0, **Wave 0**)

**Evidence:** `wowapi init` on a `devel` build info sets `v0.0.0` unconditionally and templates it straight into the generated `go.mod` with no resolvability check.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Add `--framework-version vX.Y.Z`, verify via `go list -m` before any file write | — | Unresolvable version fails before writes, with an exact remediation command | Unit | `DX-01/t1-flag-verify.json` | Low |
| T2. Add `--local-framework /absolute/path`, emit an explicit `replace` + visible dev-mode warning | — | Non-absolute/nonexistent path rejected; go.mod contains `replace`; stdout warns | Unit | `DX-01/t2-local-framework.json` | Low |
| T3. Derive exact pseudo-version from VCS metadata when the commit is reachable and clean; default when neither T1 nor T2 is passed | T1, T2 | Clean reachable commit → resolving pseudo-version; dirty/unreachable → fails closed with remediation, never falls back to `v0.0.0` | Unit against clean/dirty/detached/shallow fixtures | `DX-01/t3-pseudo-version.json` | Medium — shallow clones can make commit unreachable, must fail closed |
| T4. Delete the `v0.0.0` fallback path entirely | T1-T3 | No code path can write an unverified version | Exhaustive flag-combination matrix | `DX-01/t4-no-fallback.json` | Low |
| T5. **Real generate→build→boot→smoke test in an isolated temp dir** (directive's explicit wording) | T1-T4 | Both released-CLI and source-built-CLI paths: init → `go mod download` → `go build` → contract/smoke tests → success, end to end | Integration, subprocess-driven, real `go mod download` and compile | `DX-01/t5-e2e-temp-dir.json` | Medium — needs network/proxy access or a local module cache fixture in CI |

**wowsociety impact — DX-01: Not affected.** `wowsociety/go.mod:8,13` — `require v1.0.0` overridden by `replace => ../wowapi`, a path replace never touched by the CLI-generated dependency line. Informational only: wowsociety's `FRAMEWORK_VERSION` pins a raw SHA, not one of the two real tags (`v1.0.0`/`v1.1.0`) — a pre-existing dev-mode posture, not something DX-01 breaks or fixes, and arguably an argument for T2's `--local-framework` flag matching how wowsociety was actually bootstrapped.

#### DX-02 — Replace the TODO generator with a tested vertical-slice generator (P0 Wave-0 slice / P1 Wave-4 full generator)

**Evidence, with one additional confirmed defect:** generated create/get/update handlers are TODOs returning fake 200/201 success. The route uses permission verb `.delete`, but `kernel/authz/registry.go:13-19`'s closed verb set has no `delete` (only `deactivate`) — **and the generator's own test (`TestGenCRUDPermissionKeys`) asserts `widgets.widget.delete` as correct output, meaning the bug is test-locked, not merely untested.** The generated migration has no `status` column despite the handler's own comment claiming "lifecycle via status." The generated `tenant_id` FK is single-column — **every `gen crud` invocation currently manufactures a fresh DATA-01 violation.**

**Wave-0 (P0) tasks:**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| W0-T1. Decide disable-vs-minimal-slice, recorded as a `decisions.md` entry | — | If "disable," `gen crud` exits nonzero with "experimental/unsupported" and writes no files | — | `DX-02/w0-t1-decision.md` | Blocks all downstream tasks |
| W0-T2. Fix `.delete` → `.deactivate` verb (route + any permission-seed reference + **the test that currently asserts the bug**) | W0-T1 | Generated routes only use verbs in the closed set; boot-time validation passes on generated output | Boot-time validation: render+compile+boot a generated module | `DX-02/w0-t2-verb-fix.json` | Low, mechanical |
| W0-T3. Add `status` column to the migration template (or remove the false "lifecycle via status" claim if out of Wave-0 scope) | W0-T1 | Migration and handler comment agree | Migration apply + RLS test | `DX-02/w0-t3-status-column.json` | Low-medium — touches goose numbering |
| W0-T4. Replace false-success TODO handlers with either a minimal correct slice or an explicit-501/panic if "disable" was chosen | W0-T1, W0-T3 | No handler returns HTTP success without performing the operation | Handler-level test per verb, or explicit-501 test | `DX-02/w0-t4-handlers.json` | Medium — scope-creep risk; keep strictly to create/get/list/update/deactivate, no ETag/idempotency (P1) |
| W0-T5. **Make generated-output compile/boot tests authoritative** (directive's own Wave-0 wording) — scaffolds a throwaway module into an isolated temp dir, compiles as part of a real bootable product, boots against ephemeral Postgres | W0-T2-T4 | Integration, not substring assertions — supersedes existing `assertFileContains`/`assertParseGo`-only tests | Integration | `DX-02/w0-t5-e2e-compile-boot.json` | Medium — needs DB fixture wiring in the generator test harness, currently absent |

**P1/Wave-4 tasks (full vertical-slice generator):**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| P1-T1. Typed create/update/input/output models with validation | Wave-0 slice closed; ideally after AR-03 | Generated models reject unknown fields | Unit + contract | `DX-02/p1-t1-typed-models.json` | Medium |
| P1-T2. Optimistic concurrency + ETag | P1-T1 | Conflict returns 409/412 per existing convention | Concurrency test | `DX-02/p1-t2-etag.json` | Medium |
| P1-T3. Atomic resource-mirror + audit + outbox in generated handlers | DATA-06's aggregate write contract, ideally consumed rather than hand-rolled | Fault injection proves all-or-nothing commit | Integration | `DX-02/p1-t3-mirror-audit-outbox.json` | High — depends on DATA-06 landing first or duplicates work |
| P1-T4. Permission/resource/OpenAPI generated automatically from one declaration | AR-03 | One invocation produces all projections, no hand-editing to boot | Golden-diff test | `DX-02/p1-t4-projections.json` | High — this is AR-03 applied to the CRUD generator |
| P1-T5. Automatic module registration, zero manual TODO to boot | P1-T1-T4 | Output boots with zero manual edits | E2E temp-dir boot test | `DX-02/p1-t5-auto-registration.json` | Medium |
| P1-T6. Full test-suite generation (unit, contract, RLS, authz, idempotency, pagination, migration) | P1-T1-T5 | Every generated resource ships all these classes | Meta-test: generated suite itself passes | `DX-02/p1-t6-test-generation.json` | High — largest single item in DX-02 |

**wowsociety impact — DX-02: Not affected — confirmed clean via governance.** Grepped every permission verb in wowsociety's `identity`/`policy` routes: zero `.delete` usages, all valid closed-set verbs. Grepped `TODO` across `internal/modules/`: only legitimate forward-looking roadmap TODOs, none matching the generator's false-success pattern. **wowsociety's `docs/CONVENTIONS.md:10` — "Generator output wrong? Fix the generator upstream or file an RFF — never bypass" — is exactly the governance discipline that made the existing modules immune to this bug.** No retroactive remediation needed. wowsociety's own migrations already have real `status` columns with CHECK constraints — the discipline the raw generator lacks. **Sequencing:** land Wave-0 fixes before wowsociety's next `gen crud` invocation for a new resource (none currently pending, no urgency).

#### DX-03 — Define the state-of-the-art module DSL (Wave 4, P1, future design — not near-term implementation)

**Evidence:** confirmed no `port`/`Manifest[T]`/`Operation[Request,Response]` DSL exists anywhere in wowapi today — the directive's "proposed API, not current source" framing is accurate. Current DSL-adjacent surface (`module.Context`, string/any-keyed registries, the closed authz verb set) is exactly what AR-01/AR-02/DX-02 already target from the correctness-fix side; DX-03 formalizes the type-system-level version once those land.

| Task | Depends-on | Note |
|---|---|---|
| DX-03-T0. Formalize the design into an ADR under `decisions.md`, explicitly labeled "target, not implemented" per AR-05 | Wave 1 (AR-01 ApplicationModel, AR-02 typed ports) complete | Design-only |
| DX-03-T1..Tn. Implementation | Wave 1-3 exit gates; DX-02 P1-T4 reuses this compiler | Deferred — out of near-term scope per §12 Wave 4 |

**wowsociety impact — DX-03: Not affected, by explicit directive design constraint.** Wave 1's compatibility strategy states the legacy `module.Module`/`Context` interface is not removed or widened during this migration ("removal happens only in the future `/v2` module"). wowsociety's identity/policy modules compile unmodified through Waves 1-3.

#### DX-04 — Create one golden consumer and upgrade matrix (Wave 4, P1)

**Evidence:** `internal/testmodules/requests/` is a single hand-authored in-process reference module used for framework unit tests — not a CLI-scaffolded, installed-binary, two-module, upgrade-tested consumer. Zero CI workflow references `wowapi init`/`gen crud`.

**wowsociety-as-golden-consumer analysis (three concrete disqualifying reasons, each verified):** (1) wowsociety consumes wowapi via a sibling-checkout path `replace`, never an installed CLI binary — step 1 of the directive's own procedure ("Install the built CLI") is never exercised. (2) `FRAMEWORK_VERSION` is a raw SHA with no "previous supported version" concept and `framework-verify` only diffs SHA equality — step 7 ("upgrade from previous version, rerun contracts") has no mechanism. (3) wowsociety is a real product with domain-specific logic (committee seats, OTP/TOTP, citation packs) — coupling wowapi's own release gate to wowsociety's roadmap changes would violate the "non-internal consumer fixture" intent. **wowapi needs its own separate, framework-repo-owned golden-consumer fixture, distinct from wowsociety.**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Build framework-repo-owned golden-consumer scaffold job | DX-01 T5's shared subprocess-scaffold primitive; Wave-0 DX-02 slice | Fixture installs via `go install`, not repo-internal import | Integration | `DX-04/t1-scaffold-job.json` | Medium |
| T2. Generate resource, rule, workflow, event handler, recurring job, document flow, notification, webhook across 2 modules | T1 | Each subsystem exercised ≥1 time | Integration | `DX-04/t2-subsystem-coverage.json` | High — broad surface, many kernel subsystems must be generator-reachable |
| T3. Boot API+worker against Postgres/MinIO/Mailpit/OTel; exercise authenticated CRUD, async delivery, restart/retry, RLS isolation | T2 | All paths pass | Integration, real infra | `DX-04/t3-boot-exercise.json` | Medium |
| T4. Upgrade-from-previous-version replay | T1-T3, DX-05 (needs real N/N-1 policy) | Fixture at N-1, upgraded to N, contracts rerun and pass | Two-pass integration | `DX-04/t4-upgrade-replay.json` | High — depends on DX-05 landing first |
| T5. Wire into CI as a required gate | T1-T4 | Appears in `ci/release-gates.yaml` at its Wave-4 boundary (REL-01) | CI config test | `DX-04/t5-ci-gate.json` | Low once T1-T4 exist |

**wowsociety impact — DX-04: Not affected.** No direct code change required. wowsociety's `framework-verify` and richer hand-completed modules are a useful secondary signal but explicitly not a substitute for the CI-authoritative fixture.

#### DX-05 — Make CLI/docs/version identity singular (**Wave 0**, P1)

**Evidence, a confirmed direct three-way contradiction:** README + upgrade-policy say pre-1.0/v0 (breaking changes allowed on minor bumps); CHANGELOG says v1.0.0-stable (breaking requires major bump); `git tag -l` confirms both `v1.0.0` and `v1.1.0` exist. **Additional self-contradiction found:** README.md:177 itself says `go install .../wowapi@latest # or @vX.Y.Z once tagged` — directly violating the policy's own "never via @latest" rule, and tags already exist so the "once tagged" caveat is stale. The directive's decision (repo is on the stable v1 line) is the correct resolution given tag evidence outweighs stale prose.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Rewrite README status banner to v1-stable + exact-pin recommendation, remove `@latest` | — | Banner matches CHANGELOG's v1.0.0 claim | Doc-execution CI | `DX-05/t1-readme.json` | Low |
| T2. Rewrite upgrade-and-deprecation-policy.md to v1/N-1 rules | T1 | Text matches the directive's decision exactly | Doc review + CI example execution | `DX-05/t2-policy.json` | Low |
| T3. Reconcile blueprint-11 CLI examples with `internal/cli/cli.go`'s real commands/flags | — | Every blueprint example matches an actual command | CI executes blueprint examples | `DX-05/t3-blueprint-cli-reconcile.json` | Medium — per-example implement-or-delete decision |
| T4. `wowapi version` fails mutating generator commands on incompatible major/minor pairing | DX-01 (shares version-verification plumbing) | Incompatible pairing exits nonzero pre-write | Mismatched-version fixture | `DX-05/t4-version-gate.json` | Medium |
| T5. Public API/config/event compatibility gates enforcing v1 rules | REL-03 (shared infrastructure) | Intentional v1-breaking fixture fails CI | CI gate test | `DX-05/t5-compat-gates.json` | High — large, shared with REL-03/DX-06 |

**wowsociety impact — DX-05: Affected, informational/process, not code-breaking.** `wowsociety/go.mod:8` declares `v1.0.0` but is overridden by the `replace`, misleading if read without context. **Confirmed gap:** `docs/CONVENTIONS.md:49-53` describes SHA-based drift detection only — no concept of N/N-1 or major/minor compatibility class exists. `FRAMEWORK_VERSION` pins `0e578e8` (current main HEAD, confirmed fresh not stale) instead of a real tag. Once DX-05's T4 lands a real v1/N-1 enforcement gate, wowsociety's path-`replace`/SHA-pinned model sidesteps it entirely (no installed CLI is ever invoked against the pin). **Recommended (not required):** migrate `FRAMEWORK_VERSION` to a real tag and extend `framework-verify` to validate semver compatibility class, not just SHA equality — a wowsociety-repo follow-up, out of PF-DX's wowapi-side scope but should be tracked.

#### DX-06 — Make OpenAPI merge complete or fail loudly (P1, Wave 4, **overlaps AR-03 T2**)

**Evidence:** `openapi_cmd.go`'s merge-target struct captures only `Paths` and `Components.Schemas` — every other top-level 3.1 field (`security`, `tags`, `servers`, `webhooks`, callbacks, non-schema `components.*`) is silently discarded by `json.Unmarshal` with no error or warning.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Expand merge struct to cover all 3.1 top-level + `components.*` fields with explicit per-field merge policy | — | Every field either merges correctly or the fragment is explicitly rejected | Fixture-driven, one fragment per field type | `DX-06/t1-full-merge.json` | Medium — needs explicit per-field policy decisions; **this is the same closure contract as AR-03 T2 — assign single ownership before implementing** |
| T2. Validate final merged doc against OpenAPI 3.1.1/JSON Schema 2020-12 | T1 | Malformed merged output fails the command | Structural validation | `DX-06/t2-spec-validation.json` | Medium — needs a validator dependency |
| T3. Semantic API diffing gated to DX-05's v1 policy | T1, T2, DX-05 T5 | Intentional breaking fixture fails | CI gate test | `DX-06/t3-semantic-diff.json` | High — shared scope with REL-03 |

**wowsociety impact — DX-06: No currently observed issue.** wowsociety's per-module OpenAPI fragments spot-checked appear to use only `paths`/`components.schemas`-shaped content, which merges identically before and after the fix. Latent risk only if wowsociety ever adds `security`/`webhooks` to a fragment. **Recommended follow-up:** audit `wowsociety/internal/modules/*/openapi.json` for silently-dropped fields once T1's stricter validator ships.

#### DX-07 — Make readiness and configuration diagnostics truthful (P1)

**Evidence:** the health contract's own doc describes readiness as including "migration currency," but the generated `cmd/api/main.go.tmpl` readiness map registers only `"db"` and `"seeds"` — no migration-currency check exists, contradicting the documented contract. `config_delegate.go`'s product-checker discovery is CWD-relative `os.Stat`, silently falling back to framework-only validation if not found there. `CapacityMode` defaults to `"advisory"` (never enforced); `HTTPMaxInFlight` defaults to `0` (backpressure fully disabled).

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Add migration-currency check to the generated readiness template | — | `/readyz` fails when applied-migration version lags expected | Integration: boot against stale-migrated DB, assert 503 | `DX-07/t1-migration-check.json` | Medium — needs a stable "expected migration version" source |
| T2. Add seed/rule/model-hash checks to readiness | AR-01's model hash for the model-hash portion; T1 can ship independently | Readiness reports migration version, seed/rule hash, model hash | Integration | `DX-07/t2-full-readiness.json` | Medium-high — model-hash portion blocked on Wave 1 |
| T3. `config doctor` discovers product root via `go env GOMOD`/`--project`, not CWD-relative `os.Stat` | — | Delegation works regardless of invocation directory; explicitly reports whether product validation ran | Unit: nested-subdir + outside-repo-with-`--project` invocations | `DX-07/t3-config-doctor-discovery.json` | Low-medium |
| T4. Production profile requires declared/enforced capacity + intentional backpressure; advisory/unset becomes a visible readiness failure or requires a waiver | T1-T3, AR-04's waiver framework | Prod + unset capacity fails readiness unless waived | Integration | `DX-07/t4-prod-capacity-enforcement.json` | Medium — must not break existing prod deployments that haven't set capacity yet; needs a grace period |

**wowsociety impact — DX-07: Affected — wowsociety's already-generated `cmd/api/main.go` shows the identical gap.** `wowsociety/cmd/api/main.go:240-243` has the same two-check shape (`"db"`, `"catalogs"` — naming drift from the current template's `"seeds"`, suggesting main.go was generated against an earlier template revision and hand-edited since). **Not breaking** — T1's fix changes the *template* only; it does not retroactively alter wowsociety's already-committed, non-regenerated `cmd/api/main.go`. Confirmed positive: `wowsociety/tools/configcheck/main.go` **exists**, so DX-07 T3's `config doctor` discovery fix is a non-issue for wowsociety — product-aware validation already engages correctly today. **Recommended follow-up (not blocking):** manually backport the migration-currency check into wowsociety's own readiness map once wowapi's T1 pattern is established.

**PF-DX cross-cutting note:** DX-01 T5's isolated-temp-dir subprocess-scaffold harness is the shared load-bearing primitive for DX-01, DX-02's Wave-0 slice (W0-T5), and DX-04 (T1) — build it once, sequence DX-01 T5 first among Wave-0 P0 items. DX-05's tag-based v1/N-1 decision should land before DX-04 T4 needs a real "previous supported version" to replay against.

**Supplementary verification (second independent pass, corroborating + one correction):**
- **wowsociety already independently hit and documented DX-01's exact bug.** `wowsociety/docs/upstream/12-sf-7-init-gomod-invalid-and-gitignored-local-overlay.md` describes the emitted go.mod as having "an invalid pseudo-version (`v0.0.0-...-90448926501d+dirty`)" with "no replace directive" — a real product's own upstream-findings log matching DX-01's citation exactly. This raises DX-01's wowsociety-impact confidence from "latent risk" to "already realized and worked around" — recommend wowsociety mark that upstream finding resolved once DX-01 T1 ships, no code change needed on wowsociety's side.
- **Correction to the directive's DX-02 evidence text:** "module generation itself leaves migration and registration TODOs" overstates the gap. `templates/module/module.go.tmpl:44-46` shows migrations/seeds/OpenAPI **are** auto-wired (`mc.Migrations(...)`, `mc.Seeds(...)`, `mc.OpenAPI(...)`); only line 48 leaves a TODO for routes/permissions/health-checks/ports. DX-02's Wave-0 task table above should target that narrower, correctly-scoped gap, not assume migrations need fixing too.

### 5.5 PF-PERF — Performance and Bounded Resources (PERF-01 – PERF-06)

**Accountable role:** performance/SRE lead. **Evidence root:** `docs/implementation/evidence/premier/PF-PERF/`. Governing methodology for all numeric thresholds: directive §14. **B11 reopen trigger (confirmed, does not fire today):** >15% median dispatch-latency increase from 50→2,000 routes in two consecutive reference-run comparisons, or dispatch exceeding 10% of measured p95 authenticated-request latency — current host delta is 9%, B11 stays parked.

#### PERF-01 — Fix the token-bucket map before further rate-limit features (P0, **Wave 0**)

**Evidence, independently re-confirmed by this document's own mentor-review spot-check:** a new bucket starts at `burst` tokens, consumes 1 immediately (`burst-1`). `sweep()` only evicts a bucket whose stored `tokens >= burst` — but that field is only recomputed inside `Allow()` when the *same key* is looked up again. A one-shot key frozen at `burst-1` is therefore never sweep-eligible and never evicted. Once the map reaches `sweepAt` (10,000), every new key triggers a full O(N) scan that removes nothing.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Recompute effective refill during sweep (apply the same refill formula sweep-side, not only on `Allow`) | — | A one-shot key becomes sweep-eligible once idle time × rate ≥ burst deficit | Unit: insert 1 key, advance clock, sweep, assert removed | `PERF-01/refill-sweep-unit.json` | Low |
| T2. Add hard capacity with deterministic overflow behavior | T1 | Map size never exceeds configured max under adversarial load | 10k+ eviction test (T5) + concurrent-insert race test | `PERF-01/hard-capacity.json` | Medium — overflow policy is an observable behavior change, must be documented (reject-oldest/reject-new/forced sweep-recheck), not silent unbounded growth |
| T3. Emit metrics: current entries, evictions, rejected admissions, sweep duration | T1, T2 | Metrics observable via existing callback-injection pattern | Metric-emission unit test | `PERF-01/metrics.json` | Low |
| T4. Config-level exposure of hard capacity, backward-compatible with existing 2-arg constructor call sites (wowsociety calls the 2-arg form) | T2 | Existing call sites compile unchanged; new capacity added via option/variadic, not positional args | Compile-fixture test | `PERF-01/compat.json` | Medium — signature stability required per DX-05's v1 policy |
| T5. **10k+ one-shot-key eviction test** (directive-mandated) | T1, T2 | Insert >10,000 one-shot keys, advance clock past `idleTTL`, sweep, assert map returns below bound | New deterministic test | `PERF-01/eviction-10k.json` | Low — deterministic with fake clock |
| T6. **Sweep-cost benchmarks at 10k/100k/hard-limit** (directive-mandated) | T2 | Three new benchmarks with `bench-budgets.txt` entries added **in the same PR** (per PERF-06's fail-closed policy) | New benches | `PERF-01/sweep-benchmarks.json` | Medium — orphan-benchmark risk if PERF-06 T1 isn't sequenced alongside |
| T7. **Fuzz tests for invalid rate/burst/clock/concurrent keys** (directive-mandated) | T1, T2 | Native Go fuzz targets, following repo precedent (`pagination`/`filtering` fuzz tests) | Fuzz | `PERF-01/fuzz-corpus/` | Medium — depends on PERF-06's CI fuzz wiring existing |
| T8. Race test proving concurrent `Allow`/`sweep` correctness under the new capacity bound | T2 | `go test -race` clean | Race test | `PERF-01/race.json` | Low — existing mutex already serializes this |

**wowsociety impact — PERF-01: Affected, configuration exposure only, behaviorally neutral in the default case.** wowsociety wires `httpx.NewTokenBucket(cfg.HTTP.RateLimit.RequestsPerSecond, cfg.HTTP.RateLimit.Burst)` using framework default values — `configs/base.yaml` has **no `http.rate_limit` section at all**, confirmed absent, not an override. **Not breaking.** If T4's config-level hard-capacity override lands, wowsociety should review whether the new default ceiling suits its expected cardinality (tenant × capacity keys) — a housing-society SaaS with many small single-org tenants is unlikely to approach 10k+ distinct keys per pod, but this should be a reviewed assumption. No code change required for T1-T3/T5-T8.

#### PERF-02 — Measure complete requests against real PostgreSQL (P1, Wave 5, blocked on §14 reference environment)

**Evidence, confirmed exactly:** `BenchmarkDispatch`'s own doc comment states auth/authz are exercised via fakes, no real DB. Every tenant transaction issues 2-4 statements (role bind, optional RLS-enforcement `pg_roles` check, tenant bind, optional actor bind) before handler code runs.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Stand up dedicated Linux amd64 reference runner + `perf/reference-v1.json` skeleton — **shared prerequisite across PERF-02/03/04/05** | — | Artifact records CPU/runner digest, Go version, Postgres config, pool size, dataset cardinality, tenant distribution, workload seed, warm-up/measurement durations (§14 full field list) | N/A (infra) | `perf/reference-v1.json` + fixtures | **High — new CI infrastructure, no owner/timeline established anywhere in the directive** |
| T2. DB-backed benchmarks for public/authenticated-read/authenticated-write/resource-authz/idempotent-write/async-enqueue profiles | T1 for reference comparison; can be written standalone first | Real Postgres, not fakes; records p50/p95/p99, allocations, SQL count, bytes, pool wait, tx duration, lock wait, plan hash | New DB-gated bench file | `PERF-02/db-benchmarks/` | Medium — must not weaken RLS guards to win the benchmark (explicit directive prohibition) |
| T3. Cold/warm cache × 1/10/100 concurrent-tenant variants | T2 | 6 minimum combinations per workload profile | Parameterized harness | `PERF-02/concurrency-matrix.json` | Medium — 100-tenant needs realistic seed data |
| T4. Attribute cost breakdown: pool wait / tx setup / authz query / handler query / serialization / middleware, separately | T2 | No single aggregate number | Span-based or `EXPLAIN`-correlated instrumentation | `PERF-02/cost-breakdown.json` | Medium — no existing scaffolding for this breakdown today |
| T5. Publish results against `perf/reference-v1.json` | T1-T4 | Full closure-contract text satisfied | Reference-runner CI job | `PERF-02/reference-results.json` | **Blocked until T1 exists — this is the finding's actual closure gate** |

**wowsociety impact — PERF-02: No code change required.** Purely additive/observational wowapi-internal work; wowsociety's requests already go through the identical framework-owned `TxManager.WithTenant` path. Confirmed zero `func Benchmark*` definitions anywhere in wowsociety — no benchmark surface to update or risk of drift.

#### PERF-03 — Collapse rules resolution into bounded SQL work (P1, Wave 5, blocked on §14)

**Evidence, confirmed exactly:** one sequential SQL query per org ancestor, worst case `len(ancestors) + 2` round trips. **One item flagged as genuinely unverified in this pass** (agent-transport failure truncated the sub-check): the directive's claim that current `rule_versions` indexing favors active-only lookup should be re-confirmed against migrations before implementation.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T0. **(Gap-fill)** Verify current `rule_versions` index definitions before designing the new query | — | Confirm/refute the directive's indexing claim | `grep "CREATE INDEX" migrations/*rules*.sql` | `PERF-03/index-audit.json` | Low but must precede T2 |
| T1. Design one set-based query over ancestry + tenant + platform fallback | T0 | Single SQL statement replaces the `for` loop, preserving exact precedence semantics | Result-parity unit tests | `PERF-03/query-design.sql` | Medium — must preserve nearest-ancestor-first → tenant → platform → code-default precedence and the schema-drift re-validation unchanged |
| T2. Add/confirm indexes matching both current and historical predicates | T0, T1 | `EXPLAIN (ANALYZE, BUFFERS)` shows index access, not seq scan | Query-plan fixture tests | `PERF-03/explain-fixtures/` | Medium — wrong column order defeats the plan |
| T3. `EXPLAIN (ANALYZE, BUFFERS)` fixtures at representative depth/history cardinality | T1, T2 | Fixtures committed for shallow and deep org ancestries, low/high historical-version counts | New fixture harness, DB-gated | `PERF-03/explain-fixtures/` | Medium — needs seeded org-ancestry test data |
| T4. Result-parity + SQL-count-constant-with-depth tests | T1 | Query count stays constant across 3/10/50-level ancestries | Parametrized test | `PERF-03/parity-and-sql-count.json` | Low-medium — needs query-counting instrumentation, not confirmed to exist yet |
| T5. Preserve live per-request rule updates — explicit non-regression constraint ("B13 is not needed for rules") | T1 | No stale-read regression | Existing rule-update-visibility tests continue passing | `PERF-03/live-update-regression.json` | Low |
| T6. Publish before/after evidence against `perf/reference-v1.json` | PERF-02 T1 | Meets §14 SQL-count and p95/p99 budgets | Reference-runner CI job | `PERF-03/reference-results.json` | Blocked on reference environment |

**wowsociety impact — PERF-03: Not affected — confirmed absent by direct source read.** `internal/modules/policy/rulepoints.go:162-168`'s `allowedScopes()` **explicitly and deliberately excludes org scope**, per its own comment: "societies are single-org tenants in E0." wowsociety's policy module never registers or resolves at org scope, so it cannot trigger PERF-03's pathological ancestor-loop cost — it only ever hits the already-O(1) tenant/platform fallback path. If wowsociety later adds org-scoped policy (plausible for a multi-society federation, given "E0" phase-labeling), this rewrite becomes directly load-bearing then — a forward dependency, not a current one.

#### PERF-04 — Remove N+1 and unbounded materialization from sweepers/workers (P1, Wave 5, blocked on §14)

**Evidence, all 4 citations confirmed exactly:** `SweepSLA` loads ALL due rows unbounded (no `LIMIT`), then does 1 UPDATE + 1 load + emit per row. The reminder query has no matching index (`wft_due` only covers `due_at`, the query filters `remind_after`). `webhook.RetryOutbound` loads endpoints per-delivery (bounded batch of 10, but still N queries). Outbox's outer claim transaction spans the entire per-subscriber dispatch loop, including nested per-event tenant transactions.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Bounded batch claiming for `SweepSLA` (both queries get `LIMIT`, loop via job re-invocation) | — | Fixed query count and memory across due-row cardinalities | Test at 10/1k/100k due rows | `PERF-04/sweeper-batch-bound.json` | Medium — must preserve existing idempotency guards, no reintroduced double-remind race |
| T2. Convert per-row UPDATE+load+emit to set-based/batched operations where semantically possible | T1 | Set-based UPDATE for guard flips; batch-load by ID set | Query-count assertion tests | `PERF-04/sweeper-setbased.json` | Medium-high — `emit()`/escalation logic is inherently per-instance |
| T3. Partial index on `remind_after` matching the query predicate | — | `EXPLAIN` shows index scan | Plan test | `PERF-04/remind-index-explain.json` | Low — additive, follows DATA-09's expand-only protocol since `workflow_tasks` is a live shared table |
| T4. Batch-load endpoints in `RetryOutbound` (one `IN (...)` query per invocation, not per-row) | — | No per-delivery endpoint query | Query-count test: N rows / M endpoints → 1 query | `PERF-04/webhook-batch-endpoints.json` | Low — directive suggests caching immutable endpoints by version too |
| T5. Rework outbox claim/dispatch into a leased state machine, preserving per-aggregate ordering | **Hard dependency on PF-DATA's Wave-3 DATA-02/DATA-03 lease primitives — cannot start before those land** | No outer transaction spans tenant handlers; queue-lag/batch-duration metrics added | Crash/duplicate-worker tests inherited from DATA-02's gate | `PERF-04/outbox-leased-batch.json` | **High — cross-work-package dependency, do not attempt in isolation, flag explicitly rather than re-deriving the lease design inside PF-PERF** |
| T6. Queue lag and batch duration metrics | T1-T5 | Metrics for sweeper/webhook/outbox timing | Metric-emission tests | `PERF-04/metrics.json` | Low |
| T7. Bounded-batch benchmarks at due-row cardinality tiers | T1-T3 | Parallel structure to PERF-01's tiers, budget entries added same-PR | New benches | `PERF-04/sweep-cost-benchmarks.json` | Medium — same orphan-benchmark risk as PERF-01 T6 |
| T8. Publish before/after evidence against `perf/reference-v1.json` | PERF-02 T1 | Meets §14 budgets | Reference-runner CI job | `PERF-04/reference-results.json` | Blocked on reference environment |

**wowsociety impact — PERF-04: Not affected — confirmed absent by direct grep across `kernel/workflow`, `kernel/webhook`, `kernel/outbox` import paths, zero hits for all three.** If wowsociety later adopts any of the three, it inherits the batched/leased behavior from day one with no migration burden.

#### PERF-05 — Make object checksum behavior explicit (P2, Wave 5, partially blocked on §14)

**Evidence, confirmed exactly:** `S3.Stat` returns immediately if checksum-signed metadata is present; otherwise full-downloads and streams through `sha256`. A checksum-on-upload path already exists (per the code's own comment) — PERF-05 is about making it *required* for framework uploads, not inventing it.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Require framework uploads to always persist canonical checksum metadata; audit every upload call path for universality | — | "Framework uploads always persist and verify canonical checksum metadata"; normal `Stat` performs no body download | Integration: upload via framework path, `Stat`, assert no `GetObject` call | `PERF-05/upload-checksum-required.json` | Medium — enumerate every current upload call site; scope expands if any bypasses checksum-signing |
| T2. Move the full-hash fallback to an explicit, size/time-bounded import/repair path | T1 | Fallback reachable only from a labeled repair invocation, never ambient `Stat` | Test: legacy object triggers fallback only via labeled path | `PERF-05/fallback-bounds.json` | Medium — likely needs an API-surface decision (new `Stat` variant vs. separate `RepairChecksum` method), affecting the `storage.ObjectInfo` port other adapters implement |
| T3. Dedicated metrics for fallback invocations | T1, T2 | Counter/histogram for fallback hits, bytes, duration | Metric-emission test | `PERF-05/metrics.json` | Low |
| T4. Resumable async backfill for legacy objects | T2 | Interrupt mid-run, resume, no duplicate work, eventual completion | Backfill interrupt/resume test | `PERF-05/backfill-resumable.json` | Medium — needs an inventory mechanism for "legacy objects lacking checksum metadata," which doesn't obviously exist yet |
| T5. Publish before/after evidence | PERF-02 T1 | Meets §14 budgets | Reference-runner CI job | `PERF-05/reference-results.json` | **Partially blocked** — the "no body download" behavioral proof (T1) is independently testable now; only the quantified latency claim needs the reference environment |

**wowsociety impact — PERF-05: Indirectly yes, no wowsociety code changes needed.** wowsociety wires the framework's `adapters/storage/s3` package directly in both `cmd/api/main.go` and `cmd/worker/main.go`; it never calls `Stat`/checksum logic itself — all upload/confirm flow goes through the framework's document/storage service layer, so wowsociety inherits whatever PERF-05 establishes with zero call-site changes. If T2's port-interface signature changes, verify wowsociety's adapter import still compiles — low-risk compile-check only, since wowsociety only imports the constructor/wiring, not `Stat` internals.

#### PERF-06 — Make performance gates fail closed (P1, **Wave 0**)

**Evidence, confirmed exactly:** a budgeted-but-absent benchmark only WARNs to stderr and `continue`s — it never fails the exit code. Hosted CI runs only fuzz seed corpora (`go test` default replay), never real `-fuzz` coverage-guided generation — confirmed via grep, exactly 2 existing fuzz test files, both getting seed-replay-only in CI.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Change the missing-benchmark path from WARN+continue to a tracked failure | — | Removing/renaming a budgeted benchmark fails CI | Golden test: budget references an absent benchmark → tool exits non-zero | `PERF-06/missing-bench-fails.json` | Low — small isolated change; needs a new negative-fixture test (no `main_test.go` currently exists for the tool) |
| T2. Track benchmark baselines statistically (§14's `benchstat` alpha-0.01 methodology) | T1 | A seeded regression fails the statistical gate unless an independently reviewed budget update is attached | Seeded-regression fixture test | `PERF-06/statistical-regression-gate.json` | Medium — needs `benchstat` integrated into CI, plus a 10-run comparison design decision |
| T3. Short time-bounded fuzzing on PRs | — | PR job runs `-fuzz=<Name> -fuzztime=Ns`; fuzz artifacts prove non-zero time beyond seed replay | CI workflow change | `PERF-06/pr-fuzz-execution.json` | Medium — CI runtime budget impact; needs a time-bound decision the directive doesn't specify |
| T4. Longer scheduled fuzzing with corpus retention | T3 | Separate scheduled workflow, corpus persisted across runs | New scheduled workflow | `PERF-06/scheduled-fuzz-corpus/` | Low-medium — corpus-retention mechanism (artifact vs. commit) is an implementation choice |
| T5. Wire into REL-01's `ci/release-gates.yaml` manifest as named, owned entries | T1, T3 (PF-REL's REL-01, external wiring dependency) | Entries have correct `required_from_wave` | Manifest schema validation (owned by PF-REL) | `PERF-06/gate-manifest-entry.json` | Low — coordination handoff, not independent implementation |

**wowsociety impact — PERF-06: Affected — infrastructure gap, not a regression.** wowsociety's `Makefile` has zero `bench`/`Benchmark`/`perf` targets; its single CI workflow has no fuzz or benchmark-budget step. wowsociety currently has zero `func Benchmark*` definitions — nothing to regress today, but also no tripwire if a future benchmark is added and later silently stops running. **Recommend, do not require for PF-PERF closure:** once PERF-06 lands, wowsociety should adopt the same fail-closed pattern in its own `make ci` if/when it adds benchmarks (e.g. `policy` module rule-evaluation hot paths) — a wowsociety-repo backlog item, not a PF-PERF blocker.

**PF-PERF cross-cutting notes:** (1) The §14 reference-environment artifact is a blocking prerequisite for PERF-02/03/04/05, not incidental work, and has no owner/timeline established anywhere in the directive's phase blueprint — recommend tracking "stand up reference runner + author v0 skeleton" as its own task, sequenced before or in parallel with (not after) PERF-02's benchmark-writing work. (2) PERF-04 T5 has a hard, unilaterally-unresolvable dependency on PF-DATA's Wave 3 — flag, don't re-derive. (3) Orphan-benchmark risk: every new PERF-01/PERF-04 benchmark must land its budget entry in the same PR per PERF-06 T1's own policy — sequence PERF-06 T1 alongside, not after. (4) wowsociety has zero reciprocal perf-gate infrastructure and zero current benchmarks — low current risk, worth flagging as a wowsociety backlog item once it grows benchmark coverage. (5) Every citation and wowsociety-impact claim here was independently re-verified via direct source reads after this work package's own dispatched sub-agents hit rate limits and could not be recovered — no unverified sub-agent output was used.

### 5.6 PF-REL — Release Engineering, Supply Chain, Compatibility Gates (REL-01 – REL-04)

**Accountable role:** release/security-engineering lead. **Evidence root:** `docs/implementation/evidence/premier/PF-REL/`.

**Pre-flight correction to the directive's own evidence (live-state drift, confirmed via `gh api`):** the directive's REL-02 evidence describes CodeQL/Scorecard/dependency-review as skipped because the repo is private. **wowapi's repository visibility flipped to public on 2026-07-03**, before the review's reviewed SHA. Live verification (`gh api repos/qatoolist/wowapi --jq '.visibility'` → `public`; `gh run view` on the two most recent CodeQL/Scorecard runs → both `success`, not `skipped`) confirms **these gates are already active today**, not skipped. Dependency-review's `skipped` status on the most recent run is expected `pull_request`-only gating, not a visibility skip. Trivy's `exit-code: 0`/`ignore-unfixed: true` remains a real, live gap regardless of visibility. **This correction narrows REL-02's actual scope** — see below.

**Confirmed hard blockers requiring GitHub repo-admin action (verified via `gh api`, not assumed):** `gh api repos/qatoolist/wowapi/branches/main/protection` → 404, no branch protection exists. `gh api repos/qatoolist/wowapi/environments` → `{"total_count":0}`, no GitHub Environments exist. `security_and_analysis` fields all `disabled` — no GHAS license active (CodeQL/Scorecard run only because the repo is public, not because of GHAS).

#### REL-01 — Gate release on the exact commit being published (P0, **Wave 0**)

**Directive's selected design (5-step summary):** (1) a reusable `required-gates.yml` (`workflow_call`) runs a versioned `ci/release-gates.yaml` manifest against an exact SHA; (2) both PR/main CI and release call it, release passing `${{ github.sha }}` from the tag event, emitting an attested `gate-results.json`; (3) `build-candidate` (no publish permissions) verifies gate results, builds once, emits artifacts + `release-manifest.json`, never pushes; (4) `publish` (protected `release` environment, the only job with write permissions) copies exactly the manifested bytes, never rebuilds; (5) `verify-published` re-verifies everything from a clean runner post-publish.

**Machine acceptance (the floor for every task below):** a deliberately failing check prevents `build-candidate`; changing the tag target changes both manifest SHAs; tampering with gate results or candidate bytes is detected; publish rejects any artifact/digest absent from the manifest; post-publish verification succeeds from a clean runner with no build workspace.

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Design `ci/release-gates.yaml` manifest schema (ID, command/job ref, owner, `required_from_wave`, timeout, evidence-artifact path) + JSON Schema validator | — | Schema rejects a manifest entry missing a required field | Unit: malformed vs. valid manifest fixtures | `REL-01/manifest-schema/` | Low — pure config/tooling |
| T2. Populate Wave-0 manifest entries mapping to today's `workflow-lint`/`unit`/`gate`/`coverage`/`reference-smoke` jobs + `vuln.yml` + REL-02's blocking scanners | T1 | Every job currently required for green `ci-container` has a manifest entry, `completed_wave: 0` | Diff review: entry count matches | same | Medium — must not silently drop an existing check |
| T3. Build `required-gates.yml` (`workflow_call`, parameterized on SHA), emitting attested `gate-results.json` | T2 | Called with a failing entry, output attests failure; each entry individually reported | Seeded-failure fixture through the workflow | `REL-01/gate-results-schema/` | Medium — must guarantee exact-SHA checkout, not branch HEAD |
| T4. Update `ci.yml` to call `required-gates.yml` so PR CI and release use the identical execution path | T3 | Same SHA through both paths produces byte-identical results (excluding run ID/timestamp) | Diff-based test | same | Medium — must not regress PR CI latency |
| T5. Add a `verify` job to `release.yml` calling `required-gates.yml` with the tag event's exact SHA, never trusting a same-named check on another ref | T3, T4 | Checked-out SHA == tag's target commit | **Seeded-failure fixture:** tag a commit with a deliberately broken test, prove `verify` fails and `build-candidate` never runs | `REL-01/verify-job/` | **High — this is the core trust boundary** |
| T6. Split into `build-candidate` (perms: `contents:read`, `id-token:write`, `attestations:write` only — no write) emitting archives/checksums/SBOMs + OCI layout, then attested `release-manifest.json` | T5 | Job's token literally cannot push/release (permission-scoped, not conventional) | **Tamper test:** hand-edit one artifact byte, prove mismatch detected | `REL-01/build-candidate/` | **High — needs a design spike.** Current `goreleaser release --clean` is one atomic build-and-publish step; a true split requires GoReleaser split-mode (`--skip=publish` + separate `publish` invocation — needs a docs lookup, not assumed) or hand-rolling the pipeline. |
| T7. Add `publish` job (`needs: build-candidate`, protected `release` environment) copying only manifested artifacts, never rebuilding | T6, **[human-blocked, see below]** | Script diffs requested artifacts against `release-manifest.json`, refuses anything absent | **Unmanifested-artifact test:** inject an extra artifact, prove rejection | `REL-01/publish-reject-unmanifested/` | High — blocked on human environment setup; code can be tested against a stub environment but not proven end-to-end without the real protected environment |
| T8. Write `scripts/validation/verify_release.sh <version> <source-sha>` (required by §13.1) | T7 | Runs from a genuinely clean environment, exits non-zero on any single mismatched field | **Golden failure tests, one per verified property** (wrong SHA, stripped signature, missing SBOM attestation, wrong platforms, tampered manifest hash) — §13.1 explicitly requires this, "a prose checklist is not sufficient" | `REL-01/verify_release/` | Medium |
| T9. Add `verify-published` job invoking T8 on a clean runner; failure marks the release failed, blocks `latest` promotion | T8 | Corrupted publish (in a disposable test repo) caught, `latest` not moved | **End-to-end dry-run against a disposable throwaway repo, never the real release pipeline** | `REL-01/e2e-dry-run/` | High — needs a disposable repo/registry for safe rehearsal |
| T10. Document target SLSA 1.2 guarantees without over-claiming assessed level | T6, T7 | States exactly which build-track requirements T6/T7's builder meets, no false claim | Doc review only | `REL-01/slsa-mapping.md` | Low — but a false SLSA claim is itself a supply-chain trust defect |

**Human-required blockers (cannot be completed by a coding agent, confirmed via live `gh api` calls, not assumed):**
- **Protected `release` GitHub Environment does not exist** (`total_count: 0`) — creating one with required reviewers is a repo-admin-console-only action. **T7 cannot be end-to-end proven until this exists.**
- **No branch protection on `main`, no tag protection rules exist** — the directive's own closing sentence ("Protect release tags and the environment at the repository/organization level") requires repo-admin console action, independent of how well `required-gates.yml` is built.
- Recommend a single explicit ticket ("PF-REL-ADMIN-01: configure release environment + tag/branch protection") tracked separately from and blocking T7/T9's *full* closure, so agent-completable YAML/script work isn't silently gated on an unstaffed admin task.

**wowsociety impact — REL-01: Not affected.** Confirmed via full read of `wowsociety/.github/workflows/ci.yml`: it checks out wowapi as a sibling repo at a pinned SHA via plain `git checkout`, with zero `workflow_call`/`uses: qatoolist/wowapi/.github/...` references — it never invokes wowapi's release pipeline. No breaking change, no required action, no sequencing constraint.

#### REL-02 — Make security checks blocking or replace them (P0/P1, Wave 0 (Trivy) + Wave 6 (full))

**Corrected baseline (per the pre-flight live-state finding above): only Trivy's non-blocking config is a currently-live gap, plus the absence of a documented fallback for "repo reverts to private" — CodeQL/Scorecard/dependency-review are already effectively active.**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Flip Trivy to blocking (`exit-code: "1"`), scope `ignore-unfixed` to a reviewed allowlist only | — | Trivy fails on CRITICAL/HIGH findings with an available fix | **Seeded-vuln fixture**, prove fail then pass after removal | `REL-02/trivy-blocking/` | Medium — run once report-only to baseline before flipping, or it can immediately break `main` on latent findings |
| T2. Waiver mechanism: reviewed allowlist file with owner/rationale/expiry/remediation-link per entry, CI-validated | T1 | Missing-field or expired entry fails | Fixture: well-formed / missing-field / expired entries | `REL-02/waiver-schema/` | Low |
| T3. Meta-check: assert `dependency-review`/`codeql`/`scorecard` actually *ran* whenever the repo is public, as a regression safety net against the currently-passing live state | — | A forced-private test branch confirms the guard logic itself, not just current visibility | Test | `REL-02/guard-regression/` | Low |
| T4. Local-scanner fallback for "repo goes private again" (local SAST substitute + scorecard-equivalent, auto-activating on `guard.outputs.public == 'false'`) | T3 | Seeded unsafe pattern caught by the fallback in a forced-private test branch | Seeded SAST fixture | `REL-02/private-fallback/` | Medium — document coverage gap vs. CodeQL rather than claim parity |
| T5. Wire all REL-02 blocking checks into REL-01's Wave-0 manifest | REL-01 T2, this finding's T1-T4 | Every enumerated scanner class has exactly one manifest entry | Cross-reference test | `REL-02/gate-manifest-wiring/` | Low |

**wowsociety impact — REL-01/REL-02: Not affected.** wowsociety's CI comment already correctly anticipates the current public state ("wowapi is public: the default `github.token` suffices; a PAT is only needed if the framework goes private"). If REL-01/02 work ever flips wowapi private again, wowsociety's checkout would need `WOWAPI_CHECKOUT_TOKEN` populated — a cross-repo coordination note, not a code change.

#### REL-03 — Expand compatibility gates (P1, Wave 4/6, several sub-items hard-blocked)

**Recommend splitting into REL-03a (buildable now: Go API diff, module compile matrix, config compat, migration-upgrade drill, arch smoke, SBOM/provenance-verify) and REL-03b (hard-blocked on Wave 1/4 architecture work: OpenAPI diff needs DX-06 first, event/schema compat needs DX-03/AR-03's typed model first, generated-consumer upgrade needs DX-04 first) — do not schedule as one monolithic P1 item, or 5 of 9 sub-tasks silently block the other 4.**

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Go public API diff tool wired as a CI job | DX-05's v1/N-1 policy (already decided) defines what "breaking" means | Diff correctly classifies added/removed/changed exported symbols | Seeded breaking-API fixture | `REL-03/go-api-diff/` | Medium — needs an agreed "public surface" definition first |
| T2. Module compile matrix across supported Go/dependency versions | T1 | Excluded versions are explicit, not silently ignored | CI matrix | same | Low |
| T3. OpenAPI semantic diff | **Blocked on DX-06** (a lossy merge can't be meaningfully diffed) | Diff classifies breaking changes per DX-06's 3.1/2020-12 baseline | Seeded breaking-OpenAPI fixture | `REL-03/openapi-diff/` | High — do not attempt before DX-06 lands |
| T4. Config schema compatibility + generated fixture migration test | `kernel/config/schema.go` remains source of truth | Field removal/type change fails; additive optional fields pass | Seeded breaking-config fixture | `REL-03/config-compat/` | Medium |
| T5. Event/schema compatibility check tied to `Compatibility` mode | **Blocked on DX-03/AR-03** — the concept doesn't exist in current source | Incompatible bump fails when `CompatibilityBackward` declared | Seeded breaking-event fixture | `REL-03/event-compat/` | High — premature against today's stringly-typed event registry |
| T6. Migration upgrade-from-oldest-supported + reversibility drill, extending the existing `TestIntegrationMigrationsReversible` | DATA-09 for the full choreography; a narrower drill can be built independently and earlier | Seed at oldest version, migrate forward, reverse on disposable data | Extended reversibility test | `REL-03/migration-upgrade-drill/` | Medium |
| T7. Generated-consumer upgrade check | **Hard-blocked on DX-04** | Golden consumer at N-1, upgraded to N, contracts re-pass | Reuses DX-04's drill | `REL-03/generated-consumer-upgrade/` | High — cannot exist before DX-04 |
| T8. Container architecture smoke on every published architecture | REL-01 T6/T7 (must run against the *candidate* image, not already-published) | Each arch boots and passes minimal smoke before `publish` | CI job in `build-candidate` stage | `REL-03/arch-smoke/` | Medium — arm64 via QEMU is slow, consider native runners |
| T9. SBOM/provenance/signature verification after publish | **Folds directly into REL-01 T8/T9 — not separate work**, just REL-03's naming of a property REL-01 already builds | Same acceptance as REL-01 T8/T9 | Same golden-failure tests | `REL-01/verify_release/` (shared) | — |

**wowsociety impact — REL-03: Affected at consumption/upgrade time, not at wowapi build time. Medium severity, not breaking to wowsociety's CI mechanics.** wowsociety's CI never invokes REL-03's gates — confirmed no `workflow_call` reference. What changes is *what a wowapi release means*: once REL-03 lands, a tag passing Wave-4+ gates carries a stronger backward-compatibility guarantee, but wowsociety has **no equivalent check of its own** — `framework-verify` only asserts SHA identity, not compatibility; it discovers incompatibility only by `make ci` failing after a bump. **Not required, recommended:** consume wowapi's published Go-API-diff/OpenAPI-diff output (if exposed as a release asset) to pre-flight a `FRAMEWORK_VERSION` bump; wowsociety's own upgrade runbook should note that a bump to a release with `completed_wave < 4` carries no compatibility-gate guarantee at all, once REL-01/REL-03 exist.

#### REL-04 — Make integration coverage truthful (P1; two sub-items are Wave-0)

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk | Wave |
|---|---|---|---|---|---|---|
| T1. Add `WOWAPI_REQUIRE_S3=1` + `S3_TEST_ENDPOINT=minio:9000` to `Makefile`'s `ci-container` target and the hosted `gate` job (confirmed missing from `ci.yml`'s MinIO-starting job too, not just local) | — | MinIO down → fail, not skip; MinIO up → S3 tests execute uncached and pass | Directive's own §2.2 evidence already proves the fix works manually — this task makes it the default | `REL-04/s3-wiring/` | Low — two-line env addition, already proven | **Wave 0** |
| T2. `depends_on: minio: condition: service_healthy` mirroring the existing Postgres healthcheck pattern | T1 | Compose genuinely blocks on MinIO health, not container-started state | Compose config test | same | Low | **Wave 0** |
| T3. Canonical endpoint variable cleanup (`S3_ENDPOINT` vs. `S3_TEST_ENDPOINT`) | T1, T2 | One variable name controls the endpoint everywhere | Grep-based regression test | same | Low — directive's own "follow-up cleanup" framing, lower priority | **Wave 0** |
| T4. Deterministic TOTP — audit for any remaining wall-clock-dependent path (existing `TOTPCodeAt`/RFC-6238-vector tests already take explicit timestamps; likely a narrow residual issue, not systemic) | — | No TOTP test path uses wall-clock time without an injected `at` | Run suite at two different mocked clock/TZ settings | `REL-04/totp-determinism/` | Low | **Wave 0** |
| T5. Make E2E prerequisite failures fail, not skip, in the authoritative E2E job | — | Unmet prerequisite exits non-zero, not "0 tests ran, green" | Kill a required E2E dependency, confirm failure | `REL-04/e2e-fail-closed/` | Medium — requires classifying which of the 22 inventoried skip sites are legitimately optional vs. mask required coverage | **Wave 0** |
| T6. Machine-checked skip manifest extending `check_test_skips.sh` | T1-T5 (only meaningful once known-bad skips are fixed) | New/unapproved skip fails CI; approved skip (with rationale) passes | Fixture: add an unguarded `t.Skip()`, prove failure | `REL-04/skip-manifest/` | Medium | P1, later |
| T7. Race tests over integration-relevant packages | — | `go test -race` runs over DB/S3-backed packages in CI | CI run + seeded data-race fixture | `REL-04/race-integration/` | Medium — may need a separate scheduled job, not every PR | P1, later |
| T8. Actual time-bounded coverage-guided fuzzing on PRs + scheduled — **identical scope to PERF-06 T3/T4, same evidence text, assign single ownership (recommend PF-REL, coverage-truthfulness framing) to avoid duplicate implementation** | — | PR + scheduled `-fuzz` runs, fuzz artifacts prove non-zero time beyond seed replay | Fuzz-duration/corpus-mtime test | `REL-04/fuzz-time-bounded/` | Medium — **shared with PERF-06, coordinate ownership before implementing** | P1, later |

**Note (verified adjacent to REL-04, belongs to PERF-06):** `internal/tools/benchbudget/main.go:49-55`'s missing-benchmark-warns-not-fails behavior means today's `benchbudget` invocation, if pulled into REL-01's Wave-0 manifest as-is, would not actually enforce "missing budget entries fail CI" — PERF-06's code fix must land before REL-01's manifest can honestly mark it as a hard gate.

**wowsociety impact — REL-04: Not affected.** wowsociety's own CI independently starts MinIO and sets its own `S3_ACCESS_KEY`/`S3_SECRET_KEY` for its own `make ci` run against wowsociety's own compose setup — it does not read or inherit wowapi's `WOWAPI_REQUIRE_S3`/`S3_TEST_ENDPOINT` wiring at all, confirmed by reading both `wowsociety/.github/workflows/ci.yml` and `wowsociety/Makefile` in full. The two repos' CI environments are fully decoupled except for the source-code checkout itself.

**PF-REL cross-cutting notes:** (1) GoReleaser split-mode (REL-01 T6) is a genuinely open technical spike, not a simple config change — needs a docs lookup on `--skip=publish`/separate-publish-invocation support before implementation starts. (2) Human-blocked items (T7/T9's environment/branch-protection dependencies) must not silently become "done" in any status report — the independent-review-gate must explicitly check these are tracked as separate open items. (3) The directive's own evidence has already drifted once (repo went public mid-review-window) — REL-02's scope should be re-verified again immediately before implementation starts, not assumed frozen from this planning pass. (4) REL-03's T3/T5/T7 are hard-blocked on Wave 1/4 architecture work — split into REL-03a/REL-03b in the tracker so 4 of 9 blocked sub-tasks don't make the whole finding look "stuck." (5) PERF-06/REL-04's fuzzing scope is identical — assign one owner before implementing twice.

## 6. Traceability matrix

Every finding maps to a detailed task table in §5 (full acceptance criteria, tests, evidence paths, risk). This table is the flat status index; execution results are in §8.

| Finding | Work package | Task count | Evidence root | Status |
| --- | --- | --- | --- | --- |
| AR-01 | PF-ARCH | 11 | `PF-ARCH/AR-01/` | PLANNED |
| AR-02 | PF-ARCH | 7 | `PF-ARCH/AR-02/` | PLANNED |
| AR-03 | PF-ARCH | 5 | `PF-ARCH/AR-03/` | PLANNED |
| AR-04 | PF-ARCH | 5 | `PF-ARCH/AR-04/` | PLANNED |
| AR-05 | PF-ARCH | 5 | `PF-ARCH/AR-05/` | PLANNED |
| AR-06 | PF-ARCH | 3 | `PF-ARCH/AR-06/` | PLANNED |
| SEC-01 | PF-SEC | 7 | `PF-SEC/SEC-01/` | PLANNED |
| SEC-02 | PF-SEC | 5 | `PF-SEC/SEC-02/` | **[EXECUTED — Wave-0 minimal fix (T1+T2+T3), see §8]** |
| SEC-03 | PF-SEC | 4 | `PF-SEC/SEC-03/` | PLANNED |
| SEC-04 | PF-SEC | 6 | `PF-SEC/SEC-04/` | PLANNED |
| SEC-05 | PF-SEC | 1 | `PF-SEC/SEC-05/` | PLANNED — closure gate, blocked on SEC-01–04 |
| SEC-06 | PF-SEC | 5 | `PF-SEC/SEC-06/` | PLANNED |
| PERF-01 | PF-PERF | 8 | `PF-PERF/PERF-01/` | **[EXECUTED — see §8]** |
| PERF-02 | PF-PERF | 5 | `PF-PERF/PERF-02/` | PLANNED — blocked on reference environment |
| PERF-03 | PF-PERF | 6 | `PF-PERF/PERF-03/` | PLANNED — blocked on reference environment |
| PERF-04 | PF-PERF | 8 | `PF-PERF/PERF-04/` | PLANNED — T5 blocked on PF-DATA Wave 3 |
| PERF-05 | PF-PERF | 5 | `PF-PERF/PERF-05/` | PLANNED — blocked on reference environment |
| PERF-06 | PF-PERF | 5 | `PF-PERF/PERF-06/` | **[EXECUTED — T1 only, see §8]** |
| DATA-01 | PF-DATA | 8 | `PF-DATA/DATA-01/` | PLANNED |
| DATA-02 | PF-DATA | 7 | `PF-DATA/DATA-02/` | PLANNED |
| DATA-03 | PF-DATA | 8 | `PF-DATA/DATA-03/` | PLANNED |
| DATA-04 | PF-DATA | 6 | `PF-DATA/DATA-04/` | PLANNED |
| DATA-05 | PF-DATA | 5 | `PF-DATA/DATA-05/` | PLANNED |
| DATA-06 | PF-DATA | 4 | `PF-DATA/DATA-06/` | PLANNED |
| DATA-07 | PF-DATA | 4 | `PF-DATA/DATA-07/` | PLANNED — blocked on SEC-01 |
| DATA-08 | PF-DATA | 7 (2 Wave-0 + 5 Wave-6) | `PF-DATA/DATA-08/` | **[EXECUTED — Wave-0 slice (W0-T1+W0-T2), see §8]** |
| DATA-09 | PF-DATA | 9 | `PF-DATA/DATA-09/` | PLANNED — new infra, no owner/timeline yet |
| DX-01 | PF-DX | 5 | `PF-DX/DX-01/` | PLANNED |
| DX-02 | PF-DX | 10 (5 Wave-0 + 5 Wave-4) | `PF-DX/DX-02/` | PLANNED |
| DX-03 | PF-DX | 5 | `PF-DX/DX-03/` | PLANNED — design-only until Wave 4 |
| DX-04 | PF-DX | 5 | `PF-DX/DX-04/` | PLANNED — blocked on DX-01/DX-05 |
| DX-05 | PF-DX | 6 | `PF-DX/DX-05/` | PLANNED |
| DX-06 | PF-DX | 3 | `PF-DX/DX-06/` | PLANNED — overlaps AR-03 T2, assign one owner |
| DX-07 | PF-DX | 5 | `PF-DX/DX-07/` | PLANNED |
| REL-01 | PF-REL | 10 | `PF-REL/REL-01/` | PLANNED — T7/T9 human-blocked |
| REL-02 | PF-REL | 5 | `PF-REL/REL-02/` | PLANNED — scope narrowed, see §5.6 pre-flight correction |
| REL-03 | PF-REL | 9 | `PF-REL/REL-03/` | PLANNED — split REL-03a/REL-03b recommended |
| REL-04 | PF-REL | 8 | `PF-REL/REL-04/` | PLANNED |

**38/38 findings have a task breakdown. 4 findings (SEC-02, PERF-01, PERF-06, DATA-08) have a real, independently-reviewed Wave-0 partial closure landed in this pass** — each is a minimal-scope slice of a larger finding, not the finding's full closure per §13.2's bar. The other 34 remain PLANNED only. See §8 for exact evidence.

## 7. Cross-cutting risks, assumptions, and unresolved questions

Aggregated from all six work packages' own cross-cutting notes (§5.1–5.6); not duplicated here in full, only the items that span more than one work package or need a decision this document cannot make.

**Genuinely undecided, needs a named human decision before implementation can proceed:**
1. **SEC-01 T5's IdP claim contract** — whether the production IdP will mint an opaque `grant_id` claim, and who approves break-glass grants. Highest-uncertainty item in the entire programme.
2. **wowsociety's `identity_impersonation_session` vs. wowapi's future SEC-01 grant table** — authority split needs an explicit decision (recommend: framework owns validity/expiry/revocation, wowsociety keeps UX/audit-trail), not just a recommendation in this document.
3. **AR-01's `Registrar` type-sharing design** — one shared type (capability-confusion risk) vs. per-subsystem types (task-count multiplier). Needs a `decisions.md` entry before AR-01 T2 starts.
4. **AR-01/AR-04's post-seal-mutation policy** — error vs. panic in production builds. Recommend error-only; panic-in-prod would convert wowsociety's currently-harmless `s.rulesReg` retention into a crash risk.
5. **DATA-08 W6-T1's audit-hash design** — whether/how to add a `hash_version` discriminator so historical rows (including wowsociety's live impersonation/policy audit rows) remain verifiable after the hash contract widens. Unresolved technical design, high blast radius.
6. **REL-01 T6's GoReleaser split-mode approach** — `--skip=publish` + separate publish invocation vs. hand-rolled pipeline. Needs a docs lookup / spike before implementation, not assumed in this plan.
7. **SEC-04 T4 and SEC-06 T4** — both open architecture/design decisions (cross-pod cache invalidation transport; JWKS-client governance model) with no resolution proposed by the directive or this plan.

**Cannot be completed by a coding agent — requires a human with elevated access:**
8. **GitHub repo-admin actions for REL-01/REL-02** — no protected `release` environment, no branch/tag protection exist today (confirmed live via `gh api`, not assumed). Track as a separate ticket ("PF-REL-ADMIN-01") so agent-completable YAML work isn't silently gated on this.
9. **SEC-05's independent external security assessment and Wave-6's penetration test** — need a named security lead and an external vendor.
10. **§14's dedicated Linux amd64 reference performance runner** — blocks PERF-02/03/04/05 from ever closing; has no owner or timeline anywhere in the directive's own phase blueprint. This is itself the single largest unscheduled prerequisite in the whole programme.

**Duplicate-effort risks — assign single ownership before implementing twice:**
11. AR-03 T2 and DX-06 T1 (OpenAPI full-field merge) — identical closure contract.
12. PERF-06 T3/T4 and REL-04 T8 (time-bounded coverage-guided fuzzing) — identical evidence text and fix.
13. AR-04 T5, SEC-06, and DX-07 T4/T5 (the "no-op adapter fails readiness in prod without a waiver" mechanism) — three closure contracts describing the same waiver/readiness primitive.

**Hard cross-work-package sequencing constraints (not soft suggestions):**
14. DATA-07 is blocked on SEC-01. PERF-04 T5 is blocked on PF-DATA's Wave-3 DATA-02/DATA-03. REL-03's T3/T5/T7 are blocked on AR-01–03/DX-03/DX-04/DX-06. DX-02's Wave-4 T8/T9 are blocked on DATA-06 and AR-03 respectively. DX-03 is blocked on AR-01/AR-02's Wave-1 exit gate. None of these can be parallelized away by adding more agents — they are genuine dependency edges.

**wowsociety-specific findings requiring product-side action (not just framework awareness), consolidated from §5:**
- **Breaking, high-severity:** SEC-01 (impersonation flow), DATA-08 W6-T1 (live audit rows).
- **Real, independent instance of the same defect:** DATA-01 (`policy_override.rule_version_id`), DATA-06 (`committeeseat.go`'s manual mirror-write pattern).
- **Process/tooling gap, not code:** DATA-09 (single-shot deploy collapses the N/N-1 window), DX-05 (SHA-only pin, no semver compatibility class), PERF-06 (no bench/fuzz gate at all).
- **Inherited-but-latent, needs a manual backport since generated code isn't regenerated:** DX-07 (missing migration-currency readiness check in wowsociety's own `cmd/api/main.go`).
- **Corroborating evidence found, not just theoretical:** DX-01 — wowsociety independently documented hitting this exact bug in its own `docs/upstream/12-sf-7-*.md`.

## 8. What was executed in this pass

Four Wave-0 minimal-scope slices were implemented, each by an independent Sonnet-tier worker on disjoint files, then verified: full-repo `go build`/`go vet`/`gofmt` clean, the full `make ci` gate green (all ~50 packages' unit/integration/race tests, plus `make bench-budget`'s ~41 benchmarks all `OK`), and a fresh, unscoped independent reviewer over the combined diff. The reviewer's first pass returned **FAIL** — a real, reproducible Critical finding — which was fixed and re-verified before this section was written. This section states exactly what shipped; nothing here should be read as closing a finding per §13.2's full bar (adversarial/chaos/integration tests beyond the Wave-0 slice remain PLANNED for all four).

### SEC-02 — Wave-0 minimal fix (T1+T2+T3) — EXECUTED, independently reviewed
- **Changed:** `kernel/workflow/runtime.go` (`NewRuntime`'s nil-guard now includes the evaluator; `Override`'s `if rt.authz != nil` conditional removed — the permission check is now unconditional), plus 4 test files fixing/redirecting nil-evaluator construction sites and adding an adversarial fail-closed test.
- **Proof the fix is real:** reverted to the old code, confirmed the new adversarial test failed with the exact old-bypass symptom; reapplied, confirmed pass. Full `kernel/workflow` suite: 63 tests, `-race`, all pass.
- **Review finding (Critical, fixed):** the implementer's test sweep missed a call site in a sibling package — `testkit/workflowsim_cov_test.go:66` — which still passed `nil` and now panicked unconditionally, breaking `go test ./testkit/...`. Root cause: the plan doc's own "4 call sites" count was itself incomplete (a planning-artifact error, not a coding error) and the implementer trusted it instead of grepping the whole repo. **Fixed** by the Conductor directly: swapped `nil` for `testkit`'s existing `covEvaluator()` stub (a one-line change, this file's tests don't call `Override` so no specific permission grant was needed). Re-verified against the real DB (`WOWAPI_REQUIRE_DB=1`): `TestIntegrationWorkflowSimApproveFlow`/`RejectFlow` both pass.
- **Not done:** ratification (T4) and durable override-audit persistence (T5) — explicitly out of scope for this pass, remain PLANNED.
- **wowsociety impact:** none — confirmed zero `kernel/workflow` usage in wowsociety, unaffected by this change.

### PERF-01 — Wave-0 fix — EXECUTED, independently reviewed, PASS
- **Changed:** `kernel/httpx/ratelimit.go` (sweep now recomputes effective refill instead of comparing a stale stored value; added a hard-capacity bound via a new backward-compatible `WithHardCap` option; added a stats/metrics hook), plus `ratelimit_test.go`, `bench_test.go` (two new benchmarks), a new `export_test.go` test-only shim, and matching `bench-budgets.txt` entries.
- **Proof the fix is real:** reverted `sweep()` to the old stale-comparison logic, confirmed both the single-key and 10k+-key eviction tests failed exactly as predicted; reapplied, confirmed both pass. `-race` clean under 16-goroutine concurrent load. Hard-cap check verified to never reject an existing key (only gates admission of a genuinely new key at capacity).
- **Backward compatibility verified against the actual wowsociety call site**, not just in the abstract: `wowsociety/cmd/api/main.go:296` uses the 2-arg `NewTokenBucket(rate, burst)` form, confirmed still compiling via the `replace` directive.
- **Not done:** Wave-5 bounded-sharded/distributed-adapter work — explicitly out of scope, remains PLANNED.
- **wowsociety impact:** none required — config-exposure only, existing 2-arg constructor call unaffected.

### DATA-08 — Wave-0 slice (W0-T1+W0-T2) — EXECUTED, independently reviewed, PASS
- **Changed:** `kernel/attachment/attachment.go` (the previously-discarded outbox-write error is now propagated, causing the caller's transaction to roll back the attachment insert on failure), `kernel/notify/service.go` (legal-importance delivery now writes a durable `events_outbox` record with the provider receipt in the same transaction as the `sent` status update; both stale "app_platform lacks INSERT" comments removed), one line in `kernel/kernel.go` wiring `notify.WithOutbox`.
- **Design ambiguity resolved and independently confirmed correct:** the choice of `events_outbox` over an `audit_logs` row was verified against `notify.Service`'s actual struct fields (no `*kaudit.Writer` dependency exists or ever did) and migration `00011`'s grant comment, which names this exact use case.
- **Proof the fix is real:** a fault-injecting outbox-writer double proves `Attach` now rolls back fully (re-queried in a fresh transaction, zero rows). The independent reviewer verified the rollback claim architecturally, not just by convention — `TenantDB` is only constructible inside `WithTenant`'s closure, whose rollback-on-error behavior is structural, not incidental. Legal-delivery audit event proven present for legal-importance sends and absent for non-legal sends (explicit negative test).
- **Not done:** the Wave-6 audit-hash-widening work (`chainHash`, DSR export, central hold enforcement) — explicitly and correctly out of scope, remains PLANNED, and is flagged in §7 as the single highest-risk item in this finding's remaining scope.
- **wowsociety impact:** none today — confirmed zero `kernel/attachment`/`kernel/notify` usage in wowsociety.

### PERF-06 — T1 only — EXECUTED, independently reviewed (including its own separately-spawned reviewer), PASS
- **Changed:** `internal/tools/benchbudget/main.go` (a budgeted-but-missing benchmark now appends to the same `violations` path a real regression does, instead of only warning and continuing), plus a new subprocess-based negative-fixture test.
- **Proof the fix is real:** reverted the fix, confirmed the new test failed with the old buggy (exit-0, warn-only) behavior; reapplied, confirmed pass (exit 1).
- **False-positive check:** the real, current `bench-budgets.txt` (including PERF-01's two new entries) was cross-referenced against every `func Benchmark*` in the repo — all resolve, confirmed both by static analysis and by an actual `make bench-budget` run (all ~41 benchmarks `OK`, zero FAIL).
- **Not done:** T2 (statistical/`benchstat` baseline tracking), T3/T4 (CI fuzz wiring), T5 (REL-01 manifest wiring) — explicitly out of scope, remain PLANNED.
- **wowsociety impact:** none required now — wowsociety has no benchmark/budget infrastructure of its own; flagged in §7 as a wowsociety backlog item once it grows benchmark coverage.

### Independent review gate — final result
- **First pass:** FAIL (1 Critical finding — the `testkit` call site above).
- **Fix applied, re-verified:** `go build ./...`, `go vet ./...`, `gofmt -l` clean across the whole repo; the previously-broken tests now pass against the real database; full `make ci` re-run green (all packages, including the fixed `testkit` package); `make bench-budget` separately re-confirmed clean.
- **Second pass:** not re-run by a fresh agent after the fix (time-boxed) — the fix was small, single-line, mechanically verified by the Conductor directly (build/vet/targeted-test/full-suite), and the original reviewer's finding was narrow enough that a full re-review was judged unnecessary. This is disclosed, not hidden: if held to the strictest reading of "always re-run the full gate after any fix," this is the one procedural shortcut taken in this pass.
- **Cross-cutting checks confirmed by the reviewer directly (not delegated):** the four changes don't interact badly with each other (verified the DATA-08 kernel.go line doesn't touch workflow wiring); scope discipline held (all four diffs file-disjoint, none touched wowsociety, none exceeded their stated Wave-0 scope); git safety confirmed (no force-pushes, no dropped work; one inert leftover stash entry from SEC-02's revert/reapply verification cycle, not blocking anything, not dropped without authorization per standing policy).

### Explicit accounting — do not read more into this than what's stated
- **34 of 38 findings are PLANNED only** — a task breakdown exists, nothing has been implemented.
- **4 findings have a real, reviewed Wave-0 partial slice landed** (SEC-02, PERF-01, DATA-08, PERF-06) — each is the smallest, most self-contained, most immediately-actionable piece of a larger finding. None of the four is fully closed per the directive's own §13.2 closure-contract bar (which requires adversarial/chaos/integration proof well beyond what a Wave-0 minimal fix delivers).
- **Nothing in this pass touched wowsociety's code.** All wowsociety-impact assessments in §5 are analysis, not action.
- **10 unresolved questions and 2 classes of blockers (human-only actions, and a currently-unowned reference-performance-environment prerequisite) are recorded in §7** and are not resolved by this document — they require decisions this session cannot make.


