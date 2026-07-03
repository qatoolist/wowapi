package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// HMACVerifier implements Verifier using HMAC-SHA256. The expected signature
// is read from the header named by SignatureHeader (default "X-Signature"),
// which may carry a "sha256=" prefix that is stripped before comparison.
// The signature is the lowercase-hex HMAC-SHA256 of the raw request body
// keyed by the endpoint secret.
//
// NOTE: this verifies the common EXTERNAL-provider scheme — HMAC over the body
// alone. It is intentionally NOT the same construction as our OUTBOUND signing
// (signPayload in service.go), which authenticates "<timestamp>.<body>" so the
// X-Timestamp header is covered (SEC-52). A provider that signs a timestamped
// payload needs its own Verifier registered under its provider key.
type HMACVerifier struct {
	// SignatureHeader is the header name carrying the signature.
	// Defaults to "X-Signature" when empty.
	SignatureHeader string
}

// Verify computes HMAC-SHA256(secret, body) and compares it to the value in
// SignatureHeader using a constant-time comparison. Returns KindUnauthenticated
// on mismatch or missing header.
func (v HMACVerifier) Verify(secret string, body []byte, headers map[string]string) error {
	header := v.SignatureHeader
	if header == "" {
		header = "X-Signature"
	}
	// Header lookup is case-insensitive.
	var got string
	for k, val := range headers {
		if strings.EqualFold(k, header) {
			got = val
			break
		}
	}
	if got == "" {
		return kerr.E(kerr.KindUnauthenticated, "signature_missing", "webhook signature header absent")
	}
	// Strip "sha256=" prefix if present.
	got = strings.TrimPrefix(got, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(got), []byte(want)) {
		return kerr.E(kerr.KindUnauthenticated, "signature_mismatch", "webhook HMAC-SHA256 signature does not match")
	}
	return nil
}

// FakeVerifier is a test double that passes when the header "X-Test-Sig" equals
// the pre-configured Secret, and fails otherwise.
type FakeVerifier struct {
	// Secret is the expected value in the "X-Test-Sig" header.
	Secret string
}

// Verify passes when headers["X-Test-Sig"] == v.Secret, fails otherwise.
func (v FakeVerifier) Verify(_ string, _ []byte, headers map[string]string) error {
	if headers["X-Test-Sig"] == v.Secret {
		return nil
	}
	return kerr.E(kerr.KindUnauthenticated, "signature_mismatch", "fake verifier: signature does not match")
}
