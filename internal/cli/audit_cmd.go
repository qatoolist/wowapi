// audit_cmd.go — wowapi audit: verify a tenant's tamper-evident audit chain
// offline (roadmap S6/CA-11). Connects as app_rt, binds the tenant, walks the
// hash chain, and reports any mutation (row hash mismatch) or deletion (seq gap).
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

	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
)

func auditUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi audit verify --tenant <uuid>

Verify a tenant's tamper-evident audit chain. Requires DATABASE_URL; connects as
app_rt and binds the tenant (RLS applies). Exit 0 = intact, 1 = tamper detected.

  wowapi audit verify --tenant <uuid>
`)
}

// runAudit implements `wowapi audit`.
func runAudit(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "verify" {
		auditUsage(stderr)
		return 2
	}

	fs := flag.NewFlagSet("audit verify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tenant := fs.String("tenant", "", "tenant id (uuid)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	tenantID, err := uuid.Parse(strings.TrimSpace(*tenant))
	if err != nil {
		fmt.Fprintln(stderr, "wowapi audit verify: --tenant must be a uuid")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(stderr, "wowapi audit verify: DATABASE_URL is not set")
		return 1
	}
	pool, err := database.NewPool(ctx, dsn, config.Defaults().DB,
		database.WithSetRole("app_rt"), database.WithConnRLSGuard())
	if err != nil {
		fmt.Fprintf(stderr, "wowapi audit verify: %v\n", err)
		return 1
	}
	defer pool.Close()
	txm := database.NewManager(pool, config.Defaults().DB,
		database.WithRole("app_rt"), database.WithRLSGuard())

	w := kaudit.New(model.UUIDv7(), nil)
	var res kaudit.VerifyResult
	if err := txm.WithTenantRO(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		var e error
		res, e = w.Verify(ctx, db)
		return e
	}); err != nil {
		fmt.Fprintf(stderr, "wowapi audit verify: %v\n", err)
		return 1
	}

	if res.OK {
		fmt.Fprintf(stdout, "OK: audit chain intact (%d rows, head seq %d)\n", res.Count, res.HeadSeq)
		return 0
	}
	fmt.Fprintf(stderr, "TAMPER DETECTED at seq %d: %s (%d rows checked)\n", res.BrokenSeq, res.Reason, res.Count)
	return 1
}
