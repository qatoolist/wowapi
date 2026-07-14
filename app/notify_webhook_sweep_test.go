package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/foundation/webhook"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/secrets"
	"github.com/qatoolist/wowapi/testkit"
)

// staticSecrets resolves any ref to a fixed signing secret (webhook HMAC key).
type staticSecrets struct{ val string }

func (s staticSecrets) Resolve(_ context.Context, _ secrets.Ref) (string, error) {
	return s.val, nil
}

// TestStartWorkerRunsWebhookRetrySweep is the H3 wiring regression: the scheduled
// kernel.webhook.retry task registered by registerMaintenance enumerates active
// tenants on the leader-safe scheduler and, per tenant, re-drives a previously-
// failed outbound webhook delivery to completion — proving the async fan-out is
// SHIPPED (not built-but-not-wired) and that the delivery tenant is bound from
// the enumeration (H2-safe by construction), never a caller-supplied param.
func TestStartWorkerRunsWebhookRetrySweep(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	const secretRef = "secretref://test/key"
	sender := &webhook.FakeSender{StatusCode: 200}
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool:          h.Runtime,
		Platform:      h.Platform,
		Tx:            h.TxM,
		Secrets:       staticSecrets{val: "super-secret-key"},
		WebhookSender: sender,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	// Active tenant + an outbound endpoint + a failed, due delivery row. The retry
	// sweep must claim it (FOR UPDATE SKIP LOCKED), sign it with THIS tenant's
	// secret, POST it via the fake sender, and mark it delivered.
	tn := testkit.CreateTenant(t, h)
	epID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO webhook_endpoints
		    (id, tenant_id, direction, url, secret_ref, signature_scheme,
		     subscribed_events, status, created_by)
		 VALUES ($1, $2, 'outbound', $3, $4, 'hmac-sha256',
		         '{order.created}'::text[], 'active', $5)`,
		epID, tn.ID, "https://example.test/hook", secretRef, uuid.Nil); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	evID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO webhook_events
		    (id, tenant_id, endpoint_id, direction, external_event_id, event_type,
		     payload, received_at, delivery_status, attempts)
		 VALUES ($1, $2, $3, 'outbound', $4, 'order.created', $5, now(), 'failed', 1)`,
		uuid.New(), tn.ID, epID, evID.String(),
		`{"id":"`+evID.String()+`","type":"order.created"}`); err != nil {
		t.Fatalf("seed failed delivery: %v", err)
	}

	booted, err := app.New().Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:            80 * time.Millisecond,
			JobPoll:              80 * time.Millisecond,
			SchedulerPoll:        40 * time.Millisecond,
			SLAInterval:          time.Hour,
			IdempotencyInterval:  time.Hour,
			DLQDepthInterval:     time.Hour,
			AuditAnchorInterval:  time.Hour,
			NotifySendInterval:   time.Hour,
			WebhookRetryInterval: 100 * time.Millisecond,
			ShutdownDrain:        3 * time.Second,
		})
	}()

	// Wait until the retry sweep has driven the delivery to 'delivered'.
	deadline := time.After(12 * time.Second)
	var status string
	for status != "delivered" {
		select {
		case <-deadline:
			t.Fatalf("webhook retry sweep never delivered (last status %q)", status)
		case err := <-done:
			t.Fatalf("StartWorker returned early: %v", err)
		case <-time.After(50 * time.Millisecond):
			if err := h.Admin.QueryRow(context.Background(),
				`SELECT delivery_status FROM webhook_events WHERE endpoint_id = $1`, epID).Scan(&status); err != nil {
				t.Fatalf("query delivery row: %v", err)
			}
		}
	}

	// Exactly one POST, carrying a signature (tenant's secret) and the event id.
	if len(sender.Calls) < 1 {
		t.Fatalf("want >=1 POST from retry sweep, got %d", len(sender.Calls))
	}
	if got := sender.Calls[0].Headers["X-Event-Id"]; got != evID.String() {
		t.Fatalf("X-Event-Id = %q, want %q", got, evID.String())
	}
	if sender.Calls[0].Headers["X-Signature"] == "" {
		t.Fatal("delivery was not signed with the tenant's secret")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartWorker returned error on shutdown: %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("StartWorker did not drain within the shutdown window")
	}
}

// TestStartWorkerRunsNotifySendSweep is the H3 wiring regression for the notify
// side: the scheduled kernel.notify.send_pending task enumerates active tenants
// on the leader-safe scheduler and drives a queued notification delivery to
// 'sent' via the registered channel sender — proving the async fan-out is
// shipped and tenant-bound from the enumeration, not a caller param.
func TestStartWorkerRunsNotifySendSweep(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	// Register a template + a fake email sender on the kernel's notify service.
	k.NotifyTemplates.Register("core", notify.TemplateSpec{
		Key:      "core.notify.welcome",
		Vars:     []string{"Name"},
		Channels: []string{"email"},
	})
	if err := k.NotifyTemplates.Err(); err != nil {
		t.Fatalf("register template: %v", err)
	}
	fake := &notify.FakeSender{}
	k.Notify.RegisterSender(notify.ChannelEmail, fake)

	tn := testkit.CreateTenant(t, h)
	// Platform-default email template (tenant_id NULL).
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO notification_templates
		    (id, tenant_id, key, channel, locale, subject, body, created_by)
		 VALUES ($1, NULL, 'core.notify.welcome', 'email', 'en', 'Welcome {{.Name}}', 'Hi {{.Name}}', $2)`,
		uuid.New(), uuid.Nil); err != nil {
		t.Fatalf("seed template: %v", err)
	}

	// Seed a queued delivery through the real Send path (in the tenant's tx).
	party := uuid.New()
	actorCtx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())
	var notifID uuid.UUID
	if err := h.TxM.WithTenant(actorCtx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		notifID, e = k.Notify.Send(ctx, db, notify.Message{
			TemplateKey:      "core.notify.welcome",
			RecipientPartyID: party,
			Variables:        map[string]any{"Name": "Carlos"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "carlos@example.test"}},
		})
		return e
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	booted, err := app.New().Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:            80 * time.Millisecond,
			JobPoll:              80 * time.Millisecond,
			SchedulerPoll:        40 * time.Millisecond,
			SLAInterval:          time.Hour,
			IdempotencyInterval:  time.Hour,
			DLQDepthInterval:     time.Hour,
			AuditAnchorInterval:  time.Hour,
			NotifySendInterval:   100 * time.Millisecond,
			WebhookRetryInterval: time.Hour,
			ShutdownDrain:        3 * time.Second,
		})
	}()

	deadline := time.After(12 * time.Second)
	var status string
	for status != "sent" {
		select {
		case <-deadline:
			t.Fatalf("notify send sweep never sent (last status %q)", status)
		case err := <-done:
			t.Fatalf("StartWorker returned early: %v", err)
		case <-time.After(50 * time.Millisecond):
			if err := h.Admin.QueryRow(context.Background(),
				`SELECT status FROM notification_deliveries WHERE notification_id = $1`, notifID).Scan(&status); err != nil {
				t.Fatalf("query delivery: %v", err)
			}
		}
	}
	if fake.Count() < 1 {
		t.Fatalf("want >=1 channel send, got %d", fake.Count())
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartWorker returned error on shutdown: %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("StartWorker did not drain within the shutdown window")
	}
}
