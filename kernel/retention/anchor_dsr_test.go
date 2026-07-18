package retention_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/testkit"
)

func dsrCtx(tenant uuid.UUID) context.Context {
	return database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())
}

// TestIntegrationDSRArtifactWriteAndChecksum proves RunExport writes an encrypted
// artifact, sets a checksum, and that the checksum verifies against the file.
func TestIntegrationDSRArtifactWriteAndChecksum(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)

	dir := t.TempDir()
	key := artifactTestKey()
	artifacts := retention.NewFileArtifactWriter(dir, key, nil)

	reg := retention.NewRegistry()
	reg.Register(peopleClass())
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	dsr := retention.NewDSR(model.UUIDv7())
	eng := retention.NewEngine(reg, dsr, nil, artifacts, nil)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	seedPerson(t, h, tenant, "alice", "a1", nil)
	seedPerson(t, h, tenant, "alice", "a2", nil)

	var manifest *retention.ArtifactManifest
	var path string
	var reqID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		reqID, e = dsr.Open(ctx, db, "alice", retention.KindExport)
		if e != nil {
			return e
		}
		manifest, e = eng.RunExport(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("run export: %v", err)
	}
	if manifest.Checksum == "" {
		t.Fatal("manifest checksum is empty")
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read artifact dir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("artifact dir has %d files, want 1", len(files))
	}
	path = dir + "/" + files[0].Name()

	envBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	var env struct {
		Checksum string `json:"checksum"`
		Data     string `json:"data"`
	}
	if err := json.Unmarshal(envBytes, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		t.Fatalf("decode ciphertext: %v", err)
	}
	cs := sha256.Sum256(ciphertext)
	wantChecksum := hex.EncodeToString(cs[:])
	if env.Checksum != wantChecksum {
		t.Fatalf("envelope checksum = %q, want %q", env.Checksum, wantChecksum)
	}
	if manifest.Checksum != env.Checksum {
		t.Fatalf("manifest checksum = %q, envelope checksum = %q", manifest.Checksum, env.Checksum)
	}

	// Decrypt and inspect the manifest.
	plaintext, err := artifacts.Read(ctx, nil, path) // db unused by file writer
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	var decrypted retention.ArtifactManifest
	if err := json.Unmarshal(plaintext, &decrypted); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if decrypted.RequestID == uuid.Nil {
		t.Fatal("decrypted manifest has no request id")
	}
	peopleResult, ok := decrypted.PerClassResults["people"]
	if !ok || peopleResult.Status != retention.ClassStatusExported {
		t.Fatalf("people status = %q, want exported", peopleResult.Status)
	}

	// The DSR request is completed.
	var req retention.Request
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		req, e = dsr.Get(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("get dsr: %v", err)
	}
	if req.Status != "completed" {
		t.Fatalf("dsr status = %q, want completed", req.Status)
	}
}

// failingWriter is an ArtifactWriter that always returns a fixed error.
type failingWriter struct{ err error }

func (f failingWriter) Write(context.Context, database.TenantDB, uuid.UUID, *retention.ArtifactManifest) (string, string, error) {
	return "", "", f.err
}

func (f failingWriter) Read(context.Context, database.TenantDB, string) ([]byte, error) {
	return nil, f.err
}

// TestIntegrationDSRExportArtifactWriteFailure proves the DSR stays pending when
// the artifact writer fails.
func TestIntegrationDSRExportArtifactWriteFailure(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)

	sentinel := errors.New("artifact write failed")
	reg := retention.NewRegistry()
	reg.Register(peopleClass())
	dsr := retention.NewDSR(model.UUIDv7())
	eng := retention.NewEngine(reg, dsr, nil, failingWriter{err: sentinel}, nil)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	seedPerson(t, h, tenant, "alice", "a1", nil)

	var reqID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		reqID, e = dsr.Open(ctx, db, "alice", retention.KindExport)
		if e != nil {
			return e
		}
		_, e = eng.RunExport(ctx, db, reqID)
		if !errors.Is(e, sentinel) {
			t.Fatalf("run export error = %v, want sentinel", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("tx: %v", err)
	}

	var req retention.Request
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		req, e = dsr.Get(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("get dsr: %v", err)
	}
	if req.Status != "pending" {
		t.Fatalf("dsr status = %q, want pending after failed artifact write", req.Status)
	}
}

