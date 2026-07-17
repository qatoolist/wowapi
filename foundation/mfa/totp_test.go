package mfa_test

import (
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/foundation/mfa"
)

// RFC 6238 Appendix B published test vectors: three seeds (SHA1: 20 ASCII
// bytes "12345678901234567890"; SHA256: that string repeated to 32 bytes;
// SHA512: repeated/truncated to 64 bytes), step X=30s, T0=0, 8-digit codes.
var (
	rfc6238SeedSHA1   = []byte("12345678901234567890")
	rfc6238SeedSHA256 = []byte("12345678901234567890123456789012")
	rfc6238SeedSHA512 = []byte("1234567890123456789012345678901234567890123456789012345678901234")
)

func TestTOTP_RFC6238Vectors(t *testing.T) {
	cases := []struct {
		unixSec int64
		sha1    string
		sha256  string
		sha512  string
	}{
		{59, "94287082", "46119246", "90693936"},
		{1111111109, "07081804", "68084774", "25091201"},
		{1111111111, "14050471", "67062674", "99943326"},
		{1234567890, "89005924", "91819424", "93441116"},
		{2000000000, "69279037", "90698825", "38618901"},
		{20000000000, "65353130", "77737706", "47863826"},
	}
	opts := mfa.TOTPOptions{Step: 30 * time.Second, Digits: 8, Skew: 0}
	for _, c := range cases {
		at := time.Unix(c.unixSec, 0).UTC()

		got1, err := mfa.TOTPCodeAt(rfc6238SeedSHA1, at, mergeAlg(opts, mfa.AlgSHA1))
		if err != nil {
			t.Fatalf("t=%d SHA1: %v", c.unixSec, err)
		}
		if got1 != c.sha1 {
			t.Errorf("t=%d SHA1 = %q, want %q", c.unixSec, got1, c.sha1)
		}

		got256, err := mfa.TOTPCodeAt(rfc6238SeedSHA256, at, mergeAlg(opts, mfa.AlgSHA256))
		if err != nil {
			t.Fatalf("t=%d SHA256: %v", c.unixSec, err)
		}
		if got256 != c.sha256 {
			t.Errorf("t=%d SHA256 = %q, want %q", c.unixSec, got256, c.sha256)
		}

		got512, err := mfa.TOTPCodeAt(rfc6238SeedSHA512, at, mergeAlg(opts, mfa.AlgSHA512))
		if err != nil {
			t.Fatalf("t=%d SHA512: %v", c.unixSec, err)
		}
		if got512 != c.sha512 {
			t.Errorf("t=%d SHA512 = %q, want %q", c.unixSec, got512, c.sha512)
		}
	}
}

func mergeAlg(o mfa.TOTPOptions, alg mfa.Algorithm) mfa.TOTPOptions {
	o.Algorithm = alg
	return o
}

func TestTOTP_VerifyRFC6238Vectors(t *testing.T) {
	opts := mfa.TOTPOptions{Step: 30 * time.Second, Digits: 8, Algorithm: mfa.AlgSHA1, Skew: 0}
	at := time.Unix(59, 0).UTC()
	ok, err := mfa.VerifyTOTPAt(rfc6238SeedSHA1, "94287082", at, opts)
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if !ok {
		t.Error("VerifyTOTPAt: RFC 6238 known-good vector rejected")
	}
}

func TestTOTP_DefaultOptions(t *testing.T) {
	// Defaults: step=30s, digits=6, algorithm=SHA1, skew=1 — matches the
	// common authenticator-app convention (Google Authenticator etc).
	secret, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	code, err := mfa.TOTPCodeAt(secret, now, mfa.TOTPOptions{})
	if err != nil {
		t.Fatalf("TOTPCodeAt: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("default digit count = %d, want 6", len(code))
	}
	ok, err := mfa.VerifyTOTPAt(secret, code, now, mfa.TOTPOptions{})
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if !ok {
		t.Error("VerifyTOTPAt: valid code at issuance rejected under default options")
	}
}

func TestTOTP_GenerateSecret_Base32Encoded(t *testing.T) {
	secret, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	b32, err := mfa.EncodeSecretBase32(secret)
	if err != nil {
		t.Fatalf("EncodeSecretBase32: %v", err)
	}
	decoded, err := mfa.DecodeSecretBase32(b32)
	if err != nil {
		t.Fatalf("DecodeSecretBase32: %v", err)
	}
	if string(decoded) != string(secret) {
		t.Error("round-trip base32 encode/decode changed the secret")
	}
}

func TestTOTP_DecodeSecretBase32_MalformedRejected(t *testing.T) {
	if _, err := mfa.DecodeSecretBase32("not-valid-base32!!!"); err == nil {
		t.Error("DecodeSecretBase32: malformed secret accepted")
	}
}

func TestTOTP_DecodeSecretBase32_CaseAndWhitespaceInsensitive(t *testing.T) {
	secret, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	b32, err := mfa.EncodeSecretBase32(secret)
	if err != nil {
		t.Fatalf("EncodeSecretBase32: %v", err)
	}
	lower := "  " + toLowerSpaced(b32) + "  "
	decoded, err := mfa.DecodeSecretBase32(lower)
	if err != nil {
		t.Fatalf("DecodeSecretBase32(lowercased+padded): %v", err)
	}
	if string(decoded) != string(secret) {
		t.Error("case/whitespace-insensitive decode changed the secret")
	}
}

func toLowerSpaced(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c - 'A' + 'a'
		}
	}
	return string(b)
}

