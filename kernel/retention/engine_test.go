package retention_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/retention"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// ensurePeople creates a product-owned scratch table the record-class callbacks
// operate on, granted to app_rt (the runtime role the callbacks run as).
func ensurePeople(h *testkit.DBHandle) {
	ctx := context.Background()
	_, _ = h.Admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS rt_people (
		id uuid, tenant_id uuid, subject text, payload text, retention_until timestamptz)`)
	_, _ = h.Admin.Exec(ctx, `GRANT SELECT, INSERT, DELETE ON rt_people TO app_rt`)
}

func peopleClass() retention.RecordClass {
	return retention.RecordClass{
		Key:       "people",
		Retention: time.Hour,
		Dispose: func(ctx context.Context, db database.TenantDB, before time.Time) (int, error) {
			tag, err := db.Exec(ctx,
				`DELETE FROM rt_people WHERE tenant_id = app_tenant_id()
				   AND retention_until IS NOT NULL AND retention_until <= $1`, before)
			if err != nil {
				return 0, err
			}
			return int(tag.RowsAffected()), nil
		},
		Export: func(ctx context.Context, db database.TenantDB, subject string) (map[string]any, error) {
			rows, err := db.Query(ctx, `SELECT payload FROM rt_people WHERE tenant_id = app_tenant_id() AND subject = $1 ORDER BY payload`, subject)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var payloads []string
			for rows.Next() {
				var p string
				if err := rows.Scan(&p); err != nil {
					return nil, err
				}
				payloads = append(payloads, p)
			}
			return map[string]any{"records": payloads}, rows.Err()
		},
		Erase: func(ctx context.Context, db database.TenantDB, subject string) (int, error) {
			tag, err := db.Exec(ctx, `DELETE FROM rt_people WHERE tenant_id = app_tenant_id() AND subject = $1`, subject)
			if err != nil {
				return 0, err
			}
			return int(tag.RowsAffected()), nil
		},
	}
}

func seedPerson(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, subject, payload string, retentionUntil *time.Time) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO rt_people (id, tenant_id, subject, payload, retention_until) VALUES ($1,$2,$3,$4,$5)`,
		uuid.New(), tenant, subject, payload, retentionUntil); err != nil {
		t.Fatal(err)
	}
}

func newEngine(t *testing.T) (*retention.Engine, *retention.DSR) {
	reg := retention.NewRegistry()
	reg.Register(peopleClass())
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	dsr := retention.NewDSR(model.UUIDv7())
	artifacts := retention.NewFileArtifactWriter(t.TempDir(), retention.TestKey(), nil)
	return retention.NewEngineWithCompliance(reg, dsr, nil, artifacts, nil), dsr
}

func peopleCount(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID) int {
	t.Helper()
	var n int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM rt_people WHERE tenant_id = $1`, tenant).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestIntegrationEngineExport(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)
	eng, dsr := newEngine(t)
	tenant := uuid.New()
	ctx := tctx(tenant)
	seedPerson(t, h, tenant, "alice", "a1", nil)
	seedPerson(t, h, tenant, "alice", "a2", nil)

	var manifest *retention.ArtifactManifest
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		reqID, e := dsr.Open(ctx, db, "alice", retention.KindExport)
		if e != nil {
			return e
		}
		manifest, e = eng.RunExportDetailed(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("run export: %v", err)
	}
	peopleResult, ok := manifest.PerClassResults["people"]
	if !ok || peopleResult.Status != retention.ClassStatusExported {
		t.Fatalf("people status = %q, want exported", peopleResult.Status)
	}
	recs, _ := peopleResult.Data["records"].([]string)
	if len(recs) != 2 {
		t.Fatalf("export people.records = %v, want 2 entries", peopleResult.Data)
	}
}

func TestIntegrationEngineErasure(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)
	eng, dsr := newEngine(t)
	tenant := uuid.New()
	ctx := tctx(tenant)
	seedPerson(t, h, tenant, "bob", "b1", nil)
	seedPerson(t, h, tenant, "bob", "b2", nil)
	seedPerson(t, h, tenant, "carol", "c1", nil)

	var result *retention.ErasureResult
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		reqID, e := dsr.Open(ctx, db, "bob", retention.KindErasure)
		if e != nil {
			return e
		}
		result, e = eng.RunErasureDetailed(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("run erasure: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("erased = %d, want 2", result.Total)
	}
	if result.Statuses["people"] != retention.ClassStatusErased {
		t.Fatalf("people erasure status = %q, want erased", result.Statuses["people"])
	}
	if got := peopleCount(t, h, tenant); got != 1 {
		t.Fatalf("remaining people = %d, want 1 (only carol)", got)
	}
}

func TestIntegrationEngineDisposition(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)
	eng, _ := newEngine(t)
	tenant := uuid.New()
	ctx := tctx(tenant)
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	seedPerson(t, h, tenant, "dan", "expired", &past)
	seedPerson(t, h, tenant, "dan", "live", &future)
	seedPerson(t, h, tenant, "dan", "kept", nil) // no retention → never disposed

	var disposed int
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		disposed, e = eng.SweepDisposition(ctx, db, time.Now())
		return e
	}); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if disposed != 1 {
		t.Fatalf("disposed = %d, want 1 (only the expired row)", disposed)
	}
	if got := peopleCount(t, h, tenant); got != 2 {
		t.Fatalf("remaining = %d, want 2 (live + kept)", got)
	}
}

func TestEngineDuplicateClassRejected(t *testing.T) {
	reg := retention.NewRegistry()
	reg.Register(retention.RecordClass{Key: "x"})
	reg.Register(retention.RecordClass{Key: "x"})
	if reg.Err() == nil {
		t.Fatal("duplicate record class must be rejected")
	}
}
