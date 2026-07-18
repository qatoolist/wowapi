package httpx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// ---------- route metadata enforcement (blueprint 05 §1) ----------

func noop(http.ResponseWriter, *http.Request) {}

func TestRouterRejectsMissingMetadata(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/x", httpx.RouteMeta{}, noop) // neither Permission nor Public
	if r.Err() == nil {
		t.Fatal("a route with neither Permission nor Public must fail registration")
	}
	if !strings.Contains(r.Err().Error(), "neither Permission nor Public") {
		t.Errorf("unclear error: %v", r.Err())
	}
}

func TestRouterRejectsPublicWithPermission(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/x", httpx.RouteMeta{Public: true, Permission: "x.read"}, noop)
	if r.Err() == nil {
		t.Fatal("Public + Permission is contradictory and must fail registration")
	}
}

func TestRouterAccumulatesErrors(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/a", httpx.RouteMeta{}, noop)
	r.Handle(http.MethodGet, "/b", httpx.RouteMeta{Public: true, Permission: "p"}, noop)
	err := r.Err()
	if err == nil || !strings.Contains(err.Error(), "/a") || !strings.Contains(err.Error(), "/b") {
		t.Fatalf("both bad routes should be reported: %v", err)
	}
}

func TestRouterValidRoutesAndPermissions(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/health", httpx.RouteMeta{Public: true}, noop)
	r.Handle(http.MethodPost, "/things", httpx.RouteMeta{Permission: "things.create", NoRequestBody: true}, noop)
	r.Handle(http.MethodGet, "/things", httpx.RouteMeta{Permission: "things.read"}, noop)
	if err := r.Err(); err != nil {
		t.Fatalf("valid routes should register cleanly: %v", err)
	}
	perms := r.Permissions()
	if len(perms) != 2 || perms[0] != "things.create" || perms[1] != "things.read" {
		t.Errorf("permissions = %v", perms)
	}
	if len(r.Routes()) != 3 {
		t.Errorf("expected 3 routes, got %d", len(r.Routes()))
	}
}

func TestRouterRejectsDuplicate(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/dup", httpx.RouteMeta{Public: true}, noop)
	r.Handle(http.MethodGet, "/dup", httpx.RouteMeta{Public: true}, noop)
	if r.Err() == nil {
		t.Fatal("duplicate method+pattern must fail registration")
	}
}

// ---------- error mapping (blueprint 04 §5) ----------

func decodeProblem(t *testing.T, body []byte) httpx.ProblemError {
	t.Helper()
	var p httpx.ProblemError
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("problem body not JSON: %v (%s)", err, body)
	}
	return p
}

func TestWriteErrorMapsKinds(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{errors.E(errors.KindNotFound, "not_found", "gone"), 404, "not_found"},
		{errors.E(errors.KindVersionConflict, "version_conflict", "stale"), 412, "version_conflict"},
		{errors.E(errors.KindTenantIsolation, "tenant_mismatch", "x"), 404, "tenant_mismatch"},
		{errors.E(errors.KindRateLimited, "rate_limited", "slow down"), 429, "rate_limited"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		ctx := httpx.WithRequestID(context.Background(), "req-1")
		httpx.WriteError(ctx, rec, c.err)
		if rec.Code != c.status {
			t.Errorf("status = %d, want %d", rec.Code, c.status)
		}
		p := decodeProblem(t, rec.Body.Bytes())
		if p.Code != c.code || p.Status != c.status || p.RequestID != "req-1" {
			t.Errorf("problem = %+v", p)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
			t.Errorf("content-type = %q", ct)
		}
	}
}

func TestWriteErrorInternalNeverLeaks(t *testing.T) {
	// A plain (non-*Error) error and an explicit internal error must both
	// render as opaque 500s with no detail from the cause.
	for _, err := range []error{
		errors.E(errors.KindInternal, "internal", "DB password is hunter2", errors.Op("svc.X")),
		context.DeadlineExceeded,
	} {
		rec := httptest.NewRecorder()
		httpx.WriteError(context.Background(), rec, err)
		if rec.Code != 500 {
			t.Errorf("status = %d, want 500", rec.Code)
		}
		body := rec.Body.String()
		if strings.Contains(body, "hunter2") || strings.Contains(body, "svc.X") || strings.Contains(body, "DeadlineExceeded") {
			t.Errorf("internal error leaked detail: %s", body)
		}
		p := decodeProblem(t, rec.Body.Bytes())
		if p.Detail != "" {
			t.Errorf("internal problem must have empty detail, got %q", p.Detail)
		}
	}
}

func TestWriteErrorValidationCarriesFields(t *testing.T) {
	err := errors.Validation("invalid", errors.FieldError{Field: "email", Code: "required", Message: "required"})
	rec := httptest.NewRecorder()
	httpx.WriteError(context.Background(), rec, err)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
	p := decodeProblem(t, rec.Body.Bytes())
	if len(p.Errors) != 1 || p.Errors[0].Field != "email" {
		t.Errorf("field errors not carried: %+v", p.Errors)
	}
}

