# wowapi Research Package — Final Engineering Review & Backlog

Review lead: Fable 5 (coordination/synthesis only). Verification: 7 specialist reviews this session — i18n, scaffold/config, auth/security, rules/lifecycle (executed proofs), product migration (all five at be84ee2 against the design review), plus wowapi-internals and external-frameworks verifiers for the new benchmark doc. Package = gap-analysis (cross-link edits only), gap-design-review (all 8 gate edits verified applied), framework-competitive-architecture-benchmark (new).

> **Execution status (2026-07-10, branch `feat/backlog-p1`, unpushed):** all P0 and P1 items COMPLETE and verified.
> P0: B1 ✅ (i18n subsystem), B2–B5 ✅ (prior wave, `feat/backlog-p0`). P1: B6 ✅, B7 ✅, B8 ✅, B9 ✅, B10 ✅.
> Each item was TDD-built, independently reviewed, and coordinator-verified (acceptance tests re-run). Whole-branch gate green: lint 0, boundaries OK, 54 pkgs ok/0 FAIL, security 0 FAIL, coverage 92.1%.
> Deferred (unchanged): P2 B11 (router — parked by B5 data), B12, B13; W1 (wowsociety i18n migration) unlocked by B1. B8 MaxAge/auth_time pending IdP capability (Decision 4).

## 1. Executive summary

The research package is trustworthy and engineering-ready after three small corrections to the benchmark doc. Every external-framework claim (Gin/Spring/FastAPI/Django/Axum/Laravel, 14 URLs) verified against primary sources. Every wowapi-internal claim verified file:line except two: the doc **undersells a live SSRF exposure** (kernel/webhook's default HTTPSender is a bare http.Client — the "safe outbound client" belongs in P0, not P1) and **overclaims "no benchmarks exist"** (a bench suite with allocation budgets exists; what's missing is a full-chain, high-route-cardinality dispatch benchmark). The design review's findings were all previously verified accurate; its corrected version (sms/NIST, auth_time prerequisite, preventive MFA framing) is in the package. Net: 13 framework backlog items (5×P0, 5×P1, 3×P2-conditional), 1 product-migration item, 1 docs item.

## 2. Verified findings (claim → evidence)

- i18n is plumbing, not a subsystem: bare map catalog (kernel/i18n/catalog.go), Go-bundle-only registry, hardcoded framework strings (framework_catalog.go), no loaders/publish/validate CLI/scaffold dir/config section; worker+migrate templates have zero i18n references; product still registers Go maps via unguarded Catalog.Add (wowsociety internal/i18n/messages.go, cmd/api/main.go:185).
- Rules fail open: unknown `type` accepts anything (typeMatches), unknown keywords silently dropped at json.Unmarshal, defaults never validated at Register (all three proven by executed tests); Resolver.Resolve never validates despite rules.go:39 doc; stale "JSON Schema" claims live in rules.go field doc + migration 00008 comment (schema.go itself already discloses its subset).
- Standalone `wowapi seed sync` bypasses composed config (DATABASE_URL + config.Defaults().DB, seed_cmd.go:144) AND never runs rules.SyncDefinitions; generated migrate does both correctly.
- Step-up: boolean end-to-end; strong-factor set hardcoded incl. `sms` (evaluator.go:197 — NIST 800-63B restricted authenticator); challenge always `step_up="mfa"` (authz_gate.go:122); `auth.Claims` has no auth_time → MaxAge freshness unimplementable today.
- Privileged(): always constructs privileged.Config{} (app/context.go:280-284); no product config section for allow-lists.
- Webhook SSRF (upgraded by verification): kernel/webhook/sender.go HTTPSender = bare net/http.Client, no private-IP/loopback/metadata blocking, wired by default in generated products. LIVE gap, not hypothetical.
- Concurrency: all knob claims exact; no http max-in-flight, no cross-resource capacity validation, no backpressure; rate limiter is per-key token bucket, not a global cap.
- DI: Kernel ~33 fields, moduleContext ~35, moduleDeps triplicates; no manifest/codegen/scopes.
- Baseline strengths all confirmed (RouteMeta fail-closed boot, gateRoute order, strict decode, SET LOCAL tenant tx, SKIP LOCKED jobs, boot validation incl. i18n ownership at app/boot.go:215-217, middleware order matches template exactly).
- External-framework characterizations: 10/10 confirmed; Laravel 13.x real; Go 1.22+ ServeMux framing accurate and hedged.

