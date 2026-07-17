package notify_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/safety"
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

// TestRegisterSender_RejectsUndeclaredAdapter proves the boot-time duplicate-
// safety contract: a ChannelSender without safety.Declarer is rejected.
func TestRegisterSender_RejectsUndeclaredAdapter(t *testing.T) {
	svc := notify.New(notify.NewRegistry(), model.UUIDv7())
	defer func() {
		if recover() == nil {
			t.Fatal("expected RegisterSender to panic on adapter without safety.Declarer")
		}
	}()
	svc.RegisterSender(notify.ChannelEmail, undeclaredChannelSender{})
}

// TestRegisterSender_AcceptsDeclaredAdapter proves a correctly-declared
// ChannelSender registers successfully.
func TestRegisterSender_AcceptsDeclaredAdapter(t *testing.T) {
	svc := notify.New(notify.NewRegistry(), model.UUIDv7())
	svc.RegisterSender(notify.ChannelEmail, &declaredChannelSender{mechanism: safety.InboxEffectLedger})
}

// undeclaredChannelSender implements notify.ChannelSender but not safety.Declarer.
type undeclaredChannelSender struct{}

func (undeclaredChannelSender) Send(context.Context, notify.Delivery) (string, error) {
	return "", nil
}

// declaredChannelSender implements both notify.ChannelSender and safety.Declarer.
type declaredChannelSender struct {
	mechanism safety.Mechanism
}

func (s *declaredChannelSender) Send(context.Context, notify.Delivery) (string, error) {
	return "", nil
}

func (s *declaredChannelSender) DuplicateSafety() safety.Mechanism { return s.mechanism }

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

