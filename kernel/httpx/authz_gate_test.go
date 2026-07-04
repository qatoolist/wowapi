package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/testkit"
)

// authz_gate_test.go — the runtime enforcement gap (Finding 1). Proves the
// framework now ENFORCES the RouteMeta permission per request: Public routes are
// open, unauthenticated requests 401, unauthorized 403, and an authorized actor
// reaches the handler with the tenant + actor bound in context.

// fakeAuth returns a fixed actor (or nil to fall back to deny).
type fakeAuth struct{ actor authz.Actor }

func (f fakeAuth) Authenticate(*http.Request) (authz.Actor, error) { return f.actor, nil }

func TestIntegrationAuthzGateEnforcesRoutePermission(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)

	const perm = "core.thing.read"
	testkit.CreatePermission(t, h, perm, false)

	// A real evaluator over the DB store.
	reg := authz.NewRegistry()
	reg.Register(authz.Permission{Key: perm})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	eval := authz.New(authz.Options{Store: authz.NewStore(), Registry: reg, Policies: policy.New()})

	// A router with a Public route and a guarded route.
	router := httpx.NewRouter()
	router.Handle(http.MethodGet, "/public", httpx.RouteMeta{Public: true},
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	router.Handle(http.MethodGet, "/thing", httpx.RouteMeta{Permission: perm},
		func(w http.ResponseWriter, r *http.Request) {
			// The handler must see the bound tenant + actor.
			if tid, ok := database.TenantIDFrom(r.Context()); !ok || tid != tn.ID {
				t.Errorf("handler tenant not bound: %v ok=%v", tid, ok)
			}
			if aid, ok := database.ActorIDFrom(r.Context()); !ok || aid != capID {
				t.Errorf("handler actor not bound: %v ok=%v", aid, ok)
			}
			w.WriteHeader(http.StatusOK)
		})
	if err := router.Err(); err != nil {
		t.Fatal(err)
	}

	act := authz.Actor{Kind: authz.ActorUser, UserID: userID, CapacityID: capID, TenantID: tn.ID}

	serve := func(auth httpx.Authenticator, path string) int {
		mux := router.SecureHandler(auth, eval, h.TxM)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Code
	}

	// Public route: served without authentication.
	if code := serve(httpx.DenyAllAuthenticator{}, "/public"); code != http.StatusOK {
		t.Fatalf("public route = %d, want 200", code)
	}
	// Guarded route, no authenticator wired → 401 (fail-closed default).
	if code := serve(httpx.DenyAllAuthenticator{}, "/thing"); code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated guarded route = %d, want 401", code)
	}
	// Authenticated but NOT granted the permission → 403.
	if code := serve(fakeAuth{actor: act}, "/thing"); code != http.StatusForbidden {
		t.Fatalf("unauthorized guarded route = %d, want 403", code)
	}

	// Grant the actor a tenant-scope role carrying the permission → 200.
	role := testkit.CreateRole(t, h, tn.ID, "core.reader", perm)
	testkit.GrantRole(t, h, tn.ID, capID, role, "tenant", nil, "")
	if code := serve(fakeAuth{actor: act}, "/thing"); code != http.StatusOK {
		t.Fatalf("authorized guarded route = %d, want 200", code)
	}
}
