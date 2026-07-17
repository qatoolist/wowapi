package authz_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
)

// TestCredentialSchemeUserPermissionRejectsAPIKey proves SEC-01 T7: a
// permission scoped to CredentialUser rejects a valid, correctly-authenticated
// API-key actor. The API-key actor has the required scope but the permission's
// AllowedSchemes excludes api_key.
func TestCredentialSchemeUserPermissionRejectsAPIKey(t *testing.T) {
	const perm = "hr.payroll.export"
	reg := registry(t, authz.Permission{
		Key:            perm,
		AllowedSchemes: []authz.CredentialScheme{authz.CredentialUser},
	})
	store := &fakeStore{} // no RBAC assignments needed; API key has the scope
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	apiKeyActor := authz.Actor{
		Kind:             authz.ActorSystem,
		System:           "apikey:payroll-bot",
		TenantID:         uuid.New(),
		CredentialScheme: authz.CredentialAPIKey,
		Scopes:           []string{perm},
	}

	d, err := e.Evaluate(context.Background(), nil, apiKeyActor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || d.Reason != "credential_scheme_mismatch" {
		t.Fatalf("CredentialUser perm + api-key actor = %+v, want denied/credential_scheme_mismatch", d)
	}
}

// TestCredentialSchemeUserPermissionAllowsUser proves the positive path: a
// user actor satisfies a CredentialUser-scoped permission.
func TestCredentialSchemeUserPermissionAllowsUser(t *testing.T) {
	const perm = "hr.payroll.export"
	reg := registry(t, authz.Permission{
		Key:            perm,
		AllowedSchemes: []authz.CredentialScheme{authz.CredentialUser},
	})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "payroll", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	userActor := authz.Actor{
		Kind:             authz.ActorUser,
		UserID:           uuid.New(),
		TenantID:         uuid.New(),
		CapacityID:       uuid.New(),
		CredentialScheme: authz.CredentialUser,
	}

	d, err := e.Evaluate(context.Background(), nil, userActor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed {
		t.Fatalf("CredentialUser perm + user actor = %+v, want allowed", d)
	}
}

// TestCredentialSchemeNoRestrictionAllowsAnyScheme proves that a permission
// with no AllowedSchemes intentionally allows every explicit scheme.
func TestCredentialSchemeNoRestrictionAllowsAnyScheme(t *testing.T) {
	const perm = "gate.device.read"
	reg := registry(t, authz.Permission{Key: perm})
	store := &fakeStore{}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	apiKeyActor := authz.Actor{
		Kind:             authz.ActorSystem,
		System:           "apikey:gate-bot",
		TenantID:         uuid.New(),
		CredentialScheme: authz.CredentialAPIKey,
		Scopes:           []string{perm},
	}

	d, err := e.Evaluate(context.Background(), nil, apiKeyActor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed {
		t.Fatalf("no AllowedSchemes + api-key actor = %+v, want allowed", d)
	}
}

// TestCredentialSchemeMissingRejected proves that restricted permissions do
// not infer an authentication method from ActorKind or Scopes.
func TestCredentialSchemeMissingRejected(t *testing.T) {
	const perm = "hr.payroll.export"
	reg := registry(t, authz.Permission{
		Key:            perm,
		AllowedSchemes: []authz.CredentialScheme{authz.CredentialAPIKey},
	})
	store := &fakeStore{}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	// No explicit CredentialScheme: scopes do not imply an authentication method.
	apiKeyActor := authz.Actor{
		Kind:     authz.ActorSystem,
		System:   "apikey:payroll-bot",
		TenantID: uuid.New(),
		Scopes:   []string{perm},
	}

	d, err := e.Evaluate(context.Background(), nil, apiKeyActor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || d.Reason != "credential_scheme_mismatch" {
		t.Fatalf("missing scheme must be denied, got %+v", d)
	}
}

// TestCredentialSchemeWebhookRejectsUser proves that a webhook-only
// AllowedSchemes rejects a user actor.
func TestCredentialSchemeWebhookRejectsUser(t *testing.T) {
	const perm = "webhooks.inbox.read"
	reg := registry(t, authz.Permission{
		Key:            perm,
		AllowedSchemes: []authz.CredentialScheme{authz.CredentialWebhook},
	})
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "webhook", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	userActor := authz.Actor{
		Kind:             authz.ActorUser,
		UserID:           uuid.New(),
		TenantID:         uuid.New(),
		CapacityID:       uuid.New(),
		CredentialScheme: authz.CredentialUser,
	}

	d, err := e.Evaluate(context.Background(), nil, userActor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || d.Reason != "credential_scheme_mismatch" {
		t.Fatalf("Webhook-only perm + user actor = %+v, want denied/credential_scheme_mismatch", d)
	}
}

// TestCredentialSchemeInternalRejectsAPIKey proves that an internal-only
// AllowedSchemes rejects an API-key actor (which is ActorSystem with scopes).
func TestCredentialSchemeInternalRejectsAPIKey(t *testing.T) {
	const perm = "system.outbox.admin"
	reg := registry(t, authz.Permission{
		Key:            perm,
		AllowedSchemes: []authz.CredentialScheme{authz.CredentialInternal},
	})
	store := &fakeStore{}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	apiKeyActor := authz.Actor{
		Kind:             authz.ActorSystem,
		System:           "apikey:sweeper",
		TenantID:         uuid.New(),
		CredentialScheme: authz.CredentialAPIKey,
		Scopes:           []string{perm},
	}

	d, err := e.Evaluate(context.Background(), nil, apiKeyActor, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || d.Reason != "credential_scheme_mismatch" {
		t.Fatalf("Internal-only perm + api-key actor = %+v, want denied/credential_scheme_mismatch", d)
	}
}
