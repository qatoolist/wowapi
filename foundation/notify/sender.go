package notify

import (
	"context"

	"github.com/qatoolist/wowapi/kernel/safety"
)

// ChannelSender is the port for channel-specific delivery adapters (smtp, sms,
// whatsapp, push). Real adapters implement this interface; test doubles live in
// testkit/fakes. Adapters are responsible for fetching and rendering the template
// body (via RenderBody) from the notification_templates DB rows when needed —
// the Delivery carries routing information only, not rendered content.
type ChannelSender interface {
	// Send attempts to deliver d. It returns the provider-assigned message ID
	// on success, or an error. Errors are recorded on the delivery row and
	// retried up to maxAttempts.
	Send(ctx context.Context, d Delivery) (providerMessageID string, err error)
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

// DuplicateSafety declares that in-app delivery is duplicate-safe by domain
// compare-and-swap: the notification_deliveries row itself is the single
// durable effect, and a retried Send is a no-op against the same row.
func (inAppSender) DuplicateSafety() safety.Mechanism { return safety.DomainCAS }
