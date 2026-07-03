package httpx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// IdempotencyHeader is the client header carrying the idempotency key.
const IdempotencyHeader = "Idempotency-Key"

// RequestHash produces the stable hash stored with an idempotency key so a
// retry with the SAME key but a DIFFERENT request is rejected (409). Callers
// pass the canonical bytes of the decoded+validated command (the raw body is
// already consumed by decoding), which the kernel binds to method, path, AND
// query string (SEC-19: two requests differing only in query params must not
// share a stored response). Callers MUST pass deterministic, non-empty
// canonical bytes; an empty canonical weakens the different-request check.
func RequestHash(r *http.Request, canonical []byte) string {
	h := sha256.New()
	h.Write([]byte(r.Method))
	h.Write([]byte{0})
	h.Write([]byte(r.URL.Path))
	h.Write([]byte{0})
	h.Write([]byte(r.URL.RawQuery))
	h.Write([]byte{0})
	h.Write(canonical)
	return hex.EncodeToString(h.Sum(nil))
}

// Operation is a mutating handler body: it runs inside a tenant transaction and
// returns the HTTP status and response DTO to write (and store, when
// idempotent). Returning an error rolls the transaction back and nothing is
// stored.
type Operation func(ctx context.Context, db database.TenantDB) (status int, body any, err error)

// IdempotencyConfig configures WithIdempotency.
type IdempotencyConfig struct {
	Store      database.IdemStore
	ActorScope string        // capacity id / system actor; scopes the key
	TTL        time.Duration // how long a stored response is replayable
}

// WithIdempotency executes op at most once per (tenant, actor, Idempotency-Key)
// and writes the response. Without the header it simply runs op in one tenant
// transaction. With the header it replays a stored response on retry, rejects a
// reused key carrying a different request (409 conflict), and rejects a still
// in-flight duplicate (409 retry_later) — the key claim, the business writes,
// and the stored response all commit in the same transaction (blueprint 05
// §1–2). It owns the response so replay is transparent to the handler.
func WithIdempotency(ctx context.Context, w http.ResponseWriter, r *http.Request, tx database.TxManager, cfg IdempotencyConfig, requestHash string, op Operation) {
	key := r.Header.Get(IdempotencyHeader)

	// No key → plain tenant transaction, no storage.
	if key == "" {
		var (
			status int
			body   any
		)
		err := tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			status, body, e = op(ctx, db)
			return e
		})
		if err != nil {
			WriteError(ctx, w, err)
			return
		}
		WriteJSON(w, status, body)
		return
	}

	if cfg.Store == nil {
		WriteError(ctx, w, kerr.E(kerr.KindInternal, "internal", "idempotency store not configured"))
		return
	}

	var (
		replayStatus int
		replayBody   []byte
		freshStatus  int
		freshBody    []byte
		replayed     bool
	)
	err := tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rep, err := cfg.Store.Begin(ctx, db, cfg.ActorScope, key, requestHash, cfg.TTL)
		if err != nil {
			return err
		}
		if rep.Found {
			replayed = true
			replayStatus, replayBody = rep.ResponseStatus, rep.ResponseBody
			return nil
		}
		status, body, err := op(ctx, db)
		if err != nil {
			return err
		}
		buf, err := json.Marshal(body)
		if err != nil {
			return kerr.E(kerr.KindInternal, "internal", "encode idempotent response")
		}
		// Only successful (2xx) mutations become idempotent. A non-2xx result
		// returned without an error still commits its writes, but the claim is
		// discarded so the same key stays retryable rather than caching a
		// failure forever (review findings ARCH-32/SEC-23).
		if status >= 200 && status < 300 {
			if err := cfg.Store.Complete(ctx, db, cfg.ActorScope, key, status, buf); err != nil {
				return err
			}
		} else if err := cfg.Store.Discard(ctx, db, cfg.ActorScope, key); err != nil {
			return err
		}
		freshStatus, freshBody = status, buf
		return nil
	})
	if err != nil {
		WriteError(ctx, w, err)
		return
	}
	if replayed {
		writeRaw(w, replayStatus, replayBody)
		return
	}
	writeRaw(w, freshStatus, freshBody)
}

func writeRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
