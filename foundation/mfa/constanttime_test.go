package mfa_test

import (
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/foundation/mfa"
)

// TestConstantTimeComparisons_Semantics asserts that every security-sensitive
// comparison in this package (OTP hash verification, TOTP code verification)
// behaves exactly like the constant-time primitives it is built on
// (crypto/subtle.ConstantTimeCompare, per the same convention as
// kernel/apikey.Verify and kernel/webhook.HMACVerifier.Verify): equal inputs
// match, unequal inputs of the same length do not match, and inputs of
// different lengths do not match (and, critically, never panic or index out
// of range — a naive byte-by-byte loop that assumes equal lengths is exactly
// the kind of bug a non-constant-time rewrite could introduce). This does not
// measure wall-clock timing (that is inherently flaky in a unit test); it
// pins the *semantics* that only hold if the underlying comparison really is
// crypto/subtle.ConstantTimeCompare/hmac.Equal rather than a short-circuiting
// "==" or bytes.Equal-then-early-return that a refactor could accidentally
// introduce.
func TestConstantTimeComparisons_Semantics(t *testing.T) {
	t.Run("VerifyOTPCode", func(t *testing.T) {
		salt, code := "salt-123", "482913"
		hash := mfa.HashOTPCode(salt, code)

		if !mfa.VerifyOTPCode(salt, code, hash) {
			t.Error("equal hash must match")
		}
		if mfa.VerifyOTPCode(salt, "482914", hash) {
			t.Error("unequal same-length hash must not match")
		}
		// Length-mismatch inputs: truncated and extended hash strings must
		// both be rejected without panicking.
		if mfa.VerifyOTPCode(salt, code, hash[:len(hash)-1]) {
			t.Error("truncated (shorter) hash must not match")
		}
		if mfa.VerifyOTPCode(salt, code, hash+"0") {
			t.Error("extended (longer) hash must not match")
		}
		if mfa.VerifyOTPCode(salt, code, "") {
			t.Error("empty hash must not match")
		}
	})

	t.Run("VerifyTOTPAt", func(t *testing.T) {
		secret, err := mfa.GenerateTOTPSecret(20)
		if err != nil {
			t.Fatalf("GenerateTOTPSecret: %v", err)
		}
		now := time.Unix(1_700_000_000, 0).UTC()
		opts := mfa.TOTPOptions{Skew: 0}
		code, err := mfa.TOTPCodeAt(secret, now, opts)
		if err != nil {
			t.Fatalf("TOTPCodeAt: %v", err)
		}

		ok, err := mfa.VerifyTOTPAt(secret, code, now, opts)
		if err != nil || !ok {
			t.Fatalf("equal code must match: ok=%v err=%v", ok, err)
		}

		// Unequal, same length (6 digits).
		wrong := "000000"
		if wrong == code {
			wrong = "111111"
		}
		ok, err = mfa.VerifyTOTPAt(secret, wrong, now, opts)
		if err != nil || ok {
			t.Fatalf("unequal same-length code must not match: ok=%v err=%v", ok, err)
		}

		// Length mismatch: shorter and longer than the real code, must not
		// match and must not panic.
		ok, err = mfa.VerifyTOTPAt(secret, code[:len(code)-1], now, opts)
		if err != nil || ok {
			t.Fatalf("shorter code must not match: ok=%v err=%v", ok, err)
		}
		ok, err = mfa.VerifyTOTPAt(secret, code+"9", now, opts)
		if err != nil || ok {
			t.Fatalf("longer code must not match: ok=%v err=%v", ok, err)
		}
		ok, err = mfa.VerifyTOTPAt(secret, "", now, opts)
		if err != nil || ok {
			t.Fatalf("empty code must not match: ok=%v err=%v", ok, err)
		}
	})
}
