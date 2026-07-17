package webhook_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/foundation/webhook"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/httpclient"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/testkit"
	"github.com/qatoolist/wowapi/v2/testkit/fakes"
)

// --- extra seed helper: inbound endpoint with a caller-chosen status ---

func seedInboundEndpointStatus(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID, status string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO webhook_endpoints
		    (id, tenant_id, direction, secret_ref, signature_scheme, status, created_by)
		 VALUES ($1, $2, 'inbound', $3, 'hmac-sha256', $4, $5)`,
		id, tenantID, testSecretRef, status, uuid.Nil)
	if err != nil {
		t.Fatalf("seedInboundEndpointStatus: %v", err)
	}
	return id
}

// failingResolver is a SecretResolver that always errors — used to cover the
// secret-resolution error branches in HandleInbound and deliverToEndpoint.
type failingResolver struct{}

func (failingResolver) Resolve(_ context.Context, _ string) (string, error) {
	return "", errors.New("secret store unavailable")
}

// =============================================================================
// HTTPSender (real net/http Sender) via httptest
// =============================================================================

// testAllowlistFor builds an httpclient.Config allowlisting rawURL's host —
// the SSRF guard is on by default (backlog B2), so any test that dials a
// loopback httptest server for reasons OTHER than SSRF policy itself must
// opt that specific target in. hostOnly is defined in sender_test.go.
func testAllowlistFor(t *testing.T, rawURL string) httpclient.Config {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("testAllowlistFor: parse %q: %v", rawURL, err)
	}
	return httpclient.Config{AllowedHosts: []string{hostOnly(t, u.Host)}}
}

// TestIntegrationHTTPSender_PostSuccess proves the production Sender POSTs the
// body + headers to a live server and returns its status code.
func TestIntegrationHTTPSender_PostSuccess(t *testing.T) {
	var (
		gotMethod string
		gotSig    string
		gotBody   []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotSig = r.Header.Get("X-Signature")
		gotBody = make([]byte, r.ContentLength)
		_, _ = r.Body.Read(gotBody)
		w.WriteHeader(http.StatusAccepted) // 202
	}))
	defer srv.Close()

	// This test exercises POST mechanics (method/headers/body forwarding), not
	// SSRF policy (that's kernel/webhook/sender_test.go), so the loopback
	// httptest target must be allowlisted — the default sender now blocks it
	// dial-time (backlog B2).
	sender := webhook.NewHTTPSender(webhook.WithHTTPClientConfig(testAllowlistFor(t, srv.URL)))
	body := []byte(`{"hello":"world"}`)
	code, err := sender.Post(context.Background(), srv.URL, body,
		map[string]string{"X-Signature": "sha256=abc", "Content-Type": "application/json"})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if code != http.StatusAccepted {
		t.Fatalf("want 202, got %d", code)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("want POST, got %s", gotMethod)
	}
	if gotSig != "sha256=abc" {
		t.Fatalf("signature header not forwarded, got %q", gotSig)
	}
	if string(gotBody) != string(body) {
		t.Fatalf("body mismatch: got %q", gotBody)
	}
}

// TestIntegrationHTTPSender_PostConnectionError proves a transport failure is
// surfaced as an error (server closed → connection refused).
func TestIntegrationHTTPSender_PostConnectionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	clientCfg := testAllowlistFor(t, url) // capture the allowlist before Close
	srv.Close()                           // nothing is listening now

	sender := webhook.NewHTTPSender(webhook.WithHTTPClientConfig(clientCfg))
	code, err := sender.Post(context.Background(), url, []byte(`{}`), nil)
	if err == nil {
		t.Fatal("want transport error against a closed server, got nil")
	}
	if code != 0 {
		t.Fatalf("want status 0 on transport error, got %d", code)
	}
}

// TestUnitHTTPSender_BadURL proves request construction failure is surfaced.
func TestUnitHTTPSender_BadURL(t *testing.T) {
	sender := webhook.NewHTTPSender()
	// Control characters in the URL make http.NewRequestWithContext fail.
	code, err := sender.Post(context.Background(), "http://\x7f invalid", []byte(`{}`), nil)
	if err == nil {
		t.Fatal("want request-build error for a malformed URL, got nil")
	}
	if code != 0 {
		t.Fatalf("want status 0 on build error, got %d", code)
	}
}

// =============================================================================
// Test doubles: FakeSender default code, FakeVerifier
// =============================================================================

// TestUnitFakeSender_DefaultStatus proves the FakeSender returns 200 when its
// StatusCode field is left at zero.
func TestUnitFakeSender_DefaultStatus(t *testing.T) {
	f := &webhook.FakeSender{} // StatusCode zero
	code, err := f.Post(context.Background(), "http://x", []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("want default 200, got %d", code)
	}
	if len(f.Calls) != 1 {
		t.Fatalf("want 1 recorded call, got %d", len(f.Calls))
	}
}

// TestUnitFakeVerifier proves the FakeVerifier passes on the matching test
// header and fails otherwise.
func TestUnitFakeVerifier(t *testing.T) {
	v := webhook.FakeVerifier{Secret: "open-sesame"}
	if _, err := v.Verify("", nil, map[string]string{"X-Test-Sig": "open-sesame"}); err != nil {
		t.Fatalf("matching header should pass, got %v", err)
	}
	_, err := v.Verify("", nil, map[string]string{"X-Test-Sig": "wrong"})
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated on mismatch, got %v", err)
	}
}

// TestUnitHMACVerifier_MissingHeader proves the HMACVerifier rejects a request
// with no signature header.
func TestUnitHMACVerifier_MissingHeader(t *testing.T) {
	_, err := webhook.HMACVerifier{}.Verify(testSecret, []byte(`{}`), map[string]string{})
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated for missing header, got %v", err)
	}
}

// =============================================================================
// Constructor guard rails
// =============================================================================

// TestUnitNew_PanicsOnNilDeps proves New and NewWithClock panic when required
// dependencies are missing.
func TestUnitNew_PanicsOnNilDeps(t *testing.T) {
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	idgen := model.UUIDv7()

	mustPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatalf("%s: expected panic, got none", name)
			}
		}()
		fn()
	}

	mustPanic("New(nil sender)", func() { webhook.New(nil, resolver, idgen) })
	mustPanic("New(nil secrets)", func() { webhook.New(&webhook.FakeSender{}, nil, idgen) })
	mustPanic("New(nil idgen)", func() { webhook.New(&webhook.FakeSender{}, resolver, nil) })
	mustPanic("NewWithClock(nil clock)", func() {
		webhook.NewWithClock(&webhook.FakeSender{}, resolver, idgen, nil)
	})
}

// recordingMetrics captures SetGauge emissions.
type recordingMetrics struct{ gauges map[string]float64 }

func (m *recordingMetrics) ObserveRequest(_, _ string, _ int, _ time.Duration, _ int) {}
func (m *recordingMetrics) IncCounter(_ string, _ float64, _ map[string]string)       {}
func (m *recordingMetrics) ObserveHistogram(_ string, _ float64, _ map[string]string) {}
func (m *recordingMetrics) SetGauge(name string, v float64, _ map[string]string)      { m.gauges[name] = v }

var _ observability.Metrics = (*recordingMetrics)(nil)

// TestIntegrationWithMetrics_EmitsBreakerGauge proves WithMetrics wires a sink
// that receives the webhook_breaker_state gauge on delivery.
func TestIntegrationWithMetrics_EmitsBreakerGauge(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/metrics")

	m := &recordingMetrics{gauges: map[string]float64{}}
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	svc := webhook.New(&webhook.FakeSender{StatusCode: 200}, resolver, model.UUIDv7(), webhook.WithMetrics(m))
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	if _, ok := m.gauges["webhook_breaker_state"]; !ok {
		t.Fatal("expected webhook_breaker_state gauge to be emitted")
	}
	if got := m.gauges["webhook_breaker_state"]; got != 0 {
		t.Fatalf("want closed (0) after a successful delivery, got %v", got)
	}
}

// TestUnitWithMetrics_NilIsIgnored proves WithMetrics(nil) does not override the
// NoOp sink (service must still construct and dispatch without panicking).
func TestUnitWithMetrics_NilIsIgnored(t *testing.T) {
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	svc := webhook.New(&webhook.FakeSender{}, resolver, model.UUIDv7(), webhook.WithMetrics(nil))
	if svc == nil {
		t.Fatal("New returned nil with WithMetrics(nil)")
	}
}

// =============================================================================
// HandleInbound error branches
// =============================================================================

// TestIntegrationHandleInbound_EndpointNotFound proves an unknown endpoint id
// returns KindNotFound.
func TestIntegrationHandleInbound_EndpointNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	svc := newService(t, &webhook.FakeSender{})

	var err error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		err = svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: uuid.New(), ProviderKey: testProviderKey,
			RawBody: []byte(`{}`), EventType: "x", Timestamp: time.Now(),
		})
		return nil
	}); cerr != nil {
		t.Fatalf("tx: %v", cerr)
	}
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("want KindNotFound, got kind=%v err=%v", kerr.KindOf(err), err)
	}
}

// TestIntegrationHandleInbound_WrongDirection proves posting an inbound event to
// an OUTBOUND endpoint is rejected with KindValidation.
func TestIntegrationHandleInbound_WrongDirection(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/out")
	svc := newService(t, &webhook.FakeSender{})

	var err error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		err = svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: testProviderKey,
			RawBody: []byte(`{}`), EventType: "order.created", Timestamp: time.Now(),
		})
		return nil
	}); cerr != nil {
		t.Fatalf("tx: %v", cerr)
	}
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("want KindValidation for wrong direction, got kind=%v err=%v", kerr.KindOf(err), err)
	}
}

// TestIntegrationHandleInbound_InactiveEndpoint proves a non-active inbound
// endpoint is rejected with KindConflict.
func TestIntegrationHandleInbound_InactiveEndpoint(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpointStatus(t, h, tn.ID, "disabled")
	svc := newService(t, &webhook.FakeSender{})

	var err error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		err = svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: testProviderKey,
			RawBody: []byte(`{}`), EventType: "order.created", Timestamp: time.Now(),
		})
		return nil
	}); cerr != nil {
		t.Fatalf("tx: %v", cerr)
	}
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("want KindConflict for inactive endpoint, got kind=%v err=%v", kerr.KindOf(err), err)
	}
}

// TestIntegrationHandleInbound_NoVerifier proves an unregistered provider key is
// rejected with KindValidation.
func TestIntegrationHandleInbound_NoVerifier(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{}) // only testProviderKey registered

	var err error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		err = svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: "unregistered-provider",
			RawBody: []byte(`{}`), EventType: "order.created", Timestamp: time.Now(),
		})
		return nil
	}); cerr != nil {
		t.Fatalf("tx: %v", cerr)
	}
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("want KindValidation for missing verifier, got kind=%v err=%v", kerr.KindOf(err), err)
	}
}

// TestIntegrationHandleInbound_SecretResolveError proves a secret-store failure
// is wrapped and returned (not swallowed).
func TestIntegrationHandleInbound_SecretResolveError(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)

	svc := webhook.New(&webhook.FakeSender{}, failingResolver{}, model.UUIDv7())
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})

	var err error
	if cerr := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		err = svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: testProviderKey,
			RawBody: []byte(`{}`), EventType: "order.created", Timestamp: time.Now(),
		})
		return nil
	}); cerr != nil {
		t.Fatalf("tx: %v", cerr)
	}
	if err == nil {
		t.Fatal("want error when secret resolution fails, got nil")
	}
	if n := countEvents(t, h, tn.ID); n != 0 {
		t.Fatalf("no row should be written on secret failure, got %d", n)
	}
}

// =============================================================================
// ProcessInbound: no handler registered (no-op → processed)
// =============================================================================

// TestIntegrationProcessInbound_NoHandler proves an event with no registered
// handler is treated as a no-op and advanced to processed.
func TestIntegrationProcessInbound_NoHandler(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedInboundEndpoint(t, h, tn.ID)
	svc := newService(t, &webhook.FakeSender{}) // no handlers registered

	body := []byte(`{"event":"unhandled"}`)
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return svc.HandleInbound(ctx, db, webhook.InboundIn{
			EndpointID: epID, ProviderKey: testProviderKey,
			RawBody: body, Headers: map[string]string{"X-Signature": testSign(body)},
			ExternalEventID: "ext-nohandler", EventType: "unhandled.event", Timestamp: time.Now(),
		})
	}); err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}
	if err := svc.ProcessInbound(context.Background(), h.PlatformTxM, tn.ID, time.Now()); err != nil {
		t.Fatalf("ProcessInbound: %v", err)
	}
	if s := eventStatus(t, h, tn.ID); s != "processed" {
		t.Fatalf("want processed for unhandled event (no-op), got %s", s)
	}
}

// =============================================================================
// deliverToEndpoint: dead-letter + terminal/not-due skips
// =============================================================================

// TestIntegrationDeliverToEndpoint_DeadLetterAndSkips drives a single outbound
// event through the full failure ladder to 'dead', exercising the exponential
// backoff arms, the breaker-open degraded update, the dead-letter ceiling, and
// the terminal/not-yet-due skip branches.
func TestIntegrationDeliverToEndpoint_DeadLetterAndSkips(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/dead")

	clk := fakes.NewClock(time.Now())
	sender := &webhook.FakeSender{StatusCode: 500}
	svc := newServiceWithClock(t, sender, clk)

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{"n":1}`)}

	// First failing attempt → status 'failed', attempts=1.
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
		t.Fatalf("dispatch #1: %v", err)
	}
	callsAfterFirst := len(sender.Calls)

	// Immediate re-dispatch without advancing the clock: the failed row is not
	// yet due (next_attempt_at in the future) → skipped, no new POST.
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
		t.Fatalf("dispatch not-due: %v", err)
	}
	if len(sender.Calls) != callsAfterFirst {
		t.Fatalf("not-due row should be skipped; calls %d → %d", callsAfterFirst, len(sender.Calls))
	}

	// Drive the remaining attempts, advancing past each backoff window, until the
	// row dead-letters at MaxAttempts.
	for i := 2; i <= webhook.MaxAttempts; i++ {
		clk.Advance(6 * time.Minute) // exceeds every backoff step (max 5 m)
		if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
			t.Fatalf("dispatch #%d: %v", i, err)
		}
	}

	var status string
	var attempts int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status, attempts FROM webhook_events WHERE endpoint_id = $1`,
		epID).Scan(&status, &attempts); err != nil {
		t.Fatalf("query row: %v", err)
	}
	if status != "dead" {
		t.Fatalf("want dead after %d attempts, got %s (attempts=%d)", webhook.MaxAttempts, status, attempts)
	}
	if attempts != webhook.MaxAttempts {
		t.Fatalf("want attempts=%d, got %d", webhook.MaxAttempts, attempts)
	}

	// Endpoint was marked degraded when the breaker opened.
	var epStatus string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM webhook_endpoints WHERE id = $1`, epID).Scan(&epStatus); err != nil {
		t.Fatalf("query endpoint: %v", err)
	}
	if epStatus != "degraded" {
		t.Fatalf("want endpoint degraded after breaker open, got %s", epStatus)
	}

	// Re-dispatch a dead row: terminal → skipped, no new POST.
	callsBefore := len(sender.Calls)
	clk.Advance(6 * time.Minute)
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, clk.Now()); err != nil {
		t.Fatalf("dispatch dead row: %v", err)
	}
	if len(sender.Calls) != callsBefore {
		t.Fatalf("dead (terminal) row should be skipped; calls %d → %d", callsBefore, len(sender.Calls))
	}
}

