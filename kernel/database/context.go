package database

import (
	"context"

	"github.com/google/uuid"
)

// Tenant and actor identity travel by context from the middleware/worker
// layers into the TxManager, which binds them to the transaction with
// SET LOCAL. Repositories never read these keys — the database enforces
// isolation, not application filtering.

type tenantIDKey struct{}
type actorIDKey struct{}

// WithTenantID returns a context carrying the tenant the following database
// work is scoped to. Set by auth middleware (Phase 4) and job runners; tests
// set it directly.
func WithTenantID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, id)
}

// TenantIDFrom extracts the tenant id; ok=false when absent.
func TenantIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantIDKey{}).(uuid.UUID)
	return id, ok
}

// WithActorID returns a context carrying the acting user for audit
// attribution (app.actor_id).
func WithActorID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, actorIDKey{}, id)
}

// ActorIDFrom extracts the actor id; ok=false when absent.
func ActorIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(actorIDKey{}).(uuid.UUID)
	return id, ok
}
