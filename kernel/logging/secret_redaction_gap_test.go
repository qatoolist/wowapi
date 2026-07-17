package logging_test

// Gap-fill for secret redaction coverage (criterion #26).
//
// Existing tests in logging_test.go verify structural redaction using the key
// "db_dsn", which ALSO matches the heuristic DSN-suffix rule. These tests
// isolate the two layers:
//
//  1. Structural: config.Secret implements slog.LogValuer; the value is
//     redacted regardless of the attribute key name.
//  2. Heuristic (defense-in-depth): redactAttr catches sensitive key suffixes
//     for non-Secret types (raw strings, ints, etc.).
//
// If the structural layer were broken (e.g., LogValuer removed from Secret),
// a Secret logged under a non-sensitive key like "app_endpoint" would leak.
// The tests below detect that regression.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/logging"
)

// TestSecretStructuralRedactionNonSensitiveKey proves that config.Secret is
// ALWAYS redacted by its own LogValuer, regardless of the attribute key.
// "app_config_val" does not match any heuristic suffix, so only structural
// redaction protects it. This is the primary security mechanism.
func TestSecretStructuralRedactionNonSensitiveKey(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	const rawVal = "raw-secret-value-12345"
	s := config.NewSecret("secretref://env/API_KEY", rawVal)

	// "app_config_val" has no sensitive suffix; only structural redaction applies.
	l.Info("boot", "app_config_val", s)
	out := buf.String()

	if strings.Contains(out, rawVal) {
		t.Errorf("FINDING: Secret leaked through non-sensitive key 'app_config_val' — LogValuer structural path is broken:\n%s", out)
	}
	if !strings.Contains(out, "redacted") {
		t.Errorf("redaction marker missing for non-sensitive key:\n%s", out)
	}
}

// TestSecretStructuralRedactionMultipleKeys proves redaction holds across
// several attribute positions and key styles in the same log record.
func TestSecretStructuralRedactionMultipleKeys(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	const rawVal = "multi-key-secret-9999"
	s := config.NewSecret("secretref://env/MULTI", rawVal)

	l.Info("boot",
		"a", s, // first attr, non-sensitive key
		"password", s, // second attr, sensitive key (double-redaction, harmless)
		"z_val", s, // last attr, non-sensitive key
	)
	out := buf.String()

	if strings.Contains(out, rawVal) {
		t.Errorf("Secret raw value leaked in multi-key log record:\n%s", out)
	}
}

// TestHeuristicRedactsCatchesRawStringUnderSensitiveKey proves the heuristic
// (second line of defense) catches a raw string accidentally logged under a
// sensitive key when the value is NOT a config.Secret.
func TestHeuristicRedactsCatchesRawStringUnderSensitiveKey(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	const rawStr = "hunter2-raw-string"

	// Log a plain string (not a config.Secret) under a sensitive key.
	// Structural redaction does not apply here; heuristic must catch it.
	l.Info("boot", "db_password", rawStr)
	out := buf.String()

	if strings.Contains(out, rawStr) {
		t.Errorf("heuristic redaction missed plain string under 'db_password':\n%s", out)
	}
	if !strings.Contains(out, "[redacted]") {
		t.Errorf("redaction marker missing for heuristic path:\n%s", out)
	}
}
