# wowapi — Test Suite Design & Implemented Tests

## Design principles (applied)

1. **Non-duplication.** New tests only where the existing 477-test suite has a real behavioral gap
   (verified via `go tool cover -func` 0%-funcs cross-checked against existing test names). G6
   (idempotency) and G7 (kernel.New) were dropped because they are already covered indirectly.
2. **Behavior over structure.** Tests assert observable behavior (a guard rejects, a task advances, an
   edge becomes visible), not private field names or formatting. The two white-box tests
   (`parseISODuration`, benchbudget parsers) target meaningful unexported *logic*, not internal shape.
3. **Reuse the existing harness.** New integration tests use `testkit.NewDB`, the shared fixtures,
   `buildRuntime`/`WorkflowSim`, and the relationship seed helpers — no new scaffolding, same
   terminology.
4. **Real integration, mock only true externals.** Every DB test runs against a real Postgres via the
   testkit template; no internal component is mocked. (There are no true external systems in these gaps
   — object storage already has an in-process port + memory adapter used by the document suite.)
5. **Negative + security first.** Each new area includes the failure/deny paths (superuser rejected,
   app_rt denied, non-task step refused, empty-reason override, malformed durations, bad budget files).

## Implemented tests (8 files, 25 test functions)

| File | Pkg | Tests | What it pins |
|---|---|---|---|
| `kernel/database/rls_guard_test.go` | database_test | `TestConnRLSGuardRejectsOverPrivilegedConnection`, `TestConnRLSGuardAdmitsNonPrivilegedRole` | The fail-closed RLS pool guard: a superuser/BYPASSRLS connection is rejected (RLS would be defeated); a SET ROLE app_rt connection is admitted and genuinely RLS-enforced. **Security.** |
| `kernel/workflow/runtime_lifecycle_test.go` | workflow_test | `TestIntegrationWorkflowCompleteTask`, `…CompleteTaskRejectsNonTaskStep`, `…Delegate`, `…Override`, `…GatewayRouting` | Task-step completion (+ output merge, re-complete conflict, non-task-step refusal); delegation (open task + delegate assignee + event + delegate can approve); privileged Override (empty-reason/unknown-target/non-running negatives + task-skip + event); gateway context routing. **Workflow correctness.** |
| `kernel/workflow/sla_parse_test.go` | workflow (internal) | `TestParseISODurationValid`, `TestParseISODurationInvalid` | ISO-8601 SLA duration parsing (W/D/H/M/S, composite) + malformed inputs. **Parsing / edge.** |
| `kernel/relationship/relationship_relate_test.go` | relationship_test | `TestIntegrationRelateAsPlatformThenHas`, `…RelateAsAppRtDenied`, `…RelateTenantIsolation` | The ReBAC edge write path: correct platform-role write visible to Has; app_rt denied (SEC-24); tenant isolation. **Data integrity + security.** |
| `kernel/resource/registry_test.go` | resource_test | `TestRegistryAcceptsWellFormedType`, `…RejectsMalformedKey`, `…RejectsForeignModulePrefix`, `…RejectsDuplicate`, `…AccumulatesAllErrors`, `TestValidTypeKeyAndRefIsZero` | The resource-type registration contract (key shape, module-prefix ownership, duplicates, error accumulation) + `ValidTypeKey`/`Ref.IsZero`. **Framework contract.** |
| `internal/tools/benchbudget/main_test.go` | main | `TestBaseNameStripsGomaxprocsSuffix`, `TestLoadBudgetsValid`, `TestLoadBudgetsRejectsMalformed`, `TestParseBenchOutput` | The CI perf-gate's parsing/thresholding (suffix strip, budget file valid/malformed, bench-output parse + worst-of-duplicates). **Tooling reliability.** |
| `kernel/document/hooks_fire_test.go` | document_test | `TestIntegrationUploadHookAbortsConfirm`, `TestIntegrationAccessHookDeniesDownload` | The document extension points fire: an OnFileUpload hook aborts the confirm (no version row committed); an OnDocumentAccess hook denies the download. **Security extension point.** |
| `kernel/jobs/enqueue_global_test.go` | jobs_test | `TestIntegrationEnqueueGlobalInsertsTenantlessJob`, `TestEnqueueGlobalRejectsInvalidJob` | The tenant-less (global) job enqueue path: a NULL-tenant row is written; nil/empty-kind jobs are rejected. **Data path.** |

## Traceability

Each test file carries a header comment naming its QA gap ID (G1–G8) and risk class. Gap → test →
result is tracked in `03-coverage-matrix.md` and `05-execution-report.md`. Finding D1 (relationship
write is platform-only, previously unused/untested) is recorded in `06-gaps-and-fix-plan.md`.
