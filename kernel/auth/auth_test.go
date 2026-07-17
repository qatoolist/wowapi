package auth_test

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/auth"
	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// fakePrincipalStore maps a fixed subject → user and validates tenant access
// and capacities against allowlists, so Actor mapping is testable without a
// database. It also stubs the T4/T5 extensions: ActiveCapacityCount and
// ResolveGrant.
type fakePrincipalStore struct {
	userID   uuid.UUID
	subject  string
	okTenant uuid.UUID           // tenant that userID has live access to
	okCap    uuid.UUID           // capacity that validates for userID
	capCount int                 // active capacity count returned by ActiveCapacityCount
	grant    *auth.ResolvedGrant // resolved grant returned by ResolveGrant
	grantErr error               // error returned by ResolveGrant
}

func (f fakePrincipalStore) UserIDBySubject(_ context.Context, subject string) (uuid.UUID, error) {
	if subject != f.subject {
		return uuid.Nil, errors.E(errors.KindUnauthenticated, "unauthenticated", "no such subject")
	}
	return f.userID, nil
}

func (f fakePrincipalStore) ActiveTenantAccess(_ context.Context, userID, tenantID uuid.UUID) error {
	if userID == f.userID && tenantID == f.okTenant {
		return nil
	}
	return errors.E(errors.KindForbidden, "permission_denied", "tenant access not permitted")
}

func (f fakePrincipalStore) ActiveCapacityCount(_ context.Context, userID, tenantID uuid.UUID) (int, error) {
	if userID == f.userID && tenantID == f.okTenant {
		return f.capCount, nil
	}
	return 0, errors.E(errors.KindForbidden, "permission_denied", "tenant access not permitted")
}

func (f fakePrincipalStore) ValidateCapacity(_ context.Context, userID, _ uuid.UUID, capacityID uuid.UUID) error {
	if userID == f.userID && capacityID == f.okCap {
		return nil
	}
	return errors.E(errors.KindForbidden, "permission_denied", "capacity not yours")
}

func (f fakePrincipalStore) ResolveGrant(_ context.Context, userID, tenantID, grantID uuid.UUID) (*auth.ResolvedGrant, error) {
	_ = userID
	_ = tenantID
	_ = grantID
	if f.grantErr != nil {
		return nil, f.grantErr
	}
	if f.grant == nil {
		return nil, errors.E(errors.KindForbidden, string(auth.GrantRejectionNotFound), "grant not found")
	}
	return f.grant, nil
}

func newVerifier(ti *testkit.TokenIssuer) *auth.Verifier {
	return auth.NewVerifier(ti.KeySource(), auth.Config{
		Issuer:   "wowapi-test",
		Audience: "wowapi",
	})
}

func assertUnauthenticated(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if got := errors.KindOf(err); got != errors.KindUnauthenticated {
		t.Fatalf("want KindUnauthenticated, got %v (%v)", got, err)
	}
}

// assertNoLeak fails if the error text contains any of the sensitive fragments.
func assertNoLeak(t *testing.T, err error, secrets ...string) {
	t.Helper()
	if err == nil {
		return
	}
	msg := err.Error()
	for _, s := range secrets {
		if s != "" && strings.Contains(msg, s) {
			t.Fatalf("error leaked sensitive material %q in %q", s, msg)
		}
	}
}

func TestVerify_ValidTokenMapsToActor(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	capID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: tenantID, okCap: capID}

	tok := ti.Issue("idp|alice", tenantID, capID)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.Subject() != "idp|alice" {
		t.Fatalf("subject: got %q", claims.Subject())
	}
	if claims.TenantID != tenantID {
		t.Fatalf("tenant: got %v want %v", claims.TenantID, tenantID)
	}

	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if actor.Kind != authz.ActorUser {
		t.Fatalf("kind: got %v", actor.Kind)
	}
	if actor.UserID != userID || actor.TenantID != tenantID || actor.CapacityID != capID {
		t.Fatalf("actor mismatch: %+v", actor)
	}
}

// TestVerify_AMRPropagatesToClaims proves the standard RFC 8176 amr claim
// round-trips through Verify into Claims.AMR.
func TestVerify_AMRPropagatesToClaims(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	tenantID := uuid.New()
	capID := uuid.New()
	tok := ti.Issue("idp|alice", tenantID, capID, testkit.WithAMR("pwd", "mfa"))

	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got := claims.AMR; len(got) != 2 || got[0] != "pwd" || got[1] != "mfa" {
		t.Fatalf("Claims.AMR = %v, want [pwd mfa]", got)
	}
}

