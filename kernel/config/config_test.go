package config

import (
	"strings"
	"testing"
)

func TestDefaultsValidate(t *testing.T) {
	if err := Defaults().Validate(); err != nil {
		t.Fatalf("Defaults() must validate: %v", err)
	}
}

// TestValidateCollectsAllErrors proves fail-fast reports the complete list,
// not just the first problem (blueprint 12 §4).
func TestValidateCollectsAllErrors(t *testing.T) {
	f := Framework{} // everything invalid/zero
	err := f.Validate()
	if err == nil {
		t.Fatal("zero Framework validated")
	}
	msg := err.Error()
	for _, want := range []string{
		"environment:", "schema_version:", "http.addr:",
		"http.read_header_timeout:", "http.request_timeout:",
		"http.max_body_bytes:", "log.level:", "log.format:",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in joined error:\n%s", want, msg)
		}
	}
}

func TestProdSafetyFloor(t *testing.T) {
	f := Defaults()
	f.Environment = EnvProd
	f.Log.Format = "text"
	f.Log.Level = "debug"
	err := f.Validate()
	if err == nil {
		t.Fatal("prod with text/debug logging must fail validation")
	}
	for _, want := range []string{"prod requires json", "debug is not allowed in prod"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}
}

func TestEnvValid(t *testing.T) {
	for _, tc := range []struct {
		env  Env
		want bool
	}{
		{EnvLocal, true}, {EnvDev, true}, {EnvStage, true}, {EnvProd, true},
		{Env("production"), false}, {Env(""), false},
	} {
		if got := tc.env.Valid(); got != tc.want {
			t.Errorf("Env(%q).Valid() = %v, want %v", tc.env, got, tc.want)
		}
	}
}

func TestModuleViewStrictDecode(t *testing.T) {
	type modCfg struct {
		SLAHours int  `json:"sla_hours"`
		Enabled  bool `json:"enabled"`
	}

	var cfg modCfg
	if err := (MapView{"sla_hours": 48, "enabled": true}).Decode(&cfg); err != nil {
		t.Fatalf("valid namespace rejected: %v", err)
	}
	if cfg.SLAHours != 48 || !cfg.Enabled {
		t.Errorf("decoded %+v", cfg)
	}

	if err := (MapView{"sla_hours": 48, "typo_key": 1}).Decode(&cfg); err == nil {
		t.Fatal("unknown key accepted — ModuleView must strict-decode")
	}
}
