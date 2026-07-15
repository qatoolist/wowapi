package mfa

import "time"

// Default challenge policy values. A delivered (SMS/email) OTP challenge is
// conventionally short-lived and attempt-limited; these mirror common
// authenticator/OTP-flow defaults and are overridable per policy.
const (
	DefaultChallengeTTL         = 5 * time.Minute
	DefaultChallengeMaxAttempts = 5
)

// ChallengeState is the caller-supplied, product-persisted state of a single
// in-flight challenge (an issued TOTP/OTP code awaiting verification). This
// package has no storage schema of its own — a product loads these fields
// from wherever it persists challenges (its own table, cache, etc.) and
// passes them to ChallengePolicy for evaluation.
type ChallengeState struct {
	// IssuedAt is when the challenge was created.
	IssuedAt time.Time
	// Attempts is the number of verification attempts made so far.
	Attempts int
	// Consumed is true once the challenge has been successfully verified
	// (a consumed challenge must not be reusable).
	Consumed bool
}

// ChallengeStatus is the outcome of evaluating a ChallengeState against a
// ChallengePolicy.
type ChallengeStatus int

const (
	// ChallengeOK means the challenge is still live: not expired, not
	// attempt-exhausted, not consumed — the caller may attempt verification.
	ChallengeOK ChallengeStatus = iota
	// ChallengeExpired means the challenge's TTL has elapsed.
	ChallengeExpired
	// ChallengeAttemptsExhausted means the attempt limit has been reached
	// or exceeded.
	ChallengeAttemptsExhausted
	// ChallengeConsumed means the challenge was already successfully
	// verified and must not be reused.
	ChallengeConsumed
)

// OK reports whether the status permits a further verification attempt.
func (s ChallengeStatus) OK() bool { return s == ChallengeOK }

// String implements fmt.Stringer for readable test failures/log lines.
func (s ChallengeStatus) String() string {
	switch s {
	case ChallengeOK:
		return "ok"
	case ChallengeExpired:
		return "expired"
	case ChallengeAttemptsExhausted:
		return "attempts_exhausted"
	case ChallengeConsumed:
		return "consumed"
	default:
		return "unknown"
	}
}

// ChallengePolicy is pure TTL + attempt-limit enforcement logic for a
// delivered-code challenge (SMS/email OTP or similar). It holds no state and
// touches no storage — a product owns the challenge row/cache entry and
// calls Evaluate with the state it loaded to decide whether to accept a
// verification attempt.
type ChallengePolicy struct {
	// TTL is how long a challenge remains valid after issuance. Zero
	// resolves to DefaultChallengeTTL.
	TTL time.Duration
	// MaxAttempts is the number of verification attempts allowed before the
	// challenge is locked out. Zero resolves to DefaultChallengeMaxAttempts.
	MaxAttempts int
}

// TTLOrDefault returns p.TTL, or DefaultChallengeTTL if unset.
func (p ChallengePolicy) TTLOrDefault() time.Duration {
	if p.TTL <= 0 {
		return DefaultChallengeTTL
	}
	return p.TTL
}

// MaxAttemptsOrDefault returns p.MaxAttempts, or DefaultChallengeMaxAttempts
// if unset.
func (p ChallengePolicy) MaxAttemptsOrDefault() int {
	if p.MaxAttempts <= 0 {
		return DefaultChallengeMaxAttempts
	}
	return p.MaxAttempts
}

// ExpiresAt returns the instant a challenge issued at issuedAt expires under p.
func (p ChallengePolicy) ExpiresAt(issuedAt time.Time) time.Time {
	return issuedAt.Add(p.TTLOrDefault())
}

// Expired reports whether a challenge in state st has passed its TTL as of now.
func (p ChallengePolicy) Expired(st ChallengeState, now time.Time) bool {
	return !now.Before(p.ExpiresAt(st.IssuedAt))
}

// AttemptsExhausted reports whether st has reached or exceeded the attempt
// limit.
func (p ChallengePolicy) AttemptsExhausted(st ChallengeState) bool {
	return st.Attempts >= p.MaxAttemptsOrDefault()
}

// Evaluate returns the single ChallengeStatus describing whether st may
// still be verified as of now. Priority when multiple conditions hold:
// consumed > expired > attempts-exhausted > ok — a consumed challenge is
// permanently done regardless of timing, and an expired challenge is
// reported as expired even if attempts also happen to be exhausted (the
// caller-visible reason should be the one that would have applied first in
// time: consumption and expiry are absolute facts, exhaustion is a
// count-based fact that stops mattering once the challenge would have
// expired anyway).
func (p ChallengePolicy) Evaluate(st ChallengeState, now time.Time) ChallengeStatus {
	switch {
	case st.Consumed:
		return ChallengeConsumed
	case p.Expired(st, now):
		return ChallengeExpired
	case p.AttemptsExhausted(st):
		return ChallengeAttemptsExhausted
	default:
		return ChallengeOK
	}
}
