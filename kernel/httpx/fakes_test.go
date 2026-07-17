package httpx_test

import (
	"context"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
)

// Shared in-memory test doubles that let the httpx transaction-wrapping paths
// (authz gate, idempotency) be exercised deterministically without a database —
// avoiding both DB setup and the clock-skew that a grant-then-immediately-read
// integration flow is subject to.

// fakeTxM is an in-memory TxManager: it runs fn with a nil TenantDB so the code
// under test drives its own logic. A non-nil err short-circuits WithTenant /
// WithTenantRO, simulating a transaction begin/commit failure.
type fakeTxM struct{ err error }

func (f fakeTxM) run(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(ctx, nil)
}

func (f fakeTxM) WithTenant(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	return f.run(ctx, fn)
}

func (f fakeTxM) WithTenantRO(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	return f.run(ctx, fn)
}

func (f fakeTxM) Platform(ctx context.Context, fn func(context.Context, database.DB) error) error {
	return fn(ctx, nil)
}

// fakeEval is a fixed-decision authz.Evaluator for gate unit tests.
type fakeEval struct {
	dec authz.Decision
	err error
}

func (e fakeEval) Evaluate(context.Context, database.TenantDB, authz.Actor, string, authz.Target) (authz.Decision, error) {
	return e.dec, e.err
}

func (e fakeEval) Filter(context.Context, database.TenantDB, authz.Actor, string, string) (authz.ListFilter, error) {
	return authz.ListFilter{}, nil
}

// fakeIdem is an in-memory IdemStore recording which lifecycle calls fired so a
// test can assert Complete (2xx) vs Discard (non-2xx) behaviour.
type fakeIdem struct {
	begin     database.Replay
	beginErr  error
	completed bool
	discarded bool
}

func (f *fakeIdem) Begin(context.Context, database.TenantDB, string, string, string, time.Duration) (database.Replay, error) {
	return f.begin, f.beginErr
}

func (f *fakeIdem) Complete(context.Context, database.TenantDB, string, string, int, []byte) error {
	f.completed = true
	return nil
}

func (f *fakeIdem) Discard(context.Context, database.TenantDB, string, string) error {
	f.discarded = true
	return nil
}
