// Package errors is wowapi's error taxonomy: a closed set of Kinds that map
// deterministically to HTTP status codes and stable machine codes, plus the
// structured Error type carried across every layer. The HTTP layer (kernel/
// httpx) translates an *Error into an RFC 9457 problem-details body; anything
// that is not an *Error becomes an opaque 500 whose detail never reaches the
// wire. Contract: docs/blueprint/04 §5.
//
// The taxonomy is intentionally closed — new failure modes pick an existing
// Kind rather than inventing wire contracts ad hoc.
package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

// Kind is the closed set of error categories. Each maps to exactly one HTTP
// status and one machine code (see mapping below).
type Kind int

const (
	KindInternal Kind = iota // default zero value: unmapped → 500
	KindValidation
	KindUnauthenticated
	KindForbidden
	KindTenantIsolation
	KindNotFound
	KindConflict
	KindVersionConflict
	KindIdempotencyInFlight
	KindRuleViolation
	KindWorkflowState
	KindRateLimited
	KindExternal
	// KindIdempotencyExpired: a request presented an idempotency key whose
	// stored record has expired, so the original response can no longer be
	// replayed. Returned instead of silently re-executing the operation
	// (roadmap S5). Appended last to keep the earlier iota values stable.
	KindIdempotencyExpired
)

// FieldError is one shape-validation failure, addressed by JSON path.
type FieldError struct {
	Field   string `json:"field"`   // JSON path, e.g. "contacts[0].email"
	Code    string `json:"code"`    // "required", "max_length", "invalid_format"
	Message string `json:"message"` // safe for users
}

type kindInfo struct {
	code   string
	status int
}

// mapping is the single source of truth for Kind → (code, HTTP status).
// KindTenantIsolation is masked as 404 so a cross-tenant probe cannot learn a
// row exists (04 §5).
var mapping = map[Kind]kindInfo{
	KindValidation:          {"validation_failed", http.StatusBadRequest},
	KindUnauthenticated:     {"unauthenticated", http.StatusUnauthorized},
	KindForbidden:           {"permission_denied", http.StatusForbidden},
	KindTenantIsolation:     {"tenant_mismatch", http.StatusNotFound},
	KindNotFound:            {"not_found", http.StatusNotFound},
	KindConflict:            {"conflict", http.StatusConflict},
	KindVersionConflict:     {"version_conflict", http.StatusPreconditionFailed},
	KindIdempotencyInFlight: {"retry_later", http.StatusConflict},
	KindRuleViolation:       {"rule_violation", http.StatusUnprocessableEntity},
	KindWorkflowState:       {"invalid_transition", http.StatusConflict},
	KindRateLimited:         {"rate_limited", http.StatusTooManyRequests},
	KindExternal:            {"upstream_error", http.StatusBadGateway},
	KindIdempotencyExpired:  {"idempotency_key_expired", http.StatusGone},
	KindInternal:            {"internal", http.StatusInternalServerError},
}

func (k Kind) info() kindInfo {
	if info, ok := mapping[k]; ok {
		return info
	}
	return mapping[KindInternal]
}

// DefaultCode returns the taxonomy's machine code for k.
func (k Kind) DefaultCode() string { return k.info().code }

// HTTPStatus returns the HTTP status k maps to.
func (k Kind) HTTPStatus() int { return k.info().status }

// Error is the structured error carried across layers. Msg is user-safe; Op
// and the wrapped Err are for logs only and never reach the wire.
type Error struct {
	Kind   Kind
	Code   string // stable machine code; defaults to Kind.DefaultCode()
	Msg    string // safe, user-facing
	Op     string // "requests.Service.Approve" — logs only
	Fields []FieldError
	Err    error // wrapped cause (%w)
}

// E constructs an *Error. msg is treated as a plain string (not a format
// string) so caller-supplied values can never turn into format verbs; use
// fmt.Sprintf at the call site if you need interpolation. args, if given,
// may be a single wrapped error and/or an Op string:
//
//	errors.E(KindNotFound, "not_found", "request not found")
//	errors.E(KindInternal, "internal", "load failed", cause, Op("svc.Load"))
func E(kind Kind, code, msg string, args ...any) *Error {
	e := &Error{Kind: kind, Code: code, Msg: msg}
	if e.Code == "" {
		e.Code = kind.DefaultCode()
	}
	for _, a := range args {
		switch v := a.(type) {
		case opString:
			e.Op = string(v)
		case []FieldError:
			e.Fields = v
		case error:
			e.Err = v
		}
	}
	return e
}

// opString distinguishes an Op argument to E from a wrapped error.
type opString string

// Op tags an *Error with the operation name for logs.
func Op(op string) opString { return opString(op) }

// Validation builds a KindValidation error carrying field errors.
func Validation(msg string, fields ...FieldError) *Error {
	return &Error{Kind: KindValidation, Code: KindValidation.DefaultCode(), Msg: msg, Fields: fields}
}

func (e *Error) Error() string {
	var b []byte
	if e.Op != "" {
		b = append(b, e.Op...)
		b = append(b, ": "...)
	}
	b = append(b, e.Msg...)
	if e.Err != nil {
		b = append(b, ": "...)
		b = append(b, e.Err.Error()...)
	}
	return string(b)
}

// Unwrap exposes the wrapped cause to errors.Is/As.
func (e *Error) Unwrap() error { return e.Err }

// KindOf extracts the taxonomy Kind for any error: the nearest wrapped *Error's
// Kind, or KindInternal when none is present (unknown errors are 500s).
func KindOf(err error) Kind {
	var e *Error
	if stderrors.As(err, &e) {
		return e.Kind
	}
	return KindInternal
}

// As is a convenience wrapper over errors.As for *Error.
func As(err error) (*Error, bool) {
	var e *Error
	ok := stderrors.As(err, &e)
	return e, ok
}

// Wrapf wraps err with an operation prefix while preserving the taxonomy Kind
// of an underlying *Error (or KindInternal). It is the "every layer wraps"
// convention from 04 §5 without flattening the Kind.
func Wrapf(err error, op, format string, a ...any) *Error {
	if err == nil {
		return nil
	}
	k := KindOf(err)
	return &Error{Kind: k, Code: k.DefaultCode(), Msg: fmt.Sprintf(format, a...), Op: op, Err: err}
}
