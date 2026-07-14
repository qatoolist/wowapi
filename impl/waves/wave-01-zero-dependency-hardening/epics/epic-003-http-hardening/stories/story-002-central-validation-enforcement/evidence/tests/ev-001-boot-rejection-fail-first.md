# EV-W01-E03-S002-001 — boot-rejection fail-first pair

- **Evidence ID**: EV-W01-E03-S002-001
- **Evidence type**: unit-test report (fail-first pair)
- **Story / task**: W01-E03-S002 / W01-E03-S002-T001
- **Acceptance criteria proven**: AC-W01-E03-S002-01, AC-W01-E03-S002-02
- **Execution command**: `go test ./kernel/httpx/ -run 'RequestContract|ValidatedHandler|ContractWithWaiver|MutatingRouteWithoutContract' -count=1 -v`
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: three-stage proof. (1) At pristine 0a31186cada5c275a588c74081cf977adf346e61, `TestRouterMutatingRouteWithoutContractBootsByDefault` PASSED — a POST route with no declared contract registers cleanly today, i.e. no framework safety net exists (AC-01, the defect is real). (2) With the RouteMeta fields + RequireRequestContracts mode added but the Handle-time check deliberately stubbed, the rejection and contradiction tests FAILED (red run below). (3) After implementing `Router.checkRequestContract`, all tests PASS, including the flag-off compat test — same fixture still boots by default (RISK-W01-002).

## Stage 1 — pristine HEAD: undeclared POST route boots (AC-01)

```
=== RUN   TestRouterMutatingRouteWithoutContractBootsByDefault
--- PASS: TestRouterMutatingRouteWithoutContractBootsByDefault (0.00s)
ok  	github.com/qatoolist/wowapi/kernel/httpx	1.391s
```

## Stage 2 — enforcement stubbed (status: failed → resolved)

```
=== RUN   TestRouterMutatingRouteWithoutContractBootsByDefault
--- PASS: TestRouterMutatingRouteWithoutContractBootsByDefault (0.00s)
=== RUN   TestRouterRequireRequestContractsRejectsUndeclaredMutatingRoute
--- FAIL: TestRouterRequireRequestContractsRejectsUndeclaredMutatingRoute (0.00s)
=== RUN   TestRouterRequireRequestContractsAllowsDeclaredContract
--- PASS: TestRouterRequireRequestContractsAllowsDeclaredContract (0.00s)
=== RUN   TestRouterRequireRequestContractsWaiverExemptsBodylessMutation
--- PASS: TestRouterRequireRequestContractsWaiverExemptsBodylessMutation (0.00s)
=== RUN   TestRouterRejectsContractWithWaiverContradiction
--- FAIL: TestRouterRejectsContractWithWaiverContradiction (0.00s)
=== RUN   TestRouterRequireRequestContractsIgnoresNonMutatingMethods
--- PASS: TestRouterRequireRequestContractsIgnoresNonMutatingMethods (0.00s)
=== RUN   TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors
--- FAIL: TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors (0.00s)
=== RUN   TestValidatedHandlerPassesValidDTOToBusinessLogic
--- FAIL: TestValidatedHandlerPassesValidDTOToBusinessLogic (0.00s)
FAIL
FAIL	github.com/qatoolist/wowapi/kernel/httpx	0.578s
FAIL
```

## Stage 3 — post-implementation (status: passed)

```
=== RUN   TestRouterMutatingRouteWithoutContractBootsByDefault
--- PASS: TestRouterMutatingRouteWithoutContractBootsByDefault (0.00s)
=== RUN   TestRouterRequireRequestContractsRejectsUndeclaredMutatingRoute
--- PASS: TestRouterRequireRequestContractsRejectsUndeclaredMutatingRoute (0.00s)
=== RUN   TestRouterRequireRequestContractsAllowsDeclaredContract
--- PASS: TestRouterRequireRequestContractsAllowsDeclaredContract (0.00s)
=== RUN   TestRouterRequireRequestContractsWaiverExemptsBodylessMutation
--- PASS: TestRouterRequireRequestContractsWaiverExemptsBodylessMutation (0.00s)
=== RUN   TestRouterRejectsContractWithWaiverContradiction
--- PASS: TestRouterRejectsContractWithWaiverContradiction (0.00s)
=== RUN   TestRouterRequireRequestContractsIgnoresNonMutatingMethods
--- PASS: TestRouterRequireRequestContractsIgnoresNonMutatingMethods (0.00s)
=== RUN   TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors
--- PASS: TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors (0.00s)
=== RUN   TestValidatedHandlerPassesValidDTOToBusinessLogic
--- PASS: TestValidatedHandlerPassesValidDTOToBusinessLogic (0.00s)
ok  	github.com/qatoolist/wowapi/kernel/httpx	0.247s
```
