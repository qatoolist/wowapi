package notify_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/notify"
	"github.com/qatoolist/wowapi/testkit"
)

// harness bundles everything needed by notify integration tests.
type harness struct {
	db     *testkit.DBHandle
	reg    *notify.Registry
	fake   *notify.FakeSender
	svc    *notify.Service
	tenant uuid.UUID
	actor  uuid.UUID
	ctx    context.Context
}

// seedTemplate inserts a notification_templates row via the admin pool
// (bypassing RLS). A nil tenantID inserts a platform default (tenant_id=NULL).
func seedTemplate(t *testing.T, h *testkit.DBHandle, tenantID *uuid.UUID, key, channel, locale, subject, body string) {
	t.Helper()
	id := uuid.New()
	var tenantArg any
	if tenantID != nil {
		tenantArg = *tenantID
	}
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO notification_templates
		    (id, tenant_id, key, channel, locale, subject, body, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, tenantArg, key, channel, locale, subject, body, uuid.Nil,
	); err != nil {
		t.Fatalf("seedTemplate: %v", err)
	}
}

// newHarness creates a fresh DB, registry, service, and tenant for one test.
func newHarness(t *testing.T) *harness {
	t.Helper()
	db := testkit.NewDB(t)
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:      "core.notify.welcome",
		Vars:     []string{"Name", "Amount"},
		Channels: []string{"inapp", "email"},
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}

	fake := &notify.FakeSender{}
	svc := notify.New(reg, model.UUIDv7())
	svc.RegisterSender(notify.ChannelEmail, fake)

	tenant := testkit.CreateTenant(t, db).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tenant), actor)

	// Seed platform-default templates (tenant_id=NULL).
	seedTemplate(t, db, nil, "core.notify.welcome", "inapp", "en", "", "Hello {{.Name}}")
	seedTemplate(t, db, nil, "core.notify.welcome", "email", "en", "Welcome {{.Name}}", "Hi {{.Name}}, your amount is {{.Amount}}")

	return &harness{
		db: db, reg: reg, fake: fake, svc: svc,
		tenant: tenant, actor: actor, ctx: ctx,
	}
}

// send is a helper that calls svc.Send inside a TxM.WithTenant.
func (a *harness) send(t *testing.T, msg notify.Message) (uuid.UUID, error) {
	t.Helper()
	var id uuid.UUID
	err := a.db.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = a.svc.Send(ctx, db, msg)
		return e
	})
	return id, err
}

// ---------------------------------------------------------------------------
// Registry validation
// ---------------------------------------------------------------------------

func TestRegistryBadKeyRejected(t *testing.T) {
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:  "bad-key", // must be module.area.name
		Vars: []string{"X"},
	})
	if err := reg.Err(); err == nil {
		t.Fatal("expected error for malformed key, got nil")
	}
}

func TestRegistryForeignModuleKeyRejected(t *testing.T) {
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:  "other.notify.test", // prefix mismatch
		Vars: []string{"X"},
	})
	if err := reg.Err(); err == nil {
		t.Fatal("expected error for foreign-module key, got nil")
	}
}

func TestRegistryDuplicateKeyRejected(t *testing.T) {
	reg := notify.NewRegistry()
	spec := notify.TemplateSpec{Key: "core.notify.dup", Vars: []string{"X"}}
	reg.Register("core", spec)
	reg.Register("core", spec) // duplicate
	if err := reg.Err(); err == nil {
		t.Fatal("expected error for duplicate key, got nil")
	}
}

// ---------------------------------------------------------------------------
// ValidateBody
// ---------------------------------------------------------------------------

func TestValidateBodyAcceptsAllowlistedVars(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name", "Count"}}
	if err := notify.ValidateBody(spec, "Hello {{.Name}}, count: {{.Count}}"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateBodyRejectsUnknownVar(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.bad", Vars: []string{"Name"}}
	err := notify.ValidateBody(spec, "Hello {{.Name}} and {{.Secret}}")
	if err == nil {
		t.Fatal("expected error for template referencing undeclared var, got nil")
	}
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation, got %v", err)
	}
}

func TestValidateBodyRejectsBadTemplate(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.bad", Vars: []string{"Name"}}
	err := notify.ValidateBody(spec, "{{unclosed")
	if err == nil {
		t.Fatal("expected error for invalid template syntax, got nil")
	}
}

// ---------------------------------------------------------------------------
// RenderBody
// ---------------------------------------------------------------------------

func TestRenderBodyRejectsDisallowedVar(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name"}}
	_, err := notify.RenderBody(spec, notify.ChannelSMS, "Hello {{.Name}}", map[string]any{"Name": "Alice", "Hidden": "x"})
	if err == nil {
		t.Fatal("expected error for disallowed var in render, got nil")
	}
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation, got %v", err)
	}
}

