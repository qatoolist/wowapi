package document_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/document"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/testkit"
)

// coverage_test.go — behavior tests for the error/edge branches of the document
// service, class registry, and authz gate that the round-trip tests do not reach.

// --- fakeEvaluator: a configurable authz.Evaluator so the authorize() gate's
// evaluator branches (allow / policy-deny / hard-error) can be exercised. ---

type fakeEvaluator struct {
	decision authz.Decision
	err      error
}

func (f fakeEvaluator) Evaluate(context.Context, database.TenantDB, authz.Actor, string, authz.Target) (authz.Decision, error) {
	return f.decision, f.err
}

func (f fakeEvaluator) Filter(context.Context, database.TenantDB, authz.Actor, string, string) (authz.ListFilter, error) {
	return authz.ListFilter{}, nil
}

// --- faultyStore: a storage.Adapter that injects failures at the object-store
// boundary so the service's storage error-wrap branches (presign/stat/peek/delete
// faults) are exercised. It embeds storage.Memory so the successful setup path
// (Put + the happy calls) is real. ---

type faultyStore struct {
	*storage.Memory
	failPut    bool
	failStat   bool
	failPeek   bool
	failGet    bool
	failDelete bool
}

func (f *faultyStore) PresignPut(ctx context.Context, key string, ttl time.Duration) (storage.PresignedURL, error) {
	if f.failPut {
		return storage.PresignedURL{}, errors.New("presign put backend down")
	}
	return f.Memory.PresignPut(ctx, key, ttl)
}

func (f *faultyStore) PresignGet(ctx context.Context, key string, ttl time.Duration) (storage.PresignedURL, error) {
	if f.failGet {
		return storage.PresignedURL{}, errors.New("presign get backend down")
	}
	return f.Memory.PresignGet(ctx, key, ttl)
}

func (f *faultyStore) Stat(ctx context.Context, key string) (storage.ObjectInfo, error) {
	if f.failStat {
		return storage.ObjectInfo{}, errors.New("stat backend down") // not a NotFound
	}
	return f.Memory.Stat(ctx, key)
}

func (f *faultyStore) Peek(ctx context.Context, key string, n int) ([]byte, error) {
	if f.failPeek {
		return nil, errors.New("peek backend down")
	}
	return f.Memory.Peek(ctx, key, n)
}

func (f *faultyStore) Delete(ctx context.Context, key string) error {
	if f.failDelete {
		return errors.New("delete backend down")
	}
	return f.Memory.Delete(ctx, key)
}

// newFaultyHarness builds a harness whose service talks to a faultyStore. The
// returned harness.store shares the faultyStore's underlying memory, so the
// harness helpers (uploadVersion/download, which call store.Put) still work.
func newFaultyHarness(t *testing.T, class document.Class) (*harness, *faultyStore) {
	t.Helper()
	h := testkit.NewDB(t)
	reg := document.NewRegistry()
	reg.Register("core", class)
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	mem := storage.NewMemory()
	fs := &faultyStore{Memory: mem}
	svc := document.New(reg, fs, nil, outbox.NewWriter(model.UUIDv7()), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn), actor)
	return &harness{h: h, store: mem, svc: svc, tn: tn, actor: actor, ctx: ctx}, fs
}

// newHarnessEv is newHarness with a supplied authz evaluator wired into the service.
func newHarnessEv(t *testing.T, class document.Class, ev authz.Evaluator) *harness {
	t.Helper()
	h := testkit.NewDB(t)
	reg := document.NewRegistry()
	reg.Register("core", class)
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemory()
	svc := document.New(reg, store, ev, outbox.NewWriter(model.UUIDv7()), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn), actor)
	return &harness{h: h, store: store, svc: svc, tn: tn, actor: actor, ctx: ctx}
}

// ---------------------------------------------------------------------------
// Registry (pure, no DB)
// ---------------------------------------------------------------------------

func TestRegistryRegisterRejectsBadInput(t *testing.T) {
	cases := []struct {
		name   string
		module string
		class  document.Class
	}{
		{"malformed key", "core", document.Class{Key: "NotDotted"}},
		{"module prefix mismatch", "core", document.Class{Key: "other.doc"}},
		{"invalid default sensitivity", "core", document.Class{Key: "core.doc", DefaultSensitivity: "bogus"}},
		{"negative max bytes", "core", document.Class{Key: "core.doc", MaxBytes: -1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := document.NewRegistry()
			reg.Register(tc.module, tc.class)
			if err := reg.Err(); err == nil {
				t.Fatalf("%s must be rejected", tc.name)
			}
		})
	}
}

