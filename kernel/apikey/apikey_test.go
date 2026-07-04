package apikey_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/apikey"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

// Compile-time check: the authenticator satisfies the httpx gate's port.
var _ httpx.Authenticator = (*apikey.Authenticator)(nil)

func issue(t *testing.T, h *testkit.DBHandle, s *apikey.Store, tenant uuid.UUID, name string, scopes []string, exp *time.Time) string {
	t.Helper()
	var token string
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		token, _, e = s.Issue(ctx, db, name, scopes, exp)
		return e
	}); err != nil {
		t.Fatalf("issue: %v", err)
	}
	return token
}

func TestIntegrationApiKeyIssueVerify(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := uuid.New()

	token := issue(t, h, s, tenant, "gate-1", []string{"gate.device.read"}, nil)

	p, err := s.Verify(context.Background(), h.PlatformTxM, token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if p.TenantID != tenant || p.Name != "gate-1" {
		t.Fatalf("principal = %+v, want tenant %s / gate-1", p, tenant)
	}
	if len(p.Scopes) != 1 || p.Scopes[0] != "gate.device.read" {
		t.Fatalf("scopes = %v, want [gate.device.read]", p.Scopes)
	}
}

func TestIntegrationApiKeyWrongSecretDenied(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	token := issue(t, h, s, uuid.New(), "k", nil, nil)

	// Corrupt the secret portion.
	tampered := token + "x"
	if _, err := s.Verify(context.Background(), h.PlatformTxM, tampered); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("wrong secret must be unauthenticated, got %v", err)
	}
	// Unknown prefix.
	if _, err := s.Verify(context.Background(), h.PlatformTxM, "wowapi_deadbeef_nope"); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("unknown key must be unauthenticated, got %v", err)
	}
	// Malformed token.
	if _, err := s.Verify(context.Background(), h.PlatformTxM, "not-a-key"); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("malformed token must be unauthenticated, got %v", err)
	}
}

func TestIntegrationApiKeyRevoked(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := uuid.New()

	var token string
	var id uuid.UUID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		token, id, e = s.Issue(ctx, db, "k", []string{"x.y.read"}, nil)
		return e
	}); err != nil {
		t.Fatal(err)
	}
	// Works before revoke.
	if _, err := s.Verify(context.Background(), h.PlatformTxM, token); err != nil {
		t.Fatalf("pre-revoke verify: %v", err)
	}
	// Revoke, then it must fail.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return s.Revoke(ctx, db, id)
	}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := s.Verify(context.Background(), h.PlatformTxM, token); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("revoked key must be unauthenticated, got %v", err)
	}
}

func TestIntegrationApiKeyExpired(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	past := time.Now().Add(-time.Hour)
	token := issue(t, h, s, uuid.New(), "k", nil, &past)

	if _, err := s.Verify(context.Background(), h.PlatformTxM, token); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("expired key must be unauthenticated, got %v", err)
	}
}

func TestIntegrationApiKeyAuthenticator(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := uuid.New()
	token := issue(t, h, s, tenant, "svc", []string{"a.b.read", "a.b.update"}, nil)

	auth := apikey.NewAuthenticator(s, h.PlatformTxM)

	// A valid API key becomes an ActorSystem carrying its tenant + scopes.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	actor, err := auth.Authenticate(req)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if actor.Kind != authz.ActorSystem || actor.TenantID != tenant || len(actor.Scopes) != 2 {
		t.Fatalf("actor = %+v, want ActorSystem/tenant/2 scopes", actor)
	}

	// A non-API-key bearer (e.g. a JWT) is passed over (unauthenticated here) so a
	// composite authenticator can try OIDC.
	jwtReq := httptest.NewRequest(http.MethodGet, "/", nil)
	jwtReq.Header.Set("Authorization", "Bearer eyJhbGciOi.something.sig")
	if _, err := auth.Authenticate(jwtReq); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("non-apikey bearer should be unauthenticated, got %v", err)
	}
}
