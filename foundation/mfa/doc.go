// Package mfa provides reusable, standards-compliant multi-factor-authentication
// factor primitives: TOTP (RFC 6238) and HOTP (RFC 4226) code generation and
// verification, numeric one-time-passcode (OTP) generation with salted
// constant-time hashing, pure challenge-policy helpers (TTL + attempt-limit
// enforcement), and delivery-port interfaces for out-of-band code senders
// (SMS/email) with test/log adapters.
//
// This package is a leaf: it has no dependency on kernel/auth, kernel/authz,
// or any storage/schema concept, and nothing outside it should have to import
// it to get step-up authorization working. The framework's step-up semantics
// (kernel/authz's Decision.StepUpRequired, the `amr` claim on authz.Actor) are
// documented in docs/user-guide/auth.md's "Step-up / MFA" section and consume
// an auth-methods-reference the *product* asserts after a factor challenge
// succeeds — kernel/mfa is what a product uses to implement the factor
// challenge itself (compute/verify a TOTP or OTP code) so it doesn't have to
// hand-roll HMAC truncation or timing-safe comparisons. The two are linked by
// convention, not by an import: mfa produces "this code was valid", the
// product's auth flow turns that into an `amr` entry, kernel/authz/kernel/auth
// consume the resulting claim.
//
// Deliberately OUT of scope (product-owned):
//   - Enrollment UX (QR code rendering, backup codes, recovery flows).
//   - Factor storage schema (how/where a product persists secrets, challenge
//     rows, attempt counters — this package's ChallengeState is an in-memory
//     value the caller supplies from wherever it stores state).
//   - Delivery provider selection (which SMS/email vendor, retry policy,
//     rate limiting) — only the Sender port and a log/fake adapter are here.
//   - Policy decisions (which actions require which factor, whether a factor
//     satisfies a given permission's step_up requirement).
package mfa
