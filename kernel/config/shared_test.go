package config_test

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/config"
)

func TestSharedFingerprintIgnoresProcessSpecificSections(t *testing.T) {
	base := config.Defaults()
	api := base
	worker := base
	// The api serves HTTP on an address; the worker need not — a process-specific
	// difference that must NOT count as drift.
	api.HTTP.Addr = ":8080"
	worker.HTTP.Addr = ":9090"

	af, err := api.SharedFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	wf, err := worker.SharedFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if af != wf {
		t.Fatalf("HTTP-only difference must not change the shared fingerprint: %s vs %s", af.Short(), wf.Short())
	}
}

func TestSharedFingerprintChangesWithSharedSection(t *testing.T) {
	a := config.Defaults()
	b := config.Defaults()
	b.SchemaVersion = a.SchemaVersion + 1 // a shared section differs

	af, _ := a.SharedFingerprint()
	bf, _ := b.SharedFingerprint()
	if af == bf {
		t.Fatal("a shared-section change must change the shared fingerprint")
	}
}

func TestSharedFingerprintChangesWithOutboundAllowlist(t *testing.T) {
	base := config.Defaults()
	mutated := base
	mutated.Webhook.Outbound.AllowedHosts = []string{"relay.internal.example"}
	mutated.Webhook.Outbound.AllowedCIDRs = []string{"10.0.0.0/8"}

	bf, _ := base.SharedFingerprint()
	mf, _ := mutated.SharedFingerprint()
	if bf == mf {
		t.Fatal("SharedFingerprint must change when the outbound allowlist changes")
	}
}

func TestSharedFingerprintChangesWithTrustedIssuers(t *testing.T) {
	base := config.Defaults()
	mutated := base
	mutated.Security.TrustedIssuers = []string{"https://idp.example.com"}

	bf, _ := base.SharedFingerprint()
	mf, _ := mutated.SharedFingerprint()
	if bf == mf {
		t.Fatal("SharedFingerprint must change when security.trusted_issuers changes")
	}
}

func TestCheckSharedDrift(t *testing.T) {
	cfg := config.Defaults()
	fp, err := cfg.SharedFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	// Matching expected → no drift.
	if err := cfg.CheckSharedDrift(fp.String()); err != nil {
		t.Fatalf("matching fingerprint should not report drift: %v", err)
	}
	// Empty expected → check disabled.
	if err := cfg.CheckSharedDrift(""); err != nil {
		t.Fatalf("empty expected should disable the check: %v", err)
	}
	// A different expected → drift detected.
	if err := cfg.CheckSharedDrift("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"); err == nil {
		t.Fatal("a divergent expected fingerprint must report drift")
	}
}
