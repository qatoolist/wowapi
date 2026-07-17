// Package notify preserves the v1 import path for the notification foundation.
package notify

import (
	"github.com/qatoolist/wowapi/v2/foundation/notify"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
)

type (
	Channel         = notify.Channel
	ChannelDest     = notify.ChannelDest
	ChannelSender   = notify.ChannelSender
	Delivery        = notify.Delivery
	DeliveryReceipt = notify.DeliveryReceipt
	FakeSender      = notify.FakeSender
	Importance      = notify.Importance
	Message         = notify.Message
	Notification    = notify.Notification
	Option          = notify.Option
	Registry        = notify.Registry
	Service         = notify.Service
	TemplateSpec    = notify.TemplateSpec
)

const (
	ChannelInApp        = notify.ChannelInApp
	ChannelEmail        = notify.ChannelEmail
	ChannelSMS          = notify.ChannelSMS
	ChannelWhatsApp     = notify.ChannelWhatsApp
	ChannelPush         = notify.ChannelPush
	ImportanceNormal    = notify.ImportanceNormal
	ImportanceImportant = notify.ImportanceImportant
	ImportanceLegal     = notify.ImportanceLegal
)

func NewRegistry() *Registry                            { return notify.NewRegistry() }
func ValidateBody(spec TemplateSpec, body string) error { return notify.ValidateBody(spec, body) }
func RenderBody(spec TemplateSpec, channel Channel, body string, vars map[string]any) (string, error) {
	return notify.RenderBody(spec, channel, body, vars)
}
func WithTracer(tr observability.Tracer) Option { return notify.WithTracer(tr) }
func New(reg *Registry, idgen model.IDGen, opts ...Option) *Service {
	return notify.New(reg, idgen, opts...)
}
