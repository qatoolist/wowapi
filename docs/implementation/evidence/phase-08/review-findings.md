# Phase 8 — Review Findings

Two parallel critique agents audited the documents / storage / comments / attachments slice on
2026-07-04 with live DB probes: one SECURITY-focused (privilege boundaries, authz bypass, RLS), one
ARCHITECTURE-focused (correctness, edge cases, same-tx). They reproduced 13 defects — 4 high, 4
medium, 5 low — all fixed with regression tests. The tenant-isolation RLS, append-only enforcement,
checksum verification, and deny-first download evaluation were verified solid.

## Security findings

| ID | Sev | Finding (reproduced) | Resolution | Status |
|---|---|---|---|---|
| SEC-41 | high | app_rt could INSERT a `document_access_grants` row for a document it does NOT own, bypassing the Go `Grant` authz — self-grant on any document in the tenant | RESTRICTIVE RLS policy `document_access_grants_owner_write` pins every INSERT/UPDATE to a document whose `created_by = app_actor_id()`; new `app_actor_id()` SQL function; `TestIntegrationGrantRLSBlocksNonOwner` | **fixed** |
| SEC-42 | high | app_rt could UPDATE a grant to escalate read→write or redirect it to another document | same RESTRICTIVE policy (WITH CHECK on the new row's document ownership) | **fixed** |
| SEC-43 | high | `Revoke` had NO authorization — any tenant actor could revoke any grant | `Revoke` loads the grant's document + owner and runs `authorizeWrite` before the update; the ownership RLS is the DB backstop; `TestIntegrationRevokeRequiresWrite` | **fixed** |
| SEC-44 | high | app_rt could UPDATE `documents.legal_hold` / `status` directly — clear a legal hold or void a document to dodge retention | column-level grant: `GRANT UPDATE (title, sensitivity, version, updated_at, updated_by) ON documents TO app_rt`; `status`/`legal_hold`/`retention_until` are app_platform-only; `TestIntegrationLegalHoldColumnProtected` | **fixed** |
| SEC-45 | med | `comment.Edit`/`Void` had no author check — any actor could tamper with any comment | `ensureAuthor` gate (context actor must be the authoring capacity or creating actor; fail-closed on no actor); `TestEditByNonAuthorForbidden` | **fixed** |
| SEC-46 | med | `attachment.Detach` had no authz — any actor could void any attachment | creator-only guard (context actor must equal `created_by`); `TestDetachByNonCreatorForbidden` | **fixed** |
| SEC-47 | low | `Download` explicit-VersionNo path revealed version void/existence via differing errors BEFORE authorization | authorize immediately after loading the document (before any version query); both version paths now filter `status='active'` so voided ≡ missing | **fixed** |
| SEC-48 | low | `SweepRetention` deleted the blob before voiding the row → a mid-sweep failure left an active row pointing at a deleted blob | void rows in the tx; delete blobs only AFTER the tx commits (orphan-on-delete-failure is safe, the reverse is not) | **fixed** |

## Architecture findings

| ID | Sev | Finding (reproduced) | Resolution | Status |
|---|---|---|---|---|
| ARCH-65 | high | `Download` emitted an outbox INSERT, so it FAILED inside a `WithTenantRO` tx — the natural home for a "get download URL" GET handler | Download no longer emits an event (it is a pure read); durable download audit deferred to the audit_logs writer; `TestIntegrationDownloadInReadOnlyTx` | **fixed** |
| ARCH-66 | med | concurrent `InitiateUpload` for one document produced the SAME version-derived storage key → the two PUTs clobbered each other, and a legitimate upload could get a spurious checksum-mismatch | storage key uses a random UUID suffix, not the version number; the version_no race is still resolved cleanly by the unique constraint at confirm; `TestIntegrationDistinctUploadKeys` | **fixed** |
| ARCH-67 | med | `comment.Edit` + parent-existence check returned `KindNotFound` for ANY DB error (masking real failures) | split `errors.Is(pgx.ErrNoRows)` (→ NotFound) from `Wrapf` (→ real error), matching the service.go idiom | **fixed** |
| ARCH-68 | low | `attachment.Attach` allowed attaching to a VOIDED document version (link to a destroyed blob) | existence pre-check adds `AND status = 'active'`; `TestAttachBogusDocumentVersionError` still holds | **fixed** |
| ARCH-69 | low | `mimeConflict` compared only the top-level type, so text/html declared as text/plain passed | compare the full type/subtype (`essence`, parameters stripped); `TestIntegrationMIMEEssenceMismatch` | **fixed** |

Reviewer-verified solid (positive): strict tenant RLS on all five tables (probed cross-tenant);
`document_versions` genuinely append-only to app_rt (UPDATE/DELETE `permission denied`); checksum
verified against store-computed SHA-256 (not client-attestable); deny-first download (explicit deny
policy overrides owner+grant); write-grant satisfies read but read-grant does NOT satisfy write;
grant validity window enforced; infected never serves at any sensitivity; pending blocks
confidential+ only; `UpdateScanStatus` tenant-bound; unique(document_id, version_no) resolves the
upload race without corruption; `hasCapacityGrant` IN-clause injection-safe; the storage Adapter port
is implementable on any S3-compatible SDK.

Residual / carried forward (honest):
- The MED comment/attachment guards (SEC-45/46) are Go-level (clean Forbidden for the realistic
  user-vs-user case). A module that issues raw SQL as app_rt can still tamper with its OWN tenant's
  comments/attachments — accepted: the module is trusted in-process code; the DB-level, unbypassable
  protections are reserved for the cross-authorization and legal/retention controls (SEC-41–44).
- Role/relationship document grants are persisted but the direct-grant fast path evaluates only
  capacity grants (role/relationship access flows through the authz evaluator).
- Testkit robustness (R-testkit): a fresh template rebuild after a migration change races the global
  `app_rt`/`app_platform` LOGIN provisioning against concurrent package pool connections
  (`FATAL: role "app_rt" is not permitted to log in`). Deterministic once the template is warm;
  `go test -p 1` is race-free. A connection-retry on SQLSTATE 28000 (mirroring the existing
  "being accessed by other users" retry) is the fix — deferred as a cross-cutting testkit item.
- Hard-erasure redaction/GDPR job and an S3/minio Adapter implementation remain future work (void is
  implemented; the port is proven by the memory adapter).