// ---------- strict decode ----------

type createReq struct {
	Name string `json:"name" validate:"required"`
}

func req(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
}

func TestDecodeJSONStrict(t *testing.T) {
	if _, err := httpx.DecodeJSON[createReq](req(`{"name":"ok"}`), 1<<20); err != nil {
		t.Fatalf("valid body rejected: %v", err)
	}
	// unknown field
	_, err := httpx.DecodeJSON[createReq](req(`{"name":"ok","extra":1}`), 1<<20)
	if errors.KindOf(err) != errors.KindValidation {
		t.Errorf("unknown field should be KindValidation, got %v", err)
	}
	// trailing data
	_, err = httpx.DecodeJSON[createReq](req(`{"name":"ok"}{}`), 1<<20)
	if errors.KindOf(err) != errors.KindValidation {
		t.Errorf("trailing data should be KindValidation, got %v", err)
	}
	// oversized
	big := `{"name":"` + strings.Repeat("a", 100) + `"}`
	_, err = httpx.DecodeJSON[createReq](req(big), 16)
	if errors.KindOf(err) != errors.KindValidation {
		t.Errorf("oversized body should be KindValidation, got %v", err)
	}
}

func TestDecodeJSONRejectsNull(t *testing.T) {
	// A literal null body must be rejected like an empty body (ARCH-29),
	// not silently yield a zero-value struct.
	_, err := httpx.DecodeJSON[createReq](req(`null`), 1<<20)
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("null body should be KindValidation, got %v", err)
	}
	if _, err := httpx.DecodeJSON[createReq](req(``), 1<<20); errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("empty body should be KindValidation, got %v", err)
	}
}

func TestBindAndValidate(t *testing.T) {
	v := validation.New()
	_, err := httpx.BindAndValidate[createReq](req(`{"name":""}`), v, 1<<20)
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("empty required field should fail validation: %v", err)
	}
	got, err := httpx.BindAndValidate[createReq](req(`{"name":"ok"}`), v, 1<<20)
	if err != nil || got.Name != "ok" {
		t.Fatalf("valid bind failed: %v", err)
	}
}

// ---------- etag / if-match ----------

func TestETagRoundTrip(t *testing.T) {
	tag := httpx.ETagFrom(7)
	r := httptest.NewRequest(http.MethodPut, "/", nil)
	r.Header.Set("If-Match", tag)
	v, err := httpx.RequireIfMatch(r)
	if err != nil || v != 7 {
		t.Fatalf("If-Match round trip: v=%d err=%v", v, err)
	}
}

func TestRequireIfMatchMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/", nil)
	if _, err := httpx.RequireIfMatch(r); errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("missing If-Match should be KindValidation: %v", err)
	}
}

// ---------- middleware ----------

func TestRecoverMiddlewareReturns500WithoutLeaking(t *testing.T) {
	h := httpx.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret in panic message") }),
		httpx.RequestID(),
		httpx.Recover(nil),
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 500 {
		t.Fatalf("panic should yield 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret in panic message") {
		t.Errorf("panic message leaked to the wire: %s", rec.Body.String())
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Error("RequestID middleware should set X-Request-Id")
	}
}

func TestRecoverDoesNotCorruptWrittenResponse(t *testing.T) {
	// A handler that writes bytes and THEN panics must not have a problem body
	// appended to its already-committed response (SEC-18).
	h := httpx.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":"partial"}`))
			panic("boom after write")
		}),
		httpx.Recover(nil),
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (already written before panic)", rec.Code)
	}
	if rec.Body.String() != `{"data":"partial"}` {
		t.Errorf("recover appended to a committed body: %q", rec.Body.String())
	}
}

func TestRecoverPropagatesAbortHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != http.ErrAbortHandler {
			t.Fatalf("ErrAbortHandler must propagate, got %v", r)
		}
	}()
	h := httpx.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) }),
		httpx.Recover(nil),
	)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestRequestIDHonorsInbound(t *testing.T) {
	var seen string
	h := httpx.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = httpx.RequestIDFrom(r.Context()) }),
		httpx.RequestID(),
	)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-Id", "abc-123")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if seen != "abc-123" {
		t.Errorf("inbound request id not honored: %q", seen)
	}
}

// ---------- success envelope ----------

func TestWriteJSONEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteJSON(rec, http.StatusCreated, httpx.OK(map[string]string{"id": "1"}))
	if rec.Code != 201 {
		t.Fatalf("status = %d", rec.Code)
	}
	var env struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("envelope: %v", err)
	}
	if env.Data["id"] != "1" {
		t.Errorf("data = %v", env.Data)
	}
}