// TestIntegrationChannelPrefSkipsOptedOut is the R5 channel-preferences check: a
// recipient opted out of a channel gets no delivery on it.
func TestIntegrationChannelPrefSkipsOptedOut(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()
	if err := a.db.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		return a.svc.SetChannelPref(ctx, db, party, notify.ChannelEmail, false)
	}); err != nil {
		t.Fatalf("set pref: %v", err)
	}

	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Kai", "Amount": "1"},
		Channels: []notify.ChannelDest{
			{Channel: notify.ChannelInApp},
			{Channel: notify.ChannelEmail, Destination: "kai@example.test"},
		},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	var count int
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM notification_deliveries WHERE notification_id = $1`, id).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("deliveries = %d, want 1 (email opted out)", count)
	}
	var ch string
	if err := a.db.Admin.QueryRow(context.Background(),
		`SELECT channel FROM notification_deliveries WHERE notification_id = $1`, id).Scan(&ch); err != nil {
		t.Fatal(err)
	}
	if ch != "inapp" {
		t.Fatalf("remaining channel = %q, want inapp", ch)
	}
}

// TestIntegrationChannelPrefAllOptedOut: opting out of every requested channel
// makes Send fail loudly rather than silently sending nothing.
func TestIntegrationChannelPrefAllOptedOut(t *testing.T) {
	a := newHarness(t)
	party := uuid.New()
	if err := a.db.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		if e := a.svc.SetChannelPref(ctx, db, party, notify.ChannelInApp, false); e != nil {
			return e
		}
		return a.svc.SetChannelPref(ctx, db, party, notify.ChannelEmail, false)
	}); err != nil {
		t.Fatal(err)
	}
	_, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: party,
		Variables:        map[string]any{"Name": "Kai", "Amount": "1"},
		Channels: []notify.ChannelDest{
			{Channel: notify.ChannelInApp},
			{Channel: notify.ChannelEmail, Destination: "kai@example.test"},
		},
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("all-channels-opted-out should be a validation error, got %v", err)
	}
}

// TestIntegrationDeliveriesReceipts is the R5 receipts API: delivery status is
// queryable per notification, one receipt per channel, carrying the provider
// message id and last error.
func TestIntegrationDeliveriesReceipts(t *testing.T) {
	a := newHarness(t)
	id, err := a.send(t, notify.Message{
		TemplateKey:      "core.notify.welcome",
		RecipientPartyID: uuid.New(),
		Variables:        map[string]any{"Name": "Bo", "Amount": "9"},
		Channels: []notify.ChannelDest{
			{Channel: notify.ChannelInApp},
			{Channel: notify.ChannelEmail, Destination: "bo@example.test"},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	var receipts []notify.DeliveryReceipt
	if err := a.db.TxM.WithTenantRO(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		receipts, e = a.svc.Deliveries(ctx, db, id)
		return e
	}); err != nil {
		t.Fatalf("Deliveries: %v", err)
	}
	if len(receipts) != 2 {
		t.Fatalf("got %d receipts, want 2 (inapp + email)", len(receipts))
	}
	byChannel := make(map[notify.Channel]notify.DeliveryReceipt, 2)
	for _, r := range receipts {
		byChannel[r.Channel] = r
	}
	if _, ok := byChannel[notify.ChannelInApp]; !ok {
		t.Error("missing an inapp delivery receipt")
	}
	email, ok := byChannel[notify.ChannelEmail]
	if !ok {
		t.Fatal("missing an email delivery receipt")
	}
	if email.Destination != "bo@example.test" {
		t.Errorf("email destination = %q, want bo@example.test", email.Destination)
	}
	if email.Status != "queued" {
		t.Errorf("fresh email delivery status = %q, want queued", email.Status)
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

// TestSendPendingLegalImportanceWritesAuditEvent is the DATA-08 W0-T2
// positive regression: an ImportanceLegal delivery that sends successfully
// must produce a durable "notify.legal_delivery" outbox event carrying the
// provider's message id as receipt, written in the same transaction as the
// 'sent' status update.
func TestSendPendingLegalImportanceWritesAuditEvent(t *testing.T) {
	db := testkit.NewDB(t)
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:      "core.notify.legal",
		Vars:     []string{"Name"},
		Channels: []string{"email"},
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	fake := &notify.FakeSender{}
	ob := outbox.NewWriter(model.UUIDv7())
	svc := notify.New(reg, model.UUIDv7(), notify.WithOutbox(ob))
	svc.RegisterSender(notify.ChannelEmail, fake)

	tenant := testkit.CreateTenant(t, db).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tenant), actor)
	seedTemplate(t, db, nil, "core.notify.legal", "email", "en", "Legal notice", "Hi {{.Name}}")

	var id uuid.UUID
	if err := db.TxM.WithTenant(ctx, func(ctx context.Context, tx database.TenantDB) error {
		var e error
		id, e = svc.Send(ctx, tx, notify.Message{
			TemplateKey:      "core.notify.legal",
			RecipientPartyID: uuid.New(),
			Variables:        map[string]any{"Name": "Grace"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "grace@example.test"}},
			Importance:       notify.ImportanceLegal,
		})
		return e
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	n, err := svc.SendPending(context.Background(), db.PlatformTxM, tenant, time.Now())
	if err != nil {
		t.Fatalf("SendPending: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 sent, got %d", n)
	}

	var deliveryID uuid.UUID
	var providerMsgID string
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT id, COALESCE(provider_message_id, '') FROM notification_deliveries
		  WHERE notification_id = $1`, id,
	).Scan(&deliveryID, &providerMsgID); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	if providerMsgID == "" {
		t.Fatal("expected provider_message_id to be set")
	}

	var count int
	var payload []byte
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM events_outbox
		  WHERE event_type = 'notify.legal_delivery'
		    AND resource_id = $1`, deliveryID,
	).Scan(&count); err != nil {
		t.Fatalf("query events_outbox: %v", err)
	}
	if count != 1 {
		t.Fatalf("want exactly 1 notify.legal_delivery outbox event for delivery %s, got %d", deliveryID, count)
	}
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT payload FROM events_outbox
		  WHERE event_type = 'notify.legal_delivery' AND resource_id = $1`, deliveryID,
	).Scan(&payload); err != nil {
		t.Fatalf("read outbox payload: %v", err)
	}
	if !strings.Contains(string(payload), providerMsgID) {
		t.Fatalf("legal delivery audit payload must carry the provider message id %q, got %s", providerMsgID, payload)
	}
}

