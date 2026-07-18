package kernel

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/safety"
	"github.com/qatoolist/wowapi/kernel/secrets"
	"github.com/qatoolist/wowapi/kernel/storage"
)

// stubProvider is a secrets.Provider double so we can exercise secretRefResolver
// without a real provider adapter.
type stubProvider struct {
	val string
	err error
}

type typedNilKernelWebhookSender struct{}

func (*typedNilKernelWebhookSender) Post(context.Context, string, []byte, map[string]string) (int, error) {
	return 0, nil
}
func (*typedNilKernelWebhookSender) DuplicateSafety() safety.Mechanism { return safety.DomainCAS }

func (s stubProvider) Resolve(_ context.Context, _ secrets.Ref) (string, error) {
	return s.val, s.err
}

// TestSecretRefResolver covers all three branches of secretRefResolver.Resolve:
// no provider wired, a malformed ref (ParseRef error), and a successful resolve.
func TestSecretRefResolver(t *testing.T) {
	t.Run("no provider wired errors", func(t *testing.T) {
		r := secretRefResolver{p: nil}
		if _, err := r.Resolve(context.Background(), "secretref://env/DB_DSN"); err == nil {
			t.Fatal("want error when no secrets provider is wired")
		}
	})

	t.Run("malformed ref surfaces ParseRef error", func(t *testing.T) {
		r := secretRefResolver{p: stubProvider{val: "x"}}
		if _, err := r.Resolve(context.Background(), "not-a-ref"); err == nil {
			t.Fatal("want ParseRef error for a non-secretref string")
		}
	})

	t.Run("valid ref resolves via provider", func(t *testing.T) {
		r := secretRefResolver{p: stubProvider{val: "s3kr3t"}}
		got, err := r.Resolve(context.Background(), "secretref://env/DB_DSN")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if got != "s3kr3t" {
			t.Fatalf("Resolve = %q, want %q", got, "s3kr3t")
		}
	})

	t.Run("provider error propagates", func(t *testing.T) {
		want := errors.New("boom")
		r := secretRefResolver{p: stubProvider{err: want}}
		if _, err := r.Resolve(context.Background(), "secretref://env/DB_DSN"); !errors.Is(err, want) {
			t.Fatalf("Resolve err = %v, want %v", err, want)
		}
	})
}

func denialActor(tenant uuid.UUID) authz.Actor {
	return authz.Actor{
		Kind:       authz.ActorUser,
		UserID:     uuid.New(),
		CapacityID: uuid.New(),
		TenantID:   tenant,
	}
}

// TestLoggingAuditNilLog covers the nil-logger early return: a denial sink with
// no logger must not panic and must emit nothing.
func TestLoggingAuditNilLog(t *testing.T) {
	a := loggingAudit{log: nil}
	// Must not panic.
	a.AuthzDenial(context.Background(), denialActor(uuid.New()), "x.y.read", authz.Target{Scope: authz.ScopeTenant}, "no_grant")
}

// TestLoggingAuditWritesWarn covers the logging path including the impersonation
// field derivation.
func TestLoggingAuditWritesWarn(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	a := loggingAudit{log: log}

	actor := denialActor(uuid.New())
	actor.ImpersonatorUserID = uuid.New() // != UserID → impersonating=true branch
	a.AuthzDenial(context.Background(), actor, "x.y.read", authz.Target{Scope: authz.ScopeTenant}, "no_grant")

	out := buf.String()
	if !strings.Contains(out, "authz denial") {
		t.Fatalf("missing warn line: %q", out)
	}
	if !strings.Contains(out, "impersonating=true") {
		t.Fatalf("expected impersonating=true in: %q", out)
	}
}

// failTxM is a TxManager whose write transaction always fails, to drive the
// durableAudit error-logging branch.
type failTxM struct{ err error }

func (f failTxM) WithTenant(_ context.Context, _ func(context.Context, database.TenantDB) error) error {
	return f.err
}

func (f failTxM) WithTenantRO(_ context.Context, _ func(context.Context, database.TenantDB) error) error {
	return f.err
}

func (f failTxM) Platform(_ context.Context, _ func(context.Context, database.DB) error) error {
	return f.err
}

