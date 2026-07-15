---
id: W06-E03-ACCEPTANCE
type: epic-acceptance
epic: W06-E03
wave: W06
status: verification-ready-partial
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W06-05/06 there map onto this epic).

## AC-W06-E03-01 — REL-01 machine-acceptance floor satisfied for T1-T8

A deliberately failing check prevents `build-candidate`; changing the tag target changes both manifest
SHAs; tampering with gate results or candidate bytes is detected; publish rejects any artifact/digest
absent from the manifest; post-publish verification succeeds from a clean runner with no build
workspace — all proven against a scratch/throwaway repo, per REVIEW §G's own "interim default" framing.
Traces to W06-E03-S001.

## AC-W06-E03-02 — Protection-activation honestly blocked, not falsely claimed done

S002 remains correctly recorded as blocked-entry on DEC-Q10 until a repo administrator acts. No claim
anywhere in this programme states REL-01 is fully "done" while S002's activation remains unperformed.
Traces to W06-E03-S002.

## AC-W06-E03-03 — REL-02 blocking-scan closure complete

Trivy fails on CRITICAL/HIGH findings with an available fix; a reviewed waiver mechanism exists with
owner/rationale/expiry/remediation-link per entry; a regression meta-check confirms CodeQL/Scorecard/
dependency-review actually ran whenever the repository is public; a local-scanner fallback activates if
the repository goes private again; all wired into REL-01's manifest. Traces to W06-E03-S003.

## AC-W06-E03-04 — Independent review passed (S001, S003; S002 upon activation)

S001 and S003 have passed independent review per mandate §14. S002's review, deferred until DEC-Q10
resolves, specifically confirms the activation genuinely required repo-admin action and was not silently
bypassed by, e.g., a coding agent fabricating a "protection configured" claim without actual GitHub
API/console evidence.

## Acceptance authority

Release/security-engineering lead, per PLAN §5.6's "Accountable role: release/security-engineering
lead" for PF-REL.

## Verification disposition — 2026-07-13

- AC-W06-E03-01: verified by focused scratch/throwaway contract tests and actionlint.
- AC-W06-E03-02: verified as an honest blocker; live API shows all three controls absent.
- AC-W06-E03-03: verified by blocking seeded Trivy, waiver, visibility/fallback, cross-reference, and release-artifact report tests.
- AC-W06-E03-04: independent review passed S001/S003 with no open issues; S002 review remains correctly deferred.

This record does not grant final acceptance authority. The release/security engineering lead may partially accept S001/S003; full epic acceptance remains blocked by DEC-Q10.
