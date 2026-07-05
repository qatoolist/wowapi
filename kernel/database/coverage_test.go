package database_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/migrations"
	"github.com/qatoolist/wowapi/testkit"
)

// coverage_test.go raises kernel/database coverage with genuine, DB-asserting
// tests: TxManager commit/rollback + SET LOCAL tenant/actor binding, RLS
// isolation across tenants, read-only enforcement, the statement-timeout clamp,
// the full IdemStore lifecycle, cross-tenant sweep, and Migrate/MigrateReset.
// All exercise real Postgres effects — no mocks, no skips.

func tenantCtx(id uuid.UUID) context.Context {
	return database.WithTenantID(context.Background(), id)
}

// TestIntegrationWithTenantBindsTenantAndActor proves a read-write tenant
// transaction binds app.tenant_id AND app.actor_id with SET LOCAL, that work
// commits, and that the binding is observable via current_setting inside the tx.
func TestIntegrationWithTenantBindsTenantAndActor(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	tenant := uuid.New()
	actor := uuid.New()
	tctx := database.WithActorID(database.WithTenantID(ctx, tenant), actor)

	err := h.TxM.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		var boundTenant, boundActor string
		if err := db.QueryRow(ctx, "SELECT current_setting('app.tenant_id')").Scan(&boundTenant); err != nil {
			return err
		}
		if err := db.QueryRow(ctx, "SELECT current_setting('app.actor_id')").Scan(&boundActor); err != nil {
			return err
		}
		if boundTenant != tenant.String() {
			t.Errorf("app.tenant_id = %q, want %q", boundTenant, tenant)
		}
		if boundActor != actor.String() {
			t.Errorf("app.actor_id = %q, want %q", boundActor, actor)
		}
		// app_tenant_id() (used by every RLS policy) must resolve to the bound tenant.
		var fnTenant string
		if err := db.QueryRow(ctx, "SELECT app_tenant_id()::text").Scan(&fnTenant); err != nil {
			return err
		}
		if fnTenant != tenant.String() {
			t.Errorf("app_tenant_id() = %q, want %q", fnTenant, tenant)
		}
		// Exercise the tenant facade's Query method too.
		rows, err := db.Query(ctx, "SELECT generate_series(1,2)")
		if err != nil {
			return err
		}
		seen := 0
		for rows.Next() {
			seen++
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if seen != 2 {
			t.Errorf("tenant Query returned %d rows, want 2", seen)
		}
		// A write that must persist after commit.
		_, err = db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES (app_tenant_id(), 'system', 'commit-key', 'h', 'in_progress', now() + interval '1 hour')`)
		return err
	})
	if err != nil {
		t.Fatalf("WithTenant: %v", err)
	}

	// New tenant tx: the row committed and is visible to the same tenant.
	assertKeyCount(t, h, tenant, "commit-key", 1)
}

// TestIntegrationWithTenantRollsBackOnError proves that returning an error from
// fn rolls the whole transaction back — the inserted row must NOT survive.
func TestIntegrationWithTenantRollsBackOnError(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	sentinel := errors.New("boom")

	err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		if _, err := db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES (app_tenant_id(), 'system', 'rollback-key', 'h', 'in_progress', now() + interval '1 hour')`); err != nil {
			return err
		}
		return sentinel // triggers rollback
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
	assertKeyCount(t, h, tenant, "rollback-key", 0)
}

// TestIntegrationRLSIsolatesTenants proves FORCE row-level security: a row
// written under tenant A is invisible to tenant B in a separate transaction,
// even though both run on the same pooled connection as the same app_rt role.
func TestIntegrationRLSIsolatesTenants(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := uuid.New()
	tenantB := uuid.New()

	if err := h.TxM.WithTenant(tenantCtx(tenantA), func(ctx context.Context, db database.TenantDB) error {
		_, err := db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES (app_tenant_id(), 'system', 'secret', 'h', 'in_progress', now() + interval '1 hour')`)
		return err
	}); err != nil {
		t.Fatalf("write as A: %v", err)
	}

	// Tenant A sees its row.
	assertKeyCount(t, h, tenantA, "secret", 1)
	// Tenant B cannot see A's row — RLS filters it out entirely.
	assertKeyCount(t, h, tenantB, "secret", 0)

	// And B cannot write a row carrying A's tenant_id: the RLS WITH CHECK rejects it.
	err := h.TxM.WithTenant(tenantCtx(tenantB), func(ctx context.Context, db database.TenantDB) error {
		_, err := db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES ($1, 'system', 'spoof', 'h', 'in_progress', now() + interval '1 hour')`, tenantA)
		return err
	})
	if err == nil {
		t.Fatal("RLS WITH CHECK must reject a cross-tenant insert")
	}
}

