package mfa

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

// Default TOTP parameters (RFC 6238's own defaults, and the common
// authenticator-app convention): 30s step, 6 digits, SHA-1, ±1 step of
// clock-skew tolerance on verify.
const (
	DefaultTOTPStep   = 30 * time.Second
	DefaultTOTPDigits = 6
	DefaultTOTPSkew   = 1
)

// DefaultTOTPSecretLen is the recommended HOTP/TOTP secret length in bytes
// (160 bits, RFC 4226's recommended key size).
const DefaultTOTPSecretLen = 20

// TOTPOptions configures TOTP code generation/verification. A zero-valued
// TOTPOptions resolves to the package defaults (30s/6 digits/SHA1/skew=1) via
// withDefaults — callers only need to set the fields they want to override.
type TOTPOptions struct {
	// Step is the time-step duration (RFC 6238 calls this X). Zero resolves
	// to DefaultTOTPStep.
	Step time.Duration
	// Digits is the number of decimal digits in the code. Zero resolves to
	// DefaultTOTPDigits.
	Digits int
	// Algorithm is the HMAC hash. Empty resolves to AlgSHA1.
	Algorithm Algorithm
	// Skew is the number of steps of clock skew tolerated on verify in
	// either direction (0 = exact step only). Negative is rejected.
	Skew int
}

// withDefaults returns a copy of o with zero-valued fields replaced by
// package defaults, and validates the result.
func (o TOTPOptions) withDefaults() (TOTPOptions, error) {
	if o.Step < 0 {
		return o, fmt.Errorf("mfa: TOTP step must not be negative")
	}
	if o.Step == 0 {
		o.Step = DefaultTOTPStep
	}
	if o.Digits == 0 {
		o.Digits = DefaultTOTPDigits
	}
	if o.Digits < 1 || o.Digits > maxDigits {
		return o, fmt.Errorf("mfa: unsupported digit count %d (must be 1..%d)", o.Digits, maxDigits)
	}
	if o.Algorithm == "" {
		o.Algorithm = AlgSHA1
	}
	if _, err := newHash(o.Algorithm); err != nil {
		return o, err
	}
	if o.Skew < 0 {
		return o, fmt.Errorf("mfa: TOTP skew must not be negative")
	}
	return o, nil
}

// GenerateTOTPSecret returns n cryptographically random bytes suitable for
// use as a TOTP/HOTP shared secret (crypto/rand). n must be positive; 20
// (DefaultTOTPSecretLen) matches RFC 4226's recommended 160-bit key.
func GenerateTOTPSecret(n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("mfa: secret length must be positive, got %d", n)
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("mfa: generate secret: %w", err)
	}
	return buf, nil
}

// base32Enc is the unpadded standard base32 alphabet used by authenticator
// apps (e.g. otpauth:// URIs) to encode TOTP secrets as text.
var base32Enc = base32.StdEncoding.WithPadding(base32.NoPadding)

// EncodeSecretBase32 renders secret as unpadded base32 text, the conventional
// wire form for authenticator-app provisioning (otpauth:// URIs, QR codes).
func EncodeSecretBase32(secret []byte) (string, error) {
	if len(secret) == 0 {
		return "", fmt.Errorf("mfa: secret must not be empty")
	}
	return base32Enc.EncodeToString(secret), nil
}

// DecodeSecretBase32 parses a base32-encoded secret. It is case-insensitive
// and tolerates surrounding whitespace (common when a user copy-pastes a
// secret), matching how authenticator apps typically accept manual entry.
func DecodeSecretBase32(s string) ([]byte, error) {
	clean := strings.ToUpper(strings.TrimSpace(s))
	clean = strings.ReplaceAll(clean, " ", "")
	key, err := base32Enc.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("mfa: malformed base32 secret: %w", err)
	}
	return key, nil
}

// TOTPCodeAt computes the RFC 6238 TOTP value for secret at instant t under
// opts (zero-valued fields resolve to defaults).
func TOTPCodeAt(secret []byte, t time.Time, opts TOTPOptions) (string, error) {
	o, err := opts.withDefaults()
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix() / int64(o.Step/time.Second)) // #nosec G115 -- RFC 6238 counter: t.Unix() is non-negative for any real clock (post-1970), and Step is validated positive by withDefaults
	return HOTPCode(secret, counter, o.Algorithm, o.Digits)
}

// VerifyTOTPAt reports whether code is valid for secret at instant t,
// tolerating ±opts.Skew steps of clock skew (zero-valued fields resolve to
// defaults). The comparison against each candidate code is constant-time
// (crypto/subtle), so verification timing does not leak how close an
// incorrect guess was.
func VerifyTOTPAt(secret []byte, code string, t time.Time, opts TOTPOptions) (bool, error) {
	o, err := opts.withDefaults()
	if err != nil {
		return false, err
	}
	step := int64(o.Step / time.Second)
	counter := t.Unix() / step
	codeBytes := []byte(code)
	for delta := -o.Skew; delta <= o.Skew; delta++ {
		c := counter + int64(delta)
		if c < 0 {
			continue
		}
		want, err := HOTPCode(secret, uint64(c), o.Algorithm, o.Digits)
		if err != nil {
			return false, err
		}
		if len(want) == len(codeBytes) && subtle.ConstantTimeCompare([]byte(want), codeBytes) == 1 {
			return true, nil
		}
	}
	return false, nil
}
