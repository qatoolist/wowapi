package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationCatalogsSeededEmptyDBFailsClearly is the GAP-003 "clear failure
// mode" acceptance criterion: on a fresh, migrated-but-unseeded database, the
// check must fail with an actionable error naming the fix (wowapi seed sync /
// the generated migrate command) instead of a bare "count=0" or a downstream
// authorization denial with no diagnostic trail.
func TestIntegrationCatalogsSeededEmptyDBFailsClearly(t *testing.T) {
	h := testkit.NewDB(t)

	b := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{{Key: "widgets.widget.create", Description: "c"}},
	}
	err := app.CatalogsSeeded(context.Background(), h.Platform, b)
	if err == nil {
		t.Fatal("expected an error on an empty permissions catalog")
	}
	if !strings.Contains(err.Error(), "seed sync") {
		t.Fatalf("error must name the fix (seed sync): %v", err)
	}
}

// TestIntegrationCatalogsSeededAfterSyncPasses proves the check is satisfied once
// seeds.Sync has run — the normal post-deploy state.
func TestIntegrationCatalogsSeededAfterSyncPasses(t *testing.T) {
	h := testkit.NewDB(t)

	b := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{{Key: "widgets.widget.create", Description: "c"}},
	}
	if err := seeds.Sync(context.Background(), h.Platform, b); err != nil {
		t.Fatalf("seeds.Sync: %v", err)
	}
	if err := app.CatalogsSeeded(context.Background(), h.Platform, b); err != nil {
		t.Fatalf("CatalogsSeeded after Sync should pass: %v", err)
	}
}

// TestIntegrationCatalogsSeededPerCatalogTable exercises each of the four
// catalog tables independently, so a gap in any single one (not just
// permissions) is caught with the same actionable message.
func TestIntegrationCatalogsSeededPerCatalogTable(t *testing.T) {
	cases := []struct {
		name  string
		table string
		b     seeds.Bundle
	}{
		{"resource_types", "resource_types", seeds.Bundle{
			ResourceTypes: []seeds.ResourceTypeSeed{{Key: "widgets.widget", Description: "d"}},
		}},
		{"relationship_types", "relationship_types", seeds.Bundle{
			RelationshipTypes: []seeds.RelationshipTypeSeed{{Key: "widgets.owner_of", SubjectKind: "party", ObjectKind: "resource"}},
		}},
		{"roles", "roles", seeds.Bundle{
			Roles: []seeds.RoleSeed{{Key: "widgets.editor", Name: "Editor"}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testkit.NewDB(t)
			err := app.CatalogsSeeded(context.Background(), h.Platform, tc.b)
			if err == nil {
				t.Fatalf("expected an error on an empty %s catalog", tc.table)
			}
			if !strings.Contains(err.Error(), tc.table) || !strings.Contains(err.Error(), "seed sync") {
				t.Fatalf("error should name %s and seed sync: %v", tc.table, err)
			}
		})
	}
}

// TestCatalogsSeededEmptyBundleIsNotAnError covers a product with no seed-owning
// modules at all (no permissions/roles/resource_types/relationship_types
// declared anywhere) — there is nothing to have synced, so the check must not
// cry wolf.
func TestCatalogsSeededEmptyBundleIsNotAnError(t *testing.T) {
	h := testkit.NewDB(t)
	if err := app.CatalogsSeeded(context.Background(), h.Platform, seeds.Bundle{}); err != nil {
		t.Fatalf("an empty bundle should never fail the check: %v", err)
	}
}