// TestIntegrationWithTenantRORejectsWrites proves BEGIN READ ONLY: a write
// inside a read-only transaction fails at the database.
func TestIntegrationWithTenantRORejectsWrites(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()

	// A read is fine in RO.
	if err := h.TxM.WithTenantRO(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		var one int
		return db.QueryRow(ctx, "SELECT 1").Scan(&one)
	}); err != nil {
		t.Fatalf("RO read should succeed: %v", err)
	}

	// A write must be rejected by the read-only transaction.
	err := h.TxM.WithTenantRO(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		_, err := db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES (app_tenant_id(), 'system', 'ro-key', 'h', 'in_progress', now() + interval '1 hour')`)
		return err
	})
	if err == nil {
		t.Fatal("read-only transaction must reject a write")
	}
	// Nothing was written.
	assertKeyCount(t, h, tenant, "ro-key", 0)
}

// TestIntegrationStatementTimeoutClampAndSet proves inTx sets a tx-scoped
// statement_timeout from cfg.QueryTimeout, and that NewManager clamps a
// non-positive timeout to the compiled default (SEC-14) rather than disabling it.
func TestIntegrationStatementTimeoutClampAndSet(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()

	read := func(m database.TxManager) string {
		t.Helper()
		var setting string
		if err := m.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
			// pg_settings.setting for statement_timeout is the value in milliseconds.
			return db.QueryRow(ctx,
				"SELECT setting FROM pg_settings WHERE name = 'statement_timeout'").Scan(&setting)
		}); err != nil {
			t.Fatalf("read statement_timeout: %v", err)
		}
		return setting
	}

	// Explicit 3s -> 3000ms, tx-scoped.
	explicit := database.NewManager(h.Runtime,
		config.DB{Pool: config.Pool{MaxConns: 4, QueryTimeout: 3 * time.Second}},
		database.WithRole("app_rt"), database.WithRLSGuard())
	if got, want := read(explicit), "3000"; got != want {
		t.Fatalf("statement_timeout = %q, want %q", got, want)
	}

	// Non-positive -> clamped to the compiled default.
	clamped := database.NewManager(h.Runtime,
		config.DB{Pool: config.Pool{MaxConns: 4, QueryTimeout: 0}},
		database.WithRole("app_rt"), database.WithRLSGuard())
	wantMs := fmt.Sprintf("%d", config.Defaults().DB.QueryTimeout.Milliseconds())
	if got := read(clamped); got != wantMs {
		t.Fatalf("clamped statement_timeout = %q, want default %q", got, wantMs)
	}
}

// TestIntegrationPlatformTxRunsWithoutTenantBinding proves Platform runs the
// three DBTX methods against global scope with no app.tenant_id bound.
func TestIntegrationPlatformTxRunsWithoutTenantBinding(t *testing.T) {
	h := testkit.NewDB(t)

	if err := h.PlatformTxM.Platform(context.Background(), func(ctx context.Context, db database.DB) error {
		// No tenant is bound in a platform tx.
		var bound string
		if err := db.QueryRow(ctx, "SELECT COALESCE(current_setting('app.tenant_id', true), '')").Scan(&bound); err != nil {
			return err
		}
		if bound != "" {
			t.Errorf("platform tx bound app.tenant_id = %q, want empty", bound)
		}
		// Exercise Exec and Query on the platform facade.
		if _, err := db.Exec(ctx, "SELECT 1"); err != nil {
			return err
		}
		rows, err := db.Query(ctx, "SELECT generate_series(1,3)")
		if err != nil {
			return err
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			n++
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if n != 3 {
			t.Errorf("platform Query returned %d rows, want 3", n)
		}
		return nil
	}); err != nil {
		t.Fatalf("Platform: %v", err)
	}
}

// TestIntegrationIdemStoreLifecycle exercises Begin (Fresh -> Complete -> Found
// replay), the different-hash Conflict, the in-flight branch, Discard, and the
// Complete-on-vanished-key internal error — all against a real RLS-scoped table.
func TestIntegrationIdemStoreLifecycle(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	store := database.NewIdemStore()
	const scope, key = "system", "order-42"
	ttl := time.Hour

	// First Begin claims the key (Fresh).
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		r, err := store.Begin(ctx, db, scope, key, "hash-A", ttl)
		if err != nil {
			return err
		}
		if !r.Fresh {
			t.Errorf("first Begin = %+v, want Fresh", r)
		}
		// A second Begin on the same still-in_progress key (same tx) is in-flight.
		if _, err := store.Begin(ctx, db, scope, key, "hash-A", ttl); err == nil ||
			kerr.KindOf(err) != kerr.KindIdempotencyInFlight {
			t.Errorf("second Begin err = %v, want KindIdempotencyInFlight", err)
		}
		// Complete records the stored response in the same tx.
		return store.Complete(ctx, db, scope, key, 201, []byte(`{"ok":true}`))
	}); err != nil {
		t.Fatalf("claim+complete: %v", err)
	}

	// Replay: same key, same hash -> Found with the stored status/body.
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		r, err := store.Begin(ctx, db, scope, key, "hash-A", ttl)
		if err != nil {
			return err
		}
		if !r.Found || r.ResponseStatus != 201 || string(r.ResponseBody) != `{"ok":true}` {
			t.Errorf("replay = %+v, want Found/201/body", r)
		}
		return nil
	}); err != nil {
		t.Fatalf("replay: %v", err)
	}

	// Same key, different request hash -> Conflict.
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		_, err := store.Begin(ctx, db, scope, key, "hash-B", ttl)
		if err == nil || kerr.KindOf(err) != kerr.KindConflict {
			t.Errorf("mismatched hash err = %v, want KindConflict", err)
		}
		return nil
	}); err != nil {
		t.Fatalf("conflict tx: %v", err)
	}

	// Discard makes an in_progress claim retryable again.
	const dscope, dkey = "system", "discard-me"
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		if r, err := store.Begin(ctx, db, dscope, dkey, "h", ttl); err != nil || !r.Fresh {
			t.Fatalf("begin discardable: r=%+v err=%v", r, err)
		}
		return store.Discard(ctx, db, dscope, dkey)
	}); err != nil {
		t.Fatalf("discard tx: %v", err)
	}
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		r, err := store.Begin(ctx, db, dscope, dkey, "h", ttl)
		if err != nil {
			return err
		}
		if !r.Fresh {
			t.Errorf("after Discard, Begin = %+v, want Fresh", r)
		}
		return store.Discard(ctx, db, dscope, dkey) // leave it clean
	}); err != nil {
		t.Fatalf("re-begin after discard: %v", err)
	}

	// Complete on a key that was never claimed -> internal "vanished" error.
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		err := store.Complete(ctx, db, scope, "never-existed", 200, []byte("x"))
		if err == nil || kerr.KindOf(err) != kerr.KindInternal {
			t.Errorf("Complete on missing key err = %v, want KindInternal", err)
		}
		return nil
	}); err != nil {
		t.Fatalf("vanished-complete tx: %v", err)
	}
}

// TestIntegrationIdemStoreWriteErrorsSurface covers the DB-error branches of
// Begin, Complete, and Discard: run inside a READ ONLY transaction, each write
// is rejected by Postgres and the store must surface (wrap) that error rather
// than swallow it. Each runs in its own RO tx so the failure is the intended
// read-only rejection, not a prior aborted statement.
func TestIntegrationIdemStoreWriteErrorsSurface(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	store := database.NewIdemStore()
	const scope, key = "system", "ro-idem"

	runRO := func(op func(ctx context.Context, db database.TenantDB) error) error {
		var opErr error
		_ = h.TxM.WithTenantRO(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
			opErr = op(ctx, db)
			return nil
		})
		return opErr
	}

	if err := runRO(func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Begin(ctx, db, scope, key, "h", time.Hour)
		return e
	}); err == nil {
		t.Error("Begin in a read-only tx must surface a DB error")
	}
	if err := runRO(func(ctx context.Context, db database.TenantDB) error {
		return store.Complete(ctx, db, scope, key, 200, []byte("x"))
	}); err == nil {
		t.Error("Complete in a read-only tx must surface a DB error")
	}
	if err := runRO(func(ctx context.Context, db database.TenantDB) error {
		return store.Discard(ctx, db, scope, key)
	}); err == nil {
		t.Error("Discard in a read-only tx must surface a DB error")
	}
}

// TestIntegrationIdemStoreExpired proves an aged-out in_progress claim yields a
// defined KindIdempotencyExpired error rather than silently re-executing.
func TestIntegrationIdemStoreExpired(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()

	base := time.Now()
	clock := base
	store := database.NewIdemStoreWithClock(func() time.Time { return clock })
	const scope, key = "system", "expiring"

	// Claim with a tiny TTL at t0.
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		r, err := store.Begin(ctx, db, scope, key, "h", time.Millisecond)
		if err != nil {
			return err
		}
		if !r.Fresh {
			t.Errorf("claim = %+v, want Fresh", r)
		}
		return nil
	}); err != nil {
		t.Fatalf("claim: %v", err)
	}

	// Advance the injected clock well past expiry; the surviving in_progress row
	// is now expired.
	clock = base.Add(time.Hour)
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		_, err := store.Begin(ctx, db, scope, key, "h", time.Hour)
		if err == nil || kerr.KindOf(err) != kerr.KindIdempotencyExpired {
			t.Errorf("expired Begin err = %v, want KindIdempotencyExpired", err)
		}
		return nil
	}); err != nil {
		t.Fatalf("expired tx: %v", err)
	}
}

// TestIntegrationSweepExpired proves the cross-tenant sweep runs as app_platform
// (Platform tx, no tenant binding) and deletes only rows whose expires_at passed.
func TestIntegrationSweepExpired(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	store := database.NewIdemStore()

	// One expired row, one live row, for the same tenant.
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		if _, err := db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES (app_tenant_id(), 'system', 'stale', 'h', 'completed', now() - interval '1 hour')`); err != nil {
			return err
		}
		_, err := db.Exec(ctx,
			`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
			 VALUES (app_tenant_id(), 'system', 'live', 'h', 'completed', now() + interval '1 hour')`)
		return err
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	n, err := store.SweepExpired(context.Background(), h.PlatformTxM, time.Now())
	if err != nil {
		t.Fatalf("SweepExpired: %v", err)
	}
	if n < 1 {
		t.Fatalf("SweepExpired removed %d rows, want >= 1", n)
	}

	// The stale row is gone; the live row remains (visible to its tenant).
	assertKeyCount(t, h, tenant, "stale", 0)
	assertKeyCount(t, h, tenant, "live", 1)
}

