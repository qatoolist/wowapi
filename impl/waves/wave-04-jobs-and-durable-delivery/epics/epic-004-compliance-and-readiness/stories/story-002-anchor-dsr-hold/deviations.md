---
id: DEV-W04-E04-S002
type: deviations-record
parent_story: W04-E04-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W04-E04-S002

Per mandate §8.9/§2.6, this section records bounded design decisions that resolve `plan.md`'s
"Unresolved questions" and any implementation divergences.

## D-W04-E04-S002-001 — External anchor protocol

**Decision:** Use a local file-backed external anchor store (`audit.FileStore`) for the
implementation and tests. The anchor record is a JSON file per tenant containing the anchored seq,
head hash, and timestamp.

**Rationale:** `plan.md` left the external-anchoring vendor/protocol as an open design decision.
A public timestamping service or third-party notarization service would require new credentials,
SLA, and operational runtime configuration that this story does not have authority to procure.
The `ExternalStore` interface isolates the protocol choice; a file-backed implementation satisfies
AC-01 (anchor-then-tamper detection) and can be replaced by a remote store without changing the
consumer contract. `kernel.New` does not yet wire a scheduled external-anchor job; the interface
and `AnchorNow`/`Verify` methods are the deliverable.

## D-W04-E04-S002-002 — DSR export artifact encryption-key management

**Decision:** The production artifact writer reads a 32-byte AES key from the
`WOWAPI_DSR_ARTIFACT_KEY` environment variable (hex). When the variable is absent, `kernel.New`
falls back to the deterministic `retention.TestKey()` and logs a warning so local/test boots
succeed. Deployments must set the variable to avoid a shared fallback key.

**Rationale:** `plan.md` left the key-management scheme open. A full KMS integration would require
new infrastructure and config schema changes beyond this story's scope. The bounded decision is
(1) env-var key sourcing, (2) a clear production warning, and (3) an exported `TestKey()` for
tests. Future work can swap `NewFileArtifactWriter` for a KMS-backed `ArtifactWriter` without
changing `Engine`.

## D-W04-E04-S002-003 — Hold identifiers for non-UUID entities

**Decision:** Record-class keys and DSR subject refs are mapped to deterministic UUIDs via
`uuid.NewSHA1(uuid.NameSpaceOID, []byte("wowapi:hold:"+s))` before calling `Holds.IsHeld`. This
preserves the existing `Holds` contract (which requires a `uuid.UUID` entity id) while allowing
holds on string identifiers.

**Rationale:** AC-03 calls for checking `Holds.IsHeld(ctx, db, "record_class", c.Key)` and
`"dsr_subject", subjectRef`, but `Holds.IsHeld`'s signature requires a `uuid.UUID`. Rather than
widening the `Holds` API (which would be a larger, cross-cutting change), the Engine derives a
stable UUID from the string. This decision is documented so future consumers know the mapping.

## D-W04-E04-S002-004 — Artifact checksum covers the encrypted payload

**Decision:** The artifact checksum is `SHA256(ciphertext)`, stored in the envelope and returned in
`ArtifactManifest.Checksum`. It therefore verifies the bytes actually written to storage, not the
plaintext manifest.

**Rationale:** Embedding a self-describing checksum inside the encrypted JSON creates a circular
fixed-point. Using the ciphertext checksum is straightforward, auditable, and satisfies AC-02's
"checksum verifies against the written artifact" requirement.

## RecordClass enumeration

- **wowapi:** `grep` for `retention.NewRegistry().Register` and `Register(retention.RecordClass{...})`
  found only framework/test registrations; zero product modules register a `RecordClass` today.
- **wowsociety:** per this epic's `dependencies.md`, no `kernel/attachment`/`kernel/notify`/
  `kernel/retention` usage was found at plan time, so wowsociety has no registered classes today.

This enumeration predates the legal-hold wrapper implementation and is recorded here per PLAN
W6-T4's precondition.
