package fakes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/qatoolist/wowapi/foundation/mfa"
	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/foundation/webhook"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/safety"
)

// WebhookCall records one outbound webhook attempt.
type WebhookCall struct {
	URL     string
	Body    []byte
	Headers map[string]string
}

// WebhookSender is an in-memory webhook sender for tests.
type WebhookSender struct {
	StatusCode int
	Err        error
	Calls      []WebhookCall
}

func (f *WebhookSender) Post(_ context.Context, url string, body []byte, headers map[string]string) (int, error) {
	f.Calls = append(f.Calls, WebhookCall{URL: url, Body: body, Headers: headers})
	if f.StatusCode == 0 {
		return http.StatusOK, f.Err
	}
	return f.StatusCode, f.Err
}

func (*WebhookSender) DuplicateSafety() safety.Mechanism { return safety.None }

// WebhookSecretResolver returns one fixed secret for tests.
type WebhookSecretResolver struct{ Secret string }

func (r *WebhookSecretResolver) Resolve(context.Context, string) (string, error) {
	return r.Secret, nil
}

// WebhookVerifier is a deterministic verifier for tests.
type WebhookVerifier struct{ Secret string }

func (v WebhookVerifier) Verify(_ string, body []byte, headers map[string]string) (webhook.Envelope, error) {
	if headers["X-Test-Sig"] != v.Secret {
		return webhook.Envelope{}, kerr.E(kerr.KindUnauthenticated, "signature_mismatch", "fake verifier: signature does not match")
	}
	sum := sha256.Sum256(body)
	return webhook.Envelope{
		CanonicalBody: body, EventID: "sha256:" + hex.EncodeToString(sum[:]),
		OccurredAt: time.Now(), SignatureVersion: "test",
	}, nil
}

// NotifySender is an in-memory notification channel sender for tests.
type NotifySender struct {
	mu         sync.Mutex
	Deliveries []notify.Delivery
	Err        error
}

func (f *NotifySender) Send(_ context.Context, d notify.Delivery) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return "", f.Err
	}
	f.Deliveries = append(f.Deliveries, d)
	return "fake-msg-" + d.ID.String(), nil
}

func (*NotifySender) DuplicateSafety() safety.Mechanism { return safety.None }

func (f *NotifySender) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Deliveries)
}

func (f *NotifySender) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Deliveries = nil
	f.Err = nil
}

// MFADelivery records one out-of-band MFA delivery in tests.
type MFADelivery struct {
	Destination string
	Body        string
}

// MFASender is an in-memory MFA sender for tests.
type MFASender struct {
	mu         sync.Mutex
	Deliveries []MFADelivery
	Err        error
}

func (f *MFASender) Send(_ context.Context, destination, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.Deliveries = append(f.Deliveries, MFADelivery{Destination: destination, Body: body})
	return nil
}

func (f *MFASender) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Deliveries)
}

func (f *MFASender) LastCode() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Deliveries) == 0 {
		return ""
	}
	fields := strings.Fields(f.Deliveries[len(f.Deliveries)-1].Body)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

func (f *MFASender) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Deliveries = nil
	f.Err = nil
}

var (
	_ webhook.Sender         = (*WebhookSender)(nil)
	_ webhook.SecretResolver = (*WebhookSecretResolver)(nil)
	_ webhook.Verifier       = WebhookVerifier{}
	_ notify.ChannelSender   = (*NotifySender)(nil)
	_ mfa.Sender             = (*MFASender)(nil)
)
