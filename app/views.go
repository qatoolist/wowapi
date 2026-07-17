// Process config narrowing (blueprint 12 §7).
//
// One loaded config.Framework, three narrowed views: each process binary
// receives only the sections it actually needs. Unused sections are never
// wired, so a worker process has no HTTP section and a migrate process
// has no module namespaces. Each view carries a Fingerprint method so ops
// can detect per-section drift across processes on shared config sections.
package app

import (
	"errors"

	"github.com/qatoolist/wowapi/kernel/config"
)

// RuntimeDB is the runtime slice of config.DB handed to api/worker: the
// app_rt DSN plus the embedded pool knobs. The migration DSN is deliberately
// absent — runtime processes never hold app_migrate credentials (12 §7).
// Embedding config.Pool (not re-listing its fields) means new pool knobs
// reach every view without touching the narrowing code (ARCH-17).
type RuntimeDB struct {
	DSN         config.Secret `json:"dsn"`
	config.Pool               // flattens: max_conns, query_timeout, …
}

// MigrateDB is the migrate slice of config.DB: the app_migrate DSN plus the
// pool knobs. The runtime DSN is deliberately absent (blueprint 12 §7).
type MigrateDB struct {
	DSN config.Secret `json:"dsn"`
	config.Pool
}

// APIConfig is the narrowed view handed to cmd/api: HTTP server settings,
// runtime database, logging, and module namespaces. It deliberately omits
// the migration DSN and provider credentials the API process does not use
// (blueprint 12 §7).
type APIConfig struct {
	Environment config.Env        `json:"environment"`
	HTTP        config.HTTP       `json:"http"`
	DB          RuntimeDB         `json:"db"`
	Log         config.Log        `json:"log"`
	Modules     config.Namespaces `json:"modules"`
}

// WorkerConfig is the narrowed view handed to cmd/worker: runtime database,
// logging, and module namespaces. The HTTP server section is deliberately
// absent — a worker does not bind a port (blueprint 12 §7).
type WorkerConfig struct {
	Environment config.Env        `json:"environment"`
	DB          RuntimeDB         `json:"db"`
	Log         config.Log        `json:"log"`
	Modules     config.Namespaces `json:"modules"`
}

// MigrateConfig is the narrowed view handed to cmd/migrate: the migration
// DSN and logging only. Module namespaces, the HTTP section, and the runtime
// DSN are deliberately absent (blueprint 12 §7).
type MigrateConfig struct {
	Environment config.Env `json:"environment"`
	DB          MigrateDB  `json:"db_migrate"`
	Log         config.Log `json:"log"`
}

// NewAPIConfig constructs the api process view from a loaded Framework and
// the full module namespace map. The runtime DSN is required here — DSNs are
// validated at narrowing, not by config tags, so DB-less tooling loads stay
// possible (D-0021).
func NewAPIConfig(f config.Framework, mods config.Namespaces) (APIConfig, error) {
	if f.DB.DSN.IsZero() {
		return APIConfig{}, errors.New("app: db.dsn is required for the api process")
	}
	return APIConfig{
		Environment: f.Environment,
		HTTP:        f.HTTP,
		DB:          RuntimeDB{DSN: f.DB.DSN, Pool: f.DB.Pool},
		Log:         f.Log,
		Modules:     mods,
	}, nil
}

// NewWorkerConfig constructs the worker process view from a loaded Framework
// and the full module namespace map. Requires the runtime DSN (D-0021).
func NewWorkerConfig(f config.Framework, mods config.Namespaces) (WorkerConfig, error) {
	if f.DB.DSN.IsZero() {
		return WorkerConfig{}, errors.New("app: db.dsn is required for the worker process")
	}
	return WorkerConfig{
		Environment: f.Environment,
		DB:          RuntimeDB{DSN: f.DB.DSN, Pool: f.DB.Pool},
		Log:         f.Log,
		Modules:     mods,
	}, nil
}

// NewMigrateConfig constructs the migrate process view from a loaded
// Framework. Requires the migration DSN (D-0021).
func NewMigrateConfig(f config.Framework) (MigrateConfig, error) {
	if f.DB.MigrateDSN.IsZero() {
		return MigrateConfig{}, errors.New("app: db.migrate_dsn is required for the migrate process")
	}
	return MigrateConfig{
		Environment: f.Environment,
		DB:          MigrateDB{DSN: f.DB.MigrateDSN, Pool: f.DB.Pool},
		Log:         f.Log,
	}, nil
}

// Fingerprint returns the SHA-256 of this view's canonical redacted JSON
// rendering. Shared sections fingerprinted separately let ops detect
// api-vs-worker drift on a half-rolled deploy (blueprint 12 §7).
func (c APIConfig) Fingerprint() (config.Fingerprint, error) {
	return config.FingerprintOf(c)
}

// Fingerprint returns the SHA-256 of this view's canonical redacted JSON
// rendering.
func (c WorkerConfig) Fingerprint() (config.Fingerprint, error) {
	return config.FingerprintOf(c)
}

// Fingerprint returns the SHA-256 of this view's canonical redacted JSON
// rendering.
func (c MigrateConfig) Fingerprint() (config.Fingerprint, error) {
	return config.FingerprintOf(c)
}

// sectionFingerprints is a shared unexported helper that fingerprints each
// value in sections independently so processes sharing a subset of sections
// can compare drift on those sections alone (blueprint 12 §7).
func sectionFingerprints(sections map[string]any) (map[string]config.Fingerprint, error) {
	out := make(map[string]config.Fingerprint, len(sections))
	for k, v := range sections {
		fp, err := config.FingerprintOf(v)
		if err != nil {
			return nil, err
		}
		out[k] = fp
	}
	return out, nil
}

// SectionFingerprints returns one fingerprint per top-level section so
// processes sharing a section (api/worker both carry log + modules) can be
// compared for config drift (blueprint 12 §7); comparing whole-view
// fingerprints across differently-shaped views is meaningless.
func (c APIConfig) SectionFingerprints() (map[string]config.Fingerprint, error) {
	return sectionFingerprints(map[string]any{
		"environment": c.Environment,
		"http":        c.HTTP,
		"db":          c.DB,
		"log":         c.Log,
		"modules":     c.Modules,
	})
}

// SectionFingerprints returns one fingerprint per top-level section so
// processes sharing a section (api/worker both carry log + modules) can be
// compared for config drift (blueprint 12 §7); comparing whole-view
// fingerprints across differently-shaped views is meaningless.
func (c WorkerConfig) SectionFingerprints() (map[string]config.Fingerprint, error) {
	return sectionFingerprints(map[string]any{
		"environment": c.Environment,
		"db":          c.DB,
		"log":         c.Log,
		"modules":     c.Modules,
	})
}

// SectionFingerprints returns one fingerprint per top-level section so
// processes sharing a section (api/worker both carry log + modules) can be
// compared for config drift (blueprint 12 §7); comparing whole-view
// fingerprints across differently-shaped views is meaningless.
func (c MigrateConfig) SectionFingerprints() (map[string]config.Fingerprint, error) {
	// The migrate DB slice is named db_migrate so ops never compare it
	// against the api/worker "db" section — different shapes, different DSNs.
	return sectionFingerprints(map[string]any{
		"environment": c.Environment,
		"db_migrate":  c.DB,
		"log":         c.Log,
	})
}