func TestRegistryDuplicateRejected(t *testing.T) {
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "core.doc"})
	reg.Register("core", document.Class{Key: "core.doc"})
	if err := reg.Err(); err == nil {
		t.Fatal("duplicate class registration must be rejected")
	}
}

func TestRegistryErrJoinsMultiple(t *testing.T) {
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "BAD"})
	reg.Register("core", document.Class{Key: "also.bad.too.many"})
	err := reg.Err()
	if err == nil {
		t.Fatal("expected joined error")
	}
	if kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("registration error must be internal, got %v", err)
	}
}

func TestRegistryKeysSorted(t *testing.T) {
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "core.zeta"})
	reg.Register("core", document.Class{Key: "core.alpha"})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	keys := reg.Keys()
	want := []string{"core.alpha", "core.zeta"}
	if len(keys) != len(want) || keys[0] != want[0] || keys[1] != want[1] {
		t.Fatalf("Keys() = %v, want sorted %v", keys, want)
	}
	// Register applied the module-default sensitivity.
	c, ok := reg.Get("core.alpha")
	if !ok || c.DefaultSensitivity != document.SensitivityInternal || c.Module != "core" {
		t.Fatalf("registered class not normalized: %+v ok=%v", c, ok)
	}
}

// ---------------------------------------------------------------------------
// New (constructor guardrails)
// ---------------------------------------------------------------------------

func TestNewPanicsOnMissingDeps(t *testing.T) {
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "core.doc"})
	store := storage.NewMemory()
	ob := outbox.NewWriter(model.UUIDv7())
	idgen := model.UUIDv7()

	mustPanic := func(name string, fn func()) {
		defer func() {
			if recover() == nil {
				t.Fatalf("New must panic when %s is nil", name)
			}
		}()
		fn()
	}
	mustPanic("registry", func() { document.New(nil, store, nil, ob, nil, idgen) })
	mustPanic("store", func() { document.New(reg, nil, nil, ob, nil, idgen) })
	mustPanic("outbox", func() { document.New(reg, store, nil, nil, nil, idgen) })
	mustPanic("idgen", func() { document.New(reg, store, nil, ob, nil, nil) })
}

// ---------------------------------------------------------------------------
// Create — validation branches + resource anchor (nullStr/nullUUID non-nil)
// ---------------------------------------------------------------------------

func TestIntegrationCreateValidation(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, e := a.svc.Create(ctx, db, document.CreateInput{Class: "core.missing", Title: "T"}); kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("unknown class must be validation, got %v", e)
		}
		if _, e := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "  "}); kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("blank title must be validation, got %v", e)
		}
		if _, e := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T", Sensitivity: "bogus"}); kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("invalid sensitivity must be validation, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationCreateWithResourceAnchor(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	anchorID := uuid.New()
	var docID uuid.UUID
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{
			Class:    "core.doc",
			Title:    "anchored",
			Resource: resource.Ref{Type: "core.thing", ID: anchorID},
		})
		return e
	})
	if err != nil {
		t.Fatalf("create with anchor: %v", err)
	}
	var rtype string
	var rid uuid.UUID
	if err := a.h.Admin.QueryRow(context.Background(),
		`SELECT resource_type, resource_id FROM documents WHERE id=$1`, docID).Scan(&rtype, &rid); err != nil {
		t.Fatal(err)
	}
	if rtype != "core.thing" || rid != anchorID {
		t.Fatalf("resource anchor not persisted: type=%q id=%v", rtype, rid)
	}
}

// ---------------------------------------------------------------------------
// InitiateUpload — not found + not-active
// ---------------------------------------------------------------------------

