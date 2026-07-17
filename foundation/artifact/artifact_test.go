package artifact_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/foundation/artifact"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

func actx(tenant uuid.UUID) context.Context {
	return database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())
}

func TestIntegrationArtifactGenerateGetVerify(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := actx(tenant)

	content := []byte("%PDF-1.7 fake receipt bytes")
	var a artifact.Artifact
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		a, e = p.Generate(ctx, db, artifact.Input{
			Kind: "receipt", Content: content,
			Sidecar:         map[string]any{"total": "150"},
			TemplateVersion: "v1", EffectiveDate: time.Now(),
		})
		return e
	}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if a.Version != 1 || a.ContentHash == "" {
		t.Fatalf("artifact = %+v, want version 1 + a hash", a)
	}

	// Get returns content, hash, sidecar, template.
	var got artifact.Artifact
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		got, e = p.Get(ctx, db, a.ID)
		return e
	})
	if string(got.Content) != string(content) {
		t.Fatalf("content round-trip failed")
	}
	if got.Sidecar["total"] != "150" || got.TemplateVersion != "v1" {
		t.Fatalf("metadata = %+v", got)
	}

	// Verify: untampered artifact matches its hash.
	var ok bool
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		ok, e = p.Verify(ctx, db, a.ID)
		return e
	})
	if !ok {
		t.Fatal("a freshly generated artifact must verify")
	}

	// Tamper the content out-of-band (admin bypasses append-only) → Verify fails.
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE artifacts SET content = $2 WHERE id = $1`, a.ID, []byte("tampered")); err != nil {
		t.Fatal(err)
	}
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		ok, e = p.Verify(ctx, db, a.ID)
		return e
	})
	if ok {
		t.Fatal("Verify must detect a mutated artifact (hash mismatch)")
	}
}

func TestIntegrationArtifactVersioning(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	gen := func(kind string) int {
		var v int
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			a, e := p.Generate(ctx, db, artifact.Input{Kind: kind, Content: []byte("x")})
			v = a.Version
			return e
		}); err != nil {
			t.Fatalf("gen %s: %v", kind, err)
		}
		return v
	}
	if gen("receipt") != 1 || gen("receipt") != 2 || gen("receipt") != 3 {
		t.Fatal("receipt versions must increment 1,2,3")
	}
	// A different kind has its own version series.
	if gen("certificate") != 1 {
		t.Fatal("certificate must start at version 1")
	}
}

func TestIntegrationArtifactAppendOnly(t *testing.T) {
	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())
	var id uuid.UUID
	_ = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		a, e := p.Generate(ctx, db, artifact.Input{Kind: "receipt", Content: []byte("x")})
		id = a.ID
		return e
	})

	// app_rt must not be able to mutate or delete an artifact.
	updErr := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `UPDATE artifacts SET content_hash = 'x' WHERE id = $1`, id)
		return e
	})
	if updErr == nil || !strings.Contains(strings.ToLower(updErr.Error()), "denied") {
		t.Fatalf("app_rt UPDATE on artifacts must be denied, got %v", updErr)
	}
	delErr := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `DELETE FROM artifacts WHERE id = $1`, id)
		return e
	})
	if delErr == nil || !strings.Contains(strings.ToLower(delErr.Error()), "denied") {
		t.Fatalf("app_rt DELETE on artifacts must be denied, got %v", delErr)
	}
}

func TestTemplateResolveByEffectiveDate(t *testing.T) {
	tm := artifact.NewTemplates()
	d := func(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }
	tm.Register("receipt", "v1", d(2024))
	tm.Register("receipt", "v2", d(2026))

	// Before any version → not found.
	if _, ok := tm.Resolve("receipt", d(2023)); ok {
		t.Error("no version should be effective before 2024")
	}
	// 2025 → v1 (v2 not yet effective).
	if v, ok := tm.Resolve("receipt", d(2025)); !ok || v.Version != "v1" {
		t.Errorf("2025 resolved to %+v, want v1", v)
	}
	// 2026+ → v2.
	if v, ok := tm.Resolve("receipt", d(2027)); !ok || v.Version != "v2" {
		t.Errorf("2027 resolved to %+v, want v2", v)
	}
	// Unknown kind → not found.
	if _, ok := tm.Resolve("unknown", d(2027)); ok {
		t.Error("unknown kind must not resolve")
	}
}
