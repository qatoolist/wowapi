package webhook_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/webhook"
	"github.com/qatoolist/wowapi/testkit"
	"github.com/qatoolist/wowapi/testkit/fakes"
)

const (
	testProviderKey = "test-provider"
	testSecret      = "super-secret-key"
	testSecretRef   = "secretref://test/key"
)

// --- seed helpers (Admin pool → bypasses RLS) ---

func seedInboundEndpoint(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO webhook_endpoints
		    (id, tenant_id, direction, secret_ref, signature_scheme, status, created_by)
		 VALUES ($1, $2, 'inbound', $3, 'hmac-sha256', 'active', $4)`,
		id, tenantID, testSecretRef, uuid.Nil)
	if err != nil {
		t.Fatalf("seedInboundEndpoint: %v", err)
	}
	return id
}

// seedOutboundEndpoint inserts an outbound endpoint subscribed to "order.created".
// Uses a SQL literal for text[] to avoid pgx slice-encoding ambiguity.
func seedOutboundEndpoint(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID, url string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO webhook_endpoints
		    (id, tenant_id, direction, url, secret_ref, signature_scheme,
		     subscribed_events, status, created_by)
		 VALUES ($1, $2, 'outbound', $3, $4, 'hmac-sha256',
		         '{order.created}'::text[], 'active', $5)`,
		id, tenantID, url, testSecretRef, uuid.Nil)
	if err != nil {
		t.Fatalf("seedOutboundEndpoint: %v", err)
	}
	return id
}

// --- test service constructors ---

// newServiceWithClock wires a Service with an injectable clock and
// HMACVerifier registered under testProviderKey.
func newServiceWithClock(t *testing.T, sender *webhook.FakeSender, clk *fakes.Clock) *webhook.Service {
	t.Helper()
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	svc := webhook.NewWithClock(sender, resolver, model.UUIDv7(), clk.Now)
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})
	return svc
}

// newService wires a Service with a real (wall) clock.
func newService(t *testing.T, sender *webhook.FakeSender) *webhook.Service {
	t.Helper()
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	svc := webhook.New(sender, resolver, model.UUIDv7())
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})
	return svc
}

// --- signing helper ---

// testSign computes an inbound X-Signature header (HMAC over the body alone —
// the external-provider scheme HMACVerifier expects).
func testSign(body []byte) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// testSignOutbound computes the expected OUTBOUND X-Signature: HMAC over
// "<timestamp>.<body>" (SEC-52), matching signPayload in the service.
func testSignOutbound(ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// --- count / query helpers ---

func countEvents(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID) int {
	t.Helper()
	var n int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM webhook_events WHERE tenant_id = $1`, tenantID).Scan(&n); err != nil {
		t.Fatalf("countEvents: %v", err)
	}
	return n
}

func eventStatus(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID) string {
	t.Helper()
	var s string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE tenant_id = $1 LIMIT 1`, tenantID).Scan(&s); err != nil {
		t.Fatalf("eventStatus: %v", err)
	}
	return s
}

func eventSigOk(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID) *bool {
	t.Helper()
	var ok *bool
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT signature_ok FROM webhook_events WHERE tenant_id = $1 LIMIT 1`, tenantID).Scan(&ok); err != nil {
		t.Fatalf("eventSigOk: %v", err)
	}
	return ok
}

// =============================================================================
// Inbound tests
// =============================================================================

// TestIntegrationHandleInbound_SignatureSuccess proves a valid HMAC-SHA256
// signature persists a pending row with signature_ok=true.
func TestIntegrationHandleInbound_SignatureSuccess(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"event":"order.created"}`)
	in := webhook.InboundIn{
		EndpointID:      epID,
		ProviderKey:     testProviderKey,
		RawBody:         body,
		Headers:         map[string]string{"X-Signature": testSign(body)},
		ExternalEventID: "ext-001",
		EventType:       "order.created",
		Timestamp:       time.Now(),
	}
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, in)
	}); err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}
	if n := countEvents(t, h, tn.ID); n != 1 {
		t.Fatalf("want 1 event row, got %d", n)
	}
	if s := eventStatus(t, h, tn.ID); s != "pending" {
		t.Fatalf("want pending, got %s", s)
	}
	ok := eventSigOk(t, h, tn.ID)
	if ok == nil || !*ok {
		t.Fatalf("want signature_ok=true, got %v", ok)
	}
}