func TestIntegrationInitiateUploadNotFound(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.InitiateUpload(ctx, db, uuid.New())
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("missing document must be NotFound, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationInitiateUploadOnVoidedDoc(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", Retention: time.Hour})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("gone"), "text/plain; charset=utf-8")
	// Void the document by sweeping past its retention.
	if n, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour)); err != nil || n != 1 {
		t.Fatalf("sweep: n=%d err=%v", n, err)
	}
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.InitiateUpload(ctx, db, docID)
		if kerr.KindOf(e) != kerr.KindConflict {
			t.Fatalf("upload to a voided document must conflict, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// ConfirmUpload — every rejection branch
// ---------------------------------------------------------------------------

func TestIntegrationConfirmUploadNotFound(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: uuid.New(), VersionNo: 1, StorageKey: "nope",
			DeclaredSize: 1, DeclaredChecksum: sum([]byte("x")), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("confirm on missing document must be NotFound, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationConfirmUploadMissingObject(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		// Never Put the bytes → Stat returns NotFound → upload_missing validation.
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: 5, DeclaredChecksum: sum([]byte("hello")), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("confirm with no uploaded object must be validation, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationConfirmUploadSizeMismatch(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("actual bytes")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)) + 99, DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("size mismatch must be validation, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationConfirmUploadTooLarge(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", MaxBytes: 4})
	body := []byte("way over the four byte ceiling")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("over-limit upload must be validation, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationConfirmUploadMIMENotAllowed(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", AllowedMIME: []string{"application/pdf"}})
	body := []byte("plain text, not a pdf")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("disallowed MIME must be validation, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Binary bytes sniff as application/octet-stream, so mimeConflict trusts the
// declared type (the octet-stream short-circuit) and the version is accepted.
func TestIntegrationConfirmUploadOctetStreamTrustsDeclared(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte{0x00, 0x01, 0x02, 0x03, 0xff, 0xfe, 0x10, 0x11}
	var verID uuid.UUID
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		var e error
		verID, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "application/x-custom",
		})
		return e
	})
	if err != nil || verID == uuid.Nil {
		t.Fatalf("octet-stream sniff must trust declared MIME: verID=%v err=%v", verID, err)
	}
}

func TestIntegrationConfirmUploadDuplicateVersion(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("racing versions")
	var docID uuid.UUID
	var s1, s2 document.UploadSession

	// Tx1: create + reserve two sessions. Both compute version_no 1 because no
	// version is committed yet.
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		if docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"}); e != nil {
			return e
		}
		if s1, e = a.svc.InitiateUpload(ctx, db, docID); e != nil {
			return e
		}
		s2, e = a.svc.InitiateUpload(ctx, db, docID)
		return e
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if s1.VersionNo != 1 || s2.VersionNo != 1 {
		t.Fatalf("both sessions must reserve version 1: s1=%d s2=%d", s1.VersionNo, s2.VersionNo)
	}
	a.store.Put(s1.StorageKey, body)
	a.store.Put(s2.StorageKey, body)

	// Tx2: confirm s1 as version 1 (commits).
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: s1.VersionNo, StorageKey: s1.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		return e
	}); err != nil {
		t.Fatalf("first confirm: %v", err)
	}

	// Tx3: confirm s2 as the same version 1 → unique violation → conflict. The
	// error is returned so the aborted tx rolls back cleanly.
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: s2.VersionNo, StorageKey: s2.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		return e
	})
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("second confirm of same version_no must conflict, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Download — explicit version + missing version
// ---------------------------------------------------------------------------

func TestIntegrationDownloadExplicitVersion(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("v1"), "text/plain; charset=utf-8")
	var out document.Download
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		out, e = a.svc.Download(ctx, db,
			authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: a.tn},
			document.DownloadInput{DocumentID: docID, VersionNo: 1})
		return e
	})
	if err != nil || out.VersionNo != 1 {
		t.Fatalf("explicit-version download: out=%+v err=%v", out, err)
	}
}

