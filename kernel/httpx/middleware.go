package httpx

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// Middleware wraps an http.Handler. The kernel provides the cross-cutting
// concerns (request id, panic recovery, error-safe logging); auth/tenant
// middleware is added in Phase 4.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares so that the first listed runs outermost.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestID assigns a correlation id (honoring an inbound X-Request-Id when
// present) and stores it in the context and the response header, so every log
// line and problem body can be correlated.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = uuid.NewString()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := WithRequestID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Recover converts a panic into a 500 problem response: the stack goes to the
// logger and a panic metric, never to the wire (blueprint 04 §5). It must be
// the outermost middleware around handlers.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tw := &trackedWriter{ResponseWriter: w}
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				// http.ErrAbortHandler is the stdlib convention for aborting a
				// response silently — must propagate, not become a 500.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				if logger != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"request_id", RequestIDFrom(r.Context()),
						"method", r.Method,
						"path", r.URL.Path,
						"panic", rec,
					)
				}
				// Only emit a clean problem body if nothing was written yet;
				// appending to a partially-sent response would corrupt it
				// (review finding SEC-18).
				if !tw.wrote {
					writeInternal(r.Context(), tw)
				}
			}()
			next.ServeHTTP(tw, r)
		})
	}
}

// trackedWriter records whether the handler has begun the response, so Recover
// can tell a clean pre-write panic from one after bytes are on the wire.
type trackedWriter struct {
	http.ResponseWriter
	wrote bool
}

func (t *trackedWriter) WriteHeader(status int) {
	t.wrote = true
	t.ResponseWriter.WriteHeader(status)
}

func (t *trackedWriter) Write(b []byte) (int, error) {
	t.wrote = true
	return t.ResponseWriter.Write(b)
}
