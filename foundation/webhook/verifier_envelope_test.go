package webhook_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/foundation/webhook"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/testkit/fakes"
)

func timestampedSignature(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestUnitTimestampedHMACVerifierAuthenticatesTimestampAndBody(t *testing.T) {
	body := []byte(`{"event":"order.created"}`)
	ts := "1700000000"
	v := webhook.TimestampedHMACVerifier{}
	env, err := v.Verify(testSecret, body, map[string]string{
		"X-Timestamp": ts,
		"X-Signature": timestampedSignature(testSecret, ts, body),
	})
	if err != nil {
		t.Fatalf("verify timestamped HMAC: %v", err)
	}
	if got, want := env.OccurredAt, time.Unix(1700000000, 0).UTC(); !got.Equal(want) {
		t.Fatalf("OccurredAt = %v, want %v", got, want)
	}
	if env.SignatureVersion != "sha256-timestamped" || env.EventID == "" {
		t.Fatalf("incomplete authenticated envelope: %+v", env)
	}

	for name, headers := range map[string]map[string]string{
		"forged timestamp": {
			"X-Timestamp": "1700000001",
			"X-Signature": timestampedSignature(testSecret, ts, body),
		},
		"forged body": {
			"X-Timestamp": ts,
			"X-Signature": timestampedSignature(testSecret, ts, body),
		},
		"malformed timestamp": {
			"X-Timestamp": "not-unix-seconds",
			"X-Signature": timestampedSignature(testSecret, "not-unix-seconds", body),
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidateBody := body
			if name == "forged body" {
				candidateBody = []byte(`{"event":"order.cancelled"}`)
			}
			if _, err := v.Verify(testSecret, candidateBody, headers); kerr.KindOf(err) != kerr.KindUnauthenticated {
				t.Fatalf("tampered timestamped envelope = %v, want unauthenticated", err)
			}
		})
	}
}

// TestUnitHMACVerifier_Envelope proves HMACVerifier returns an Envelope whose
// fields are derived from the authenticated body and receipt time only.
func TestUnitHMACVerifier_Envelope(t *testing.T) {
	body := []byte(`{"event":"order.created","id":"provider-123"}`)
	v := webhook.HMACVerifier{}

	env, err := v.Verify(testSecret, body, map[string]string{"X-Signature": testSign(body)})
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if string(env.CanonicalBody) != string(body) {
		t.Fatalf("CanonicalBody mismatch: got %q, want %q", env.CanonicalBody, body)
	}

	wantID := "sha256:" + hex.EncodeToString(sha256sum(body))
	if env.EventID != wantID {
		t.Fatalf("EventID mismatch: got %q, want %q", env.EventID, wantID)
	}

	if env.SignatureVersion != "sha256" {
		t.Fatalf("want SignatureVersion=sha256, got %q", env.SignatureVersion)
	}

	if env.KeyID != "" {
		t.Fatalf("want empty KeyID for body-only HMAC, got %q", env.KeyID)
	}

	// OccurredAt must be a recent receipt time, not zero.
	if env.OccurredAt.IsZero() {
		t.Fatal("OccurredAt must not be zero")
	}
	if delta := time.Since(env.OccurredAt); delta > time.Second {
		t.Fatalf("OccurredAt too old: %v", delta)
	}
}

// TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader proves the verifier's
// receipt-time synthesis is independent of any timestamp header the caller
// supplies. A body-only verifier cannot attest to caller-supplied time.
func TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader(t *testing.T) {
	body := []byte(`{"event":"order.created"}`)
	v := webhook.HMACVerifier{}
	headers := map[string]string{
		"X-Signature":         testSign(body),
		"X-Timestamp":         "0",
		"X-Event-Id":          "evt-attacker",
		"X-Signature-Version": "v99",
	}

	env, err := v.Verify(testSecret, body, headers)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if env.OccurredAt.IsZero() {
		t.Fatal("OccurredAt must not be zero")
	}
	if delta := time.Since(env.OccurredAt); delta > time.Second {
		t.Fatalf("OccurredAt too old: %v", delta)
	}
}

// TestUnitHMACVerifier_BadSignature proves a mismatched signature returns
// KindUnauthenticated and an undefined (zero) Envelope.
func TestUnitHMACVerifier_BadSignature(t *testing.T) {
	body := []byte(`{"event":"order.created"}`)
	v := webhook.HMACVerifier{}
	env, err := v.Verify(testSecret, body, map[string]string{"X-Signature": "sha256=badhex"})
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated, got %v", err)
	}
	if env.EventID != "" || !env.OccurredAt.IsZero() || env.SignatureVersion != "" {
		t.Fatalf("error Envelope must be zero-valued, got %+v", env)
	}
}

// TestUnitFakeVerifier_Envelope proves FakeVerifier returns a valid Envelope
// on success and a zero Envelope on failure.
func TestUnitFakeVerifier_Envelope(t *testing.T) {
	body := []byte(`{"event":"order.created"}`)
	v := fakes.WebhookVerifier{Secret: "open-sesame"}

	env, err := v.Verify("", body, map[string]string{"X-Test-Sig": "open-sesame"})
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if string(env.CanonicalBody) != string(body) {
		t.Fatalf("CanonicalBody mismatch: got %q, want %q", env.CanonicalBody, body)
	}

	wantID := "sha256:" + hex.EncodeToString(sha256sum(body))
	if env.EventID != wantID {
		t.Fatalf("EventID mismatch: got %q, want %q", env.EventID, wantID)
	}

	if env.SignatureVersion != "test" {
		t.Fatalf("want SignatureVersion=test, got %q", env.SignatureVersion)
	}

	if env.OccurredAt.IsZero() {
		t.Fatal("OccurredAt must not be zero")
	}

	_, err = v.Verify("", body, map[string]string{"X-Test-Sig": "wrong"})
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated on mismatch, got %v", err)
	}
}

func sha256sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
