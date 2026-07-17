package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// Fourth closure-audit regression (2026-07-17): the boot-materialized
// migration snapshot must work end to end through the REAL migration engine
// (goose over PostgreSQL) — apply, idempotent rerun, and reset — while the
// module-owned source filesystem is mutated underneath it. A snapshot that
// only satisfies fs.ReadFile in unit tests proves nothing about the migrate
// path's fs.FS expectations.
func TestIntegrationMaterializedSnapshotMigratesRerunsAndResets(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	migFS := fstest.MapFS{"00001_probe.sql": &fstest.MapFile{Data: []byte(
		"-- +goose Up\nCREATE TABLE widgets_snapshot_probe (id int primary key);\n-- +goose Down\nDROP TABLE widgets_snapshot_probe;\n")}}
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Migrations(migFS)
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	// Corrupt the module-owned source AFTER boot: the snapshot must be
	// unaffected through every engine operation below.
	migFS["00001_probe.sql"].Data = []byte("-- +goose Up\nTHIS IS NOT SQL;\n-- +goose Down\nNOR THIS;\n")
	migFS["00002_evil.sql"] = &fstest.MapFile{Data: []byte("-- +goose Up\nDROP TABLE tenants;\n-- +goose Down\nSELECT 1;\n")}

	snap := booted.RuntimeMigrations()["widgets"]
	ctx := context.Background()

	res, err := database.Migrate(ctx, h.Admin, snap, "widgets")
	if err != nil {
		t.Fatalf("Migrate through the materialized snapshot: %v", err)
	}
	if res.Applied != 1 {
		t.Fatalf("applied %d migrations, want exactly the 1 boot-validated file (the post-boot addition must be invisible)", res.Applied)
	}
	var exists bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'widgets_snapshot_probe')`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("migration did not create the probe table")
	}

	// Idempotent rerun: nothing new to apply, no error from re-reading the
	// snapshot.
	rerun, err := database.Migrate(ctx, h.Admin, snap, "widgets")
	if err != nil {
		t.Fatalf("idempotent rerun: %v", err)
	}
	if rerun.Applied != 0 {
		t.Fatalf("rerun applied %d migrations, want 0", rerun.Applied)
	}

	// Reset: down migrations run from the same snapshot bytes.
	if _, err := database.MigrateReset(ctx, h.Admin, snap, "widgets"); err != nil {
		t.Fatalf("MigrateReset through the materialized snapshot: %v", err)
	}
	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'widgets_snapshot_probe')`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("reset did not drop the probe table")
	}
}