// TestDurableAuditEarlyReturn covers the guard that skips the durable write when
// there is no TxManager (still logs the WARN, never attempts a durable write).
func TestDurableAuditEarlyReturn(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	writer := kaudit.New(model.UUIDv7(), nil)

	a := durableAudit{log: log, txm: nil, writer: writer}
	a.AuthzDenial(context.Background(), denialActor(uuid.New()), "x.y.read", authz.Target{Scope: authz.ScopeTenant}, "no_grant")

	out := buf.String()
	if !strings.Contains(out, "authz denial") {
		t.Fatalf("expected the denial WARN, got: %q", out)
	}
	if strings.Contains(out, "durable audit write failed") {
		t.Fatalf("must not attempt a durable write with a nil TxManager: %q", out)
	}
}

// TestDurableAuditNilTenant covers the guard for an actor with no tenant: even
// with a TxManager and writer present, an unattributable denial is not durably
// written.
func TestDurableAuditNilTenant(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	writer := kaudit.New(model.UUIDv7(), nil)

	a := durableAudit{log: log, txm: failTxM{err: errors.New("should not be called")}, writer: writer}
	actor := denialActor(uuid.Nil) // TenantID == uuid.Nil → skip durable write
	a.AuthzDenial(context.Background(), actor, "x.y.read", authz.Target{Scope: authz.ScopeTenant}, "no_grant")

	if strings.Contains(buf.String(), "durable audit write failed") {
		t.Fatalf("must not attempt durable write without a tenant: %q", buf.String())
	}
}

// TestDurableAuditWriteError covers the error-logging branch: a failing durable
// write is logged (best-effort) and never panics.
func TestDurableAuditWriteError(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	writer := kaudit.New(model.UUIDv7(), nil)

	a := durableAudit{log: log, txm: failTxM{err: errors.New("db down")}, writer: writer}
	a.AuthzDenial(context.Background(), denialActor(uuid.New()), "x.y.read", authz.Target{Scope: authz.ScopeTenant}, "no_grant")

	if !strings.Contains(buf.String(), "durable audit write failed") {
		t.Fatalf("expected the durable-write failure to be logged, got: %q", buf.String())
	}
}

// TestNewDurableAuditWithTx covers New's durable-audit branch: when a TxManager
// is present (but no injected sink), the durable sink is selected. Also asserts
// the metrics/tracer NoOp defaults and off-by-default caching.
func TestNewDurableAuditWithTx(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := New(config.Defaults(), log, Deps{Tx: failTxM{err: errors.New("x")}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := k.auditSink.(durableAudit); !ok {
		t.Fatalf("auditSink = %T, want durableAudit when a TxManager is present", k.auditSink)
	}
	if k.Metrics == nil {
		t.Fatal("Metrics must default to a NoOp sink, never nil")
	}
	if k.Tracer == nil {
		t.Fatal("Tracer must default to a NoOp tracer, never nil")
	}
	if k.Documents != nil {
		t.Fatal("Documents must be nil when no storage adapter is wired")
	}
	if k.AuthzCache != nil {
		t.Fatal("AuthzCache must be nil when AuthzCacheTTL is zero")
	}
}

func TestNewRejectsTypedNilWebhookSender(t *testing.T) {
	var sender *typedNilKernelWebhookSender
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := New(config.Defaults(), log, Deps{Tx: failTxM{err: errors.New("x")}, WebhookSender: sender})
	if err == nil || k != nil || !strings.Contains(err.Error(), "WebhookSender must not be typed nil") {
		t.Fatalf("New(typed-nil WebhookSender) = (%v, %v), want nil kernel and explicit error", k, err)
	}
}

// TestNewStorageWiresDocuments covers New's Storage != nil branch: providing an
// object-storage adapter wires the document service. (A non-nil TxManager is
// required — the workflow runtime is built unconditionally.)
func TestNewStorageWiresDocuments(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := New(config.Defaults(), log, Deps{
		Tx:      failTxM{err: errors.New("x")},
		Storage: storage.NewMemory(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if k.Documents == nil {
		t.Fatal("Documents must be wired when a storage adapter is provided")
	}
}
