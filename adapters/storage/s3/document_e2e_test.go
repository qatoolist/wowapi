// End-to-end proof that the s3 adapter satisfies the document framework's
// storage.Adapter port under real usage: a minimal TEST-LOCAL module
// (testDocModule, below) that registers a single document class, booted with
// THIS s3 adapter (real minio), walks the full document upload round trip —
// create → presigned session → the client's real HTTP PUT to minio →
// checksum-verified confirm — and confirm rejects a lying checksum. It swaps
// storage.NewMemory (and its synthetic store.Put) for the s3 adapter and a
// real presigned upload.
//
// This exercises the framework documents capability against a real adapter
// through exported API only — no product module is required.
//
// Gates: testkit.NewDB handles the Postgres gate (skip, or fail under
// WOWAPI_REQUIRE_DB=1) and requireMinio handles the S3/minio gate the same way
// (skip, or fail under WOWAPI_REQUIRE_S3=1).
package s3_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// docClassAttachment is the test-local document class key, namespaced under
// the test module's own name so it cannot collide with a product module.
const docClassAttachment = "testdoc.attachment"

// testDocModule is the smallest possible module.Module: it registers one
// document class (no routes, no migrations, no seeds — the documents
// capability needs none of those to exercise the s3 adapter end-to-end).
type testDocModule struct{}

var _ module.Module = (*testDocModule)(nil)

func (testDocModule) Name() string        { return "testdoc" }
func (testDocModule) DependsOn() []string { return nil }

func (testDocModule) Register(mc module.Context) error {
	mc.DocumentClasses().Register("testdoc", document.Class{
		Key:                docClassAttachment,
		DefaultSensitivity: document.SensitivityInternal,
		MaxBytes:           1 << 20,
	})
	return nil
}

// docEnv is one booted testDocModule over the s3 adapter.
type docEnv struct {
	h  *testkit.DBHandle
	k  *kernel.Kernel
	tn uuid.UUID
}

// bootDocModuleWithS3 boots testDocModule with the s3 adapter wired as
// Deps.Storage, so the document class it registers has real object storage
// behind it.
func bootDocModuleWithS3(t *testing.T, store storage.Adapter) *docEnv {
	t.Helper()
	h := testkit.NewDB(t)
	ctx := context.Background()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM, Storage: store,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(&testDocModule{})
	booted, err := a.Boot(ctx, k, config.Namespaces{})
	if err != nil {
		t.Fatalf("app.Boot: %v", err)
	}
	// testDocModule declares no migrations/seeds of its own — the tenants /
	// acting_capacities tables used below come from the kernel baseline
	// migrations testkit.NewDB already applied.
	if err := seeds.Sync(ctx, h.Platform, booted.RuntimeSeeds()); err != nil {
		t.Fatalf("seed sync: %v", err)
	}

	tn := uuid.New()
	if _, err := h.Admin.Exec(ctx,
		`INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1,$2,$3,$4)`,
		tn, "t-"+uuid.NewString()[:8], "Tenant", uuid.Nil); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	return &docEnv{h: h, k: k, tn: tn}
}

// newCapacity creates a user + active acting capacity and returns the capacity
// id the document service records as the acting principal.
func newCapacity(t *testing.T, e *docEnv) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	if _, err := e.h.Admin.Exec(ctx,
		`INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		userID, "idp-"+uuid.NewString()[:8], uuid.NewString()[:8]+"@example.test", uuid.Nil); err != nil {
		t.Fatalf("create user: %v", err)
	}
	capID := uuid.New()
	if _, err := e.h.Admin.Exec(ctx,
		`INSERT INTO acting_capacities (id, tenant_id, user_id, label, created_by)
		 VALUES ($1,$2,$3,'member',$4)`, capID, e.tn, userID, uuid.Nil); err != nil {
		t.Fatalf("create capacity: %v", err)
	}
	return capID
}

// presignedPut is the CLIENT leg: a real HTTP PUT of body to the session's
// presigned URL — bytes travel to minio directly, never through the kernel.
func presignedPut(t *testing.T, sess document.UploadSession, body []byte) {
	t.Helper()
	req, err := http.NewRequest(sess.Upload.Method, sess.Upload.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	for name, value := range sess.Upload.Headers {
		req.Header.Set(name, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT to presigned URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("presigned PUT = %d\n%s", resp.StatusCode, b)
	}
}

func TestDocument_UploadRoundTrip_S3(t *testing.T) {
	store := requireMinio(t, testConfig()) // minio gate first: no DB churn when storage is absent
	e := bootDocModuleWithS3(t, store)
	if e.k.Documents == nil {
		t.Fatal("kernel.Documents is nil despite a wired s3 storage adapter")
	}

	ctx := database.WithActorID(testkit.TenantCtx(e.tn), newCapacity(t, e))
	body := []byte("test attachment payload over real minio")
	sum := sha256.Sum256(body)

	var verID uuid.UUID
	err := e.h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, err := e.k.Documents.Create(ctx, db, document.CreateInput{
			Class: docClassAttachment, Title: "roundtrip-s3",
		})
		if err != nil {
			return err
		}
		sess, err := e.k.Documents.InitiateUploadChecksum(ctx, db, docID, hex.EncodeToString(sum[:]))
		if err != nil {
			return err
		}
		if sess.Upload.URL == "" || sess.StorageKey == "" {
			t.Fatalf("presigned session incomplete: %+v", sess)
		}
		if sess.Upload.ExpiresAt.Before(time.Now()) {
			t.Fatalf("presigned URL already expired: %+v", sess.Upload)
		}
		presignedPut(t, sess, body) // the client's REAL PUT to minio
		verID, err = e.k.Documents.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: hex.EncodeToString(sum[:]),
			DeclaredMIME: "text/plain; charset=utf-8",
		})
		return err
	})
	if err != nil {
		t.Fatalf("upload round trip: %v", err)
	}
	if verID == uuid.Nil {
		t.Fatal("confirm returned a nil version id")
	}

	// Negative: a lying checksum must be refused — the confirm verifies the
	// bytes minio actually holds via the adapter's Stat.
	err = e.h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, err := e.k.Documents.Create(ctx, db, document.CreateInput{
			Class: docClassAttachment, Title: "tampered-s3",
		})
		if err != nil {
			return err
		}
		sess, err := e.k.Documents.InitiateUploadChecksum(ctx, db, docID, hex.EncodeToString(sum[:]))
		if err != nil {
			return err
		}
		presignedPut(t, sess, body)
		_, err = e.k.Documents.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize:     int64(len(body)),
			DeclaredChecksum: hex.EncodeToString(make([]byte, sha256.Size)), // wrong
			DeclaredMIME:     "text/plain; charset=utf-8",
		})
		return err
	})
	if err == nil {
		t.Fatal("confirm accepted a wrong declared checksum")
	}
}
