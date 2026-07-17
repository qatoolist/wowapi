// Package webhook implements wowapi's webhook subsystem: inbound signature
// verification + replay protection + async processing, and outbound signed
// HTTP delivery with per-endpoint circuit breakers.
// Contract: docs/blueprint/07 §6.
package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/safety"
)

// Direction values mirror the DB check constraint.
const (
	DirectionInbound  = "inbound"
	DirectionOutbound = "outbound"
)

// DeliveryStatus values mirror the DB check constraint.
const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusDelivered = "delivered"
	StatusFailed    = "failed"
	StatusDead      = "dead"
)

// MaxAttempts is the DLQ ceiling for both inbound processing and outbound delivery.
const MaxAttempts = 5

// TimestampWindow is the replay-protection window (±5 m per blueprint 07 §6).
const TimestampWindow = 5 * time.Minute

// OutboundTimeout is the per-delivery HTTP call timeout.
const OutboundTimeout = 10 * time.Second

// BreakerFailureThreshold is the number of consecutive delivery failures that
// opens the circuit breaker.
const BreakerFailureThreshold = 5

// BreakerCooldown is the half-open probe interval after the circuit opens.
const BreakerCooldown = 5 * time.Minute

// --- domain types ---

// Endpoint is the service-layer view of a webhook_endpoints row.
type Endpoint struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	Direction        string
	ProviderID       *uuid.UUID
	URL              *string
	SecretRef        string
	SignatureScheme  string
	SubscribedEvents []string
	Status           string
}

// Event is the service-layer view of a webhook_events row.
type Event struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	EndpointID      uuid.UUID
	Direction       string
	ExternalEventID string
	EventType       string
	Payload         json.RawMessage
	SignatureOk     *bool
	ReceivedAt      time.Time
	DeliveryStatus  string
	Attempts        int
	NextAttemptAt   *time.Time
	LastError       *string
}

// InboundIn is the input envelope the HTTP layer fills when it receives a
// provider webhook POST.
type InboundIn struct {
	EndpointID      uuid.UUID
	ProviderKey     string // key to look up the registered Verifier
	RawBody         []byte
	Headers         map[string]string
	ExternalEventID string
	EventType       string
	Timestamp       time.Time // provider-supplied timestamp from headers
}

// --- port interfaces ---

// Verifier verifies a provider's signature over a raw body + headers.
// Implementations are registered per provider key.
//
// Verify returns an Envelope containing only fields derived from authenticated
// data. Callers must check the error first; Envelope is undefined when error
// is non-nil. The Envelope contract is documented in the provider-verifier
// contract (see docs/blueprint/webhook-provider-verifier-contract.md).
type Verifier interface {
	Verify(secret string, body []byte, headers map[string]string) (Envelope, error)
}

// Sender delivers a signed HTTP POST to a webhook URL.
type Sender interface {
	Post(ctx context.Context, url string, body []byte, headers map[string]string) (statusCode int, err error)
}

// SecretResolver resolves a secret_ref string (as stored in the secret_ref
// column) to its plaintext value. Only the composition root wires a real
// implementation; the kernel/secrets adapter satisfies this interface.
type SecretResolver interface {
	Resolve(ctx context.Context, ref string) (string, error)
}

// InboundHandler processes a verified, persisted inbound webhook event.
// Registered per event_type; called asynchronously by ProcessInbound.
type InboundHandler func(ctx context.Context, db database.TenantDB, e Event) error

// --- Service ---

// Service is the webhook framework. HandleInbound runs on the caller's tenant
// DB (app_rt); ProcessInbound and DispatchOutbound run on a platform TxManager
// (app_platform, tenant-bound), following the document.Service pattern.
type Service struct {
	verifiers map[string]Verifier
	handlers  map[string]InboundHandler
	sender    Sender
	secrets   SecretResolver
	breaker   *breakerRegistry
	idgen     model.IDGen
	now       func() time.Time
	metrics   observability.Metrics
}

// Option customizes a Service at construction.
type Option func(*Service)

// WithMetrics wires an observability sink so the outbound breaker state is
// exported as the webhook_breaker_state gauge (0=closed, 1=open, 2=half-open).
func WithMetrics(m observability.Metrics) Option {
	return func(s *Service) {
		if m != nil {
			s.metrics = m
		}
	}
}

// WithClock supplies the service clock. It is primarily useful for
// deterministic tests; production composition should use New's real clock.
func WithClock(now func() time.Time) Option {
	if now == nil {
		panic("webhook.WithClock: clock is required")
	}
	return func(s *Service) { s.now = now }
}

// New wires the Service. sender, secrets, and idgen are required.
func New(sender Sender, secrets SecretResolver, idgen model.IDGen, opts ...Option) *Service {
	if sender == nil || secrets == nil || idgen == nil {
		panic("webhook.New: sender, secrets, and idgen are required")
	}
	return newService(sender, secrets, idgen, time.Now, opts...)
}

func newService(sender Sender, secrets SecretResolver, idgen model.IDGen, nowFn func() time.Time, opts ...Option) *Service {
	if _, ok := sender.(safety.Declarer); !ok {
		panic(fmt.Sprintf("webhook.newService: sender %T must implement safety.Declarer", sender))
	}
	s := &Service{
		verifiers: make(map[string]Verifier),
		handlers:  make(map[string]InboundHandler),
		sender:    sender,
		secrets:   secrets,
		breaker:   newBreakerRegistry(),
		idgen:     idgen,
		now:       nowFn,
		metrics:   observability.NoOp,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// emitBreakerState exports the endpoint's circuit-breaker state as a gauge
// (0=closed, 1=open, 2=half-open) after a delivery outcome (roadmap CA-1). NoOp
// unless a metrics adapter is wired.
func (s *Service) emitBreakerState(endpointID uuid.UUID, br *breakerState) {
	s.metrics.SetGauge("webhook_breaker_state", br.stateValue(s.now()),
		map[string]string{"endpoint_id": endpointID.String()})
}

// RegisterVerifier registers a Verifier for the given provider key.
// Call before serving requests.
func (s *Service) RegisterVerifier(providerKey string, v Verifier) {
	s.verifiers[providerKey] = v
}

// RegisterHandler registers an InboundHandler for the given event type.
// Only one handler per event type; call before serving.
func (s *Service) RegisterHandler(eventType string, h InboundHandler) {
	s.handlers[eventType] = h
}
