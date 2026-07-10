package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

// TestBackpressurePassesThroughUnderCap proves requests within the cap reach
// the handler normally and the in-flight gauge callback reports occupancy.
func TestBackpressurePassesThroughUnderCap(t *testing.T) {
	served := 0
	h := httpx.Backpressure(2, config.Overload{Status: http.StatusServiceUnavailable, RetryAfter: 2 * time.Second})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			served++
			w.WriteHeader(http.StatusOK)
		}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if served != 1 {
		t.Fatalf("handler must be reached under cap, served=%d", served)
	}
}

// TestBackpressureRejectsOverCap uses a blocking handler to hold the single
// slot open, proving a second concurrent request is rejected with the
// configured overload status and Retry-After BEFORE reaching the handler.
func TestBackpressureRejectsOverCap(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	var handlerHits int
	var mu sync.Mutex

	h := httpx.Backpressure(1, config.Overload{Status: http.StatusServiceUnavailable, RetryAfter: 3 * time.Second})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			mu.Lock()
			handlerHits++
			mu.Unlock()
			close(entered)
			<-release
			w.WriteHeader(http.StatusOK)
		}))

	// First request occupies the only slot and blocks.
	done := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done <- rec.Code
	}()
	<-entered // wait until the first request is inside the handler, holding the slot

	// Second request must be rejected without reaching the handler.
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("over-cap request status = %d, want 503", rec2.Code)
	}
	if ra := rec2.Header().Get("Retry-After"); ra != "3" {
		t.Errorf("Retry-After = %q, want %q", ra, "3")
	}
	if ct := rec2.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}

	close(release)
	if code := <-done; code != http.StatusOK {
		t.Fatalf("first (in-flight) request status = %d, want 200", code)
	}

	mu.Lock()
	defer mu.Unlock()
	if handlerHits != 1 {
		t.Fatalf("handler must be reached exactly once (not by the rejected request), got %d", handlerHits)
	}
}

// TestBackpressureDisabledWhenCapZero proves cap<=0 disables the limiter
// entirely (pass-through, no semaphore) — the safe default for existing
// deployments that haven't opted in (backlog B6 rollout guard).
func TestBackpressureDisabledWhenCapZero(t *testing.T) {
	served := 0
	h := httpx.Backpressure(0, config.Overload{Status: http.StatusServiceUnavailable, RetryAfter: time.Second})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			served++
			w.WriteHeader(http.StatusOK)
		}))
	for i := 0; i < 50; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200 (limiter disabled)", i, rec.Code)
		}
	}
	if served != 50 {
		t.Fatalf("served = %d, want 50", served)
	}
}

// TestBackpressure429Status proves the configured overload status (429
// instead of the 503 default) is honored.
func TestBackpressure429Status(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	h := httpx.Backpressure(1, config.Overload{Status: http.StatusTooManyRequests, RetryAfter: 500 * time.Millisecond})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(entered)
			<-release
			w.WriteHeader(http.StatusOK)
		}))

	go func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	<-entered

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec2.Code)
	}
	// ceil(500ms) = 1s
	if ra := rec2.Header().Get("Retry-After"); ra != "1" {
		t.Errorf("Retry-After = %q, want %q", ra, "1")
	}
	close(release)
}

// TestBackpressureMetricsHooks proves OnOverload fires exactly once per
// rejection and OnInFlightChange reports gauge deltas as requests enter/leave
// (the metrics wiring point the composition root uses for the rejected-overload
// counter and in-flight gauge).
func TestBackpressureMetricsHooks(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	var overloadCount int
	var mu sync.Mutex
	var gaugeValues []int

	h := httpx.Backpressure(1, config.Overload{Status: http.StatusServiceUnavailable, RetryAfter: time.Second},
		httpx.OnBackpressureOverload(func(route string) {
			mu.Lock()
			overloadCount++
			mu.Unlock()
		}),
		httpx.OnInFlightChange(func(n int) {
			mu.Lock()
			gaugeValues = append(gaugeValues, n)
			mu.Unlock()
		}),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	go func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	<-entered

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec2.Code)
	}

	close(release)
	// allow the background goroutine's deferred decrement to run
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if overloadCount != 1 {
		t.Fatalf("OnOverload must fire exactly once, got %d", overloadCount)
	}
	if len(gaugeValues) == 0 {
		t.Fatal("OnInFlightChange must fire at least once")
	}
	sawOne := false
	for _, v := range gaugeValues {
		if v == 1 {
			sawOne = true
		}
	}
	if !sawOne {
		t.Errorf("expected an in-flight gauge value of 1 while the blocking handler held its slot, got %v", gaugeValues)
	}
}

// TestBackpressureRetryAfterRoundsUp proves sub-second retry-after values are
// rounded up to at least 1 second (Retry-After is defined in whole seconds).
func TestBackpressureRetryAfterRoundsUp(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	h := httpx.Backpressure(1, config.Overload{Status: http.StatusServiceUnavailable, RetryAfter: 100 * time.Millisecond})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(entered)
			<-release
			w.WriteHeader(http.StatusOK)
		}))
	go func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	<-entered

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if ra := rec2.Header().Get("Retry-After"); ra != "1" {
		t.Errorf("Retry-After = %q, want %q (rounded up from 100ms)", ra, "1")
	}
	close(release)
}

// TestBackpressureZeroStatusDefaultsTo503 proves passing a zero-value
// config.Overload (Status unset) falls back to 503, so a caller that
// forgets to set Status still gets a sane overload response instead of
// writing a nonsensical 0 status.
func TestBackpressureZeroStatusDefaultsTo503(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	h := httpx.Backpressure(1, config.Overload{RetryAfter: time.Second})( // Status left zero
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(entered)
			<-release
			w.WriteHeader(http.StatusOK)
		}))
	go func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	<-entered

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (zero-value Overload.Status default)", rec2.Code)
	}
	close(release)
}

// TestBackpressureAllowsSerializedRequests proves the semaphore correctly
// releases slots: N sequential requests under a cap of 1 all succeed.
func TestBackpressureAllowsSerializedRequests(t *testing.T) {
	h := httpx.Backpressure(1, config.Overload{Status: http.StatusServiceUnavailable, RetryAfter: time.Second})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("sequential request %d: status = %d, want 200", i, rec.Code)
		}
	}
}
