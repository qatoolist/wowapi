package config

import (
	"strings"
	"testing"
)

// TestConcurrencyDefaultsValidate proves the compiled defaults (disabled
// limiter, advisory capacity mode) validate cleanly on their own — a
// zero-touch deployment must never fail boot because of this feature
// (backlog B6 rollout guard: advisory-then-enforced).
func TestConcurrencyDefaultsValidate(t *testing.T) {
	f := Defaults()
	if err := f.Validate(); err != nil {
		t.Fatalf("Defaults() with Concurrency must validate: %v", err)
	}
	if f.Concurrency.CapacityMode != CapacityModeAdvisory {
		t.Errorf("Concurrency.CapacityMode default = %q, want advisory", f.Concurrency.CapacityMode)
	}
	if f.Concurrency.HTTPMaxInFlight != 0 {
		t.Errorf("Concurrency.HTTPMaxInFlight default = %d, want 0 (disabled unless configured)", f.Concurrency.HTTPMaxInFlight)
	}
	if f.Concurrency.Overload.Status != 503 {
		t.Errorf("Concurrency.Overload.Status default = %d, want 503", f.Concurrency.Overload.Status)
	}
	if f.Concurrency.Overload.RetryAfter <= 0 {
		t.Errorf("Concurrency.Overload.RetryAfter default must be > 0, got %v", f.Concurrency.Overload.RetryAfter)
	}
}

// TestConcurrencyFieldValidation covers the basic per-field range checks:
// negative/zero worker caps, invalid overload status, and non-positive
// retry-after are all rejected regardless of capacity mode.
func TestConcurrencyFieldValidation(t *testing.T) {
	f := Defaults()
	f.Concurrency.HTTPMaxInFlight = -1
	f.Concurrency.WorkerMaxJobs = -5
	f.Concurrency.PlatformMaxInFlight = -2
	f.Concurrency.Replicas = -1
	f.Concurrency.RuntimePoolMax = -1
	f.Concurrency.PlatformPoolMax = -1
	f.Concurrency.MigratePoolMax = -1
	f.Concurrency.ReservedAdmin = -1
	f.Concurrency.Overload.Status = 200
	f.Concurrency.Overload.RetryAfter = 0
	err := f.Validate()
	if err == nil {
		t.Fatal("negative caps / bad overload status must fail validation")
	}
	for _, want := range []string{
		"concurrency.http_max_in_flight:",
		"concurrency.worker_max_jobs:",
		"concurrency.platform_max_in_flight:",
		"concurrency.replicas:",
		"concurrency.runtime_pool_max:",
		"concurrency.platform_pool_max:",
		"concurrency.migrate_pool_max:",
		"concurrency.reserved_admin:",
		"concurrency.overload.status:",
		"concurrency.overload.retry_after:",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}
}

// TestConcurrencyCapacityModeInvalid proves an unrecognized capacity mode
// string is rejected.
func TestConcurrencyCapacityModeInvalid(t *testing.T) {
	f := Defaults()
	f.Concurrency.CapacityMode = "yolo"
	err := f.Validate()
	if err == nil || !strings.Contains(err.Error(), "concurrency.capacity_mode:") {
		t.Fatalf("invalid capacity_mode must fail validation, got: %v", err)
	}
}

// TestCapacityBudgetOversubscribed proves the deployment-shape formula
// (backlog B6 / benchmark §Concurrency):
//
//	replicas*(runtime_pool_max+platform_pool_max) + migrate_pool_max + reserved_admin <= db_max_connections
//
// flags an oversubscribed shape, and that ADVISORY mode reports the problem
// as a warning without failing Validate, while ENFORCED mode fails closed.
func TestCapacityBudgetOversubscribed(t *testing.T) {
	f := Defaults()
	f.DB.MaxConns = 20
	f.Concurrency.Replicas = 3
	f.Concurrency.RuntimePoolMax = 16
	f.Concurrency.PlatformPoolMax = 8
	f.Concurrency.MigratePoolMax = 2
	f.Concurrency.ReservedAdmin = 4
	// 3*(16+8) + 2 + 4 = 78 > 20 → oversubscribed.

	// Advisory (default): Validate must still succeed, but CheckCapacity
	// must report the problem so callers (CLI, boot warnings) can surface it.
	f.Concurrency.CapacityMode = CapacityModeAdvisory
	if err := f.Validate(); err != nil {
		t.Fatalf("advisory mode must not fail Validate on an oversubscribed shape: %v", err)
	}
	warn := CheckCapacity(f)
	if warn == nil {
		t.Fatal("CheckCapacity must flag an oversubscribed shape")
	}
	if !strings.Contains(warn.Error(), "78") || !strings.Contains(warn.Error(), "20") {
		t.Errorf("capacity problem should cite computed demand and db_max_connections: %v", warn)
	}

	// Enforced: Validate must fail closed on the same oversubscribed shape.
	f.Concurrency.CapacityMode = CapacityModeEnforced
	err := f.Validate()
	if err == nil {
		t.Fatal("enforced mode must fail Validate on an oversubscribed shape")
	}
	if !strings.Contains(err.Error(), "concurrency: capacity budget exceeded") {
		t.Errorf("enforced validation error missing capacity message: %v", err)
	}
}

// TestCapacityBudgetWithinBounds proves a correctly-sized shape passes in
// both advisory and enforced mode, and CheckCapacity returns nil (no
// warning) when there is no problem.
func TestCapacityBudgetWithinBounds(t *testing.T) {
	f := Defaults()
	f.DB.MaxConns = 100
	f.Concurrency.Replicas = 2
	f.Concurrency.RuntimePoolMax = 16
	f.Concurrency.PlatformPoolMax = 8
	f.Concurrency.MigratePoolMax = 2
	f.Concurrency.ReservedAdmin = 4
	// 2*(16+8) + 2 + 4 = 54 <= 100 → fine.

	if warn := CheckCapacity(f); warn != nil {
		t.Errorf("within-budget shape must not warn: %v", warn)
	}

	f.Concurrency.CapacityMode = CapacityModeEnforced
	if err := f.Validate(); err != nil {
		t.Fatalf("within-budget shape must validate under enforced mode: %v", err)
	}
}

// TestCapacityBudgetZeroReplicasSkipsCheck proves the budget check is a
// no-op when Replicas is left at its zero value (not configured) — the
// formula is undefined without a deployment shape, so it must not produce
// spurious failures for products that haven't opted into this yet.
func TestCapacityBudgetZeroReplicasSkipsCheck(t *testing.T) {
	f := Defaults()
	f.DB.MaxConns = 2 // absurdly small, would fail if the formula ran
	if warn := CheckCapacity(f); warn != nil {
		t.Errorf("zero Replicas (not configured) must skip the capacity check, got: %v", warn)
	}
	f.Concurrency.CapacityMode = CapacityModeEnforced
	if err := f.Validate(); err != nil {
		t.Fatalf("zero Replicas must not fail enforced validation: %v", err)
	}
}
