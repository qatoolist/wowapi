package authz_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
)

// A StepUp permission challenges an otherwise-allowed actor that lacks a strong
// auth factor, and admits it once the AMR carries one (roadmap S3).
func TestStepUpChallenge(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	// The actor HAS an RBAC grant for the permission at tenant scope.
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	// Without a strong factor: allowed-but-for-MFA → step-up challenge, not allowed.
	noMFA := authz.Actor{Kind: authz.ActorUser, TenantID: actor.TenantID, UserID: actor.UserID, CapacityID: actor.CapacityID}
	d, err := e.Evaluate(context.Background(), nil, noMFA, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired || d.Reason != "step_up_required" {
		t.Fatalf("no-MFA on a StepUp perm = %+v, want denied+StepUpRequired", d)
	}

	// With a strong factor in the AMR: allowed.
	withMFA := noMFA
	withMFA.AMR = []string{"pwd", "mfa"}
	d, err = e.Evaluate(context.Background(), nil, withMFA, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("MFA-satisfied on a StepUp perm = %+v, want allowed", d)
	}
}

// A StepUp permission the actor was NOT granted is a plain deny — step-up only
// gates an otherwise-allowed decision, it never surfaces for an unauthorized one.
func TestStepUpDoesNotMaskPlainDeny(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	e := newEval(t, &fakeStore{}, reg, nil, nil) // no assignments → not granted

	d, err := e.Evaluate(context.Background(), nil, actor, perm, authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || d.StepUpRequired || d.Reason != "default_deny" {
		t.Fatalf("ungranted StepUp perm = %+v, want plain default_deny (no step-up)", d)
	}
}

// A permission requiring a hardware key SPECIFICALLY (StepUpPolicy.RequiredAMR
// = ["hwk"]) is NOT satisfied by a lesser factor like otp — the per-permission
// AMR subset is enforced, not just "any strong factor" (B8 acceptance:
// per-permission AMR subsets).
func TestStepUpPolicyRequiresSpecificFactor(t *testing.T) {
	const perm = "vault.secret.export"
	reg := registry(t, authz.Permission{
		Key: perm, StepUp: true,
		StepUpPolicy: &authz.StepUpPolicy{RequiredAMR: []string{"hwk"}, Challenge: "hwk"},
	})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "vault.admin", ScopeKind: "tenant", Perms: []string{perm}},
	}}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	// otp is a strong factor in general, but NOT the specifically-required hwk.
	withOTP := authz.Actor{Kind: authz.ActorUser, TenantID: actor.TenantID, UserID: actor.UserID, CapacityID: actor.CapacityID, AMR: []string{"pwd", "otp"}}
	d, err := e.Evaluate(context.Background(), nil, withOTP, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired {
		t.Fatalf("otp against a hwk-only StepUpPolicy = %+v, want denied+StepUpRequired", d)
	}
	if d.StepUpChallenge != "hwk" {
		t.Errorf("StepUpChallenge = %q, want %q (the policy's own challenge, not the default)", d.StepUpChallenge, "hwk")
	}

	// hwk itself satisfies it.
	withHWK := withOTP
	withHWK.AMR = []string{"pwd", "hwk"}
	d, err = e.Evaluate(context.Background(), nil, withHWK, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("hwk against a hwk-only StepUpPolicy = %+v, want allowed", d)
	}
}

// sms is EXCLUDED from the default strong-factor set (Decision 5): a plain
// `step_up: true` permission is NOT satisfied by amr=[sms] under the default
// configuration.
func TestStepUpDefaultExcludesSMS(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}},
	}}
	e := newEval(t, store, reg, nil, nil) // default Options: no StrongFactors override
	withSMS := authz.Actor{Kind: authz.ActorUser, TenantID: actor.TenantID, UserID: actor.UserID, CapacityID: actor.CapacityID, AMR: []string{"pwd", "sms"}}

	d, err := e.Evaluate(context.Background(), nil, withSMS, perm, authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired {
		t.Fatalf("sms-only AMR under the DEFAULT strong-factor set = %+v, want denied+StepUpRequired (sms is opt-in only)", d)
	}
}

// A deployment can opt sms BACK IN via configuration alone (Options.StrongFactors
// / kernel.Deps.StepUpStrongFactors) — no code changes — and amr=[sms] then
// satisfies the same `step_up: true` permission.
func TestStepUpSMSOptInViaConfig(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}},
	}}
	// Deployment config: default set PLUS sms, still no code change to the
	// evaluator or the permission — just the configured strong-factor list.
	factors := append([]string{"sms"}, authz.DefaultStrongFactors...)
	e := newEvalWithFactors(t, store, reg, factors, "")
	withSMS := authz.Actor{Kind: authz.ActorUser, TenantID: actor.TenantID, UserID: actor.UserID, CapacityID: actor.CapacityID, AMR: []string{"pwd", "sms"}}

	d, err := e.Evaluate(context.Background(), nil, withSMS, perm, authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("sms-only AMR with sms opted into the configured strong-factor set = %+v, want allowed", d)
	}
}

// The default strong-factor set itself excludes sms (documents the exact
// membership Decision 5 mandates, independent of the evaluator's behavior).
func TestDefaultStrongFactorsExcludeSMS(t *testing.T) {
	for _, f := range authz.DefaultStrongFactors {
		if f == "sms" {
			t.Fatalf("DefaultStrongFactors must not include sms (opt-in only): %v", authz.DefaultStrongFactors)
		}
	}
	want := map[string]bool{"mfa": true, "otp": true, "totp": true, "hwk": true, "fpt": true, "face": true}
	if len(authz.DefaultStrongFactors) != len(want) {
		t.Fatalf("DefaultStrongFactors = %v, want exactly %v", authz.DefaultStrongFactors, want)
	}
	for _, f := range authz.DefaultStrongFactors {
		if !want[f] {
			t.Errorf("unexpected default strong factor %q", f)
		}
	}
}

// A plain `step_up: true` permission (no StepUpPolicy) advertises the
// deployment's DEFAULT challenge hint when it can be overridden via config.
func TestStepUpDefaultChallengeIsConfigurable(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}},
	}}
	e := newEvalWithFactors(t, store, reg, nil, "otp")
	noFactor := authz.Actor{Kind: authz.ActorUser, TenantID: actor.TenantID, UserID: actor.UserID, CapacityID: actor.CapacityID}

	d, err := e.Evaluate(context.Background(), nil, noFactor, perm, authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if !d.StepUpRequired || d.StepUpChallenge != "otp" {
		t.Fatalf("StepUpChallenge = %q (StepUpRequired=%v), want %q", d.StepUpChallenge, d.StepUpRequired, "otp")
	}
}
