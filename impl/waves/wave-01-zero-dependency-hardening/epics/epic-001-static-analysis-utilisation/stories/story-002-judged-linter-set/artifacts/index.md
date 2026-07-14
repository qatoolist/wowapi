---
id: W01-E01-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E01-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries produced 2026-07-13 (W01Lint).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E01-S002-001 | Updated `.golangci.yml` (judged block) | configuration example | implementation | Enables gosec, errorlint, exhaustive, forcetypeassert, usestdlibvars | FBL-07 | W01-E01-S002-T007 (final enablement step) | `.golangci.yml` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-002 | `kernel/auth/jwks.go` (G704 annotation) | source-code change | implementation | `#nosec` justification comment at lines 204, 210, referencing SEC-06 | FBL-07 | W01-E01-S002-T001 | `kernel/auth/jwks.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-003 | G115 site fixes/annotations | source-code change | implementation | Per-site annotation or bounds check across audit/database/jobs/mfa/pagination packages | FBL-07 | W01-E01-S002-T002 | audit.go, database.go, jobs.go, totp.go annotated; cursor.go FIXED (bounds checks + test) | produced 2026-07-13 |
| ART-W01-E01-S002-004 | G304 site annotation (buildinfo file read) | source-code change | implementation | Tool-only/low-risk annotation | FBL-07 | W01-E01-S002-T003 | `internal/buildinfo/buildinfo.go` (+3 more G304-class sites: openapi_cmd.go, benchbudget/main.go, config/tree.go) | produced 2026-07-13 |
| ART-W01-E01-S002-005 | `kernel/httpx/middleware.go` (errorlint fix) | source-code change | implementation | `errors.Is` in place of `==` against `http.ErrAbortHandler` at line 54 | FBL-07 | W01-E01-S002-T004 | `kernel/httpx/middleware.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-006 | `kernel/workflow/definition.go` (exhaustive annotation) | source-code change | implementation | Suppression annotation preserving fail-closed `default:` arm at line 313 | FBL-07 | W01-E01-S002-T005 | `kernel/workflow/definition.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-007 | `kernel/workflow/runtime.go` (exhaustive annotation) | source-code change | implementation | Suppression annotation preserving fail-closed `default:` arm at line 170 | FBL-07 | W01-E01-S002-T005 | `kernel/workflow/runtime.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-008 | `kernel/auth/jwks.go` (forcetypeassert fix) | source-code change | implementation | Checked (comma-ok) type assertion at line 112 | FBL-07 | W01-E01-S002-T006 | `kernel/auth/jwks.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-009 | `kernel/config/bind.go` (forcetypeassert fix) | source-code change | implementation | Checked (comma-ok) type assertion at line 150 | FBL-07 | W01-E01-S002-T006 | `kernel/config/bind.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-010 | usestdlibvars site fixes | source-code change | implementation | Literal-to-stdlib-constant replacements across whatever sites the fresh run enumerates | FBL-07 | W01-E01-S002-T007 | 9 sites: document/service_test.go, httpx (composite_auth, csrf_internal, ratelimit tests), storage/memory_test.go | produced 2026-07-13 |
| ART-W01-E01-S002-011 | `kernel/policy/policy.go` (nilerr annotation) | source-code change | implementation | Fail-closed-intent explanation comment at line 166, no logic change | FBL-07 | W01-E01-S002-T007 | `kernel/policy/policy.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S002-012 | Judged-linter-set triage record | design document | implementation | Per-hit disposition record for every gosec/errorlint/exhaustive/forcetypeassert/usestdlibvars hit at the execution commit, plus the nilerr non-finding and wrapcheck/revive rejection | FBL-07 | All tasks (aggregated in `implementation.md`) | `implementation.md` (this story) | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
