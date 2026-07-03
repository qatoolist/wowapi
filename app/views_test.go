package app

import (
	"reflect"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
)

var testFramework = config.Framework{
	Environment:   config.EnvDev,
	SchemaVersion: config.SchemaVersion,
	HTTP: config.HTTP{
		Addr:              ":9090",
		ReadHeaderTimeout: config.Defaults().HTTP.ReadHeaderTimeout,
		RequestTimeout:    config.Defaults().HTTP.RequestTimeout,
		MaxBodyBytes:      config.Defaults().HTTP.MaxBodyBytes,
	},
	Log: config.Log{Level: "debug", Format: "text"},
}

var testMods = config.Namespaces{
	"catalog": config.MapView{"price_ttl": "5m"},
	"billing": config.MapView{"currency": "USD"},
}

// TestWorkerConfig_NoHTTPField asserts via reflection that WorkerConfig has
// no field of type config.HTTP — the HTTP section is deliberately absent from
// the worker view (blueprint 12 §7).
func TestWorkerConfig_NoHTTPField(t *testing.T) {
	wType := reflect.TypeOf(WorkerConfig{})
	httpType := reflect.TypeOf(config.HTTP{})
	for i := 0; i < wType.NumField(); i++ {
		if wType.Field(i).Type == httpType {
			t.Fatalf("WorkerConfig must not have a field of type config.HTTP (found: %s)", wType.Field(i).Name)
		}
	}
}

// TestMigrateConfig_NoModulesField asserts via reflection that MigrateConfig
// has no field of type config.Namespaces — module namespaces are deliberately
// absent from the migrate view (blueprint 12 §7).
func TestMigrateConfig_NoModulesField(t *testing.T) {
	mType := reflect.TypeOf(MigrateConfig{})
	nsType := reflect.TypeOf(config.Namespaces{})
	for i := 0; i < mType.NumField(); i++ {
		if mType.Field(i).Type == nsType {
			t.Fatalf("MigrateConfig must not have a field of type config.Namespaces (found: %s)", mType.Field(i).Name)
		}
	}
}

// TestMigrateConfig_NoHTTPField asserts that MigrateConfig also has no HTTP
// server section.
func TestMigrateConfig_NoHTTPField(t *testing.T) {
	mType := reflect.TypeOf(MigrateConfig{})
	httpType := reflect.TypeOf(config.HTTP{})
	for i := 0; i < mType.NumField(); i++ {
		if mType.Field(i).Type == httpType {
			t.Fatalf("MigrateConfig must not have a field of type config.HTTP (found: %s)", mType.Field(i).Name)
		}
	}
}

// TestNewAPIConfig_Fields verifies the constructor copies the expected sections.
func TestNewAPIConfig_Fields(t *testing.T) {
	c := NewAPIConfig(testFramework, testMods)
	if c.Environment != testFramework.Environment {
		t.Errorf("Environment = %v, want %v", c.Environment, testFramework.Environment)
	}
	if c.HTTP != testFramework.HTTP {
		t.Errorf("HTTP = %v, want %v", c.HTTP, testFramework.HTTP)
	}
	if c.Log != testFramework.Log {
		t.Errorf("Log = %v, want %v", c.Log, testFramework.Log)
	}
	if len(c.Modules) != len(testMods) {
		t.Errorf("Modules len = %d, want %d", len(c.Modules), len(testMods))
	}
}

// TestNewWorkerConfig_Fields verifies the constructor omits HTTP.
func TestNewWorkerConfig_Fields(t *testing.T) {
	c := NewWorkerConfig(testFramework, testMods)
	if c.Environment != testFramework.Environment {
		t.Errorf("Environment = %v, want %v", c.Environment, testFramework.Environment)
	}
	if c.Log != testFramework.Log {
		t.Errorf("Log = %v, want %v", c.Log, testFramework.Log)
	}
	if len(c.Modules) != len(testMods) {
		t.Errorf("Modules len = %d, want %d", len(c.Modules), len(testMods))
	}
}

