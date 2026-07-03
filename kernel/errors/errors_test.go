package errors_test

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
)

func TestKindMappingClosedSet(t *testing.T) {
	cases := []struct {
		kind   errors.Kind
		code   string
		status int
	}{
		{errors.KindValidation, "validation_failed", 400},
		{errors.KindUnauthenticated, "unauthenticated", 401},
		{errors.KindForbidden, "permission_denied", 403},
		{errors.KindTenantIsolation, "tenant_mismatch", 404},
		{errors.KindNotFound, "not_found", 404},
		{errors.KindConflict, "conflict", 409},
		{errors.KindVersionConflict, "version_conflict", 412},
		{errors.KindIdempotencyInFlight, "retry_later", 409},
		{errors.KindRuleViolation, "rule_violation", 422},
		{errors.KindWorkflowState, "invalid_transition", 409},
		{errors.KindRateLimited, "rate_limited", 429},
		{errors.KindExternal, "upstream_error", 502},
		{errors.KindInternal, "internal", 500},
	}
	for _, c := range cases {
		if got := c.kind.DefaultCode(); got != c.code {
			t.Errorf("kind %d code = %q, want %q", c.kind, got, c.code)
		}
		if got := c.kind.HTTPStatus(); got != c.status {
			t.Errorf("kind %d status = %d, want %d", c.kind, got, c.status)
		}
	}
}

func TestTenantIsolationMaskedAs404(t *testing.T) {
	if errors.KindTenantIsolation.HTTPStatus() != http.StatusNotFound {
		t.Fatal("tenant isolation must be masked as 404 to avoid leaking existence")
	}
}

func TestUnknownKindIsInternal(t *testing.T) {
	var k errors.Kind = 9999
	if k.HTTPStatus() != 500 || k.DefaultCode() != "internal" {
		t.Fatalf("unmapped kind must fall back to internal 500, got %d/%s", k.HTTPStatus(), k.DefaultCode())
	}
}

func TestEDefaultsCodeFromKind(t *testing.T) {
	e := errors.E(errors.KindNotFound, "", "gone")
	if e.Code != "not_found" {
		t.Errorf("empty code should default from kind: %q", e.Code)
	}
}

func TestEMsgIsNotAFormatString(t *testing.T) {
	// A user-supplied value containing verbs must not be interpreted.
	e := errors.E(errors.KindValidation, "validation_failed", "bad value: 100%s of items")
	if e.Msg != "bad value: 100%s of items" {
		t.Errorf("msg was reinterpreted as a format string: %q", e.Msg)
	}
}

func TestEWrapsCauseAndOp(t *testing.T) {
	cause := stderrors.New("db exploded")
	e := errors.E(errors.KindInternal, "internal", "load failed", cause, errors.Op("svc.Load"))
	if !stderrors.Is(e, cause) {
		t.Error("wrapped cause must be reachable via errors.Is")
	}
	if e.Op != "svc.Load" {
		t.Errorf("Op = %q", e.Op)
	}
	if got := e.Error(); got != "svc.Load: load failed: db exploded" {
		t.Errorf("Error() = %q", got)
	}
}

func TestKindOfThroughWrapping(t *testing.T) {
	base := errors.E(errors.KindVersionConflict, "version_conflict", "stale")
	wrapped := fmt.Errorf("service layer: %w", base)
	if errors.KindOf(wrapped) != errors.KindVersionConflict {
		t.Error("KindOf must see through fmt.Errorf wrapping")
	}
	if errors.KindOf(stderrors.New("plain")) != errors.KindInternal {
		t.Error("a plain error must be treated as internal")
	}
}

func TestValidationHelper(t *testing.T) {
	e := errors.Validation("invalid", errors.FieldError{Field: "email", Code: "required", Message: "required"})
	if e.Kind != errors.KindValidation || len(e.Fields) != 1 {
		t.Fatalf("unexpected: %+v", e)
	}
}

func TestWrapfPreservesKind(t *testing.T) {
	base := errors.E(errors.KindNotFound, "not_found", "missing")
	w := errors.Wrapf(base, "repo.Get", "loading id %d", 7)
	if w.Kind != errors.KindNotFound {
		t.Errorf("Wrapf flattened the kind: %v", w.Kind)
	}
	if !stderrors.Is(w, base) {
		t.Error("Wrapf must preserve the chain")
	}
	if errors.Wrapf(nil, "x", "y") != nil {
		t.Error("Wrapf(nil) must be nil")
	}
}
