package mfa

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // RFC 4226/6238 mandate SHA-1 as the default HOTP/TOTP algorithm
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"hash"
)

// Algorithm is the HMAC hash function underlying an HOTP/TOTP code. RFC 4226
// defines SHA-1; RFC 6238 §5.2 extends the same construction to SHA-256 and
// SHA-512 for TOTP. This is a closed set — an unrecognized value is rejected
// rather than silently falling back to a default, since silently using the
// wrong algorithm produces a code that verifies against nothing.
type Algorithm string

const (
	AlgSHA1   Algorithm = "SHA1"
	AlgSHA256 Algorithm = "SHA256"
	AlgSHA512 Algorithm = "SHA512"
)

// newHash returns the constructor for alg, or an error for an unrecognized
// algorithm.
func newHash(alg Algorithm) (func() hash.Hash, error) {
	switch alg {
	case AlgSHA1:
		return sha1.New, nil
	case AlgSHA256:
		return sha256.New, nil
	case AlgSHA512:
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("mfa: unsupported algorithm %q", alg)
	}
}

// maxDigits is the largest digit count HOTPCode supports: the RFC 4226 §5.3
// dynamic-truncation binary code is a 31-bit unsigned value (max
// 2147483647), which has 10 decimal digits, so an 11-digit request can never
// be satisfied by the algorithm and is rejected rather than silently
// producing a code with structurally short leading-zero padding for a
// digit count the math cannot actually fill.
const maxDigits = 10

// HOTPCode computes the RFC 4226 HOTP value for key and counter using the
// given algorithm, truncated to digits decimal digits (left-zero-padded).
// digits must be in [1,10]; key must be non-empty.
func HOTPCode(key []byte, counter uint64, alg Algorithm, digits int) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("mfa: HOTP key must not be empty")
	}
	if digits < 1 || digits > maxDigits {
		return "", fmt.Errorf("mfa: unsupported digit count %d (must be 1..%d)", digits, maxDigits)
	}
	hf, err := newHash(alg)
	if err != nil {
		return "", err
	}

	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)

	mac := hmac.New(hf, key)
	mac.Write(counterBytes[:])
	sum := mac.Sum(nil)

	// RFC 4226 §5.3 dynamic truncation.
	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)

	mod := uint32(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	code := binCode % mod
	return fmt.Sprintf("%0*d", digits, code), nil
}
