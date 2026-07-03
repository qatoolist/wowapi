package httpx

import "context"

// Request-scoped values that the middleware chain sets and handlers/helpers
// read. Kept minimal in Phase 3 (request id); auth/tenant/actor land in
// Phase 4 with their middleware.

type requestIDKey struct{}

// WithRequestID returns a context carrying the request correlation id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFrom returns the request id, or "" if none was set.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