func TestIntegrationDownloadMissingVersion(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("only-v1"), "text/plain; charset=utf-8")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Download(ctx, db,
			authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: a.tn},
			document.DownloadInput{DocumentID: docID, VersionNo: 99})
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("nonexistent version must be NotFound, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationDownloadMissingDocument(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Download(ctx, db,
			authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: a.tn},
			document.DownloadInput{DocumentID: uuid.New()})
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("missing document download must be NotFound, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// authorize() — evaluator allow / policy-deny / hard-error branches
// ---------------------------------------------------------------------------

// An evaluator "allow" lets a stranger (no ownership, no grant) download; the
// actor here carries only a UserID (no capacity) to exercise the actorID UserID
// branch fed to the access hook.
func TestIntegrationEvaluatorAllowsStranger(t *testing.T) {
	ev := fakeEvaluator{decision: authz.Decision{Allowed: true, Reason: "role:reader"}}
	a := newHarnessEv(t, document.Class{Key: "core.doc"}, ev)
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("policy-open"), "text/plain; charset=utf-8")

	stranger := uuid.New()
	var out document.Download
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		out, e = a.svc.Download(ctx, db,
			authz.Actor{UserID: stranger, TenantID: a.tn}, // no CapacityID
			document.DownloadInput{DocumentID: docID})
		return e
	})
	if err != nil || out.VersionNo != 1 {
		t.Fatalf("evaluator allow must let a stranger download: out=%+v err=%v", out, err)
	}
}

// A policy deny is authoritative: even the owner is refused.
func TestIntegrationEvaluatorPolicyDenyBeatsOwner(t *testing.T) {
	ev := fakeEvaluator{decision: authz.Decision{Allowed: false, Reason: "policy:locked"}}
	a := newHarnessEv(t, document.Class{Key: "core.doc"}, ev)
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("locked"), "text/plain; charset=utf-8")

	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Download(ctx, db,
			authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: a.tn},
			document.DownloadInput{DocumentID: docID})
		if kerr.KindOf(e) != kerr.KindForbidden {
			t.Fatalf("policy deny must forbid even the owner, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A real (non-internal) evaluator error is propagated, not swallowed.
func TestIntegrationEvaluatorHardErrorPropagates(t *testing.T) {
	ev := fakeEvaluator{err: kerr.E(kerr.KindConflict, "boom", "evaluator exploded")}
	a := newHarnessEv(t, document.Class{Key: "core.doc"}, ev)
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("err"), "text/plain; charset=utf-8")

	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Download(ctx, db,
			authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: a.tn},
			document.DownloadInput{DocumentID: docID})
		if kerr.KindOf(e) != kerr.KindConflict {
			t.Fatalf("evaluator hard error must propagate, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A write grant satisfies a read request (accepted-access set includes write).
func TestIntegrationWriteGrantSatisfiesRead(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("rw"), "text/plain; charset=utf-8")
	stranger := uuid.New()
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Grant(ctx, db, document.GrantInput{
			DocumentID: docID, GranteeKind: "capacity", GranteeRef: stranger.String(), Access: "write",
		})
		return e
	}); err != nil {
		t.Fatalf("grant write: %v", err)
	}
	if _, err := a.download(t, stranger, docID); err != nil {
		t.Fatalf("a write grant must satisfy a read download: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Grant — validation + not-found
// ---------------------------------------------------------------------------

func TestIntegrationGrantValidation(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("g"), "text/plain; charset=utf-8")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, e := a.svc.Grant(ctx, db, document.GrantInput{DocumentID: docID, GranteeKind: "banana", GranteeRef: "x", Access: "read"}); kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("bad grantee kind must be validation, got %v", e)
		}
		if _, e := a.svc.Grant(ctx, db, document.GrantInput{DocumentID: docID, GranteeKind: "capacity", GranteeRef: "x", Access: "delete"}); kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("bad access must be validation, got %v", e)
		}
		if _, e := a.svc.Grant(ctx, db, document.GrantInput{DocumentID: docID, GranteeKind: "capacity", GranteeRef: "   ", Access: "read"}); kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("blank grantee ref must be validation, got %v", e)
		}
		if _, e := a.svc.Grant(ctx, db, document.GrantInput{DocumentID: uuid.New(), GranteeKind: "capacity", GranteeRef: "x", Access: "read"}); kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("grant on missing document must be NotFound, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Revoke — success, double-revoke, not-found
// ---------------------------------------------------------------------------

func TestIntegrationRevokeLifecycle(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("rev"), "text/plain; charset=utf-8")
	stranger := uuid.New()

	var grantID uuid.UUID
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		grantID, e = a.svc.Grant(ctx, db, document.GrantInput{
			DocumentID: docID, GranteeKind: "capacity", GranteeRef: stranger.String(), Access: "read",
		})
		return e
	}); err != nil {
		t.Fatalf("grant: %v", err)
	}
	// Grant is live → stranger can download.
	if _, err := a.download(t, stranger, docID); err != nil {
		t.Fatalf("granted download: %v", err)
	}
	// Owner revokes.
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		return a.svc.Revoke(ctx, db, grantID)
	}); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	// Stranger is now forbidden again.
	if _, err := a.download(t, stranger, docID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("revoked grant must forbid, got %v", err)
	}
	// Re-revoking an already-closed grant → NotFound (no active grant).
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		e := a.svc.Revoke(ctx, db, grantID)
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("double revoke must be NotFound, got %v", e)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// Revoking an unknown grant id → NotFound.
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		e := a.svc.Revoke(ctx, db, uuid.New())
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("revoke of unknown grant must be NotFound, got %v", e)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// UpdateScanStatus — invalid result + settle-once
// ---------------------------------------------------------------------------