// TestIntegrationHandleInbound_BadSignature proves a wrong signature returns
// KindUnauthenticated and records a signature_ok=false audit row.
func TestIntegrationHandleInbound_BadSignature(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"event":"order.created"}`)
	in := webhook.InboundIn{
		EndpointID:      epID,
		ProviderKey:     testProviderKey,
		RawBody:         body,
		Headers:         map[string]string{"X-Signature": "sha256=badhex"},
		ExternalEventID: "ext-badsig",
		EventType:       "order.created",
		Timestamp:       time.Now(),
	}

	var sigErr error
	// Commit despite error so the best-effort audit row persists.
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		sigErr = svc.HandleInbound(ctx, db, in)
		return nil
	}); cerr != nil {
		t.Fatalf("tx commit: %v", cerr)
	}
	if kerr.KindOf(sigErr) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated, got kind=%v err=%v", kerr.KindOf(sigErr), sigErr)
	}
	if n := countEvents(t, h, tn.ID); n != 1 {
		t.Fatalf("want 1 audit row, got %d", n)
	}
	ok := eventSigOk(t, h, tn.ID)
	if ok == nil || *ok {
		t.Fatalf("want signature_ok=false, got %v", ok)
	}
}

// TestIntegrationHandleInbound_Replay proves a duplicate external_event_id
// is idempotent: KindConflict returned, exactly one row stored.
func TestIntegrationHandleInbound_Replay(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"event":"order.created"}`)
	makeIn := func() webhook.InboundIn {
		return webhook.InboundIn{
			EndpointID:      epID,
			ProviderKey:     testProviderKey,
			RawBody:         body,
			Headers:         map[string]string{"X-Signature": testSign(body)},
			ExternalEventID: "ext-replay",
			EventType:       "order.created",
			Timestamp:       time.Now(),
		}
	}

	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, makeIn())
	}); err != nil {
		t.Fatalf("first call: %v", err)
	}

	var replayErr error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		replayErr = svc.HandleInbound(ctx, db, makeIn())
		return nil
	}); cerr != nil {
		t.Fatalf("tx commit: %v", cerr)
	}
	if kerr.KindOf(replayErr) != kerr.KindConflict {
		t.Fatalf("want KindConflict on replay, got kind=%v err=%v", kerr.KindOf(replayErr), replayErr)
	}
	if n := countEvents(t, h, tn.ID); n != 1 {
		t.Fatalf("want 1 row after replay, got %d", n)
	}
}

