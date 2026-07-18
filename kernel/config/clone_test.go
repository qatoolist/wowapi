package config

import "testing"

// Fifth closure-audit gate finding (2026-07-17): the reflective deep copy
// must PRESERVE unexported-field value types — an earlier iteration created
// fresh zero structs and skipped unsettable fields, silently zeroing every
// config.Secret (DB.DSN, MigrateDSN, PlatformDSN) in the cloned config.
func TestClonePreservesSecretsAndIsolatesMutables(t *testing.T) {
	f := Defaults()
	f.DB.DSN = NewSecret("secretref://env/DB_DSN", "postgres://real-dsn")
	f.HTTP.CORSAllowedOrigins = []string{"https://trusted.example"}
	f.Webhook.Outbound.AllowedHosts = []string{"hooks.example"}

	c := f.Clone()

	if c.DB.DSN.Reveal() != "postgres://real-dsn" || c.DB.DSN.Ref() != "secretref://env/DB_DSN" {
		t.Fatalf("Clone zeroed the Secret: ref=%q revealed=%q", c.DB.DSN.Ref(), c.DB.DSN.Reveal())
	}

	// Mutating the original's nested slices must not reach the clone.
	f.HTTP.CORSAllowedOrigins[0] = "https://evil.example"
	f.Webhook.Outbound.AllowedHosts[0] = "evil.example"
	if c.HTTP.CORSAllowedOrigins[0] != "https://trusted.example" {
		t.Fatalf("clone shares CORS slice storage: %v", c.HTTP.CORSAllowedOrigins)
	}
	if c.Webhook.Outbound.AllowedHosts[0] != "hooks.example" {
		t.Fatalf("clone shares webhook allowlist storage: %v", c.Webhook.Outbound.AllowedHosts)
	}
}
