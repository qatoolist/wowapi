package app

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/config"
)

var fpZero config.Fingerprint

// TestMigrateConfig_FingerprintAndSections covers the migrate view's whole-view
// Fingerprint and its per-section fingerprints, which the api/worker cases in
// views_test.go exercise but the migrate case did not. The migrate DB slice is
// deliberately named "db_migrate" so ops never compare it against api/worker
// "db" (different shape, different DSN).
func TestMigrateConfig_FingerprintAndSections(t *testing.T) {
	c := mustMigrate(t, testFramework)

	fp, err := c.Fingerprint()
	if err != nil {
		t.Fatalf("MigrateConfig.Fingerprint() error: %v", err)
	}
	if fp == (fpZero) {
		t.Error("MigrateConfig.Fingerprint() returned the zero fingerprint")
	}

	secs, err := c.SectionFingerprints()
	if err != nil {
		t.Fatalf("MigrateConfig.SectionFingerprints() error: %v", err)
	}
	for _, want := range []string{"environment", "db_migrate", "log"} {
		if secs[want] == (fpZero) {
			t.Errorf("section %q missing from MigrateConfig.SectionFingerprints()", want)
		}
	}
	// The migrate view must NOT expose a plain "db" or "modules"/"http" section.
	for _, forbidden := range []string{"db", "modules", "http"} {
		if _, ok := secs[forbidden]; ok {
			t.Errorf("migrate view must not carry a %q section", forbidden)
		}
	}
}

// TestSectionFingerprints_MarshalError covers the error branch of the shared
// sectionFingerprints helper: a value that cannot be JSON-marshaled (a func)
// must surface FingerprintOf's error rather than being silently dropped.
func TestSectionFingerprints_MarshalError(t *testing.T) {
	_, err := sectionFingerprints(map[string]any{
		"bad": func() {}, // funcs are not JSON-marshalable
	})
	if err == nil {
		t.Fatal("sectionFingerprints must error on an unmarshalable section value")
	}
}
