# SEC-05 version-pinned control map

> **Canonical machine-readable artifact:** [`control-map.json`](control-map.json). This Markdown view is descriptive; the JSON plus `validate_control_map.py` is authoritative.

## Scope

This profile is deliberately bounded to the four framework security slices named by SEC-05: SEC-01 server-side session grants, SEC-03 authenticated webhook replay/deduplication, SEC-04 bounded and invalidated authorization caching, and SEC-06 outbound-security governance. It does not certify a consuming application, browser UI, identity provider, or product repository.

Every source inventory item has exactly one disposition in `control-map.json`: `applicable`, `not_applicable` with a non-empty rationale, or `waived` with an actually approved owner/approver/rationale/expiry record. There are no waivers in this profile.

## Pinned sources

| Standard | Version | Canonical source | Local inventory | Source pin | Inventory SHA-256 | Inventory | Applicable | N/A | Waived |
|---|---|---|---|---|---|---:|---:|---:|---:|
| ASVS | `5.0.0` | https://raw.githubusercontent.com/OWASP/ASVS/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv | `SEC-05/sources/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv` | OWASP/ASVS git tag v5.0.0; official CSV SHA-256 98c8fe911b9edb403af8ee05d3ce8201ecac2659e313b053890a62847cdcf680 | `98c8fe911b9edb403af8ee05d3ce8201ecac2659e313b053890a62847cdcf680` | 345 | 24 | 321 | 0 |
| OWASP-API-TOP-10 | `2023` | https://owasp.org/API-Security/editions/2023/en/0x11-t10/ | `SEC-05/sources/owasp-api-security-top-10-2023.json` | OWASP API Security edition 2023 canonical category list; local ten-category inventory SHA-256 pinned separately | `772078bd4be4293a5efe8b693671c0506efd38a746c4a69bc9b49a9e5b50bfa6` | 10 | 7 | 3 | 0 |
| NIST-SP-800-63 | `final-2025-07` | https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-63-4.pdf | `SEC-05/sources/nist-sp-800-63-4-normative-inventory.json` | DOI 10.6028/NIST.SP.800-63-4; final PDF SHA-256 2f5b107de218a0fc3fe7f7e91b2ece5babb001eda5745b7c568a3cc876002f2d | `31f39cefd2033a9661fecd75674153430392710d7a94e16b58a8e730c157586e` | 57 | 2 | 55 | 0 |

### NIST inventory granularity

The final July 2025 **main** NIST SP 800-63-4 publication has no external control-ID catalog. The pinned local inventory therefore assigns stable profile IDs to each official HTML paragraph or list item containing normative `SHALL` or `MUST` language (57 units), while pinning the final PDF by DOI and SHA-256. SP 800-63A/B/C technical companion requirements are not silently relabeled as controls from the main SP 800-63-4 publication. The external assessor must confirm this applicability method; until that professional assessment exists, this map is an internally prepared input, not an external certification.

## Applicability method

- ASVS 5.0.0: all 345 official CSV requirements are inventoried. Controls fully exercised by the four SEC slices are mapped to focused executable regressions. Partial or consumer-owned controls are explicitly N/A rather than overstated.
- OWASP API Security Top 10 2023: all ten risk categories are inventoried. Seven categories apply to this bounded framework profile; API3, API4, and API9 are explicitly consumer/application-owned and N/A here.
- NIST SP 800-63-4: organizational DIRM duties are N/A for a framework library. The two profile-governance duties applicable to SEC-05—baseline-control inventory and compensating-control documentation—map to executable validator regression tests.

## Applicable controls

