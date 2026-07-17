package httpx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/filtering"
	"github.com/qatoolist/wowapi/v2/kernel/httpx"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/pagination"
	"github.com/qatoolist/wowapi/v2/kernel/validation"
)

// --- listing helpers ---

func TestParsePagination(t *testing.T) {
	def := pagination.Defaults{PerPage: 20, MaxPerPage: 100}

	req := httptest.NewRequest(http.MethodGet, "/?per_page=50", nil)
	got, err := httpx.ParsePagination(req, def)
	if err != nil {
		t.Fatalf("valid per_page: %v", err)
	}
	if got.Limit != 50 {
		t.Fatalf("limit = %d, want 50", got.Limit)
	}

	bad := httptest.NewRequest(http.MethodGet, "/?per_page=abc", nil)
	if _, err := httpx.ParsePagination(bad, def); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("non-integer per_page must be KindValidation, got %v", err)
	}
}

func TestParseFilters(t *testing.T) {
	allow := filtering.Allowlist{"status": {Col: "status", Ops: []filtering.Op{filtering.OpEq}}}

	req := httptest.NewRequest(http.MethodGet, "/?filter.status=eq:active", nil)
	if _, err := httpx.ParseFilters(req, allow); err != nil {
		t.Fatalf("allowlisted filter must parse: %v", err)
	}

	bad := httptest.NewRequest(http.MethodGet, "/?filter.bogus=eq:x", nil)
	if _, err := httpx.ParseFilters(bad, allow); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("unknown filter field must be KindValidation, got %v", err)
	}
}

func TestParseSort(t *testing.T) {
	allow := filtering.SortAllowlist{"created_at": {Col: "created_at"}}

	req := httptest.NewRequest(http.MethodGet, "/?sort=created_at:desc", nil)
	if _, err := httpx.ParseSort(req, allow); err != nil {
		t.Fatalf("allowlisted sort must parse: %v", err)
	}

	bad := httptest.NewRequest(http.MethodGet, "/?sort=bogus", nil)
	if _, err := httpx.ParseSort(bad, allow); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("unknown sort key must be KindValidation, got %v", err)
	}
}

// --- etag helpers ---

func TestParseResourceID(t *testing.T) {
	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/things/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	got, err := httpx.ParseResourceID(req, "id")
	if err != nil || got != id {
		t.Fatalf("valid id: got %v err %v", got, err)
	}

	missing := httptest.NewRequest(http.MethodGet, "/things", nil)
	if _, err := httpx.ParseResourceID(missing, "id"); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("missing path param must be KindValidation, got %v", err)
	}

	malformed := httptest.NewRequest(http.MethodGet, "/things/not-a-uuid", nil)
	malformed.SetPathValue("id", "not-a-uuid")
	if _, err := httpx.ParseResourceID(malformed, "id"); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("malformed id must be KindValidation, got %v", err)
	}
}

func TestRequireIfMatchVariants(t *testing.T) {
	mk := func(v string) *http.Request {
		r := httptest.NewRequest(http.MethodPut, "/", nil)
		if v != "" {
			r.Header.Set("If-Match", v)
		}
		return r
	}

	if v, err := httpx.RequireIfMatch(mk(`"v5"`)); err != nil || v != 5 {
		t.Fatalf("strong tag: v=%d err=%v, want 5", v, err)
	}
	if v, err := httpx.RequireIfMatch(mk(`W/"v3"`)); err != nil || v != 3 {
		t.Fatalf("weak validator must be tolerated: v=%d err=%v, want 3", v, err)
	}
	if _, err := httpx.RequireIfMatch(mk("*")); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("wildcard If-Match must be rejected, got %v", err)
	}
	if _, err := httpx.RequireIfMatch(mk("not-a-number")); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("malformed If-Match must be KindValidation, got %v", err)
	}
}

func TestAuditMetaFrom(t *testing.T) {
	created := time.Now().Add(-time.Hour)
	updated := time.Now()
	by := uuid.New()
	a := model.Auditable{CreatedAt: created, CreatedBy: by, UpdatedAt: &updated}

	meta := httpx.AuditMetaFrom(a, 7)
	if meta.Version != 7 {
		t.Fatalf("version = %d, want 7", meta.Version)
	}
	if meta.CreatedBy != by || !meta.CreatedAt.Equal(created) {
		t.Fatalf("author metadata not carried through: %+v", meta)
	}
	if meta.UpdatedAt == nil || !meta.UpdatedAt.Equal(updated) {
		t.Fatalf("updated_at not carried through: %+v", meta.UpdatedAt)
	}
}