// TestIntegrationHandleInbound_TimestampOutOfWindow proves events with a
// timestamp outside ±5 m are rejected.
func TestIntegrationHandleInbound_TimestampOutOfWindow(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"event":"old"}`)
	in := webhook.InboundIn{
		EndpointID:      epID,
		ProviderKey:     testProviderKey,
		RawBody:         body,
		Headers:         map[string]string{"X-Signature": testSign(body)},
		ExternalEventID: "ext-old",
		EventType:       "order.created",
		Timestamp:       time.Now().Add(-10 * time.Minute), // 10 m ago
	}

	var tsErr error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		tsErr = svc.HandleInbound(ctx, db, in)
		return nil
	}); cerr != nil {
		t.Fatalf("tx commit: %v", cerr)
	}
	if kerr.KindOf(tsErr) != kerr.KindValidation {
		t.Fatalf("want KindValidation, got kind=%v err=%v", kerr.KindOf(tsErr), tsErr)
	}
	if n := countEvents(t, h, tn.ID); n != 0 {
		t.Fatalf("want 0 rows after out-of-window rejection, got %d", n)
	}
}

// =============================================================================
// Inbound processing tests
// =============================================================================

// TestIntegrationProcessInbound_Success proves ProcessInbound runs the handler
// and advances delivery_status to processed.
func TestIntegrationProcessInbound_Success(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	var handled []string
	svc.RegisterHandler("order.created", func(_ context.Context, _ database.TenantDB, e webhook.Event) error {
		handled = append(handled, e.EventType)
		return nil
	})

	body := []byte(`{"event":"order.created"}`)
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: testProviderKey,
			RawBody:         body,
			Headers:         map[string]string{"X-Signature": testSign(body)},
			ExternalEventID: "ext-proc-ok",
			EventType:       "order.created",
			Timestamp:       time.Now(),
		})
	}); err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}
	if err := svc.ProcessInbound(context.Background(), h.PlatformTxM, tn.ID, time.Now()); err != nil {
		t.Fatalf("ProcessInbound: %v", err)
	}
	if len(handled) != 1 || handled[0] != "order.created" {
		t.Fatalf("handler not called or wrong type: %v", handled)
	}
	if s := eventStatus(t, h, tn.ID); s != "processed" {
		t.Fatalf("want processed, got %s", s)
	}
	_ = epID // used via seedInboundEndpoint
}

// TestIntegrationProcessInbound_HandlerErrorDeadLetters proves that a handler
// that always fails reaches dead after MaxAttempts.
func TestIntegrationProcessInbound_HandlerErrorDeadLetters(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})
	svc.RegisterHandler("broken.event", func(_ context.Context, _ database.TenantDB, _ webhook.Event) error {
		return kerr.E(kerr.KindInternal, "internal", "handler always fails")
	})

	body := []byte(`{}`)
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: testProviderKey,
			RawBody:         body,
			Headers:         map[string]string{"X-Signature": testSign(body)},
			ExternalEventID: "ext-dlq",
			EventType:       "broken.event",
			Timestamp:       time.Now(),
		})
	}); err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}

	// Each pass advances processNow by 10 m so next_attempt_at is always past.
	base := time.Now()
	for i := 0; i <= webhook.MaxAttempts; i++ {
		processNow := base.Add(time.Duration(i+1) * 10 * time.Minute)
		if err := svc.ProcessInbound(context.Background(), h.PlatformTxM, tn.ID, processNow); err != nil {
			t.Fatalf("ProcessInbound[%d]: %v", i, err)
		}
	}
	if s := eventStatus(t, h, tn.ID); s != "dead" {
		t.Fatalf("want dead after DLQ ceiling, got %s", s)
	}
}

// =============================================================================
// Outbound dispatch tests
// =============================================================================

// TestIntegrationDispatchOutbound_MatchingEndpoint proves DispatchOutbound
// signs and POSTs to a matching endpoint and marks the delivery row delivered.
func TestIntegrationDispatchOutbound_MatchingEndpoint(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/hook")
	sender := &webhook.FakeSender{StatusCode: 200}
	svc := newService(t, sender)

	ev := outbox.Event{
		ID:      uuid.New(),
		Type:    "order.created",
		Payload: json.RawMessage(`{"order_id":"abc"}`),
	}
	now := time.Now()
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, now); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	if len(sender.Calls) != 1 {
		t.Fatalf("want 1 POST, got %d", len(sender.Calls))
	}
	call := sender.Calls[0]
	if call.URL != "https://example.test/hook" {
		t.Fatalf("wrong URL: %s", call.URL)
	}
	if call.Headers["X-Signature"] == "" || call.Headers["X-Timestamp"] == "" {
		t.Fatalf("missing signature headers: %v", call.Headers)
	}
	if call.Headers["X-Event-Id"] != ev.ID.String() {
		t.Fatalf("X-Event-Id = %q, want %q", call.Headers["X-Event-Id"], ev.ID.String())
	}

	// Verify X-Signature is the HMAC-SHA256 of "<X-Timestamp>.<body>" (SEC-52).
	wantSig := testSignOutbound(call.Headers["X-Timestamp"], call.Body)
	if call.Headers["X-Signature"] != wantSig {
		t.Fatalf("X-Signature mismatch\n got  %s\n want %s", call.Headers["X-Signature"], wantSig)
	}

	// Delivery row must be 'delivered'.
	var status string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE endpoint_id = $1`, epID).Scan(&status); err != nil {
		t.Fatalf("query delivery row: %v", err)
	}
	if status != "delivered" {
		t.Fatalf("want delivered, got %s", status)
	}
}

