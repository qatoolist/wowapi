// Process config narrowing (blueprint 12 §7).
//
// One loaded config.Framework, three narrowed views: each process binary
// receives only the sections it actually needs. Unused sections are never
// wired, so a worker process has no HTTP section and a migrate process
// has no module namespaces. Each view carries a Fingerprint method so ops
// can detect per-section drift across processes on shared config sections.
package app

import "github.com/qatoolist/wowapi/kernel/config"

// APIConfig is the narrowed view handed to cmd/api: HTTP server settings,
// logging, and module namespaces. It deliberately omits migration DSN and
// provider credentials the API process does not use (blueprint 12 §7).
type APIConfig struct {
	Environment config.Env        `json:"environment"`
	HTTP        config.HTTP       `json:"http"`
	Log         config.Log        `json:"log"`
	Modules     config.Namespaces `json:"modules"`
}

// WorkerConfig is the narrowed view handed to cmd/worker: logging and module
// namespaces. The HTTP server section is deliberately absent — a worker does
// not bind a port (blueprint 12 §7).
type WorkerConfig struct {
	Environment config.Env        `json:"environment"`
	Log         config.Log        `json:"log"`
	Modules     config.Namespaces `json:"modules"`
}

// MigrateConfig is the narrowed view handed to cmd/migrate: logging only.
// Module namespaces and the HTTP section are deliberately absent; a migration
// DSN (app_migrate secret ref) arrives in Phase 2 and will widen this struct.
type MigrateConfig struct {
	Environment config.Env `json:"environment"`
	Log         config.Log `json:"log"`
}

// NewAPIConfig constructs the api process view from a loaded Framework and
// the full module namespace map.
func NewAPIConfig(f config.Framework, mods config.Namespaces) APIConfig {
	return APIConfig{
		Environment: f.Environment,
		HTTP:        f.HTTP,
		Log:         f.Log,
		Modules:     mods,
	}
}

// NewWorkerConfig constructs the worker process view from a loaded Framework
// and the full module namespace map.
func NewWorkerConfig(f config.Framework, mods config.Namespaces) WorkerConfig {
	return WorkerConfig{
		Environment: f.Environment,
		Log:         f.Log,
		Modules:     mods,
	}
}

// NewMigrateConfig constructs the migrate process view from a loaded Framework.
func NewMigrateConfig(f config.Framework) MigrateConfig {
	return MigrateConfig{
		Environment: f.Environment,
		Log:         f.Log,
	}
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
		"log":         c.Log,
		"modules":     c.Modules,
	})
}

// SectionFingerprints returns one fingerprint per top-level section so
// processes sharing a section (api/worker both carry log + modules) can be
// compared for config drift (blueprint 12 §7); comparing whole-view
// fingerprints across differently-shaped views is meaningless.
func (c MigrateConfig) SectionFingerprints() (map[string]config.Fingerprint, error) {
	return sectionFingerprints(map[string]any{
		"environment": c.Environment,
		"log":         c.Log,
	})
}
