package webhook_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/foundation/webhook"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// =============================================================================
// H-9: SEC-03's adversarial tamper matrix requires body, timestamp, event-ID,
// key-ID, and signature-version to each independently fail verification when
// manipulated (docs/implementation/architecture-directive-2026-07-11.md
// SEC-03 AC). TestIntegrationHandleInbound_BadSignature covers body,
// TestIntegrationHandleInbound_TimestampOutOfWindow covers timestamp, and
// TestIntegrationHandleInbound_Replay/_IdlessDedup cover event-ID — this file
// adds the two remaining fields: key-ID and signature-version.
//
// HMACVerifier's body-only scheme (webhook/verifier.go) does not authenticate
// a key id or signature version at all, so it cannot exercise this part of
// the matrix. keyedVerifier is a test-local Verifier — the same "swap in a
// purpose-built Verifier" technique webhook_test.go already uses for
// envelopeVerifier's timestamp case — that signs "<key_id>.<sig_version>.
// <body>", mirroring a realistic provider scheme that binds its signature to
// a specific key id and signature version (guarding against key-confusion and
// signature-downgrade attacks). Tampering either field without re-signing
// must be rejected exactly like a tampered body.
// =============================================================================

const testKeyedProviderKey = "test-keyed-provider"

// keyedVerifier authenticates "<key_id>.<sig_version>.<body>" via
// HMAC-SHA256, read from the X-Key-Id / X-Signature-Version / X-Signature
// headers.
type keyedVerifier struct{}

func (keyedVerifier) Verify(secret string, body []byte, headers map[string]string) (webhook.Envelope, error) {
	keyID := headers["X-Key-Id"]
	sigVersion := headers["X-Signature-Version"]
	got := headers["X-Signature"]
	if keyID == "" || sigVersion == "" || got == "" {
		return webhook.Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_missing", "keyed webhook signature headers absent")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(keyID + "." + sigVersion + "."))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(got), []byte(want)) {
		return webhook.Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_mismatch", "keyed webhook signature does not match")
	}

	sum := sha256.Sum256(body)
	return webhook.Envelope{
		CanonicalBody:    body,
		EventID:          "sha256:" + hex.EncodeToString(sum[:]),
		OccurredAt:       time.Now(),
		SignatureVersion: sigVersion,
		KeyID:            keyID,
	}, nil
}

// testSignKeyed computes a valid keyedVerifier signature for (keyID,
// sigVersion, body).
func testSignKeyed(keyID, sigVersion string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(keyID + "." + sigVersion + "."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// TestIntegrationHandleInbound_TamperedKeyID proves the key-ID field of the
// tamper matrix: a request whose X-Key-Id has been swapped to a different
// key after signing (the X-Signature still reflects the ORIGINAL key id) is
// rejected — the signature no longer matches "<new_key_id>.<sig_version>.
// <body>". No attacker-controlled key id may bypass verification.
func TestIntegrationHandleInbound_TamperedKeyID(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})
	svc.RegisterVerifier(testKeyedProviderKey, keyedVerifier{})

	body := []byte(`{"event":"order.created"}`)
	sig := testSignKeyed("key-1", "v1", body)

	in := webhook.InboundIn{
		EndpointID:  epID,
		ProviderKey: testKeyedProviderKey,
		RawBody:     body,
		Headers: map[string]string{
			"X-Signature":         sig,
			"X-Key-Id":            "key-2", // tampered: signed for key-1
			"X-Signature-Version": "v1",
		},
		ExternalEventID: "ext-keyid-tamper",
		EventType:       "order.created",
		Timestamp:       time.Now(),
	}

	var tamperErr error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		tamperErr = svc.HandleInbound(ctx, db, in)
		return nil // commit so the best-effort audit row persists
	}); cerr != nil {
		t.Fatalf("tx commit: %v", cerr)
	}
	if kerr.KindOf(tamperErr) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated for tampered key-ID, got kind=%v err=%v", kerr.KindOf(tamperErr), tamperErr)
	}
	if n := countEvents(t, h, tn.ID); n != 1 {
		t.Fatalf("want 1 audit row, got %d", n)
	}
	ok := eventSigOk(t, h, tn.ID)
	if ok == nil || *ok {
		t.Fatalf("want signature_ok=false for tampered key-ID, got %v", ok)
	}
}

// TestIntegrationHandleInbound_TamperedSignatureVersion proves the
// signature-version field of the tamper matrix: a request whose
// X-Signature-Version has been downgraded/altered after signing is rejected
// — the signature no longer matches "<key_id>.<new_sig_version>.<body>".
// This guards against a signature-downgrade attack where a caller claims a
// weaker/different scheme version than what was actually signed.
func TestIntegrationHandleInbound_TamperedSignatureVersion(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})
	svc.RegisterVerifier(testKeyedProviderKey, keyedVerifier{})

	body := []byte(`{"event":"order.created"}`)
	sig := testSignKeyed("key-1", "v2", body)

	in := webhook.InboundIn{
		EndpointID:  epID,
		ProviderKey: testKeyedProviderKey,
		RawBody:     body,
		Headers: map[string]string{
			"X-Signature":         sig,
			"X-Key-Id":            "key-1",
			"X-Signature-Version": "v1", // tampered: signed for v2
		},
		ExternalEventID: "ext-sigver-tamper",
		EventType:       "order.created",
		Timestamp:       time.Now(),
	}

	var tamperErr error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		tamperErr = svc.HandleInbound(ctx, db, in)
		return nil // commit so the best-effort audit row persists
	}); cerr != nil {
		t.Fatalf("tx commit: %v", cerr)
	}
	if kerr.KindOf(tamperErr) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated for tampered signature-version, got kind=%v err=%v", kerr.KindOf(tamperErr), tamperErr)
	}
	if n := countEvents(t, h, tn.ID); n != 1 {
		t.Fatalf("want 1 audit row, got %d", n)
	}
	ok := eventSigOk(t, h, tn.ID)
	if ok == nil || *ok {
		t.Fatalf("want signature_ok=false for tampered signature-version, got %v", ok)
	}
}
