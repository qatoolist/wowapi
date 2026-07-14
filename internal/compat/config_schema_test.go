package compat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckConfigSchemaCompatibility(t *testing.T) {
	fixtures := filepath.Join("testdata", "config-schema")
	baseline, err := os.ReadFile(filepath.Join(fixtures, "baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		fixture     string
		wantErr     bool
		wantMessage string
	}{
		{"identical", "baseline.json", false, ""},
		{"additive optional field", "additive-optional.json", false, ""},
		{"removed field", "breaking-removed.json", true, "database.host"},
		{"changed type", "breaking-type.json", true, "database.port"},
		{"new required field", "breaking-required.json", true, "database.sslmode"},
		{"narrowed enum", "breaking-enum.json", true, "environment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := os.ReadFile(filepath.Join(fixtures, tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			err = CheckConfigSchemaCompatibility(baseline, current)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error=%v, wantErr=%v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error %q missing path %q", err, tt.wantMessage)
			}
		})
	}
}

func TestCheckConfigSchemaCompatibilityRequiredDirection(t *testing.T) {
	baseline := []byte(`{
		"type":"object",
		"properties":{"database":{"type":"object","properties":{"host":{"type":"string"},"port":{"type":"integer"}},"required":["host"]}},
		"required":["database"]
	}`)
	requiredMadeOptional := []byte(`{
		"type":"object",
		"properties":{"database":{"type":"object","properties":{"host":{"type":"string"},"port":{"type":"integer"}}}},
		"required":["database"]
	}`)
	if err := CheckConfigSchemaCompatibility(baseline, requiredMadeOptional); err != nil {
		t.Fatalf("making a required property optional must remain compatible: %v", err)
	}

	optionalMadeRequired := []byte(`{
		"type":"object",
		"properties":{"database":{"type":"object","properties":{"host":{"type":"string"},"port":{"type":"integer"}},"required":["host","port"]}},
		"required":["database"]
	}`)
	err := CheckConfigSchemaCompatibility(baseline, optionalMadeRequired)
	if err == nil || !strings.Contains(err.Error(), "database.port became required") {
		t.Fatalf("making an optional property required must fail at its path, got %v", err)
	}
}

func TestCheckConfigSchemaCompatibilityRejectsInvalidSchemas(t *testing.T) {
	valid := []byte(`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object","properties":{}}`)
	for _, tc := range []struct {
		name     string
		baseline []byte
		current  []byte
		want     string
	}{
		{"invalid baseline JSON", []byte(`{"type":`), valid, "baseline"},
		{"invalid current JSON", valid, []byte(`null`), "current"},
		{"non-object properties", []byte(`{"type":"object","properties":[]}`), valid, "baseline.properties"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckConfigSchemaCompatibility(tc.baseline, tc.current)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want path %q", err, tc.want)
			}
		})
	}
}
