package mfa_test

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/mfa"
)

// TestHOTP_RFC4226Vectors pins HOTPCode against RFC 4226 Appendix D's published
// test vectors: secret "12345678901234567890" (ASCII, 20 bytes), SHA-1, 6
// digits, counters 0..9.
func TestHOTP_RFC4226Vectors(t *testing.T) {
	key := []byte("12345678901234567890")
	want := []string{
		"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489",
	}
	for counter, w := range want {
		got, err := mfa.HOTPCode(key, uint64(counter), mfa.AlgSHA1, 6)
		if err != nil {
			t.Fatalf("HOTPCode(counter=%d): %v", counter, err)
		}
		if got != w {
			t.Errorf("HOTPCode(counter=%d) = %q, want %q", counter, got, w)
		}
	}
}

func TestHOTP_RejectsUnsupportedDigits(t *testing.T) {
	key := []byte("12345678901234567890")
	if _, err := mfa.HOTPCode(key, 0, mfa.AlgSHA1, 0); err == nil {
		t.Error("HOTPCode: expected error for digits=0")
	}
	if _, err := mfa.HOTPCode(key, 0, mfa.AlgSHA1, 11); err == nil {
		t.Error("HOTPCode: expected error for digits=11 (uint32 truncation overflow)")
	}
}

func TestHOTP_RejectsUnknownAlgorithm(t *testing.T) {
	key := []byte("12345678901234567890")
	if _, err := mfa.HOTPCode(key, 0, mfa.Algorithm("md5"), 6); err == nil {
		t.Error("HOTPCode: expected error for unsupported algorithm")
	}
}

func TestHOTP_RejectsEmptyKey(t *testing.T) {
	if _, err := mfa.HOTPCode(nil, 0, mfa.AlgSHA1, 6); err == nil {
		t.Error("HOTPCode: expected error for empty key")
	}
}

// TestHOTP_DifferentAlgorithmsProduceDifferentCodes is a smoke test that SHA256
// and SHA512 truncation paths actually execute distinctly from SHA1 (they are
// not RFC-vectored the way SHA1 is since RFC 4226 only defines SHA-1, but RFC
// 6238 §5.2 extends the construction to SHA-256/SHA-512 with the same
// truncation algorithm).
func TestHOTP_DifferentAlgorithmsProduceDifferentCodes(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes, long enough for SHA512 test vectors elsewhere
	c1, err := mfa.HOTPCode(key, 42, mfa.AlgSHA1, 8)
	if err != nil {
		t.Fatalf("HOTPCode SHA1: %v", err)
	}
	c256, err := mfa.HOTPCode(key, 42, mfa.AlgSHA256, 8)
	if err != nil {
		t.Fatalf("HOTPCode SHA256: %v", err)
	}
	c512, err := mfa.HOTPCode(key, 42, mfa.AlgSHA512, 8)
	if err != nil {
		t.Fatalf("HOTPCode SHA512: %v", err)
	}
	if c1 == c256 || c1 == c512 || c256 == c512 {
		t.Errorf("expected distinct codes per algorithm, got sha1=%s sha256=%s sha512=%s", c1, c256, c512)
	}
	for _, c := range []string{c1, c256, c512} {
		if len(c) != 8 {
			t.Errorf("code %q length = %d, want 8", c, len(c))
		}
	}
}
