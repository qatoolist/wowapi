package kernel

import (
	"io"
	"log/slog"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
)

// Sixth review regression (2026-07-17, C-05): a missing or malformed DSR
// artifact key must FAIL BOOT in production, never silently fall back to the
// deterministic shared test key. Non-production keeps the test-key
// convenience.
func TestArtifactWriterFailsClosedInProd(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	t.Setenv("WOWAPI_DSR_ARTIFACT_KEY", "") // missing

	if _, err := newArtifactWriter(log, nil, config.EnvProd); err == nil {
		t.Fatal("production boot accepted a missing artifact key (would encrypt with the public test key)")
	}
	t.Setenv("WOWAPI_DSR_ARTIFACT_KEY", "not-hex-and-too-short")
	if _, err := newArtifactWriter(log, nil, config.EnvProd); err == nil {
		t.Fatal("production boot accepted a malformed artifact key")
	}
	// Non-production: the test-key convenience is allowed (warned).
	w, err := newArtifactWriter(log, nil, config.EnvLocal)
	if err != nil || w == nil {
		t.Fatalf("local boot rejected the test-key convenience: writer=%v err=%v", w, err)
	}
}
