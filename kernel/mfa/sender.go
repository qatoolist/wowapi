package mfa

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

// Sender is the delivery port for an out-of-band factor code (SMS or email
// body text — the destination address format is caller-defined: a phone
// number for SMS, an email address for email). Real provider adapters
// (Twilio, SES, etc.) are product territory; this package only defines the
// shape and ships a log adapter (for local/dev) and a fake (for tests),
// mirroring how kernel/notify.ChannelSender separates the port from any
// concrete transport. mfa.Sender is intentionally NOT kernel/notify's
// ChannelSender: notify's port is keyed to a persisted Delivery row from the
// notification_deliveries schema, which is exactly the storage coupling this
// leaf package must not take on.
type Sender interface {
	// Send delivers body to destination. Errors are returned to the caller;
	// this package does not retry — retry/backoff policy is product-owned.
	Send(ctx context.Context, destination, body string) error
}

// delivery is one recorded Send call, captured by FakeSender.
type delivery struct {
	Destination string
	Body        string
}

// FakeSender is an in-memory Sender for tests: it records every Send call
// and returns the configured Err (if any) instead of delivering anything.
type FakeSender struct {
	mu         sync.Mutex
	Deliveries []delivery
	// Err, if non-nil, is returned by every Send call (and the call is NOT
	// recorded, matching "delivery failed" semantics).
	Err error
}

// Send records the delivery, or returns f.Err without recording.
func (f *FakeSender) Send(_ context.Context, destination, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.Deliveries = append(f.Deliveries, delivery{Destination: destination, Body: body})
	return nil
}

// Count returns the number of successfully recorded deliveries.
func (f *FakeSender) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Deliveries)
}

// LastCode extracts the last whitespace-separated token of the most recent
// delivery's body — a convenience for tests that send a "...code is 123456"
// style message and want the code back without re-deriving it via a second
// channel.
func (f *FakeSender) LastCode() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Deliveries) == 0 {
		return ""
	}
	body := f.Deliveries[len(f.Deliveries)-1].Body
	fields := strings.Fields(body)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// Reset clears recorded deliveries and the configured Err.
func (f *FakeSender) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Deliveries = nil
	f.Err = nil
}

// logSender logs the code instead of delivering it — the safe-by-default
// dev/local adapter: it never touches the network, so it can be wired as the
// default when no real provider is configured without risk of an accidental
// production delivery path.
type logSender struct {
	log *slog.Logger
}

// NewLogSender returns the logging dev/local Sender adapter.
func NewLogSender(log *slog.Logger) Sender {
	return &logSender{log: log}
}

func (s *logSender) Send(ctx context.Context, destination, body string) error {
	s.log.InfoContext(ctx, "mfa: sender (dev/log adapter, not actually delivered)",
		"destination", redactDestination(destination), "body", body)
	return nil
}

// redactDestination keeps only the last 4 characters, matching the
// convention already used for phone redaction elsewhere in the framework
// (kernel/... adapters redact PII destinations in logs by default).
func redactDestination(destination string) string {
	if len(destination) <= 4 {
		return "****"
	}
	return "****" + destination[len(destination)-4:]
}
