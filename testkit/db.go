// Package testkit is wowapi's public integration-test harness: the one package
// permitted to compose everything (kernel, app, adapters, modules) so that both
// the framework and external product repositories can exercise their code
// against a real Postgres with the same fixtures, fakes, and assertions.
//
// Production packages MUST NOT import testkit (boundary lint). testkit MAY
// import kernel/*, app, adapters, and module.
//
// # Database strategy (D-0022)
//
// No testcontainers. The admin DSN comes from WOWAPI_TEST_DSN (fallback
// DATABASE_URL); tests skip with a clear message when neither is set. Kernel
// migrations run once per process into a content-addressed template database
// (wowapi_tmpl_<hash>); every test then gets an exclusive database cloned with
// CREATE DATABASE … TEMPLATE and dropped on cleanup. See docs/blueprint/08 §2
// and decisions D-0022/D-0023/D-0025.
package testkit

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/migrations"
)

// DBHandle is what NewDB hands a test: an admin pool (owner privileges, for
// fixtures/DDL) and a runtime pool (SET ROLE app_rt, RLS-enforced) on a
// database that is EXCLUSIVELY this test's.
type DBHandle struct {
	Name     string             // per-test database name
	Admin    *pgxpool.Pool      // owner credentials — fixtures, probe DDL
	Runtime  *pgxpool.Pool      // connects AS app_rt — what production code sees
	Platform *pgxpool.Pool      // connects AS app_platform — kernel/seed catalog writes
	TxM      database.TxManager // manager over Runtime

	// PlatformTxM is a tenant-bound TxManager over Platform (SET ROLE app_platform)
	// for kernel background work that mutates append-only-to-app_rt tables under a
	// bound tenant — e.g. document scan-status + retention voiding.
	PlatformTxM database.TxManager
}

// identRE guards every identifier this kit interpolates into DDL/DML. Test
// helpers still validate: a table name that reaches SQL string-building is a
// programming error, not user input, and we fail loudly rather than quote-and-pray.
var identRE = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// tmplAdvisoryKey serializes template creation across processes (parallel test
// packages) on the admin connection. Arbitrary fixed key ("wowa" bytes).
const tmplAdvisoryKey int64 = 0x776F7761

var (
	tmplOnce sync.Once
	tmplName string
	tmplErr  error

	// tmplCopyMu serializes CREATE DATABASE … TEMPLATE copies within this
	// process: the template is briefly source-locked during each copy.
	tmplCopyMu sync.Mutex
)

// NewDB provisions an exclusive database for t cloned from the migrated kernel
// template, returning admin + runtime pools and a TxManager over runtime. It
// skips (never fails) when no admin DSN is configured.
func NewDB(t *testing.T) *DBHandle {
	t.Helper()
	dsn := adminDSN(t)
	ctx := context.Background()

	tmpl, err := ensureTemplate(ctx, dsn)
	if err != nil {
		t.Fatalf("testkit: prepare template database: %v", err)
	}

	name := testDBName(t)
	createTestDB(ctx, t, dsn, name, tmpl)

	admin, err := newPoolDB(ctx, dsn, name, 4)
	if err != nil {
		dropTestDB(context.Background(), dsn, name)
		t.Fatalf("testkit: admin pool: %v", err)
	}
	// Runtime pool mirrors a deployed process: it connects AS the non-superuser
	// app_rt login (not a superuser doing SET ROLE), so even hostile in-tx SQL
	// like RESET ROLE cannot climb back to a privileged role (SEC-11).
	// WithConnRLSGuard refuses the connection if the effective role somehow
	// bypasses RLS (SEC-12); the TxManager re-asserts the role and guards each
	// transaction — the same wiring product api/worker processes use.
	runtime, err := runtimePoolDB(ctx, dsn, name, 4,
		database.WithConnRLSGuard())
	if err != nil {
		admin.Close()
		dropTestDB(context.Background(), dsn, name)
		t.Fatalf("testkit: runtime pool (as %s): %v", runtimeRole, err)
	}
	// Platform pool connects AS the non-superuser app_platform login: it holds
	// the catalog grants (SEC-13) but NOT app_rt's, so a contract seed sync runs
	// under exactly the privilege a real platform sync has — a seed needing a
	// grant app_platform lacks fails here instead of in production (SEC-33).
	platform, err := platformPoolDB(ctx, dsn, name, 2)
	if err != nil {
		runtime.Close()
		admin.Close()
		dropTestDB(context.Background(), dsn, name)
		t.Fatalf("testkit: platform pool (as %s): %v", platformRole, err)
	}

	h := &DBHandle{
		Name:     name,
		Admin:    admin,
		Runtime:  runtime,
		Platform: platform,
		TxM: database.NewManager(runtime, config.DB{Pool: config.Pool{MaxConns: 4, QueryTimeout: 5 * time.Second}},
			database.WithRole(runtimeRole), database.WithRLSGuard()),
		PlatformTxM: database.NewManager(platform, config.DB{Pool: config.Pool{MaxConns: 2, QueryTimeout: 5 * time.Second}},
			database.WithRole(platformRole), database.WithRLSGuard()),
	}

	t.Cleanup(func() {
		platform.Close()
		runtime.Close()
		admin.Close()
		dropTestDB(context.Background(), dsn, name)
	})
	return h
}

