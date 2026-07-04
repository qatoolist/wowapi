# miscellaneous/ — reusable internal check scripts

Reusable, read-only project checks used during development and the review gate. **All internal
check/audit/validation scripts live here** — do not scatter one-off scripts across the repo (policy:
[../docs/working/internal-scripts-policy.md](../docs/working/internal-scripts-policy.md)). Each script is
read-only (never modifies project files), self-locating (runs from anywhere), and documents what it checks
in its header.

Run the whole mechanical pass with `miscellaneous/review_gate.sh` (add `--full` to also run
`make ci-container`).

| Script | Checks | When to run |
|---|---|---|
| `review_gate.sh` | Aggregates all checks below + gofmt/vet/boundary lint + stray artifacts; `--full` adds the authoritative gate | Every review gate, before declaring done |
| `check_migrations.sh` | Migrations registered in `migrations_test.go`, have Up+Down, contiguous numbering | After adding/editing a migration |
| `check_test_skips.sh` | Lists `t.Skip` sites (green-but-hollow guard) | Review gate; when touching integration tests |
| `find_duplicate_tests.sh` | Duplicate Test/Benchmark/Fuzz function names | Before adding tests; review gate |
| `check_unwired.sh [pkg]` | Exported constructors/services with no caller outside their package (built-but-not-wired candidates) | After adding a kernel service; review gate |
| `check_overclaims.sh` | Doc lines claiming "complete/done" next to a deferral word | Before updating decisions/CHANGELOG; review gate |

Grounding: each script targets a real recurring risk from
[../docs/working/review-learning-register.md](../docs/working/review-learning-register.md) (unregistered
migrations, hollow coverage, unwired primitives, overclaims). Exit codes: 0 = clean, 1 = issues (advisory
scripts always exit 0 and just print).

To add a script: follow the policy — clear name, reusable, documented header (what/when/usage), read-only,
no hardcoded temp assumptions, and add a row to this table.
