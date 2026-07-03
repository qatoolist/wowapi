# Phase 8 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-04.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `make migrate` (00010 on fresh schema) | 0 | migration 10 applied: documents, document_versions (append-only), document_access_grants, comments, attachments + strict tenant RLS + grants |
| 2 | `go build ./kernel/document/ ./kernel/storage/ ./kernel/comment/ ./kernel/attachment/` | 0 | four new packages compile |
| 3 | `DATABASE_URL=… go test ./kernel/storage/ ./kernel/document/ ./kernel/comment/ ./kernel/attachment/` | 0 | storage memory adapter; document upload round-trip, byte verification, MIME mismatch, scan gate (confidential blocked→clean), infected never serves, access grant, retention sweep, legal-hold blocks sweep, tenant isolation; comment create/list/parent-mismatch/edit-CAS/void; attachment attach/list/bogus-fk/detach/isolation |
| 4 | `go build ./...` | 0 | full module builds after wiring document/comment/attachment into kernel + module.Context + app boot |
| 5 | `make ci` (host) | 0 | vet, boundary lint, unit, race, build green |
| 6 | `sh scripts/lint_boundaries.sh` | 0 | OK — document/storage/comment/attachment import kernel/* only; domain-neutral (no leaked domain terms) |
| 7 | `go test -race -count=5 ./kernel/document/ ./kernel/storage/ ./kernel/comment/ ./kernel/attachment/` | 0 | no races; stable across 5 runs |
| 8 | `make test-integration` (host) | 0 | all integration green |
| 9 | `make ci-container` | 0 | green in the tools container (one transient template-clone contention retry — ARCH-21 — then green) |
| 10 | (review pass) `go test ./kernel/document/ ./kernel/comment/ ./kernel/attachment/ ./kernel/storage/` after SEC-41…48 + ARCH-65…69 fixes | 0 | + regressions: grant RLS blocks non-owner (SEC-41), legal_hold column protected (SEC-44), revoke requires write (SEC-43), non-author edit forbidden (SEC-45), non-creator detach forbidden (SEC-46), RO-tx download (ARCH-65), distinct upload keys (ARCH-66), MIME essence (ARCH-69) |
| 11 | `make ci` (host, post-fix) | 0 | vet, boundary lint, unit, race, build green |
| 12 | `docker compose run --rm tools go test -p 1 ./...` | 0 | full suite green in-container, SERIAL (proves Phase 8 correct; isolates the parallel template/role-provisioning race) |
| 13 | `make ci-container` (warm template) | 0 | parallel container CI green once the migration-hash-changed template is built. NOTE: the first parallel run after a migration change can flake `FATAL: role "app_rt" is not permitted to log in` — a testkit template/role-provisioning race (not Phase 8 code; it hit unrelated packages: relationship, rules, workflow). Deferred testkit hardening: retry pool connect on SQLSTATE 28000. |
