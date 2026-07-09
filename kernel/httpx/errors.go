package httpx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/qatoolist/wowapi/kernel/errors"
)

// ProblemError is the RFC 9457 problem-details body — the ONLY error shape the
// API emits (blueprint 04 §4). It never carries internal detail: Op, wrapped
// causes, and stack traces stay in logs.
type ProblemError struct {
	Type      string              `json:"type"`
	Title     string              `json:"title"`
	Status    int                 `json:"status"`
	Detail    string              `json:"detail,omitempty"`
	Instance  string              `json:"instance,omitempty"`
	Code      string              `json:"code"`
	RequestID string              `json:"request_id"`
	Errors    []errors.FieldError `json:"errors,omitempty"`
}

// problemTypeBase is the URI prefix for the machine-readable problem type.
const problemTypeBase = "https://errors.wowapi.dev/"

// titles are short, safe, human titles per Kind — the English fallback used
// when no i18n catalog is bound to the request context (zero-config) or a
// locale has no translation for a title. Localized rendering pulls from the
// framework catalog via localizeTitle (kernel/i18n keys the SAME English
// strings under KeyProblemTitle, so this map and the catalog stay in lockstep).
// Never derived from the (potentially sensitive) error message.
var titles = map[errors.Kind]string{
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

// WriteError translates any error into a problem-details response. An *Error
// contributes its Kind (→ status/code), user-safe Msg, and field errors; any
// other error is rendered as an opaque 500 whose cause never reaches the wire
// (blueprint 04 §5). Observability logging is wired by the recover/log
// middleware; this function only shapes the response.
func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	reqID := RequestIDFrom(ctx)

	e, ok := errors.As(err)
	if !ok {
		writeProblem(w, ProblemError{
			Type:      problemTypeBase + "internal",
			Title:     localizeTitle(ctx, errors.KindInternal, titles[errors.KindInternal]),
			Status:    http.StatusInternalServerError,
			Code:      errors.KindInternal.DefaultCode(),
			RequestID: reqID,
		})
		return
	}

	kind := e.Kind
	status := kind.HTTPStatus()
	code := e.Code
	if code == "" {
		code = kind.DefaultCode()
	}
	p := ProblemError{
		Type:      problemTypeBase + kind.DefaultCode(),
		Title:     localizeTitle(ctx, kind, titles[kind]),
		Status:    status,
		Code:      code,
		RequestID: reqID,
		Errors:    e.Fields,
	}
	// Internal errors never expose their message; everything else exposes the
	// deliberately user-safe Msg only (never Op or the wrapped cause). Detail is
	// the user-safe Msg the producing layer already localized (or left English);
	// WriteError does not translate it — kernel/validation localizes field
	// messages at production time, and other callers pass locale-appropriate Msg.
	if kind != errors.KindInternal {
		p.Detail = e.Msg
	}
	if p.Title == "" {
		p.Title = localizeTitle(ctx, errors.KindInternal, titles[errors.KindInternal])
	}
	writeProblem(w, p)
}

func writeInternal(ctx context.Context, w http.ResponseWriter) {
	writeProblem(w, ProblemError{
		Type:      problemTypeBase + "internal",
		Title:     localizeTitle(ctx, errors.KindInternal, titles[errors.KindInternal]),
		Status:    http.StatusInternalServerError,
		Code:      errors.KindInternal.DefaultCode(),
		RequestID: RequestIDFrom(ctx),
	})
}

func writeProblem(w http.ResponseWriter, p ProblemError) {
	buf, err := json.Marshal(p)
	if err != nil {
		// Should be impossible; fall back to a minimal hardcoded body.
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status":500,"code":"internal","title":"Internal error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_, _ = w.Write(buf)
}
