package webhook

import (
	"context"
	"reflect"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/safety"
)

type registrationSender struct{}

func (registrationSender) Post(context.Context, string, []byte, map[string]string) (int, error) {
	return 200, nil
}
func (registrationSender) DuplicateSafety() safety.Mechanism { return safety.DomainCAS }

type registrationSecrets struct{}

func (registrationSecrets) Resolve(context.Context, string) (string, error) { return "", nil }

type identityVerifier struct{ id string }

func (*identityVerifier) Verify(string, []byte, map[string]string) (Envelope, error) {
	return Envelope{}, nil
}

func TestDuplicateRegistrationsPreserveOriginals(t *testing.T) {
	svc := New(registrationSender{}, registrationSecrets{}, model.UUIDv7())
	firstVerifier, secondVerifier := &identityVerifier{id: "first"}, &identityVerifier{id: "second"}
	firstHandler := func(context.Context, database.TenantDB, Event) error { return nil }
	secondHandler := func(context.Context, database.TenantDB, Event) error { return nil }
	svc.RegisterVerifier("provider", firstVerifier)
	svc.RegisterHandler("event", firstHandler)

	for name, duplicate := range map[string]func(){
		"verifier": func() { svc.RegisterVerifier("provider", secondVerifier) },
		"handler":  func() { svc.RegisterHandler("event", secondHandler) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("duplicate registration did not panic")
				}
			}()
			duplicate()
		})
	}
	if svc.verifiers["provider"] != firstVerifier {
		t.Fatal("duplicate verifier replaced the original")
	}
	if reflect.ValueOf(svc.handlers["event"]).Pointer() != reflect.ValueOf(firstHandler).Pointer() {
		t.Fatal("duplicate handler replaced the original")
	}
}
