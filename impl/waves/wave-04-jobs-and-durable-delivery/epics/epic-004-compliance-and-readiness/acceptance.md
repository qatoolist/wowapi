---
id: W04-E04-ACCEPTANCE
type: epic-acceptance
epic: W04-E04
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern.

## AC-W04-E04-01 — Audit hash chain widened and version-discriminated (D-04 enacted)

The widened `chainHash` covers every persisted field, including canonicalized `metadata` and
`tx_id`; a tamper test mutating each declared field independently confirms every one fails
verification. A `hash_version smallint NOT NULL DEFAULT 1` column exists in the same migration that
widens field coverage; historical rows verify under the v1 branch, new rows verify under the v2
branch (metadata + tx_id included), per D-04's ratified design. Traces to W04-E04-S001.

## AC-W04-E04-02 — Anchor, DSR export, legal-hold, and explicit-status delivered

The audit chain is periodically anchored externally; a test anchoring then tampering the chain
confirms detection even where local `head_hash` were compromised. DSR export completes only after an
encrypted immutable artifact (manifest, per-class results, checksum, expiry, access policy, download
audit) is successfully written, proven by a test confirming export completion is gated on artifact-
write success and checksum verification. A deliberately non-compliant `Dispose`/`Erase` callback is
still blocked by the central legal-hold wrapper, proven by a negative test. The DSR result set
explicitly lists every registered record class with a status, never a silent omission. Traces to
W04-E04-S002.

## AC-W04-E04-03 — Readiness and configuration diagnostics truthful (DX-07 T1-T3)

`/readyz` fails (503) when the applied-migration version lags the expected version, proven by an
integration test booting against a stale-migrated database. Readiness reports migration version,
seed/rule hash, and model hash. `config doctor` discovers the product root via `go env GOMOD`/
`--project` regardless of invocation directory, proven by nested-subdirectory and outside-repo-with-
`--project` test cases, and explicitly reports whether product validation ran. Traces to
W04-E04-S003. DX-07 T4 is explicitly out of scope for this epic — see `epic.md` "Out of scope" and
`risks.md` RISK-W04-004.

## AC-W04-E04-04 — Independent review passed

All three stories (S001, S002, S003) have passed independent review per mandate §14. S001's review
specifically confirms the per-field tamper test genuinely covers every declared field independently
(not a generic tamper test), and that D-04's version-branch design was implemented exactly as
ratified, not a divergent interpretation silently substituted. S002's review specifically confirms
W6-T4's `RecordClass` callback enumeration was genuinely completed before the wrapper landed, and
that W6-T3's DSR export artifact is genuinely gated on write success (not a partial or best-effort
gate). S003's review specifically confirms DX-07 T4 was correctly and explicitly scoped out — no
task silently attempting T4's capacity-enforcement behavior, and no silent dropping of the forward
reference to W05-E03-S002.

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA, extended to DX-07's readiness/diagnostics scope given its shared
deployment-readiness closure-spec relationship with DATA-08 within this wave).