// adminDSN resolves the owner DSN or skips the test with actionable guidance.
func adminDSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("WOWAPI_TEST_DSN"); dsn != "" {
		return dsn
	}
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	t.Skip("testkit: no admin DSN. Run `make up`, then export " +
		"DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable " +
		"(or set WOWAPI_TEST_DSN).")
	return "" // unreachable; t.Skip halts the goroutine
}

// ensureTemplate builds the migrated template once per process.
func ensureTemplate(ctx context.Context, dsn string) (string, error) {
	tmplOnce.Do(func() {
		tmplName, tmplErr = buildTemplate(ctx, dsn)
	})
	return tmplName, tmplErr
}

// buildTemplate computes the content-addressed template name and, under a
// cross-process advisory lock, creates + migrates it if it does not yet exist.
func buildTemplate(ctx context.Context, dsn string) (string, error) {
	name, err := templateName()
	if err != nil {
		return "", err
	}

	conn, err := connectDB(ctx, dsn, "")
	if err != nil {
		return "", fmt.Errorf("connect admin: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	// Serialize template creation across processes: parallel packages race here.
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", tmplAdvisoryKey); err != nil {
		return "", fmt.Errorf("advisory lock: %w", err)
	}
	defer func() { _, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", tmplAdvisoryKey) }()

	var exists bool
	if err := conn.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return "", fmt.Errorf("check template exists: %w", err)
	}
	if !exists {
		if _, err := conn.Exec(ctx, "CREATE DATABASE "+quoteIdent(name)); err != nil {
			return "", fmt.Errorf("create template %s: %w", name, err)
		}
		// Migrate into the template through its own short-lived pool, then close
		// it so the template carries no open sessions (required for it to be a
		// copy source). Migrations create app_rt/app_platform cluster-wide.
		pool, err := newPoolDB(ctx, dsn, name, 2)
		if err != nil {
			return "", fmt.Errorf("template pool: %w", err)
		}
		_, mErr := database.Migrate(ctx, pool, migrations.Kernel(), migrations.SourceName)
		pool.Close()
		if mErr != nil {
			return "", fmt.Errorf("migrate template: %w", mErr)
		}
	}

	// Give app_rt a LOGIN so the runtime pool connects AS a genuine
	// non-superuser role — the production posture the review requires
	// (SEC-11/SEC-12). Runs after migration (which creates the role) in both
	// the fresh and reused paths; cluster-wide and idempotent. Kept out of the
	// committed migration on purpose: no runtime password ships to production,
	// where ops provision the app_rt login their own way.
	if err := alterRoleWithRetry(ctx, conn,
		"ALTER ROLE "+quoteIdent(runtimeRole)+" LOGIN PASSWORD "+quoteLiteral(runtimeRolePassword)); err != nil {
		return "", fmt.Errorf("provision %s login: %w", runtimeRole, err)
	}
	// Same for app_platform: the seed/catalog role, which holds the catalog
	// grants but not app_rt's (SEC-13/SEC-33).
	if err := alterRoleWithRetry(ctx, conn,
		"ALTER ROLE "+quoteIdent(platformRole)+" LOGIN PASSWORD "+quoteLiteral(platformRolePassword)); err != nil {
		return "", fmt.Errorf("provision %s login: %w", platformRole, err)
	}
	return name, nil
}

// alterRoleWithRetry runs a role-DDL statement, retrying the transient "tuple
// concurrently updated" (XX000) catalog race: app roles are CLUSTER-GLOBAL, so a
// parallel package's fresh migration (which also ALTERs the same role) can update
// the pg_authid tuple concurrently. The change is idempotent, so re-running after
// the winner commits succeeds.
func alterRoleWithRetry(ctx context.Context, conn *pgx.Conn, sql string) error {
	const attempts = 20
	var err error
	for i := 0; i < attempts; i++ {
		if _, err = conn.Exec(ctx, sql); err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "XX000" || !strings.Contains(pgErr.Message, "tuple concurrently updated") {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return err
}

// templateName is wowapi_tmpl_<8 hex> where the hash covers the sorted kernel
// migration filenames and contents, so the template regenerates when migrations
// change.
func templateName() (string, error) {
	src := migrations.Kernel()
	var names []string
	err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".sql") {
			names = append(names, p)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk migrations: %w", err)
	}
	sort.Strings(names)

	h := sha256.New()
	for _, n := range names {
		b, err := fs.ReadFile(src, n)
		if err != nil {
			return "", fmt.Errorf("read migration %s: %w", n, err)
		}
		h.Write([]byte(n))
		h.Write([]byte{0})
		h.Write(b)
	}
	// 8 bytes (64 bits) of the digest — wide enough that a collision serving a
	// stale template for changed migrations is not a practical concern (ARCH-21).
	return "wowapi_tmpl_" + hex.EncodeToString(h.Sum(nil)[:8]), nil
}

// createTestDB clones the template into name, retrying transient "being
// accessed by other users" contention on the shared source.
func createTestDB(ctx context.Context, t *testing.T, dsn, name, tmpl string) {
	t.Helper()
	conn, err := connectDB(ctx, dsn, "")
	if err != nil {
		t.Fatalf("testkit: connect admin: %v", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	stmt := fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", quoteIdent(name), quoteIdent(tmpl))

	tmplCopyMu.Lock()
	defer tmplCopyMu.Unlock()
	for attempt := 0; ; attempt++ {
		_, err = conn.Exec(ctx, stmt)
		if err == nil {
			return
		}
		if attempt >= 5 || !strings.Contains(err.Error(), "being accessed by other users") {
			t.Fatalf("testkit: create database %s: %v", name, err)
		}
		time.Sleep(time.Duration(50*(attempt+1)) * time.Millisecond)
	}
}

// createEmptyDB creates a brand-new, unmigrated database (no template) — used
// to exercise the genuinely-fresh migration path under assertion (ARCH-18).
func createEmptyDB(ctx context.Context, t *testing.T, dsn, name string) {
	t.Helper()
	conn, err := connectDB(ctx, dsn, "")
	if err != nil {
		t.Fatalf("testkit: connect admin: %v", err)
	}
	defer func() { _ = conn.Close(ctx) }()
	if _, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdent(name))); err != nil {
		t.Fatalf("testkit: create empty database %s: %v", name, err)
	}
}

// dropTestDB removes a per-test database. WITH (FORCE) (PG16) evicts any
// lingering sessions so cleanup never hangs.
func dropTestDB(ctx context.Context, dsn, name string) {
	conn, err := connectDB(ctx, dsn, "")
	if err != nil {
		return
	}
	defer func() { _ = conn.Close(ctx) }()
	_, _ = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(name)))
}