// TestActor_AMRPropagates proves Verifier.Actor carries Claims.AMR through to
// authz.Actor.AMR — the plumbing the evaluator's step-up check depends on.
func TestActor_AMRPropagates(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	capID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: tenantID, okCap: capID}

	tok := ti.Issue("idp|alice", tenantID, capID, testkit.WithAMR("pwd", "mfa"))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if got := actor.AMR; len(got) != 2 || got[0] != "pwd" || got[1] != "mfa" {
		t.Fatalf("Actor.AMR = %v, want [pwd mfa]", got)
	}
}

// TestVerify_AuthTimeAndACRPropagatesToClaims proves the OIDC auth_time and
// acr claims round-trip through Verify into Claims (SEC-01 T6).
func TestVerify_AuthTimeAndACRPropagatesToClaims(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	tenantID := uuid.New()
	capID := uuid.New()
	authTime := time.Date(2026, 7, 3, 11, 30, 0, 0, time.UTC)
	tok := ti.Issue("idp|alice", tenantID, capID, testkit.WithAuthTime(authTime), testkit.WithACR("urn:mace:incommon:iap:silver"))

	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.AuthTime == nil || !claims.AuthTime.Equal(authTime) {
		t.Fatalf("Claims.AuthTime = %v, want %v", claims.AuthTime, authTime)
	}
	if claims.ACR != "urn:mace:incommon:iap:silver" {
		t.Fatalf("Claims.ACR = %q, want silver ACR", claims.ACR)
	}
}

// TestActor_AuthTimeACRAndCredentialSchemePropagates proves Verifier.Actor
// carries AuthTime, ACR, and the CredentialUser scheme through to the
// authz.Actor (SEC-01 T6/T7).
func TestActor_AuthTimeACRAndCredentialSchemePropagates(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	capID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: tenantID, okCap: capID}

	authTime := time.Date(2026, 7, 3, 11, 30, 0, 0, time.UTC)
	tok := ti.Issue("idp|alice", tenantID, capID, testkit.WithAuthTime(authTime), testkit.WithACR("silver"))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if !actor.AuthTime.Equal(authTime) {
		t.Fatalf("Actor.AuthTime = %v, want %v", actor.AuthTime, authTime)
	}
	if actor.ACR != "silver" {
		t.Fatalf("Actor.ACR = %q, want silver", actor.ACR)
	}
	if actor.CredentialScheme != authz.CredentialUser {
		t.Fatalf("Actor.CredentialScheme = %q, want user", actor.CredentialScheme)
	}
}

// TestVerify_NoAMRIsEmpty proves a token without an amr claim maps to a nil/
// empty Claims.AMR — no strong factor is asserted absent explicit evidence.
func TestVerify_NoAMRIsEmpty(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	tok := ti.Issue("idp|alice", uuid.New(), uuid.New())
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(claims.AMR) != 0 {
		t.Fatalf("Claims.AMR = %v, want empty", claims.AMR)
	}
}

// malformedAMRToken hand-assembles a JWT with a header+claims+signature of our
// choosing so a test can pin a claims JSON shape the typed TokenIssuer API
// cannot produce (e.g. amr as a bare string or an array of numbers). The JWT
// library decodes claims via json.Unmarshal in ParseUnverified BEFORE the
// signature is ever checked (golang-jwt/jwt/v5 parser.go), so a malformed amr
// payload fails the parse at the JSON-decode step — the signature bytes here
// never need to verify for this test to pin that behavior.
func malformedAMRToken(t *testing.T, claimsJSON string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT","kid":"test-key-1"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	sig := base64.RawURLEncoding.EncodeToString([]byte("not-a-real-signature"))
	return header + "." + payload + "." + sig
}

func TestVerify_MalformedAMRStringInsteadOfArray(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	now := time.Now().Unix()
	claimsJSON := fmt.Sprintf(`{"sub":"idp|alice","iss":"wowapi-test","aud":"wowapi",
		"iat":%d,"nbf":%d,"exp":%d,"tenant_id":%q,"amr":"mfa"}`,
		now, now, now+3600, uuid.New().String())

	_, err := v.Verify(context.Background(), malformedAMRToken(t, claimsJSON))
	// encoding/json refuses to unmarshal a JSON string into a []string field —
	// the whole Claims decode fails, so Verify must reject the token rather than
	// silently drop or coerce the malformed amr.
	assertUnauthenticated(t, err)
}

