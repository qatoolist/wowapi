package apikey_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/apikey"
	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// fixedGen is an IDGen that always returns the same UUID, so tests can force a
// primary-key collision on the second insert and drive the DB error branches.
type fixedGen struct{ id uuid.UUID }

func (f fixedGen) New() uuid.UUID { return f.id }

// TestIntegrationApiKeyNewStoreNilIDGen covers NewStore's nil-idgen default:
// passing nil must fall back to the production UUIDv7 generator, and the store
// must still mint a verifiable key.
func TestIntegrationApiKeyNewStoreNilIDGen(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(nil) // exercises the `if idgen == nil` fallback
	if s == nil {
		t.Fatal("NewStore(nil) returned nil")
	}
	tenant := uuid.New()
	token := issue(t, h, s, tenant, "nil-gen", []string{"a.b.read"}, nil)
	p, err := s.Verify(context.Background(), h.PlatformTxM, token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if p.TenantID != tenant || p.KeyID == uuid.Nil {
		t.Fatalf("principal = %+v, want tenant %s with non-nil key id", p, tenant)
	}
}

// TestIntegrationApiKeyIssueValidation covers Issue's name-required guard.
func TestIntegrationApiKeyIssueValidation(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	var token string
	var id uuid.UUID
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		token, id, e = s.Issue(ctx, db, "", []string{"a.b.read"}, nil)
		return e
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty name must be a validation error, got %v", err)
	}
	if token != "" || id != uuid.Nil {
		t.Fatalf("failed Issue must return zero token/id, got %q / %s", token, id)
	}
}