// testDBName derives a lowercase, <=63-char database name from the test name
// plus a random suffix for uniqueness across parallel runs.
func testDBName(t *testing.T) string {
	base := strings.ToLower(t.Name())
	base = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(base, "_")
	base = strings.Trim(base, "_")

	suffix := "_" + randHex(6)
	const prefix = "wt_"
	room := 63 - len(prefix) - len(suffix)
	if len(base) > room {
		base = base[:room]
	}
	return prefix + base + suffix
}

// runtimeRole is the non-superuser role the runtime pool authenticates as, so
// RLS binds it and it cannot escalate — the production posture (SEC-11/SEC-12).
// The password is local-test-only and never leaves the harness (it is set by
// buildTemplate via ALTER ROLE, not by any committed migration).
const (
	runtimeRole         = "app_rt"
	runtimeRolePassword = "app_rt_testkit_local"
	// platformRole is the catalog/seed role (app_platform): it holds the global
	// catalog grants but NOT app_rt's, so seed sync runs under the real
	// platform privilege (SEC-33). Local-test-only password, never committed.
	platformRole         = "app_platform"
	platformRolePassword = "app_platform_testkit_local"
)

// newPoolDB builds a pool against the base DSN with the database swapped and a
// small MaxConns; opts apply to the pgx config.
func newPoolDB(ctx context.Context, dsn, dbname string, maxConns int32, opts ...database.Option) (*pgxpool.Pool, error) {
	return buildPool(ctx, dsn, dbname, "", "", maxConns, opts...)
}

