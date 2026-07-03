# Phase 8 — Acceptance Map

Phase 8 exit criteria (Goal 2 Phase 8 + phase-plan row 8 + blueprint 07 §4) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Object-storage port (kernel never touches S3 directly) | `kernel/storage/storage.go` `Adapter` (PresignPut/Get, Stat, Peek, Delete); `kernel/storage/memory.go` in-memory adapter; `TestMemory*` |
| 2 | Document class registry (module-owned, policy envelope) | `kernel/document/registry.go` Class{DefaultSensitivity, MaxBytes, AllowedMIME, Retention}; module-prefix + duplicate + sensitivity validation |
| 3 | Document metadata + optional resource anchor | `documents` table; `Service.Create`; `TestIntegrationUploadRoundTrip` |
| 4 | Presigned upload session (initiate → confirm) | `Service.InitiateUpload` (presigned PUT, tenant-prefixed key, version reserved) → `ConfirmUpload` |
| 5 | **Confirm verifies size + checksum + MIME sniff + class limits** | `TestIntegrationConfirmVerifiesBytes` (checksum mismatch rejected), `TestIntegrationMIMEMismatchRejected` (text sniffed vs image/png declared) |
| 6 | Immutable versioned file pointers (append-only) | `document_versions` unique(document_id,version_no); app_rt SELECT+INSERT only (UPDATE = app_platform) |
| 7 | OnFileUpload / OnDocumentAccess hooks | `kernel/document/hooks.go`; runUpload aborts confirm, runAccess denies download |
| 8 | **Authorized download (policy + grants + owner) → presigned GET + audit** | `Service.Download`; `TestIntegrationAccessGrant` (stranger forbidden → granted → allowed); deny-first via authz evaluator |
| 9 | Explicit access grants (beyond policy, time-boxed) | `document_access_grants`; `Service.Grant`/`Revoke`; capacity-grant validity window enforced |
| 10 | **Scan gate: infected never serves; pending blocks confidential+** | `TestIntegrationScanGateBlocksConfidential`, `TestIntegrationInfectedNeverServes`; `Service.UpdateScanStatus` (app_platform) |
| 11 | **Retention sweep voids expired versions (legal hold blocks)** | `Service.SweepRetention` (app_platform); `TestIntegrationRetentionSweep`, `TestIntegrationLegalHoldBlocksSweep` |
| 12 | Deletion = voiding, not hard erase (void ≠ delete) | version/document status='voided' + voided_at tombstone; blob deleted, row retained |
| 13 | Comments (threaded, edit-with-history, void) | `kernel/comment/comment.go`; `TestCreateAndList`, `TestEditChangesBodyStatusVersion`, `TestParentMismatchRejected`, `TestVoidThenEditConflict` |
| 14 | Attachments (document_version ↔ resource/comment/task) | `kernel/attachment/attachment.go`; `TestAttachAndList`, `TestAttachBogusDocumentVersionError`, `TestDetachVoids` |
| 15 | **Tenant isolation across all five tables** | strict RLS `tenant_id = app_tenant_id()`; `TestIntegrationTenantIsolation` (document), comment/attachment `TestTenantIsolation` |
| 16 | Privilege boundary (append-only to module; scan/retention platform) | migration grants: app_rt no UPDATE/DELETE on document_versions; app_platform SELECT+INSERT+UPDATE; retention UPDATE on documents to app_platform |
| 17 | Module.Context accessors wired + boot gates | DocumentClasses/DocumentHooks/Documents/Comments/Attachments on `module.Context`; boot gates `DocumentClasses().Err()` + fails when a class is registered without a storage adapter |
| 18 | Container-first verification | host `make ci` + `make test-integration`; `make ci-container` green |
| 19 | Evidence bundle + parallel review | this directory; review-findings.md (security + architecture agents) |

Carried forward: role/relationship document grants are persisted but evaluated through the authz
evaluator (the direct-grant fast path covers capacity grants); a durable audit_logs writer for
download events (currently outbox events); a hard-erasure redaction/GDPR job (blueprint 07 §4 — void
is implemented, redaction deferred); an S3/minio Adapter implementation (memory adapter ships; the
port is proven). Graphify `extract` blocked on LLM key (R11).
