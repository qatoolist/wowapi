package notify

import (
	"context"
	"sync"
)

// ChannelSender is the port for channel-specific delivery adapters (smtp, sms,
// whatsapp, push). Real adapters implement this interface; the fake is provided
// for tests. Adapters are responsible for fetching and rendering the template
// body (via RenderBody) from the notification_templates DB rows when needed —
// the Delivery carries routing information only, not rendered content.
type ChannelSender interface {
	// Send attempts to deliver d. It returns the provider-assigned message ID
	// on success, or an error. Errors are recorded on the delivery row and
	// retried up to maxAttempts.
	Send(ctx context.Context, d Delivery) (providerMessageID string, err error)
}

// FakeSender is an in-memory ChannelSender for integration tests. It records
// every Delivery passed to Send and returns a deterministic provider message ID
// ("fake-msg-" + delivery ID). Set Err to make Send return an error.
type FakeSender struct {
	mu         sync.Mutex
	Deliveries []Delivery
	Err        error // if non-nil, returned by every Send call
}

// Send records the delivery and returns a fake provider message ID, or the
// configured Err.
func (f *FakeSender) Send(_ context.Context, d Delivery) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return "", f.Err
	}
	f.Deliveries = append(f.Deliveries, d)
	return "fake-msg-" + d.ID.String(), nil
}

// Count returns the number of deliveries recorded (thread-safe).
func (f *FakeSender) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Deliveries)
}

// Reset clears recorded deliveries and the configured Err.
func (f *FakeSender) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Deliveries = nil
	f.Err = nil
}

// inAppSender is the built-in ChannelSender for the inapp channel. "Sending"
// an in-app notification is a no-op at the transport level — the row is already
// written by Send and queried by the /notifications API. The delivery is
// immediately considered delivered.
type inAppSender struct{}

func (inAppSender) Send(_ context.Context, d Delivery) (string, error) {
	// In-app "delivery" is just the row written by Send; no transport needed.
	return "inapp-" + d.ID.String(), nil
}

// noopSender is a fallback for channels with no registered sender.
type noopSender struct{}

func (noopSender) Send(_ context.Context, _ Delivery) (string, error) {
	return "", nil
}
