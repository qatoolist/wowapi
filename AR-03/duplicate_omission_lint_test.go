package ar03_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/appmodel"
)

// TestLint_DuplicateIdentities asserts the lint rule fails on duplicate routes,
// permissions, or resources (hand-maintained duplicate identity).
func TestLint_DuplicateIdentities(t *testing.T) {
	m := appmodel.Manifest{
		ID:        "requests",
		Version:   "1.0.0",
		DependsOn: []string{},
		Routes: []appmodel.RouteDecl{
			{Method: "GET", Path: "/requests", Public: true},
			{Method: "GET", Path: "/requests", Public: true}, // Duplicate
		},
		Permissions: []appmodel.PermissionDecl{
			{Key: "requests.request.create", Description: "Create"},
			{Key: "requests.request.create", Description: "Create again"}, // Duplicate
		},
		Resources: []appmodel.ResourceDecl{
			{Key: "requests.request", Description: "Resource"},
			{Key: "requests.request", Description: "Resource again"}, // Duplicate
		},
	}

	violations := appmodel.LintManifest(m)
	if len(violations) == 0 {
		t.Fatal("expected lint failures for duplicates, got none")
	}

	var hasRoute, hasPerm, hasRes bool
	for _, v := range violations {
		if strings.Contains(v, "duplicate route identity") {
			hasRoute = true
		}
		if strings.Contains(v, "duplicate permission identity") {
			hasPerm = true
		}
		if strings.Contains(v, "duplicate resource identity") {
			hasRes = true
		}
	}

	if !hasRoute || !hasPerm || !hasRes {
		t.Fatalf("missing expected duplicate violations: route=%t, perm=%t, res=%t\nall violations:\n%v", hasRoute, hasPerm, hasRes, violations)
	}
}

// TestLint_OmittedProjection asserts the lint rule fails on an omitted projection
// where a route references an undeclared permission in the catalog.
func TestLint_OmittedProjection(t *testing.T) {
	m := appmodel.Manifest{
		ID:        "requests",
		Version:   "1.0.0",
		DependsOn: []string{},
		Routes: []appmodel.RouteDecl{
			{Method: "POST", Path: "/requests", Permission: "requests.request.create"},
		},
		Permissions: []appmodel.PermissionDecl{
			// requests.request.create is omitted!
		},
	}

	violations := appmodel.LintManifest(m)
	if len(violations) == 0 {
		t.Fatal("expected lint failure for omitted projection, got none")
	}

	foundOmitted := false
	for _, v := range violations {
		if strings.Contains(v, "omitted projection") {
			foundOmitted = true
			break
		}
	}

	if !foundOmitted {
		t.Fatalf("expected omitted projection violation, got:\n%v", violations)
	}
}
