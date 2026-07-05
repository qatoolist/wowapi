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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
