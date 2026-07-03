# Phase 1 — Review Findings

Two parallel critique agents reviewed the Phase 1 slice (2026-07-03):
**S** = security/config reviewer (findings SEC-3…SEC-10 + 1 info; three findings reproduced by
live loader probes); **A** = architecture/API reviewer (ARCH-6…ARCH-15 + acceptance-coverage map).
Resolutions were split between the lead (kernel/config, internal/cli, docs) and a fix agent
(app, kernel/logging, lint script); all fixes carry regression tests.

| ID | Sev | Finding (file) | Resolution | Status |
|---|---|---|---|---|
| SEC-3 | high | unsafe-knob prod refusal bypassed when the unsafe value is the compiled default (check lived on the tree-present path only) — reproduced | enforcement moved to post-bind `enforceUnsafe` pass over final bound values (D-0019); `TestLoadUnsafeDefaultRefusedInProd` | **fixed** |
| SEC-4 | high | `unsafe` tag silently unenforced on struct/Secret/slice/pointer fields; schema still advertised x-unsafe — reproduced | same post-bind pass covers all field kinds (zero-value = not set); `TestLoadUnsafeStructKnobRefusedInProd` + unset-is-fine test | **fixed** |
| SEC-5 | medium | `environment` downgradable by env var over committed prod file; flag-set environment escaped the flags-in-prod guard — reproduced | trust rules (D-0017): flags may never set environment; env-var/file mismatch is an error; prod checks key off file-layer value; 3 regression tests | **fixed** |
| SEC-6 | medium | `config validate --env prod` only picked the overlay file; a mis-declaring prod.yaml validated under laxer rules (CI gate under-checks) | `assertEnv` in validate/print/doctor: loaded environment must equal `--env`; `TestConfigValidateEnvMismatchFails` | **fixed** |
| SEC-7 | low/med | yaml parse errors echo file content (backtick node payloads) — a malformed raw secret would print | `scrubYAMLError` strips backtick payloads, keeps positions; `TestLoadYAMLParseErrorScrubbed` | **fixed** |
| SEC-8 | low | binder conversion errors echoed raw values (`cannot parse %q as integer`) | all conversion errors report path + expected type only; `TestLoadConversionErrorsDoNotEchoValues`; `TestLoadErrorsAccumulate` asserts no echo | **fixed** |
| SEC-9 | low | logging redaction only rewrote string/any kinds; `"token", 123456` passed through | sensitive keys now redacted regardless of value kind; integer+duration tests | **fixed** |
| SEC-10 | low | module namespace maps aliased the parse tree; `MapView` down-cast allows mutation of own view | namespaces deep-copied at capture (no aliasing, no cross-view sharing). Down-cast mutation of a module's *own* copy remains possible: accepted for Phase 1 — isolation (the security property) holds; unexported view type reconsidered with the Phase 5 module SDK | **fixed (copy) / accepted (own-view mutability)** |
| SEC-info | info | fail-closed env guard reads root `tree["environment"]` positionally, decoupled from field binding | documented in load.go: the environment key must live at the document root (where Framework binds it); nesting/renaming unsupported | **documented** |
| ARCH-6 | high | binder rejected pointer fields — next config-bearing phase (DB/auth) would hit a wall or contort shapes | pointer branch: nil when absent, allocate+bind when present, default-tag support, *Secret/*Namespaces rejected; `TestLoadPointerFields` | **fixed** |
| ARCH-7 | medium | whole-view fingerprints make §7 shared-section drift detection impossible (api vs worker always differ) | `SectionFingerprints()` per view via shared helper; tests prove shared sections agree and diverge correctly | **fixed** |
| ARCH-8 | medium | module namespace values from env vars arrive as strings and fail the module's typed Decode confusingly | loader rejects env-var/flag-sourced `modules.*` keys with a clear error (D-0018, revisit Phase 5); `TestLoadModuleNamespaceViaEnvVarRejected` | **fixed** |
| ARCH-9 | low | generated schema omitted `environment` from `required` while the loader hard-requires it | schema special-cases the fail-closed key: required + `x-fail-closed: true`; asserted in `TestSchemaFromTags` | **fixed** |
| ARCH-10 | low | product `Validate()` shadows promoted `Framework.Validate()`; prod-safety floor silently skipped unless product delegates | composition contract documented at the loader hook; `wowapi init` scaffold (Phase 10) must generate the delegating shape. Loader-side auto-call rejected: detecting promotion vs. own method needs magic beyond R5's budget | **documented (accepted)** |
| ARCH-11 | low | `moduleContext.Logger()` re-tagged (allocated) per call | logger tagged once in `newModuleContext` | **fixed** |
| ARCH-12 | info | `config.Options` diverged from blueprint sketch without a decision | D-0016 recorded; blueprint 12 §2 updated to the final shape | **fixed** |
| ARCH-13 | info | D-0005 wording ("no config imports needed") contradicted by kernel/logging importing config types | D-0005 amended: types-only kernel→kernel edges are the intended reading | **fixed** |
| ARCH-14 | info | `internal/cli` is a second composition root but the boundary lint didn't govern it | lint rules added: internal/cli must not import module/testkit/examples/testmodules | **fixed** |
| ARCH-15 | info | no test for non-struct `Load[T]`; embedded conf-key collisions resolved silently | collision guard in `bindStructInto` (claimed-key map) + `TestLoadEmbeddedKeyCollision`; `TestLoadNonStructTarget` | **fixed** |

Reviewer-confirmed solid (no action): structural `Secret` redaction across all render paths;
`Loaded` never returns partially-bound config or resolved secrets on error; fingerprint hashes
redacted JSON only; unknown-key defense (incl. empty maps, nil, env-var typos); envprovider and
`secrets.ParseRef` error hygiene.

Residual risk:
- Blueprint 12 §5's CLI redaction snapshot test cannot exist until `config.Framework` gains its
  first `Secret` field — **Phase 2 exit item** (DB DSN).
- The logging ReplaceAttr denylist stays leaf-key/heuristic (group traversal out of scope) — the
  security boundary remains the structural `Secret` type.
- ARCH-10's contract is documentation until the Phase 10 scaffold enforces it by generation.
