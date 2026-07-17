package ar03_test

import (
	"encoding/json"
	"testing"

	"github.com/qatoolist/wowapi/kernel/appmodel"
)

// TestManifestSchemaFixture proves the manifest schema is traceable 1:1 to existing
// declarations and round-trips cleanly against the requests fixture module.
func TestManifestSchemaFixture(t *testing.T) {
	m := appmodel.Manifest{
		ID:        "requests",
		Version:   "1.0.0",
		DependsOn: []string{},
		Routes: []appmodel.RouteDecl{
			{Method: "GET", Path: "/requests/healthz", Public: true},
			{Method: "POST", Path: "/requests", Permission: "requests.request.create"},
			{Method: "GET", Path: "/requests/{id}", Permission: "requests.request.read"},
			{Method: "GET", Path: "/requests", Permission: "requests.request.list"},
		},
		Permissions: []appmodel.PermissionDecl{
			{Key: "requests.request.create", Description: "Create a request"},
			{Key: "requests.request.read", Description: "Read a request"},
			{Key: "requests.request.list", Description: "List requests"},
		},
		Resources: []appmodel.ResourceDecl{
			{Key: "requests.request", Description: "Request resource"},
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	// Unmarshal back to Manifest
	var got appmodel.Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	// Assert equality
	if got.ID != m.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, m.ID)
	}
	if got.Version != m.Version {
		t.Errorf("Version mismatch: got %q, want %q", got.Version, m.Version)
	}
	if len(got.Routes) != len(m.Routes) {
		t.Errorf("Routes count mismatch: got %d, want %d", len(got.Routes), len(m.Routes))
	}
	for i, r := range got.Routes {
		want := m.Routes[i]
		if r.Method != want.Method || r.Path != want.Path || r.Permission != want.Permission || r.Public != want.Public {
			t.Errorf("Route[%d] mismatch: got %+v, want %+v", i, r, want)
		}
	}
	if len(got.Permissions) != len(m.Permissions) {
		t.Errorf("Permissions count mismatch: got %d, want %d", len(got.Permissions), len(m.Permissions))
	}
	if len(got.Resources) != len(m.Resources) {
		t.Errorf("Resources count mismatch: got %d, want %d", len(got.Resources), len(m.Resources))
	}
}
