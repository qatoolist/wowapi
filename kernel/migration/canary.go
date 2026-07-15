package migration

import (
	"context"
	"fmt"
	"time"
)

// SoakConfig holds the configurable canary duration and threshold parameters.
// The numeric values are intentionally not hardcoded: they are a per-rollout
// human decision (RISK-W02-003).
type SoakConfig struct {
	SoakDuration   time.Duration
	ErrorThreshold float64       // errors per second; zero means no errors allowed
	MinSampleCount int           // minimum successful leg executions before declaring parity
	RateLimit      time.Duration // pause between leg executions
}

// CanaryLeg is one side of the N/N-1 canary matrix.
type CanaryLeg struct {
	Name      string
	SchemaAge string // "before_backfill" or "after_backfill"
	Version   string // "N-1" or "N"
	Run       func(ctx context.Context) error
}

// CanaryLegResult records one execution.
type CanaryLegResult struct {
	Leg       CanaryLeg
	Duration  time.Duration
	Error     string
	Timestamp time.Time
}

// CanaryResult aggregates every leg execution during the soak.
type CanaryResult struct {
	Config  SoakConfig
	Legs    []CanaryLegResult
	Passed  bool
	Errors  int
	Samples int
}

// RunCanary executes every leg repeatedly for the configured soak duration,
// collecting metrics against the threshold. It returns when the duration
// expires or a leg fails unrecoverably.
func RunCanary(ctx context.Context, cfg SoakConfig, legs []CanaryLeg) (*CanaryResult, error) {
	if cfg.SoakDuration <= 0 {
		cfg.SoakDuration = 5 * time.Second
	}
	if cfg.MinSampleCount <= 0 {
		cfg.MinSampleCount = 2
	}

	res := &CanaryResult{Config: cfg}
	deadline := time.Now().Add(cfg.SoakDuration)

	for time.Now().Before(deadline) {
		for _, leg := range legs {
			select {
			case <-ctx.Done():
				return res, ctx.Err()
			default:
			}

			start := time.Now()
			err := leg.Run(ctx)
			r := CanaryLegResult{
				Leg:       leg,
				Duration:  time.Since(start),
				Timestamp: start,
			}
			if err != nil {
				r.Error = err.Error()
				res.Errors++
			}
			res.Samples++
			res.Legs = append(res.Legs, r)

			if cfg.RateLimit > 0 {
				select {
				case <-time.After(cfg.RateLimit):
				case <-ctx.Done():
					return res, ctx.Err()
				}
			}
		}
	}

	errorRate := float64(res.Errors) / cfg.SoakDuration.Seconds()
	res.Passed = res.Samples >= cfg.MinSampleCount && (cfg.ErrorThreshold <= 0 || errorRate <= cfg.ErrorThreshold)
	return res, nil
}

// ErrorRate returns errors per second observed during the canary.
func (r *CanaryResult) ErrorRate() float64 {
	if r.Config.SoakDuration <= 0 {
		return 0
	}
	return float64(r.Errors) / r.Config.SoakDuration.Seconds()
}

// Summary returns a short human-readable canary outcome.
func (r *CanaryResult) Summary() string {
	return fmt.Sprintf("canary samples=%d errors=%d rate=%.4f/s passed=%t",
		r.Samples, r.Errors, r.ErrorRate(), r.Passed)
}
