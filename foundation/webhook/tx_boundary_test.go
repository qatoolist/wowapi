package webhook_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/foundation/webhook"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/kernel/safety"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// =============================================================================
// C-1 regression: no remote I/O (secret resolution, HTTP POST) while a DB
// transaction is open (W04-E02-S001 AC-02/03).
//
// txDepthTracker wraps a real database.TxManager and counts how many
// WithTenant/WithTenantRO/Platform calls are currently open (nested calls are
// possible if the fix regresses to running the effect stage from inside a
// callback). txAssertingSender / txAssertingSecretResolver read that counter
// at the exact moment Post / Resolve run and fail the test if it is nonzero.
//
// Revert-sensitivity: with the pre-fix code, DispatchOutbound/RetryOutbound
// call deliverToEndpoint (which resolves the secret and POSTs) FROM INSIDE
// the plat.WithTenant callback, so depth==1 at both call sites and these
// tests go red. The staged claim/deliver/finalize implementation calls
// effectDeliver — secret resolution + POST — strictly between the claim
// transaction's commit and the finalize transaction's begin, so depth==0.
// =============================================================================

// txDepthTracker counts currently-open transactions started through it.
type txDepthTracker struct {
	inner database.TxManager
	depth int32
}

func (p *txDepthTracker) WithTenant(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	atomic.AddInt32(&p.depth, 1)
	defer atomic.AddInt32(&p.depth, -1)
	return p.inner.WithTenant(ctx, fn)
}

func (p *txDepthTracker) WithTenantRO(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	atomic.AddInt32(&p.depth, 1)
	defer atomic.AddInt32(&p.depth, -1)
	return p.inner.WithTenantRO(ctx, fn)
}

func (p *txDepthTracker) Platform(ctx context.Context, fn func(context.Context, database.DB) error) error {
	atomic.AddInt32(&p.depth, 1)
	defer atomic.AddInt32(&p.depth, -1)
	return p.inner.Platform(ctx, fn)
}

// txAssertingSender wraps a Sender and fails t if Post is invoked while
// tracker reports an open transaction.
type txAssertingSender struct {
	t       *testing.T
	tracker *txDepthTracker
	inner   *webhook.FakeSender
}

func (s *txAssertingSender) Post(ctx context.Context, url string, body []byte, headers map[string]string) (int, error) {
	s.t.Helper()
	if d := atomic.LoadInt32(&s.tracker.depth); d != 0 {
		s.t.Errorf("C-1 regression: Sender.Post called while a DB transaction is open (depth=%d)", d)
	}
	return s.inner.Post(ctx, url, body, headers)
}

func (s *txAssertingSender) DuplicateSafety() safety.Mechanism { return safety.None }

// txAssertingSecretResolver wraps a SecretResolver and fails t if Resolve is
// invoked while tracker reports an open transaction.
type txAssertingSecretResolver struct {
	t       *testing.T
	tracker *txDepthTracker
	secret  string
}

func (r *txAssertingSecretResolver) Resolve(_ context.Context, _ string) (string, error) {
	r.t.Helper()
	if d := atomic.LoadInt32(&r.tracker.depth); d != 0 {
		r.t.Errorf("C-1 regression: SecretResolver.Resolve called while a DB transaction is open (depth=%d)", d)
	}
	return r.secret, nil
}

// TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO proves DispatchOutbound
// never resolves the endpoint secret nor POSTs while a DB transaction is open
// (C-1 / W04-E02-S001 AC-02/03).
func TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/txcheck")

	tracker := &txDepthTracker{inner: h.PlatformTxM}
	sender := &txAssertingSender{t: t, tracker: tracker, inner: &webhook.FakeSender{StatusCode: 200}}
	resolver := &txAssertingSecretResolver{t: t, tracker: tracker, secret: testSecret}

	svc := webhook.New(sender, resolver, model.UUIDv7())
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})

	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{"k":"v"}`)}
	if err := svc.DispatchOutbound(context.Background(), tracker, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("DispatchOutbound: %v", err)
	}
	if len(sender.inner.Calls) != 1 {
		t.Fatalf("want 1 POST, got %d", len(sender.inner.Calls))
	}
}

// TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO proves RetryOutbound
// never resolves the endpoint secret nor POSTs while a DB transaction is open
// (C-1 / W04-E02-S001 AC-02/03).
func TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	seedOutboundEndpoint(t, h, tn.ID, "https://example.test/txcheck-retry")

	// First, dispatch with a failing sender (untracked) to seed a 'failed' row.
	failSender := &webhook.FakeSender{StatusCode: 500}
	seedSvc := webhook.New(failSender, &webhook.FakeSecretResolver{Secret: testSecret}, model.UUIDv7())
	seedSvc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})
	ev := outbox.Event{ID: uuid.New(), Type: "order.created", Payload: json.RawMessage(`{"k":"v"}`)}
	if err := seedSvc.DispatchOutbound(context.Background(), h.PlatformTxM, tn.ID, ev, time.Now()); err != nil {
		t.Fatalf("seed dispatch: %v", err)
	}

	tracker := &txDepthTracker{inner: h.PlatformTxM}
	sender := &txAssertingSender{t: t, tracker: tracker, inner: &webhook.FakeSender{StatusCode: 200}}
	resolver := &txAssertingSecretResolver{t: t, tracker: tracker, secret: testSecret}

	svc := webhook.New(sender, resolver, model.UUIDv7())
	svc.RegisterVerifier(testProviderKey, webhook.HMACVerifier{})

	// Advance well past the first backoff step so the row is due for retry.
	retryNow := time.Now().Add(time.Hour)
	if err := svc.RetryOutbound(context.Background(), tracker, tn.ID, retryNow); err != nil {
		t.Fatalf("RetryOutbound: %v", err)
	}
	if len(sender.inner.Calls) != 1 {
		t.Fatalf("want 1 retry POST, got %d", len(sender.inner.Calls))
	}
}
