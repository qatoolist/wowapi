package integration

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/secrets"
)

// Store reads/writes integration_providers rows and resolves a provider's
// config + credential. Config reads run on the caller's TenantDB (RLS: a tenant
// sees its own rows + platform defaults); writes run with platform privilege
// (provider config is behavior-changing, kept off the module role — SEC-13).
type Store struct {
	reg     *Registry
	secrets secrets.Provider
	idgen   model.IDGen
}

// NewStore wires the store. secrets may be nil (credential resolution then errors
// if a row carries a credential_ref).
func NewStore(reg *Registry, sec secrets.Provider, idgen model.IDGen) *Store {
	return &Store{reg: reg, secrets: sec, idgen: idgen}
}

// UpsertIn describes a provider config row. TenantID zero → a platform-wide row
// (tenant_id NULL); non-zero → a tenant override (the DBTX must be bound to that
// tenant so RLS admits the write).
type UpsertIn struct {
	TenantID      uuid.UUID
	Key           string
	Kind          string
	Settings      map[string]any
	CredentialRef string // a secretref://... string, or "" for none
}

// Upsert inserts or updates a provider config row (platform privilege). Returns
// the row id.
func (s *Store) Upsert(ctx context.Context, db database.DBTX, in UpsertIn) (uuid.UUID, error) {
	if !keyRE.MatchString(in.Key) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "provider_invalid", "provider key must be module.name")
	}
	if !validKinds[in.Kind] {
		return uuid.Nil, kerr.E(kerr.KindValidation, "provider_invalid", "invalid provider kind: "+in.Kind)
	}
	if in.CredentialRef != "" && !secrets.IsRef(in.CredentialRef) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "provider_invalid", "credential_ref must be a secretref:// reference, never plaintext")
	}
	settings, err := json.Marshal(orEmpty(in.Settings))
	if err != nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "provider_invalid", "settings not JSON-encodable")
	}
	var tenantArg any
	if in.TenantID != uuid.Nil {
		tenantArg = in.TenantID
	}
	id := s.idgen.New()
	actor := actorFrom(ctx)
	// RETURNING id so the conflict (DO UPDATE) path returns the EXISTING row's id,
	// not the freshly-generated one that was discarded by the conflict (ARCH-71).
	var rowID uuid.UUID
	err = db.QueryRow(ctx,
		`INSERT INTO integration_providers (id, tenant_id, key, kind, config, credential_ref, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key)
		 DO UPDATE SET kind = EXCLUDED.kind, config = EXCLUDED.config,
		               credential_ref = EXCLUDED.credential_ref, status = 'active',
		               version = integration_providers.version + 1,
		               updated_at = now(), updated_by = EXCLUDED.created_by
		 RETURNING id`,
		id, tenantArg, in.Key, in.Kind, settings, nullStr(in.CredentialRef), actor).Scan(&rowID)
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "integration.Upsert", "upsert provider %s", in.Key)
	}
	return rowID, nil
}

// Resolve loads the active config for key (a tenant override wins over the
// platform default) and resolves its credential. Runs on the caller's TenantDB.
func (s *Store) Resolve(ctx context.Context, db database.TenantDB, key string) (Config, error) {
	var (
		kind      string
		configRaw []byte
		credRef   *string
		tenantID  *uuid.UUID
	)
	err := db.QueryRow(ctx,
		`SELECT kind, config, credential_ref, tenant_id
		   FROM integration_providers
		  WHERE key = $1 AND status = 'active'
		  ORDER BY tenant_id NULLS LAST
		  LIMIT 1`, key).Scan(&kind, &configRaw, &credRef, &tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Config{}, kerr.E(kerr.KindNotFound, "provider_not_configured", "no active provider configured for "+key)
	}
	if err != nil {
		return Config{}, kerr.Wrapf(err, "integration.Resolve", "load provider %s", key)
	}
	cfg := Config{Key: key, Kind: kind, IsPlatform: tenantID == nil}
	if len(configRaw) > 0 {
		if err := json.Unmarshal(configRaw, &cfg.Settings); err != nil {
			return Config{}, kerr.Wrapf(err, "integration.Resolve", "decode config for %s", key)
		}
	}
	if credRef != nil && *credRef != "" {
		if s.secrets == nil {
			return Config{}, kerr.E(kerr.KindInternal, "no_secrets_provider", "provider "+key+" has a credential_ref but no secrets provider is wired")
		}
		ref, err := secrets.ParseRef(*credRef)
		if err != nil {
			return Config{}, kerr.E(kerr.KindInternal, "invalid_credential_ref", "provider "+key+" has a malformed credential_ref")
		}
		val, err := s.secrets.Resolve(ctx, ref)
		if err != nil {
			return Config{}, kerr.Wrapf(err, "integration.Resolve", "resolve credential for %s", key)
		}
		cfg.Credential = config.NewSecret(*credRef, val)
	}
	return cfg, nil
}

// HealthChecks probes every registered provider that has an active config row
// visible to the caller's tenant. Returns key → error (nil = healthy). A provider
// with no config row is skipped (not enabled for this tenant). Non-fatal:
// intended for readiness detail.
func (s *Store) HealthChecks(ctx context.Context, db database.TenantDB) map[string]error {
	out := map[string]error{}
	for _, key := range s.reg.Keys() {
		p, _ := s.reg.Get(key)
		cfg, err := s.Resolve(ctx, db, key)
		if kerr.KindOf(err) == kerr.KindNotFound {
			continue // not configured for this tenant
		}
		if err != nil {
			out[key] = err
			continue
		}
		out[key] = p.HealthCheck(ctx, cfg)
	}
	return out
}

func actorFrom(ctx context.Context) uuid.UUID {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return uuid.Nil
}

func orEmpty(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
