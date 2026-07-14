# EV-W01-E03-S002-002 — adversarial invalid-DTO 400 with field errors

- **Evidence ID**: EV-W01-E03-S002-002
- **Evidence type**: unit-test report (adversarial, fail-first)
- **Story / task**: W01-E03-S002 / W01-E03-S002-T002
- **Acceptance criteria proven**: AC-W01-E03-S002-03
- **Execution command**: `go test ./kernel/httpx/ -run 'ValidatedHandler' -count=1 -v` (part of the ev-001 command's runs)
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: `TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors` posts `{"name":""}` (violating `validate:"required"`) to a route registered through `httpx.ValidatedHandler` with `RouteMeta.Request: createWidgetRequest{}` and enforcement ON: asserts HTTP 400, non-empty `errors` array with `field == "name"` (the existing KindValidation problem-details shape), and that the business handler never ran. Fail-first: with the adaptor stubbed to skip validation (exactly the defect class FBL-08 closes), this test was RED (stage 2 log in EV-W01-E03-S002-001); it went green only when BindAndValidate was actually wired. `TestValidatedHandlerPassesValidDTOToBusinessLogic` covers the happy path (also red under the stub — the stub passed a zero value, proving the test detects a non-binding adaptor).
- **Log**: see EV-W01-E03-S002-001 stages 2–3 (same runs).
