package outbox

import (
	"context"
	"time"
)

// SetRequeueHook overrides the relay's failed-event requeue (test-only): the
// F-07 starvation regression wraps the REAL RequeueFailed with a short cooldown
// so a row becomes due only after sustained draining has begun.
func SetRequeueHook(r *Relay, fn func(ctx context.Context, cooldown time.Duration) error) {
	if r.hooks == nil {
		r.hooks = &relayTestHooks{}
	}
	r.hooks.requeue = fn
}
