package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIMergePreservesEveryOpenAPI31Field(t *testing.T) {
	dir := t.TempDir()
	fragments := map[string]string{
		"openapi.json":              `{"openapi":"3.1.1"}`,
		"info.json":                 `{"info":{"title":"T","version":"1.2.3"}}`,
		"json-schema-dialect.json":  `{"jsonSchemaDialect":"https://json-schema.org/draft/2020-12/schema"}`,
		"servers.json":              `{"servers":[{"url":"https://api.example.test"}]}`,
		"paths.json":                `{"paths":{"/pets":{"get":{"responses":{"200":{"description":"ok"}}}}}}`,
		"webhooks.json":             `{"webhooks":{"petChanged":{"$ref":"#/components/pathItems/PetChanged"}}}`,
		"security.json":             `{"security":[{"Bearer":[]}]}`,
		"tags.json":                 `{"tags":[{"name":"pets","description":"Pet operations"}]}`,
		"external-docs.json":        `{"externalDocs":{"url":"https://docs.example.test"}}`,
		"extension.json":            `{"x-owner":"platform"}`,
		"schemas.json":              `{"components":{"schemas":{"Pet":{"type":"object","properties":{"id":{"type":"string"}}}}}}`,
		"responses.json":            `{"components":{"responses":{"NotFound":{"description":"missing"}}}}`,
		"parameters.json":           `{"components":{"parameters":{"PetID":{"name":"id","in":"path","required":true,"schema":{"type":"string"}}}}}`,
		"examples.json":             `{"components":{"examples":{"Pet":{"value":{"id":"p1"}}}}}`,
		"request-bodies.json":       `{"components":{"requestBodies":{"Pet":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Pet"}}}}}}}`,
		"headers.json":              `{"components":{"headers":{"RequestID":{"schema":{"type":"string"}}}}}`,
		"security-schemes.json":     `{"components":{"securitySchemes":{"Bearer":{"type":"http","scheme":"bearer"}}}}`,
		"links.json":                `{"components":{"links":{"PetByID":{"operationId":"getPet"}}}}`,
		"callbacks.json":            `{"components":{"callbacks":{"OnPet":{"{$request.body#/callback}":{"post":{"responses":{"200":{"description":"ok"}}}}}}}}`,
		"path-items.json":           `{"components":{"pathItems":{"PetChanged":{"post":{"responses":{"200":{"description":"ok"}}}}}}}`,
		"components-extension.json": `{"components":{"x-owner":{"team":"platform"}}}`,
	}
	for name, body := range fragments {
		writeFile(t, dir, name, body)
	}

	var stdout, stderr bytes.Buffer
	code := runOpenAPI([]string{"merge", "--dir", dir, "--title", "T", "--version", "1.2.3"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("merge exit %d: %s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode merged document: %v", err)
	}
	for _, field := range []string{"openapi", "info", "jsonSchemaDialect", "servers", "paths", "webhooks", "security", "tags", "externalDocs", "x-owner"} {
		if _, ok := got[field]; !ok {
			t.Errorf("top-level field %q was silently dropped", field)
		}
	}
	components, ok := got["components"].(map[string]any)
	if !ok {
		t.Fatalf("components missing or wrong type: %#v", got["components"])
	}
	for _, field := range []string{"schemas", "responses", "parameters", "examples", "requestBodies", "headers", "securitySchemes", "links", "callbacks", "pathItems", "x-owner"} {
		if _, ok := components[field]; !ok {
			t.Errorf("components.%s was silently dropped", field)
		}
	}
}

func TestOpenAPIDocumentCapacityRejectsIntegerOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if _, err := openAPIDocumentCapacity(maxInt - 2); err == nil {
		t.Fatal("capacity calculation must reject a top-level field count that would overflow")
	}
	if got, err := openAPIDocumentCapacity(7); err != nil || got != 10 {
		t.Fatalf("safe capacity = %d, %v; want 10, nil", got, err)
	}
}

