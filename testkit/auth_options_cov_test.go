package testkit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/auth"
)

// TestTokenOptionsDriveClaims proves each TokenOption is applied to the minted
// token by verifying observable effects: default claims verify; a wrong issuer
// or audience is rejected; an expired token is rejected; and a grant identity
// round-trips through the verifier.
func TestTokenOptionsDriveClaims(t *testing.T) {
	ti := NewTokenIssuer()
	v := auth.NewVerifier(ti.KeySource(), auth.Config{Issuer: defaultTestIssuer, Audience: defaultTestAudience})
	ctx := context.Background()
	tenantID := uuid.New()
	capID := uuid.New()

	grantID := uuid.New()
	tok := ti.Issue("idp|sub", tenantID, capID, WithGrantID(grantID))
	claims, err := v.Verify(ctx, tok)
	if err != nil {
		t.Fatalf("verify default+options token: %v", err)
	}
	if claims.GrantID != grantID {
		t.Fatalf("grant = %v, want %v", claims.GrantID, grantID)
	}

	// WithIssuer overrides the iss claim → verifier rejects the mismatch.
	if _, err := v.Verify(ctx, ti.Issue("s", tenantID, capID, WithIssuer("someone-else"))); err == nil {
		t.Fatal("token with wrong issuer verified; want rejection")
	}

	// WithAudience overrides the aud claim → verifier rejects the mismatch.
	if _, err := v.Verify(ctx, ti.Issue("s", tenantID, capID, WithAudience("wrong-aud"))); err == nil {
		t.Fatal("token with wrong audience verified; want rejection")
	}

	// WithExpiry negative → already-expired token → verifier rejects.
	if _, err := v.Verify(ctx, ti.Issue("s", tenantID, capID, WithExpiry(-time.Hour))); err == nil {
		t.Fatal("expired token verified; want rejection")
	}
}

// TestTokenIssuerPublicKey covers PublicKey and confirms it is the public half
// of the keypair the KeySource publishes.
func TestTokenIssuerPublicKey(t *testing.T) {
	ti := NewTokenIssuer()
	pub := ti.PublicKey()
	if pub == nil {
		t.Fatal("PublicKey() = nil")
	}
	if pub.N.Sign() == 0 {
		t.Fatal("PublicKey modulus is zero")
	}
	if pub != &ti.key.PublicKey {
		t.Fatal("PublicKey did not return the issuer's own public key")
	}
}
