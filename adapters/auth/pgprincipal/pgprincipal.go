// Package pgprincipal implements auth.PrincipalStore over Postgres: it resolves
// the framework user id from an IdP subject on the global identity spine and
// validates an acting capacity within the token's tenant. It is the DB leg of
// the generated api's OIDC/JWT authenticator (kernel/auth), wired in the
// composition root.
//
// Role split (blueprint 03 §1/§2): the global users table is readable only by
// app_platform, so UserIDBySubject runs on the platform TxManager. Tenant-scoped
// acting_capacities is FORCE-RLS and readable only by app_rt within a tenant tx,
// so ValidateCapacity binds the token's tenant and reads read-only under RLS —
// a capacity from another tenant is invisible, never an oracle.
package pgprincipal

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/auth"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Store resolves principals for kernel/auth.
type Store struct {
	platform database.TxManager // app_platform: global identity tables (users)
	runtime  database.TxManager // app_rt: tenant-scoped acting_capacities under RLS
}

// New builds a Store over the platform pool (global identity reads) and the
// runtime pool (tenant-scoped capacity reads under RLS).
func New(platform, runtime database.TxManager) *Store {
	return &Store{platform: platform, runtime: runtime}
}

// UserIDBySubject returns the active framework user id for an IdP subject. An
// unknown or disabled subject is KindUnauthenticated with the same opaque
// message (no oracle distinguishing missing from disabled).
func (s *Store) UserIDBySubject(ctx context.Context, subject string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.platform.Platform(ctx, func(ctx context.Context, db database.DB) error {
		return db.QueryRow(ctx,
			`SELECT id FROM users WHERE idp_subject = $1 AND status = 'active'`, subject,
		).Scan(&id)
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, kerr.E(kerr.KindUnauthenticated, "unauthenticated",
				"unknown subject", kerr.Op("pgprincipal.UserIDBySubject"))
		}
		return uuid.Nil, kerr.Wrapf(err, "pgprincipal.UserIDBySubject", "load user")
	}
	return id, nil
}

// ActiveTenantAccess confirms userID has a live membership row in tenantID.
// A "live" row is status='active' with valid_to IS NULL. Missing, revoked, or
// expired membership is KindForbidden; a DB/connectivity error is wrapped.
func (s *Store) ActiveTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) error {
	var ok bool
	err := s.platform.Platform(ctx, func(ctx context.Context, db database.DB) error {
		return db.QueryRow(ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM user_tenant_access
			    WHERE user_id = $1 AND tenant_id = $2
			      AND status = 'active' AND valid_to IS NULL
			 )`, userID, tenantID,
		).Scan(&ok)
	})
	if err != nil {
		return kerr.Wrapf(err, "pgprincipal.ActiveTenantAccess", "load tenant access")
	}
	if !ok {
		return kerr.E(kerr.KindForbidden, "permission_denied",
			"tenant access not permitted", kerr.Op("pgprincipal.ActiveTenantAccess"))
	}
	return nil
}

// ActiveCapacityCount returns the number of active acting capacities userID
// holds in tenantID. It reads under RLS in the tenant, so capacities from other
// tenants are invisible.
func (s *Store) ActiveCapacityCount(ctx context.Context, userID, tenantID uuid.UUID) (int, error) {
	ctx = database.WithTenantID(ctx, tenantID)
	var count int
	err := s.runtime.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT count(*) FROM acting_capacities
			  WHERE user_id = $1 AND status = 'active' AND valid_to IS NULL`, userID,
		).Scan(&count)
	})
	if err != nil {
		return 0, kerr.Wrapf(err, "pgprincipal.ActiveCapacityCount", "load capacity count")
	}
	return count, nil
}

