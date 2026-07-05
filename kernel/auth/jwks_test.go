package auth_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/auth"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/errors"
)

// This file exercises the production JWKS-over-HTTPS KeySource and the
// httpx.Authenticator adapter against a locally-generated signing key served by
// an httptest JWKS server: valid RS256/ES256 tokens are accepted; wrong
// iss/aud, expired, bad signature, unknown kid, and HS256 (alg-confusion) are
// rejected; an unreachable JWKS endpoint is a HARD fault (not a 401); and the
// cache serves a known kid without refetching.

const (
	jwksIssuer   = "https://idp.example.test"
	jwksAudience = "wowapi"
	rsaKID       = "rsa-1"
	ecKID        = "ec-1"
)

func b64u(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func rsaJWK(kid string, pub *rsa.PublicKey) map[string]any {
	return map[string]any{
		"kty": "RSA", "kid": kid, "use": "sig", "alg": "RS256",
		"n": b64u(pub.N.Bytes()),
		"e": b64u(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func ecJWK(t *testing.T, kid string, pub *ecdsa.PublicKey) map[string]any {
	t.Helper()
	// Uncompressed SEC1 point (0x04 || X || Y); avoids the Go 1.26-deprecated
	// direct .X/.Y coordinate reads.
	raw, err := pub.Bytes()
	if err != nil {
		t.Fatalf("ec pub bytes: %v", err)
	}
	byteLen := (len(raw) - 1) / 2
	return map[string]any{
		"kty": "EC", "kid": kid, "use": "sig", "alg": "ES256", "crv": "P-256",
		"x": b64u(raw[1 : 1+byteLen]),
		"y": b64u(raw[1+byteLen:]),
	}
}

// jwksServer serves a JWKS document and an OIDC discovery document, counting
// hits so tests can assert caching and discovery behavior.
type jwksServer struct {
	*httptest.Server
	jwks          []byte
	status        int // status code for /jwks (200 unless overridden)
	jwksHits      atomic.Int64
	discoveryHits atomic.Int64
}

func newJWKSServer(t *testing.T, keys ...map[string]any) *jwksServer {
	t.Helper()
	doc, err := json.Marshal(map[string]any{"keys": keys})
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	s := &jwksServer{jwks: doc, status: http.StatusOK}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			s.discoveryHits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issuer":   s.URL,
				"jwks_uri": s.URL + "/jwks",
			})
		case "/jwks":
			s.jwksHits.Add(1)
			if s.status != http.StatusOK {
				w.WriteHeader(s.status)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(s.jwks)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.Close)
	return s
}

func stdClaims(sub, iss, aud string, tenant uuid.UUID, exp time.Duration) auth.Claims {
	now := time.Now()
	return auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    iss,
			Audience:  jwt.ClaimStrings{aud},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
		},
		TenantID: tenant,
	}
}

func signRS(t *testing.T, key *rsa.PrivateKey, kid string, c auth.Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign RS256: %v", err)
	}
	return s
}

func signES(t *testing.T, key *ecdsa.PrivateKey, kid string, c auth.Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, c)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign ES256: %v", err)
	}
	return s
}

// jwksFixture bundles a signing keypair, its JWKS server, and a Verifier wired
// to the server's explicit jwks_uri.
type jwksFixture struct {
	rsa    *rsa.PrivateKey
	ec     *ecdsa.PrivateKey
	server *jwksServer
	verif  *auth.Verifier
}

func newJWKSFixture(t *testing.T) *jwksFixture {
	t.Helper()
	rk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa keygen: %v", err)
	}
	ek, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ec keygen: %v", err)
	}
	srv := newJWKSServer(t, rsaJWK(rsaKID, &rk.PublicKey), ecJWK(t, ecKID, &ek.PublicKey))
	keys, err := auth.NewJWKSKeySource(auth.JWKSConfig{
		Issuer:  jwksIssuer,
		JWKSURI: srv.URL + "/jwks",
	})
	if err != nil {
		t.Fatalf("NewJWKSKeySource: %v", err)
	}
	v := auth.NewVerifier(keys, auth.Config{Issuer: jwksIssuer, Audience: jwksAudience})
	return &jwksFixture{rsa: rk, ec: ek, server: srv, verif: v}
}

func TestJWKS_VerifyTable(t *testing.T) {
	f := newJWKSFixture(t)
	tenant := uuid.New()

	// A forged HS256 token whose HMAC secret is the RSA modulus bytes — the
	// classic algorithm-confusion attack — must be rejected, not verified.
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, stdClaims("idp|a", jwksIssuer, jwksAudience, tenant, time.Hour))
	forged.Header["kid"] = rsaKID
	hs256, err := forged.SignedString(f.rsa.N.Bytes())
	if err != nil {
		t.Fatalf("sign forged HS256: %v", err)
	}

	// Tamper the signature of an otherwise-valid RS256 token. Flip the FIRST
	// signature char (always carries significant bits) so the decoded signature
	// is guaranteed to differ.
	valid := signRS(t, f.rsa, rsaKID, stdClaims("idp|a", jwksIssuer, jwksAudience, tenant, time.Hour))
	parts := strings.Split(valid, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token shape")
	}
	sig := []byte(parts[2])
	if sig[0] == 'A' {
		sig[0] = 'B'
	} else {
		sig[0] = 'A'
	}
	tampered := parts[0] + "." + parts[1] + "." + string(sig)

	cases := []struct {
		name       string
		token      string
		wantOK     bool
		wantSubmap string
	}{
		{"valid RS256", valid, true, "idp|a"},
		{"valid ES256", signES(t, f.ec, ecKID, stdClaims("idp|e", jwksIssuer, jwksAudience, tenant, time.Hour)), true, "idp|e"},
		{"wrong issuer", signRS(t, f.rsa, rsaKID, stdClaims("idp|a", "https://evil.test", jwksAudience, tenant, time.Hour)), false, ""},
		{"wrong audience", signRS(t, f.rsa, rsaKID, stdClaims("idp|a", jwksIssuer, "other-api", tenant, time.Hour)), false, ""},
		{"expired", signRS(t, f.rsa, rsaKID, stdClaims("idp|a", jwksIssuer, jwksAudience, tenant, -time.Hour)), false, ""},
		{"unknown kid", signRS(t, f.rsa, "no-such-kid", stdClaims("idp|a", jwksIssuer, jwksAudience, tenant, time.Hour)), false, ""},
		{"bad signature", tampered, false, ""},
		{"alg confusion HS256", hs256, false, ""},
		{"empty", "", false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := f.verif.Verify(context.Background(), tc.token)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("want accept, got %v", err)
				}
				if claims.Subject() != tc.wantSubmap {
					t.Fatalf("subject: got %q want %q", claims.Subject(), tc.wantSubmap)
				}
				return
			}
			assertUnauthenticated(t, err)
			if tc.token != "" {
				assertNoLeak(t, err, tc.token)
			}
		})
	}
}

