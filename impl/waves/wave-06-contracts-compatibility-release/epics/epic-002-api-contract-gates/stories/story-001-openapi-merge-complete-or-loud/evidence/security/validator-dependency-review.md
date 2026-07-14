---
id: EV-W06-E02-S001-004
type: dependency-security-licence-review
parent_story: W06-E02-S001
task: W06-E02-S001-T002
status: produced
created_at: 2026-07-13
updated_at: 2026-07-13
---

# Validator dependency decision — `libopenapi` + `libopenapi-validator`

## Evidence record

- **Evidence ID:** EV-W06-E02-S001-004
- **Evidence type:** dependency security/licence review and decision record
- **Story and task:** W06-E02-S001 / W06-E02-S001-T002
- **Acceptance criteria proven:** AC-W06-E02-S001-04
- **Execution command:** `go list -m -json github.com/pb33f/libopenapi@v0.38.6`; `go list -m -json github.com/pb33f/libopenapi-validator@v0.14.0`; primary-source review of repository licences, module metadata, Context7 API documentation, and deps.dev version metadata
- **Code revision or commit SHA:** `733ef3e930cbb3f89f5bbc53d8f562c60e426513`
- **Branch or tag:** `main`
- **Execution environment:** Darwin arm64; Go 1.26.5
- **Relevant tool versions:** Go 1.26.5; libopenapi v0.38.6; libopenapi-validator v0.14.0
- **Date and time:** 2026-07-13T16:52:37Z
- **Result:** PASS — pin `github.com/pb33f/libopenapi` v0.38.6 for parsing/model build/semantic comparison and `github.com/pb33f/libopenapi-validator` v0.14.0 for complete OpenAPI 3.1.1 document validation
- **File or URI:** this record; https://github.com/pb33f/libopenapi/blob/v0.38.6/LICENSE; https://github.com/pb33f/libopenapi-validator/blob/v0.14.0/LICENSE.md; https://api.deps.dev/v3/systems/go/packages/github.com%2Fpb33f%2Flibopenapi/versions/v0.38.6; https://api.deps.dev/v3/systems/go/packages/github.com%2Fpb33f%2Flibopenapi-validator/versions/v0.14.0
- **Checksum:** not applicable (review record)
- **Reviewer:** W06E02Impl (implementation-time dependency reviewer); independent story review remains required by W06-E02-S001-T004
- **Superseded evidence:** not applicable

## Candidate assessment

`github.com/pb33f/libopenapi` v0.38.6 plus `github.com/pb33f/libopenapi-validator` v0.14.0 were selected over bespoke validation. `libopenapi` supplies the OpenAPI 3.1 model, reference index, and semantic `CompareDocuments` API. The companion validator's `ValidateDocument` API performs specification-level structure, required-field, enum, and JSON Schema validation; this closes the gap intentionally exposed by the malformed-parameter fixture, which a model build alone does not reject. Both modules declare Go 1.25.7 and are compatible with this repository's Go 1.26.5 toolchain.

## Licence review

- Both tagged projects are MIT licensed. Their licences permit use, modification, redistribution, sublicensing, and sale with preservation of the copyright and licence notice.
- deps.dev reports MIT for both pinned versions and marks neither deprecated.
- No copyleft or source-distribution obligation is introduced by either direct dependency.

## Security review

- deps.dev reports no advisory keys for either pinned version as of the review timestamp.
- A targeted public advisory/CVE search found no published advisory specific to either module at the selected version.
- Neither upstream repository publishes a `SECURITY.md`; this is a residual process risk, not a known vulnerability.
- The dependencies parse contract files only in developer/CI tooling. They are not linked into an HTTP request path and perform no network access in the selected APIs.
- Versions are pinned rather than tracking `main` or `latest`. Repository vulnerability scanning remains the ongoing control for the direct and transitive dependency graph.

## Decision

**Approved with low residual risk.** Wire only v0.38.6 and v0.14.0 after this dated record. Reject future upgrades until licence and advisory metadata are re-reviewed. The missing upstream security policies are accepted as a documented low process risk because there are no published advisories, execution is build-time only, versions are pinned, and automated vulnerability scanning remains in force.
