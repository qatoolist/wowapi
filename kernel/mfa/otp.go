package mfa

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// DefaultOTPDigits is the conventional length of a delivered (SMS/email)
// one-time code — shorter than a TOTP code since it is single-use,
// short-lived, and attempt-limited rather than continuously re-derived.
const DefaultOTPDigits = 6

// GenerateOTPCode returns a cryptographically random (crypto/rand) numeric
// code of exactly digits decimal characters, left-zero-padded. digits must
// be in [1,10] (bounded by the same uint32 truncation limit as HOTPCode).
func GenerateOTPCode(digits int) (string, error) {
	if digits < 1 || digits > maxDigits {
		return "", fmt.Errorf("mfa: unsupported digit count %d (must be 1..%d)", digits, maxDigits)
	}
	mod := uint32(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("mfa: generate OTP code: %w", err)
	}
	n := binary.BigEndian.Uint32(buf[:]) % mod
	return fmt.Sprintf("%0*d", digits, n), nil
}

// HashOTPCode returns a salted SHA-256 hash of code, hex-encoded. salt should
// be a value unique to the challenge (e.g. the challenge's random ID) so an
// offline attacker cannot precompute a rainbow table across challenges; it is
// not a substitute for TTL + attempt-limit enforcement (see ChallengePolicy),
// which is what actually bounds brute-force exposure for a short numeric
// code space.
func HashOTPCode(salt, code string) string {
	sum := sha256.Sum256([]byte(salt + ":" + code))
	return hex.EncodeToString(sum[:])
}

// VerifyOTPCode reports whether code, salted the same way, matches wantHash.
// The comparison is constant-time (crypto/subtle) to avoid leaking
// hash-prefix-match timing to an attacker probing over the network.
func VerifyOTPCode(salt, code, wantHash string) bool {
	got := HashOTPCode(salt, code)
	return len(got) == len(wantHash) && subtle.ConstantTimeCompare([]byte(got), []byte(wantHash)) == 1
}
