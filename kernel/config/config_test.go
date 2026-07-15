package config

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultsValidate(t *testing.T) {
	if err := Defaults().Validate(); err != nil {
		t.Fatalf("Defaults() must validate: %v", err)
	}
}

// TestPoolLifetimeKeysValidate pins the W01-E01-S001 contract for the
// MaxConnLifetime/MaxConnIdleTime pool keys: the defaults are pgx v5's own
// internal defaults (1h / 30m), explicit pgx-default values are accepted, and
// zero (an omitted key on a hand-built Pool literal) is accepted so that
// pre-existing configurations keep pre-story pool behavior (pgx defaults).
func TestPoolLifetimeKeysValidate(t *testing.T) {
	f := Defaults()
	if f.DB.MaxConnLifetime != time.Hour || f.DB.MaxConnIdleTime != 30*time.Minute {
		t.Fatalf("Defaults() pool lifetimes = (%v, %v), want pgx defaults (1h, 30m)",
			f.DB.MaxConnLifetime, f.DB.MaxConnIdleTime)
	}
	if err := f.Validate(); err != nil {
		t.Fatalf("pgx-default lifetime values must validate: %v", err)
	}

	// Omitted (zero) values = "use the pgx default" — must stay valid so
	// existing Pool literals that never set the keys keep working unchanged.
	f.DB.MaxConnLifetime = 0
	f.DB.MaxConnIdleTime = 0
	if err := f.Validate(); err != nil {
		t.Fatalf("zero lifetime values (pgx-default sentinel) must validate: %v", err)
	}
}

// TestValidateCollectsAllErrors proves fail-fast reports the complete list,
// not just the first problem (blueprint 12 §4).
func TestValidateCollectsAllErrors(t *testing.T) {
	f := Framework{} // everything invalid/zero
	err := f.Validate()
	if err == nil {
		t.Fatal("zero Framework validated")
	}
	msg := err.Error()
	for _, want := range []string{
		"environment:", "schema_version:", "http.addr:",
		"http.read_header_timeout:", "http.request_timeout:",
		"http.max_body_bytes:", "log.level:", "log.format:",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in joined error:\n%s", want, msg)
		}
	}
}

// TestTelemetryAndRateLimitValidation covers the CA-2 config keys: the trace
// sample ratio must be in [0,1], and an ENABLED rate limiter needs a positive
// rate and burst — while a DISABLED limiter skips those checks entirely.
func TestTelemetryAndRateLimitValidation(t *testing.T) {
	f := Defaults()
	f.Telemetry.TraceSampleRatio = 1.5
	f.HTTP.RateLimit.RequestsPerSecond = 0
	f.HTTP.RateLimit.Burst = 0
	err := f.Validate()
	if err == nil {
		t.Fatal("out-of-range ratio and zero rate/burst (enabled) must fail")
	}
	for _, want := range []string{
		"telemetry.trace_sample_ratio:",
		"http.rate_limit.requests_per_second:",
		"http.rate_limit.burst:",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}

	// Disabling the limiter waives its rate/burst checks (opt-out with no knobs).
	off := Defaults()
	off.HTTP.RateLimit = RateLimit{Disabled: true}
	if err := off.Validate(); err != nil {
		t.Fatalf("disabled rate limiter must validate without rate/burst: %v", err)
	}
}

// TestWebhookOutboundValidation covers the B2 config surface: the outbound
// SSRF allowlist. Disabled by default (SSRFAllowUnsafe=false); an invalid
// CIDR entry must fail validation, and a bad hostname entry (empty string)
// must fail too.
func TestWebhookOutboundValidation(t *testing.T) {
	f := Defaults()
	f.Webhook.Outbound.AllowedCIDRs = []string{"not-a-cidr"}
	err := f.Validate()
	if err == nil {
		t.Fatal("invalid outbound allowlist CIDR must fail validation")
	}
	if !strings.Contains(err.Error(), "webhook.outbound.allowed_cidrs:") {
		t.Errorf("missing %q in: %v", "webhook.outbound.allowed_cidrs:", err)
	}

	f2 := Defaults()
	f2.Webhook.Outbound.AllowedHosts = []string{""}
	err2 := f2.Validate()
	if err2 == nil {
		t.Fatal("empty outbound allowlist host must fail validation")
	}
	if !strings.Contains(err2.Error(), "webhook.outbound.allowed_hosts:") {
		t.Errorf("missing %q in: %v", "webhook.outbound.allowed_hosts:", err2)
	}
}

func TestWebhookOutboundDefaultsAreSafe(t *testing.T) {
	f := Defaults()
	if f.Webhook.Outbound.SSRFProtectionDisabled {
		t.Fatal("SSRF protection must be enabled by default")
	}
	if len(f.Webhook.Outbound.AllowedHosts) != 0 || len(f.Webhook.Outbound.AllowedCIDRs) != 0 {
		t.Fatal("the default allowlist must be empty")
	}
}

func TestProdSafetyFloor(t *testing.T) {
	f := Defaults()
	f.Environment = EnvProd
	f.Log.Format = "text"
	f.Log.Level = "debug"
	err := f.Validate()
	if err == nil {
		t.Fatal("prod with text/debug logging must fail validation")
	}
	for _, want := range []string{"prod requires json", "debug is not allowed in prod"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}
}

// TestProdSafetyFloorRejectsSSRFProtectionDisabled proves the B2 escape hatch
// cannot be used to disable the guard entirely in production — only the
// scoped host/CIDR allowlist is available there.
func TestProdSafetyFloorRejectsSSRFProtectionDisabled(t *testing.T) {
	f := Defaults()
	f.Environment = EnvProd
	f.Webhook.Outbound.SSRFProtectionDisabled = true
	err := f.Validate()
	if err == nil {
		t.Fatal("prod with SSRF protection disabled must fail validation")
	}
	if !strings.Contains(err.Error(), "webhook.outbound.ssrf_protection_disabled:") {
		t.Errorf("missing %q in: %v", "webhook.outbound.ssrf_protection_disabled:", err)
	}
}

func TestEnvValid(t *testing.T) {
	for _, tc := range []struct {
		env  Env
		want bool
	}{
		{EnvLocal, true},
		{EnvDev, true},
		{EnvStage, true},
		{EnvProd, true},
		{Env("production"), false},
		{Env(""), false},
	} {
		if got := tc.env.Valid(); got != tc.want {
			t.Errorf("Env(%q).Valid() = %v, want %v", tc.env, got, tc.want)
		}
	}
}

func TestModuleViewStrictDecode(t *testing.T) {
	type modCfg struct {
		SLAHours int  `json:"sla_hours"`
		Enabled  bool `json:"enabled"`
	}

	var cfg modCfg
	if err := (MapView{"sla_hours": 48, "enabled": true}).Decode(&cfg); err != nil {
		t.Fatalf("valid namespace rejected: %v", err)
	}
	if cfg.SLAHours != 48 || !cfg.Enabled {
		t.Errorf("decoded %+v", cfg)
	}

	if err := (MapView{"sla_hours": 48, "typo_key": 1}).Decode(&cfg); err == nil {
		t.Fatal("unknown key accepted — ModuleView must strict-decode")
	}
}
