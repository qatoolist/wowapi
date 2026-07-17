package app

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
)

// TestNewModuleContext_Logger verifies that Logger() returns a logger
// pre-tagged with "module=<name>" so all lines from Module.Register carry
// the module identity without the module having to add it manually.
func TestNewModuleContext_Logger(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(h)

	ctx := newModuleContext("mymod", logger, nil, moduleDeps{})
	ctx.Logger().Info("test message")

	if !strings.Contains(buf.String(), "module=mymod") {
		t.Errorf("logged output must contain module=mymod, got: %s", buf.String())
	}
}

// TestNewModuleContext_NilViewReturnsEmptyMapView verifies that a nil view
// yields an empty (but non-nil) MapView so modules with no namespace can
// still call Decode without a guard.
func TestNewModuleContext_NilViewReturnsEmptyMapView(t *testing.T) {
	ctx := newModuleContext("a", slog.Default(), nil, moduleDeps{})
	mv := ctx.Config()
	if mv == nil {
		t.Fatal("Config() must not return nil for a nil view")
	}
	var out map[string]any
	if err := mv.Decode(&out); err != nil {
		t.Fatalf("empty MapView must decode cleanly into map[string]any: %v", err)
	}
}

// TestNewModuleContext_ConfigIsolation verifies that a module context built
// from module "a"'s namespace can decode its own keys and — critically —
// cannot see keys from module "b".
func TestNewModuleContext_ConfigIsolation(t *testing.T) {
	ns := config.Namespaces{
		"a": config.MapView{"key1": "val1", "key2": "val2"},
		"b": config.MapView{"key3": "val3"},
	}

	ctx := newModuleContext("a", slog.Default(), ns["a"], moduleDeps{})

	var out map[string]any
	if err := ctx.Config().Decode(&out); err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if _, ok := out["key1"]; !ok {
		t.Error("module a must see key1")
	}
	if _, ok := out["key2"]; !ok {
		t.Error("module a must see key2")
	}
	if _, ok := out["key3"]; ok {
		t.Error("module a must NOT see key3 (belongs to module b)")
	}
}

// TestNewModuleContext_DecodeTypedStruct verifies that Config().Decode works
// with a typed struct, matching how modules actually use it in Register.
func TestNewModuleContext_DecodeTypedStruct(t *testing.T) {
	type modCfg struct {
		PriceTTL string `json:"price_ttl"`
	}
	ns := config.Namespaces{
		"catalog": config.MapView{"price_ttl": "5m"},
	}
	ctx := newModuleContext("catalog", slog.Default(), ns["catalog"], moduleDeps{})

	var cfg modCfg
	if err := ctx.Config().Decode(&cfg); err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if cfg.PriceTTL != "5m" {
		t.Errorf("PriceTTL = %q, want %q", cfg.PriceTTL, "5m")
	}
}