func TestJWKS_Discovery(t *testing.T) {
	rk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa keygen: %v", err)
	}
	srv := newJWKSServer(t, rsaJWK(rsaKID, &rk.PublicKey))
	// No explicit JWKSURI: the source must discover it from the issuer.
	keys, err := auth.NewJWKSKeySource(auth.JWKSConfig{Issuer: srv.URL})
	if err != nil {
		t.Fatalf("NewJWKSKeySource: %v", err)
	}
	v := auth.NewVerifier(keys, auth.Config{Issuer: srv.URL, Audience: jwksAudience})

	tok := signRS(t, rk, rsaKID, stdClaims("idp|a", srv.URL, jwksAudience, uuid.New(), time.Hour))
	if _, err := v.Verify(context.Background(), tok); err != nil {
		t.Fatalf("discovery verify: %v", err)
	}
	if srv.discoveryHits.Load() == 0 {
		t.Fatal("expected the discovery endpoint to be hit")
	}
}

func TestJWKS_CachesKnownKID(t *testing.T) {
	f := newJWKSFixture(t)
	tok := signRS(t, f.rsa, rsaKID, stdClaims("idp|a", jwksIssuer, jwksAudience, uuid.New(), time.Hour))
	for i := 0; i < 3; i++ {
		if _, err := f.verif.Verify(context.Background(), tok); err != nil {
			t.Fatalf("verify #%d: %v", i, err)
		}
	}
	if got := f.server.jwksHits.Load(); got != 1 {
		t.Fatalf("expected a single JWKS fetch (cache hit thereafter), got %d", got)
	}
}

func TestJWKS_UnreachableIsHardFault(t *testing.T) {
	f := newJWKSFixture(t)
	f.server.status = http.StatusInternalServerError // JWKS endpoint returns 500

	tok := signRS(t, f.rsa, rsaKID, stdClaims("idp|a", jwksIssuer, jwksAudience, uuid.New(), time.Hour))
	_, err := f.verif.Verify(context.Background(), tok)
	if err == nil {
		t.Fatal("want error, got nil")
	}
	// A transport fault must NOT be flattened into a 401 — it is a hard fault so a
	// composite authenticator short-circuits instead of masking an outage.
	if k := errors.KindOf(err); k == errors.KindUnauthenticated {
		t.Fatalf("JWKS outage must be a hard fault, got KindUnauthenticated: %v", err)
	}
	if k := errors.KindOf(err); k != errors.KindExternal {
		t.Fatalf("want KindExternal, got %v (%v)", k, err)
	}
}

func TestJWKS_RejectsNonHTTPSURL(t *testing.T) {
	if _, err := auth.NewJWKSKeySource(auth.JWKSConfig{JWKSURI: "http://idp.example.test/jwks"}); err == nil {
		t.Fatal("expected non-loopback http jwks_uri to be rejected")
	}
	if _, err := auth.NewJWKSKeySource(auth.JWKSConfig{}); err == nil {
		t.Fatal("expected empty config (no issuer/jwks_uri) to be rejected")
	}
}

// --- Authenticator adapter (httpx.Authenticator) ---

func TestAuthenticator_Accepts(t *testing.T) {
	f := newJWKSFixture(t)
	userID := uuid.New()
	tenant := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|human"}
	a := auth.NewAuthenticator(f.verif, ps)

	tok := signES(t, f.ec, ecKID, stdClaims("idp|human", jwksIssuer, jwksAudience, tenant, time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	actor, err := a.Authenticate(req)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if actor.Kind != authz.ActorUser || actor.UserID != userID || actor.TenantID != tenant {
		t.Fatalf("actor mismatch: %+v", actor)
	}
}

func TestAuthenticator_DeclinesNonJWT(t *testing.T) {
	f := newJWKSFixture(t)
	a := auth.NewAuthenticator(f.verif, fakePrincipalStore{subject: "idp|human"})

	// No Authorization header → decline so a composite falls through.
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if _, err := a.Authenticate(req); errors.KindOf(err) != errors.KindUnauthenticated {
		t.Fatalf("missing bearer must decline as unauthenticated, got %v", err)
	}

	// An API-key-shaped bearer (not a JWT) → decline, not a hard fault.
	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req2.Header.Set("Authorization", "Bearer wowapi_deadbeef_secret")
	if _, err := a.Authenticate(req2); errors.KindOf(err) != errors.KindUnauthenticated {
		t.Fatalf("api-key token must decline as unauthenticated, got %v", err)
	}
}
