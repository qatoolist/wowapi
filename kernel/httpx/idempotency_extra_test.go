package httpx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

// dto is a small marshalable response body used by the idempotency tests.
type dto struct {
	Value string `json:"value"`
}

func idemReq(key string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/things?x=1", nil)
	if key != "" {
		r.Header.Set(httpx.IdempotencyHeader, key)
	}
	return r
}

// TestRequestHashStableAndSensitive proves the hash is deterministic for the same
// (method, path, query, canonical) and changes when any of them change (SEC-19).
func TestRequestHashStableAndSensitive(t *testing.T) {
	r1 := httptest.NewRequest(http.MethodPost, "/things?x=1", nil)
	r2 := httptest.NewRequest(http.MethodPost, "/things?x=1", nil)
	r3 := httptest.NewRequest(http.MethodPost, "/things?x=2", nil) // different query

	base := httpx.RequestHash(r1, []byte(`{"a":1}`))
	if base != httpx.RequestHash(r2, []byte(`{"a":1}`)) {
		t.Fatal("same request + canonical must hash identically")
	}
	if base == httpx.RequestHash(r1, []byte(`{"a":2}`)) {
		t.Fatal("a different canonical body must change the hash")
	}
	if base == httpx.RequestHash(r3, []byte(`{"a":1}`)) {
		t.Fatal("a different query string must change the hash (SEC-19)")
	}
}

// TestWithIdempotencyNoKeyRunsOnce covers the header-less path: op runs in one
// tenant transaction and its response is written directly, nothing stored.
func TestWithIdempotencyNoKeyRunsOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	store := &fakeIdem{}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}

	httpx.WithIdempotency(context.Background(), rec, idemReq(""), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			return http.StatusCreated, dto{Value: "made"}, nil
		})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	var body dto
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Value != "made" {
		t.Fatalf("body = %q err=%v, want {made}", rec.Body.String(), err)
	}
	if store.completed || store.discarded {
		t.Fatal("a keyless request must not touch the idempotency store")
	}
}

// TestWithIdempotencyMissingStore covers the guard: a key is present but no store
// is configured → a 500 problem, never a silent unguarded execution.
func TestWithIdempotencyMissingStore(t *testing.T) {
	rec := httptest.NewRecorder()
	cfg := httpx.IdempotencyConfig{Store: nil}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			t.Fatal("op must not run without a store")
			return 0, nil, nil
		})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

// TestWithIdempotencyFreshCompletes covers the fresh-claim happy path: a 2xx
// result is stored (Complete) and written to the client.
func TestWithIdempotencyFreshCompletes(t *testing.T) {
	rec := httptest.NewRecorder()
	store := &fakeIdem{begin: database.Replay{Fresh: true}}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			return http.StatusOK, dto{Value: "fresh"}, nil
		})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !store.completed {
		t.Fatal("a successful (2xx) idempotent op must be recorded via Complete")
	}
	var body dto
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Value != "fresh" {
		t.Fatalf("body = %q err=%v", rec.Body.String(), err)
	}
}

// TestWithIdempotencyReplaysStored covers the replay path: Begin returns a stored
// response, which is written verbatim without re-running op.
func TestWithIdempotencyReplaysStored(t *testing.T) {
	rec := httptest.NewRecorder()
	stored := []byte(`{"value":"stored"}`)
	store := &fakeIdem{begin: database.Replay{Found: true, ResponseStatus: http.StatusAccepted, ResponseBody: stored}}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			t.Fatal("op must not run on a replay")
			return 0, nil, nil
		})

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (replayed)", rec.Code)
	}
	if rec.Body.String() != string(stored) {
		t.Fatalf("body = %q, want the stored response verbatim", rec.Body.String())
	}
}

// TestWithIdempotencyTxErrorBecomesProblem covers the transaction-failure branch:
// when the tenant transaction fails, the error is surfaced as a problem response.
func TestWithIdempotencyTxErrorBecomesProblem(t *testing.T) {
	rec := httptest.NewRecorder()
	store := &fakeIdem{begin: database.Replay{Fresh: true}}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}
	txm := fakeTxM{err: kerr.E(kerr.KindInternal, "tx_failed", "transaction aborted")}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), txm, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			return http.StatusOK, dto{Value: "x"}, nil
		})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (tx failure)", rec.Code)
	}
}

// TestWithIdempotencyBeginErrorBecomesProblem covers the Begin-error branch: a
// store failure during the claim propagates out of the transaction as a problem.
func TestWithIdempotencyBeginErrorBecomesProblem(t *testing.T) {
	rec := httptest.NewRecorder()
	store := &fakeIdem{beginErr: kerr.E(kerr.KindConflict, "conflict", "key reused with a different request")}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			t.Fatal("op must not run when Begin fails")
			return 0, nil, nil
		})

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (Begin conflict)", rec.Code)
	}
}

// TestWithIdempotencyUnmarshalableBody covers the response-encode failure branch:
// a body that cannot be JSON-encoded becomes a 500 rather than a corrupt store.
func TestWithIdempotencyUnmarshalableBody(t *testing.T) {
	rec := httptest.NewRecorder()
	store := &fakeIdem{begin: database.Replay{Fresh: true}}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			return http.StatusOK, make(chan int), nil // unmarshalable
		})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (encode failure)", rec.Code)
	}
	if store.completed {
		t.Fatal("an unencodable response must not be stored")
	}
}

// TestWithIdempotencyNon2xxDiscards covers the non-2xx branch: a non-success
// result commits its writes but the key claim is discarded so it stays retryable
// (ARCH-32/SEC-23), never caching a failure.
func TestWithIdempotencyNon2xxDiscards(t *testing.T) {
	rec := httptest.NewRecorder()
	store := &fakeIdem{begin: database.Replay{Fresh: true}}
	cfg := httpx.IdempotencyConfig{Store: store, ActorScope: "cap-1"}

	httpx.WithIdempotency(context.Background(), rec, idemReq("k-1"), fakeTxM{}, cfg, "hash",
		func(context.Context, database.TenantDB) (int, any, error) {
			return http.StatusBadRequest, dto{Value: "nope"}, nil
		})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !store.discarded {
		t.Fatal("a non-2xx result must Discard the key claim (stay retryable)")
	}
	if store.completed {
		t.Fatal("a non-2xx result must not Complete (cache) the key")
	}
}
