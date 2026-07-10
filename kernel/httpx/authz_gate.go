package httpx

import (
	"context"
	"net/http"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// authz_gate.go — runtime enforcement of the RouteMeta permission gate (blueprint
// 07 §5, criterion #18). RouteMeta is validated at boot, but a route is only
// SAFE if its permission is also enforced PER REQUEST. This wraps each non-Public
// route with the deny-by-default chain: Authenticate → bind tenant+actor →
// evaluate the route permission at tenant scope → serve. A route the framework
// cannot authenticate or authorize never reaches the handler.

// Authenticator resolves a request's authenticated actor (carrying its tenant).
// The identity/tenant strategy (OIDC verification, tenant resolution) is
// deployment-specific, so a product supplies the concrete implementation; the
// framework never hardcodes the identity source. A returned error becomes the
// HTTP response (return a KindUnauthenticated error for a 401).
type Authenticator interface {
	Authenticate(r *http.Request) (authz.Actor, error)
}

// DenyAllAuthenticator rejects every request as unauthenticated. It is the
// SECURE DEFAULT: a freshly-scaffolded API enforces deny-by-default (every
// non-Public route → 401) until a real Authenticator is wired, rather than
// serving business routes unguarded.
type DenyAllAuthenticator struct{}

// Authenticate always fails closed.
func (DenyAllAuthenticator) Authenticate(*http.Request) (authz.Actor, error) {
	return authz.Actor{}, kerr.E(kerr.KindUnauthenticated, "unauthenticated",
		"no Authenticator is wired — deny-by-default (wire an OIDC/tenant Authenticator to serve business routes)")
}

// Composite tries each authenticator in order and returns the first successful
// actor — the way a product runs API-key and OIDC auth side by side (roadmap
// S1/CA-2). An authenticator that returns a KindUnauthenticated error is treated
// as "not my scheme" and the next is tried; any OTHER error (e.g. the key store
// is unreachable) short-circuits so a transient fault is not misreported as a
// clean 401. If every authenticator declines, the last unauthenticated error is
// returned. With no authenticators it fails closed.
func Composite(auths ...Authenticator) Authenticator {
	return compositeAuthenticator{auths: auths}
}

type compositeAuthenticator struct{ auths []Authenticator }

func (c compositeAuthenticator) Authenticate(r *http.Request) (authz.Actor, error) {
	var lastErr error = kerr.E(kerr.KindUnauthenticated, "unauthenticated", "no credentials accepted")
	for _, a := range c.auths {
		actor, err := a.Authenticate(r)
		if err == nil {
			return actor, nil
		}
		if kerr.KindOf(err) != kerr.KindUnauthenticated {
			return authz.Actor{}, err // hard fault: do not mask as a 401
		}
		lastErr = err
	}
	return authz.Actor{}, lastErr
}

// SecureHandler builds the serving mux with every route wrapped by the
// authN→authZ(RouteMeta) gate. Public routes are served directly; every other
// route is authenticated and authorized against its declared permission before
// its handler runs. Health endpoints (added by the caller after this) are Public
// infrastructure and are mounted directly.
func (r *Router) SecureHandler(auth Authenticator, eval authz.Evaluator, txm database.TxManager) *http.ServeMux {
	mux := http.NewServeMux()
	for _, rt := range r.routes {
		mux.Handle(rt.Method+" "+rt.Pattern, gateRoute(rt.Meta, rt.Handler, auth, eval, txm))
	}
	return mux
}

// gateRoute wraps one route handler with the deny-by-default gate.
func gateRoute(meta RouteMeta, h http.Handler, auth Authenticator, eval authz.Evaluator, txm database.TxManager) http.Handler {
	if meta.Public {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// 1. AuthN — resolve the actor + its tenant (401 on failure).
		actor, err := auth.Authenticate(req)
		if err != nil {
			WriteError(ctx, w, err)
			return
		}

		// 2. Bind tenant + actor so the tenant tx (RLS) and audit attribution are
		//    set for both the authz decision and the downstream handler. The full
		//    principal is also bound so per-actor guardrails (KeyByActor) can key on
		//    its strongest identifier rather than the uuid.Nil audit capacity a
		//    machine caller carries.
		ctx = database.WithTenantID(ctx, actor.TenantID)
		ctx = database.WithActorID(ctx, actor.CapacityID)
		ctx = WithActor(ctx, actor)

		// 3. AuthZ — evaluate the route permission at tenant scope in the request's
		//    tenant snapshot (deny-by-default; 403 on deny). Resource-scoped checks
		//    remain the handler's job against the concrete target.
		var decision authz.Decision
		aerr := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			decision, e = eval.Evaluate(ctx, db, actor, meta.Permission, authz.Target{Scope: authz.ScopeTenant})
			return e
		})
		if aerr != nil {
			WriteError(ctx, w, aerr)
			return
		}
		// Step-up: the actor is otherwise permitted but the permission demands an
		// elevated auth factor it has not satisfied — challenge for re-auth (401 +
		// WWW-Authenticate) rather than a flat 403 (roadmap S3). The advertised
		// factor comes from the policy's Decision.StepUpChallenge (per-permission
		// StepUpPolicy.Challenge, or the deployment's default), never hardcoded —
		// a permission requiring a hardware key advertises step_up="hwk", not "mfa".
		if decision.StepUpRequired {
			challenge := decision.StepUpChallenge
			if challenge == "" {
				challenge = "mfa" // defensive fallback; the evaluator always sets one
			}
			w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_user_authentication", step_up="`+challenge+`"`)
			WriteError(ctx, w, kerr.E(kerr.KindUnauthenticated, "step_up_required",
				"elevated authentication required for: "+meta.Permission))
			return
		}
		if !decision.Allowed {
			WriteError(ctx, w, kerr.E(kerr.KindForbidden, "permission_denied",
				"not permitted: "+meta.Permission))
			return
		}

		h.ServeHTTP(w, req.WithContext(ctx))
	})
}