func TestVerify_MalformedAMRNonStringElements(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	now := time.Now().Unix()
	claimsJSON := fmt.Sprintf(`{"sub":"idp|alice","iss":"wowapi-test","aud":"wowapi",
		"iat":%d,"nbf":%d,"exp":%d,"tenant_id":%q,"amr":[1,2]}`,
		now, now, now+3600, uuid.New().String())

	_, err := v.Verify(context.Background(), malformedAMRToken(t, claimsJSON))
	assertUnauthenticated(t, err)
}

func TestActor_UnknownSubjectUnauthenticated(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	ps := fakePrincipalStore{userID: uuid.New(), subject: "idp|known"}

	tok := ti.Issue("idp|stranger", uuid.New(), uuid.Nil)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	assertUnauthenticated(t, err)
}

func TestActor_CapacityNotYoursForbidden(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	tenantID := uuid.New()
	ps := fakePrincipalStore{userID: uuid.New(), subject: "idp|carol", okTenant: tenantID, okCap: uuid.New()}

	tok := ti.Issue("idp|carol", tenantID, uuid.New()) // capacity not in allowlist
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if err == nil || errors.KindOf(err) != errors.KindForbidden {
		t.Fatalf("want KindForbidden, got %v", err)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	tok := ti.Issue("idp|alice", uuid.New(), uuid.Nil, testkit.WithExpiry(-time.Hour))
	_, err := v.Verify(context.Background(), tok)
	assertUnauthenticated(t, err)
	assertNoLeak(t, err, tok)
}

func TestVerify_WrongIssuer(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	tok := ti.Issue("idp|alice", uuid.New(), uuid.Nil, testkit.WithIssuer("evil-idp"))
	_, err := v.Verify(context.Background(), tok)
	assertUnauthenticated(t, err)
}

func TestVerify_WrongAudience(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	tok := ti.Issue("idp|alice", uuid.New(), uuid.Nil, testkit.WithAudience("other-app"))
	_, err := v.Verify(context.Background(), tok)
	assertUnauthenticated(t, err)
}

func TestVerify_UnknownKID(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	// Verifier wired to a different issuer's KeySource → kid won't resolve.
	other := testkit.NewTokenIssuer()
	v := auth.NewVerifier(other.KeySource(), auth.Config{Issuer: "wowapi-test", Audience: "wowapi"})

	tok := ti.Issue("idp|alice", uuid.New(), uuid.Nil)
	_, err := v.Verify(context.Background(), tok)
	assertUnauthenticated(t, err)
}

// TestVerify_AlgConfusionHS256 mints an HS256 token whose HMAC secret is the
// verifier's RSA public key bytes — the classic algorithm-confusion attack — and
// asserts the verifier rejects it rather than validating it as HMAC.
func TestVerify_AlgConfusionHS256(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	// Recover the public key the verifier trusts and use its DER bytes as the
	// forged HMAC secret.
	der, err := x509.MarshalPKIXPublicKey(ti.PublicKey())
	if err != nil {
		t.Fatalf("marshal pub: %v", err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "idp|attacker",
			Issuer:    "wowapi-test",
			Audience:  jwt.ClaimStrings{"wowapi"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		TenantID: uuid.New(),
	}
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	forged.Header["kid"] = "test-key-1"
	signed, err := forged.SignedString(der)
	if err != nil {
		t.Fatalf("sign forged: %v", err)
	}

	_, err = v.Verify(context.Background(), signed)
	assertUnauthenticated(t, err)
}

func TestVerify_TamperedSignature(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	tok := ti.Issue("idp|alice", uuid.New(), uuid.Nil)

	// Flip the FIRST character of the signature segment. The first base64url
	// char always contributes 6 significant bits (the top of byte 0), so the
	// decoded signature is guaranteed to differ — unlike the trailing char,
	// which may only carry discarded padding bits and leave the signature intact.
	parts := strings.Split(tok, ".")
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

	_, err := v.Verify(context.Background(), tampered)
	assertUnauthenticated(t, err)
	assertNoLeak(t, err, tampered)
}

func TestVerify_MissingToken(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	_, err := v.Verify(context.Background(), "")
	assertUnauthenticated(t, err)
}

// TestActor_ZeroTenantRejected proves that a token carrying a zero TenantID is
// rejected before any tenant-bound database work begins.
func TestActor_ZeroTenantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	ps := fakePrincipalStore{userID: uuid.New(), subject: "idp|alice"}

	tok := ti.Issue("idp|alice", uuid.Nil, uuid.Nil)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if err == nil || errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("want KindValidation, got %v", err)
	}
}

// TestActor_GarbageTenantRejected proves that a token carrying a non-existent
// tenant UUID is rejected by the unconditional ActiveTenantAccess check, even
// though the token is validly signed.
func TestActor_GarbageTenantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: uuid.New()} // okTenant != tenantID

	tok := ti.Issue("idp|alice", tenantID, uuid.Nil)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if err == nil || errors.KindOf(err) != errors.KindForbidden {
		t.Fatalf("want KindForbidden, got %v", err)
	}
}

