// Package config defines wowapi's typed configuration contracts: the
// framework-owned Framework struct, the Secret type with structural
// redaction, and the ModuleView through which modules receive their
// namespaced configuration.
//
// Ownership model, precedence, and loader behavior are specified in
// docs/blueprint/12-configuration-and-deployment.md. Phase 0 ships the core
// types and validation; the layered loader (files → overlay → env vars →
// secret resolution) lands in Phase 1.
//
// Product applications compose rather than fork: they define their own
// Config type embedding Framework in an internal/appcfg package (scaffolded
// by `wowapi init`).
package config

import (
	"errors"
	"fmt"
	"time"
)

// Env is the deployment environment. It gates dev-only behavior: anything
// marked unsafe refuses to run when the environment is Prod.
type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvStage Env = "stage"
	EnvProd  Env = "prod"
)

// Valid reports whether e is one of the known environments.
func (e Env) Valid() bool {
	switch e {
	case EnvLocal, EnvDev, EnvStage, EnvProd:
		return true
	}
	return false
}

// IsProd reports whether production safety rules apply.
func (e Env) IsProd() bool { return e == EnvProd }

// SchemaVersion is the current config file format version. Loaders reject
// files declaring a newer version (config written for a newer wowapi) and
// files older than the supported floor.
const SchemaVersion = 1

// Framework is the framework-owned configuration. It is loaded and validated
// once at boot and is immutable afterwards; hot paths read precomputed
// values, never stores. Fields grow phase by phase with the components that
// consume them (DB in Phase 2, Auth in Phase 4, …).
type Framework struct {
	Environment   Env  `json:"environment"`
	SchemaVersion int  `json:"schema_version"`
	HTTP          HTTP `json:"http"`
	Log           Log  `json:"log"`
}

// HTTP holds server guardrails. Zero values are replaced by Defaults.
type HTTP struct {
	Addr              string        `json:"addr"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MaxBodyBytes      int64         `json:"max_body_bytes"`
}

// Log configures structured logging.
type Log struct {
	Level  string `json:"level"`  // debug|info|warn|error
	Format string `json:"format"` // json|text
}

// Defaults returns the compiled framework defaults — the always-present,
// always-safe bottom layer of the precedence chain.
func Defaults() Framework {
	return Framework{
		Environment:   EnvLocal,
		SchemaVersion: SchemaVersion,
		HTTP: HTTP{
			Addr:              ":8080",
			ReadHeaderTimeout: 5 * time.Second,
			RequestTimeout:    30 * time.Second,
			MaxBodyBytes:      1 << 20, // 1 MiB
		},
		Log: Log{Level: "info", Format: "json"},
	}
}

// Validate checks the whole struct and returns ALL problems joined, not just
// the first — boot failures must list everything wrong at once.
func (f Framework) Validate() error {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if !f.Environment.Valid() {
		add("environment: %q is not one of local|dev|stage|prod", string(f.Environment))
	}
	if f.SchemaVersion < 1 || f.SchemaVersion > SchemaVersion {
		add("schema_version: %d unsupported (supported: 1..%d)", f.SchemaVersion, SchemaVersion)
	}
	if f.HTTP.Addr == "" {
		add("http.addr: required")
	}
	if f.HTTP.ReadHeaderTimeout <= 0 {
		add("http.read_header_timeout: must be > 0")
	}
	if f.HTTP.RequestTimeout <= 0 {
		add("http.request_timeout: must be > 0")
	}
	if f.HTTP.MaxBodyBytes <= 0 {
		add("http.max_body_bytes: must be > 0")
	}
	switch f.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		add("log.level: %q is not one of debug|info|warn|error", f.Log.Level)
	}
	switch f.Log.Format {
	case "json", "text":
	default:
		add("log.format: %q is not one of json|text", f.Log.Format)
	}

	// Production safety floor. Dev-only conveniences added in later phases
	// carry an `unsafe` marker and are rejected here when Environment is prod.
	if f.Environment.IsProd() {
		if f.Log.Format != "json" {
			add("log.format: prod requires json")
		}
		if f.Log.Level == "debug" {
			add("log.level: debug is not allowed in prod")
		}
	}

	return errors.Join(errs...)
}