// --- response helpers ---

func TestOKWithMeta(t *testing.T) {
	resp := httpx.OKWithMeta("payload", &httpx.Meta{RequestID: "req-1"})
	if resp.Data != "payload" {
		t.Fatalf("data = %q, want payload", resp.Data)
	}
	if resp.Meta == nil || resp.Meta.RequestID != "req-1" {
		t.Fatalf("meta not attached: %+v", resp.Meta)
	}
}

// TestWriteJSONMarshalFailureDegradesTo500 covers the marshal-error branch of
// WriteJSON: an unmarshalable value (a channel) degrades to a 500 problem body
// rather than a partial write.
func TestWriteJSONMarshalFailureDegradesTo500(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteJSON[any](rec, http.StatusOK, make(chan int))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("marshal failure = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}

// --- error writer ---

// TestWriteErrorUnmappedKindFallsBackTitle covers a Kind absent from the title
// table (KindIdempotencyExpired): the status/code are the kind's, but the title
// falls back to the internal title rather than being blank.
func TestWriteErrorUnmappedKindFallsBackTitle(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(context.Background(), rec, kerr.E(kerr.KindIdempotencyExpired, "", "key expired"))

	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d, want 410", rec.Code)
	}
	var p httpx.ProblemError
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body not problem json: %v", err)
	}
	if p.Code != "idempotency_key_expired" {
		t.Fatalf("code = %q, want idempotency_key_expired", p.Code)
	}
	if p.Title != "Internal error" {
		t.Fatalf("unmapped-kind title must fall back to %q, got %q", "Internal error", p.Title)
	}
}

// TestWriteErrorHonorsExplicitCode covers the non-empty Code branch of WriteError.
func TestWriteErrorHonorsExplicitCode(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(context.Background(), rec, kerr.E(kerr.KindConflict, "widget_locked", "the widget is locked"))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	var p httpx.ProblemError
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body not problem json: %v", err)
	}
	if p.Code != "widget_locked" {
		t.Fatalf("code = %q, want widget_locked (explicit override)", p.Code)
	}
	if p.Detail != "the widget is locked" {
		t.Fatalf("detail = %q, want the user-safe message", p.Detail)
	}
}

// --- router ---

// TestRouterRejectsNilHandler covers the nil-handler guard in Handle.
func TestRouterRejectsNilHandler(t *testing.T) {
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/x", httpx.RouteMeta{Public: true}, nil)
	if r.Err() == nil {
		t.Fatal("a nil handler must record a registration error")
	}
}

// --- decode ---

type decodePayload struct {
	Name string `json:"name" validate:"required"`
}

func TestDecodeJSONNilBody(t *testing.T) {
	r := &http.Request{Method: http.MethodPost} // Body == nil
	if _, err := httpx.DecodeJSON[decodePayload](r, 1024); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("nil body must be KindValidation, got %v", err)
	}
}

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"a"} {"name":"b"}`))
	if _, err := httpx.DecodeJSON[decodePayload](r, 1024); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("trailing data must be KindValidation, got %v", err)
	}
}

func TestDecodeJSONBodyTooLarge(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"aaaaaaaaaaaaaaaaaaaa"}`))
	if _, err := httpx.DecodeJSON[decodePayload](r, 4); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("oversized body must be KindValidation, got %v", err)
	}
}

func TestBindAndValidateDecodeError(t *testing.T) {
	v := validation.New()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`)) // malformed json
	if _, err := httpx.BindAndValidate[decodePayload](r, v, 1024); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("malformed json must be KindValidation, got %v", err)
	}
}

func TestBindAndValidateValidationError(t *testing.T) {
	v := validation.New()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`)) // fails required
	_, err := httpx.BindAndValidate[decodePayload](r, v, 1024)
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("validation failure must be KindValidation, got %v", err)
	}
	e, ok := kerr.As(err)
	if !ok || len(e.Fields) == 0 {
		t.Fatalf("validation error must carry field errors, got %+v", err)
	}
}
