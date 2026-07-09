package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// DecodeJSON strict-decodes the request body into T: unknown fields are
// rejected (typo defense), the body is size-capped, and exactly one JSON value
// is required. All failures map to KindValidation — a malformed body is a
// client error, never a 500. The raw body content is never echoed into the
// error (redaction discipline).
func DecodeJSON[T any](r *http.Request, maxBytes int64) (T, error) {
	var out T
	if r.Body == nil {
		return out, kerr.E(kerr.KindValidation, "validation_failed", "request body is required")
	}
	limited := http.MaxBytesReader(nil, r.Body, maxBytes)
	dec := json.NewDecoder(limited)
	dec.DisallowUnknownFields()

	// Decode into a pointer so a literal `null` body decodes to nil and is
	// rejected like an empty body (review finding ARCH-29) rather than
	// silently yielding a zero-value struct.
	var ptr *T
	if err := dec.Decode(&ptr); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return out, kerr.E(kerr.KindValidation, "validation_failed", "request body too large")
		}
		// Deliberately generic: never surface parser internals or body bytes.
		return out, kerr.E(kerr.KindValidation, "validation_failed", "request body is not valid JSON")
	}
	if ptr == nil {
		return out, kerr.E(kerr.KindValidation, "validation_failed", "request body is required")
	}
	// Reject trailing data after the first JSON value.
	if dec.More() {
		return out, kerr.E(kerr.KindValidation, "validation_failed", "request body must contain a single JSON object")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return out, kerr.E(kerr.KindValidation, "validation_failed", "request body must contain a single JSON object")
	}
	return *ptr, nil
}

// BindAndValidate decodes the body (strict) and then runs struct-tag
// validation, returning KindValidation with field errors on failure
// (blueprint 05 §1). maxBytes bounds the body.
func BindAndValidate[T any](r *http.Request, v *validation.Validator, maxBytes int64) (T, error) {
	out, err := DecodeJSON[T](r, maxBytes)
	if err != nil {
		return out, err
	}
	// StructCtx localizes field messages against the locale/catalog the
	// httpx.Locale middleware bound to r.Context(); with no catalog wired it is
	// byte-identical to v.Struct (English). Field paths and codes stay stable.
	if err := v.StructCtx(r.Context(), out); err != nil {
		return out, err
	}
	return out, nil
}
