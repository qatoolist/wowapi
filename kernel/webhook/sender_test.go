package webhook_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/httpclient"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/webhook"
	"github.com/qatoolist/wowapi/testkit"
)

// TestHTTPSenderBlocksLoopbackByDefault proves NewHTTPSender is SSRF-safe by
// default (backlog B2): a webhook endpoint URL pointing at loopback (which
// httptest always binds to) must be refused, not delivered.
func TestHTTPSenderBlocksLoopbackByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := webhook.NewHTTPSender()
	_, err := sender.Post(context.Background(), srv.URL, []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected the default sender to block a loopback delivery target")
	}
	if !errors.Is(err, httpclient.ErrBlockedAddress) {
		t.Errorf("expected the error chain to contain httpclient.ErrBlockedAddress, got %v", err)
	}
}

// TestHTTPSenderAllowlistEscapeHatch proves the config/opt escape hatch: a
// product that intentionally wants to deliver to an internal target can
// allowlist it, and delivery then succeeds.
func TestHTTPSenderAllowlistEscapeHatch(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 2)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	host := srv.Listener.Addr().String()
	sender := webhook.NewHTTPSender(webhook.WithHTTPClientConfig(httpclient.Config{
		AllowedHosts: []string{hostOnly(t, host)},
	}))

	code, err := sender.Post(context.Background(), srv.URL, []byte(`{}`), map[string]string{"X-Test": "1"})
	if err != nil {
		t.Fatalf("expected the allowlisted target to succeed, got %v", err)
	}
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", code)
	}
	_ = gotBody
}

func hostOnly(t *testing.T, hostport string) string {
	t.Helper()
	for i := len(hostport) - 1; i >= 0; i-- {
		if hostport[i] == ':' {
			return hostport[:i]
		}
	}
	return hostport
}

// TestHTTPSenderDisabledProtectionEscapeHatch proves the blanket opt-out knob
// (mirrors config.WebhookOutbound.SSRFProtectionDisabled) reaches the sender:
// with protection disabled, a loopback target is delivered to.
func TestHTTPSenderDisabledProtectionEscapeHatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := webhook.NewHTTPSender(webhook.WithSSRFProtectionDisabled())
	_, err := sender.Post(context.Background(), srv.URL, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("expected delivery to succeed with SSRF protection disabled, got %v", err)
	}
}

// TestHTTPSenderPublicDeliveryUnaffected proves a public-looking destination
// (simulated by allowlisting the loopback httptest addr, standing in for a
// real public host) still delivers exactly as before: same status code, same
// error-free path — the guard adds no behavior change beyond the address check.
func TestHTTPSenderPublicDeliveryUnaffected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	sender := webhook.NewHTTPSender(webhook.WithHTTPClientConfig(httpclient.Config{
		AllowedHosts: []string{hostOnly(t, srv.Listener.Addr().String())},
	}))
	code, err := sender.Post(context.Background(), srv.URL, []byte(`{"ok":true}`), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", code)
	}
}

// =============================================================================
// End-to-end: Service.DispatchOutbound wired with the REAL default sender
// (webhook.NewHTTPSender() with no options — exactly what kernel.New builds
// when deps.WebhookSender is nil and SSRF protection is not disabled by
// config), proving default-on protection through the full delivery path, not
// just at the HTTPSender.Post layer.
// =============================================================================

// TestIntegrationDispatchOutbound_DefaultSenderBlocksLoopback proves that an
// outbound webhook endpoint pointed at a loopback target (the only kind an
// httptest server can offer) is refused by the DEFAULT sender end-to-end:
// DispatchOutbound records the delivery as failed (not delivered), backlog B2's
// "default-on" requirement holding all the way from kernel wiring down to the
// DB row.
func TestIntegrationDispatchOutbound_DefaultSenderBlocksLoopback(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	epID := seedOutboundEndpoint(t, h, tn.ID, srv.URL)
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	// The exact construction kernel.New performs by default: no allowlist, SSRF
	// protection on.
	svc := webhook.New(webhook.NewHTTPSender(), resolver, model.UUIDv7())

	ev := outbox.Event{ID: uuid.New(), Type: "order.created"}
	now := time.Now()
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, now); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}

	var status string
	var lastErr *string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status, last_error FROM webhook_events WHERE endpoint_id = $1`,
		epID).Scan(&status, &lastErr); err != nil {
		t.Fatalf("query delivery row: %v", err)
	}
	if status == "delivered" {
		t.Fatalf("delivery_status = %q, want anything but delivered — the default sender must block the loopback target", status)
	}
	if status != "failed" {
		t.Fatalf("delivery_status = %q, want failed", status)
	}
	if lastErr == nil || !containsSSRFBlockedMessage(*lastErr) {
		t.Fatalf("last_error = %v, want it to mention the SSRF block", lastErr)
	}
}

// TestIntegrationDispatchOutbound_AllowlistedLoopbackDelivers is the control:
// the same setup as above, but with the loopback target allowlisted via
// WithHTTPClientConfig — proving the escape hatch reaches all the way through
// the real delivery path and the row is marked delivered.
func TestIntegrationDispatchOutbound_AllowlistedLoopbackDelivers(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	epID := seedOutboundEndpoint(t, h, tn.ID, srv.URL)
	resolver := &webhook.FakeSecretResolver{Secret: testSecret}
	sender := webhook.NewHTTPSender(webhook.WithHTTPClientConfig(httpclient.Config{
		AllowedHosts: []string{hostOnly(t, srv.Listener.Addr().String())},
	}))
	svc := webhook.New(sender, resolver, model.UUIDv7())

	ev := outbox.Event{ID: uuid.New(), Type: "order.created"}
	now := time.Now()
	if err := svc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, now); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}

	var status string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT delivery_status FROM webhook_events WHERE endpoint_id = $1`,
		epID).Scan(&status); err != nil {
		t.Fatalf("query delivery row: %v", err)
	}
	if status != "delivered" {
		t.Fatalf("delivery_status = %q, want delivered", status)
	}
}

func containsSSRFBlockedMessage(s string) bool {
	return strings.Contains(s, "SSRF") || strings.Contains(s, "blocked")
}
