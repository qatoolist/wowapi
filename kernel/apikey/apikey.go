// Package apikey provides machine authentication (roadmap S1): issuable, scoped,
// rotatable, revocable, expirable API keys / service principals so non-human
// callers authenticate without a user token. Only the sha256 of a key's secret
// is stored; the public prefix is the lookup handle. A verified key becomes an
// ActorSystem whose Scopes authorize it through the authz evaluator's machine
// fast-path (a scope acts like an RBAC grant, still subject to ABAC deny).
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
)

// scheme prefixes every token so an API key is distinguishable from a JWT.
const scheme = "wowapi"

// Store issues, verifies, and revokes API keys.
type Store struct {
	idgen model.IDGen
	now   func() time.Time
}

// NewStore builds a Store.
func NewStore(idgen model.IDGen) *Store {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Store{idgen: idgen, now: time.Now}
}

// Principal is a verified machine caller.
type Principal struct {
	KeyID    uuid.UUID
	TenantID uuid.UUID
	Name     string
	Scopes   []string
}

// Issue mints a key for the current tenant with the given scopes and optional
// expiry, records only the secret's hash, and returns the plaintext token ONCE —
// it cannot be recovered later. Runs in the caller's tenant transaction.
func (s *Store) Issue(ctx context.Context, db database.TenantDB, name string, scopes []string, expiresAt *time.Time) (token string, id uuid.UUID, err error) {
	if name == "" {
		return "", uuid.Nil, kerr.E(kerr.KindValidation, "invalid_api_key", "api key name is required")
	}
	prefix, secret, err := randParts()
	if err != nil {
		return "", uuid.Nil, kerr.Wrapf(err, "apikey.Issue", "generate key")
	}
	if scopes == nil {
		scopes = []string{}
	}
	id = s.idgen.New()
	if _, err := db.Exec(ctx,
		`INSERT INTO api_keys (id, tenant_id, name, key_prefix, key_hash, scopes, expires_at, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7)`,
		id, name, prefix, hashSecret(secret), scopes, expiresAt, actorOrNil(ctx)); err != nil {
		return "", uuid.Nil, kerr.Wrapf(err, "apikey.Issue", "insert key")
	}
	return scheme + "_" + prefix + "_" + secret, id, nil
}

// Verify authenticates a token cross-tenant (as app_platform, since the tenant is
// unknown pre-auth): it parses the prefix, loads the key, constant-time compares
// the secret hash, checks revocation/expiry, bumps last_used_at, and returns the
// Principal. Every failure is a KindUnauthenticated error with a non-specific
// message (no oracle for which check failed).
func (s *Store) Verify(ctx context.Context, plat database.TxManager, token string) (Principal, error) {
	prefix, secret, ok := parseToken(token)
	if !ok {
		return Principal{}, unauth()
	}
	var (
		p         Principal
		storedH   string
		scopes    []string
		expiresAt *time.Time
		revokedAt *time.Time
		found     bool
	)
	err := plat.Platform(ctx, func(ctx context.Context, db database.DB) error {
		row := db.QueryRow(ctx,
			`SELECT id, tenant_id, name, key_hash, scopes, expires_at, revoked_at
			   FROM api_keys WHERE key_prefix = $1`, prefix)
		if err := row.Scan(&p.KeyID, &p.TenantID, &p.Name, &storedH, &scopes, &expiresAt, &revokedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // found stays false
			}
			return err
		}
		found = true
		return nil
	})
	if err != nil {
		return Principal{}, kerr.Wrapf(err, "apikey.Verify", "load key")
	}
	// Constant-time hash compare even when the key is missing, to avoid timing
	// differences between "unknown prefix" and "wrong secret".
	ok = subtle.ConstantTimeCompare([]byte(hashSecret(secret)), []byte(storedH)) == 1
	if !found || !ok || revokedAt != nil || (expiresAt != nil && !expiresAt.After(s.now())) {
		return Principal{}, unauth()
	}
	p.Scopes = scopes
	// Best-effort last-used bump; a failure here must not fail authentication.
	_ = plat.Platform(ctx, func(ctx context.Context, db database.DB) error {
		_, _ = db.Exec(ctx, `UPDATE api_keys SET last_used_at = now() WHERE id = $1`, p.KeyID)
		return nil
	})
	return p, nil
}

