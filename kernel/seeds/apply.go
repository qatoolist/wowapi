package seeds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// ApplyOptions tunes the behaviour of Apply.
type ApplyOptions struct {
	// DryRun, when true, computes and prints a change plan without writing
	// anything to the database. No audit record is produced.
	DryRun bool
	// Actor is recorded in the audit row (e.g. "migrate", "wowapi-cli").
	Actor string
	// Out receives the human-readable dry-run plan. Ignored when DryRun is false.
	Out io.Writer
	// Invalidators are invoked after catalog writes commit (same contract as Sync).
	Invalidators []SpineInvalidator
}

// ApplyCounts records how many catalog entities the bundle declares. It is
// stored as the audit row's counts JSONB.
type ApplyCounts struct {
	Permissions       int `json:"permissions"`
	ResourceTypes     int `json:"resource_types"`
	RelationshipTypes int `json:"relationship_types"`
	Roles             int `json:"roles"`
}

// ChangePlan is the read-only diff used for dry-run reporting.
type ChangePlan struct {
	Permissions       TablePlan
	ResourceTypes     TablePlan
	RelationshipTypes TablePlan
	Roles             RolePlan
}

// TablePlan is the per-entity change classification for a single catalog table.
type TablePlan struct {
	Insert    int `json:"insert"`
	Update    int `json:"update"`
	Unchanged int `json:"unchanged"`
}

// RolePlan extends TablePlan with grant reconcile details.
type RolePlan struct {
	TablePlan
	GrantAdds   int `json:"grant_adds"`
	GrantPrunes int `json:"grant_prunes"`
}

// Report describes the outcome of an Apply run.
type Report struct {
	Hash         string
	VersionLabel string
	Outcome      string // applied | noop | failed | dry_run
	Counts       ApplyCounts
	ChangePlan   ChangePlan
}

const (
	advisoryKey1 int64 = 0x73656564 // "seed"
	advisoryKey2 int64 = 0x73796E63 // "sync"
)

// Apply is the production seed-sync entrypoint: it computes the bundle's
// content hash, serializes concurrent callers with a transaction-scoped
// advisory lock, short-circuits to a no-op when the database already reflects
// the same hash, and otherwise upserts the catalogs and records an audit row
// atomically in one transaction.
//
// Apply accepts either *pgxpool.Pool (it will begin and commit its own tx) or
// an already-open pgx.Tx (it will neither commit nor roll it back). A bare
// database.DBTX that is neither is rejected.
func Apply(ctx context.Context, db database.DBTX, b Bundle, opts ApplyOptions) (Report, error) {
	hash := Hash(b)
	report := Report{
		Hash:         hash,
		VersionLabel: b.Version,
		Outcome:      "noop",
	}

	if opts.DryRun {
		report.Outcome = "dry_run"
		plan, err := computeDryRun(ctx, db, b, opts.Out)
		report.ChangePlan = plan
		return report, err
	}

	tx, own, err := beginTx(ctx, db)
	if err != nil {
		return report, err
	}
	if own {
		defer func() { _ = tx.Rollback(ctx) }()
	}

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1, $2)", advisoryKey1, advisoryKey2); err != nil {
		return report, kerr.Wrapf(err, "seeds.Apply", "advisory lock")
	}

	latest := latestRunHash(ctx, tx)
	if latest != "" && latest == hash && !bundleEmpty(b) {
		report.Outcome = "noop"
		if own {
			if cerr := tx.Commit(ctx); cerr != nil {
				return report, kerr.Wrapf(cerr, "seeds.Apply", "commit")
			}
		}
		return report, nil
	}

	if err := Sync(ctx, tx, b, opts.Invalidators...); err != nil {
		_ = recordFailedRun(ctx, db, report, opts.Actor, err)
		return report, err
	}

	counts := countBundle(b)
	report.Outcome = "applied"
	report.Counts = counts
	if err := insertRun(ctx, tx, report, opts.Actor, counts); err != nil {
		return report, err
	}

	if own {
		if cerr := tx.Commit(ctx); cerr != nil {
			return report, kerr.Wrapf(cerr, "seeds.Apply", "commit")
		}
	}
	return report, nil
}

// HashOf reports the canonical content hash of the bundle without applying it.
func HashOf(b Bundle) string { return Hash(b) }

func beginTx(ctx context.Context, db database.DBTX) (pgx.Tx, bool, error) {
	switch d := db.(type) {
	case pgx.Tx:
		return d, false, nil
	case *pgxpool.Pool:
		tx, err := d.Begin(ctx)
		if err != nil {
			return nil, false, kerr.Wrapf(err, "seeds.Apply", "begin tx")
		}
		return tx, true, nil
	default:
		return nil, false, fmt.Errorf("seeds.Apply: db must be *pgxpool.Pool or pgx.Tx, got %T", db)
	}
}

func latestRunHash(ctx context.Context, db database.DBTX) string {
	var hash string
	err := db.QueryRow(ctx,
		`SELECT manifest_hash FROM seed_sync_runs
		  WHERE outcome IN ('applied','noop')
		  ORDER BY created_at DESC LIMIT 1`).Scan(&hash)
	if err != nil {
		// No rows is the expected state on a fresh database.
		return ""
	}
	return hash
}

