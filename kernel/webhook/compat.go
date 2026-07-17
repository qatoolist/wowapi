// Package webhook preserves the v1 import path for the webhook foundation.
package webhook

import (
	"time"

	"github.com/qatoolist/wowapi/v2/foundation/webhook"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
)

type (
	Endpoint           = webhook.Endpoint
	Event              = webhook.Event
	FakeSecretResolver = webhook.FakeSecretResolver
	FakeSender         = webhook.FakeSender
	FakeVerifier       = webhook.FakeVerifier
	HMACVerifier       = webhook.HMACVerifier
	HTTPSender         = webhook.HTTPSender
	InboundHandler     = webhook.InboundHandler
	InboundIn          = webhook.InboundIn
	Option             = webhook.Option
	SecretResolver     = webhook.SecretResolver
	Sender             = webhook.Sender
	SentCall           = webhook.SentCall
	Service            = webhook.Service
	Verifier           = webhook.Verifier
)

const (
	DirectionInbound        = webhook.DirectionInbound
	DirectionOutbound       = webhook.DirectionOutbound
	StatusPending           = webhook.StatusPending
	StatusProcessed         = webhook.StatusProcessed
	StatusDelivered         = webhook.StatusDelivered
	StatusFailed            = webhook.StatusFailed
	StatusDead              = webhook.StatusDead
	MaxAttempts             = webhook.MaxAttempts
	TimestampWindow         = webhook.TimestampWindow
	OutboundTimeout         = webhook.OutboundTimeout
	BreakerFailureThreshold = webhook.BreakerFailureThreshold
	BreakerCooldown         = webhook.BreakerCooldown
)

func NewHTTPSender() *HTTPSender                 { return webhook.NewHTTPSender() }
func WithMetrics(m observability.Metrics) Option { return webhook.WithMetrics(m) }
func New(sender Sender, secrets SecretResolver, idgen model.IDGen, opts ...Option) *Service {
	return webhook.New(sender, secrets, idgen, opts...)
}

func NewWithClock(sender Sender, secrets SecretResolver, idgen model.IDGen, nowFn func() time.Time, opts ...Option) *Service {
	return webhook.NewWithClock(sender, secrets, idgen, nowFn, opts...)
}
