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
	"net"
	"strings"
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
	// Environment carries NO default tag: it is fail-closed (D-0010/SEC-1) —
	// the loader errors when it is absent from every layer. The compiled
	// `local` value exists only through Defaults() for tests/local tooling.
	Environment   Env       `conf:"environment" json:"environment" doc:"deployment environment (local|dev|stage|prod); must be set explicitly in deployed processes"`
	SchemaVersion int       `conf:"schema_version" default:"1" json:"schema_version" doc:"config file format version"`
	HTTP          HTTP      `conf:"http" json:"http"`
	Log           Log       `conf:"log" json:"log"`
	DB            DB        `conf:"db" json:"db"`
	Telemetry     Telemetry `conf:"telemetry" json:"telemetry"`
	Webhook       Webhook   `conf:"webhook" json:"webhook"`
}

// Telemetry configures distributed tracing (roadmap O1). Tracing is OFF by
// default (zero-cost NoOp tracer) and becomes active only when the sample ratio
// is > 0 — the composition root then wires the OTel adapter with this ratio,
// exporting to the OTLP endpoint named by the standard OTEL_EXPORTER_OTLP_ENDPOINT
// environment variable (e.g. http://jaeger:4318 in the compose stack). This is
// the real config key that replaces the previously-documented-but-nonexistent
// cfg.TraceSampleRatio (roadmap CA-2/CA-7).
type Telemetry struct {
	TraceSampleRatio float64 `conf:"trace_sample_ratio" default:"0" json:"trace_sample_ratio" doc:"OTel trace head-sampling ratio 0.0..1.0 (0 disables tracing; export target is OTEL_EXPORTER_OTLP_ENDPOINT)"`
}

// DB configures the Postgres pools. DSNs are optional at load time and
// validated at process-view narrowing instead: api/worker require DSN,
// migrate requires MigrateDSN (D-0021) — the framework repo's config tooling
// and DB-less tests must stay loadable.
type DB struct {
	DSN         Secret `conf:"dsn" json:"dsn" doc:"runtime database DSN (app_rt role) as a secretref:// reference"`
	MigrateDSN  Secret `conf:"migrate_dsn" json:"migrate_dsn" doc:"migration DSN (app_migrate role) as a secretref:// reference; only the migrate process receives it"`
	PlatformDSN Secret `conf:"platform_dsn" json:"platform_dsn" doc:"cross-tenant platform DSN (a DEDICATED app_platform login) as a secretref:// reference; used by the api (API-key auth) and worker (relay/runner/scheduler). REQUIRED — the api/worker fail closed if it is unset. It must not reuse the runtime (app_rt) DSN: doing so would require app_rt to be a member of app_platform, a cluster-global grant that defeats runtime/platform privilege separation (CF-1)"`
	Pool               // embedded: pool knobs stay flat under db.* and flow to every process view wholesale
}

// Pool holds the connection-pool knobs shared by every process view. New
// pool fields belong HERE, never directly on DB: the app views embed Pool,
// so additions propagate to api/worker/migrate narrowing automatically
// instead of silently dropping out of a hand-copied field list (ARCH-17).
type Pool struct {
	MaxConns     int           `conf:"max_conns" default:"16" json:"max_conns" doc:"maximum pool connections"`
	QueryTimeout time.Duration `conf:"query_timeout" default:"5s" json:"query_timeout" doc:"per-query context deadline"`
}

// HTTP holds server guardrails. Zero values are replaced by Defaults.
type HTTP struct {
	Addr              string        `conf:"addr" default:":8080" json:"addr" doc:"HTTP listen address"`
	ReadHeaderTimeout time.Duration `conf:"read_header_timeout" default:"5s" json:"read_header_timeout" doc:"maximum time to read request headers"`
	RequestTimeout    time.Duration `conf:"request_timeout" default:"30s" json:"request_timeout" doc:"per-request handler timeout"`
	MaxBodyBytes      int64         `conf:"max_body_bytes" default:"1048576" json:"max_body_bytes" doc:"maximum request body size in bytes"`
	// CORSAllowedOrigins is the exact-match CORS allowlist (deny-by-default when
	// empty). Set per environment, e.g. modules-free base leaves it empty and the
	// prod overlay lists the product's web origins.
	CORSAllowedOrigins []string  `conf:"cors_allowed_origins" json:"cors_allowed_origins" doc:"exact-match CORS origin allowlist (empty = deny all cross-origin)"`
	RateLimit          RateLimit `conf:"rate_limit" json:"rate_limit"`
}

