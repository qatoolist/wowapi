package notify_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
	"github.com/qatoolist/wowapi/testkit/fakes"
)

// ---------------------------------------------------------------------------
// Registry.Keys / Registry.Err
// ---------------------------------------------------------------------------

func TestRegistryKeysSorted(t *testing.T) {
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{Key: "core.notify.zeta", Vars: []string{"X"}})
	reg.Register("core", notify.TemplateSpec{Key: "core.notify.alpha", Vars: []string{"X"}})
	reg.Register("core", notify.TemplateSpec{Key: "core.notify.mid", Vars: []string{"X"}})
	if err := reg.Err(); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	keys := reg.Keys()
	want := []string{"core.notify.alpha", "core.notify.mid", "core.notify.zeta"}
	if len(keys) != len(want) {
		t.Fatalf("Keys() = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("Keys() = %v, want sorted %v", keys, want)
		}
	}
}

// TestRegistryErrJoinsMultiple covers the multi-error join branch of Err (two
// accumulated errors are joined with "; ").
func TestRegistryErrJoinsMultiple(t *testing.T) {
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{Key: "bad key one", Vars: []string{"X"}})    // malformed
	reg.Register("core", notify.TemplateSpec{Key: "other.notify.x", Vars: []string{"X"}}) // foreign prefix
	err := reg.Err()
	if err == nil {
		t.Fatal("expected joined error for two bad registrations")
	}
	if kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("expected KindInternal, got %v", err)
	}
	if !strings.Contains(err.Error(), "; ") {
		t.Fatalf("multiple errors must be joined with '; ', got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// New — nil-argument guard
// ---------------------------------------------------------------------------

func TestNewPanicsOnNilArgs(t *testing.T) {
	assertPanics(t, "nil reg", func() { notify.New(nil, model.UUIDv7()) })
	assertPanics(t, "nil idgen", func() { notify.New(notify.NewRegistry(), nil) })
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

// ---------------------------------------------------------------------------
// FakeSender.Reset
// ---------------------------------------------------------------------------

func TestFakeSenderReset(t *testing.T) {
	f := &fakes.NotifySender{}
	if _, err := f.Send(context.Background(), notify.Delivery{ID: uuid.New()}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if f.Count() != 1 {
		t.Fatalf("Count after send = %d, want 1", f.Count())
	}
	f.Err = context.Canceled
	f.Reset()
	if f.Count() != 0 {
		t.Fatalf("Count after Reset = %d, want 0", f.Count())
	}
	// Err must be cleared too: a post-Reset Send succeeds and records again.
	if _, err := f.Send(context.Background(), notify.Delivery{ID: uuid.New()}); err != nil {
		t.Fatalf("post-Reset Send should succeed, got %v", err)
	}
	if f.Count() != 1 {
		t.Fatalf("Count after post-Reset send = %d, want 1", f.Count())
	}
}

// ---------------------------------------------------------------------------
// RenderBody — parse-error branches (both html and text paths)
// ---------------------------------------------------------------------------

func TestRenderBodyParseErrorEmail(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name"}}
	_, err := notify.RenderBody(spec, notify.ChannelEmail, "Hi {{.Name", map[string]any{})
	if err == nil {
		t.Fatal("expected parse error for malformed email template")
	}
	if kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("email parse failure should be KindInternal, got %v", err)
	}
}

func TestRenderBodyParseErrorText(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name"}}
	_, err := notify.RenderBody(spec, notify.ChannelSMS, "Hi {{.Name", map[string]any{})
	if err == nil {
		t.Fatal("expected parse error for malformed text template")
	}
	if kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("text parse failure should be KindInternal, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Send — input-validation branches
// ---------------------------------------------------------------------------

func TestSendMissingRecipient(t *testing.T) {
	a := newHarness(t)
	_, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: uuid.Nil, // missing
		Variables:        map[string]any{"Name": "X", "Amount": "1"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation for missing recipient, got %v", err)
	}
	if !strings.Contains(err.Error(), "RecipientPartyID") {
		t.Fatalf("error should mention RecipientPartyID, got %q", err.Error())
	}
}

func TestSendNoChannels(t *testing.T) {
	a := newHarness(t)
	_, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: uuid.New(),
		Variables:        map[string]any{"Name": "X", "Amount": "1"},
		Channels:         nil, // none
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation for no channels, got %v", err)
	}
}

func TestSendNoTemplateFound(t *testing.T) {
	a := newHarness(t)
	// The sms channel has no seeded template (only inapp + email exist), so after
	// channel resolution nothing remains and Send fails loudly.
	_, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: uuid.New(),
		Variables:        map[string]any{"Name": "X", "Amount": "1"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelSMS, Destination: "+15550000000"}},
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation for no template found, got %v", err)
	}
	if !strings.Contains(err.Error(), "no template found") {
		t.Fatalf("error should explain missing template, got %q", err.Error())
	}
}

// TestSendSkipsEmptyChannelAndNilVariables covers the empty-Channel `continue`
// branch and the nil-Variables default: an empty ChannelDest is skipped while a
// real one succeeds, and passing no Variables map still renders and writes a row.
// It uses its own registry/template so the body references no variables (so a nil
// Variables map passes the dry-run render).
func TestSendSkipsEmptyChannelAndNilVariables(t *testing.T) {
	h := newVarlessHarness(t)
	party := uuid.New()

	var id uuid.UUID
	err := h.db.TxM.WithTenant(h.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = h.svc.Send(ctx, db, notify.Message{
			TemplateKey:      "core.notify.plain",
			RecipientPartyID: party,
			Variables:        nil, // exercises the renderVars default
			Channels: []notify.ChannelDest{
				{Channel: ""}, // skipped by the empty-channel guard
				{Channel: notify.ChannelInApp},
			},
		})
		return e
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	var count int
	if err := h.db.Admin.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM notification_deliveries WHERE notification_id = $1`, id,
	).Scan(&count); err != nil {
		t.Fatalf("count deliveries: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 delivery (empty channel skipped), got %d", count)
	}
}

// varlessHarness is a minimal service whose template body references no
// variables, so Send succeeds with a nil Variables map.
type varlessHarness struct {
	db  *testkit.DBHandle
	svc *notify.Service
	ctx context.Context
}

func newVarlessHarness(t *testing.T) *varlessHarness {
	t.Helper()
	db := testkit.NewDB(t)
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{Key: "core.notify.plain", Channels: []string{"inapp"}})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	svc := notify.New(reg, model.UUIDv7())
	tenant := testkit.CreateTenant(t, db).ID
	ctx := database.WithActorID(testkit.TenantCtx(tenant), uuid.New())
	seedTemplate(t, db, nil, "core.notify.plain", "inapp", "en", "", "Static hello")
	return &varlessHarness{db: db, svc: svc, ctx: ctx}
}

// TestSendWithResourceAnchor covers the optional resource-anchor branch: a
// non-zero Resource ref is persisted into notifications.resource_type/id.
func TestSendWithResourceAnchor(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()
	ref := resource.Ref{Type: "invoices.invoice", ID: uuid.New()}

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Ivy", "Amount": "12"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
		Resource:         ref,
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	var gotType string
	var gotID uuid.UUID
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT resource_type, resource_id FROM notifications WHERE id = $1`, id,
	).Scan(&gotType, &gotID); err != nil {
		t.Fatalf("read notification: %v", err)
	}
	if gotType != ref.Type || gotID != ref.ID {
		t.Fatalf("resource anchor = (%s,%s), want (%s,%s)", gotType, gotID, ref.Type, ref.ID)
	}
}
