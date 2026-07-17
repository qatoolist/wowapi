package app

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/config"
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
	DB: config.DB{
		DSN:        config.NewSecret("secretref://env/APP_DSN", "postgres://rt:pw@h/db"),
		MigrateDSN: config.NewSecret("secretref://env/MIGRATE_DSN", "postgres://mig:pw@h/db"),
		Pool:       config.Defaults().DB.Pool,
	},
}

func mustAPI(t *testing.T, f config.Framework, mods config.Namespaces) APIConfig {
	t.Helper()
	c, err := NewAPIConfig(f, mods)
	if err != nil {
		t.Fatalf("NewAPIConfig: %v", err)
	}
	return c
}

func mustWorker(t *testing.T, f config.Framework, mods config.Namespaces) WorkerConfig {
	t.Helper()
	c, err := NewWorkerConfig(f, mods)
	if err != nil {
		t.Fatalf("NewWorkerConfig: %v", err)
	}
	return c
}

func mustMigrate(t *testing.T, f config.Framework) MigrateConfig {
	t.Helper()
	c, err := NewMigrateConfig(f)
	if err != nil {
		t.Fatalf("NewMigrateConfig: %v", err)
	}
	return c
}

// D-0021: DSNs are validated at narrowing, not by config tags.
func TestViewsRequireTheirDSN(t *testing.T) {
	noRT := testFramework
	noRT.DB.DSN = config.Secret{}
	if _, err := NewAPIConfig(noRT, nil); err == nil {
		t.Error("api view must require db.dsn")
	}
	if _, err := NewWorkerConfig(noRT, nil); err == nil {
		t.Error("worker view must require db.dsn")
	}
	noMig := testFramework
	noMig.DB.MigrateDSN = config.Secret{}
	if _, err := NewMigrateConfig(noMig); err == nil {
		t.Error("migrate view must require db.migrate_dsn")
	}
}

// 12 §7: runtime processes never hold app_migrate credentials, and the
// migrate process never holds the runtime DSN. The redaction markers carry
// the refs, so the rendered JSON proves which secrets a view can even name.
func TestViewsCarryOnlyTheirDSN(t *testing.T) {
	api, _ := json.Marshal(mustAPI(t, testFramework, testMods))
	worker, _ := json.Marshal(mustWorker(t, testFramework, testMods))
	migrate, _ := json.Marshal(mustMigrate(t, testFramework))

	for name, js := range map[string][]byte{"api": api, "worker": worker} {
		if strings.Contains(string(js), "MIGRATE_DSN") {
			t.Errorf("%s view must not carry the migrate DSN: %s", name, js)
		}
		if !strings.Contains(string(js), "APP_DSN") {
			t.Errorf("%s view should carry the runtime DSN ref: %s", name, js)
		}
	}
	if strings.Contains(string(migrate), "APP_DSN") {
		t.Errorf("migrate view must not carry the runtime DSN: %s", migrate)
	}
	if !strings.Contains(string(migrate), "MIGRATE_DSN") {
		t.Errorf("migrate view should carry the migrate DSN ref: %s", migrate)
	}
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
	c := mustAPI(t, testFramework, testMods)
	if c.Environment != testFramework.Environment {
		t.Errorf("Environment = %v, want %v", c.Environment, testFramework.Environment)
	}
	if !reflect.DeepEqual(c.HTTP, testFramework.HTTP) {
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
	c := mustWorker(t, testFramework, testMods)
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
	c := mustMigrate(t, testFramework)
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
	api := mustAPI(t, testFramework, testMods)
	worker := mustWorker(t, testFramework, testMods)

	apiFPs, err := api.SectionFingerprints()
	if err != nil {
		t.Fatalf("APIConfig.SectionFingerprints() error: %v", err)
	}
	workerFPs, err := worker.SectionFingerprints()
	if err != nil {
		t.Fatalf("WorkerConfig.SectionFingerprints() error: %v", err)
	}

	for _, section := range []string{"environment", "db", "log", "modules"} {
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

	fps1, err := mustAPI(t, f1, testMods).SectionFingerprints()
	if err != nil {
		t.Fatalf("SectionFingerprints() error: %v", err)
	}
	fps2, err := mustAPI(t, f2, testMods).SectionFingerprints()
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
	api := mustAPI(t, testFramework, testMods)
	worker := mustWorker(t, testFramework, testMods)

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
	api1 := mustAPI(t, testFramework, testMods)
	api2 := mustAPI(t, testFramework, testMods)

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

	fp1, err := mustAPI(t, f1, testMods).Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := mustAPI(t, f2, testMods).Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if fp1 == fp2 {
		t.Error("fingerprints must differ when config content changes")
	}
}