// Revoke marks a key revoked (tenant-scoped). Idempotent; KindNotFound if the id
// is not an active key of this tenant.
func (s *Store) Revoke(ctx context.Context, db database.TenantDB, id uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`UPDATE api_keys SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return kerr.Wrapf(err, "apikey.Revoke", "revoke key")
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no active key with that id")
	}
	return nil
}

// KeyInfo is a non-secret view of a key for listing.
type KeyInfo struct {
	ID        uuid.UUID
	Name      string
	Prefix    string
	Scopes    []string
	ExpiresAt *time.Time
	RevokedAt *time.Time
	LastUsed  *time.Time
}

// List returns the current tenant's keys (never the secret).
func (s *Store) List(ctx context.Context, db database.TenantDB) ([]KeyInfo, error) {
	rows, err := db.Query(ctx,
		`SELECT id, name, key_prefix, scopes, expires_at, revoked_at, last_used_at
		   FROM api_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, kerr.Wrapf(err, "apikey.List", "query keys")
	}
	defer rows.Close()
	var out []KeyInfo
	for rows.Next() {
		var k KeyInfo
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Scopes, &k.ExpiresAt, &k.RevokedAt, &k.LastUsed); err != nil {
			return nil, kerr.Wrapf(err, "apikey.List", "scan key")
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// --- Authenticator: satisfies httpx.Authenticator structurally ---

// Authenticator verifies a Bearer API key and maps it to an ActorSystem. Wire it
// (or a composite that also tries OIDC) into httpx.SecureHandler.
type Authenticator struct {
	store *Store
	plat  database.TxManager
}

// NewAuthenticator builds the authenticator over the platform TxManager used for
// cross-tenant key verification.
func NewAuthenticator(store *Store, plat database.TxManager) *Authenticator {
	return &Authenticator{store: store, plat: plat}
}

// Authenticate reads the Bearer token; if it is an API key (wowapi_ prefix) it
// verifies it and returns the machine actor, else KindUnauthenticated so a
// composite authenticator can try another scheme.
func (a *Authenticator) Authenticate(r *http.Request) (authz.Actor, error) {
	tok := bearer(r)
	if !strings.HasPrefix(tok, scheme+"_") {
		return authz.Actor{}, unauth()
	}
	p, err := a.store.Verify(r.Context(), a.plat, tok)
	if err != nil {
		return authz.Actor{}, err
	}
	return authz.Actor{
		Kind:     authz.ActorSystem,
		System:   "apikey:" + p.Name,
		TenantID: p.TenantID,
		Scopes:   p.Scopes,
	}, nil
}

// --- helpers ---

func randParts() (prefix, secret string, err error) {
	pb := make([]byte, 8)
	sb := make([]byte, 24)
	if _, err = rand.Read(pb); err != nil {
		return "", "", err
	}
	if _, err = rand.Read(sb); err != nil {
		return "", "", err
	}
	return hex.EncodeToString(pb), base64.RawURLEncoding.EncodeToString(sb), nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// parseToken splits "wowapi_<prefix>_<secret>". SplitN on "_" with 3 parts so the
// base64url secret (which never contains "_") stays intact.
func parseToken(token string) (prefix, secret string, ok bool) {
	parts := strings.SplitN(token, "_", 3)
	if len(parts) != 3 || parts[0] != scheme || parts[1] == "" || parts[2] == "" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func bearer(r *http.Request) string {
	if after, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return after
	}
	return ""
}

func unauth() error {
	return kerr.E(kerr.KindUnauthenticated, "unauthenticated", "invalid or missing API key")
}

func actorOrNil(ctx context.Context) any {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return nil
}
