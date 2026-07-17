package testkit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/auth"
)

// TestIssueVerifyRoundTrip proves a token minted by TokenIssuer verifies under
// an auth.Verifier wired to its KeySource — the fixture contract every
// authenticated test relies on.
func TestIssueVerifyRoundTrip(t *testing.T) {
	ti := NewTokenIssuer()
	v := auth.NewVerifier(ti.KeySource(), auth.Config{Issuer: "wowapi-test", Audience: "wowapi"})

	tenantID := uuid.New()
	capID := uuid.New()
	tok := ti.Issue("idp|round-trip", tenantID, capID)

	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.Subject() != "idp|round-trip" {
		t.Fatalf("subject: got %q", claims.Subject())
	}
	if claims.TenantID != tenantID {
		t.Fatalf("tenant: got %v want %v", claims.TenantID, tenantID)
	}
	if claims.CapacityID != capID {
		t.Fatalf("capacity: got %v want %v", claims.CapacityID, capID)
	}
}
