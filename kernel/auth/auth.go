// Package auth is wowapi's authentication kernel: it verifies OIDC/JWT bearer
// tokens against an injectable KeySource (JWKS-over-HTTPS in production, a local
// signer in tests) and maps validated claims onto an authz.Actor after the app
// resolves the framework user id and active capacity (D-0037, 01 §3). The
// Authenticator type adapts a Verifier to the framework's structural
// httpx.Authenticator so it can be the user leg of a product's composite.
//
// Two properties are structural, not configurable:
//   - asymmetric signatures only: the verifier asserts the token's signing method
//     is RSA or ECDSA (RS256/ES256) before touching the key, so "alg":"none" and
//     HMAC tokens are rejected outright (algorithm-confusion defense);
//   - opaque failures: every bad/missing/expired/wrong-issuer/wrong-audience/
//     unknown-kid/bad-signature case returns errors.E(KindUnauthenticated, ...)
//     and never echoes the token or key material. A KeySource transport fault
//     (unreachable JWKS) surfaces as a KindExternal HARD error, not a 401, so a
//     composite authenticator does not mask a transient outage as a clean reject.
//
// Import law (blueprint 04 §1; boundary lint): this package imports only stdlib,
// kernel/errors, kernel/authz, google/uuid and the jwt library — never module,
// app, adapters or testkit. The DB-backed user/capacity lookup is injected as
// the PrincipalStore port, wired by the app.
package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/errors"
)

// defaultLeeway is the clock-skew tolerance applied to exp/nbf/iat when the
// Config leaves Leeway unset.
const defaultLeeway = 30 * time.Second

// KeySource resolves a token-verification key by its key id (kid). Production
// wires a caching JWKS-over-HTTPS adapter; tests wire a static in-memory source.
type KeySource interface {
	// Key returns the verification key (e.g. *rsa.PublicKey) for the given kid.
	Key(ctx context.Context, kid string) (any, error)
}

// Config parameterizes a Verifier.
type Config struct {
	Issuer   string        // expected iss
	Audience string        // expected aud
	Leeway   time.Duration // clock-skew tolerance (default 30s)
}

// Claims carries the wowapi-specific token payload alongside the standard
// registered claims. Subject (sub) maps to a user's idp_subject; TenantID and
// the optional CapacityID/ImpersonatorUserID/BreakGlass drive the authz.Actor.
type Claims struct {
	jwt.RegisteredClaims
	TenantID           uuid.UUID `json:"tenant_id"`
	CapacityID         uuid.UUID `json:"capacity_id,omitempty"`
	ImpersonatorUserID uuid.UUID `json:"impersonator_user_id,omitempty"`
	BreakGlass         bool      `json:"break_glass,omitempty"`
}

// Subject returns the token subject (sub), which maps to a user's idp_subject.
func (c Claims) Subject() string { return c.RegisteredClaims.Subject }

// Verifier parses and validates bearer tokens and maps their claims to actors.
type Verifier struct {
	keys   KeySource
	issuer string
	aud    string
	leeway time.Duration
}

// NewVerifier builds a Verifier over keys and cfg. A zero cfg.Leeway defaults to
// 30s.
func NewVerifier(keys KeySource, cfg Config) *Verifier {
	leeway := cfg.Leeway
	if leeway <= 0 {
		leeway = defaultLeeway
	}
	return &Verifier{
		keys:   keys,
		issuer: cfg.Issuer,
		aud:    cfg.Audience,
		leeway: leeway,
	}
}

// unauth builds an opaque KindUnauthenticated error. The wrapped cause is kept
// for logs only (never reaches the wire, 04 §5); callers must never pass the
// token or key material as msg.
func unauth(msg string, cause error) error {
	if cause != nil {
		return errors.E(errors.KindUnauthenticated, "unauthenticated", msg, cause, errors.Op("auth.Verify"))
	}
	return errors.E(errors.KindUnauthenticated, "unauthenticated", msg, errors.Op("auth.Verify"))
}

