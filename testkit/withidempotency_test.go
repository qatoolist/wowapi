package testkit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/httpx"
)

// TestIntegrationWithIdempotency drives the full httpx.WithIdempotency
// composition against real Postgres: the operation runs exactly once across two
// requests carrying the same Idempotency-Key, and the second request replays
// the stored response without re-running the op.
func TestIntegrationWithIdempotency(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	tenant := uuid.New()

	var runs int
	op := func(ctx context.Context, db database.TenantDB) (int, any, error) {
		runs++
		return http.StatusCreated, map[string]string{"id": "created"}, nil
	}

	doRequest := func() *httptest.ResponseRecorder {
		ctx := database.WithTenantID(context.Background(), tenant)
		r := httptest.NewRequest(http.MethodPost, "/things", nil)
		r.Header.Set(httpx.IdempotencyHeader, "key-xyz")
		rec := httptest.NewRecorder()
		cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "actor-1", TTL: time.Hour}
		httpx.WithIdempotency(ctx, rec, r, h.TxM, cfg, httpx.RequestHash(r, []byte(`{}`)), op)
		return rec
	}

	first := doRequest()
	if first.Code != http.StatusCreated || !strings.Contains(first.Body.String(), "created") {
		t.Fatalf("first request: code=%d body=%s", first.Code, first.Body.String())
	}
	second := doRequest()
	if second.Code != http.StatusCreated || !strings.Contains(second.Body.String(), "created") {
		t.Fatalf("second request: code=%d body=%s", second.Code, second.Body.String())
	}
	if runs != 1 {
		t.Fatalf("operation ran %d times; idempotency must run it exactly once", runs)
	}
}

// TestIntegrationWithIdempotencyConflict proves a reused key with a different
// request body is rejected as a 409 conflict problem.
func TestIntegrationWithIdempotencyConflict(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	tenant := uuid.New()
	op := func(ctx context.Context, db database.TenantDB) (int, any, error) {
		return http.StatusOK, map[string]string{"ok": "1"}, nil
	}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "actor-1", TTL: time.Hour}

	send := func(hash string) *httptest.ResponseRecorder {
		ctx := database.WithTenantID(context.Background(), tenant)
		r := httptest.NewRequest(http.MethodPost, "/things", nil)
		r.Header.Set(httpx.IdempotencyHeader, "reused")
		rec := httptest.NewRecorder()
		httpx.WithIdempotency(ctx, rec, r, h.TxM, cfg, hash, op)
		return rec
	}

	if rec := send("hash-A"); rec.Code != http.StatusOK {
		t.Fatalf("first: %d", rec.Code)
	}
	rec := send("hash-B") // same key, different request hash
	if rec.Code != http.StatusConflict {
		t.Fatalf("reused key with different request should be 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
