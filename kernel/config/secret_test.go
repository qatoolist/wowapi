package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

const rawValue = "super-sensitive-value"

func newTestSecret() Secret { return NewSecret("secretref://env/DB_DSN", rawValue) }

// TestSecretNeverLeaks drives every standard rendering path and asserts the
// raw value is absent — this is acceptance criterion 12 §11.5 in miniature.
func TestSecretNeverLeaks(t *testing.T) {
	s := newTestSecret()

	renders := map[string]string{
		"%v":       fmt.Sprintf("%v", s),
		"%+v":      fmt.Sprintf("%+v", s),
		"%#v":      fmt.Sprintf("%#v", s),
		"%s":       fmt.Sprintf("%s", s),
		"%q":       fmt.Sprintf("%q", s),
		"%x":       fmt.Sprintf("%x", s),
		"String()": s.String(),
	}
	if b, err := json.Marshal(s); err != nil {
		t.Fatalf("json.Marshal: %v", err)
	} else {
		renders["json"] = string(b)
	}
	if b, err := s.MarshalText(); err != nil {
		t.Fatalf("MarshalText: %v", err)
	} else {
		renders["text"] = string(b)
	}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("boot", "dsn", s)
	renders["slog"] = buf.String()

	// Struct-embedded rendering must be safe too.
	type wrapper struct{ DSN Secret }
	renders["struct %+v"] = fmt.Sprintf("%+v", wrapper{DSN: s})

	for name, got := range renders {
		if strings.Contains(got, rawValue) {
			t.Errorf("%s leaked the secret value: %s", name, got)
		}
		if !strings.Contains(got, "redacted") {
			t.Errorf("%s missing redaction marker: %s", name, got)
		}
	}

	if s.Reveal() != rawValue {
		t.Errorf("Reveal() = %q, want raw value", s.Reveal())
	}
}

func TestSecretUnmarshalTextAcceptsOnlyRefs(t *testing.T) {
	var s Secret
	if err := s.UnmarshalText([]byte("secretref://env/API_KEY")); err != nil {
		t.Fatalf("valid ref rejected: %v", err)
	}
	if got := s.Ref(); got != "secretref://env/API_KEY" {
		t.Errorf("Ref() = %q", got)
	}

	err := s.UnmarshalText([]byte("hunter2-raw-password"))
	if err == nil {
		t.Fatal("raw value accepted as secret config — must be rejected")
	}
	if strings.Contains(err.Error(), "hunter2-raw-password") {
		t.Errorf("error message echoed the raw candidate value: %v", err)
	}
}

func TestSecretJSONArrayAndMapContexts(t *testing.T) {
	s := newTestSecret()
	b, err := json.Marshal(map[string]any{"nested": []any{s}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), rawValue) {
		t.Errorf("nested JSON leaked secret: %s", b)
	}
}
