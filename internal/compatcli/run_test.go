package compatcli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConfig(t *testing.T) {
	fixtures := filepath.Join("..", "compat", "testdata", "config-schema")
	for _, tt := range []struct {
		name     string
		current  string
		wantCode int
		wantText string
	}{
		{"additive", "additive-optional.json", 0, "compatible"},
		{"breaking", "breaking-type.json", 1, "database.port"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{"config", "--baseline", filepath.Join(fixtures, "baseline.json"), "--current", filepath.Join(fixtures, tt.current)}, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("exit %d, want %d; stdout=%s stderr=%s", code, tt.wantCode, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String()+stderr.String(), tt.wantText) {
				t.Fatalf("output missing %q: stdout=%s stderr=%s", tt.wantText, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunConfigRequiresFiles(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"config"}, &stdout, &stderr); code != 2 {
		t.Fatalf("missing flags exit %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "--baseline") {
		t.Fatalf("missing actionable error: %s", stderr.String())
	}
}
