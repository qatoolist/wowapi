package httpx_test

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/auth"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/testkit"
)

// dbPrincipalStore satisfies auth.PrincipalStore over real Postgres with the
// exact query logic of adapters/auth/pgprincipal.Store (platform pool for the
// global identity spine, tenant-bound RLS-read pool for acting_capacities).
// It is a test-local stand-in so this file can exercise the real DB round-trip
// without kernel tests importing adapters/ (boundary lint: kernel tests must
// not import adapters).
type dbPrincipalStore struct {
	platform database.TxManager
	runtime  database.TxManager
}

func (s dbPrincipalStore) UserIDBySubject(ctx context.Context, subject string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.platform.Platform(ctx, func(ctx context.Context, db database.DB) error {
		return db.QueryRow(
			ctx,
			`SELECT id FROM users WHERE idp_subject = $1 AND status = 'active'`, subject,
		).Scan(&id)
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, kerr.E(kerr.KindUnauthenticated, "unauthenticated",
				"unknown subject", kerr.Op("dbPrincipalStore.UserIDBySubject"))
		}
		return uuid.Nil, kerr.Wrapf(err, "dbPrincipalStore.UserIDBySubject", "load user")
	}
	return id, nil
}

func (s dbPrincipalStore) ActiveTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) error {
	var ok bool
	err := s.platform.Platform(ctx, func(ctx context.Context, db database.DB) error {
		return db.QueryRow(
			ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM user_tenant_access
			    WHERE user_id = $1 AND tenant_id = $2
			      AND status = 'active' AND valid_to IS NULL
			 )`, userID, tenantID,
		).Scan(&ok)
	})
	if err != nil {
		return kerr.Wrapf(err, "dbPrincipalStore.ActiveTenantAccess", "load tenant access")
	}
	if !ok {
		return kerr.E(kerr.KindForbidden, "permission_denied",
			"tenant access not permitted", kerr.Op("dbPrincipalStore.ActiveTenantAccess"))
	}
	return nil
}

func (s dbPrincipalStore) ValidateCapacity(ctx context.Context, userID, tenantID, capacityID uuid.UUID) error {
	ctx = database.WithTenantID(ctx, tenantID)
	var ok bool
	err := s.runtime.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(
			ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM acting_capacities
			    WHERE id = $1 AND user_id = $2 AND status = 'active' AND valid_to IS NULL
			 )`, capacityID, userID,
		).Scan(&ok)
	})
	if err != nil {
		return kerr.Wrapf(err, "dbPrincipalStore.ValidateCapacity", "load capacity")
	}
	if !ok {
		return kerr.E(kerr.KindForbidden, "permission_denied",
			"capacity not permitted", kerr.Op("dbPrincipalStore.ValidateCapacity"))
	}
	return nil
}

// ActiveCapacityCount returns the number of active capacities for the user in
// the tenant. This test store does not currently seed multiple capacities, so
// returning zero avoids triggering T4 capacity-selection enforcement.
func (s dbPrincipalStore) ActiveCapacityCount(ctx context.Context, userID, tenantID uuid.UUID) (int, error) {
	ctx = database.WithTenantID(ctx, tenantID)
	var count int
	err := s.runtime.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(
			ctx,
			`SELECT count(*) FROM acting_capacities
			  WHERE user_id = $1 AND status = 'active' AND valid_to IS NULL`, userID,
		).Scan(&count)
	})
	if err != nil {
		return 0, kerr.Wrapf(err, "dbPrincipalStore.ActiveCapacityCount", "load capacity count")
	}
	return count, nil
}

// ResolveGrant looks up a privileged-session grant by ID. The step-up e2e tests
// do not exercise privileged sessions, so any grant ID is treated as not found.
func (s dbPrincipalStore) ResolveGrant(ctx context.Context, userID, tenantID, grantID uuid.UUID) (*auth.ResolvedGrant, error) {
	_ = ctx
	_ = userID
	_ = tenantID
	_ = grantID
	return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionNotFound),
		"grant not found", kerr.Op("dbPrincipalStore.ResolveGrant"))
}

