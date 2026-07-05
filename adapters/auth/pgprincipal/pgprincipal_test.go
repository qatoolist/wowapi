package pgprincipal_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/adapters/auth/pgprincipal"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/testkit"
)

// These are real DB integration tests (no mocks): they run against the migrated
// kernel schema over the app_platform (global users) and app_rt (RLS-scoped
// acting_capacities) pools, proving the role split and that cross-tenant
// capacities are invisible under RLS.

// seedUser inserts a global user with a known idp subject via the owner pool.
func seedUser(t *testing.T, h *testkit.DBHandle, subject, status string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO users (id, idp_subject, email, status, created_by) VALUES ($1,$2,$3,$4,$5)`,
		id, subject, uuid.NewString()[:8]+"@example.test", status, uuid.Nil)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestUserIDBySubject(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	subject := "idp|active-" + uuid.NewString()[:8]
	userID := seedUser(t, h, subject, "active")

	got, err := store.UserIDBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("UserIDBySubject(active): %v", err)
	}
	if got != userID {
		t.Fatalf("user id: got %v want %v", got, userID)
	}

	// Unknown subject → opaque unauthenticated.
	if _, err := store.UserIDBySubject(ctx, "idp|nobody-"+uuid.NewString()[:8]); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("unknown subject: want KindUnauthenticated, got %v", err)
	}

	// Disabled user → unauthenticated with the SAME message (no oracle).
	disSubject := "idp|disabled-" + uuid.NewString()[:8]
	seedUser(t, h, disSubject, "disabled")
	if _, err := store.UserIDBySubject(ctx, disSubject); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("disabled subject: want KindUnauthenticated, got %v", err)
	}
}

func TestValidateCapacity(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	userID := seedUser(t, h, "idp|cap-"+uuid.NewString()[:8], "active")
	capID := testkit.CreateCapacity(t, h, tenant.ID, userID)

	// Valid capacity for this user in this tenant.
	if err := store.ValidateCapacity(ctx, userID, tenant.ID, capID); err != nil {
		t.Fatalf("ValidateCapacity(valid): %v", err)
	}

	// Unknown capacity id → forbidden.
	if err := store.ValidateCapacity(ctx, userID, tenant.ID, uuid.New()); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("unknown capacity: want KindForbidden, got %v", err)
	}

	// Capacity for a different user → forbidden.
	otherUser := seedUser(t, h, "idp|other-"+uuid.NewString()[:8], "active")
	if err := store.ValidateCapacity(ctx, otherUser, tenant.ID, capID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("foreign-user capacity: want KindForbidden, got %v", err)
	}

	// Cross-tenant: the capacity belongs to `tenant`, so under RLS it is invisible
	// when validating within a different tenant → forbidden (no cross-tenant leak).
	tenant2 := testkit.CreateTenant(t, h)
	if err := store.ValidateCapacity(ctx, userID, tenant2.ID, capID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("cross-tenant capacity must be invisible: want KindForbidden, got %v", err)
	}
}
