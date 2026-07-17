package ar03_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/appmodel"
)

// TestFullProjectionGolden verifies that the combined output of all 9 projections
// for the canonical requests module matches a strict golden baseline.
func TestFullProjectionGolden(t *testing.T) {
	m := appmodel.Manifest{
		ID:        "requests",
		Version:   "1.0.0",
		DependsOn: []string{},
		Routes: []appmodel.RouteDecl{
			{Method: "GET", Path: "/requests/healthz", Public: true},
			{Method: "POST", Path: "/requests", Permission: "requests.request.create"},
		},
		Permissions: []appmodel.PermissionDecl{
			{Key: "requests.request.create", Description: "Create a request"},
		},
		Resources: []appmodel.ResourceDecl{
			{Key: "requests.request", Description: "Request resource"},
		},
	}

	actualProjections := appmodel.GenerateProjections(m).String()

	expectedGolden := strings.TrimSpace(`
=== Routes ===
GET /requests/healthz (Permission: , Public: true)
POST /requests (Permission: requests.request.create, Public: false)

=== Permissions ===
requests.request.create: Create a request

=== Resources ===
requests.request: Request resource

=== Schema ===
Schema ref for POST /requests: RequestPayload

=== OpenAPI ===
paths:
  /requests/healthz:
    get:
      security: []
  /requests:
    post:
      security:
        - OAuth2: [requests.request.create]

=== Lifecycle ===
Module: requests
DependsOn: 

=== Profile ===
Profiles:
  API: routes=2
  Worker: jobs=0
  Migrate: migrations=true

=== Test ===
TestRoute: GET /requests/healthz (public=true)
TestRoute: POST /requests (public=false)

=== Doc ===
| Method | Path | Permission | Public |
|---|---|---|---|
| GET | /requests/healthz |  | true |
| POST | /requests | requests.request.create | false |
`)

	if strings.TrimSpace(actualProjections) != expectedGolden {
		t.Fatalf("Full Projection Golden Mismatch!\n--- GOT ---\n%s\n--- WANT ---\n%s", actualProjections, expectedGolden)
	}
}
