package artifact_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/foundation/artifact"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestNewDefaultsIDGen exercises New(nil): a nil generator must fall back to the
// built-in UUIDv7 source, and the resulting pipeline must still mint usable IDs.
func TestIntegrationNewDefaultsIDGen(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(nil) // nil → model.UUIDv7() fallback branch
	ctx := actx(uuid.New())

	var a artifact.Artifact
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		a, e = p.Generate(ctx, db, artifact.Input{Kind: "receipt", Content: []byte("x")})
		return e
	}); err != nil {
		t.Fatalf("generate with default idgen: %v", err)
	}
	if a.ID == uuid.Nil {
		t.Fatal("default idgen must produce a non-nil UUID")
	}
	if a.Version != 1 {
		t.Fatalf("version = %d, want 1", a.Version)
	}
}

// TestGenerateWithoutActor covers actorOrNil's no-actor branch: a tenant-bound
// context that carries no actor id stores created_by as NULL.
func TestIntegrationGenerateWithoutActor(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	tenant := uuid.New()
	// Tenant binding only — deliberately no WithActorID, so actorOrNil returns nil.
	ctx := database.WithTenantID(context.Background(), tenant)

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		a, e := p.Generate(ctx, db, artifact.Input{Kind: "receipt", Content: []byte("x")})
		id = a.ID
		return e
	}); err != nil {
		t.Fatalf("generate without actor: %v", err)
	}

	var createdBy *uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT created_by FROM artifacts WHERE id = $1`, id).Scan(&createdBy); err != nil {
		t.Fatalf("read created_by: %v", err)
	}
	if createdBy != nil {
		t.Fatalf("created_by = %v, want NULL when no actor is bound", createdBy)
	}
}

// TestGenerateValidation covers the fast-fail guard: empty kind or empty content
// is a KindValidation error and never touches the database.
func TestIntegrationGenerateValidation(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	cases := []struct {
		name string
		in   artifact.Input
	}{
		{"empty kind", artifact.Input{Kind: "", Content: []byte("x")}},
		{"empty content", artifact.Input{Kind: "receipt", Content: nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
				_, e := p.Generate(ctx, db, tc.in)
				return e
			})
			if err == nil {
				t.Fatal("want validation error, got nil")
			}
			if kerr.KindOf(err) != kerr.KindValidation {
				t.Fatalf("kind = %v, want KindValidation", kerr.KindOf(err))
			}
			if e, ok := kerr.As(err); !ok || e.Code != "invalid_artifact" {
				t.Fatalf("code = %+v, want invalid_artifact", e)
			}
		})
	}
}

// TestGenerateSidecarMarshalError covers the json.Marshal failure branch: a
// sidecar carrying an unencodable value (a channel) fails before the INSERT.
func TestIntegrationGenerateSidecarMarshalError(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := p.Generate(ctx, db, artifact.Input{
			Kind:    "receipt",
			Content: []byte("x"),
			Sidecar: map[string]any{"bad": make(chan int)}, // json.Marshal cannot encode a channel
		})
		return e
	})
	if err == nil {
		t.Fatal("want marshal error, got nil")
	}
	if !strings.Contains(err.Error(), "marshal sidecar") {
		t.Fatalf("error = %v, want it to mention marshal sidecar", err)
	}
}

// TestGenerateInsertError covers the generic (non-unique) INSERT failure path:
// a table CHECK constraint rejects the row, so Generate reports an insert error
// and isUniqueViolation returns false (SQLSTATE 23514, not 23505).
func TestIntegrationGenerateInsertError(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	// Add a CHECK constraint that the INSERT will violate. Owner-only DDL on this
	// test's exclusive database; app_rt still does the INSERT under RLS.
	if _, err := h.Admin.Exec(context.Background(),
		`ALTER TABLE artifacts ADD CONSTRAINT no_forbidden_kind CHECK (kind <> 'forbidden')`); err != nil {
		t.Fatalf("add check constraint: %v", err)
	}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := p.Generate(ctx, db, artifact.Input{Kind: "forbidden", Content: []byte("x")})
		return e
	})
	if err == nil {
		t.Fatal("want insert error from CHECK violation, got nil")
	}
	if !strings.Contains(err.Error(), "insert artifact") {
		t.Fatalf("error = %v, want it to mention insert artifact", err)
	}
	// A CHECK violation (23514) must NOT be classified as a version conflict.
	if kerr.KindOf(err) == kerr.KindConflict {
		t.Fatalf("a CHECK violation must not be reported as a version conflict: %v", err)
	}
}

// TestIntegrationGenerateOverlappingSucceeds is the counter-fix regression:
// a first transaction inserts version 1 but holds open (uncommitted); a second
// transaction used to compute the same next version via MAX(version)+1 and
// collide on the unique index. With the locked per-(tenant,kind) counter, the
// second transaction blocks on the counter row, then receives version 2 once
// the holder commits — both generates succeed.
func TestIntegrationGenerateOverlappingSucceeds(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	inserted := make(chan struct{})
	release := make(chan struct{})

	var holderErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		holderErr = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			if _, e := p.Generate(ctx, db, artifact.Input{Kind: "race", Content: []byte("first")}); e != nil {
				return e
			}
			close(inserted) // version 1 inserted, still uncommitted
			<-release       // hold the transaction open so the second blocks on the counter
			return nil
		})
	}()

	<-inserted

	secondCh := make(chan error, 1)
	go func() {
		secondCh <- h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			_, e := p.Generate(ctx, db, artifact.Input{Kind: "race", Content: []byte("second")})
			return e
		})
	}()

	// Give the second transaction time to reach and block on the counter row,
	// then let the holder commit so the second can allocate version 2.
	time.Sleep(300 * time.Millisecond)
	close(release)

	secondErr := <-secondCh
	wg.Wait()

	if holderErr != nil {
		t.Fatalf("holder transaction must commit version 1: %v", holderErr)
	}
	if secondErr != nil {
		t.Fatalf("second generate must succeed with counter allocation, got %v", secondErr)
	}

	var versions []int
	rows, err := h.Admin.Query(context.Background(),
		`SELECT version FROM artifacts WHERE kind='race' ORDER BY version`)
	if err != nil {
		t.Fatalf("list artifacts: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan version: %v", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	if len(versions) != 2 || versions[0] != 1 || versions[1] != 2 {
		t.Fatalf("want versions [1 2], got %v", versions)
	}
}

// TestGetNotFoundAndVerifyPropagates covers Get's ErrNoRows → KindNotFound branch
// and Verify's error passthrough (it returns false and the Get error unchanged).
func TestIntegrationGetNotFoundAndVerifyPropagates(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())
	missing := uuid.New()

	var getErr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, getErr = p.Get(ctx, db, missing)
		return nil
	})
	if getErr == nil {
		t.Fatal("Get of a missing id must error")
	}
	if kerr.KindOf(getErr) != kerr.KindNotFound {
		t.Fatalf("kind = %v, want KindNotFound", kerr.KindOf(getErr))
	}
	if e, ok := kerr.As(getErr); !ok || e.Code != "not_found" {
		t.Fatalf("code = %+v, want not_found", e)
	}

	var ok bool
	var verifyErr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		ok, verifyErr = p.Verify(ctx, db, missing)
		return nil
	})
	if ok {
		t.Fatal("Verify of a missing id must not report ok")
	}
	if verifyErr == nil || kerr.KindOf(verifyErr) != kerr.KindNotFound {
		t.Fatalf("Verify must propagate the not-found error, got %v", verifyErr)
	}
}

// TestList covers the happy List path end to end: newest-version-first ordering,
// per-row template_version hydration (the tmpl != nil branch and its skip), and
// kind scoping.
func TestIntegrationList(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	gen := func(kind, tmpl string) {
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			_, e := p.Generate(ctx, db, artifact.Input{Kind: kind, Content: []byte(kind + tmpl), TemplateVersion: tmpl})
			return e
		}); err != nil {
			t.Fatalf("gen %s/%s: %v", kind, tmpl, err)
		}
	}
	gen("receipt", "")   // version 1, NULL template → tmpl == nil skip branch
	gen("receipt", "v2") // version 2, template set → tmpl != nil branch
	gen("receipt", "v3") // version 3
	gen("certificate", "")

	var list []artifact.Artifact
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		list, e = p.List(ctx, db, "receipt")
		return e
	}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len(list) = %d, want 3", len(list))
	}
	// Newest version first.
	if list[0].Version != 3 || list[1].Version != 2 || list[2].Version != 1 {
		t.Fatalf("versions = [%d %d %d], want [3 2 1]", list[0].Version, list[1].Version, list[2].Version)
	}
	if list[0].TemplateVersion != "v3" || list[1].TemplateVersion != "v2" {
		t.Fatalf("template versions = %q,%q, want v3,v2", list[0].TemplateVersion, list[1].TemplateVersion)
	}
	if list[2].TemplateVersion != "" {
		t.Fatalf("version 1 template = %q, want empty (NULL)", list[2].TemplateVersion)
	}
	for _, a := range list {
		if a.Kind != "receipt" {
			t.Fatalf("List returned foreign kind %q", a.Kind)
		}
		if a.ContentHash == "" {
			t.Fatal("List rows must carry the content hash")
		}
		if a.Content != nil {
			t.Fatal("List must not include content bytes")
		}
	}
}

// TestGetAndListQueryErrors covers the non-ErrNoRows read-error branch in Get and
// the query-error branch in List: dropping a selected column makes both SELECTs
// fail at plan time with a wrapped error that is not KindNotFound.
func TestIntegrationGetAndListQueryErrors(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		a, e := p.Generate(ctx, db, artifact.Input{Kind: "receipt", Content: []byte("x")})
		id = a.ID
		return e
	}); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Remove a column both queries SELECT; the statements now fail to plan.
	if _, err := h.Admin.Exec(context.Background(),
		`ALTER TABLE artifacts DROP COLUMN content_type`); err != nil {
		t.Fatalf("drop column: %v", err)
	}

	var getErr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, getErr = p.Get(ctx, db, id)
		return nil
	})
	if getErr == nil || !strings.Contains(getErr.Error(), "read artifact") {
		t.Fatalf("Get error = %v, want it to mention read artifact", getErr)
	}
	if kerr.KindOf(getErr) == kerr.KindNotFound {
		t.Fatalf("a plan-time read error must not be reported as not-found: %v", getErr)
	}

	var listErr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, listErr = p.List(ctx, db, "receipt")
		return nil
	})
	if listErr == nil || !strings.Contains(listErr.Error(), "query artifacts") {
		t.Fatalf("List error = %v, want it to mention query artifacts", listErr)
	}
}

// TestTemplateRegisterUpdatesExisting covers Register's in-place update branch:
// re-registering an existing version replaces its effective date rather than
// appending a duplicate, which Resolve then reflects.
func TestTemplateRegisterUpdatesExisting(t *testing.T) {
	tm := artifact.NewTemplates()
	d := func(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }

	tm.Register("receipt", "v1", d(2020))
	tm.Register("receipt", "v1", d(2025)) // same version → update EffectiveFrom in place

	// With the updated date, v1 is not effective in 2024...
	if _, ok := tm.Resolve("receipt", d(2024)); ok {
		t.Error("after update, v1 must not be effective before 2025")
	}
	// ...but is from 2025 onward, and remains the only registered version.
	v, ok := tm.Resolve("receipt", d(2026))
	if !ok || v.Version != "v1" {
		t.Fatalf("resolved %+v, ok=%v; want v1", v, ok)
	}
	if !v.EffectiveFrom.Equal(d(2025)) {
		t.Fatalf("EffectiveFrom = %v, want %v (updated in place)", v.EffectiveFrom, d(2025))
	}
}