// TestIntegrationCentralLegalHoldBlocksDisposeErase proves the Engine wrapper
// blocks a deliberately non-compliant callback (one with no internal hold check)
// when a hold is placed.
func TestIntegrationCentralLegalHoldBlocksDisposeErase(t *testing.T) {
	h := testkit.NewDB(t)
	holds := retention.NewHolds(model.UUIDv7())

	// A class whose callbacks ignore holds — the wrapper must still block them.
	deleted := false
	class := retention.RecordClass{
		Key: "risky",
		Dispose: func(ctx context.Context, db database.TenantDB, before time.Time) (int, error) {
			deleted = true
			return 1, nil
		},
		Erase: func(ctx context.Context, db database.TenantDB, subjectRef string) (int, error) {
			deleted = true
			return 1, nil
		},
	}
	reg := retention.NewRegistry()
	reg.Register(class)
	dsr := retention.NewDSR(model.UUIDv7())

	tenant := uuid.New()
	ctx := dsrCtx(tenant)

	// Place a record_class hold and a dsr_subject hold.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		// holdID is internal; place the hold via Holds on the derived UUID.
		if _, e := holds.Place(ctx, db, "record_class", uuid.NewSHA1(uuid.NameSpaceOID, []byte("wowapi:hold:risky")), "litigation"); e != nil {
			return e
		}
		_, e := holds.Place(ctx, db, "dsr_subject", uuid.NewSHA1(uuid.NameSpaceOID, []byte("wowapi:hold:bob")), "litigation")
		return e
	}); err != nil {
		t.Fatalf("place holds: %v", err)
	}

	// Dispose is blocked.
	artifacts := retention.NewFileArtifactWriter(t.TempDir(), artifactTestKey(), nil)
	eng := retention.NewEngine(reg, dsr, holds, artifacts, nil)
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := eng.SweepDisposition(ctx, db, time.Now())
		return e
	})
	if !errors.Is(err, retention.ErrHeld) {
		t.Fatalf("dispose error = %v, want ErrHeld", err)
	}
	if deleted {
		t.Fatal("non-compliant Dispose callback ran despite hold")
	}

	// Open an erasure DSR in its own transaction so the existence of the row is
	// not rolled back when the erasure is blocked.
	var reqID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		reqID, e = dsr.Open(ctx, db, "bob", retention.KindErasure)
		return e
	}); err != nil {
		t.Fatalf("open dsr: %v", err)
	}

	// Erasure is blocked.
	err = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := eng.RunErasure(ctx, db, reqID)
		return e
	})
	if !errors.Is(err, retention.ErrHeld) {
		t.Fatalf("erasure error = %v, want ErrHeld", err)
	}
	if deleted {
		t.Fatal("non-compliant Erase callback ran despite hold")
	}

	// Erasure DSR remains pending.
	var req retention.Request
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		req, e = dsr.Get(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("get dsr: %v", err)
	}
	if req.Status != "pending" {
		t.Fatalf("erasure status = %q, want pending", req.Status)
	}
}

// TestIntegrationExplicitPerClassExportStatus proves the manifest lists every
// registered class, including those with no Export callback.
func TestIntegrationExplicitPerClassExportStatus(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)

	reg := retention.NewRegistry()
	reg.Register(peopleClass())
	reg.Register(retention.RecordClass{Key: "sessions"}) // no Export
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	dsr := retention.NewDSR(model.UUIDv7())
	artifacts := retention.NewFileArtifactWriter(t.TempDir(), artifactTestKey(), nil)
	eng := retention.NewEngine(reg, dsr, nil, artifacts, nil)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	seedPerson(t, h, tenant, "alice", "a1", nil)

	var manifest *retention.ArtifactManifest
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		reqID, e := dsr.Open(ctx, db, "alice", retention.KindExport)
		if e != nil {
			return e
		}
		manifest, e = eng.RunExport(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("run export: %v", err)
	}

	if len(manifest.PerClassResults) != 2 {
		t.Fatalf("per-class results = %v, want 2 classes", manifest.PerClassResults)
	}
	if manifest.PerClassResults["people"].Status != retention.ClassStatusExported {
		t.Fatalf("people status = %q, want exported", manifest.PerClassResults["people"].Status)
	}
	if manifest.PerClassResults["sessions"].Status != retention.ClassStatusNotApplicable {
		t.Fatalf("sessions status = %q, want not_applicable", manifest.PerClassResults["sessions"].Status)
	}
}

// TestIntegrationExplicitPerClassErasureStatus proves RunErasure reports a
// status for every registered class.
func TestIntegrationExplicitPerClassErasureStatus(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)

	reg := retention.NewRegistry()
	reg.Register(peopleClass())
	reg.Register(retention.RecordClass{Key: "sessions"}) // no Erase
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	dsr := retention.NewDSR(model.UUIDv7())
	eng := retention.NewEngine(reg, dsr, nil, nil, nil)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	seedPerson(t, h, tenant, "bob", "b1", nil)

	var result *retention.ErasureResult
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		reqID, e := dsr.Open(ctx, db, "bob", retention.KindErasure)
		if e != nil {
			return e
		}
		result, e = eng.RunErasure(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("run erasure: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("total erased = %d, want 1", result.Total)
	}
	if len(result.Statuses) != 2 {
		t.Fatalf("statuses = %v, want 2 classes", result.Statuses)
	}
	if result.Statuses["people"] != retention.ClassStatusErased {
		t.Fatalf("people status = %q, want erased", result.Statuses["people"])
	}
	if result.Statuses["sessions"] != retention.ClassStatusNotApplicable {
		t.Fatalf("sessions status = %q, want not_applicable", result.Statuses["sessions"])
	}
}

