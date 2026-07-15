package requestbench

import (
	"testing"
	"time"
)

func TestReferenceWarmupDurationConfiguration(t *testing.T) {
	t.Setenv("PERF_WARMUP_DURATION", "5m")
	got, err := configuredWarmupDuration()
	if err != nil {
		t.Fatalf("configured warmup: %v", err)
	}
	if got != 5*time.Minute {
		t.Fatalf("warmup = %s, want 5m", got)
	}

	t.Setenv("PERF_WARMUP_DURATION", "not-a-duration")
	if _, err := configuredWarmupDuration(); err == nil {
		t.Fatal("invalid warmup duration accepted")
	}
}
