package notify_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/notify"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/testkit"
)

// fakeTracer records the carriers passed to Extract and the span names started,
// and injects a fixed traceparent, so a test can prove notify carries trace
// context across the async send boundary (O1/CA-9). Mirrors the outbox test.
type fakeTracer struct {
	inject    string
	mu        sync.Mutex
	extracted []string
	spans     []string
}

func (f *fakeTracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	f.mu.Lock()
	f.spans = append(f.spans, name)
	f.mu.Unlock()
	return ctx, fakeSpan{}
}
func (f *fakeTracer) Inject(context.Context) string { return f.inject }
func (f *fakeTracer) Extract(ctx context.Context, carrier string) context.Context {
	f.mu.Lock()
	f.extracted = append(f.extracted, carrier)
	f.mu.Unlock()
	return ctx
}

type fakeSpan struct{}

func (fakeSpan) End()                {}
func (fakeSpan) SetAttr(_, _ string) {}
func (fakeSpan) RecordError(error)   {}

// TestIntegrationNotifyTracePropagation is the O1/CA-9 regression for the notify
// framework: a queued delivery captures the sender's trace context at Send, and
// SendPending extracts it and delivers under a child span — so an async delivery
// continues the originating request's trace. Mirrors the outbox test.
func TestIntegrationNotifyTracePropagation(t *testing.T) {
	db := testkit.NewDB(t)
	const carrier = "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	tr := &fakeTracer{inject: carrier}

	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:      "core.notify.welcome",
		Vars:     []string{"Name"},
		Channels: []string{"email"},
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}

	fake := &notify.FakeSender{}
	svc := notify.New(reg, model.UUIDv7(), notify.WithTracer(tr))
	svc.RegisterSender(notify.ChannelEmail, fake)

	tenant := testkit.CreateTenant(t, db).ID
	ctx := database.WithActorID(testkit.TenantCtx(tenant), uuid.New())
	seedTemplate(t, db, nil, "core.notify.welcome", "email", "en", "Welcome {{.Name}}", "Hi {{.Name}}")

	party := uuid.New()
	var notifID uuid.UUID
	if err := db.TxM.WithTenant(ctx, func(ctx context.Context, tdb database.TenantDB) error {
		var e error
		notifID, e = svc.Send(ctx, tdb, notify.Message{
			TemplateKey:      "core.notify.welcome",
			RecipientPartyID: party,
			Variables:        map[string]any{"Name": "Dana"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "dana@example.test"}},
		})
		return e
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// The delivery row stored the injected trace context.
	var stored string
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT coalesce(trace_context,'') FROM notification_deliveries WHERE notification_id = $1`, notifID).Scan(&stored); err != nil {
		t.Fatalf("read trace_context: %v", err)
	}
	if stored != carrier {
		t.Fatalf("stored trace_context = %q, want the injected carrier %q", stored, carrier)
	}

	// SendPending extracts that carrier and opens a child span when delivering.
	n, err := svc.SendPending(context.Background(), db.PlatformTxM, tenant, time.Now())
	if err != nil {
		t.Fatalf("SendPending: %v", err)
	}
	if n != 1 {
		t.Fatalf("SendPending sent %d, want 1", n)
	}
	if fake.Count() != 1 {
		t.Fatalf("FakeSender calls = %d, want 1", fake.Count())
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()
	found := false
	for _, c := range tr.extracted {
		if c == carrier {
			found = true
		}
	}
	if !found {
		t.Fatalf("SendPending must Extract the stored trace context to continue the trace; extracted=%v", tr.extracted)
	}
	spanFound := false
	for _, name := range tr.spans {
		if name == "notify.send email" {
			spanFound = true
		}
	}
	if !spanFound {
		t.Fatalf("SendPending must start a child span %q; spans=%v", "notify.send email", tr.spans)
	}
}

// TestIntegrationNotifyNoTracerNoContext proves backward compatibility: without a
// tracer, Send stores NULL trace_context (no behavior change) and SendPending
// still delivers.
func TestIntegrationNotifyNoTracerNoContext(t *testing.T) {
	db := testkit.NewDB(t)

	reg := notify.NewRegistry()
	reg.Register("core", notify.TemplateSpec{
		Key:      "core.notify.welcome",
		Vars:     []string{"Name"},
		Channels: []string{"email"},
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}

	fake := &notify.FakeSender{}
	svc := notify.New(reg, model.UUIDv7()) // no tracer
	svc.RegisterSender(notify.ChannelEmail, fake)

	tenant := testkit.CreateTenant(t, db).ID
	ctx := database.WithActorID(testkit.TenantCtx(tenant), uuid.New())
	seedTemplate(t, db, nil, "core.notify.welcome", "email", "en", "Welcome {{.Name}}", "Hi {{.Name}}")

	var notifID uuid.UUID
	if err := db.TxM.WithTenant(ctx, func(ctx context.Context, tdb database.TenantDB) error {
		var e error
		notifID, e = svc.Send(ctx, tdb, notify.Message{
			TemplateKey:      "core.notify.welcome",
			RecipientPartyID: uuid.New(),
			Variables:        map[string]any{"Name": "Eve"},
			Channels:         []notify.ChannelDest{{Channel: notify.ChannelEmail, Destination: "eve@example.test"}},
		})
		return e
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	var stored *string
	if err := db.Admin.QueryRow(context.Background(),
		`SELECT trace_context FROM notification_deliveries WHERE notification_id = $1`, notifID).Scan(&stored); err != nil {
		t.Fatalf("read trace_context: %v", err)
	}
	if stored != nil {
		t.Fatalf("trace_context = %q, want NULL with no tracer", *stored)
	}

	if n, err := svc.SendPending(context.Background(), db.PlatformTxM, tenant, time.Now()); err != nil || n != 1 {
		t.Fatalf("SendPending = (%d, %v), want (1, nil)", n, err)
	}
}