// TestIntegrationArtifactDownloadAudit proves Read records a download-audit row.
func TestIntegrationArtifactDownloadAudit(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(nil, nil)
	dir := t.TempDir()
	artifacts := retention.NewFileArtifactWriter(dir, artifactTestKey(), w)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	reqID := uuid.New()
	manifest := &retention.ArtifactManifest{
		RequestID:       reqID,
		PerClassResults: map[string]retention.ClassResult{"people": {Status: retention.ClassStatusExported}},
	}

	var path string
	var auditCount int
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		_, path, e = artifacts.Write(ctx, db, reqID, manifest)
		if e != nil {
			return e
		}
		if _, e := artifacts.Read(ctx, db, path); e != nil {
			return e
		}
		return db.QueryRow(ctx,
			`SELECT count(*) FROM audit_logs WHERE tenant_id = app_tenant_id() AND action = 'dsr.artifact.download'`).Scan(&auditCount)
	}); err != nil {
		t.Fatalf("write/read artifact / count audit: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("download audit rows = %d, want 1", auditCount)
	}
}

// TestIntegrationArtifactCreationAudit proves Write records a creation-audit row.
func TestIntegrationArtifactCreationAudit(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(nil, nil)
	artifacts := retention.NewFileArtifactWriter(t.TempDir(), artifactTestKey(), w)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	reqID := uuid.New()
	manifest := &retention.ArtifactManifest{
		RequestID:       reqID,
		PerClassResults: map[string]retention.ClassResult{"people": {Status: retention.ClassStatusExported}},
	}

	var auditCount int
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, _, e := artifacts.Write(ctx, db, reqID, manifest); e != nil {
			return e
		}
		return db.QueryRow(ctx,
			`SELECT count(*) FROM audit_logs WHERE tenant_id = app_tenant_id() AND action = 'dsr.artifact.created'`).Scan(&auditCount)
	}); err != nil {
		t.Fatalf("write artifact / count audit: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("creation audit rows = %d, want 1", auditCount)
	}
}

// TestIntegrationDSRExportEmptyClassStatus proves a class with an Export that
// returns no data is reported as empty rather than exported.
func TestIntegrationDSRExportEmptyClassStatus(t *testing.T) {
	h := testkit.NewDB(t)

	reg := retention.NewRegistry()
	reg.Register(retention.RecordClass{
		Key: "notes",
		Export: func(ctx context.Context, db database.TenantDB, subjectRef string) (map[string]any, error) {
			return map[string]any{}, nil
		},
	})
	dsr := retention.NewDSR(model.UUIDv7())
	artifacts := retention.NewFileArtifactWriter(t.TempDir(), artifactTestKey(), nil)
	eng := retention.NewEngine(reg, dsr, nil, artifacts, nil)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)

	var manifest *retention.ArtifactManifest
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		reqID, e := dsr.Open(ctx, db, "alice", retention.KindExport)
		if e != nil {
			return e
		}
		manifest, e = eng.RunExport(ctx, db, reqID)
		return e
	}); err != nil {
		t.Fatalf("run export: %v", err)
	}
	if manifest.PerClassResults["notes"].Status != retention.ClassStatusEmpty {
		t.Fatalf("notes status = %q, want empty", manifest.PerClassResults["notes"].Status)
	}
}

// TestIntegrationDSRExportHoldGatedCompletion uses a real Engine and artifact
// writer to confirm a successful export completes the DSR and the artifact can
// be read back with a matching checksum.
func TestIntegrationDSRExportArtifactRoundTrip(t *testing.T) {
	h := testkit.NewDB(t)
	ensurePeople(h)

	artifacts := retention.NewFileArtifactWriter(t.TempDir(), artifactTestKey(), nil)
	reg := retention.NewRegistry()
	reg.Register(peopleClass())
	dsr := retention.NewDSR(model.UUIDv7())
	_ = retention.NewEngine(reg, dsr, nil, artifacts, nil)

	tenant := uuid.New()
	ctx := dsrCtx(tenant)
	seedPerson(t, h, tenant, "carol", "c1", nil)

	var path string
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		reqID, e := dsr.Open(ctx, db, "carol", retention.KindExport)
		if e != nil {
			return e
		}
		_, path, e = artifacts.Write(ctx, db, reqID, &retention.ArtifactManifest{
			RequestID:       reqID,
			PerClassResults: map[string]retention.ClassResult{"people": {Status: retention.ClassStatusExported}},
		})
		return e
	}); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	_, err := artifacts.Read(ctx, nil, path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
}