func TestIntegrationUpdateScanStatusInvalidResult(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	_, verID := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("scan"), "text/plain; charset=utf-8")
	if e := a.svc.UpdateScanStatus(context.Background(), a.h.PlatformTxM, a.tn, verID, "maybe"); kerr.KindOf(e) != kerr.KindValidation {
		t.Fatalf("invalid scan result must be validation, got %v", e)
	}
}

func TestIntegrationUpdateScanStatusSettlesOnce(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	_, verID := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("scan"), "text/plain; charset=utf-8")
	if err := a.svc.UpdateScanStatus(context.Background(), a.h.PlatformTxM, a.tn, verID, "clean"); err != nil {
		t.Fatalf("first scan settle: %v", err)
	}
	// A second settle (or a settle of an unknown version) affects no rows → conflict.
	if e := a.svc.UpdateScanStatus(context.Background(), a.h.PlatformTxM, a.tn, verID, "infected"); kerr.KindOf(e) != kerr.KindConflict {
		t.Fatalf("re-settling a scan must conflict, got %v", e)
	}
	if e := a.svc.UpdateScanStatus(context.Background(), a.h.PlatformTxM, a.tn, uuid.New(), "clean"); kerr.KindOf(e) != kerr.KindConflict {
		t.Fatalf("settling an unknown version must conflict, got %v", e)
	}
}

// ---------------------------------------------------------------------------
// SweepRetention — multi-document/version sweep + idempotent re-run
// ---------------------------------------------------------------------------

func TestIntegrationSweepMultipleDocsAndReRun(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", Retention: time.Hour})
	// Two documents; one carries two versions.
	d1, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("d1v1"), "text/plain; charset=utf-8")
	// Add a second version to d1.
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		sess, e := a.svc.InitiateUpload(ctx, db, d1)
		if e != nil {
			return e
		}
		body := []byte("d1v2")
		a.store.Put(sess.StorageKey, body)
		_, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: d1, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		return e
	}); err != nil {
		t.Fatalf("second version: %v", err)
	}
	a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("d2v1"), "text/plain; charset=utf-8")

	n, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour))
	if err != nil || n != 3 {
		t.Fatalf("sweep must void 3 versions across 2 docs: n=%d err=%v", n, err)
	}
	// Idempotent: a second sweep voids nothing.
	if n2, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour)); err != nil || n2 != 0 {
		t.Fatalf("re-sweep must be a no-op: n=%d err=%v", n2, err)
	}
}

// ---------------------------------------------------------------------------
// Storage-boundary faults (faultyStore)
// ---------------------------------------------------------------------------