// ValidateCapacity confirms capacityID is an active acting capacity of userID in
// tenantID. It binds the tenant and reads acting_capacities under RLS, so the row
// is confined to the token's tenant; a missing or foreign capacity is
// KindForbidden.
func (s *Store) ValidateCapacity(ctx context.Context, userID, tenantID, capacityID uuid.UUID) error {
	ctx = database.WithTenantID(ctx, tenantID)
	var ok bool
	err := s.runtime.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM acting_capacities
			    WHERE id = $1 AND user_id = $2 AND status = 'active' AND valid_to IS NULL
			 )`, capacityID, userID,
		).Scan(&ok)
	})
	if err != nil {
		return kerr.Wrapf(err, "pgprincipal.ValidateCapacity", "load capacity")
	}
	if !ok {
		return kerr.E(kerr.KindForbidden, "permission_denied",
			"capacity not permitted", kerr.Op("pgprincipal.ValidateCapacity"))
	}
	return nil
}

// ResolveGrant looks up a privileged-session grant by opaque grantID and
// validates it authorizes userID in tenantID. It returns a verified
// auth.ResolvedGrant or one of the auth.GrantRejection reasons (detectable with
// auth.IsGrantRejection). The lookup runs on the platform manager because
// identity_grant is writable/readable only by app_platform.
func (s *Store) ResolveGrant(ctx context.Context, userID, tenantID, grantID uuid.UUID) (*auth.ResolvedGrant, error) {
	var row struct {
		tenantID           uuid.UUID
		actorID            uuid.UUID
		impersonatedUserID *uuid.UUID
		approverID         *uuid.UUID
		status             string
		expiresAt          *time.Time
		revokedAt          *time.Time
	}

	err := s.platform.Platform(ctx, func(ctx context.Context, db database.DB) error {
		return db.QueryRow(ctx,
			`SELECT tenant_id, actor_id, impersonated_user_id, approver_id, status, expires_at, revoked_at
			   FROM identity_grant
			  WHERE id = $1`,
			grantID,
		).Scan(&row.tenantID, &row.actorID, &row.impersonatedUserID, &row.approverID, &row.status, &row.expiresAt, &row.revokedAt)
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionNotFound),
				"grant not found", kerr.Op("pgprincipal.ResolveGrant"))
		}
		return nil, kerr.Wrapf(err, "pgprincipal.ResolveGrant", "load grant")
	}
	if row.tenantID != tenantID {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionWrongTenant),
			"grant tenant mismatch", kerr.Op("pgprincipal.ResolveGrant"))
	}

	if row.status == "revoked" || row.revokedAt != nil {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionRevoked),
			"grant revoked", kerr.Op("pgprincipal.ResolveGrant"))
	}
	if row.status == "expired" || (row.expiresAt != nil && row.expiresAt.Before(time.Now())) {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionExpired),
			"grant expired", kerr.Op("pgprincipal.ResolveGrant"))
	}
	if row.status != "active" {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionNotFound),
			"grant not active", kerr.Op("pgprincipal.ResolveGrant"))
	}

	grant := &auth.ResolvedGrant{}
	if row.impersonatedUserID != nil {
		// Impersonation: the authenticated user is the person being impersonated;
		// the grant's actor_id is the support operator performing the action.
		if *row.impersonatedUserID != userID {
			return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionWrongActor),
				"grant actor mismatch", kerr.Op("pgprincipal.ResolveGrant"))
		}
		if row.actorID == userID {
			return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionWrongActor),
				"self-impersonation not permitted", kerr.Op("pgprincipal.ResolveGrant"))
		}
		grant.ImpersonatorUserID = row.actorID
	} else {
		// Break-glass: the authenticated user is the actor named in the grant.
		if row.actorID != userID {
			return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionWrongActor),
				"grant actor mismatch", kerr.Op("pgprincipal.ResolveGrant"))
		}
		grant.BreakGlass = true
	}

	// Authority model (interim, pending DEC-Q1): a privileged grant must name a
	// distinct approver who holds active membership in the same tenant.
	if row.approverID == nil {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionUnauthorizedApprover),
			"grant approver required", kerr.Op("pgprincipal.ResolveGrant"))
	}
	if *row.approverID == row.actorID || *row.approverID == userID {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionUnauthorizedApprover),
			"grant approver must be distinct", kerr.Op("pgprincipal.ResolveGrant"))
	}
	if err := s.ActiveTenantAccess(ctx, *row.approverID, tenantID); err != nil {
		return nil, kerr.E(kerr.KindForbidden, string(auth.GrantRejectionUnauthorizedApprover),
			"grant approver not authorized", kerr.Op("pgprincipal.ResolveGrant"))
	}

	return grant, nil
}