| Standard | Control | Executable test(s) |
|---|---|---|
| ASVS `V1.3.6` | Verify that the application protects against Server-side Request Forgery (SSRF) attacks, by validating untrusted data against an allowlist of protocols, domains, paths and ports and sanitizing potentially dangerous characters before using the data to call another service. | `kernel/httpclient/client_test.go::TestClientBlocksLoopbackByDefault` |
| ASVS `V2.2.1` | Verify that input is validated to enforce business or functional expectations for that input. This should either use positive validation against an allow list of values, patterns, and ranges, or be based on comparing the input to an expected structure and logical limits according to predefined rules. For L1, this can focus on input which is used to make specific business or security decisions. For L2 and up, this should apply to all input. | `kernel/auth/auth_test.go::TestActor_ExplicitCapacityValidatedServerSide` |
| ASVS `V2.2.2` | Verify that the application is designed to enforce input validation at a trusted service layer. While client-side validation improves usability and should be encouraged, it must not be relied upon as a security control. | `kernel/auth/auth_test.go::TestActor_ExplicitCapacityValidatedServerSide` |
| ASVS `V2.3.3` | Verify that transactions are being used at the business logic level such that either a business logic operation succeeds in its entirety or it is rolled back to the previous correct state. | `kernel/database/coverage_test.go::TestIntegrationWithTenantRollsBackOnError` |
| ASVS `V4.1.5` | Verify that per-message digital signatures are used to provide additional assurance on top of transport protections for requests or transactions which are highly sensitive or which traverse a number of systems. | `foundation/webhook/webhook_test.go::TestIntegrationOutboundSignatureCoversTimestamp` |
| ASVS `V6.8.2` | Verify that the presence and integrity of digital signatures on authentication assertions (for example on JWTs or SAML assertions) are always validated, rejecting any assertions that are unsigned or have invalid signatures. | `kernel/auth/auth_test.go::TestVerify_TamperedSignature`<br>`kernel/auth/auth_test.go::TestVerify_AlgConfusionHS256` |
| ASVS `V6.8.4` | Verify that, if an application uses a separate Identity Provider (IdP) and expects specific authentication strength, methods, or recentness for specific functions, the application verifies this using the information returned by the IdP. For example, if OIDC is used, this might be achieved by validating ID Token claims such as 'acr', 'amr', and 'auth_time' (if present). If the IdP does not provide this information, the application must have a documented fallback approach that assumes that the minimum strength authentication mechanism was used (for example, single-factor authentication using username and password). | `kernel/auth/auth_test.go::TestVerify_AuthTimeAndACRPropagatesToClaims`<br>`kernel/authz/assurance_freshness_test.go::TestStepUpFreshnessStaleAuthTimeFails` |
| ASVS `V7.2.1` | Verify that the application performs all session token verification using a trusted, backend service. | `kernel/auth/auth_test.go::TestActor_PrivilegedSessionResolvedFromGrant` |
| ASVS `V7.5.3` | Verify that the application requires further authentication with at least one factor or secondary verification before performing highly sensitive transactions or operations. | `kernel/auth/auth_test.go::TestVerify_AuthTimeAndACRPropagatesToClaims`<br>`kernel/authz/assurance_freshness_test.go::TestStepUpFreshnessStaleAuthTimeFails` |
| ASVS `V8.2.1` | Verify that the application ensures that function-level access is restricted to consumers with explicit permissions. | `kernel/authz/evaluator_test.go::TestDenyByDefault` |
| ASVS `V8.2.2` | Verify that the application ensures that data-specific access is restricted to consumers with explicit permissions to specific data items to mitigate insecure direct object reference (IDOR) and broken object level authorization (BOLA). | `kernel/authz/evaluator_test.go::TestRBACResourceScopeExact` |
| ASVS `V8.2.4` | Verify that adaptive security controls based on a consumer's environmental and contextual attributes (such as time of day, location, IP address, or device) are implemented for authentication and authorization decisions, as defined in the application's documentation. These controls must be applied when the consumer tries to start a new session and also during an existing session. | `kernel/authz/evaluator_test.go::TestABACDenyUnresolvedAttributeFailsClosed` |
| ASVS `V8.3.1` | Verify that the application enforces authorization rules at a trusted service layer and doesn't rely on controls that an untrusted consumer could manipulate, such as client-side JavaScript. | `kernel/authz/evaluator_test.go::TestDenyByDefault` |
| ASVS `V8.3.2` | Verify that changes to values on which authorization decisions are made are applied immediately. Where changes cannot be applied immediately, (such as when relying on data in self-contained tokens), there must be mitigating controls to alert when a consumer performs an action when they are no longer authorized to do so and revert the change. Note that this alternative would not mitigate information leakage. | `kernel/authz/caching_pg_test.go::TestIntegrationCachingStoreRevokeInvalidate` |
| ASVS `V8.3.3` | Verify that access to an object is based on the originating subject's (e.g. consumer's) permissions, not on the permissions of any intermediary or service acting on their behalf. For example, if a consumer calls a web service using a self-contained token for authentication, and the service then requests data from a different service, the second service will use the consumer's token, rather than a machine-to-machine token from the first service, to make permission decisions. | `kernel/authz/machine_scope_test.go::TestMachineScopeStillSubjectToABACDeny` |
| ASVS `V8.4.1` | Verify that multi-tenant applications use cross-tenant controls to ensure consumer operations will never affect tenants with which they do not have permissions to interact. | `kernel/database/coverage_test.go::TestIntegrationRLSIsolatesTenants` |
| ASVS `V9.1.1` | Verify that self-contained tokens are validated using their digital signature or MAC to protect against tampering before accepting the token's contents. | `kernel/auth/auth_test.go::TestVerify_TamperedSignature`<br>`kernel/auth/auth_test.go::TestVerify_AlgConfusionHS256` |
| ASVS `V9.1.2` | Verify that only algorithms on an allowlist can be used to create and verify self-contained tokens, for a given context. The allowlist must include the permitted algorithms, ideally only either symmetric or asymmetric algorithms, and must not include the 'None' algorithm. If both symmetric and asymmetric must be supported, additional controls will be needed to prevent key confusion. | `kernel/auth/auth_test.go::TestVerify_TamperedSignature`<br>`kernel/auth/auth_test.go::TestVerify_AlgConfusionHS256` |
| ASVS `V9.1.3` | Verify that key material that is used to validate self-contained tokens is from trusted pre-configured sources for the token issuer, preventing attackers from specifying untrusted sources and keys. For JWTs and other JWS structures, headers such as 'jku', 'x5u', and 'jwk' must be validated against an allowlist of trusted sources. | `kernel/auth/jwks_governance_test.go::TestNewJWKSKeySource_ProdCustomClientRequiresTrustedIssuers` |
| ASVS `V9.2.1` | Verify that, if a validity time span is present in the token data, the token and its content are accepted only if the verification time is within this validity time span. For example, for JWTs, the claims 'nbf' and 'exp' must be verified. | `kernel/auth/auth_test.go::TestVerify_ExpiredToken` |
| ASVS `V9.2.2` | Verify that the service receiving a token validates the token to be the correct type and is meant for the intended purpose before accepting the token's contents. For example, only access tokens can be accepted for authorization decisions and only ID Tokens can be used for proving user authentication. | `kernel/auth/auth_test.go::TestVerify_WrongAudience` |
| ASVS `V9.2.3` | Verify that the service only accepts tokens which are intended for use with that service (audience). For JWTs, this can be achieved by validating the 'aud' claim against an allowlist defined in the service. | `kernel/auth/auth_test.go::TestVerify_WrongAudience` |
| ASVS `V11.4.1` | Verify that only approved hash functions are used for general cryptographic use cases, including digital signatures, HMAC, KDF, and random bit generation. Disallowed hash functions, such as MD5, must not be used for any cryptographic purpose. | `foundation/webhook/webhook_test.go::TestIntegrationOutboundSignatureCoversTimestamp` |
| ASVS `V13.2.2` | Verify that communications between backend application components, including local or operating system services, APIs, middleware, and data layers, are performed with accounts assigned the least necessary privileges. | `kernel/authz/escalation_test.go::TestIntegrationRuntimeRoleNotMemberOfPlatform` |
| OWASP-API-TOP-10 `API1:2023` | Broken Object Level Authorization | `kernel/authz/evaluator_test.go::TestRBACResourceScopeExact` |
| OWASP-API-TOP-10 `API2:2023` | Broken Authentication | `kernel/auth/auth_test.go::TestVerify_TamperedSignature`<br>`kernel/auth/auth_test.go::TestVerify_AlgConfusionHS256` |
| OWASP-API-TOP-10 `API5:2023` | Broken Function Level Authorization | `kernel/authz/evaluator_test.go::TestDenyByDefault` |
| OWASP-API-TOP-10 `API6:2023` | Unrestricted Access to Sensitive Business Flows | `kernel/auth/auth_test.go::TestVerify_AuthTimeAndACRPropagatesToClaims`<br>`kernel/authz/assurance_freshness_test.go::TestStepUpFreshnessStaleAuthTimeFails` |
| OWASP-API-TOP-10 `API7:2023` | Server Side Request Forgery | `kernel/httpclient/client_test.go::TestClientBlocksLoopbackByDefault` |
| OWASP-API-TOP-10 `API8:2023` | Security Misconfiguration | `kernel/auth/jwks_test.go::TestJWKS_RejectsNonHTTPSURL` |
| OWASP-API-TOP-10 `API10:2023` | Unsafe Consumption of APIs | `foundation/webhook/webhook_test.go::TestIntegrationHandleInbound_FailedSigDoesNotBlockValid` |
| NIST-SP-800-63 `NIST-SP800-63-4-NORM-025` | Using the initial xALs selected in Sec. 3.3.3, the organization SHALL identify the applicable baseline controls for each user group as follows: | `SEC-05/test_validate_control_map.py::test_rejects_unmapped_control` |
| NIST-SP-800-63 `NIST-SP800-63-4-NORM-036` | Where compensating controls are implemented, organizations SHALL document the compensating control, the rationale for the deviation, comparability of the chosen alternative, and any resulting residual risks. CSPs and IdPs that implement compensating controls SHALL communicate this information to all potential RPs prior to integration to allow the RP to assess and determine the acceptability of the compensating controls for their use cases. | `SEC-05/test_validate_control_map.py::test_rejects_invalid_waiver` |

## Machine validation and focused execution

```sh
python3 SEC-05/validate_control_map.py
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 \
python3 SEC-05/validate_control_map.py --run-tests
```

The validator verifies each committed inventory digest, rejects source-inventory omissions, duplicate or unknown controls, version drift, dangling test names/files, applicable controls without executable tests, N/A controls without rationale, and waivers that are unapproved, ownerless, approverless, rationale-free, or expired. `--run-tests` de-duplicates mapped tests by package and executes the focused set.

## External-assessment boundary

This control map does **not** constitute the independent professional-services assessment required by AC-W07-E02-S001-02. No external assessor, vendor, engagement identifier, assessment report, findings register, or acceptance-authority waiver approval was supplied or discoverable in the repository as of 2026-07-14. See `external-assessment-status.md`.