func TestRenderBodyMissingKeyErrors(t *testing.T) {
	// missingkey=error: template referencing .Name but Name not in vars → error.
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name"}}
	_, err := notify.RenderBody(spec, notify.ChannelSMS, "Hello {{.Name}}", map[string]any{})
	if err == nil {
		t.Fatal("expected error when template references var not in vars map, got nil")
	}
}

func TestRenderBodySuccess(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name"}}
	out, err := notify.RenderBody(spec, notify.ChannelSMS, "Hello {{.Name}}", map[string]any{"Name": "Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Hello Alice" {
		t.Fatalf("unexpected output: %q", out)
	}
}

// TestRenderBodyEmailEscapesHTML is the SEC-51 regression: a variable value
// carrying markup is auto-escaped when rendered for the email (HTML) channel,
// but left raw for a plain-text channel (sms).
func TestRenderBodyEmailEscapesHTML(t *testing.T) {
	spec := notify.TemplateSpec{Key: "core.notify.ok", Vars: []string{"Name"}}
	payload := map[string]any{"Name": "<script>alert(1)</script>"}

	// Email → html/template auto-escapes.
	email, err := notify.RenderBody(spec, notify.ChannelEmail, "Hi {{.Name}}", payload)
	if err != nil {
		t.Fatalf("email render: %v", err)
	}
	if strings.Contains(email, "<script>") {
		t.Fatalf("email body must escape HTML, got raw markup: %q", email)
	}
	if !strings.Contains(email, "&lt;script&gt;") {
		t.Fatalf("email body must contain escaped markup, got: %q", email)
	}

	// SMS → text/template leaves it raw (plain-text transport).
	sms, err := notify.RenderBody(spec, notify.ChannelSMS, "Hi {{.Name}}", payload)
	if err != nil {
		t.Fatalf("sms render: %v", err)
	}
	if !strings.Contains(sms, "<script>alert(1)</script>") {
		t.Fatalf("sms body must keep the value raw, got: %q", sms)
	}
}

// ---------------------------------------------------------------------------
// Send — integration
// ---------------------------------------------------------------------------

func TestSendRejectsUnknownTemplateKey(t *testing.T) {
	a := newHarness(t)
	_, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.nonexistent",
		RecipientPartyID: uuid.New(),
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation for unknown key, got %v", err)
	}
}

func TestSendRejectsDisallowedVariable(t *testing.T) {
	a := newHarness(t)
	_, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: uuid.New(),
		Variables:        map[string]any{"Name": "Alice", "Secret": "hack"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation for disallowed var, got %v", err)
	}
}

// TestSendRejectsIncompleteVariables is the ARCH-77 regression: the email
// template body references {{.Amount}}, but the caller supplies only Name. Send
// must fail SYNCHRONOUSLY (KindValidation) via the dry-run render and write no
// rows — rather than committing a notification that fails at delivery time.
func TestSendRejectsIncompleteVariables(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Alice"}, // missing Amount
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "alice@example.test"}},
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("expected KindValidation for missing referenced var, got %v", err)
	}
	if id != uuid.Nil {
		t.Fatalf("expected no notification id, got %s", id)
	}

	// No rows must have been written (dry-run happens before any INSERT).
	var count int
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM notifications WHERE recipient_party_id = $1`, party,
	).Scan(&count); err != nil {
		t.Fatalf("count check: %v", err)
	}
	if count != 0 {
		t.Fatalf("no notification must be written on incomplete-vars rejection, got %d", count)
	}
}

func TestSendWritesNotificationAndDeliveries(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Alice", "Amount": "50"},
		Channels: []notify.ChannelDest{
			{Channel: notify.ChannelInApp},
			{Channel: notify.ChannelEmail, Destination: "alice@example.test"},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("expected non-nil notification id")
	}

	// Verify notification row via admin (bypasses RLS for cross-check).
	var templateKey string
	var importance string
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT template_key, importance FROM notifications WHERE id = $1`, id,
	).Scan(&templateKey, &importance); err != nil {
		t.Fatalf("read notification: %v", err)
	}
	if templateKey != "core.notify.welcome" {
		t.Fatalf("unexpected template_key: %s", templateKey)
	}
	if importance != "normal" {
		t.Fatalf("unexpected importance: %s", importance)
	}

	// Verify two delivery rows (inapp + email).
	var count int
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM notification_deliveries WHERE notification_id = $1`, id,
	).Scan(&count); err != nil {
		t.Fatalf("count deliveries: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 deliveries, got %d", count)
	}
}

func TestSendAtomicRollback(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	// Send inside a tx that forces a rollback.
	var notifID uuid.UUID
	_ = a.db.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		notifID, e = a.svc.Send(ctx, db, notify.Message{
			TemplateKey:      "core.notify.welcome",
			RecipientPartyID: party,
			Variables:        map[string]any{"Name": "Alice", "Amount": "10"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
		})
		if e != nil {
			return e
		}
		return errors.New("force rollback") // rolls back the entire tx
	})

	// The notification must not exist after rollback.
	var count int
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM notifications WHERE id = $1`, notifID,
	).Scan(&count); err != nil {
		t.Fatalf("count check: %v", err)
	}
	if count != 0 {
		t.Fatalf("rolled-back notification persists: count=%d", count)
	}
}

