package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/qatoolist/wowapi/kernel/errors"
)

// jwks.go implements the production KeySource: a JWKS-over-HTTPS adapter that
// fetches an issuer's JSON Web Key Set, caches the parsed verification keys by
// kid with a TTL, and refetches to pick up key rotation. It complements the
// in-memory staticKeySource (keysource.go) the tests use.
//
// Failure posture (SEC / 04 §5): an unknown kid is an opaque
// KindUnauthenticated (fail closed, no oracle). A transport/parse failure of the
// JWKS or discovery endpoint is a KindExternal HARD fault, deliberately NOT a
// 401 — the verifier and any composite authenticator must not report a transient
// upstream outage as a clean "invalid token". No key material or URL body ever
// reaches the wire.

const (
	// defaultJWKSTTL caps how long fetched keys are served before a refetch.
	defaultJWKSTTL = 15 * time.Minute
	// defaultJWKSMinRefresh throttles rotation-triggered refetches: within a
	// fresh cache window, an unknown kid triggers at most one refetch per this
	// interval so unknown-kid tokens cannot force unbounded JWKS fetches (DoS).
	defaultJWKSMinRefresh = time.Minute
	// defaultJWKSTimeout bounds a single discovery/JWKS HTTP request.
	defaultJWKSTimeout = 10 * time.Second
	// maxJWKSBytes caps discovery/JWKS response bodies read into memory.
	maxJWKSBytes = 1 << 20 // 1 MiB
)

// JWKSConfig parameterizes a JWKS-over-HTTPS KeySource.
type JWKSConfig struct {
	// Issuer is the token issuer (the iss claim). When JWKSURI is empty it is
	// used for OIDC discovery at <issuer>/.well-known/openid-configuration.
	Issuer string
	// JWKSURI is the explicit JWKS endpoint. When empty it is discovered from
	// Issuer. It must be https (loopback http is permitted for tests/local IdPs).
	JWKSURI string
	// TTL bounds how long fetched keys are cached before a refetch (default 15m).
	TTL time.Duration
	// Client is the HTTP client used to fetch discovery/JWKS documents. A nil
	// client defaults to one with a 10s timeout.
	Client *http.Client
	// Now is the clock, injectable for tests (default time.Now). It drives the
	// cache TTL and the rotation-refetch throttle.
	Now func() time.Time
}

// NewJWKSKeySource builds a caching JWKS-over-HTTPS KeySource. It requires an
// Issuer (for discovery) or an explicit JWKSURI; a non-loopback URL must be
// https. Returns an error only for static misconfiguration; network fetches
// happen lazily on the first Key call.
func NewJWKSKeySource(cfg JWKSConfig) (KeySource, error) {
	if cfg.Issuer == "" && cfg.JWKSURI == "" {
		return nil, errors.E(errors.KindInternal, "internal",
			"auth: JWKS source requires an issuer or jwks_uri", errors.Op("auth.NewJWKSKeySource"))
	}
	if cfg.JWKSURI != "" {
		if err := validateHTTPSURL(cfg.JWKSURI); err != nil {
			return nil, err
		}
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultJWKSTTL
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: defaultJWKSTimeout, Transport: jwksTransport()}
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &jwksKeySource{
		issuer:     strings.TrimRight(cfg.Issuer, "/"),
		jwksURI:    cfg.JWKSURI,
		ttl:        ttl,
		minRefresh: defaultJWKSMinRefresh,
		client:     client,
		now:        now,
	}, nil
}

// jwksTransport is the default outbound transport for JWKS/discovery fetches.
// It clones the stdlib default transport but DISABLES proxy use
// (http.DefaultTransport.Proxy is http.ProxyFromEnvironment): the issuer/JWKS
// URI is trusted boot-time config already constrained by validateHTTPSURL, and
// honoring an ambient HTTP(S)_PROXY would route auth-critical key fetches
// through an unvalidated hop — the same egress policy kernel/httpclient enforces
// for webhook delivery. Unlike that client, no private-IP dial guard is applied:
// an internal https OIDC issuer (e.g. a cluster-private IdP) is a legitimate,
// supported deployment that validateHTTPSURL permits. A deployment that wants a
// fully guarded client can inject one via JWKSConfig.Client.
func jwksTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = nil
	return t
}

// jwksKeySource is a concurrency-safe caching KeySource.
type jwksKeySource struct {
	issuer     string
	ttl        time.Duration
	minRefresh time.Duration
	client     *http.Client
	now        func() time.Time

	mu        sync.Mutex
	jwksURI   string         // explicit or discovered endpoint (cached once resolved)
	keys      map[string]any // kid -> *rsa.PublicKey / *ecdsa.PublicKey
	lastFetch time.Time
	expiry    time.Time
}

// Key returns the verification key for kid, fetching/refreshing the JWKS as
// needed. An unknown kid is KindUnauthenticated; a fetch/parse fault is a
// KindExternal hard error (see file header).
func (s *jwksKeySource) Key(ctx context.Context, kid string) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.now().Before(s.expiry) {
		if k, ok := s.keys[kid]; ok {
			return k, nil
		}
		// Fresh cache but unknown kid: possibly a just-rotated signing key.
		// Refetch at most once per minRefresh window; otherwise fail closed.
		if s.now().Sub(s.lastFetch) < s.minRefresh {
			return nil, unknownKID()
		}
	}

	keys, err := s.load(ctx)
	if err != nil {
		return nil, err // hard fault: propagate (must not be masked as a 401)
	}
	s.keys = keys
	s.lastFetch = s.now()
	s.expiry = s.lastFetch.Add(s.ttl)

	if k, ok := keys[kid]; ok {
		return k, nil
	}
	return nil, unknownKID()
}