// baseOnlyPrincipalStore implements only the base auth.PrincipalStore
// interface (no AssurancePrincipalStore extension), simulating a legacy or
// misconfigured store wired into a Verifier.
type baseOnlyPrincipalStore struct {
	userID  uuid.UUID
	subject string
	okCap   uuid.UUID
}

func (b baseOnlyPrincipalStore) UserIDBySubject(_ context.Context, subject string) (uuid.UUID, error) {
	if subject != b.subject {
		return uuid.Nil, errors.E(errors.KindUnauthenticated, "unauthenticated", "no such subject")
	}
	return b.userID, nil
}

func (b baseOnlyPrincipalStore) ValidateCapacity(_ context.Context, userID, _ uuid.UUID, capacityID uuid.UUID) error {
	if userID == b.userID && capacityID == b.okCap {
		return nil
	}
	return errors.E(errors.KindForbidden, "permission_denied", "capacity not yours")
}

// TestActor_BaseOnlyStoreFailsClosedOnTenantClaim proves SEC-01's unconditional
// membership verification cannot be silently bypassed by wiring a
// PrincipalStore that lacks the AssurancePrincipalStore extension: a token
// carrying a TenantID must be rejected, not waved through unverified.
func TestActor_BaseOnlyStoreFailsClosedOnTenantClaim(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)
	userID := uuid.New()
	tenantID := uuid.New()
	ps := baseOnlyPrincipalStore{userID: userID, subject: "idp|alice"}

	tok := ti.Issue("idp|alice", tenantID, uuid.Nil)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if err == nil || errors.KindOf(err) != errors.KindForbidden {
		t.Fatalf("want KindForbidden (fail closed on missing assurance store), got %v", err)
	}
}

// T4 — capacity-selection enforcement.

// TestActor_NoCapacitySingleCapacityAllowed proves that a capacity-less token
// is accepted when the actor holds exactly one active capacity: no ambiguous
// choice exists.
func TestActor_NoCapacitySingleCapacityAllowed(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: tenantID, capCount: 1}

	tok := ti.Issue("idp|alice", tenantID, uuid.Nil)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if actor.CapacityID != uuid.Nil {
		t.Fatalf("expected zero capacity, got %v", actor.CapacityID)
	}
}

// TestActor_NoCapacityMultipleCapacitiesRejected proves that a capacity-less
// actor with more than one active capacity is rejected pending an explicit,
// server-side-validated capacity choice (SEC-01 T4).
func TestActor_NoCapacityMultipleCapacitiesRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: tenantID, capCount: 2}

	tok := ti.Issue("idp|alice", tenantID, uuid.Nil)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if err == nil || errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("want KindValidation, got %v", err)
	}
}

// TestActor_ExplicitCapacityValidatedServerSide proves that a token carrying an
// explicit CapacityID is validated against the principal store, not merely
// accepted from the claim.
func TestActor_ExplicitCapacityValidatedServerSide(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	capID := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okTenant: tenantID, okCap: capID}

	tok := ti.Issue("idp|alice", tenantID, capID)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if actor.CapacityID != capID {
		t.Fatalf("capacity: got %v want %v", actor.CapacityID, capID)
	}
}

// T5 — privileged-session resolver.

// TestActor_PrivilegedSessionResolvedFromGrant proves that
// ImpersonatorUserID/BreakGlass are populated from a verified grant row when a
// grant_id claim is present.
func TestActor_PrivilegedSessionResolvedFromGrant(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	grantID := uuid.New()
	imp := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grant:    &auth.ResolvedGrant{ImpersonatorUserID: imp, BreakGlass: true},
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(grantID))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if actor.ImpersonatorUserID != imp {
		t.Fatalf("impersonator: got %v want %v", actor.ImpersonatorUserID, imp)
	}
	if !actor.BreakGlass {
		t.Fatalf("expected break-glass true")
	}
}