// TestIntegrationDeliverToEndpoint_DeliveredSkip proves a delivered row is
// terminal and re-dispatch of the same event id issues no further POST.
func TestIntegrationDeliverToEndpoint_DeliveredSkip(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/delivered")

	sender := &webhook.FakeSender{StatusCode: 200}
	svc := newService(t, sender)

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(sender.Calls) != 1 {
		t.Fatalf("want 1 POST, got %d", len(sender.Calls))
	}
	// Re-dispatch same event id → row already 'delivered' → skipped.
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("re-dispatch: %v", err)
	}
	if len(sender.Calls) != 1 {
		t.Fatalf("delivered row must not re-POST, got %d calls", len(sender.Calls))
	}
}

// TestIntegrationDeliverToEndpoint_SecretResolveError proves a secret-store
// failure during outbound delivery is surfaced (the delivery row stays pending,
// no POST is made).
func TestIntegrationDeliverToEndpoint_SecretResolveError(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/nosecret")

	sender := &webhook.FakeSender{StatusCode: 200}
	svc := webhook.New(sender, failingResolver{}, model.UUIDv7())
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{}`)}
	// DispatchOutbound swallows per-endpoint errors, so it returns nil even though
	// delivery failed to resolve the secret.
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	if len(sender.Calls) != 0 {
		t.Fatalf("no POST should happen when the secret cannot be resolved, got %d", len(sender.Calls))
	}
	var status string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE endpoint_id = $1`, epID).Scan(&status); err != nil {
		t.Fatalf("query row: %v", err)
	}
	if status != "pending" {
		t.Fatalf("row should remain pending after secret failure, got %s", status)
	}
}