// TestIntegrationDispatchOutbound_NonMatchingEventType proves that an event
// type not in subscribed_events results in zero POSTs and no delivery rows.
func TestIntegrationDispatchOutbound_NonMatchingEventType(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/hook")
	sender := &webhook.FakeSender{StatusCode: 200}
	svc := newService(t, sender)

	ev := outbox.Event{
		ID:      uuid.New(),
		Type:    "invoice.paid", // not in subscribed_events (order.created only)
		Payload: json.RawMessage(`{}`),
	}
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	if len(sender.Calls) != 0 {
		t.Fatalf("want 0 POSTs for non-matching type, got %d", len(sender.Calls))
	}
	if n := countEvents(t, h, tn.ID); n != 0 {
		t.Fatalf("want 0 delivery rows, got %d", n)
	}
}

// =============================================================================
// Circuit breaker tests (use injectable clock via fakes.Clock)
// =============================================================================

// TestIntegrationBreakerOpensAfterNFailures proves the circuit breaker opens after
// BreakerFailureThreshold consecutive delivery failures and blocks further POSTs.
func TestIntegrationBreakerOpensAfterNFailures(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/cb")

	clk := fakes.NewClock(time.Now())
	sender := &webhook.FakeSender{StatusCode: 500}
	svc := newServiceWithClock(t, sender, clk)

	// Drive BreakerFailureThreshold failures (each with a distinct event ID).
	for i := 0; i < webhook.BreakerFailureThreshold; i++ {
		ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
		if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
			t.Fatalf("dispatch[%d]: %v", i, err)
		}
	}
	if got := len(sender.Calls); got != webhook.BreakerFailureThreshold {
		t.Fatalf("want %d POSTs before open, got %d", webhook.BreakerFailureThreshold, got)
	}

	// Next attempt — breaker is open, no POST.
	before := len(sender.Calls)
	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
		t.Fatalf("dispatch after open: %v", err)
	}
	if len(sender.Calls) != before {
		t.Fatalf("breaker should block POST while open; calls went %d → %d", before, len(sender.Calls))
	}
}

// TestIntegrationBreakerHalfOpenAfterCooldown proves the breaker allows exactly one
// probe after BreakerCooldown elapses, and closes on a successful probe.
func TestIntegrationBreakerHalfOpenAfterCooldown(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/cb2")

	clk := fakes.NewClock(time.Now())
	sender := &webhook.FakeSender{StatusCode: 500}
	svc := newServiceWithClock(t, sender, clk)

	// Open the breaker.
	for i := 0; i < webhook.BreakerFailureThreshold; i++ {
		ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
		_ = svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now())
	}

	// Endpoint must be marked degraded.
	var epStatus string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM webhook_endpoints WHERE id = $1`, epID).Scan(&epStatus); err != nil {
		t.Fatalf("query endpoint status: %v", err)
	}
	if epStatus != "degraded" {
		t.Fatalf("want endpoint status=degraded after breaker opens, got %s", epStatus)
	}

	// Advance past cooldown → half-open probe allowed.
	clk.Advance(webhook.BreakerCooldown + time.Second)
	sender.StatusCode = 200 // probe succeeds

	probe := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
	callsBefore := len(sender.Calls)
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, probe, clk.Now()); err != nil {
		t.Fatalf("probe dispatch: %v", err)
	}
	if len(sender.Calls) == callsBefore {
		t.Fatal("want one probe POST after cooldown, got none")
	}

	// ARCH-72: a successful probe must return the endpoint to 'active'.
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM webhook_endpoints WHERE id = $1`, epID).Scan(&epStatus); err != nil {
		t.Fatalf("re-query endpoint status: %v", err)
	}
	if epStatus != "active" {
		t.Fatalf("want endpoint status back to active after recovery, got %s", epStatus)
	}

	// Breaker is now closed — next attempt goes through without waiting.
	follow := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
	callsBefore = len(sender.Calls)
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, follow, clk.Now()); err != nil {
		t.Fatalf("post-probe dispatch: %v", err)
	}
	if len(sender.Calls) == callsBefore {
		t.Fatal("want POST after breaker closes, got none")
	}
}

