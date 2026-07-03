package auth_test

import (
	"context"
	"crypto/x509"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/auth"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/testkit"
)

// fakePrincipalStore maps a fixed subject → user and validates capacities
// against an allowlist, so Actor mapping is testable without a database.
type fakePrincipalStore struct {
	userID  uuid.UUID
	subject string
	okCap   uuid.UUID // capacity that validates for userID
}

func (f fakePrincipalStore) UserIDBySubject(_ context.Context, subject string) (uuid.UUID, error) {
	if subject != f.subject {
		return uuid.Nil, errors.E(errors.KindUnauthenticated, "unauthenticated", "no such subject")
	}
	return f.userID, nil
}

func (f fakePrincipalStore) ValidateCapacity(_ context.Context, userID, _ uuid.UUID, capacityID uuid.UUID) error {
	if userID == f.userID && capacityID == f.okCap {
		return nil
	}
	return errors.E(errors.KindForbidden, "permission_denied", "capacity not yours")
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
	ps := fakePrincipalStore{userID: userID, subject: "idp|alice", okCap: capID}

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

func TestActor_ImpersonationAndBreakGlassCarry(t *testing.T) {
	ti := testkit.NewTokenIssuer()
	v := newVerifier(ti)

	userID := uuid.New()
	tenantID := uuid.New()
	imp := uuid.New()
	ps := fakePrincipalStore{userID: userID, subject: "idp|bob"}

	// No capacity claim → ValidateCapacity is skipped.
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
	if actor.ImpersonatorUserID != imp || !actor.BreakGlass {
		t.Fatalf("impersonation/break-glass not carried: %+v", actor)
	}
	if actor.CapacityID != uuid.Nil {
		t.Fatalf("expected zero capacity, got %v", actor.CapacityID)
	}
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
	ps := fakePrincipalStore{userID: uuid.New(), subject: "idp|carol", okCap: uuid.New()}

	tok := ti.Issue("idp|carol", uuid.New(), uuid.New()) // capacity not in allowlist
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