// =============================================================================
// RetryOutbound: non-UUID external_event_id is skipped, not fatal
// =============================================================================

// TestIntegrationRetryOutbound_SkipsNonUUID proves a failed outbound row whose
// external_event_id is not a UUID is skipped by RetryOutbound (defensive
// branch) rather than crashing the batch.
func TestIntegrationRetryOutbound_SkipsNonUUID(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	epID := seedOutboundEndpoint(t, h, tn.ID, "https://example.test/badid")

	// Seed a failed outbound delivery row with a non-UUID external id.
	rowID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO webhook_events
		    (id, tenant_id, endpoint_id, direction, external_event_id, event_type,
		     payload, received_at, delivery_status, attempts)
		 VALUES ($1, $2, $3, 'outbound', 'not-a-uuid', 'order.created',
		         '{}'::jsonb, now(), 'failed', 1)`,
		rowID, tn.ID, epID); err != nil {
		t.Fatalf("seed bad-id row: %v", err)
	}

	sender := &webhook.FakeSender{StatusCode: 200}
	svc := newService(t, sender)

	if err := svc.RetryOutbound(context.Background(), h.PlatformTxM, tn.ID, time.Now()); err != nil {
		t.Fatalf("RetryOutbound must not fail on a non-UUID id: %v", err)
	}
	if len(sender.Calls) != 0 {
		t.Fatalf("non-UUID row must be skipped (no POST), got %d", len(sender.Calls))
	}
	// The row is untouched — still 'failed'.
	var status string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE id = $1`, rowID).Scan(&status); err != nil {
		t.Fatalf("re-query row: %v", err)
	}
	if status != "failed" {
		t.Fatalf("non-UUID row should stay failed, got %s", status)
	}
}
