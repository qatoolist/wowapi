// Package mfa is the deprecated forwarding shim for foundation/mfa.
// It will be removed in a future minor version.
package mfa

import (
	"log/slog"
	"time"

	"github.com/qatoolist/wowapi/v2/foundation/mfa"
)

// Default challenge policy values.
const (
	DefaultChallengeTTL         = mfa.DefaultChallengeTTL
	DefaultChallengeMaxAttempts = mfa.DefaultChallengeMaxAttempts
)

// ChallengeState type alias.
type ChallengeState = mfa.ChallengeState

// ChallengeStatus type alias.
type ChallengeStatus = mfa.ChallengeStatus

// ChallengeStatus constants.
const (
	ChallengeOK                = mfa.ChallengeOK
	ChallengeExpired           = mfa.ChallengeExpired
	ChallengeAttemptsExhausted = mfa.ChallengeAttemptsExhausted
	ChallengeConsumed          = mfa.ChallengeConsumed
)

// ChallengePolicy type alias.
type ChallengePolicy = mfa.ChallengePolicy

// Default TOTP parameters.
const (
	DefaultTOTPStep      = mfa.DefaultTOTPStep
	DefaultTOTPDigits    = mfa.DefaultTOTPDigits
	DefaultTOTPSkew      = mfa.DefaultTOTPSkew
	DefaultTOTPSecretLen = mfa.DefaultTOTPSecretLen
)

// TOTPOptions type alias.
type TOTPOptions = mfa.TOTPOptions

// GenerateTOTPSecret forwards to foundation/mfa.
func GenerateTOTPSecret(n int) ([]byte, error) {
	return mfa.GenerateTOTPSecret(n)
}

// EncodeSecretBase32 forwards to foundation/mfa.
func EncodeSecretBase32(secret []byte) (string, error) {
	return mfa.EncodeSecretBase32(secret)
}

// DecodeSecretBase32 forwards to foundation/mfa.
func DecodeSecretBase32(s string) ([]byte, error) {
	return mfa.DecodeSecretBase32(s)
}

// TOTPCodeAt forwards to foundation/mfa.
func TOTPCodeAt(secret []byte, t time.Time, opts TOTPOptions) (string, error) {
	return mfa.TOTPCodeAt(secret, t, opts)
}

// VerifyTOTPAt forwards to foundation/mfa.
func VerifyTOTPAt(secret []byte, code string, t time.Time, opts TOTPOptions) (bool, error) {
	return mfa.VerifyTOTPAt(secret, code, t, opts)
}

// DefaultOTPDigits constant.
const DefaultOTPDigits = mfa.DefaultOTPDigits

// GenerateOTPCode forwards to foundation/mfa.
func GenerateOTPCode(digits int) (string, error) {
	return mfa.GenerateOTPCode(digits)
}

// HashOTPCode forwards to foundation/mfa.
func HashOTPCode(salt, code string) string {
	return mfa.HashOTPCode(salt, code)
}

// VerifyOTPCode forwards to foundation/mfa.
func VerifyOTPCode(salt, code, wantHash string) bool {
	return mfa.VerifyOTPCode(salt, code, wantHash)
}

// Sender type alias.
type Sender = mfa.Sender

// FakeSender type alias.
type FakeSender = mfa.FakeSender

// NewLogSender forwards to foundation/mfa.
func NewLogSender(log *slog.Logger) Sender {
	return mfa.NewLogSender(log)
}

// Algorithm type alias.
type Algorithm = mfa.Algorithm

// Algorithm constants.
const (
	AlgSHA1   = mfa.AlgSHA1
	AlgSHA256 = mfa.AlgSHA256
	AlgSHA512 = mfa.AlgSHA512
)

// HOTPCode forwards to foundation/mfa.
func HOTPCode(key []byte, counter uint64, alg Algorithm, digits int) (string, error) {
	return mfa.HOTPCode(key, counter, alg, digits)
}