// ---------------------------------------------------------------------------
// Template resolution
// ---------------------------------------------------------------------------

func TestTemplateResolutionTenantOverridesBeatsPlatform(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	// Seed a tenant-specific email template that overrides the platform default.
	seedTemplate(t, a.db, &a.tenant, "core.notify.welcome", "email", "en",
		"Tenant subject", "Tenant body {{.Name}}")

	// Send should pick the tenant row (not error).
	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Bob", "Amount": "99"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "bob@example.test"}},
	})
	if err != nil {
		t.Fatalf("Send with tenant override: %v", err)
	}

	// Verify exactly one delivery for the email channel.
	var count int
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM notification_deliveries WHERE notification_id = $1 AND channel = 'email'`, id,
	).Scan(&count); err != nil {
		t.Fatalf("count deliveries: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 email delivery, got %d", count)
	}
}

func TestTemplateLocaleWithFallback(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	// Only "en" template exists for email (seeded in newHarness). Requesting
	// "hi-IN" should fall back through "hi" → "en" and succeed.
	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Priya", "Amount": "500"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "priya@example.test"}},
		Locale:           "hi-IN",
	})
	if err != nil {
		t.Fatalf("Send with locale hi-IN fallback: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("expected non-nil id")
	}
}

func TestTemplateLocaleExactMatch(t *testing.T) {
	a := newHarness(t)

	// Seed a hi locale template.
	seedTemplate(t, a.db, nil, "core.notify.welcome", "email", "hi",
		"Namaste {{.Name}}", "Namaste {{.Name}}, rakam {{.Amount}}")

	party := uuid.New()
	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Raj", "Amount": "200"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "raj@example.test"}},
		Locale:           "hi",
	})
	if err != nil {
		t.Fatalf("Send with hi locale: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("expected non-nil id")
	}
}

// ---------------------------------------------------------------------------
// SendPending — integration
// ---------------------------------------------------------------------------

func TestSendPendingDeliversQueued(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Carlos", "Amount": "5"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "carlos@example.test"}},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	n, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, time.Now())
	if err != nil {
		t.Fatalf("SendPending: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 sent, got %d", n)
	}
	if a.fake.Count() != 1 {
		t.Fatalf("expected 1 FakeSender call, got %d", a.fake.Count())
	}

	// Delivery row must be 'sent' with a provider_message_id.
	var status string
	var providerID string
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT status, COALESCE(provider_message_id, '') FROM notification_deliveries
		  WHERE notification_id = $1`, id,
	).Scan(&status, &providerID); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	if status != "sent" {
		t.Fatalf("expected status=sent, got %s", status)
	}
	if providerID == "" {
		t.Fatal("expected provider_message_id to be set")
	}
}

