package webhook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// HTTPSender implements Sender using a standard net/http client. The client
// carries OutboundTimeout as its hard ceiling; the caller's context deadline
// may further constrain it.
type HTTPSender struct {
	client *http.Client
}

// NewHTTPSender returns the production Sender backed by net/http.
func NewHTTPSender() *HTTPSender {
	return &HTTPSender{client: &http.Client{Timeout: OutboundTimeout}}
}

// Post sends a POST request with the given body and headers and returns the
// HTTP status code on success.
func (s *HTTPSender) Post(ctx context.Context, url string, body []byte, headers map[string]string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("webhook sender: build request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("webhook sender: POST: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

// FakeSender is a test double that records every Post call and returns a
// pre-configured status code and optional error.
type FakeSender struct {
	// StatusCode is returned from every Post call (default 200 when zero).
	StatusCode int
	// Err is returned from every Post call when non-nil.
	Err error

	// Calls accumulates the arguments of every Post invocation.
	Calls []SentCall
}

// SentCall records one Post invocation.
type SentCall struct {
	URL     string
	Body    []byte
	Headers map[string]string
}

// Post records the call and returns the pre-configured response.
func (f *FakeSender) Post(_ context.Context, url string, body []byte, headers map[string]string) (int, error) {
	f.Calls = append(f.Calls, SentCall{URL: url, Body: body, Headers: headers})
	code := f.StatusCode
	if code == 0 {
		code = http.StatusOK
	}
	return code, f.Err
}

// FakeSecretResolver is a test double that returns a fixed secret for any ref.
type FakeSecretResolver struct {
	Secret string
}

// Resolve returns the configured Secret regardless of ref.
func (r *FakeSecretResolver) Resolve(_ context.Context, _ string) (string, error) {
	return r.Secret, nil
}