// TestIntegrationRLSGuardRefusesSuperuser covers the tenant-transaction RLS
// guard's fail-closed branch: a Manager wired with WithRLSGuard over an
// over-privileged (superuser) connection must refuse to run tenant work, because
// FORCE RLS does not constrain a superuser (SEC-12).
func TestIntegrationRLSGuardRefusesSuperuser(t *testing.T) {
	h := testkit.NewDB(t)
	// h.Admin authenticates as the DSN's owner login (superuser here); with no
	// role re-bound, the effective role bypasses RLS.
	m := database.NewManager(h.Admin, config.Defaults().DB, database.WithRLSGuard())
	err := m.WithTenant(tenantCtx(uuid.New()), func(ctx context.Context, db database.TenantDB) error {
		t.Fatal("guard must refuse before fn runs")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "superuser or BYPASSRLS") {
		t.Fatalf("err = %v, want an RLS-guard refusal", err)
	}
}

// TestIntegrationSweepExpiredErrorsWithoutPlatformScope covers SweepExpired's
// error branch: run under a runtime (app_rt) manager, the cross-tenant DELETE's
// RLS policy calls app_tenant_id() with no tenant bound, which fails closed.
func TestIntegrationSweepExpiredErrorsWithoutPlatformScope(t *testing.T) {
	h := testkit.NewDB(t)
	store := database.NewIdemStore()
	// h.TxM is the app_rt manager; its Platform tx binds no tenant, so the
	// tenant-isolation policy's app_tenant_id() raises rather than sweeping.
	if _, err := store.SweepExpired(context.Background(), h.TxM, time.Now()); err == nil {
		t.Fatal("SweepExpired under app_rt (no tenant bound) must error, not sweep")
	}
}

// TestIntegrationMigrateAndReset exercises Migrate (idempotent head re-apply and
// fresh forward apply) and MigrateReset (full rollback) on an isolated database,
// asserting real schema effects. Safe for concurrent tests: bootstrap-down keeps
// the cluster-global roles.
func TestIntegrationMigrateAndReset(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	// Re-running Migrate on a head database is a no-op (Applied == 0).
	head, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("Migrate (head): %v", err)
	}
	if head.Applied != 0 {
		t.Errorf("Migrate on head Applied = %d, want 0 (idempotent)", head.Applied)
	}
	if head.Version == 0 {
		t.Fatal("head Version should be non-zero")
	}

	// Roll all the way back to 0.
	v, err := database.MigrateReset(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("MigrateReset: %v", err)
	}
	if v != 0 {
		t.Fatalf("after reset Version = %d, want 0", v)
	}

	// Roll forward from empty: a genuine apply (Applied > 0) back to head.
	reup, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("Migrate (re-up): %v", err)
	}
	if reup.Version != head.Version {
		t.Fatalf("re-up Version = %d, want %d", reup.Version, head.Version)
	}
	if reup.Applied != int(head.Version) {
		t.Errorf("re-up Applied = %d, want %d", reup.Applied, head.Version)
	}
}

