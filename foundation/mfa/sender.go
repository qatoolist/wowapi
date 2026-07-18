package mfa

import (
	"context"
	"log/slog"
)

// Sender is the delivery port for an out-of-band factor code (SMS or email
// body text — the destination address format is caller-defined: a phone
// number for SMS, an email address for email). Real provider adapters
// (Twilio, SES, etc.) are product territory; this package only defines the
// shape and ships a log adapter for local/dev; test doubles live in testkit/fakes,
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
