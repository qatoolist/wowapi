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
// be in [1,10] (bounded by the same limit as HOTPCode's dynamic-truncation
// output width).
//
// The modulus and the random draw are both computed in uint64: digits=10
// requires a modulus of 10^10 (10000000000), which exceeds uint32's range
// (max 4294967295) and would silently wrap to a much smaller modulus if
// accumulated in 32 bits — the same class of bug this package's HOTPCode
// once had. 8 random bytes read into a uint64 give ~64 bits of entropy,
// comfortably enough to reduce mod 10^10 without the last-bucket modulo bias
// a narrower read would introduce.
func GenerateOTPCode(digits int) (string, error) {
	if digits < 1 || digits > maxDigits {
		return "", fmt.Errorf("mfa: unsupported digit count %d (must be 1..%d)", digits, maxDigits)
	}
	mod := uint64(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("mfa: generate OTP code: %w", err)
	}
	n := binary.BigEndian.Uint64(buf[:]) % mod
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
