package mfa_test

import (
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/mfa"
)

var benchmarkTOTPCode string

// BenchmarkTOTPDerive measures the HMAC-based TOTP derivation used for every
// authenticator-code challenge, with the framework's production defaults.
func BenchmarkTOTPDerive(b *testing.B) {
	secret := []byte("0123456789abcdefghij")
	at := time.Unix(1_800_000_000, 0)

	b.ReportAllocs()
	for b.Loop() {
		code, err := mfa.TOTPCodeAt(secret, at, mfa.TOTPOptions{})
		if err != nil {
			b.Fatalf("derive TOTP: %v", err)
		}
		benchmarkTOTPCode = code
	}
}
