package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// failingReader always errors, letting the tests below deterministically
// exercise newCSRFToken/ensureCSRFCookie's entropy-failure branch without
// depending on crypto/rand actually failing (which does not happen in
// practice on any supported platform).
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("entropy source unavailable") }

func TestNewCSRFTokenPropagatesReadError(t *testing.T) {
	old := csrfRandReader
	csrfRandReader = failingReader{}
	defer func() { csrfRandReader = old }()

	if _, err := newCSRFToken(); err == nil {
		t.Fatal("newCSRFToken must propagate the entropy source's error")
	}
}

// TestEnsureCSRFCookieSkipsOnTokenGenFailure proves that when token
// generation fails, ensureCSRFCookie simply does not set a cookie rather
// than panicking or writing a broken one — the safe methods must still be
// servable even if entropy generation transiently fails.
func TestEnsureCSRFCookieSkipsOnTokenGenFailure(t *testing.T) {
	old := csrfRandReader
	csrfRandReader = failingReader{}
	defer func() { csrfRandReader = old }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ensureCSRFCookie(rec, req, CSRFPolicy{CookieName: "csrf_token"})

	if got := rec.Result().Cookies(); len(got) != 0 {
		t.Errorf("no cookie should be set when token generation fails, got %v", got)
	}
}
