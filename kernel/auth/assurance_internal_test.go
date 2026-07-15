package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/errors"
)

// fakePSForAssurance is a minimal PrincipalStore for these internal tests.
type fakePSForAssurance struct {
	userID   uuid.UUID
	subject  string
	okTenant uuid.UUID
	okCap    uuid.UUID
}

func (f fakePSForAssurance) UserIDBySubject(_ context.Context, subject string) (uuid.UUID, error) {
	if subject != f.subject {
		return uuid.Nil, errors.E(errors.KindUnauthenticated, "unauthenticated", "no such subject")
	}
	return f.userID, nil
}

func (f fakePSForAssurance) ActiveTenantAccess(_ context.Context, userID, tenantID uuid.UUID) error {
	if userID == f.userID && tenantID == f.okTenant {
		return nil
	}
	return errors.E(errors.KindForbidden, "permission_denied", "tenant access not permitted")
}

func (f fakePSForAssurance) ActiveCapacityCount(_ context.Context, userID, tenantID uuid.UUID) (int, error) {
	if userID == f.userID && tenantID == f.okTenant {
		return 1, nil
	}
	return 0, errors.E(errors.KindForbidden, "permission_denied", "tenant access not permitted")
}

func (f fakePSForAssurance) ValidateCapacity(_ context.Context, userID, _ uuid.UUID, capacityID uuid.UUID) error {
	if userID == f.userID && capacityID == f.okCap {
		return nil
	}
	return errors.E(errors.KindForbidden, "permission_denied", "capacity not yours")
}

func (f fakePSForAssurance) ResolveGrant(_ context.Context, _, _, _ uuid.UUID) (*ResolvedGrant, error) {
	return nil, errors.E(errors.KindForbidden, string(GrantRejectionNotFound), "grant not found")
}

// TestActorInternal_AssuranceFieldsPropagate verifies that AuthTime, ACR, AMR,
// and the CredentialUser scheme are propagated from Claims to Actor without
// needing testkit (which currently cannot be imported because the root kernel
// package has a pre-existing build error).
func TestActorInternal_AssuranceFieldsPropagate(t *testing.T) {
	v := NewVerifier(nilStaticKeySource(), Config{Issuer: "wowapi-test", Audience: "wowapi"})
	userID := uuid.New()
	tenantID := uuid.New()
	capID := uuid.New()
	ps := fakePSForAssurance{userID: userID, subject: "idp|alice", okTenant: tenantID, okCap: capID}

	authTime := time.Date(2026, 7, 3, 11, 30, 0, 0, time.UTC)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "idp|alice"},
		TenantID:         tenantID,
		CapacityID:       capID,
		AuthTime:         jwt.NewNumericDate(authTime),
		ACR:              "silver",
		AMR:              []string{"pwd", "mfa"},
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
	if len(actor.AMR) != 2 || actor.AMR[0] != "pwd" || actor.AMR[1] != "mfa" {
		t.Fatalf("Actor.AMR = %v, want [pwd mfa]", actor.AMR)
	}
	if actor.CredentialScheme != authz.CredentialUser {
		t.Fatalf("Actor.CredentialScheme = %q, want user", actor.CredentialScheme)
	}
}

func nilStaticKeySource() KeySource {
	return NewStaticKeySource(map[string]any{})
}
