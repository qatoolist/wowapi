package notify

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/safety"
)

type identitySender struct{ id string }

func (*identitySender) Send(context.Context, Delivery) (string, error) { return "", nil }
func (*identitySender) DuplicateSafety() safety.Mechanism              { return safety.DomainCAS }

func TestRegisterSenderDuplicatePreservesOriginal(t *testing.T) {
	svc := New(NewRegistry(), model.UUIDv7())
	first, second := &identitySender{id: "first"}, &identitySender{id: "second"}
	svc.RegisterSender(ChannelEmail, first)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("duplicate sender registration did not panic")
			}
		}()
		svc.RegisterSender(ChannelEmail, second)
	}()
	got, ok := svc.senderFor(ChannelEmail)
	if !ok || got != first {
		t.Fatalf("duplicate replaced original sender: got %#v, want %#v", got, first)
	}
}