// =============================================================================
// Tenant isolation test
// =============================================================================

// TestIntegrationTenantIsolation proves webhook_events rows for tenant A are
// invisible to tenant B through the RLS-enforced runtime pool.
func TestIntegrationTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tnA := testkit.CreateTenant(t, h)
	tnB := testkit.CreateTenant(t, h)
	epA := seedInboundEndpoint(t, h, tnA.ID)

	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"x":1}`)
	if err := h.TxM.WithTenant(testkit.TenantCtx(tnA.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID:      epA,
			ProviderKey:     testProviderKey,
			RawBody:         body,
			Headers:         map[string]string{"X-Signature": testSign(body)},
			ExternalEventID: "ext-iso",
			EventType:       "x.event",
			Timestamp:       time.Now(),
		})
	}); err != nil {
		t.Fatalf("HandleInbound for A: %v", err)
	}

	// Tenant B sees zero events through the runtime (RLS).
	var nB int
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tnB.ID), func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx, `SELECT count(*) FROM webhook_events`).Scan(&nB)
	}); err != nil {
		t.Fatalf("tenant B count: %v", err)
	}
	if nB != 0 {
		t.Fatalf("tenant B sees %d of tenant A's events (RLS leak)", nB)
	}
}

// =============================================================================
// Regression tests (Phase 9 review findings)
// =============================================================================

// TestIntegrationRetryOutbound_RedeliversFailed proves ARCH-70 is fixed: a
// 'failed' outbound delivery is re-driven by RetryOutbound once its backoff has
// elapsed and reaches 'delivered'. Without RetryOutbound the failed row would
// never be retried (the outbox relay already marked its source event dispatched).
func TestIntegrationRetryOutbound_RedeliversFailed(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/retry")

	clk := fakes.NewClock(time.Now())
	sender := &webhook.FakeSender{StatusCode: 500} // first attempt fails
	svc := newServiceWithClock(t, sender, clk)

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{"n":1}`)}
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	var status string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE endpoint_id = $1`, epID).Scan(&status); err != nil {
		t.Fatalf("query row: %v", err)
	}
	if status != "failed" {
		t.Fatalf("want failed after first attempt, got %s", status)
	}
	callsAfterFirst := len(sender.Calls)

	// Recover the upstream and advance past the backoff window.
	sender.StatusCode = 200
	clk.Advance(time.Hour)
	if err := svc.RetryOutbound(context.Background(), h.PlatformTxM, tn.ID, clk.Now()); err != nil {
		t.Fatalf("RetryOutbound: %v", err)
	}
	if len(sender.Calls) != callsAfterFirst+1 {
		t.Fatalf("want one retry POST, calls went %d → %d", callsAfterFirst, len(sender.Calls))
	}
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE endpoint_id = $1`, epID).Scan(&status); err != nil {
		t.Fatalf("re-query row: %v", err)
	}
	if status != "delivered" {
		t.Fatalf("want delivered after RetryOutbound, got %s", status)
	}
}

