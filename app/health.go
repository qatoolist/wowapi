package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/seeds"
)

// Readiness assembles the /readyz aggregator from the booted app: every module's
// registered readiness check (ctx.Health) plus the framework checks the caller
// supplies as `extra` (typically a DB ping and a "migrations current" probe,
// which the composition root wires because it owns the pool). The redacted config
// fingerprint is reported in the response for drift correlation. Mount
// h.Liveness() at /healthz and h.Readiness() at /readyz in the product's api and
// worker mains (blueprint 07 §9).
func Readiness(b *Booted, fingerprint config.Fingerprint, extra map[string]httpx.HealthCheck) *httpx.Health {
	h := httpx.NewHealth(fingerprint.String())
	// Read the boot-validated view, not the reassignable Health field (second
	// closure audit 2026-07-17, F-10).
	for name, chk := range b.runtimeHealth() {
		h.Register("module."+name, chk)
	}
	for name, chk := range extra {
		h.Register(name, chk)
	}
	return h
}

// ReadinessWithCatalogs is the framework-assembled readiness aggregator used by
// the generated api/worker mains. It is identical to Readiness but additionally
// registers the `seed_catalogs` readiness check when the booted product declares
// seeds, reports the latest recorded seed/catalog hash, checks migration
// currency, and reports migration/rule/model hashes in the readiness payload
// details. This closes the CS-21 fail-first gap: a prod-profile boot against an
// empty catalog database or stale-migrated database fails readiness with an
// actionable message until seed-sync / migrate has run.
func ReadinessWithCatalogs(b *Booted, fingerprint config.Fingerprint, db database.DBTX, src fs.FS, source string, extra map[string]httpx.HealthCheck) *httpx.Health {
	h := Readiness(b, fingerprint, extra)

	// Read the boot-validated seed catalog, never the reassignable Seeds field
	// (third closure audit 2026-07-17, F-10): the readiness hash and the
	// seeded-catalog check must reflect what boot validated.
	validatedSeeds := b.runtimeSeeds()
	if bundleHasSeeds(validatedSeeds) {
		h.Register("seed_catalogs", func(ctx context.Context) error {
			return CatalogsSeeded(ctx, db, validatedSeeds)
		})
	}

	if src != nil && source != "" {
		h.Register("migration_currency", MigrationCurrencyCheck(db, src, source))
		h.Detail(MigrationVersionDetail(db, source))
	}

	h.Detail(func(ctx context.Context) (string, any) {
		hash, err := latestSeedHash(ctx, db)
		if err != nil || hash == "" {
			return "", nil
		}
		return "seed_catalog_hash", hash
	})
	h.Detail(func(ctx context.Context) (string, any) {
		if b.runtimeKernel().Rules == nil {
			return "", nil
		}
		hash := RuleHash(b.runtimeKernel().Rules)
		if hash == "" {
			return "", nil
		}
		return "rule_hash", hash
	})
	h.Detail(func(ctx context.Context) (string, any) {
		if b.runtimeKernel().ModelHash == "" {
			return "", nil
		}
		return "model_hash", b.runtimeKernel().ModelHash
	})

	return h
}

// MigrationCurrencyCheck returns a readiness check that fails when the migration
// source's expected version (highest numbered migration in src) is ahead of the
// version recorded in the database's goose_version_<source> table.
func MigrationCurrencyCheck(db database.DBTX, src fs.FS, source string) httpx.HealthCheck {
	expected, err := MaxMigrationVersion(src)
	if err != nil {
		return func(ctx context.Context) error {
			return fmt.Errorf("migration currency: cannot read expected version: %w", err)
		}
	}
	return func(ctx context.Context) error {
		var applied int64
		err := db.QueryRow(ctx,
			"SELECT version_id FROM goose_version_"+source+" WHERE is_applied ORDER BY version_id DESC LIMIT 1").Scan(&applied)
		if err != nil {
			return fmt.Errorf("migration currency: cannot read applied version: %w", err)
		}
		if applied < expected {
			return fmt.Errorf("migration currency: applied version %d lags expected %d; run migrations", applied, expected)
		}
		return nil
	}
}

// MaxMigrationVersion returns the highest numbered migration version found in src.
func MaxMigrationVersion(src fs.FS) (int64, error) {
	var max int64
	err := fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || len(path) < 4 || path[len(path)-4:] != ".sql" {
			return nil
		}
		v, err := parseMigrationVersion(path)
		if err != nil {
			return nil // non-goose files are ignored
		}
		if int64(v) > max {
			max = int64(v)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return max, nil
}

// MigrationVersionDetail returns a detail provider that reports the applied
// migration version from the database's goose_version_<source> table.
func MigrationVersionDetail(db database.DBTX, source string) httpx.DetailProvider {
	return func(ctx context.Context) (string, any) {
		var version int64
		err := db.QueryRow(ctx,
			"SELECT version_id FROM goose_version_"+source+" WHERE is_applied ORDER BY version_id DESC LIMIT 1").Scan(&version)
		if err != nil {
			return "", nil
		}
		return "migration_version", version
	}
}

// RuleHash returns a deterministic hash of the registered rule points. It is
// stable across map iteration order and JSON formatting differences.
func RuleHash(r *rules.Registry) string {
	if r == nil {
		return ""
	}
	keys := r.Keys()
	if len(keys) == 0 {
		return ""
	}
	pts := r.Points()
	h := sha256.New()
	for _, k := range keys {
		p := pts[k]
		b, err := json.Marshal(struct {
			Key              string          `json:"key"`
			Module           string          `json:"module"`
			ValueSchema      json.RawMessage `json:"value_schema"`
			Default          json.RawMessage `json:"default"`
			AllowedScopes    []string        `json:"allowed_scopes"`
			RequiresApproval bool            `json:"requires_approval"`
			Description      string          `json:"description"`
		}{
			Key:              p.Key,
			Module:           p.Module,
			ValueSchema:      p.ValueSchema,
			Default:          p.Default,
			AllowedScopes:    scopeStrings(p.AllowedScopes),
			RequiresApproval: p.RequiresApproval,
			Description:      p.Description,
		})
		if err != nil {
			continue // should never happen
		}
		_, _ = h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// parseMigrationVersion extracts the leading NNNNN number from a goose-style
// migration filename. It is duplicated here to avoid an app -> migration test
// import cycle (migration tests import testkit, which imports app).
func parseMigrationVersion(filename string) (int, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("migration filename %q is not NNNNN_name.sql", filename)
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("migration filename %q: %w", filename, err)
	}
	return v, nil
}

func scopeStrings(scopes []rules.ScopeKind) []string {
	out := make([]string, len(scopes))
	for i, s := range scopes {
		out[i] = string(s)
	}
	sort.Strings(out)
	return out
}

func bundleHasSeeds(b seeds.Bundle) bool {
	return len(b.Permissions) > 0 || len(b.ResourceTypes) > 0 ||
		len(b.RelationshipTypes) > 0 || len(b.Roles) > 0
}
