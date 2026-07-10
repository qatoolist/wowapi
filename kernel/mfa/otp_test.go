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
		t.Error("digits=11 accepted (exceeds the 10-digit supported range)")
	}
}

// TestOTP_MaxDigitsSpansFullRange is a regression test for the same class of
// modulus-overflow bug fixed in HOTPCode: a uint32 accumulator for 10^10
// silently wraps to ~1.41e9, capping every digits=10 code well below
// 10,000,000,000 without an error. It draws enough digits=10 codes that, if
// the modulus were still computed in uint32, none would ever reach the
// wrapped ceiling — so seeing one prove the true (uint64) range is in use.
func TestOTP_MaxDigitsSpansFullRange(t *testing.T) {
	const wrappedUint32Modulus = 1410065408 // 10^10 mod 2^32, the bug's effective ceiling
	found := false
	for i := 0; i < 2000; i++ {
		code, err := mfa.GenerateOTPCode(10)
		if err != nil {
			t.Fatalf("GenerateOTPCode(10): %v", err)
		}
		if len(code) != 10 {
			t.Fatalf("GenerateOTPCode(10) length = %d, want 10", len(code))
		}
		var asInt uint64
		for _, c := range code {
			asInt = asInt*10 + uint64(c-'0')
		}
		if asInt >= wrappedUint32Modulus {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no digits=10 code in 2000 draws reached the range a uint32-modulus bug would have made " +
			"unreachable (~1-in-3 odds per draw if fixed) — either the regression reappeared or this is a " +
			"1-in-10^600ish fluke; re-run before assuming regression")
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