// TestIntegrationHandleInbound_IdlessDedup proves SEC-49/ARCH-74 is fixed: two
// id-less inbound calls with the same body dedup to a single row (the dedup id
// is synthesized from the body).
func TestIntegrationHandleInbound_IdlessDedup(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"event":"order.created","n":7}`)
	makeIn := func() webhook.InboundIn {
		return webhook.InboundIn{
			EndpointID:  epID,
			ProviderKey: testProviderKey,
			RawBody:     body,
			Headers:     map[string]string{"X-Signature": testSign(body)},
			// no ExternalEventID → synthesized from the body
			EventType: "order.created",
			Timestamp: time.Now(),
		}
	}

	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, makeIn())
	}); err != nil {
		t.Fatalf("first id-less call: %v", err)
	}
	var replayErr error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		replayErr = svc.HandleInbound(ctx, db, makeIn())
		return nil
	}); cerr != nil {
		t.Fatalf("tx commit: %v", cerr)
	}
	if kerr.KindOf(replayErr) != kerr.KindConflict {
		t.Fatalf("want KindConflict on id-less replay, got kind=%v err=%v", kerr.KindOf(replayErr), replayErr)
	}
	if n := countEvents(t, h, tn.ID); n != 1 {
		t.Fatalf("want 1 row for two id-less calls with the same body, got %d", n)
	}
}

// TestIntegrationHandleInbound_FailedSigDoesNotBlockValid proves SEC-50 is
// fixed: a spoofed unsigned request carrying a legitimate event's id must NOT
// occupy that id's dedup slot, so the real signed event still lands.
func TestIntegrationHandleInbound_FailedSigDoesNotBlockValid(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{})

	body := []byte(`{"event":"order.created"}`)

	// Spoofed request with a bad signature pre-claiming id 'evt-1'.
	badIn := webhook.InboundIn{
		EndpointID:      epID,
		ProviderKey:     testProviderKey,
		RawBody:         body,
		Headers:         map[string]string{"X-Signature": "sha256=deadbeef"},
		ExternalEventID: "evt-1",
		EventType:       "order.created",
		Timestamp:       time.Now(),
	}
	var badErr error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		badErr = svc.HandleInbound(ctx, db, badIn)
		return nil // commit so the (NULL-id) audit row persists
	}); cerr != nil {
		t.Fatalf("tx commit (bad): %v", cerr)
	}
	if kerr.KindOf(badErr) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated for spoofed request, got kind=%v", kerr.KindOf(badErr))
	}

	// The legitimate signed event with the SAME id must still succeed.
	goodIn := webhook.InboundIn{
		EndpointID:      epID,
		ProviderKey:     testProviderKey,
		RawBody:         body,
		Headers:         map[string]string{"X-Signature": testSign(body)},
		ExternalEventID: "evt-1",
		EventType:       "order.created",
		Timestamp:       time.Now(),
	}
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, goodIn)
	}); err != nil {
		t.Fatalf("valid event blocked by prior failed-sig row: %v", err)
	}

	var n int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM webhook_events
		  WHERE tenant_id = $1 AND external_event_id = 'evt-1'
		    AND delivery_status = 'pending' AND signature_ok = true`,
		tn.ID).Scan(&n); err != nil {
		t.Fatalf("query valid row: %v", err)
	}
	if n != 1 {
		t.Fatalf("want exactly 1 valid pending evt-1 row, got %d", n)
	}
}

// TestIntegrationOutboundSignatureCoversTimestamp proves SEC-52 is fixed: the outbound
// X-Signature authenticates "<timestamp>.<body>", not the body alone, so a
// forged X-Timestamp invalidates the signature.
func TestIntegrationOutboundSignatureCoversTimestamp(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/ts")

	sender := &webhook.FakeSender{StatusCode: 200}
	svc := newService(t, sender)

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{"k":"v"}`)}
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	call := sender.Calls[0]
	ts := call.Headers["X-Timestamp"]
	gotSig := call.Headers["X-Signature"]

	// Matches HMAC over "<ts>.<body>".
	if gotSig != testSignOutbound(ts, call.Body) {
		t.Fatalf("signature does not match HMAC over <ts>.<body>")
	}
	// Does NOT match HMAC over the body alone → timestamp is authenticated.
	bodyOnly := "sha256=" + func() string {
		mac := hmac.New(sha256.New, []byte(testSecret))
		mac.Write(call.Body)
		return hex.EncodeToString(mac.Sum(nil))
	}()
	if gotSig == bodyOnly {
		t.Fatal("signature covers only the body — X-Timestamp not authenticated (SEC-52)")
	}
	// A forged timestamp yields a different signature.
	if testSignOutbound(ts+"9", call.Body) == gotSig {
		t.Fatal("signature unchanged when timestamp is altered")
	}
}