// Verify parses and validates an RS256 bearer token: it asserts the signing
// method is RSA (rejecting "alg":"none"/HMAC), resolves the key via KeySource by
// the token's kid header, and checks iss/aud/exp/nbf with the configured leeway.
// On success it returns the validated Claims; every failure mode returns
// errors.E(KindUnauthenticated, ...) with no token or key material in the error.
func (v *Verifier) Verify(ctx context.Context, tokenString string) (Claims, error) {
	if tokenString == "" {
		return Claims{}, unauth("missing token", nil)
	}

	claims := &Claims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256", "ES256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.aud),
		jwt.WithLeeway(v.leeway),
		jwt.WithExpirationRequired(),
	)

	keyfunc := func(token *jwt.Token) (any, error) {
		// Algorithm-confusion defense: assert a concrete asymmetric signing
		// method before we ever look up a key. WithValidMethods above already
		// gates on the "alg" string; this second check binds to the *type* so an
		// attacker cannot coerce an HMAC verification against public-key bytes.
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA, *jwt.SigningMethodECDSA:
		default:
			return nil, unauth("unexpected signing method", nil)
		}
		kid, _ := token.Header["kid"].(string)
		key, err := v.keys.Key(ctx, kid)
		if err != nil {
			// An unknown-kid (KindUnauthenticated) stays an opaque 401; any other
			// kind (e.g. KindExternal when the JWKS endpoint is unreachable) is a
			// hard fault that must propagate rather than be flattened to a 401.
			if errors.KindOf(err) == errors.KindUnauthenticated {
				return nil, unauth("unknown key id", err)
			}
			return nil, err
		}
		return key, nil
	}

	if _, err := parser.ParseWithClaims(tokenString, claims, keyfunc); err != nil {
		// A kernel *Error in the chain was surfaced deliberately by keyfunc/the
		// KeySource: keep its Kind (an opaque 401 for unknown-kid, or a hard fault
		// such as KindExternal for an unreachable JWKS). A raw jwt parse/validation
		// failure (bad signature, expired, wrong iss/aud) carries no Kind — wrap it
		// as opaque unauthenticated.
		if e, ok := errors.As(err); ok {
			return Claims{}, e
		}
		return Claims{}, unauth("invalid token", err)
	}

	return *claims, nil
}

// PrincipalStore resolves the framework user id from the IdP subject and
// confirms the capacity belongs to that user in the tenant. Implemented in the
// app/adapters DB layer (kernel/auth may not import a database).
type PrincipalStore interface {
	// UserIDBySubject returns the framework user id for an IdP subject.
	UserIDBySubject(ctx context.Context, subject string) (uuid.UUID, error)
	// ValidateCapacity returns a non-nil error if capacityID is not an active
	// capacity of userID in tenantID.
	ValidateCapacity(ctx context.Context, userID, tenantID, capacityID uuid.UUID) error
}

// Actor maps validated Claims onto an authz.Actor. It resolves the framework
// user id from the subject via ps (an unknown subject → KindUnauthenticated) and,
// when a capacity is present, confirms it belongs to that user in the tenant
// (a mismatch → KindForbidden). Impersonation and break-glass carry through.
func (v *Verifier) Actor(ctx context.Context, claims Claims, ps PrincipalStore) (authz.Actor, error) {
	subject := claims.Subject()
	if subject == "" {
		return authz.Actor{}, unauth("missing subject", nil)
	}

	userID, err := ps.UserIDBySubject(ctx, subject)
	if err != nil {
		return authz.Actor{}, unauth("unknown subject", err)
	}

	if claims.CapacityID != uuid.Nil {
		if err := ps.ValidateCapacity(ctx, userID, claims.TenantID, claims.CapacityID); err != nil {
			return authz.Actor{}, errors.E(errors.KindForbidden, "permission_denied",
				"capacity not permitted", err, errors.Op("auth.Actor"))
		}
	}

	return authz.Actor{
		Kind:               authz.ActorUser,
		UserID:             userID,
		CapacityID:         claims.CapacityID,
		TenantID:           claims.TenantID,
		ImpersonatorUserID: claims.ImpersonatorUserID,
		BreakGlass:         claims.BreakGlass,
	}, nil
}

// Authenticator adapts a Verifier to the framework's structural
// httpx.Authenticator: it reads an "Authorization: Bearer <jwt>" header,
// verifies the token, and maps its claims to an authz.Actor via the
// PrincipalStore. It is the OIDC/JWT user leg of a product's composite
// authenticator (roadmap S1/CA-2). It satisfies httpx.Authenticator
// structurally — kernel/auth never imports kernel/httpx (import law).
//
// Decline vs. fault: a missing bearer token or a non-JWT token (e.g. an API key)
// yields KindUnauthenticated so a composite falls through to the next scheme; a
// JWKS/transport fault propagates as a hard error so it is not masked as a 401.
type Authenticator struct {
	v  *Verifier
	ps PrincipalStore
}

// NewAuthenticator builds the OIDC/JWT authenticator over a Verifier and the
// app-supplied PrincipalStore (subject → framework user id + capacity check).
func NewAuthenticator(v *Verifier, ps PrincipalStore) *Authenticator {
	return &Authenticator{v: v, ps: ps}
}

// Authenticate resolves the user actor from the request's bearer JWT. It
// declines (KindUnauthenticated) when no bearer token is present.
func (a *Authenticator) Authenticate(r *http.Request) (authz.Actor, error) {
	tok := bearerToken(r)
	if tok == "" {
		return authz.Actor{}, unauth("missing bearer token", nil)
	}
	claims, err := a.v.Verify(r.Context(), tok)
	if err != nil {
		return authz.Actor{}, err
	}
	return a.v.Actor(r.Context(), claims, a.ps)
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(r *http.Request) string {
	if after, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return after
	}
	return ""
}
