package app

import (
	"context"

	"github.com/qatoolist/wowapi/v2/kernel/config"
)

// CapturedKernelConfig exposes the boot-captured kernel view's config for the
// deep-isolation regression. Compiled only into test binaries; the aggregate
// pointer itself stays unexported (fifth closure audit 2026-07-17).
func CapturedKernelConfig(b *Booted) config.Framework {
	return b.runtime.kernel.Cfg
}

// CapturedHealth exposes the boot-validated health-check set for tests.
func CapturedHealth(b *Booted) map[string]func(context.Context) error {
	return b.runtimeHealth()
}

// CapturedRecurring exposes the boot-validated recurring jobs for tests.
func CapturedRecurring(b *Booted) []RecurringJob {
	return b.runtimeRecurring()
}
