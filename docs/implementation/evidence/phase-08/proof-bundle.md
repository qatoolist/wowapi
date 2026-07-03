# Phase 8 — Proof Bundle

Scope (phase-plan row 8): document/file framework (object-storage port, class registry, presigned
upload sessions, versioned append-only file pointers, authorized presigned downloads, access grants,
scan gate, retention sweep) + comments + attachments, migration 00010. Date: 2026-07-04.

## 1. Decision evidence
D-0055 (Phase 8: document framework design + review fixes — storage Adapter port, append-only
document_versions with app_platform scan/retention, deny-first + owner + capacity-grant download
authorization, ownership-enforced grant RLS via app_actor_id(), column-restricted documents UPDATE,
comment/attachment author guards).

## 2. Discussion evidence
- **Object storage as a port**: the kernel never talks to S3/minio directly. `storage.Adapter`
  (PresignPut/Get, Stat, Peek, Delete) keeps blob bytes off the API process; the memory adapter backs
  tests and proves the port, and a real S3 adapter implements the same five methods (HEAD/ranged-GET/
  DeleteObject). Rejected: streaming bytes through the service (couples the kernel to storage + moves
  large payloads through the API tier).
- **Append-only versions + privilege split**: `document_versions` is INSERT-only to the module role;
  scan-status settlement and retention voiding are app_platform (tenant-bound), mirroring the Phase
  6/7 relay/activate posture — a module can never rewrite a file pointer or clear an infected flag.
- **Grant authorization**: the review reproduced that a Go-only gate on `Grant` is bypassable
  because app_rt holds raw table INSERT (SEC-41/42). Resolved with a RESTRICTIVE RLS policy keyed on
  `app_actor_id()` document ownership — unbypassable at the DB — rather than moving grants to a
  platform path (which would break business-tx composition). The behavior-changing/legal controls
  (legal_hold, status, retention_until) are column-restricted to app_platform (SEC-44).
- **Download is a read**: emitting an outbox event there broke read-only-tx callers (ARCH-65);
  removed. Durable download audit belongs to the audit_logs writer (a later phase).

## 3. Critique/review evidence
`review-findings.md`: 13 reproduced defects — 4 high (SEC-41 grant self-grant, SEC-43 unguarded
revoke, SEC-44 legal-hold/status escalation, ARCH-65 read-only-tx break), 4 med (SEC-42 grant
escalation, SEC-45 comment tamper, ARCH-66 upload-key collision, ARCH-67 error-masking), 5 low. All
fixed with regression tests. Two parallel review agents (security + architecture); RLS isolation,
append-only enforcement, checksum verification, and deny-first evaluation verified solid.

## 4. Implementation evidence
New: `kernel/storage/` (Adapter port + memory adapter), `kernel/document/` (registry, hooks,
service), `kernel/comment/`, `kernel/attachment/`, migration `00010_documents.sql` (+ `app_actor_id()`).
Changed: `kernel/kernel.go` (wire Documents/Comments/Attachments + kernel-owned document permissions),
`module/module.go` + `app/context.go` + `app/boot.go` (accessors + boot gates), `testkit/db.go`
(PlatformTxM tenant-bound app_platform manager), `migrations/migrations_test.go`.
Team: 1 implementation agent (comment + attachment) + lead (storage, document, migration, wiring, all
review fixes); 2 review agents (security, architecture).

## 5. Verification evidence
`command-log.md`: storage + document + comment + attachment integration (upload round-trip, byte
verification, MIME sniff, scan gate, infected-never-serves, access grant, retention sweep, legal-hold
block, tenant isolation) + review regressions (grant RLS, legal-hold column, revoke authz, author
guards, RO-tx download, distinct upload keys, MIME essence); full `make ci` + `make test-integration`
host; `-race -count=5` clean; `make ci-container` (warm) and `go test -p 1 ./...` in-container green.
The parallel-container template/role-provisioning race is documented (deferred testkit hardening).

## 6. Acceptance evidence
`acceptance-map.md`: all 19 Phase 8 exit criteria mapped to named tests. Carried forward: role/
relationship grants via the authz evaluator, durable download audit (audit_logs writer), hard-erasure
redaction/GDPR job, an S3/minio Adapter, and the testkit SQLSTATE-28000 connect-retry. Graphify
`extract` blocked on LLM key (R11).
