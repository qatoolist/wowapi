package ar03_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/appmodel"
)

// TestGoldenDeclarationDelta proves that a manifest change deterministically produces
// the expected full projection diff (across all 9 projection types) with no other files.
func TestGoldenDeclarationDelta(t *testing.T) {
	mA := appmodel.Manifest{
		ID:        "requests",
		Version:   "1.0.0",
		DependsOn: []string{},
		Routes: []appmodel.RouteDecl{
			{Method: "GET", Path: "/requests/healthz", Public: true},
			{Method: "POST", Path: "/requests", Permission: "requests.request.create"},
			{Method: "GET", Path: "/requests/{id}", Permission: "requests.request.read"},
		},
		Permissions: []appmodel.PermissionDecl{
			{Key: "requests.request.create", Description: "Create a request"},
			{Key: "requests.request.read", Description: "Read a request"},
		},
		Resources: []appmodel.ResourceDecl{
			{Key: "requests.request", Description: "Request resource"},
		},
	}

	// mB is mA with a new route and its corresponding permission added
	mB := appmodel.Manifest{
		ID:        "requests",
		Version:   "1.1.0",
		DependsOn: []string{},
		Routes: []appmodel.RouteDecl{
			{Method: "GET", Path: "/requests/healthz", Public: true},
			{Method: "POST", Path: "/requests", Permission: "requests.request.create"},
			{Method: "GET", Path: "/requests/{id}", Permission: "requests.request.read"},
			{Method: "PATCH", Path: "/requests/{id}/cancel", Permission: "requests.request.cancel"}, // Added
		},
		Permissions: []appmodel.PermissionDecl{
			{Key: "requests.request.create", Description: "Create a request"},
			{Key: "requests.request.read", Description: "Read a request"},
			{Key: "requests.request.cancel", Description: "Cancel a request"}, // Added
		},
		Resources: []appmodel.ResourceDecl{
			{Key: "requests.request", Description: "Request resource"},
		},
	}
	pA := appmodel.GenerateProjections(mA).String()
	pB := appmodel.GenerateProjections(mB).String()
	actualDiff := computeDiff(pA, pB)
	expectedGoldenDiff := strings.TrimSpace(`
--- Base (v1.0.0)
+++ Mutated (v1.1.0)
@@ Routes @@
+ PATCH /requests/{id}/cancel (Permission: requests.request.cancel, Public: false)
@@ Permissions @@
+ requests.request.cancel: Cancel a request
@@ Schema @@
+ Schema ref for PATCH /requests/{id}/cancel: RequestPayload
@@ OpenAPI @@
+   /requests/{id}/cancel:
+     patch:
+         - OAuth2: [requests.request.cancel]
@@ Profile @@
-   API: routes=3
+   API: routes=4
@@ Test @@
+ TestRoute: PATCH /requests/{id}/cancel (public=false)
@@ Doc @@
+ | PATCH | /requests/{id}/cancel | requests.request.cancel | false |
`)

	if actualDiff != expectedGoldenDiff {
		t.Fatalf("Golden Delta Mismatch!\n--- GOT ---\n%s\n--- WANT ---\n%s", actualDiff, expectedGoldenDiff)
	}
}

func computeDiff(a, b string) string {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	var diff []string
	diff = append(diff, "--- Base (v1.0.0)", "+++ Mutated (v1.1.0)")

	// Simple structural diffing helper targeted at our custom header sections
	headers := []string{
		"=== Routes ===",
		"=== Permissions ===",
		"=== Resources ===",
		"=== Schema ===",
		"=== OpenAPI ===",
		"=== Lifecycle ===",
		"=== Profile ===",
		"=== Test ===",
		"=== Doc ===",
	}

	for _, h := range headers {
		sectionA := getSection(linesA, h)
		sectionB := getSection(linesB, h)

		hasDiff := false
		var secDiff []string

		// Check for deleted or changed lines
		for _, la := range sectionA {
			found := false
			for _, lb := range sectionB {
				if la == lb {
					found = true
					break
				}
			}
			if !found && la != "" {
				secDiff = append(secDiff, "- "+la)
				hasDiff = true
			}
		}

		// Check for added lines
		for _, lb := range sectionB {
			found := false
			for _, la := range sectionA {
				if la == lb {
					found = true
					break
				}
			}
			if !found && lb != "" {
				secDiff = append(secDiff, "+ "+lb)
				hasDiff = true
			}
		}

		if hasDiff {
			headerName := strings.Trim(h, "= ")
			diff = append(diff, "@@ "+headerName+" @@")
			diff = append(diff, secDiff...)
		}
	}

	return strings.TrimSpace(strings.Join(diff, "\n"))
}

func getSection(lines []string, header string) []string {
	var out []string
	capture := false
	for _, l := range lines {
		if strings.HasPrefix(l, "===") {
			if l == header {
				capture = true
			} else {
				capture = false
			}
			continue
		}
		if capture {
			out = append(out, l)
		}
	}
	return out
}