## 3. Corrected / inaccurate findings

| Where | Correction | Severity |
|---|---|---|
| Benchmark matrix row 1 ("No route-dispatch benchmark or high-cardinality routing budget exists") | Overclaims. `kernel/httpx/bench_test.go` (BenchmarkRouterHandle/Routes/RouteMetaValidate/TokenBucketAllow/EdgeMiddlewareChain) + `bench-budgets.txt` + `make bench` exist and run. Reword: "benchmarks exist for registration, token bucket, and edge chain, but none exercise full ServeMux→SecureHandler→authz dispatch at high route cardinality." | Medium |
| Benchmark §Security + P1.3 (safe outbound client as future profile work) | Undersells: the framework ALREADY ships an outbound path (kernel/webhook.HTTPSender) with zero SSRF protection, default-wired. Move to P0 and name the sender. | High |
| Benchmark §Concurrency knob list | `reclaimTimeout` is `stalledTimeout` in code (set via WithReclaimTimeout). | Low |
| Design review | All 8 previously-required gate edits verified applied; no open corrections. | — |
| Prior-session corrections retained | FG-POST-005 "marketed as MFA support" premise unsupported (docs consistently say primitives); seed-sync "no framing exists" softened (source comment frames it; --help/docs don't). | — |

## 4. Final backlog

### Framework — P0 (before broad merge/announcement)

**B1. GAP-001B: i18n source/loading/tooling subsystem** — Effort L
- Problem: locale negotiation + catalog exist, but no source-of-truth, loader, override, validation, or scaffold story; framework strings hardcoded in Go; product carries Go-map registration via unguarded Catalog.Add.
- Evidence: §2 bullets 1; design-review FG-POST-001..003 (all verified).
- Fix: `i18n.Source` loader contract; framework defaults as embedded per-locale YAML (kernel/i18n/locales/<locale>/kernel.yaml); publish command; product YAML + JSON loaders + .go catalog bundles; optional DB overlay last; precedence = embedded defaults → product framework-overrides → product/module catalogs → Go bundles → DB overlay, intra-layer duplicates fail; guarded RegisterFrameworkLocale path (kill raw Add usage); catalogs frozen after boot (Decision 3); `wowapi i18n validate` (coverage, ownership, placeholder drift, duplicates); scaffold locales/ dir + i18n config section; catalogs load in api, worker, AND migrate before serving.
- Impact: unblocks every future product's localization; removes the largest "plumbing-not-capability" gap.
- Dependencies: none (unlocks W1 product migration).
- Acceptance: benchmark doc's 5 i18n acceptance tests (init renders 3 binaries loading same sources; product locales/mr/kernel.yaml overrides embedded kernel.* while en falls back; YAML/JSON/Go bundles share one lifecycle; DB overlay wins last post-validation; i18n validate fails on the 4 defect classes) + kernel.* coverage check for every product-declared locale.
- Risks: precedence/override design is one-way door;  scope creep into interpolation/plurals — decide support or explicitly document static-only.

**B2. SSRF-safe outbound HTTP client, applied to webhook sender** — Effort M (escalated from P1)
- Problem: kernel/webhook.HTTPSender delivers webhooks with a bare http.Client — user-configurable URLs can reach loopback/RFC1918/link-local/cloud-metadata endpoints.
- Evidence: kernel/webhook/sender.go (verified: no IP filtering, no allowlist); default-wired in generated products.
- Fix: kernel/httpclient with dial-time DNS/IP blocking (loopback, link-local, RFC1918, metadata) + allowlist; make HTTPSender use it by default; config escape hatch for intentional internal targets.
- Impact: closes a live SSRF vector in every deployed product.
- Dependencies: none. B7 SecurityProfile later consumes the same client.
- Acceptance: unit tests proving blocked dial per address class (incl. DNS-rebind: resolve-then-verify), allowlist override works, webhook delivery to public hosts unaffected; docs updated.
- Risks: breaks legitimate internal webhook targets — ship allowlist + clear error naming the config key; redirect-following must re-check each hop.

**B3. Rules schema fail-closed + contract honesty** — Effort S/M
- Problem/Evidence: §2 bullet 2 (proofs on file).
- Fix: reject unknown `type` and unknown keywords at Register/SyncDefinitions (boot-accumulation pattern exists — confirmed implementable); validate defaults at Register; validate at resolve or fix rules.go:39 doc; correct rules.go ValueSchema doc + migration 00008 comment to the documented subset; per Decision 2: rename to a strict RuleValueSchema grammar (no JSON-Schema library).
- Impact: silently-unenforced rule points become impossible.
- Dependencies: none (Decision 2 resolved the design fork). Acceptance: registration-time rejection tests (unknown type/keyword/default-violates-schema), resolve-path test, existing suites green. Risks: existing registered points with sloppy schemas will now fail boot — sweep wowsociety's rule points first (verified currently clean).

**B4. Lifecycle CLI alignment** — Effort S
- Problem/Evidence: §2 bullet 3.
- Fix: `--help` + user docs mark `wowapi seed sync` as low-level escape hatch; add rules.SyncDefinitions to it (or generate product tools/seedsync via appcfg.Load — pick one, recommend the CLI fix now, generated command with B6 config work).
- Acceptance: standalone path syncs both catalogs; docs/help updated; drift test comparing CLI vs generated-migrate behavior. Risks: none material.

**B5. Full-chain dispatch benchmark gate** — Effort S
- Problem: existing bench suite lacks a high-route-cardinality ServeMux→SecureHandler→authz dispatch benchmark (the doc's matrix overclaimed; this is the real residue).
- Fix: add BenchmarkDispatch (static/param/wildcard mix at N routes), authz-gate cached/uncached, JSON decode/body-limit benches to bench_test.go + budgets to bench-budgets.txt.
- Impact: makes P2 router decisions data-driven; guards regressions. Dependencies: none. Acceptance: `make bench` runs them; budgets recorded. Risks: none.

### Framework — P1 (next cycle)

**B6. ConcurrencyProfile + capacity validation + backpressure** — Effort M — per benchmark §Concurrency: config.ConcurrencyProfile (http_max_in_flight, worker caps, platform reservations, overload status/retry-after); fail-closed validation `replicas*(runtime+platform)+migrate+reserve <= db max_connections`; backpressure middleware answering 503/429 before pool exhaustion; overload/pool-wait/saturation metrics; `wowapi config capacity` lint. Evidence: knob inventory verified; negatives verified. Acceptance: validation fails oversubscribed shapes; load test proves 503 before DB exhaustion. Risks: default budget too tight breaks existing deploys — ship advisory-then-enforced.

**B7. SecurityProfile (API vs browser/session)** — Effort M — profiles per benchmark §Security; CSRF/SameSite/CSP only in browser mode; consumes B2 client; config validate + scaffold tests prove selected profile wired. Dependencies: B2. Risks: browser mode is meaningful new surface — keep API-only default untouched.

**B8. Step-up policy registry** — Effort M — first task: confirm the production IdP reliably emits auth_time — if yes wire it end-to-end and expose MaxAge, if no ship AMR-only policy WITHOUT a MaxAge field (Decision 4); StepUpPolicy{RequiredAMR, Challenge(, MaxAge iff auth_time)} with `step_up: true` shorthand; strong-factor set moves to config/policy with sms EXCLUDED from the default set, opt-in only (Decision 5); challenge advertises the policy's factor. Evidence: §2 bullet 4. Acceptance: per-permission AMR subsets; sms absent from defaults + opt-in works without code change; freshness test iff auth_time wired. Risks: token-shape dependency on IdP.

**B9. Static provider/lifecycle manifest** — Effort M/L — descriptors (process/request/tenant_tx/job/migrate scopes) generated from kernel/app/module wiring; CI validation of scope leaks/missing providers; surface in `wowapi doctor`. Evidence: field counts verified (~33/~35/triplicated). Risks: codegen maintenance; start with manifest+lint, not full DI.

**B10. Privileged allow-list config** — Effort S — product config section, boot-validated, passed into Privileged(); scaffold docs/tests for an allow-listed module. Evidence: app/context.go:280-284. Risks: widening is security-sensitive — require explicit key enumeration, never wildcards.

### Framework — P2 (only if data/need demands)

**B11.** Radix-router strategy — only if B5 budgets show ServeMux dispatch matters; must preserve RouteMeta/boot-validation/OpenAPI contracts. **B12.** Typed schema unification (validation/OpenAPI/codegen single source). **B13.** Hot-reloadable DB overlays for i18n/rules — opt-in overlay only, gated on a demonstrated operational need (Decision 3), after immutability/validation/metrics/invalidation semantics are defined.

### Product migration (wowsociety)

**W1. Consume i18n loaders** — after B1: move internal/i18n Go maps to locales/ files (en + mr incl. kernel.* mr coverage), delete manual Register/Catalog.Add path, adopt publish/override flow, run `wowapi i18n validate` in product CI. Effort S/M. Acceptance: no Go-map catalogs; mr framework strings localized (no silent English fallback); product suite green.

### Docs/workflow

**D1. Benchmark doc corrections** — apply §3 rows 1-3 (matrix row-1 rewording, SSRF→P0 with kernel/webhook.HTTPSender named, stalledTimeout naming). The doc is uncommitted — apply before circulating. Effort S.

## 5. Decisions (ratified 2026-07-10)
1. MFA: primitives-only. kernel/mfa stays crypto + helpers; the service layer (factor persistence, recovery codes, challenge rows, replay protection, rate-limited enrollment) is an explicit non-goal until someone is ready to own it. No P1 MFA item.
2. Rules: formal limited grammar. Rename the contract off "JSON Schema" to a strict RuleValueSchema, fail closed on unsupported keywords/types. No JSON-Schema library unless external compatibility becomes a requirement (B3 scoped accordingly).
3. i18n: freeze catalogs at boot by default. Load + validate during boot, immutable for request handling. Hot-reload is a separate opt-in overlay concern (B13 stays P2, needs demonstrated operational need).
4. auth_time: wire it only if the production IdP emits it reliably. If not, step-up stays AMR-only — do NOT ship a decorative MaxAge (B8 amended; first task is the IdP capability check).
5. sms: removed from the default strong-factor set. Available only as an explicit opt-in policy choice for legacy compatibility (B8 amended).
6. B2 (webhook SSRF client): first post-merge engineering task; the current merge is not blocked on it.

## 6. Recommended implementation order
merge the current branches as-is → B2 first post-merge (Decision 6) → B3 + B4 + B5 in parallel (small correctness gates) → B1 (large; then W1 immediately) → B8 + B10 (auth hardening) → B6 → B7 → B9 → P2 items only on evidence. D1 immediately (doc edit).

## Tradeoffs stated
- Design purity vs rollout: B6 enforcement and B3 fail-closed can break existing lax configs/schemas — both ship advisory-first with a documented enforcement flip.
- B1 is a one-way-door API design; spend review there, not on P2 speculation.
- Keeping ServeMux until B5 data exists is deliberate: safety contracts (RouteMeta, boot validation) outrank raw dispatch speed at current scale.

## Specialist artifacts
design-review-{i18n,scaffold,security,rules,product}.md · benchmark-{internals,external}-review.md · design-review-gate-report.md (same directory).
