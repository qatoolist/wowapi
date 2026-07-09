package mfa_test

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/mfa"
)

func TestOTP_GenerateCode_Length(t *testing.T) {
	for _, digits := range []int{4, 6, 8, 10} {
		code, err := mfa.GenerateOTPCode(digits)
		if err != nil {
			t.Fatalf("GenerateOTPCode(%d): %v", digits, err)
		}
		if len(code) != digits {
			t.Errorf("GenerateOTPCode(%d) length = %d, want %d", digits, len(code), digits)
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Fatalf("GenerateOTPCode(%d) = %q, contains non-digit", digits, code)
			}
		}
	}
}

func TestOTP_GenerateCode_RejectsInvalidDigits(t *testing.T) {
	if _, err := mfa.GenerateOTPCode(0); err == nil {
		t.Error("digits=0 accepted")
	}
	if _, err := mfa.GenerateOTPCode(-1); err == nil {
		t.Error("digits=-1 accepted")
	}
	if _, err := mfa.GenerateOTPCode(11); err == nil {
		t.Error("digits=11 accepted (exceeds uint32 range)")
	}
}

func TestOTP_GenerateCode_Randomness(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		code, err := mfa.GenerateOTPCode(6)
		if err != nil {
			t.Fatalf("GenerateOTPCode: %v", err)
		}
		seen[code] = true
	}
	if len(seen) < 40 { // extremely generous floor; true collisions are ~1e-6
		t.Errorf("only %d unique codes out of 50 draws, suspiciously low entropy", len(seen))
	}
}

func TestOTP_HashAndVerify_RoundTrip(t *testing.T) {
	salt := "challenge-id-123"
	code := "482913"
	hash := mfa.HashOTPCode(salt, code)
	if hash == "" {
		t.Fatal("HashOTPCode returned empty hash")
	}
	if !mfa.VerifyOTPCode(salt, code, hash) {
		t.Error("VerifyOTPCode: correct code+salt rejected")
	}
}

func TestOTP_HashAndVerify_WrongCodeRejected(t *testing.T) {
	salt := "challenge-id-123"
	hash := mfa.HashOTPCode(salt, "482913")
	if mfa.VerifyOTPCode(salt, "482914", hash) {
		t.Error("VerifyOTPCode: wrong code accepted")
	}
}

func TestOTP_HashAndVerify_WrongSaltRejected(t *testing.T) {
	hash := mfa.HashOTPCode("salt-a", "482913")
	if mfa.VerifyOTPCode("salt-b", "482913", hash) {
		t.Error("VerifyOTPCode: wrong salt accepted (hash not actually salted)")
	}
}

func TestOTP_Hash_DifferentSaltsProduceDifferentHashes(t *testing.T) {
	h1 := mfa.HashOTPCode("salt-a", "482913")
	h2 := mfa.HashOTPCode("salt-b", "482913")
	if h1 == h2 {
		t.Error("same code with different salts produced identical hashes")
	}
}

func TestOTP_VerifyOTPCode_LengthMismatchRejected(t *testing.T) {
	hash := mfa.HashOTPCode("salt", "482913")
	if mfa.VerifyOTPCode("salt", "482913", hash[:len(hash)-4]) {
		t.Error("VerifyOTPCode: truncated hash incorrectly verified")
	}
}
