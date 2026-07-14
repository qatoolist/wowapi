---
id: DEV-INDEX-W00-E01-S003
type: deviations-index
parent_story: W00-E01-S003
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations — W00-E01-S003

*Per mandate §8.9.* Story executed 2026-07-13 at `0a31186cada5c275a588c74081cf977adf346e61`.

**No deviations.** The verification runs followed `plan.md`'s approach; the plan was not rewritten.
Two execution notes that are within-plan choices, recorded here for full transparency (neither
alters the approved plan, so neither is registered as a deviation):

1. **Environment choice within the plan's stated options.** Test suites ran host-side (`go test`
   with `WOWAPI_REQUIRE_DB=1`/`WOWAPI_REQUIRE_S3=1`, `DATABASE_URL`/`S3_TEST_ENDPOINT` pointing at
   the compose services on localhost) rather than inside `make ci-container`. `plan.md` "Testing
   strategy", both task definitions, and the ACs themselves name this alternative explicitly
   ("or `docker compose` + `go test` with `WOWAPI_REQUIRE_S3=1`"). The required-env posture
   (fail-not-skip) is identical to the container gate's.
2. **Infrastructure lifecycle.** The compose Postgres/MinIO stack was already up and healthy
   before this story started (shared with concurrent sibling W00 workers), so `make up` was not
   re-run and the stack was deliberately **not** torn down afterwards — tearing it down would have
   disrupted sibling workers mid-run (coordinated with W00E02S001 over IRC). The plan's
   implementation-sequence step 1 ("confirm testkit Postgres and MinIO are both available") was
   satisfied by direct health confirmation (`docker compose ps`: both `healthy`).
