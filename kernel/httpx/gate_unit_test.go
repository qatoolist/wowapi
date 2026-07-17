package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

// TestGateAllowsAndBindsContext drives the gate's success path deterministically
// (fixed-allow evaluator, in-memory tx): an authorized actor reaches the handler
// with the tenant + actor bound in context. This is the non-DB companion to the
// DB-backed enforcement test and does not depend on grant-visibility timing.
func TestGateAllowsAndBindsContext(t *testing.T) {
	tenant := uuid.New()
	capID := uuid.New()
	act := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: capID, TenantID: tenant}

	router := httpx.NewRouter()
	handlerRan := false
	router.Handle(http.MethodGet, "/thing", httpx.RouteMeta{Permission: "core.thing.read"},
		func(w http.ResponseWriter, r *http.Request) {
			handlerRan = true
			if tid, ok := database.TenantIDFrom(r.Context()); !ok || tid != tenant {
				t.Errorf("handler tenant not bound: %v ok=%v", tid, ok)
			}
			if aid, ok := database.ActorIDFrom(r.Context()); !ok || aid != capID {
				t.Errorf("handler actor not bound: %v ok=%v", aid, ok)
			}
			w.WriteHeader(http.StatusNoContent)
		})
	if err := router.Err(); err != nil {
		t.Fatal(err)
	}

	mux := router.SecureHandler(fakeAuth{actor: act}, fakeEval{dec: authz.Decision{Allowed: true}}, fakeTxM{})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if !handlerRan {
		t.Fatal("authorized request must reach the handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("authorized route = %d, want 204", rec.Code)
	}
}

// TestGateEvaluatorErrorBecomesProblem proves an evaluator/transaction failure is
// surfaced as a problem-details response (not a silent allow): the gate returns
// the mapped status and the RFC 9457 content type.
func TestGateEvaluatorErrorBecomesProblem(t *testing.T) {
	act := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: uuid.New()}
	router := httpx.NewRouter()
	router.Handle(http.MethodGet, "/thing", httpx.RouteMeta{Permission: "core.thing.read"},
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if err := router.Err(); err != nil {
		t.Fatal(err)
	}

	evalErr := kerr.E(kerr.KindInternal, "boom", "evaluator exploded")
	mux := router.SecureHandler(fakeAuth{actor: act}, fakeEval{err: evalErr}, fakeTxM{})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("evaluator error = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}
