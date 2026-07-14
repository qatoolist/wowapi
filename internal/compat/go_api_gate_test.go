package compat

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoAPIDiffGateFixtures(t *testing.T) {
	root := filepath.Join("testdata", "go-api")
	script := filepath.Join("..", "..", "scripts", "check_go_api_compat.sh")
	for _, tt := range []struct {
		name        string
		current     string
		wantSuccess bool
		wantText    string
	}{
		{"identical", "baseline", true, "compatible"},
		{"additive", "additive", true, "compatible"},
		{"removed exported method", "breaking-removed", false, "Get"},
		{"changed exported type", "breaking-changed", false, "Timeout"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("sh", script, filepath.Join(root, "baseline"), filepath.Join(root, tt.current))
			cmd.Env = append(os.Environ(), "GO_API_COMPAT_ALLOWLIST=/dev/null")
			output, err := cmd.CombinedOutput()
			if (err == nil) != tt.wantSuccess {
				t.Fatalf("success=%v want=%v; output=%s", err == nil, tt.wantSuccess, output)
			}
			if !strings.Contains(string(output), tt.wantText) {
				t.Fatalf("output %q missing %q", output, tt.wantText)
			}
		})
	}
}