func TestTOTP_ClockSkewWindow(t *testing.T) {
	secret, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	opts := mfa.TOTPOptions{Step: 30 * time.Second, Digits: 6, Algorithm: mfa.AlgSHA1, Skew: 1}
	now := time.Unix(1_700_000_000, 0).UTC()
	code, err := mfa.TOTPCodeAt(secret, now, opts)
	if err != nil {
		t.Fatalf("TOTPCodeAt: %v", err)
	}

	// One step (30s) later: within the ±1 step window.
	ok, err := mfa.VerifyTOTPAt(secret, code, now.Add(30*time.Second), opts)
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if !ok {
		t.Error("code rejected within ±1 step skew window")
	}

	// Three steps (90s) later: outside the window.
	ok, err = mfa.VerifyTOTPAt(secret, code, now.Add(90*time.Second), opts)
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if ok {
		t.Error("code accepted outside the skew window")
	}
}

func TestTOTP_ZeroSkewRejectsAdjacentStep(t *testing.T) {
	secret, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	opts := mfa.TOTPOptions{Step: 30 * time.Second, Digits: 6, Algorithm: mfa.AlgSHA1, Skew: 0}
	now := time.Unix(1_700_000_000, 0).UTC()
	code, err := mfa.TOTPCodeAt(secret, now, opts)
	if err != nil {
		t.Fatalf("TOTPCodeAt: %v", err)
	}
	ok, err := mfa.VerifyTOTPAt(secret, code, now.Add(30*time.Second), opts)
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if ok {
		t.Error("skew=0 must reject the adjacent step's code")
	}
}

func TestTOTP_WrongCodeRejected(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	real, err := mfa.TOTPCodeAt(rfc6238SeedSHA1, now, mfa.TOTPOptions{})
	if err != nil {
		t.Fatalf("TOTPCodeAt: %v", err)
	}
	wrong := "000000"
	if wrong == real {
		wrong = "000001"
	}
	ok, err := mfa.VerifyTOTPAt(rfc6238SeedSHA1, wrong, now, mfa.TOTPOptions{})
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if ok {
		t.Fatalf("wrong code %q accepted (real code %q)", wrong, real)
	}
}

func TestTOTP_InvalidOptionsRejected(t *testing.T) {
	secret, _ := mfa.GenerateTOTPSecret(20)
	now := time.Now()
	if _, err := mfa.TOTPCodeAt(secret, now, mfa.TOTPOptions{Step: -1}); err == nil {
		t.Error("negative step accepted")
	}
	if _, err := mfa.TOTPCodeAt(secret, now, mfa.TOTPOptions{Digits: 11}); err == nil {
		t.Error("digits=11 accepted")
	}
	if _, err := mfa.TOTPCodeAt(secret, now, mfa.TOTPOptions{Algorithm: "bogus"}); err == nil {
		t.Error("unknown algorithm accepted")
	}
}

func TestTOTP_GenerateSecret_LengthAndRandomness(t *testing.T) {
	s1, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	if len(s1) != 20 {
		t.Fatalf("secret length = %d, want 20", len(s1))
	}
	s2, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	if string(s1) == string(s2) {
		t.Error("two generated secrets must not be identical")
	}
}

func TestTOTP_GenerateSecret_RejectsNonPositiveLength(t *testing.T) {
	if _, err := mfa.GenerateTOTPSecret(0); err == nil {
		t.Error("GenerateTOTPSecret(0) accepted")
	}
	if _, err := mfa.GenerateTOTPSecret(-1); err == nil {
		t.Error("GenerateTOTPSecret(-1) accepted")
	}
}

func TestTOTP_NegativeSkewRejected(t *testing.T) {
	secret, _ := mfa.GenerateTOTPSecret(20)
	if _, err := mfa.VerifyTOTPAt(secret, "000000", time.Now(), mfa.TOTPOptions{Skew: -1}); err == nil {
		t.Error("negative skew accepted")
	}
}

func TestTOTP_VerifyTOTPAt_PropagatesInvalidOptions(t *testing.T) {
	secret, _ := mfa.GenerateTOTPSecret(20)
	if _, err := mfa.VerifyTOTPAt(secret, "000000", time.Now(), mfa.TOTPOptions{Digits: 11}); err == nil {
		t.Error("VerifyTOTPAt: invalid digits accepted")
	}
}

// TestTOTP_VerifyTOTPAt_NearEpochSkipsNegativeCounters exercises the guard
// that skips a would-be-negative HOTP counter when skew pushes the window
// before the Unix epoch — a real (if unusual) input, not just a defensive
// dead branch, since nothing stops a caller from checking a code against a
// clock set near 1970.
func TestTOTP_VerifyTOTPAt_NearEpochSkipsNegativeCounters(t *testing.T) {
	secret, err := mfa.GenerateTOTPSecret(20)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	opts := mfa.TOTPOptions{Step: 30 * time.Second, Digits: 6, Algorithm: mfa.AlgSHA1, Skew: 2}
	near0 := time.Unix(10, 0).UTC() // counter=0; skew of 2 would probe counter -2 and -1
	code, err := mfa.TOTPCodeAt(secret, near0, opts)
	if err != nil {
		t.Fatalf("TOTPCodeAt: %v", err)
	}
	ok, err := mfa.VerifyTOTPAt(secret, code, near0, opts)
	if err != nil {
		t.Fatalf("VerifyTOTPAt: %v", err)
	}
	if !ok {
		t.Error("VerifyTOTPAt: valid code near the epoch with skew rejected")
	}
}

func TestTOTP_EncodeSecretBase32_RejectsEmpty(t *testing.T) {
	if _, err := mfa.EncodeSecretBase32(nil); err == nil {
		t.Error("EncodeSecretBase32(nil) accepted")
	}
	if _, err := mfa.EncodeSecretBase32([]byte{}); err == nil {
		t.Error("EncodeSecretBase32(empty) accepted")
	}
}