// runtimePoolDB builds a pool that authenticates as runtimeRole instead of the
// admin login, so tenant queries run under a genuinely RLS-bound identity.
func runtimePoolDB(ctx context.Context, dsn, dbname string, maxConns int32, opts ...database.Option) (*pgxpool.Pool, error) {
	return buildPool(ctx, dsn, dbname, runtimeRole, runtimeRolePassword, maxConns, opts...)
}

// platformPoolDB builds a pool that authenticates as app_platform — the catalog
// role — so seed sync runs under real platform privilege (SEC-33).
func platformPoolDB(ctx context.Context, dsn, dbname string, maxConns int32, opts ...database.Option) (*pgxpool.Pool, error) {
	return buildPool(ctx, dsn, dbname, platformRole, platformRolePassword, maxConns, opts...)
}

func buildPool(ctx context.Context, dsn, dbname, user, password string, maxConns int32, opts ...database.Option) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn (credentials withheld)")
	}
	pc.ConnConfig.Database = dbname
	if user != "" {
		pc.ConnConfig.User = user
		pc.ConnConfig.Password = password
	}
	pc.MaxConns = maxConns
	for _, o := range opts {
		o(pc)
	}
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, err
	}
	// Retry the first connection on SQLSTATE 28000 ("role is not permitted to log
	// in"): under parallel test packages, a fresh template build re-runs the
	// bootstrap migration which DROP+CREATEs the CLUSTER-GLOBAL app_rt/app_platform
	// roles as NOLOGIN before buildTemplate re-grants LOGIN — a connection landing
	// in that brief window transiently fails. It is not a real auth error, so we
	// wait it out rather than fail the whole package.
	if err := pingWithRoleRetry(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// pingWithRoleRetry pings, retrying only the transient "role not permitted to log
// in" (28000) provisioning race for up to ~2s.
func pingWithRoleRetry(ctx context.Context, pool *pgxpool.Pool) error {
	const attempts = 20
	var err error
	for i := 0; i < attempts; i++ {
		if err = pool.Ping(ctx); err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "28000" {
			return err // a real error, not the provisioning race
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return err
}

// quoteLiteral renders a single-quoted SQL string literal (for ALTER ROLE …
// PASSWORD, which cannot be parameterized). Doubles embedded quotes.
func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// connectDB opens a single (non-pooled) admin connection for CREATE/DROP
// DATABASE, which cannot run in a pooled transaction. dbname="" keeps the DSN's
// own database.
func connectDB(ctx context.Context, dsn, dbname string) (*pgx.Conn, error) {
	cc, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn (credentials withheld)")
	}
	if dbname != "" {
		cc.Database = dbname
	}
	return pgx.ConnectConfig(ctx, cc)
}

// quoteIdent double-quotes a validated identifier for interpolation.
func quoteIdent(id string) string { return pgx.Identifier{id}.Sanitize() }

// randHex returns n hex characters from crypto/rand.
func randHex(n int) string {
	b := make([]byte, (n+1)/2)
	if _, err := rand.Read(b); err != nil {
		panic("testkit: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)[:n]
}

// sortedKeys returns the map keys in stable order for deterministic SQL.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
