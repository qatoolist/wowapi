package testkit

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
)

// AssertAllowed fails the test unless the evaluator allows actor to exercise
// perm on target. The check runs inside a tenant transaction (the production
// path) scoped to the actor's tenant.
func AssertAllowed(t *testing.T, h *DBHandle, e authz.Evaluator, a authz.Actor, perm string, target authz.Target) {
	t.Helper()
	d := evaluate(t, h, e, a, perm, target)
	if !d.Allowed {
		t.Fatalf("expected ALLOW for %s on %v, got deny (%s)", perm, target.Scope, d.Reason)
	}
}

// AssertDenied fails the test unless the evaluator denies actor perm on target.
func AssertDenied(t *testing.T, h *DBHandle, e authz.Evaluator, a authz.Actor, perm string, target authz.Target) {
	t.Helper()
	d := evaluate(t, h, e, a, perm, target)
	if d.Allowed {
		t.Fatalf("expected DENY for %s on %v, got allow (%s)", perm, target.Scope, d.Reason)
	}
}

func evaluate(t *testing.T, h *DBHandle, e authz.Evaluator, a authz.Actor, perm string, target authz.Target) authz.Decision {
	t.Helper()
	var d authz.Decision
	err := h.TxM.WithTenantRO(database.WithTenantID(context.Background(), a.TenantID),
		func(ctx context.Context, db database.TenantDB) error {
			var e2 error
			d, e2 = e.Evaluate(ctx, db, a, perm, target)
			return e2
		})
	if err != nil {
		t.Fatalf("evaluate %s: %v", perm, err)
	}
	return d
}
