package i18n

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
)

// This golden test is the guard that migrating the framework's English strings
// out of hardcoded Go maps and into embedded per-locale YAML (B1) did NOT change
// a single shipped message. The maps below are the frozen historical values that
// kernel/httpx and kernel/validation produced before the YAML migration; the
// test asserts the embedded-YAML-loaded framework catalog reproduces them
// byte-for-byte. If you intentionally change a framework English string, update
// BOTH the YAML and this golden map in the same commit.

var goldenProblemTitles = map[errors.Kind]string{
	errors.KindValidation:          "Validation failed",
	errors.KindUnauthenticated:     "Authentication required",
	errors.KindForbidden:           "Permission denied",
	errors.KindTenantIsolation:     "Not found",
	errors.KindNotFound:            "Not found",
	errors.KindConflict:            "Conflict",
	errors.KindVersionConflict:     "Version conflict",
	errors.KindIdempotencyInFlight: "Retry later",
	errors.KindRuleViolation:       "Rule violation",
	errors.KindWorkflowState:       "Invalid transition",
	errors.KindRateLimited:         "Rate limited",
	errors.KindExternal:            "Upstream error",
	errors.KindInternal:            "Internal error",
}

var goldenValidationMessages = map[string]string{
	"required": "this field is required",
	"email":    "must be a valid email address",
	"min":      "must be at least %s",
	"max":      "must be at most %s",
	"len":      "must be exactly %s characters long",
	"oneof":    "must be one of: %s",
	"uuid":     "must be a valid UUID",
	"gte":      "must be at least %s",
	"lte":      "must be at most %s",
}

var goldenDetails = map[string]string{
	"validation_failed": "validation failed",
}

func TestFrameworkYAMLMatchesGoldenMaps(t *testing.T) {
	cat := NewRegistry().Catalog()

	for kind, want := range goldenProblemTitles {
		key := KeyProblemTitle(kind)
		got, loc := cat.Lookup(DefaultLocale, key)
		if got != want || loc != DefaultLocale {
			t.Errorf("problem title %v (%s): got (%q,%q), want (%q,en)", kind, key, got, loc, want)
		}
	}
	for tag, want := range goldenValidationMessages {
		key := KeyValidationMessage(tag)
		got, _ := cat.Lookup(DefaultLocale, key)
		if got != want {
			t.Errorf("validation %s (%s): got %q, want %q", tag, key, got, want)
		}
	}
	for code, want := range goldenDetails {
		key := KeyDetail(code)
		got, _ := cat.Lookup(DefaultLocale, key)
		if got != want {
			t.Errorf("detail %s (%s): got %q, want %q", code, key, got, want)
		}
	}
}

// TestFrameworkYAMLHasNoExtraKeys guards the other direction: the embedded YAML
// must not ship a kernel.* key that no golden entry accounts for (a stray key
// would silently change behavior for some code path).
func TestFrameworkYAMLHasNoExtraKeys(t *testing.T) {
	cat := NewRegistry().Catalog()
	want := map[string]bool{}
	for kind := range goldenProblemTitles {
		want[KeyProblemTitle(kind)] = true
	}
	for tag := range goldenValidationMessages {
		want[KeyValidationMessage(tag)] = true
	}
	for code := range goldenDetails {
		want[KeyDetail(code)] = true
	}
	for key := range cat.messages[DefaultLocale] {
		if !want[key] {
			t.Errorf("embedded framework YAML ships unexpected key %q not in the golden set", key)
		}
	}
}