// load resolves the JWKS endpoint (discovering it from the issuer on first use)
// and fetches + parses the key set.
func (s *jwksKeySource) load(ctx context.Context) (map[string]any, error) {
	uri := s.jwksURI
	if uri == "" {
		discovered, err := s.discover(ctx)
		if err != nil {
			return nil, err
		}
		if err := validateHTTPSURL(discovered); err != nil {
			return nil, err
		}
		s.jwksURI = discovered // cache the discovered endpoint
		uri = discovered
	}
	body, err := s.get(ctx, uri)
	if err != nil {
		return nil, err
	}
	return parseJWKS(body)
}

// discover reads the issuer's OIDC discovery document and returns its jwks_uri.
func (s *jwksKeySource) discover(ctx context.Context) (string, error) {
	body, err := s.get(ctx, s.issuer+"/.well-known/openid-configuration")
	if err != nil {
		return "", err
	}
	var doc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if uerr := json.Unmarshal(body, &doc); uerr != nil || doc.JWKSURI == "" {
		return "", errors.E(errors.KindExternal, "upstream_error",
			"auth: OIDC discovery document missing jwks_uri", errors.Op("auth.JWKS.discover"))
	}
	return doc.JWKSURI, nil
}

// get performs a size-capped GET and returns the body, or a KindExternal error.
func (s *jwksKeySource) get(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, errors.E(errors.KindExternal, "upstream_error",
			"auth: build JWKS request", err, errors.Op("auth.JWKS.get"))
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.E(errors.KindExternal, "upstream_error",
			"auth: fetch JWKS", err, errors.Op("auth.JWKS.get"))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.E(errors.KindExternal, "upstream_error",
			"auth: JWKS endpoint returned non-200", errors.Op("auth.JWKS.get"))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxJWKSBytes))
	if err != nil {
		return nil, errors.E(errors.KindExternal, "upstream_error",
			"auth: read JWKS body", err, errors.Op("auth.JWKS.get"))
	}
	return body, nil
}

// jwk is one JSON Web Key (the subset needed for RS256/ES256 verification).
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"` // RSA modulus
	E   string `json:"e"` // RSA exponent
	Crv string `json:"crv"`
	X   string `json:"x"` // EC coordinate
	Y   string `json:"y"` // EC coordinate
}

// parseJWKS parses a JWKS document into a kid->public-key map. Keys without a
// kid, non-signing keys, unsupported types, and individually malformed keys are
// skipped so one bad entry cannot poison the whole set (the missing kid then
// simply fails closed at lookup time).
func parseJWKS(body []byte) (map[string]any, error) {
	var set struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, errors.E(errors.KindExternal, "upstream_error",
			"auth: parse JWKS document", err, errors.Op("auth.JWKS.parse"))
	}
	out := make(map[string]any, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kid == "" || (k.Use != "" && k.Use != "sig") {
			continue
		}
		var (
			key any
			err error
		)
		switch k.Kty {
		case "RSA":
			key, err = rsaFromJWK(k.N, k.E)
		case "EC":
			key, err = ecFromJWK(k.Crv, k.X, k.Y)
		default:
			continue
		}
		if err != nil {
			continue
		}
		out[k.Kid] = key
	}
	return out, nil
}

// errBadJWK marks a structurally invalid JWK; parseJWKS skips such keys.
var errBadJWK = errors.E(errors.KindExternal, "upstream_error", "auth: malformed JWK", errors.Op("auth.JWKS.parse"))

// rsaFromJWK builds an *rsa.PublicKey from base64url modulus/exponent.
func rsaFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := b64uDecode(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := b64uDecode(eB64)
	if err != nil {
		return nil, err
	}
	if len(nBytes) == 0 || len(eBytes) == 0 {
		return nil, errBadJWK
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	if e < 2 {
		return nil, errBadJWK
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

// ecFromJWK builds an *ecdsa.PublicKey from a named curve and base64url X/Y. It
// rejects coordinates outside the field so an obviously invalid point never
// reaches signature verification; keys come from the trusted issuer over https.
func ecFromJWK(crv, xB64, yB64 string) (*ecdsa.PublicKey, error) {
	var curve elliptic.Curve
	switch crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, errBadJWK
	}
	xB, err := b64uDecode(xB64)
	if err != nil {
		return nil, err
	}
	yB, err := b64uDecode(yB64)
	if err != nil {
		return nil, err
	}
	x := new(big.Int).SetBytes(xB)
	y := new(big.Int).SetBytes(yB)
	p := curve.Params().P
	if len(xB) == 0 || len(yB) == 0 || x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
		return nil, errBadJWK
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

// b64uDecode decodes base64url, tolerating optional padding some IdPs emit.
func b64uDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

// unknownKID is the opaque fail-closed error for a kid the JWKS does not carry.
func unknownKID() error {
	return errors.E(errors.KindUnauthenticated, "unauthenticated", "unknown key id", errors.Op("auth.JWKS.Key"))
}

// validateHTTPSURL requires https, allowing http only for loopback hosts so the
// httptest-based tests and local IdPs work without weakening production.
func validateHTTPSURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.E(errors.KindInternal, "internal",
			"auth: invalid JWKS/issuer URL", errors.Op("auth.JWKS"))
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && isLoopbackHost(u.Hostname()) {
		return nil
	}
	return errors.E(errors.KindInternal, "internal",
		"auth: JWKS/issuer URL must use https", errors.Op("auth.JWKS"))
}

// isLoopbackHost reports whether host is localhost or a loopback IP literal.
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