func TestIntegrationInitiateUploadPresignFault(t *testing.T) {
	a, fs := newFaultyHarness(t, document.Class{Key: "core.doc"})
	var docID uuid.UUID
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	fs.failPut = true
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.InitiateUpload(ctx, db, docID)
		if e == nil {
			t.Fatal("a presign-put failure must surface as an error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationConfirmUploadStatFault(t *testing.T) {
	a, fs := newFaultyHarness(t, document.Class{Key: "core.doc"})
	body := []byte("payload")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		fs.failStat = true
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if e == nil {
			t.Fatal("a non-NotFound stat failure must surface as an error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationConfirmUploadPeekFault(t *testing.T) {
	a, fs := newFaultyHarness(t, document.Class{Key: "core.doc"})
	body := []byte("payload")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		fs.failPeek = true // stat succeeds, the MIME-sniff peek fails
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if e == nil {
			t.Fatal("a peek failure must surface as an error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationDownloadPresignFault(t *testing.T) {
	a, fs := newFaultyHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("dl"), "text/plain; charset=utf-8")
	fs.failGet = true
	if _, err := a.download(t, a.actor, docID); err == nil {
		t.Fatal("a presign-get failure must surface as an error")
	}
}

func TestIntegrationSweepDeleteFault(t *testing.T) {
	a, fs := newFaultyHarness(t, document.Class{Key: "core.doc", Retention: time.Hour})
	a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("orphan"), "text/plain; charset=utf-8")
	fs.failDelete = true
	// The row voiding commits; the post-commit blob delete fails and is reported.
	n, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour))
	if err == nil {
		t.Fatal("a post-commit blob-delete failure must be reported")
	}
	if n != 1 {
		t.Fatalf("the version was still voided pre-commit: n=%d", n)
	}
}

// A document row referencing a class no longer in the registry fails confirm
// with an internal error (defensive: the class envelope is required to validate
// the bytes).
func TestIntegrationConfirmUploadUnregisteredClass(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("orphaned class")
	docID := uuid.New()
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, e := db.Exec(ctx,
			`INSERT INTO documents (id, tenant_id, document_class, title, sensitivity, created_by)
			 VALUES ($1, app_tenant_id(), 'core.ghost', 'T', 'internal', $2)`, docID, a.actor); e != nil {
			return e
		}
		sess, e := a.svc.InitiateUpload(ctx, db, docID)
		if e != nil {
			return e
		}
		a.store.Put(sess.StorageKey, body)
		_, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindInternal {
			t.Fatalf("confirm against an unregistered class must be internal, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Remaining reachable branches
// ---------------------------------------------------------------------------

// An empty DeclaredMIME falls back to the sniffed type.
func TestIntegrationConfirmUploadEmptyDeclaredMIME(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("plain text body")
	var verID uuid.UUID
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUpload(ctx, db, docID)
		a.store.Put(sess.StorageKey, body)
		var e error
		verID, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "",
		})
		return e
	})
	if err != nil || verID == uuid.Nil {
		t.Fatalf("empty declared MIME must fall back to the sniff: verID=%v err=%v", verID, err)
	}
	// The stored MIME is the sniffed text type.
	var mime string
	if err := a.h.Admin.QueryRow(context.Background(),
		`SELECT mime_type FROM document_versions WHERE id=$1`, verID).Scan(&mime); err != nil {
		t.Fatal(err)
	}
	if mime == "" {
		t.Fatal("sniffed MIME was not persisted")
	}
}

// Create in a context with no actor id records the zero uuid (actorFromCtx miss).
func TestIntegrationCreateWithoutActor(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	noActor := testkit.TenantCtx(a.tn) // no WithActorID
	var docID uuid.UUID
	err := a.h.TxM.WithTenant(noActor, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		return e
	})
	if err != nil {
		t.Fatalf("create without an actor id: %v", err)
	}
	var createdBy uuid.UUID
	if err := a.h.Admin.QueryRow(context.Background(),
		`SELECT created_by FROM documents WHERE id=$1`, docID).Scan(&createdBy); err != nil {
		t.Fatal(err)
	}
	if createdBy != uuid.Nil {
		t.Fatalf("missing actor must record the zero uuid, got %v", createdBy)
	}
}

// An access hook that returns nil is invoked and lets the download proceed
// (the success path through runAccess).
func TestIntegrationAccessHookAllows(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn), actor)

	var seen bool
	hooks := document.NewHooks()
	hooks.OnDocumentAccess(func(_ context.Context, e document.AccessEvent) error {
		seen = true
		if e.VersionNo != 1 {
			t.Fatalf("access event carried version %d", e.VersionNo)
		}
		return nil // allow
	})
	svc, store := newHookedService(t, hooks)

	var docID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}
		body := []byte("hook allows")
		sess, e := svc.InitiateUpload(ctx, db, docID)
		if e != nil {
			return e
		}
		store.Put(sess.StorageKey, body)
		_, e = svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain; charset=utf-8",
		})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Download(ctx, db,
			authz.Actor{CapacityID: actor, UserID: actor, TenantID: tn},
			document.DownloadInput{DocumentID: docID})
		return e
	}); err != nil {
		t.Fatalf("download with an allowing access hook: %v", err)
	}
	if !seen {
		t.Fatal("the access hook was never invoked")
	}
}
