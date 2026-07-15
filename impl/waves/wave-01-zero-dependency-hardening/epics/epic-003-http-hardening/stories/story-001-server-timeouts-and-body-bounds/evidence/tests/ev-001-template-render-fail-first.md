# EV-W01-E03-S001-001 — template-render assertion (fail-first pair)

- **Evidence ID**: EV-W01-E03-S001-001
- **Evidence type**: unit-test report (fail-first pair)
- **Story / task**: W01-E03-S001 / W01-E03-S001-T001
- **Acceptance criteria proven**: AC-W01-E03-S001-02
- **Execution command**: `go test ./internal/cli/ -run 'TestInitAPIMainConfiguresAllServerTimeouts|TestInitConfigsBaseDocumentsServerTimeouts' -count=1 -v`
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: FAILED pre-fix (template lacked ReadTimeout/WriteTimeout/IdleTimeout and yaml lacked the four keys), PASSED post-fix. Failed run preserved below per the failed-evidence preservation rule (status of the failing half: `resolved` — superseded by the passing half of this same record pair).

## Pre-fix run (status: failed → resolved) — captured at 0a31186cada5c275a588c74081cf977adf346e61 before any template/config change

```
=== RUN   TestInitAPIMainConfiguresAllServerTimeouts
    scaffold_test.go:526: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitAPIMainConfiguresAllServerTimeouts3932833730/001/cmd/api/main.go: expected to match "ReadTimeout:\\s+cfg\\.HTTP\\.ReadTimeout"
    scaffold_test.go:526: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitAPIMainConfiguresAllServerTimeouts3932833730/001/cmd/api/main.go: expected to match "WriteTimeout:\\s+cfg\\.HTTP\\.WriteTimeout"
    scaffold_test.go:526: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitAPIMainConfiguresAllServerTimeouts3932833730/001/cmd/api/main.go: expected to match "IdleTimeout:\\s+cfg\\.HTTP\\.IdleTimeout"
--- FAIL: TestInitAPIMainConfiguresAllServerTimeouts (0.01s)
=== RUN   TestInitConfigsBaseDocumentsServerTimeouts
    scaffold_test.go:547: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitConfigsBaseDocumentsServerTimeouts138484261/001/configs/base.yaml: expected to contain "read_header_timeout: 10s"
    scaffold_test.go:547: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitConfigsBaseDocumentsServerTimeouts138484261/001/configs/base.yaml: expected to contain "read_timeout: 30s"
    scaffold_test.go:547: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitConfigsBaseDocumentsServerTimeouts138484261/001/configs/base.yaml: expected to contain "write_timeout: 60s"
    scaffold_test.go:547: /var/folders/7v/2bvxj5q50kl8fljg5qm0j8_h0000gp/T/TestInitConfigsBaseDocumentsServerTimeouts138484261/001/configs/base.yaml: expected to contain "idle_timeout: 120s"
--- FAIL: TestInitConfigsBaseDocumentsServerTimeouts (0.01s)
FAIL
FAIL	github.com/qatoolist/wowapi/internal/cli	0.245s
FAIL
```

## Post-fix run (status: passed)

```
=== RUN   TestInitAPIMainConfiguresAllServerTimeouts
--- PASS: TestInitAPIMainConfiguresAllServerTimeouts (0.01s)
=== RUN   TestInitConfigsBaseDocumentsServerTimeouts
--- PASS: TestInitConfigsBaseDocumentsServerTimeouts (0.01s)
ok  	github.com/qatoolist/wowapi/internal/cli	0.271s
```