// TestActor_DirectImpersonationClaimIgnoredWithoutGrantID proves that a token
// with legacy impersonator/break-glass claims but no grant_id does NOT populate
// those Actor fields (T5 never trusts the claim directly).
func TestActor_DirectImpersonationClaimIgnoredWithoutGrantID(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	imp := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|bob", okTenant: tenantID}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil,
		testkit.WithImpersonator(imp), testkit.WithBreakGlass(true))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	actor, err := v.Actor(context.Background(), claims, ps)
	if err != nil {
		t.Fatalf("Actor: %v", err)
	}
	if actor.ImpersonatorUserID != uuid.Nil {
		t.Fatalf("impersonator must be ignored without grant_id, got %v", actor.ImpersonatorUserID)
	}
	if actor.BreakGlass {
		t.Fatalf("break-glass must be ignored without grant_id")
	}
}

// TestActor_ForgedGrantIDRejected proves that an unknown/forged grant ID is
// rejected with GrantRejectionNotFound.
func TestActor_ForgedGrantIDRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grantErr: errors.E(errors.KindForbidden, string(auth.GrantRejectionNotFound), "grant not found"),
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(uuid.New()))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if !auth.IsGrantRejection(err, auth.GrantRejectionNotFound) {
		t.Fatalf("want GrantRejectionNotFound, got %v", err)
	}
}

// TestActor_ExpiredGrantRejected proves that an expired grant is rejected with
// GrantRejectionExpired.
func TestActor_ExpiredGrantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grantErr: errors.E(errors.KindForbidden, string(auth.GrantRejectionExpired), "grant expired"),
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(uuid.New()))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if !auth.IsGrantRejection(err, auth.GrantRejectionExpired) {
		t.Fatalf("want GrantRejectionExpired, got %v", err)
	}
}

// TestActor_RevokedGrantRejected proves that a revoked grant is rejected with
// GrantRejectionRevoked.
func TestActor_RevokedGrantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grantErr: errors.E(errors.KindForbidden, string(auth.GrantRejectionRevoked), "grant revoked"),
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(uuid.New()))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if !auth.IsGrantRejection(err, auth.GrantRejectionRevoked) {
		t.Fatalf("want GrantRejectionRevoked, got %v", err)
	}
}

// TestActor_WrongTenantGrantRejected proves that a grant from another tenant is
// rejected with GrantRejectionWrongTenant.
func TestActor_WrongTenantGrantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grantErr: errors.E(errors.KindForbidden, string(auth.GrantRejectionWrongTenant), "grant wrong tenant"),
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(uuid.New()))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if !auth.IsGrantRejection(err, auth.GrantRejectionWrongTenant) {
		t.Fatalf("want GrantRejectionWrongTenant, got %v", err)
	}
}

// TestActor_WrongActorGrantRejected proves that a grant authorizing a different
// actor is rejected with GrantRejectionWrongActor.
func TestActor_WrongActorGrantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grantErr: errors.E(errors.KindForbidden, string(auth.GrantRejectionWrongActor), "grant wrong actor"),
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(uuid.New()))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if !auth.IsGrantRejection(err, auth.GrantRejectionWrongActor) {
		t.Fatalf("want GrantRejectionWrongActor, got %v", err)
	}
}

// TestActor_UnauthorizedApproverGrantRejected proves that a grant with an
// unauthorized approver is rejected with GrantRejectionUnauthorizedApprover.
func TestActor_UnauthorizedApproverGrantRejected(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	ps := fakePrincipalStore{
		userID:   userID,
		subject:  "idp|bob",
		okTenant: tenantID,
		grantErr: errors.E(errors.KindForbidden, string(auth.GrantRejectionUnauthorizedApprover), "grant unauthorized approver"),
	}

	tok := ti.Issue("idp|bob", tenantID, uuid.Nil, testkit.WithGrantID(uuid.New()))
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	_, err = v.Actor(context.Background(), claims, ps)
	if !auth.IsGrantRejection(err, auth.GrantRejectionUnauthorizedApprover) {
		t.Fatalf("want GrantRejectionUnauthorizedApprover, got %v", err)
	}
}
