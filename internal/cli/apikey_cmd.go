// apikey_cmd.go — wowapi apikey: issue, list, rotate, and revoke machine API
// keys / service principals (roadmap S1/CA-3). Tenant-scoped: connects as app_rt
// and binds the given tenant, so RLS applies exactly as at runtime. Issue and
// rotate print the plaintext token ONCE — it cannot be recovered later.
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/apikey"
	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
)

func apikeyUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi apikey <issue|list|rotate|revoke> --tenant <uuid> [args]

Manage machine API keys. Requires DATABASE_URL; connects as app_rt and binds the
given tenant (RLS applies). Issue/rotate print the token ONCE.

  wowapi apikey issue  --tenant <uuid> --name <name> [--scopes a.b.read,c.d.write] [--expires 720h]
  wowapi apikey list   --tenant <uuid>
  wowapi apikey rotate --tenant <uuid> --id <uuid>     mint a new secret (old stays valid until revoked)
  wowapi apikey revoke --tenant <uuid> --id <uuid>
`)
}

// runApikey implements `wowapi apikey`.
func runApikey(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		apikeyUsage(stderr)
		return 2
	}
	action := args[0]
	switch action {
	case "issue", "list", "rotate", "revoke":
	default:
		apikeyUsage(stderr)
		return 2
	}

	fs := flag.NewFlagSet("apikey "+action, flag.ContinueOnError)
	fs.SetOutput(stderr)
	tenant := fs.String("tenant", "", "tenant id (uuid)")
	name := fs.String("name", "", "key name (issue)")
	scopesRaw := fs.String("scopes", "", "comma-separated scopes (issue)")
	idRaw := fs.String("id", "", "key id (rotate/revoke)")
	expires := fs.Duration("expires", 0, "optional expiry from now, e.g. 720h (issue)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	tenantID, err := uuid.Parse(strings.TrimSpace(*tenant))
	if err != nil {
		fmt.Fprintln(stderr, "wowapi apikey: --tenant must be a uuid")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(stderr, "wowapi apikey: DATABASE_URL is not set")
		return 1
	}
	pool, err := database.NewPool(ctx, dsn, config.Defaults().DB,
		database.WithSetRole("app_rt"), database.WithConnRLSGuard())
	if err != nil {
		fmt.Fprintf(stderr, "wowapi apikey: %v\n", err)
		return 1
	}
	defer pool.Close()

	txm := database.NewManager(pool, config.Defaults().DB,
		database.WithRole("app_rt"), database.WithRLSGuard())
	store := apikey.NewStore(model.UUIDv7(), apikey.WithAudit(kaudit.New(model.UUIDv7(), nil)))
	tctx := database.WithTenantID(ctx, tenantID)

	switch action {
	case "issue":
		return apikeyIssue(tctx, txm, store, *name, *scopesRaw, *expires, stdout, stderr)
	case "rotate":
		return apikeyRotate(tctx, txm, store, *idRaw, tenantID, stdout, stderr)
	case "revoke":
		return apikeyRevoke(tctx, txm, store, *idRaw, stdout, stderr)
	case "list":
		return apikeyList(tctx, txm, store, stdout, stderr)
	}
	return 0
}

func apikeyIssue(ctx context.Context, txm database.TxManager, store *apikey.Store, name, scopesRaw string, expires time.Duration, stdout, stderr io.Writer) int {
	if name == "" {
		fmt.Fprintln(stderr, "wowapi apikey issue: --name is required")
		return 2
	}
	var scopes []string
	if s := strings.TrimSpace(scopesRaw); s != "" {
		scopes = strings.Split(s, ",")
	}
	var exp *time.Time
	if expires > 0 {
		t := time.Now().Add(expires)
		exp = &t
	}
	var token string
	var id uuid.UUID
	if err := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		token, id, e = store.Issue(ctx, db, name, scopes, exp)
		return e
	}); err != nil {
		fmt.Fprintf(stderr, "wowapi apikey issue: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "issued key %s\ntoken (shown once): %s\n", id, token)
	return 0
}

func apikeyRotate(ctx context.Context, txm database.TxManager, store *apikey.Store, idRaw string, tenantID uuid.UUID, stdout, stderr io.Writer) int {
	id, err := uuid.Parse(strings.TrimSpace(idRaw))
	if err != nil {
		fmt.Fprintln(stderr, "wowapi apikey rotate: --id must be a uuid")
		return 2
	}
	var token string
	var newID uuid.UUID
	if err := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		token, newID, e = store.Rotate(ctx, db, id)
		return e
	}); err != nil {
		fmt.Fprintf(stderr, "wowapi apikey rotate: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "rotated key %s -> %s\ntoken (shown once): %s\n", id, newID, token)
	fmt.Fprintf(stdout, "once callers are migrated, revoke the old key:\n  wowapi apikey revoke --tenant %s --id %s\n", tenantID, id)
	return 0
}

func apikeyRevoke(ctx context.Context, txm database.TxManager, store *apikey.Store, idRaw string, stdout, stderr io.Writer) int {
	id, err := uuid.Parse(strings.TrimSpace(idRaw))
	if err != nil {
		fmt.Fprintln(stderr, "wowapi apikey revoke: --id must be a uuid")
		return 2
	}
	if err := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return store.Revoke(ctx, db, id)
	}); err != nil {
		fmt.Fprintf(stderr, "wowapi apikey revoke: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "revoked key %s\n", id)
	return 0
}

func apikeyList(ctx context.Context, txm database.TxManager, store *apikey.Store, stdout, stderr io.Writer) int {
	var keys []apikey.KeyInfo
	if err := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		keys, e = store.List(ctx, db)
		return e
	}); err != nil {
		fmt.Fprintf(stderr, "wowapi apikey list: %v\n", err)
		return 1
	}
	if len(keys) == 0 {
		fmt.Fprintln(stdout, "no keys")
		return 0
	}
	fmt.Fprintf(stdout, "%-36s  %-20s  %-10s  %s\n", "ID", "NAME", "STATUS", "SCOPES")
	for _, k := range keys {
		status := "active"
		switch {
		case k.RevokedAt != nil:
			status = "revoked"
		case k.ExpiresAt != nil && !k.ExpiresAt.After(time.Now()):
			status = "expired"
		}
		fmt.Fprintf(stdout, "%-36s  %-20s  %-10s  %s\n", k.ID, k.Name, status, strings.Join(k.Scopes, ","))
	}
	return 0
}
