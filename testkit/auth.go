package testkit

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/auth"
)

// Default token parameters the auth.Verifier is expected to accept. Tests
// override them via TokenOptions to exercise the verifier's checks.
const (
	defaultTestIssuer   = "wowapi-test"
	defaultTestAudience = "wowapi"
	defaultTestExpiry   = time.Hour
	defaultTestKID      = "test-key-1"
)

// TokenIssuer holds a locally-generated RSA keypair and mints RS256 JWTs the
// auth.Verifier accepts. It is the fixture every authenticated test uses
// (blueprint 08 §2, D-0037): pair KeySource() with an auth.Verifier, then Issue
// tokens for the subjects/tenants/capacities under test.
type TokenIssuer struct {
	key *rsa.PrivateKey
	kid string
}

// NewTokenIssuer generates a fresh 2048-bit RSA keypair and returns an issuer
// keyed under a stable test kid. It panics on key-generation failure since a
// test harness cannot proceed without it.
func NewTokenIssuer() *TokenIssuer {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("testkit: generating RSA test key: " + err.Error())
	}
	return &TokenIssuer{key: key, kid: defaultTestKID}
}

// KeySource returns an auth.KeySource exposing this issuer's public key under
// its kid, ready to wire into an auth.Verifier.
func (ti *TokenIssuer) KeySource() auth.KeySource {
	return auth.NewStaticKeySource(map[string]any{ti.kid: &ti.key.PublicKey})
}

// PublicKey returns the issuer's RSA public key. Tests use it to construct
// negative fixtures (e.g. algorithm-confusion forgeries).
func (ti *TokenIssuer) PublicKey() *rsa.PublicKey {
	return &ti.key.PublicKey
}

// tokenConfig accumulates TokenOption mutations.
type tokenConfig struct {
	issuer   string
	audience string
	expiry   time.Duration
	grantID  uuid.UUID
	authTime *time.Time
	acr      string
	amr      []string
}

// TokenOption customizes a minted token so tests can drive the verifier's
// issuer/audience/expiry/assurance checks.
type TokenOption func(*tokenConfig)

// WithIssuer overrides the iss claim (default "wowapi-test").
func WithIssuer(iss string) TokenOption {
	return func(c *tokenConfig) { c.issuer = iss }
}

// WithAudience overrides the aud claim (default "wowapi").
func WithAudience(aud string) TokenOption {
	return func(c *tokenConfig) { c.audience = aud }
}

// WithExpiry sets the token lifetime relative to now (default +1h). A negative
// value mints an already-expired token.
func WithExpiry(d time.Duration) TokenOption {
	return func(c *tokenConfig) { c.expiry = d }
}

// WithGrantID sets the grant_id claim used by the privileged-session resolver
// (SEC-01 T5). The framework resolves attribution server-side.
func WithGrantID(id uuid.UUID) TokenOption {
	return func(c *tokenConfig) { c.grantID = id }
}

// WithAuthTime sets the standard auth_time claim (OIDC Core §2). When not
// used, auth_time is omitted from the token, so freshness enforcement sees a
// zero AuthTime. Pass a pointer so nil/unset is distinguishable from the
// zero instant.
func WithAuthTime(at time.Time) TokenOption {
	return func(c *tokenConfig) { c.authTime = &at }
}

// WithACR sets the standard acr (authentication-context-class-reference)
// claim (OIDC Core §2), propagated to authz.Actor.ACR.
func WithACR(acr string) TokenOption {
	return func(c *tokenConfig) { c.acr = acr }
}

// WithAMR sets the standard amr (authentication-methods-references) claim
// (RFC 8176, e.g. WithAMR("pwd", "mfa")), driving the auth.Verifier's
// propagation into authz.Actor.AMR and the evaluator's step-up check.
func WithAMR(amr ...string) TokenOption {
	return func(c *tokenConfig) { c.amr = amr }
}

// Issue mints a signed RS256 JWT for subject in tenantID with the given
// capacityID (pass uuid.Nil to omit the capacity claim). Standard claims
// (iss/aud/exp/iat/nbf) default to the values the auth.Verifier expects and are
// tunable via opts. The kid header is set so KeySource resolves the key.
func (ti *TokenIssuer) Issue(subject string, tenantID, capacityID uuid.UUID, opts ...TokenOption) string {
	cfg := tokenConfig{
		issuer:   defaultTestIssuer,
		audience: defaultTestAudience,
		expiry:   defaultTestExpiry,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	now := time.Now()
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    cfg.issuer,
			Audience:  jwt.ClaimStrings{cfg.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.expiry)),
		},
		TenantID:   tenantID,
		CapacityID: capacityID,
		GrantID:    cfg.grantID,
		ACR:        cfg.acr,
		AMR:        cfg.amr,
	}
	if cfg.authTime != nil {
		claims.AuthTime = jwt.NewNumericDate(*cfg.authTime)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = ti.kid
	signed, err := token.SignedString(ti.key)
	if err != nil {
		panic("testkit: signing test token: " + err.Error())
	}
	return signed
}
