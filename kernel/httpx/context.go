package httpx

import (
	"context"

	"github.com/qatoolist/wowapi/kernel/authz"
)

// Request-scoped values that the middleware chain sets and handlers/helpers
// read. Kept minimal in Phase 3 (request id); auth/tenant/actor land in
// Phase 4 with their middleware.

type (
	requestIDKey struct{}
	actorKey     struct{}
)

// WithRequestID returns a context carrying the request correlation id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFrom returns the request id, or "" if none was set.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// WithActor returns a context carrying the full authenticated principal. The
// authz gate sets it after AuthN so downstream middleware/helpers (e.g.
// KeyByActor) can identify the caller by its strongest identifier — not just the
// audit CapacityID bound via database.WithActorID, which is uuid.Nil for every
// machine principal.
func WithActor(ctx context.Context, a authz.Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, a)
}

// ActorFrom extracts the authenticated principal; ok=false when none is bound.
func ActorFrom(ctx context.Context) (authz.Actor, bool) {
	a, ok := ctx.Value(actorKey{}).(authz.Actor)
	return a, ok
}
