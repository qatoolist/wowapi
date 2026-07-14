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
//
// T5 (SEC-01): ImpersonatorUserID and BreakGlass are no longer trusted from
// these claim fields. When GrantID is present, Verifier.Actor resolves the
// verified grant row and populates those Actor fields from that row only.
// The claim fields are retained for claim-shape compatibility during cutover.
type Claims struct {
	jwt.RegisteredClaims
	TenantID           uuid.UUID `json:"tenant_id"`
	CapacityID         uuid.UUID `json:"capacity_id,omitempty"`
	GrantID            uuid.UUID `json:"grant_id,omitempty"`
	ImpersonatorUserID uuid.UUID `json:"impersonator_user_id,omitempty"`
	BreakGlass         bool      `json:"break_glass,omitempty"`
	// AuthTime is the standard auth_time claim (OIDC Core §2, ISO8601 numeric
	// date). It records when the user authenticated at the IdP and drives
	// step-up freshness enforcement (SEC-01 T6). Absent or malformed auth_time
	// is treated as an unset claim by the verifier; freshness enforcement
	// interprets an unset AuthTime as failing freshness.
	AuthTime *jwt.NumericDate `json:"auth_time,omitempty"`
	// ACR is the standard authentication-context-class-reference claim (OIDC
	// Core §2). It is propagated to authz.Actor.ACR so policies can reason
	// about the assurance class, but it does not by itself gate step-up.
	ACR string `json:"acr,omitempty"`
	// AMR is the standard authentication-methods-references claim (RFC 8176,
	// e.g. ["pwd","mfa"]) surfaced by the IdP. Verifier.Actor propagates it to
	// authz.Actor.AMR, which drives step-up (MFA) enforcement (roadmap S3). A
	// malformed amr in the token (wrong JSON shape) fails the claims decode in
	// Verify, so Actor never sees a token whose amr could not be parsed.
	AMR []string `json:"amr,omitempty"`
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

// ResolvedGrant carries the verified privileged-session state returned by a
// PrincipalStore grant lookup. It is the only source of truth for populating
// authz.Actor.ImpersonatorUserID and authz.Actor.BreakGlass (SEC-01 T5).
type ResolvedGrant struct {
	ImpersonatorUserID uuid.UUID
	BreakGlass         bool
}

// GrantRejection names why a privileged-session grant was rejected by the
// resolver. Each value maps to one of the adversarial conditions in SEC-01 T5.
type GrantRejection string

const (
	// GrantRejectionExpired is returned when the grant's expires_at has passed
	// or its status is 'expired'.
	GrantRejectionExpired GrantRejection = "grant_expired"
	// GrantRejectionRevoked is returned when the grant's status is 'revoked' or
	// revoked_at is set.
	GrantRejectionRevoked GrantRejection = "grant_revoked"
	// GrantRejectionWrongTenant is returned when the grant row's tenant_id does
	// not match the actor's tenant. (The lookup itself is tenant-scoped, so this
	// condition also covers a forged grant ID from another tenant.)
	GrantRejectionWrongTenant GrantRejection = "grant_wrong_tenant"
	// GrantRejectionWrongActor is returned when the grant does not authorize
	// this actor (e.g. an impersonation grant for a different user, or a
	// break-glass grant issued to a different actor).
	GrantRejectionWrongActor GrantRejection = "grant_wrong_actor"
	// GrantRejectionNotFound is returned when the grant ID does not identify any
	// grant row (forged/unknown ID).
	GrantRejectionNotFound GrantRejection = "grant_not_found"
	// GrantRejectionUnauthorizedApprover is returned when the grant's approver
	// is missing or is not entitled to approve privileged sessions.
	GrantRejectionUnauthorizedApprover GrantRejection = "grant_unauthorized_approver"
)

// IsGrantRejection reports whether err is a privileged-session rejection with
// reason r. It walks the error chain and inspects each structured error's code;
// a nil error is false.
func IsGrantRejection(err error, r GrantRejection) bool {
	for err != nil {
		if e, ok := errors.As(err); ok {
			if e.Code == string(r) {
				return true
			}
			// Descend past this *Error's wrapped cause to catch rejections
			// re-wrapped by Verifier.Actor.
			err = e.Err
			continue
		}
		break
	}
	return false
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

// AssurancePrincipalStore is the additive v1 extension used for live tenant
// membership, capacity-count, and privileged-session checks. Legacy
// PrincipalStore implementations remain source-compatible; production stores
// should implement this interface to enable the stricter assurance posture.
type AssurancePrincipalStore interface {
	PrincipalStore
	ActiveTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) error
	ActiveCapacityCount(ctx context.Context, userID, tenantID uuid.UUID) (int, error)
	ResolveGrant(ctx context.Context, userID, tenantID, grantID uuid.UUID) (*ResolvedGrant, error)
}

// Actor maps validated Claims onto an authz.Actor. It resolves the framework
// user id from the subject via ps (an unknown subject → KindUnauthenticated) and
// verifies the user's live tenant membership unconditionally whenever the token
// carries a TenantID. A zero tenant claim is rejected before any tenant-bound
// database work.
//
// T4 (capacity selection): when the token does not carry an explicit CapacityID,
// Actor counts the user's active capacities in the tenant. If the count is
// greater than one, the request is rejected with KindValidation — the actor must
// present an explicit, server-side-validated capacity choice. When CapacityID is
// present, it is validated in the tenant (a mismatch → KindForbidden).
//
// T5 (privileged-session resolver): ImpersonatorUserID and BreakGlass are
// populated only from a verified identity_grant row looked up by claims.GrantID.
// Direct claim-copy of those fields is never used. A forged or adversarial grant
// (expired, revoked, wrong-tenant, wrong-actor, unknown, unauthorized-approver)
// is rejected with KindForbidden and a GrantRejection code.
func (v *Verifier) Actor(ctx context.Context, claims Claims, ps PrincipalStore) (authz.Actor, error) {
	subject := claims.Subject()
	if subject == "" {
		return authz.Actor{}, unauth("missing subject", nil)
	}

	userID, err := ps.UserIDBySubject(ctx, subject)
	if err != nil {
		return authz.Actor{}, unauth("unknown subject", err)
	}

	assurance, hasAssurance := ps.(AssurancePrincipalStore)
	if claims.TenantID != uuid.Nil {
		if hasAssurance {
			if err := assurance.ActiveTenantAccess(ctx, userID, claims.TenantID); err != nil {
				return authz.Actor{}, errors.E(errors.KindForbidden, "permission_denied",
					"tenant access not permitted", err, errors.Op("auth.Actor"))
			}
		}
	} else {
		return authz.Actor{}, errors.E(errors.KindValidation, "validation_failed",
			"tenant claim required", errors.Op("auth.Actor"))
	}

	if claims.CapacityID != uuid.Nil {
		if err := ps.ValidateCapacity(ctx, userID, claims.TenantID, claims.CapacityID); err != nil {
			return authz.Actor{}, errors.E(errors.KindForbidden, "permission_denied",
				"capacity not permitted", err, errors.Op("auth.Actor"))
		}
	} else {
		count := 0
		if hasAssurance {
			var err error
			count, err = assurance.ActiveCapacityCount(ctx, userID, claims.TenantID)
			if err != nil {
				return authz.Actor{}, errors.E(errors.KindForbidden, "permission_denied",
					"capacity count unavailable", err, errors.Op("auth.Actor"))
			}
		}
		if count > 1 {
			return authz.Actor{}, errors.E(errors.KindValidation, "validation_failed",
				"explicit capacity selection required", errors.Op("auth.Actor"))
		}
	}

	actor := authz.Actor{
		Kind:             authz.ActorUser,
		UserID:           userID,
		CapacityID:       claims.CapacityID,
		TenantID:         claims.TenantID,
		CredentialScheme: authz.CredentialUser,
		AMR:              claims.AMR,
		ACR:              claims.ACR,
	}
	if claims.AuthTime != nil {
		actor.AuthTime = claims.AuthTime.Time
	}

	if claims.GrantID != uuid.Nil {
		if !hasAssurance {
			return authz.Actor{}, errors.E(errors.KindForbidden, "permission_denied",
				"principal store does not support privileged sessions", errors.Op("auth.Actor"))
		}
		grant, err := assurance.ResolveGrant(ctx, userID, claims.TenantID, claims.GrantID)
		if err != nil {
			return authz.Actor{}, errors.E(errors.KindForbidden, "permission_denied",
				"privileged session not permitted", err, errors.Op("auth.Actor"))
		}
		actor.GrantID = claims.GrantID
		actor.ImpersonatorUserID = grant.ImpersonatorUserID
		actor.BreakGlass = grant.BreakGlass
	}

	return actor, nil
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
