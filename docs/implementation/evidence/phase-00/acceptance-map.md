# Phase 0 — Acceptance Map

Phase 0 exit criteria (Goal 2 §Phase 0) → evidence:

| Criterion | Status | Evidence |
|---|---|---|
| `make help` works | ✅ | command-log #12 |
| `make up` starts local infra | ⚠️ deferred | compose validated statically (command-log #14); daemon unavailable (R4) — to be demonstrated at first CI/daemon availability, hard requirement before Phase 2 |
| `make lint` and `make test-unit` run | ✅ | command-log #13 (`make ci` includes vet/lint-boundaries/test-unit/test-race/build) |
| Package graph rules encoded | ✅ | scripts/lint_boundaries.sh (import law + vocabulary + Reveal() rules); command-log #9 |
| go.mod / repo structure / Makefile / Dockerfile / compose / lint tooling | ✅ | files in commit; command-log #4–#14 |
| Phase plan + risk register | ✅ | ../../phase-plan.md, ../../risk-register.md |
| Initial Graphify check | ✅ | command-log #3 |
| Preflight items 1–6 (Goal 2 command) | ✅ | decisions.md D-0001…D-0008; ../../readiness-checklist.md; blueprint diffs (04/06/11/12) in commit |

Blueprint acceptance criteria touched this phase (10-delivery §2): #24 partially demonstrated
(graph rules encoded + lint green on current packages); 12 §11.5 partially demonstrated
(Secret redaction test suite: kernel/config/secret_test.go).
