package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Envelope carries the fields a Verifier derives from authenticated data.
// Every field must come from data covered by the provider's signature; no
// caller-supplied request field may be surfaced here. On Verify error the
// Envelope is undefined and callers must not read it.
type Envelope struct {
	// CanonicalBody is the byte sequence the signature actually authenticated.
	// For body-only HMAC schemes this is the raw request body; for schemes that
	// sign a canonicalized envelope it is the canonical form.
	CanonicalBody []byte

	// EventID is a stable identifier for this event. It must be derived from
	// authenticated data so that replay/dedup decisions cannot be influenced by
	// an attacker-supplied id.
	EventID string

	// OccurredAt is the event timestamp the verifier is willing to attest to.
	// For schemes that authenticate a provider-asserted timestamp this is that
	// timestamp; for schemes that do not authenticate a timestamp it must be a
	// locally-generated receipt time.
	OccurredAt time.Time

	// SignatureVersion is the signature scheme version (e.g. "sha256"). It may
	// be empty for schemes that do not version their signatures.
	SignatureVersion string

	// KeyID is the identifier of the key used to verify the signature, when the
	// scheme authenticates it. For schemes that do not authenticate a key id it
	// is empty.
	KeyID string
}

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

// TimestampedHMACVerifier implements the framework's timestamped HMAC wire
// contract. It authenticates the exact byte sequence "<unix-seconds>.<body>"
// using HMAC-SHA256, matching outbound delivery. The timestamp is therefore
// verifier-attested and can safely drive HandleInbound's replay window.
type TimestampedHMACVerifier struct {
	SignatureHeader string
	TimestampHeader string
}

// Verify validates a lowercase hexadecimal HMAC (an optional "sha256=" prefix
// is accepted), strictly parses the authenticated Unix timestamp, and derives
// a stable event identity from the complete authenticated message.
func (v TimestampedHMACVerifier) Verify(secret string, body []byte, headers map[string]string) (Envelope, error) {
	sigHeader := v.SignatureHeader
	if sigHeader == "" {
		sigHeader = "X-Signature"
	}
	tsHeader := v.TimestampHeader
	if tsHeader == "" {
		tsHeader = "X-Timestamp"
	}
	got := headerValue(headers, sigHeader)
	if got == "" {
		return Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_missing", "webhook signature header absent")
	}
	ts := headerValue(headers, tsHeader)
	if ts == "" {
		return Envelope{}, kerr.E(kerr.KindUnauthenticated, "timestamp_missing", "webhook timestamp header absent")
	}
	seconds, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return Envelope{}, kerr.E(kerr.KindUnauthenticated, "timestamp_invalid", "webhook timestamp must be Unix seconds")
	}
	got = strings.TrimPrefix(got, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts + "."))
	_, _ = mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(got), []byte(want)) {
		return Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_mismatch", "webhook timestamped HMAC-SHA256 signature does not match")
	}
	identity := sha256.New()
	_, _ = identity.Write([]byte(ts + "."))
	_, _ = identity.Write(body)
	return Envelope{
		CanonicalBody:    body,
		EventID:          "sha256:" + hex.EncodeToString(identity.Sum(nil)),
		OccurredAt:       time.Unix(seconds, 0).UTC(),
		SignatureVersion: "sha256-timestamped",
	}, nil
}

func headerValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

// Verify computes HMAC-SHA256(secret, body) and compares it to the value in
// SignatureHeader using a constant-time comparison. On success it returns an
// Envelope synthesized from the authenticated body and receipt time. On
// mismatch or missing header it returns KindUnauthenticated.
//
// Because this verifier authenticates the body only, it cannot attest to a
// provider-asserted timestamp or event id. OccurredAt is set to the local
// receipt time and EventID is a stable hash of the body. This makes
// HMACVerifier unsuitable for provider protocols that require a
// provider-asserted timestamp; such protocols need their own Verifier.
func (v HMACVerifier) Verify(secret string, body []byte, headers map[string]string) (Envelope, error) {
	header := v.SignatureHeader
	if header == "" {
		header = "X-Signature"
	}
	// Header lookup is case-insensitive.
	got := headerValue(headers, header)
	if got == "" {
		return Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_missing", "webhook signature header absent")
	}
	// Strip "sha256=" prefix if present.
	got = strings.TrimPrefix(got, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(got), []byte(want)) {
		return Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_mismatch", "webhook HMAC-SHA256 signature does not match")
	}

	// Synthesize envelope from authenticated body and receipt time only.
	// The body is the only authenticated datum, so we cannot surface any
	// provider-asserted timestamp or id.
	sum := sha256.Sum256(body)
	return Envelope{
		CanonicalBody:    body,
		EventID:          "sha256:" + hex.EncodeToString(sum[:]),
		OccurredAt:       time.Now(),
		SignatureVersion: "sha256",
		KeyID:            "", // body-only HMAC does not authenticate a key id
	}, nil
}