// TestSendPendingNonLegalImportanceWritesNoAuditEvent is the DATA-08 W0-T2
// negative regression: a successful delivery for a NON-legal importance must
// NOT write the "notify.legal_delivery" outbox event, even when the Service is
// wired with an outbox writer — proving the audit write is conditional on
// ImportanceLegal, not unconditional.
func TestSendPendingNonLegalImportanceWritesNoAuditEvent(t *testing.T) {
	db := testkit.NewDB(t)
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:      "core.notify.normal",
		Vars:     []string{"Name"},
		Channels: []string{"email"},
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	fake := &notify.FakeSender{}
	ob := outbox.NewWriter(model.UUIDv7())
	svc := notify.New(reg, model.UUIDv7(), notify.WithOutbox(ob))
	svc.RegisterSender(notify.ChannelEmail, fake)

	tenant := testkit.CreateTenant(t, db).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tenant), actor)
	seedTemplate(t, db, nil, "core.notify.normal", "email", "en", "Notice", "Hi {{.Name}}")

	var id uuid.UUID
	if err := db.TxM.WithTenant(ctx, func(ctx context.Context, tx database.TenantDB) error {
		var e error
		id, e = svc.Send(ctx, tx, notify.Message{
			TemplateKey:      "core.notify.normal",
			RecipientPartyID: uuid.New(),
			Variables:        map[string]any{"Name": "Hank"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "hank@example.test"}},
			// Importance intentionally left as default (normal), not legal.
		})
		return e
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	n, err := svc.SendPending(context.Background(), db.PlatformTxM, tenant, time.Now())
	if err != nil {
		t.Fatalf("SendPending: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 sent, got %d", n)
	}

	var deliveryID uuid.UUID
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT id FROM notification_deliveries WHERE notification_id = $1`, id,
	).Scan(&deliveryID); err != nil {
		t.Fatalf("read delivery: %v", err)
	}

	var count int
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM events_outbox
		  WHERE event_type = 'notify.legal_delivery' AND resource_id = $1`, deliveryID,
	).Scan(&count); err != nil {
		t.Fatalf("query events_outbox: %v", err)
	}
	if count != 0 {
		t.Fatalf("non-legal delivery must NOT write a notify.legal_delivery outbox event, got %d", count)
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

// TestSendPendingUnregisteredChannelFailsLoudly is the CA-15 regression: a
// queued delivery for a channel with NO registered sender must be recorded as a
// terminal failure ('dead', with a clear last_error) — never silently marked
// 'sent' by a no-op fallback sender (the prior silent-success hole).
func TestSendPendingUnregisteredChannelFailsLoudly(t *testing.T) {
	db := testkit.NewDB(t)
	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:      "core.notify.alert",
		Vars:     []string{"Name"},
		Channels: []string{"sms"},
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	// Deliberately register NO sender for the sms channel.
	svc := notify.New(reg, model.UUIDv7())

	tenant := testkit.CreateTenant(t, db).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tenant), actor)
	seedTemplate(t, db, nil, "core.notify.alert", "sms", "en", "", "Hi {{.Name}}")

	var id uuid.UUID
	if err := db.TxM.WithTenant(ctx, func(ctx context.Context, tx database.TenantDB) error {
		var e error
		id, e = svc.Send(ctx, tx, notify.Message{
			TemplateKey:      "core.notify.alert",
			RecipientPartyID: uuid.New(),
			Variables:        map[string]any{"Name": "Frank"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelSMS, Destination: "+15551234567"}},
		})
		return e
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	n, err := svc.SendPending(context.Background(), db.PlatformTxM, tenant, time.Now())
	if err != nil {
		t.Fatalf("SendPending: %v", err)
	}
	if n != 0 {
		t.Fatalf("unregistered-channel delivery must NOT count as sent, got %d", n)
	}

	var status, lastErr string
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT status, COALESCE(last_error,'') FROM notification_deliveries WHERE notification_id = $1`, id,
	).Scan(&status, &lastErr); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	if status != "dead" {
		t.Fatalf("unregistered channel must go terminal 'dead', got %q", status)
	}
	if !strings.Contains(lastErr, "no sender registered") {
		t.Fatalf("last_error should explain the misconfiguration, got %q", lastErr)
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