// seedUserWithSubject inserts a global user with a known idp_subject via the
// owner pool, so a test can mint a JWT whose sub the real auth.Verifier/
// PrincipalStore round-trip resolves to a known user (testkit.CreateUser
// generates a random, unobservable subject and cannot be used to mint tokens).
func seedUserWithSubject(t *testing.T, h *testkit.DBHandle, subject string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		id, subject, uuid.NewString()[:8]+"@example.test", uuid.Nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

// seedUserTenantAccess inserts a live membership row so the real auth.Verifier
// can pass the unconditional ActiveTenantAccess check introduced in W03-E01-S001.
func seedUserTenantAccess(t *testing.T, h *testkit.DBHandle, userID, tenantID uuid.UUID) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO user_tenant_access (id, user_id, tenant_id, status, created_by)
		 VALUES ($1, $2, $3, 'active', $4)`,
		uuid.New(), userID, tenantID, uuid.Nil); err != nil {
		t.Fatalf("seed user_tenant_access: %v", err)
	}
}

// TestIntegrationStepUpEndToEnd is the GAP-004 end-to-end proof: a permission
// declared step_up: true in seed YAML, synced through seeds.Sync into the real
// permissions catalog and the shared authz registry (mirroring what app.Boot
// does), gated behind a real JWT-authenticated route. An actor holding the RBAC
// grant but presenting a JWT with NO strong auth factor gets the step-up
// challenge (not a flat 403, not a flat allow); the SAME actor presenting a JWT
// minted with WithAMR("pwd","mfa") is allowed through. This proves the full
// seed → boot-propagation → JWT-claims → Actor.AMR → evaluator → HTTP-gate
// chain that wowsociety previously had to hand-wire (module.go direct
// registration + a JWT-reparsing authenticator wrapper).
func TestIntegrationStepUpEndToEnd(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	subject := "idp|stepup-" + uuid.NewString()[:8]
	userID := seedUserWithSubject(t, h, subject)
	seedUserTenantAccess(t, h, userID, tenant.ID)
	capID := testkit.CreateCapacity(t, h, tenant.ID, userID)

	const perm = "identity.impersonation.assign"

	// 1. Seed declares step_up: true and syncs into the real catalog — this is
	// the exact wowsociety workaround (manual migration + direct registry call)
	// GAP-004 replaces with a plain seed field.
	bundle := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{Key: perm, Description: "assign impersonation", StepUp: true},
		},
	}
	if err := seeds.Sync(ctx, h.Platform, bundle); err != nil {
		t.Fatalf("seeds.Sync: %v", err)
	}
	var stepUpCol bool
	if err := h.Admin.QueryRow(ctx, `SELECT step_up FROM permissions WHERE key = $1`, perm).Scan(&stepUpCol); err != nil {
		t.Fatalf("read step_up column: %v", err)
	}
	if !stepUpCol {
		t.Fatal("seeds.Sync did not persist step_up=true")
	}

	// 2. Registry propagation mirrors app.Boot's PermissionSeed -> authz.Permission
	// wiring (app/boot.go): StepUp must reach the registry entry the evaluator reads.
	reg := authz.NewRegistry()
	for _, p := range bundle.Permissions {
		reg.Register(authz.Permission{Key: p.Key, Sensitive: p.Sensitive, GrantedVia: p.GrantedVia, StepUp: p.StepUp})
	}
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	if got, _ := reg.Get(perm); !got.StepUp {
		t.Fatal("registry entry lost StepUp during propagation")
	}

	// 3. RBAC grant: the actor is otherwise permitted (real evaluator, real store).
	role := testkit.CreateRole(t, h, tenant.ID, "identity.impersonator", perm)
	testkit.GrantRole(t, h, tenant.ID, capID, role, "tenant", nil, "")
	eval := authz.New(authz.Options{Store: authz.NewStore(), Registry: reg, Policies: policy.New()})

	// 4. A real router + a real JWT authenticator (auth.Authenticator satisfies
	// httpx.Authenticator structurally) — no faked actor.
	router := httpx.NewRouter()
	router.Handle(http.MethodPost, "/impersonation/assign", httpx.RouteMeta{Permission: perm},
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if err := router.Err(); err != nil {
		t.Fatal(err)
	}

	ti := testkit.NewTokenIssuer()
	verifier := auth.NewVerifier(ti.KeySource(), auth.Config{Issuer: "wowapi-test", Audience: "wowapi"})
	principals := dbPrincipalStore{platform: h.PlatformTxM, runtime: h.TxM}
	authenticator := auth.NewAuthenticator(verifier, principals)

	mux := router.SecureHandler(authenticator, eval, h.TxM)
	do := func(tok string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/impersonation/assign", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	// --- Without a strong factor: step-up challenge, exact shape. ---
	noMFATok := ti.Issue(subject, tenant.ID, capID)
	rec := do(noMFATok)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no-AMR request: status = %d, want 401", rec.Code)
	}
	wantWA := `Bearer error="insufficient_user_authentication", step_up="mfa"`
	if got := rec.Header().Get("WWW-Authenticate"); got != wantWA {
		t.Fatalf("WWW-Authenticate = %q, want %q", got, wantWA)
	}
	if !strings.Contains(rec.Body.String(), "step_up_required") {
		t.Fatalf("body = %q, want it to mention step_up_required", rec.Body.String())
	}

	// --- With a strong factor via WithAMR: allowed. ---
	mfaTok := ti.Issue(subject, tenant.ID, capID, testkit.WithAMR("pwd", "mfa"))
	rec = do(mfaTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("with-AMR request: status = %d, want 200, body=%q", rec.Code, rec.Body.String())
	}
}

// TestIntegrationStepUpPolicyEndToEnd is TestIntegrationStepUpEndToEnd's B8
// extension: a permission declaring the RICHER seed form (step_up_amr: [hwk],
// step_up_challenge: hwk) propagates through the exact same
// seed→boot-propagation→JWT-claims→Actor.AMR→evaluator→HTTP-gate chain, and
// the evaluator enforces the SPECIFIC factor — a generic strong factor (otp)
// is NOT enough, only hwk satisfies it, and the challenge advertised is "hwk"
// (not a hardcoded "mfa").
func TestIntegrationStepUpPolicyEndToEnd(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	subject := "idp|stepup-hwk-" + uuid.NewString()[:8]
	userID := seedUserWithSubject(t, h, subject)
	seedUserTenantAccess(t, h, userID, tenant.ID)
	capID := testkit.CreateCapacity(t, h, tenant.ID, userID)

	const perm = "vault.secret.export"

	// 1. Seed declares the richer step-up form.
	bundle := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{
				Key: perm, Description: "export a vault secret", StepUp: true,
				StepUpAMR: []string{"hwk"}, StepUpChallenge: "hwk",
			},
		},
	}
	if err := seeds.Sync(ctx, h.Platform, bundle); err != nil {
		t.Fatalf("seeds.Sync: %v", err)
	}
	// permissions.step_up persists the plain bool; the AMR subset does not
	// round-trip through the DB (registry-declared only — see
	// authz.Permission.StepUpPolicy doc comment for the rationale).
	var stepUpCol bool
	if err := h.Admin.QueryRow(ctx, `SELECT step_up FROM permissions WHERE key = $1`, perm).Scan(&stepUpCol); err != nil {
		t.Fatalf("read step_up column: %v", err)
	}
	if !stepUpCol {
		t.Fatal("seeds.Sync did not persist step_up=true")
	}

	// 2. Registry propagation mirrors app.Boot's PermissionSeed -> authz.Permission
	// wiring, extended for the richer form (app/boot.go).
	reg := authz.NewRegistry()
	for _, p := range bundle.Permissions {
		perm := authz.Permission{Key: p.Key, Sensitive: p.Sensitive, GrantedVia: p.GrantedVia, StepUp: p.StepUp}
		if len(p.StepUpAMR) > 0 || p.StepUpChallenge != "" {
			perm.StepUpPolicy = &authz.StepUpPolicy{RequiredAMR: p.StepUpAMR, Challenge: p.StepUpChallenge}
		}
		reg.Register(perm)
	}
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}

	// 3. RBAC grant: the actor is otherwise permitted (real evaluator, real store).
	role := testkit.CreateRole(t, h, tenant.ID, "vault.exporter", perm)
	testkit.GrantRole(t, h, tenant.ID, capID, role, "tenant", nil, "")
	eval := authz.New(authz.Options{Store: authz.NewStore(), Registry: reg, Policies: policy.New()})

	router := httpx.NewRouter()
	router.Handle(http.MethodPost, "/vault/export", httpx.RouteMeta{Permission: perm},
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if err := router.Err(); err != nil {
		t.Fatal(err)
	}

	ti := testkit.NewTokenIssuer()
	verifier := auth.NewVerifier(ti.KeySource(), auth.Config{Issuer: "wowapi-test", Audience: "wowapi"})
	principals := dbPrincipalStore{platform: h.PlatformTxM, runtime: h.TxM}
	authenticator := auth.NewAuthenticator(verifier, principals)

	mux := router.SecureHandler(authenticator, eval, h.TxM)
	do := func(tok string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/vault/export", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	// --- A generic strong factor (otp) is NOT the specifically-required hwk. ---
	otpTok := ti.Issue(subject, tenant.ID, capID, testkit.WithAMR("pwd", "otp"))
	rec := do(otpTok)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("otp against a hwk-only policy: status = %d, want 401", rec.Code)
	}
	wantWA := `Bearer error="insufficient_user_authentication", step_up="hwk"`
	if got := rec.Header().Get("WWW-Authenticate"); got != wantWA {
		t.Fatalf("WWW-Authenticate = %q, want %q", got, wantWA)
	}

	// --- hwk satisfies it. ---
	hwkTok := ti.Issue(subject, tenant.ID, capID, testkit.WithAMR("pwd", "hwk"))
	rec = do(hwkTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("hwk against a hwk-only policy: status = %d, want 200, body=%q", rec.Code, rec.Body.String())
	}
}