func insertRun(ctx context.Context, db database.DBTX, r Report, actor string, counts ApplyCounts) error {
	data, err := json.Marshal(counts)
	if err != nil {
		return kerr.Wrapf(err, "seeds.Apply", "marshal counts")
	}
	_, err = db.Exec(ctx,
		`INSERT INTO seed_sync_runs (manifest_hash, version_label, actor, outcome, counts)
		 VALUES ($1, $2, $3, $4, $5)`,
		r.Hash, r.VersionLabel, actor, r.Outcome, data)
	if err != nil {
		return kerr.Wrapf(err, "seeds.Apply", "record sync run")
	}
	return nil
}

func recordFailedRun(ctx context.Context, db database.DBTX, r Report, actor string, runErr error) error {
	// If the caller supplied a transaction, the failure is already rolling back;
	// inserting outside that tx is not safe/sensible here.
	pool, ok := db.(*pgxpool.Pool)
	if !ok {
		return nil
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO seed_sync_runs (manifest_hash, version_label, actor, outcome, error)
		 VALUES ($1, $2, $3, 'failed', $4)`,
		r.Hash, r.VersionLabel, actor, runErr.Error())
	return err // best-effort
}

func countBundle(b Bundle) ApplyCounts {
	return ApplyCounts{
		Permissions:       len(b.Permissions),
		ResourceTypes:     len(b.ResourceTypes),
		RelationshipTypes: len(b.RelationshipTypes),
		Roles:             len(b.Roles),
	}
}

func bundleEmpty(b Bundle) bool {
	return len(b.Permissions) == 0 && len(b.ResourceTypes) == 0 &&
		len(b.RelationshipTypes) == 0 && len(b.Roles) == 0
}

func computeDryRun(ctx context.Context, db database.DBTX, b Bundle, out io.Writer) (ChangePlan, error) {
	plan := ChangePlan{}
	if bundleEmpty(b) {
		return plan, nil
	}

	permPlan, err := diffPermissions(ctx, db, b.Permissions)
	if err != nil {
		return plan, err
	}
	plan.Permissions = permPlan

	rtPlan, err := diffResourceTypes(ctx, db, b.ResourceTypes)
	if err != nil {
		return plan, err
	}
	plan.ResourceTypes = rtPlan

	relPlan, err := diffRelationshipTypes(ctx, db, b.RelationshipTypes)
	if err != nil {
		return plan, err
	}
	plan.RelationshipTypes = relPlan

	rolePlan, err := diffRoles(ctx, db, b.Roles)
	if err != nil {
		return plan, err
	}
	plan.Roles = rolePlan

	if out != nil {
		_, _ = fmt.Fprintf(out, "dry-run: manifest hash %s\n", Hash(b))
		_, _ = fmt.Fprintf(out, "permissions: +%d ~%d =%d\n", permPlan.Insert, permPlan.Update, permPlan.Unchanged)
		_, _ = fmt.Fprintf(out, "resource_types: +%d ~%d =%d\n", rtPlan.Insert, rtPlan.Update, rtPlan.Unchanged)
		_, _ = fmt.Fprintf(out, "relationship_types: +%d ~%d =%d\n", relPlan.Insert, relPlan.Update, relPlan.Unchanged)
		_, _ = fmt.Fprintf(out, "roles: +%d ~%d =%d (grant adds %d, prunes %d)\n",
			rolePlan.Insert, rolePlan.Update, rolePlan.Unchanged, rolePlan.GrantAdds, rolePlan.GrantPrunes)
	}
	return plan, nil
}

func diffPermissions(ctx context.Context, db database.DBTX, seeds []PermissionSeed) (TablePlan, error) {
	plan := TablePlan{}
	if len(seeds) == 0 {
		return plan, nil
	}
	keys := make([]string, len(seeds))
	for i, p := range seeds {
		keys[i] = p.Key
	}
	existing := make(map[string]PermissionSeed)
	rows, err := db.Query(ctx,
		`SELECT key, description, sensitive, step_up FROM permissions WHERE key = ANY($1)`, keys)
	if err != nil {
		return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run query permissions")
	}
	defer rows.Close()
	for rows.Next() {
		var p PermissionSeed
		if err := rows.Scan(&p.Key, &p.Description, &p.Sensitive, &p.StepUp); err != nil {
			return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run scan permission")
		}
		existing[p.Key] = p
	}
	if err := rows.Err(); err != nil {
		return plan, err
	}
	for _, p := range seeds {
		cur, ok := existing[p.Key]
		if !ok {
			plan.Insert++
			continue
		}
		if cur.Description != p.Description || cur.Sensitive != p.Sensitive || cur.StepUp != p.StepUp {
			plan.Update++
		} else {
			plan.Unchanged++
		}
	}
	return plan, nil
}

func diffResourceTypes(ctx context.Context, db database.DBTX, seeds []ResourceTypeSeed) (TablePlan, error) {
	plan := TablePlan{}
	if len(seeds) == 0 {
		return plan, nil
	}
	keys := make([]string, len(seeds))
	for i, rt := range seeds {
		keys[i] = rt.Key
	}
	existing := make(map[string]ResourceTypeSeed)
	rows, err := db.Query(ctx,
		`SELECT key, description FROM resource_types WHERE key = ANY($1)`, keys)
	if err != nil {
		return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run query resource_types")
	}
	defer rows.Close()
	for rows.Next() {
		var rt ResourceTypeSeed
		if err := rows.Scan(&rt.Key, &rt.Description); err != nil {
			return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run scan resource_type")
		}
		existing[rt.Key] = rt
	}
	if err := rows.Err(); err != nil {
		return plan, err
	}
	for _, rt := range seeds {
		cur, ok := existing[rt.Key]
		if !ok {
			plan.Insert++
		} else if cur.Description != rt.Description {
			plan.Update++
		} else {
			plan.Unchanged++
		}
	}
	return plan, nil
}

func diffRelationshipTypes(ctx context.Context, db database.DBTX, seeds []RelationshipTypeSeed) (TablePlan, error) {
	plan := TablePlan{}
	if len(seeds) == 0 {
		return plan, nil
	}
	keys := make([]string, len(seeds))
	for i, rt := range seeds {
		keys[i] = rt.Key
	}
	existing := make(map[string]RelationshipTypeSeed)
	rows, err := db.Query(ctx,
		`SELECT key, subject_kind, object_kind, cardinality, description FROM relationship_types WHERE key = ANY($1)`, keys)
	if err != nil {
		return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run query relationship_types")
	}
	defer rows.Close()
	for rows.Next() {
		var rt RelationshipTypeSeed
		if err := rows.Scan(&rt.Key, &rt.SubjectKind, &rt.ObjectKind, &rt.Cardinality, &rt.Description); err != nil {
			return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run scan relationship_type")
		}
		existing[rt.Key] = rt
	}
	if err := rows.Err(); err != nil {
		return plan, err
	}
	for _, rt := range seeds {
		cur, ok := existing[rt.Key]
		if !ok {
			plan.Insert++
			continue
		}
		card := rt.Cardinality
		if card == "" {
			card = "many"
		}
		if cur.SubjectKind != rt.SubjectKind || cur.ObjectKind != rt.ObjectKind ||
			cur.Cardinality != card || cur.Description != rt.Description {
			plan.Update++
		} else {
			plan.Unchanged++
		}
	}
	return plan, nil
}

func diffRoles(ctx context.Context, db database.DBTX, seeds []RoleSeed) (RolePlan, error) {
	plan := RolePlan{}
	if len(seeds) == 0 {
		return plan, nil
	}
	keys := make([]string, len(seeds))
	for i, r := range seeds {
		keys[i] = r.Key
	}
	type roleRow struct {
		key  string
		name string
	}
	existing := make(map[string]roleRow)
	rows, err := db.Query(ctx,
		`SELECT key, name FROM roles WHERE key = ANY($1) AND tenant_id IS NULL`, keys)
	if err != nil {
		return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run query roles")
	}
	defer rows.Close()
	for rows.Next() {
		var r roleRow
		if err := rows.Scan(&r.key, &r.name); err != nil {
			return plan, kerr.Wrapf(err, "seeds.Apply", "dry-run scan role")
		}
		existing[r.key] = r
	}
	if err := rows.Err(); err != nil {
		return plan, err
	}

	grants, err := currentGrants(ctx, db, keys)
	if err != nil {
		return plan, err
	}

	for _, r := range seeds {
		cur, ok := existing[r.Key]
		if !ok {
			plan.Insert++
			plan.GrantAdds += len(r.Permissions)
			continue
		}
		if cur.name != r.Name {
			plan.Update++
		} else {
			plan.Unchanged++
		}
		want := sortedCopy(r.Permissions)
		have := sortedCopy(grants[r.Key])
		plan.GrantAdds += len(setDifference(want, have))
		plan.GrantPrunes += len(setDifference(have, want))
	}
	return plan, nil
}

func currentGrants(ctx context.Context, db database.DBTX, roleKeys []string) (map[string][]string, error) {
	out := make(map[string][]string)
	rows, err := db.Query(ctx,
		`SELECT r.key, rp.permission_key
		   FROM roles r
		   JOIN role_permissions rp ON rp.role_id = r.id
		  WHERE r.key = ANY($1) AND r.tenant_id IS NULL`, roleKeys)
	if err != nil {
		return nil, kerr.Wrapf(err, "seeds.Apply", "dry-run query role_permissions")
	}
	defer rows.Close()
	for rows.Next() {
		var roleKey, permKey string
		if err := rows.Scan(&roleKey, &permKey); err != nil {
			return nil, kerr.Wrapf(err, "seeds.Apply", "dry-run scan role_permission")
		}
		out[roleKey] = append(out[roleKey], permKey)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for k := range out {
		sort.Strings(out[k])
	}
	return out, nil
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func setDifference(a, b []string) []string {
	out := []string{}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			out = append(out, a[i])
			i++
		case a[i] > b[j]:
			j++
		default:
			i++
			j++
		}
	}
	out = append(out, a[i:]...)
	return out
}
