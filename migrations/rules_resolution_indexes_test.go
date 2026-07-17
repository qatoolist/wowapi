package migrations_test

import (
	"context"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/testkit"
)

func TestRuleVersionResolutionIndexes(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	var historyDefinition string
	if err := h.Admin.QueryRow(ctx,
		`SELECT indexdef
		   FROM pg_indexes
		  WHERE schemaname = 'public'
		    AND tablename = 'rule_versions'
		    AND indexname = 'rule_versions_history_resolution_idx'`).
		Scan(&historyDefinition); err != nil {
		t.Fatalf("query historical rule_versions index: %v", err)
	}
	for _, fragment := range []string{
		"(rule_key, scope_kind, scope_id, tenant_id, effective_from DESC)",
		"WHERE (status = ANY (ARRAY['active'::text, 'superseded'::text]))",
	} {
		if !strings.Contains(historyDefinition, fragment) {
			t.Errorf("historical index definition %q does not contain %q", historyDefinition, fragment)
		}
	}

	var currentConstraint string
	if err := h.Admin.QueryRow(ctx,
		`SELECT pg_get_constraintdef(oid)
		   FROM pg_constraint
		  WHERE conrelid = 'rule_versions'::regclass
		    AND contype = 'x'`).
		Scan(&currentConstraint); err != nil {
		t.Fatalf("query active rule_versions exclusion index: %v", err)
	}
	for _, fragment := range []string{"status = 'active'::text", "tstzrange(effective_from, effective_to)"} {
		if !strings.Contains(currentConstraint, fragment) {
			t.Errorf("active exclusion constraint %q does not contain %q", currentConstraint, fragment)
		}
	}

	var obsoleteLookupExists bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (
		     SELECT 1 FROM pg_indexes
		      WHERE schemaname = 'public'
		        AND tablename = 'rule_versions'
		        AND indexname = 'rule_versions_lookup'
		 )`).Scan(&obsoleteLookupExists); err != nil {
		t.Fatalf("query obsolete rule_versions lookup index: %v", err)
	}
	if obsoleteLookupExists {
		t.Error("obsolete active-only rule_versions_lookup index was not removed")
	}
}