// TestMigrateRequiresSourceName covers the source-name guard on both Migrate and
// MigrateReset (no database needed).
func TestMigrateRequiresSourceName(t *testing.T) {
	if _, err := database.Migrate(context.Background(), nil, migrations.Kernel(), ""); err == nil {
		t.Error("Migrate with empty source must error")
	}
	if _, err := database.MigrateReset(context.Background(), nil, migrations.Kernel(), ""); err == nil {
		t.Error("MigrateReset with empty source must error")
	}
}

// TestExpectOneRowTooBroad covers the "affected more than one row" branch:
// a versioned UPDATE that matches many rows is a programming bug, not a
// benign conflict, and must surface as a non-ErrVersionConflict error.
func TestExpectOneRowTooBroad(t *testing.T) {
	err := database.ExpectOneRow(pgconn.NewCommandTag("UPDATE 2"), "request")
	if err == nil {
		t.Fatal("2 rows must be an error")
	}
	if errors.Is(err, database.ErrVersionConflict) {
		t.Fatal("2 rows must NOT be masked as a version conflict")
	}
	if !strings.Contains(err.Error(), "WHERE clause too broad") {
		t.Fatalf("message = %q, want too-broad diagnostic", err.Error())
	}
}

// TestIntegrationInTxBeginErrorOnClosedPool covers inTx's BeginTx-error branch:
// a Manager over a closed pool fails to open the transaction, so fn never runs
// and the "database: begin" error is returned (fail closed).
func TestIntegrationInTxBeginErrorOnClosedPool(t *testing.T) {
	dsn := guardTestDSN(t)
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	pool.Close()

	m := database.NewManager(pool, config.Defaults().DB)
	err = m.WithTenant(tenantCtx(uuid.New()), func(ctx context.Context, db database.TenantDB) error {
		t.Fatal("fn must not run on a closed pool")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "begin") {
		t.Fatalf("err = %v, want a begin error", err)
	}

	// Platform path takes the same inTx door.
	if err := m.Platform(context.Background(), func(ctx context.Context, db database.DB) error {
		t.Fatal("platform fn must not run on a closed pool")
		return nil
	}); err == nil || !strings.Contains(err.Error(), "begin") {
		t.Fatalf("platform err = %v, want a begin error", err)
	}
}

// TestIntegrationMigrateErrorsOnClosedPool covers the Up/Down failure branches of
// Migrate and MigrateReset: a closed pool cannot run migrations.
func TestIntegrationMigrateErrorsOnClosedPool(t *testing.T) {
	dsn := guardTestDSN(t)
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	pool.Close()

	if _, err := database.Migrate(context.Background(), pool, migrations.Kernel(), migrations.SourceName); err == nil {
		t.Error("Migrate on a closed pool must error")
	}
	if _, err := database.MigrateReset(context.Background(), pool, migrations.Kernel(), migrations.SourceName); err == nil {
		t.Error("MigrateReset on a closed pool must error")
	}
}

// TestNewPoolChainedAfterConnectPropagatesError covers chainAfterConnect's
// prev-error branch: WithSetRole to a non-existent role fails during connect,
// and the chained WithConnRLSGuard step must never run — NewPool returns the
// underlying error.
func TestIntegrationNewPoolChainedAfterConnectPropagatesError(t *testing.T) {
	dsn := guardTestDSN(t)
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB,
		database.WithSetRole("no_such_role_xyz"), database.WithConnRLSGuard())
	if err == nil {
		pool.Close()
		t.Fatal("NewPool must fail when the first AfterConnect step errors")
	}
}

// assertKeyCount opens a fresh tenant transaction and asserts the number of
// idempotency_keys rows the tenant can see for idem_key — the canonical RLS
// visibility check.
func assertKeyCount(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, key string, want int) {
	t.Helper()
	if err := h.TxM.WithTenant(tenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		var n int
		if err := db.QueryRow(ctx,
			"SELECT count(*) FROM idempotency_keys WHERE idem_key = $1", key).Scan(&n); err != nil {
			return err
		}
		if n != want {
			t.Errorf("tenant %s sees %d rows for %q, want %d", tenant, n, key, want)
		}
		return nil
	}); err != nil {
		t.Fatalf("count query: %v", err)
	}
}
