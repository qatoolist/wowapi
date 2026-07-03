package testkit

import (
	"context"
	"fmt"
	"testing"
)

// CreateProbeTable creates a minimal tenant-scoped table following every
// convention from 03 §1 (tenant_id, ENABLE + FORCE RLS, standard policy, grants
// to app_rt) so RLS mechanics can be proven before real tenant tables ship
// (D-0025). Returns the table name.
//
// The table is created through h.Admin (owner). FORCE ROW LEVEL SECURITY is
// what makes RLS apply even though the runtime role reaches the table via
// SET ROLE from a superuser login — the policy is enforced against app_rt.
func CreateProbeTable(t *testing.T, h *DBHandle) string {
	t.Helper()
	name := "probe_" + randHex(8)
	q := quoteIdent(name)

	stmts := []string{
		fmt.Sprintf(`CREATE TABLE %s (
			id        uuid PRIMARY KEY,
			tenant_id uuid NOT NULL,
			note      text NOT NULL DEFAULT ''
		)`, q),
		fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL SECURITY`, q),
		fmt.Sprintf(`ALTER TABLE %s FORCE ROW LEVEL SECURITY`, q),
		fmt.Sprintf(`CREATE POLICY %s ON %s
			USING (tenant_id = app_tenant_id())
			WITH CHECK (tenant_id = app_tenant_id())`, quoteIdent(name+"_tenant_isolation"), q),
		fmt.Sprintf(`GRANT SELECT, INSERT, UPDATE, DELETE ON %s TO app_rt`, q),
	}
	for _, s := range stmts {
		if _, err := h.Admin.Exec(context.Background(), s); err != nil {
			t.Fatalf("testkit: create probe table: %v\n%s", err, s)
		}
	}
	return name
}