func TestSendPendingInAppChannel(t *testing.T) {
	// In-app channel: built-in sender succeeds immediately, no external call.
	a := newHarness(t)
	party := uuid.New()

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Diane", "Amount": "0"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	n, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, time.Now())
	if err != nil {
		t.Fatalf("SendPending: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 sent, got %d", n)
	}

	var status string
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT status FROM notification_deliveries WHERE notification_id = $1`, id,
	).Scan(&status); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	if status != "sent" {
		t.Fatalf("expected status=sent for inapp, got %s", status)
	}
}

func TestSendPendingRetriesAndDeadLetters(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	// Configure the fake sender to always fail.
	a.fake.Err = errors.New("smtp down")

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Eve", "Amount": "1"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "eve@example.test"}},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	base := time.Now()

	// First attempt → failed, attempts = 1, next_attempt_at set (backoff).
	if _, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, base); err != nil {
		t.Fatalf("SendPending attempt 1: %v", err)
	}
	checkDelivery(t, a.db, id, "failed", 1)

	// ARCH-75: immediately re-running with the SAME now must NOT re-claim the
	// failed delivery — its backoff has not elapsed.
	if n, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, base); err != nil {
		t.Fatalf("SendPending before backoff: %v", err)
	} else if n != 0 {
		t.Fatalf("failed delivery must not be re-claimed before backoff elapses, got %d", n)
	}
	checkDelivery(t, a.db, id, "failed", 1) // still attempt 1

	// Advance now past the first backoff → re-claimed, attempts = 2.
	if _, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, base.Add(1*time.Hour)); err != nil {
		t.Fatalf("SendPending attempt 2: %v", err)
	}
	checkDelivery(t, a.db, id, "failed", 2)

	// Advance again → third attempt (= maxAttempts) → dead.
	if _, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, base.Add(2*time.Hour)); err != nil {
		t.Fatalf("SendPending attempt 3: %v", err)
	}
	checkDelivery(t, a.db, id, "dead", 3)

	// Even far in the future, a 'dead' delivery must not be re-claimed.
	n, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, a.tenant, base.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("SendPending after dead: %v", err)
	}
	if n != 0 {
		t.Fatalf("dead delivery must not be re-sent, got %d sent", n)
	}
}

// checkDelivery asserts the status and attempts of the delivery for notification id.
func checkDelivery(t *testing.T, db *testkit.DBHandle, notifID uuid.UUID, wantStatus string, wantAttempts int) {
	t.Helper()
	var status string
	var attempts int
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT status, attempts FROM notification_deliveries WHERE notification_id = $1`, notifID,
	).Scan(&status, &attempts); err != nil {
		t.Fatalf("checkDelivery: %v", err)
	}
	if status != wantStatus {
		t.Errorf("expected status=%s, got %s", wantStatus, status)
	}
	if attempts != wantAttempts {
		t.Errorf("expected attempts=%d, got %d", wantAttempts, attempts)
	}
}

// ---------------------------------------------------------------------------
// ListForParty
// ---------------------------------------------------------------------------

func TestListForParty(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	// Send two notifications to the same party.
	id1, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Frank", "Amount": "10"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if err != nil {
		t.Fatalf("Send 1: %v", err)
	}
	id2, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Frank", "Amount": "20"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if err != nil {
		t.Fatalf("Send 2: %v", err)
	}

	var notifs []notify.Notification
	if err := a.db.TxM.WithTenantRO(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		notifs, e = a.svc.ListForParty(ctx, db, party)
		return e
	}); err != nil {
		t.Fatalf("ListForParty: %v", err)
	}

	if len(notifs) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notifs))
	}
	// Newest first: id2 should come before id1.
	if notifs[0].ID != id2 {
		t.Errorf("expected newest first: got %s, want %s", notifs[0].ID, id2)
	}
	if notifs[1].ID != id1 {
		t.Errorf("expected second: got %s, want %s", notifs[1].ID, id1)
	}
}

// ---------------------------------------------------------------------------
// Tenant isolation
// ---------------------------------------------------------------------------

func TestTenantIsolation(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()

	// Send a notification as tenant A.
	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Grace", "Amount": "30"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelInApp}},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Create a second tenant.
	tenantB := testkit.CreateTenant(t, a.db).ID
	ctxB := testkit.TenantCtx(tenantB)

	// Tenant B queries for the same party — must see nothing.
	var notifs []notify.Notification
	if err := a.db.TxM.WithTenantRO(ctxB, func(ctx context.Context, db database.TenantDB) error {
		var e error
		notifs, e = a.svc.ListForParty(ctx, db, party)
		return e
	}); err != nil {
		t.Fatalf("ListForParty as tenant B: %v", err)
	}
	if len(notifs) != 0 {
		t.Fatalf("tenant B must not see tenant A's notifications: got %d", len(notifs))
	}

	// Also verify via admin that the row exists (sanity check).
	var count int
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM notifications WHERE id = $1`, id,
	).Scan(&count); err != nil {
		t.Fatalf("admin count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 notification in admin view, got %d", count)
	}
}

func TestSendPendingTenantIsolation(t *testing.T) {
	// SendPending for tenant B must not process tenant A's queued deliveries.
	a := newHarness(t)
	party := uuid.New()

	if _, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Hugo", "Amount": "7"},
		Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "hugo@example.test"}},
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Run SendPending as a different tenant — must process zero deliveries.
	tenantB := testkit.CreateTenant(t, a.db).ID
	n, err := a.svc.SendPending(context.Background(), a.db.PlatformTxM, tenantB, time.Now())
	if err != nil {
		t.Fatalf("SendPending tenant B: %v", err)
	}
	if n != 0 {
		t.Fatalf("tenant B processed tenant A's deliveries: n=%d", n)
	}

	// Tenant A's delivery must still be queued.
	var status string
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT nd.status FROM notification_deliveries nd
		   JOIN notifications n ON n.id = nd.notification_id
		  WHERE n.recipient_party_id = $1`, party,
	).Scan(&status); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	if status != "queued" {
		t.Fatalf("expected delivery to still be queued, got %s", status)
	}
}
