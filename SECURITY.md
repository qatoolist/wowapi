# Security Policy

wowapi is a security-sensitive backend framework (multi-tenant isolation, authorization, audit,
secret handling). We take vulnerabilities seriously and appreciate responsible disclosure.

## Reporting a vulnerability

**Do not open a public issue for security problems.**

Report privately via GitHub Security Advisories:
<https://github.com/qatoolist/wowapi/security/advisories/new>

Please include:
- affected version(s) and component (e.g. `kernel/authz`, RLS, idempotency, outbox),
- a description and, ideally, a minimal reproduction,
- the impact you foresee (tenant crossing, authz bypass, secret disclosure, etc.).

### Response targets
| Stage | Target |
|---|---|
| Acknowledgement | within 3 business days |
| Triage + severity assessment | within 7 business days |
| Fix or mitigation plan | depends on severity; critical issues prioritized immediately |

We will coordinate a disclosure timeline with you and credit reporters unless you prefer to remain anonymous.

## Supported versions

wowapi is pre-1.0 (`v0.x`). Until `v1.0.0`, only the **latest minor** receives security fixes.

| Version | Supported |
|---|---|
| latest `v0.x` minor | ✅ |
| older `v0.x` | ❌ (upgrade to latest) |

After `v1.0.0`, this table will list the supported release lines.

## Security properties & assurance

The framework's security controls are structurally enforced (types, middleware, DB) **and** tested. Key
invariants (deny-by-default authz, fail-closed RLS, append-only tamper-evident audit, structural secret
redaction) are described in [`docs/SRS.md`](docs/SRS.md) §5.1 and exercised by the test suite behind the CI gate.

### Operational runbooks
- **Rotating a per-tenant integration-provider credential** (zero-downtime, secretref-only):
  [`docs/operations/integration-credential-rotation.md`](docs/operations/integration-credential-rotation.md).

### Supply-chain integrity
Releases are built by GitHub Actions and published with:
- **cosign keyless signature** (Sigstore/Fulcio/Rekor) on the `checksums.txt` file, which in turn pins the
  SHA-256 of every release archive,
- **SLSA build-provenance attestations** on each binary archive and on the container image digest
  (`actions/attest-build-provenance`),
- **SBOMs** (syft) for the binary archives and the container image (the image SBOM is attested).

**Verify a released binary archive** — verify the signed checksums, then confirm your download's hash is listed:
```bash
# 1) verify the cosign keyless signature on the checksums file
cosign verify-blob \
  --certificate-identity-regexp 'https://github.com/qatoolist/wowapi/.github/workflows/release.yml@.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --signature checksums.txt.sig --certificate checksums.txt.pem checksums.txt
# 2) confirm the archive you downloaded matches a listed checksum
sha256sum -c checksums.txt --ignore-missing
```
Or verify the archive's SLSA provenance directly (no checksums step needed):
```bash
gh attestation verify wowapi_<version>_<os>_<arch>.tar.gz --owner qatoolist
```

**Verify the container image and its provenance:**
```bash
gh attestation verify oci://ghcr.io/qatoolist/wowapi:<tag> --owner qatoolist
```

All third-party GitHub Actions are pinned to full commit SHAs and updated via Dependabot with a 7-day cooldown.