func TestOpenAPIMergePoliciesAreCompleteOrLoud(t *testing.T) {
	tests := []struct {
		name      string
		first     string
		second    string
		wantField string
	}{
		{"openapi", `{"openapi":"3.1.1"}`, `{"openapi":"3.0.3"}`, "openapi"},
		{"info", `{"info":{"title":"T","version":"1.2.3"}}`, `{"info":{"title":"other","version":"1.2.3"}}`, "info"},
		{"jsonSchemaDialect", `{"jsonSchemaDialect":"https://json-schema.org/draft/2020-12/schema"}`, `{"jsonSchemaDialect":"https://example.test/dialect"}`, "jsonSchemaDialect"},
		{"paths", `{"paths":{"/pets":{}}}`, `{"paths":{"/pets":{}}}`, "paths./pets"},
		{"webhooks", `{"webhooks":{"changed":{}}}`, `{"webhooks":{"changed":{}}}`, "webhooks.changed"},
		{"security", `{"security":[{"Bearer":[]}]}`, `{"security":[{"OAuth":["read"]}]}`, "security"},
		{"tags", `{"tags":[{"name":"pets","description":"one"}]}`, `{"tags":[{"name":"pets","description":"two"}]}`, "tags.pets"},
		{"externalDocs", `{"externalDocs":{"url":"https://one.test"}}`, `{"externalDocs":{"url":"https://two.test"}}`, "externalDocs"},
		{"extension", `{"x-owner":"one"}`, `{"x-owner":"two"}`, "x-owner"},
		{"unknown top-level", `{"paths":{}}`, `{"futureField":{}}`, "futureField"},
		{"unknown components", `{"paths":{}}`, `{"components":{"futureField":{}}}`, "components.futureField"},
	}
	componentFields := []string{"schemas", "responses", "parameters", "examples", "requestBodies", "headers", "securitySchemes", "links", "callbacks", "pathItems"}
	for _, field := range componentFields {
		tests = append(tests, struct {
			name      string
			first     string
			second    string
			wantField string
		}{"components." + field, `{"components":{"` + field + `":{"Duplicate":{}}}}`, `{"components":{"` + field + `":{"Duplicate":{}}}}`, "components." + field + ".Duplicate"})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "a.json", tt.first)
			writeFile(t, dir, "b.json", tt.second)
			var stdout, stderr bytes.Buffer
			if code := runOpenAPI([]string{"merge", "--dir", dir, "--title", "T", "--version", "1.2.3"}, &stdout, &stderr); code != 1 {
				t.Fatalf("conflict must fail, got exit %d and output %s", code, stdout.String())
			}
			if !strings.Contains(stderr.String(), tt.wantField) {
				t.Fatalf("error %q does not identify field %q", stderr.String(), tt.wantField)
			}
		})
	}
}

func TestOpenAPIMergeUnionsServersAndRejectsMalformedOutput(t *testing.T) {
	t.Run("servers union and deduplicate", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.json", `{"servers":[{"url":"https://one.test"},{"url":"https://shared.test"}]}`)
		writeFile(t, dir, "b.json", `{"servers":[{"url":"https://shared.test"},{"url":"https://two.test"}]}`)
		var stdout, stderr bytes.Buffer
		if code := runOpenAPI([]string{"merge", "--dir", dir}, &stdout, &stderr); code != 0 {
			t.Fatalf("merge exit %d: %s", code, stderr.String())
		}
		if count := strings.Count(stdout.String(), "shared.test"); count != 1 {
			t.Fatalf("shared server should be deduplicated, appeared %d times: %s", count, stdout.String())
		}
	})

	t.Run("malformed parameter fails structural validation", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "bad.json", `{"components":{"parameters":{"Broken":{"name":"id","in":"not-a-location","schema":{"type":"string"}}}}}`)
		var stdout, stderr bytes.Buffer
		if code := runOpenAPI([]string{"merge", "--dir", dir}, &stdout, &stderr); code != 1 {
			t.Fatalf("malformed merged output must fail, got exit %d: %s", code, stdout.String())
		}
		if !strings.Contains(stderr.String(), "validation") {
			t.Fatalf("expected structural validation error, got %q", stderr.String())
		}
	})
}

func TestOpenAPIDiffSemanticCompatibility(t *testing.T) {
	fixtures := filepath.Join("testdata", "openapi-diff")
	tests := []struct {
		name        string
		current     string
		wantCode    int
		wantMessage string
	}{
		{"additive passes", "additive.json", 0, "0 breaking"},
		{"request requirement fails", "breaking-required-request.json", 1, "breaking"},
		{"response removal fails", "breaking-response-removal.json", 1, "breaking"},
		{"security weakening fails", "breaking-security.json", 1, "breaking"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runOpenAPI([]string{"diff", "--baseline", filepath.Join(fixtures, "baseline.json"), "--current", filepath.Join(fixtures, tt.current)}, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("diff exit %d, want %d; stdout=%s stderr=%s", code, tt.wantCode, stdout.String(), stderr.String())
			}
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, tt.wantMessage) {
				t.Fatalf("diff output %q missing %q", combined, tt.wantMessage)
			}
		})
	}
}
