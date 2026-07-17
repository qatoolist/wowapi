package logging_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/logging"
)

// TestJSONFormatProducesValidJSON verifies that Format:"json" emits valid JSON lines.
func TestJSONFormatProducesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("hello", "key", "val")
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, buf.String())
	}
	if m["msg"] != "hello" {
		t.Errorf("msg field = %v, want hello", m["msg"])
	}
}

// TestTextFormatProducesText verifies that Format:"text" emits key=value text, not JSON.
func TestTextFormatProducesText(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "text"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("hello")
	out := buf.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("text format output looks like JSON: %s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("output missing message: %s", out)
	}
}

// TestLevelFilteringDropsInfoAtWarn verifies that level=warn silently drops Info records.
func TestLevelFilteringDropsInfoAtWarn(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "warn", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("should-be-dropped")
	l.Warn("should-appear")
	out := buf.String()
	if strings.Contains(out, "should-be-dropped") {
		t.Errorf("info record leaked through warn-level filter: %s", out)
	}
	if !strings.Contains(out, "should-appear") {
		t.Errorf("warn record missing from output: %s", out)
	}
}

// TestUnknownLevelReturnsError verifies that an unrecognized level string errors.
func TestUnknownLevelReturnsError(t *testing.T) {
	_, err := logging.New(new(bytes.Buffer), config.Log{Level: "verbose", Format: "json"})
	if err == nil {
		t.Fatal("expected error for unknown level, got nil")
	}
}

// TestUnknownFormatReturnsError verifies that an unrecognized format string errors.
func TestUnknownFormatReturnsError(t *testing.T) {
	_, err := logging.New(new(bytes.Buffer), config.Log{Level: "info", Format: "logfmt"})
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}

// TestRedactionPasswordExact verifies exact key "password" is redacted.
func TestRedactionPasswordExact(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("x", "password", "hunter2")
	out := buf.String()
	if strings.Contains(out, "hunter2") {
		t.Errorf("password value leaked: %s", out)
	}
	if !strings.Contains(out, "[redacted]") {
		t.Errorf("redaction marker missing: %s", out)
	}
}

// TestRedactionTokenExact verifies exact key "token" is redacted.
func TestRedactionTokenExact(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("x", "token", "bearer-abc123")
	out := buf.String()
	if strings.Contains(out, "bearer-abc123") {
		t.Errorf("token value leaked: %s", out)
	}
	if !strings.Contains(out, "[redacted]") {
		t.Errorf("redaction marker missing: %s", out)
	}
}

// TestRedactionSuffixMatch verifies that a key with a sensitive suffix (db_password) is redacted.
func TestRedactionSuffixMatch(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("x", "db_password", "s3cr3t!")
	out := buf.String()
	if strings.Contains(out, "s3cr3t!") {
		t.Errorf("db_password value leaked: %s", out)
	}
	if !strings.Contains(out, "[redacted]") {
		t.Errorf("redaction marker missing: %s", out)
	}
}

// TestNonSensitiveKeyPassesThrough verifies that "username" is not redacted.
func TestNonSensitiveKeyPassesThrough(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("x", "username", "alice")
	out := buf.String()
	if !strings.Contains(out, "alice") {
		t.Errorf("non-sensitive value was redacted: %s", out)
	}
}

// TestRedactionIntegerValue verifies that a numeric value under a sensitive key
// is redacted regardless of kind (SEC-9: redaction must not be limited to strings).
func TestRedactionIntegerValue(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("x", "token", 123456)
	out := buf.String()
	if strings.Contains(out, "123456") {
		t.Errorf("integer token value leaked through redaction: %s", out)
	}
	if !strings.Contains(out, "[redacted]") {
		t.Errorf("redaction marker missing for integer token: %s", out)
	}
}

// TestRedactionDurationValue verifies that a time.Duration value under a
// sensitive key is redacted regardless of kind (SEC-9).
func TestRedactionDurationValue(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Info("x", "secret", 5*time.Second)
	out := buf.String()
	if strings.Contains(out, "5s") || strings.Contains(out, "5000000000") {
		t.Errorf("duration secret value leaked through redaction: %s", out)
	}
	if !strings.Contains(out, "[redacted]") {
		t.Errorf("redaction marker missing for duration secret: %s", out)
	}
}

// TestConfigSecretStructuralRedaction verifies that a config.Secret attr never
// exposes its raw value regardless of which key name is used.
func TestConfigSecretStructuralRedaction(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	s := config.NewSecret("secretref://env/X", "raw-value")
	// db_dsn exercises both the Secret.LogValue structural path and the heuristic DSN suffix.
	l.Info("boot", "db_dsn", s)
	out := buf.String()
	if strings.Contains(out, "raw-value") {
		t.Errorf("config.Secret leaked raw value: %s", out)
	}
	if !strings.Contains(out, "redacted") {
		t.Errorf("config.Secret redaction marker missing: %s", out)
	}
}

// TestLogStartupEmitsRequiredFields verifies the blueprint §7 startup record.
func TestLogStartupEmitsRequiredFields(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	fp, err := config.FingerprintOf(config.Defaults())
	if err != nil {
		t.Fatalf("FingerprintOf: %v", err)
	}
	logging.LogStartup(l, "api", config.EnvDev, fp)
	out := buf.String()

	checks := map[string]string{
		"process":                    "api",
		"environment":                "dev",
		"config_fingerprint (full)":  fp.String(),
		"config_fingerprint (short)": fp.Short(),
		"message":                    "starting",
	}
	for label, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("LogStartup output missing %s (%q): %s", label, want, out)
		}
	}
}