// TestIntegrationApiKeyOnlyHashStored proves the plaintext secret never lands in
// the database: the stored key_hash equals sha256(secret) and the raw secret
// substring is absent from the row. Also covers actorOrNil's nil branch by
// issuing with no ActorID in context (created_by must be NULL).
func TestIntegrationApiKeyOnlyHashStored(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := testkit.CreateTenant(t, h).ID

	// Context carries a tenant but deliberately NO actor id -> actorOrNil returns nil.
	ctx := database.WithTenantID(context.Background(), tenant)
	var token string
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		token, id, e = s.Issue(ctx, db, "hash-only", []string{"a.b.read"}, nil)
		return e
	}); err != nil {
		t.Fatalf("issue: %v", err)
	}

	// token = wowapi_<prefix>_<secret>; extract the secret portion.
	parts := strings.SplitN(token, "_", 3)
	if len(parts) != 3 {
		t.Fatalf("malformed token %q", token)
	}
	secret := parts[2]
	wantHash := hex.EncodeToString(func() []byte { h := sha256.Sum256([]byte(secret)); return h[:] }())

	var gotHash string
	var createdBy *uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT key_hash, created_by FROM api_keys WHERE id = $1`, id).Scan(&gotHash, &createdBy); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if gotHash != wantHash {
		t.Fatalf("key_hash = %s, want sha256(secret) %s", gotHash, wantHash)
	}
	if gotHash == secret {
		t.Fatal("stored hash equals plaintext secret")
	}
	if strings.Contains(gotHash, secret) {
		t.Fatal("plaintext secret substring present in stored hash")
	}
	if createdBy != nil {
		t.Fatalf("created_by must be NULL when no actor in ctx, got %s", createdBy)
	}

	// Sanity: the secret still authenticates via the hash comparison.
	if _, err := s.Verify(context.Background(), h.PlatformTxM, token); err != nil {
		t.Fatalf("verify hash-stored key: %v", err)
	}
}

// TestIntegrationApiKeyList covers List: it returns non-secret KeyInfo rows for
// the tenant's keys ordered by creation, and never exposes a secret.
func TestIntegrationApiKeyList(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	exp := time.Now().Add(time.Hour)
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, _, e := s.Issue(ctx, db, "list-a", []string{"a.b.read"}, &exp); e != nil {
			return e
		}
		_, _, e := s.Issue(ctx, db, "list-b", []string{"c.d.read", "c.d.update"}, nil)
		return e
	}); err != nil {
		t.Fatalf("issue: %v", err)
	}

	var infos []apikey.KeyInfo
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		infos, e = s.List(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(infos) != 2 {
		t.Fatalf("List returned %d keys, want 2", len(infos))
	}
	byName := map[string]apikey.KeyInfo{}
	for _, ki := range infos {
		byName[ki.Name] = ki
		if ki.Prefix == "" {
			t.Fatalf("KeyInfo.Prefix empty for %s", ki.Name)
		}
	}
	a, ok := byName["list-a"]
	if !ok {
		t.Fatal("list-a missing from List")
	}
	if a.ExpiresAt == nil {
		t.Fatal("list-a must carry its expiry")
	}
	if len(a.Scopes) != 1 || a.Scopes[0] != "a.b.read" {
		t.Fatalf("list-a scopes = %v", a.Scopes)
	}
	if a.RevokedAt != nil {
		t.Fatal("list-a must not be revoked")
	}
	b := byName["list-b"]
	if len(b.Scopes) != 2 {
		t.Fatalf("list-b scopes = %v, want 2", b.Scopes)
	}
	if b.ExpiresAt != nil {
		t.Fatal("list-b has no expiry, want nil")
	}
}

// TestIntegrationApiKeyRotateNotFound covers Rotate's KindNotFound branch when
// the id is not an active key of this tenant.
func TestIntegrationApiKeyRotateNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Rotate(ctx, db, uuid.New()) // unknown id
		return e
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("rotate of unknown id must be KindNotFound, got %v", err)
	}
}

// TestIntegrationApiKeyRevokeNotFoundAndIdempotent covers Revoke's
// RowsAffected==0 branch: revoking an unknown id, and re-revoking an already
// revoked key, both return KindNotFound (the UPDATE matches no active row).
func TestIntegrationApiKeyRevokeNotFoundAndIdempotent(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	// Unknown id -> NotFound.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return s.Revoke(ctx, db, uuid.New())
	}); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("revoke unknown must be KindNotFound, got %v", err)
	}

	// Issue, revoke once (ok), revoke again -> NotFound (no active row left).
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		_, id, e = s.Issue(ctx, db, "rev", []string{"a.b.read"}, nil)
		return e
	}); err != nil {
		t.Fatalf("issue: %v", err)
	}
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return s.Revoke(ctx, db, id)
	}); err != nil {
		t.Fatalf("first revoke: %v", err)
	}
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return s.Revoke(ctx, db, id)
	}); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("second revoke must be KindNotFound, got %v", err)
	}
}

// TestIntegrationApiKeyIssueInsertError covers Issue's insert-error branch by
// forcing a primary-key collision with a fixed IDGen.
func TestIntegrationApiKeyIssueInsertError(t *testing.T) {
	h := testkit.NewDB(t)
	dup := fixedGen{id: uuid.New()}
	s := apikey.NewStore(dup)
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	// First issue succeeds and commits row id=dup.id.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Issue(ctx, db, "dup-1", []string{"a.b.read"}, nil)
		return e
	}); err != nil {
		t.Fatalf("first issue: %v", err)
	}
	// Second issue reuses the same id -> INSERT violates the primary key.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Issue(ctx, db, "dup-2", []string{"a.b.read"}, nil)
		return e
	})
	if err == nil {
		t.Fatal("second issue with duplicate id must fail on insert")
	}
}

// TestIntegrationApiKeyRotateInsertError covers Rotate's insert-error branch:
// the fixed IDGen makes the rotated row's id collide with the already-issued key.
func TestIntegrationApiKeyRotateInsertError(t *testing.T) {
	h := testkit.NewDB(t)
	dup := fixedGen{id: uuid.New()}
	s := apikey.NewStore(dup)
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	var oldID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		_, oldID, e = s.Issue(ctx, db, "rot-src", []string{"a.b.read"}, nil)
		return e
	}); err != nil {
		t.Fatalf("issue: %v", err)
	}
	// Rotate loads oldID (active) then inserts newID == dup.id == oldID -> collision.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Rotate(ctx, db, oldID)
		return e
	})
	if err == nil {
		t.Fatal("rotate with colliding new id must fail on insert")
	}
}

// TestIntegrationApiKeyAuditRecordError covers the recordAudit error branch
// (and the wrap in recordAudit) by giving the audit Writer a fixed IDGen so the
// second issuance's audit_logs insert collides on its primary key. The store's
// own IDGen stays unique so the api_keys insert succeeds and the failure is
// isolated to the audit write.
func TestIntegrationApiKeyAuditRecordError(t *testing.T) {
	h := testkit.NewDB(t)
	auditGen := fixedGen{id: uuid.New()}
	s := apikey.NewStore(model.UUIDv7(), apikey.WithAudit(kaudit.New(auditGen, nil)))
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	// First issuance writes audit row id=auditGen.id.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Issue(ctx, db, "aud-1", []string{"a.b.read"}, nil)
		return e
	}); err != nil {
		t.Fatalf("first issue: %v", err)
	}
	// Second issuance: api_keys insert ok (unique id), audit insert collides.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Issue(ctx, db, "aud-2", []string{"a.b.read"}, nil)
		return e
	})
	if err == nil {
		t.Fatal("second issue must surface the audit insert failure")
	}
}

// TestIntegrationApiKeyRotateAuditError covers Rotate's recordAudit error
// branch: the api_keys row for the rotated key inserts fine (unique store id),
// but the audit Writer's fixed IDGen collides on the audit_logs primary key.
func TestIntegrationApiKeyRotateAuditError(t *testing.T) {
	h := testkit.NewDB(t)
	auditGen := fixedGen{id: uuid.New()}
	s := apikey.NewStore(model.UUIDv7(), apikey.WithAudit(kaudit.New(auditGen, nil)))
	tenant := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())

	// First issuance writes audit row id=auditGen.id.
	var oldID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		_, oldID, e = s.Issue(ctx, db, "rot-aud", []string{"a.b.read"}, nil)
		return e
	}); err != nil {
		t.Fatalf("issue: %v", err)
	}
	// Rotate: new api_keys row ok, but the audit insert collides on its id.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, _, e := s.Rotate(ctx, db, oldID)
		return e
	})
	if err == nil {
		t.Fatal("rotate must surface the audit insert failure")
	}
}

// TestIntegrationApiKeyAuthenticatorNoBearer covers bearer()'s no-Authorization
// branch: a request without a Bearer token yields KindUnauthenticated.
func TestIntegrationApiKeyAuthenticatorNoBearer(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	auth := apikey.NewAuthenticator(s, h.PlatformTxM)

	req := httptest.NewRequest(http.MethodGet, "/", nil) // no Authorization header
	if _, err := auth.Authenticate(req); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("missing bearer must be unauthenticated, got %v", err)
	}
}

// TestIntegrationApiKeyAuthenticatorVerifyError covers Authenticate's error path
// when the token has the apikey scheme but fails verification.
func TestIntegrationApiKeyAuthenticatorVerifyError(t *testing.T) {
	h := testkit.NewDB(t)
	s := apikey.NewStore(model.UUIDv7())
	auth := apikey.NewAuthenticator(s, h.PlatformTxM)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wowapi_deadbeef_notarealsecret")
	actor, err := auth.Authenticate(req)
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("bad apikey must be unauthenticated, got %v", err)
	}
	if actor.Kind != "" || len(actor.Scopes) != 0 {
		t.Fatalf("failed auth must return zero actor, got %+v", actor)
	}
}
