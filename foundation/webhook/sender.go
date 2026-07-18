package webhook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/qatoolist/wowapi/kernel/httpclient"
	"github.com/qatoolist/wowapi/kernel/safety"
)

// HTTPSender implements Sender using a standard net/http client. The client
// carries OutboundTimeout as its hard ceiling; the caller's context deadline
// may further constrain it.
//
// Outbound webhook URLs are USER-CONFIGURABLE (tenants register their own
// endpoints), so by default the underlying client is kernel/httpclient's
// SSRF-safe client (backlog B2): loopback, link-local (incl. the cloud
// metadata address), RFC1918/ULA private, and unspecified addresses are all
// refused at dial time. WithHTTPClientConfig and WithSSRFProtectionDisabled
// are the escape hatches for intentional internal targets — wire them from
// config.WebhookOutbound (kernel/config), never hard-code an override.
type HTTPSender struct {
	client *http.Client
}

// HTTPSenderOption customizes NewHTTPSender.
type HTTPSenderOption func(*httpSenderCfg)

type httpSenderCfg struct {
	clientCfg         httpclient.Config
	ssrfProtectionOff bool
}

// WithHTTPClientConfig sets the SSRF-guard config (allowlist hosts/CIDRs,
// timeout) the default sender's client is built with. Corresponds to
// config.WebhookOutbound.AllowedHosts/AllowedCIDRs.
func WithHTTPClientConfig(cfg httpclient.Config) HTTPSenderOption {
	return func(c *httpSenderCfg) { c.clientCfg = cfg }
}

// WithSSRFProtectionDisabled removes the SSRF guard entirely, falling back to
// a bare net/http client. Corresponds to
// config.WebhookOutbound.SSRFProtectionDisabled — Validate() refuses that
// config key in prod, so this option should only ever be reached in
// local/dev wiring.
func WithSSRFProtectionDisabled() HTTPSenderOption {
	return func(c *httpSenderCfg) { c.ssrfProtectionOff = true }
}

// NewHTTPSender returns the production Sender backed by net/http. By default
// it is SSRF-safe (kernel/httpclient, dial-time address-class blocking); pass
// WithHTTPClientConfig for an allowlist or WithSSRFProtectionDisabled to opt
// out entirely (local/dev only).
func NewHTTPSender(opts ...HTTPSenderOption) *HTTPSender {
	cfg := httpSenderCfg{}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.ssrfProtectionOff {
		return &HTTPSender{client: &http.Client{Timeout: OutboundTimeout}}
	}
	clientCfg := cfg.clientCfg
	if clientCfg.Timeout <= 0 {
		clientCfg.Timeout = OutboundTimeout
	}
	return &HTTPSender{client: httpclient.New(clientCfg)}
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

// DuplicateSafety declares that HTTPSender has no built-in duplicate-safety
// mechanism; the caller/framework must suppress duplicates via lease fencing
// and idempotency keys.
func (s *HTTPSender) DuplicateSafety() safety.Mechanism { return safety.None }
