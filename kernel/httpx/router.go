package httpx

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/qatoolist/wowapi/internal/sealer"

	"github.com/qatoolist/wowapi/kernel/authz"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// ScopeExtractor derives the authorization target (org/resource id) from a
// request; the auth middleware (Phase 5) passes it to authz.Evaluate. Returning
// an error fails the request as a client error (review finding ARCH-43 — now
// typed to authz.Target).
type ScopeExtractor func(r *http.Request) (authz.Target, error)

// RouteMeta is mandatory metadata for every route. There is deliberately no
// registration path without it: a route is either guarded by a Permission or
// explicitly Public, never neither and never both (blueprint 05 §1).
type RouteMeta struct {
	// Permission is the permission key required to call the route. Empty only
	// when Public is true.
	Permission string
	// Public opts a route out of authz (health, pre-verification webhooks).
	Public bool
	// Scope derives the authz target from the request (optional).
	Scope ScopeExtractor
	// Idempotent enables idempotency-key handling for unsafe methods.
	Idempotent bool
	// Sensitive forces an audit record even for reads.
	Sensitive bool
	// Request declares the route's request-body contract for mutating verbs
	// (POST/PUT/PATCH) as a zero value of the request DTO, e.g.
	// `Request: CreateThingRequest{}` (FBL-08 / MATRIX CS-08). Wire the
	// handler through ValidatedHandler[CreateThingRequest] so declaring the
	// contract and running BindAndValidate are the same act. When the router
	// runs, a mutating route that declares neither Request nor NoRequestBody fails
	// registration. The declared value is deliberately a concrete prototype,
	// not a bool, so OpenAPI generation derives the request schema from the same
	// route table that runtime validation consumes.
	Request any
	// NoRequestBody waives the Request-contract requirement for a genuinely
	// body-less mutation (e.g. POST /things/{id}/archive). It is mutually
	// exclusive with Request. Kept a minimal marker on purpose: AR-04 T5
	// (W05, requirement-inventory.md row AR-04) will introduce the general
	// boot-time waiver mechanism; this field is designed additively so that
	// work can absorb or wrap it without breaking registrations.
	NoRequestBody bool
}

// validate enforces the metadata invariants. Returns a descriptive error used
// at registration time (which fails application boot).
func (m RouteMeta) validate() error {
	switch {
	case m.Public && m.Permission != "":
		return fmt.Errorf("route is marked Public but also sets Permission %q — choose one", m.Permission)
	case !m.Public && m.Permission == "":
		return fmt.Errorf("route has neither Permission nor Public — every route must declare one")
	}
	return nil
}

// Route is a registered route with its metadata, exposed for permission-sync
// and OpenAPI generation (later phases).
type Route struct {
	Method  string
	Pattern string
	Meta    RouteMeta
	Handler http.HandlerFunc
}

// Router collects routes with enforced metadata. Registration errors are
// accumulated and surfaced by Err/Build, so application boot fails with the
// full list (consistent with config/app validation).
type Router struct {
	routes []Route
	errs   []error
	seen   map[string]bool // method+pattern dedupe
	sealed bool
}

// Seal freezes the router once boot validation completes: any later Handle
// call panics rather than silently adding a route whose permission was never
// boot-validated (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (r *Router) Seal(sealer.Authority) { r.sealed = true }

// NewRouter returns an empty Router.
func NewRouter() *Router {
	return &Router{seen: map[string]bool{}}
}

// Handle registers a route. Invalid metadata (or a duplicate method+pattern)
// records an error retrievable via Err() — it does not panic, so a module's
// whole route set is validated at once.
func (r *Router) Handle(method, pattern string, meta RouteMeta, h http.HandlerFunc) {
	if r.sealed {
		panic("httpx: route registration after boot: the extension model is sealed")
	}
	if err := meta.validate(); err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s %s: %w", method, pattern, err))
		return
	}
	if err := r.checkRequestContract(method, meta); err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s %s: %w", method, pattern, err))
		return
	}
	if h == nil {
		r.errs = append(r.errs, fmt.Errorf("%s %s: nil handler", method, pattern))
		return
	}
	key := method + " " + pattern
	if r.seen[key] {
		r.errs = append(r.errs, fmt.Errorf("%s: registered more than once", key))
		return
	}
	r.seen[key] = true
	r.routes = append(r.routes, Route{Method: method, Pattern: pattern, Meta: meta, Handler: h})
}

// mutatingMethods are the verbs whose routes carry a request body by
// convention and therefore need a declared request contract. DELETE is
// deliberately absent: it is
// body-less by convention here, like GET/HEAD.
var mutatingMethods = map[string]bool{
	http.MethodPost:  true,
	http.MethodPut:   true,
	http.MethodPatch: true,
}

// checkRequestContract enforces the mutating-route contract invariants. The
// Request/NoRequestBody contradiction and a missing contract are invalid.
// This lives on Router rather than RouteMeta.validate() because the metadata
// alone does not know the HTTP method.
func (r *Router) checkRequestContract(method string, meta RouteMeta) error {
	if !mutatingMethods[method] {
		return nil
	}
	if meta.Request != nil && meta.NoRequestBody {
		return fmt.Errorf("route declares both a Request contract and NoRequestBody — choose one")
	}
	if meta.Request == nil && !meta.NoRequestBody {
		return fmt.Errorf("mutating route declares no request contract: set RouteMeta.Request to the request DTO's zero value (and wire the handler through ValidatedHandler), or set NoRequestBody for a genuinely body-less mutation")
	}
	return nil
}

// Routes returns the registered routes in a deterministic order (for
// permission sync, OpenAPI, and tests).
func (r *Router) Routes() []Route {
	out := append([]Route(nil), r.routes...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pattern != out[j].Pattern {
			return out[i].Pattern < out[j].Pattern
		}
		return out[i].Method < out[j].Method
	})
	return out
}

// Permissions returns the set of non-empty permission keys the routes require,
// sorted — the input to the Phase 4 permission-registration sync.
func (r *Router) Permissions() []string {
	set := map[string]struct{}{}
	for _, rt := range r.routes {
		if rt.Meta.Permission != "" {
			set[rt.Meta.Permission] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// Err returns the accumulated registration errors joined, or nil. Callers
// (app boot) must check this before serving.
func (r *Router) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	// Registration failures are programming/config errors, surfaced at boot.
	return kerr.E(kerr.KindInternal, "route_registration_failed", "route registration failed", joinErrs(r.errs))
}

func joinErrs(errs []error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	msg := ""
	for i, e := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += e.Error()
	}
	return fmt.Errorf("%s", msg)
}