// RateLimit configures the in-process per-client rate limiter that the generated
// api installs in its default middleware chain (roadmap S2/CA-2). It is OPT-OUT:
// enabled unless Disabled is set, so a scaffolded product is protected against
// resource-exhaustion by default. Limits are guardrails, not billing.
type RateLimit struct {
	Disabled          bool    `conf:"disabled" json:"disabled" doc:"set true to remove the default per-client rate limiter from the chain"`
	RequestsPerSecond float64 `conf:"requests_per_second" default:"20" json:"requests_per_second" doc:"sustained requests/sec per client key (per replica)"`
	Burst             int     `conf:"burst" default:"40" json:"burst" doc:"burst capacity per client key"`
}

// Webhook configures the webhook framework (kernel/webhook).
type Webhook struct {
	Outbound WebhookOutbound `conf:"outbound" json:"outbound"`
}

// WebhookOutbound configures outbound webhook delivery's SSRF protection
// (backlog B2). Outbound delivery targets are USER-CONFIGURABLE URLs (tenants
// register their own webhook endpoints), so by default every dial is guarded
// by kernel/httpclient: loopback, link-local (incl. the 169.254.169.254 cloud
// metadata address), RFC1918/ULA private ranges, and unspecified addresses are
// all refused. AllowedHosts/AllowedCIDRs are the escape hatch for intentional
// internal targets (e.g. a tenant's own internal relay); SSRFProtectionDisabled
// exists only for local/dev convenience against a hand-rolled test receiver —
// Validate() below refuses it in prod.
type WebhookOutbound struct {
	SSRFProtectionDisabled bool     `conf:"ssrf_protection_disabled" json:"ssrf_protection_disabled" doc:"DANGEROUS: disables ALL dial-time SSRF protection for outbound webhook delivery. Refused in prod by Validate()"`
	AllowedHosts           []string `conf:"allowed_hosts" json:"allowed_hosts" doc:"exact-match hostname allowlist bypassing the blocked-address-class check for outbound webhook delivery"`
	AllowedCIDRs           []string `conf:"allowed_cidrs" json:"allowed_cidrs" doc:"CIDR allowlist (e.g. 10.20.0.0/16) for resolved outbound webhook delivery addresses"`
}

// Log configures structured logging.
type Log struct {
	Level  string `conf:"level" default:"info" json:"level" doc:"log level: debug|info|warn|error"`
	Format string `conf:"format" default:"json" json:"format" doc:"log output format: json|text (prod requires json)"`
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
			RateLimit:         RateLimit{Disabled: false, RequestsPerSecond: 20, Burst: 40},
		},
		Log:       Log{Level: "info", Format: "json"},
		DB:        DB{Pool: Pool{MaxConns: 16, QueryTimeout: 5 * time.Second}},
		Telemetry: Telemetry{TraceSampleRatio: 0},
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
	if f.DB.MaxConns < 2 || f.DB.MaxConns > 200 {
		add("db.max_conns: %d outside safe range 2..200", f.DB.MaxConns)
	}
	if f.DB.QueryTimeout < 100*time.Millisecond || f.DB.QueryTimeout > 60*time.Second {
		add("db.query_timeout: %v outside safe range 100ms..60s", f.DB.QueryTimeout)
	}
	if f.Telemetry.TraceSampleRatio < 0 || f.Telemetry.TraceSampleRatio > 1 {
		add("telemetry.trace_sample_ratio: %v outside range 0.0..1.0", f.Telemetry.TraceSampleRatio)
	}
	if !f.HTTP.RateLimit.Disabled {
		if f.HTTP.RateLimit.RequestsPerSecond <= 0 {
			add("http.rate_limit.requests_per_second: must be > 0 when the limiter is enabled")
		}
		if f.HTTP.RateLimit.Burst < 1 {
			add("http.rate_limit.burst: must be >= 1 when the limiter is enabled")
		}
	}
	for _, h := range f.Webhook.Outbound.AllowedHosts {
		if strings.TrimSpace(h) == "" {
			add("webhook.outbound.allowed_hosts: entries must not be blank")
			break
		}
	}
	for _, c := range f.Webhook.Outbound.AllowedCIDRs {
		if _, _, err := net.ParseCIDR(strings.TrimSpace(c)); err != nil {
			add("webhook.outbound.allowed_cidrs: %q is not a valid CIDR: %v", c, err)
		}
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
		if f.Webhook.Outbound.SSRFProtectionDisabled {
			add("webhook.outbound.ssrf_protection_disabled: must not be true in prod")
		}
	}

	return errors.Join(errs...)
}