// TestNewMigrateConfig_Fields verifies the constructor includes only env+log.
func TestNewMigrateConfig_Fields(t *testing.T) {
	c := NewMigrateConfig(testFramework)
	if c.Environment != testFramework.Environment {
		t.Errorf("Environment = %v, want %v", c.Environment, testFramework.Environment)
	}
	if c.Log != testFramework.Log {
		t.Errorf("Log = %v, want %v", c.Log, testFramework.Log)
	}
}

// TestSectionFingerprints_SharedSectionsAgree verifies that for the same
// Framework+Namespaces, APIConfig and WorkerConfig SectionFingerprints agree
// on the sections they share ("environment", "log", "modules") — the
// per-section approach is what makes cross-process drift detection possible
// (blueprint 12 §7).
func TestSectionFingerprints_SharedSectionsAgree(t *testing.T) {
	api := NewAPIConfig(testFramework, testMods)
	worker := NewWorkerConfig(testFramework, testMods)

	apiFPs, err := api.SectionFingerprints()
	if err != nil {
		t.Fatalf("APIConfig.SectionFingerprints() error: %v", err)
	}
	workerFPs, err := worker.SectionFingerprints()
	if err != nil {
		t.Fatalf("WorkerConfig.SectionFingerprints() error: %v", err)
	}

	for _, section := range []string{"environment", "log", "modules"} {
		if apiFPs[section] != workerFPs[section] {
			t.Errorf("section %q fingerprints differ between API and Worker for identical input: api=%s worker=%s",
				section, apiFPs[section].Short(), workerFPs[section].Short())
		}
	}
}

// TestSectionFingerprints_ChangedSectionDiffers verifies that after mutating
// Log.Level, the "log" section fingerprints differ while "environment" still
// agrees — confirming independent per-section hashing (blueprint 12 §7).
func TestSectionFingerprints_ChangedSectionDiffers(t *testing.T) {
	f1 := testFramework
	f2 := testFramework
	f2.Log.Level = "warn" // mutate only the log section

	fps1, err := NewAPIConfig(f1, testMods).SectionFingerprints()
	if err != nil {
		t.Fatalf("SectionFingerprints() error: %v", err)
	}
	fps2, err := NewAPIConfig(f2, testMods).SectionFingerprints()
	if err != nil {
		t.Fatalf("SectionFingerprints() error: %v", err)
	}

	if fps1["log"] == fps2["log"] {
		t.Error(`"log" section fingerprints must differ after changing Log.Level`)
	}
	if fps1["environment"] != fps2["environment"] {
		t.Error(`"environment" section fingerprints must agree when only Log changed`)
	}
}

// TestAPIAndWorkerFingerprints_Differ verifies that for the same Framework
// the API and Worker fingerprints differ, because the API view includes HTTP.
func TestAPIAndWorkerFingerprints_Differ(t *testing.T) {
	api := NewAPIConfig(testFramework, testMods)
	worker := NewWorkerConfig(testFramework, testMods)

	apiFP, err := api.Fingerprint()
	if err != nil {
		t.Fatalf("APIConfig.Fingerprint() error: %v", err)
	}
	workerFP, err := worker.Fingerprint()
	if err != nil {
		t.Fatalf("WorkerConfig.Fingerprint() error: %v", err)
	}
	if apiFP == workerFP {
		t.Error("APIConfig and WorkerConfig fingerprints must differ (different sections)")
	}
}

// TestFingerprint_Deterministic verifies same input always produces the same
// fingerprint (fingerprinting is a pure function of the view's content).
func TestFingerprint_Deterministic(t *testing.T) {
	api1 := NewAPIConfig(testFramework, testMods)
	api2 := NewAPIConfig(testFramework, testMods)

	fp1, err := api1.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := api2.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if fp1 != fp2 {
		t.Errorf("same input produced different fingerprints: %s vs %s", fp1.Short(), fp2.Short())
	}
}

// TestFingerprint_ChangesWithContent verifies that mutating a field changes
// the fingerprint.
func TestFingerprint_ChangesWithContent(t *testing.T) {
	f1 := testFramework
	f2 := testFramework
	f2.Log.Level = "warn"

	fp1, err := NewAPIConfig(f1, testMods).Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := NewAPIConfig(f2, testMods).Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if fp1 == fp2 {
		t.Error("fingerprints must differ when config content changes")
	}
}
