package authz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/policy"
)

// TestStepUpFreshnessStaleAuthTimeFails proves SEC-01 T6: a stale auth_time
// with an otherwise-valid strong AMR still fails step-up. The permission uses
// a StepUpPolicy with an explicit MaxAge; the actor's AuthTime is older than
// MaxAge but its AMR contains a qualifying strong factor.
func TestStepUpFreshnessStaleAuthTimeFails(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{
		Key: perm,
		StepUpPolicy: &authz.StepUpPolicy{
			RequiredAMR: []string{"mfa"},
			MaxAge:      5 * time.Minute,
			Challenge:   "mfa",
		},
	})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	stale := now.Add(-10 * time.Minute)

	actor := authz.Actor{
		Kind:       authz.ActorUser,
		UserID:     uuid.New(),
		TenantID:   uuid.New(),
		CapacityID: uuid.New(),
		AuthTime:   stale,
		AMR:        []string{"pwd", "mfa"},
	}

	d, err := e.Evaluate(context.Background(), nil, actor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired {
		t.Fatalf("stale auth_time with valid amr = %+v, want denied+StepUpRequired", d)
	}
	if d.Reason != "step_up_freshness_required" {
		t.Fatalf("reason = %q, want step_up_freshness_required", d.Reason)
	}
}

// TestStepUpFreshnessFreshAuthTimeAllows proves the positive path: an
// AuthTime within MaxAge with a valid AMR is allowed.
func TestStepUpFreshnessFreshAuthTimeAllows(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{
		Key: perm,
		StepUpPolicy: &authz.StepUpPolicy{
			RequiredAMR: []string{"mfa"},
			MaxAge:      5 * time.Minute,
			Challenge:   "mfa",
		},
	})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-2 * time.Minute)

	actor := authz.Actor{
		Kind:       authz.ActorUser,
		UserID:     uuid.New(),
		TenantID:   uuid.New(),
		CapacityID: uuid.New(),
		AuthTime:   fresh,
		AMR:        []string{"pwd", "mfa"},
	}

	d, err := e.Evaluate(context.Background(), nil, actor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("fresh auth_time with valid amr = %+v, want allowed", d)
	}
}

// TestStepUpFreshnessZeroAuthTimeFails proves that an unset (zero) AuthTime
// fails freshness when MaxAge is configured, because freshness cannot be
// verified.
func TestStepUpFreshnessZeroAuthTimeFails(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{
		Key: perm,
		StepUpPolicy: &authz.StepUpPolicy{
			RequiredAMR: []string{"mfa"},
			MaxAge:      5 * time.Minute,
			Challenge:   "mfa",
		},
	})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	actor := authz.Actor{
		Kind:       authz.ActorUser,
		UserID:     uuid.New(),
		TenantID:   uuid.New(),
		CapacityID: uuid.New(),
		AuthTime:   time.Time{}, // zero / unset
		AMR:        []string{"pwd", "mfa"},
	}

	d, err := e.Evaluate(context.Background(), nil, actor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired || d.Reason != "step_up_freshness_required" {
		t.Fatalf("zero auth_time with valid amr = %+v, want step_up_freshness_required", d)
	}
}

// TestStepUpFreshnessDefaultMaxAgeForShorthand proves the deployment default
// StepUpMaxAge applies to the plain `step_up: true` shorthand.
func TestStepUpFreshnessDefaultMaxAgeForShorthand(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := authz.New(authz.Options{
		Store: store, Registry: reg, Policies: policy.New(),
		Now:          func() time.Time { return time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC) },
		StepUpMaxAge: 5 * time.Minute,
	})
	target := authz.Target{Scope: authz.ScopeTenant}

	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	stale := now.Add(-10 * time.Minute)

	actor := authz.Actor{
		Kind:       authz.ActorUser,
		UserID:     uuid.New(),
		TenantID:   uuid.New(),
		CapacityID: uuid.New(),
		AuthTime:   stale,
		AMR:        []string{"pwd", "mfa"},
	}

	d, err := e.Evaluate(context.Background(), nil, actor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired || d.Reason != "step_up_freshness_required" {
		t.Fatalf("stale auth_time under default max age = %+v, want step_up_freshness_required", d)
	}
}

// TestStepUpFreshnessNoMaxAgeAllowsBackwardCompatibility proves that when no
// MaxAge is configured (neither per-policy nor deployment default), a zero
// AuthTime still allows step-up as long as AMR is satisfied. This preserves
// the pre-T6 behavior for deployments that do not enable freshness.
func TestStepUpFreshnessNoMaxAgeAllowsBackwardCompatibility(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	actor := authz.Actor{
		Kind:       authz.ActorUser,
		UserID:     uuid.New(),
		TenantID:   uuid.New(),
		CapacityID: uuid.New(),
		AuthTime:   time.Time{}, // zero / unset
		AMR:        []string{"pwd", "mfa"},
	}

	d, err := e.Evaluate(context.Background(), nil, actor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("no max age + zero auth_time + valid amr = %+v, want allowed", d)
	}
}
