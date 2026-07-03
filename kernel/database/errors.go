package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors. These map into the kernel error taxonomy when
// kernel/errors lands in Phase 3 (D-0024).
var (
	// ErrNoTenantContext: a tenant-scoped transaction was requested without
	// database.WithTenantID on the context. Fails closed — there is no
	// "default tenant".
	ErrNoTenantContext = errors.New("database: no tenant in context (WithTenant requires database.WithTenantID)")

	// ErrVersionConflict: an optimistic-locking UPDATE matched zero rows —
	// the aggregate changed since it was read (HTTP 409/412 upstream).
	ErrVersionConflict = errors.New("database: version conflict")
)

// ExpectOneRow asserts a versioned UPDATE/DELETE matched exactly one row:
//
//	tag, err := db.Exec(ctx, "UPDATE … WHERE id=$1 AND version=$2", id, v)
//	if err != nil { return err }
//	if err := database.ExpectOneRow(tag, "request"); err != nil { return err }
//
// 0 rows is the optimistic-lock conflict (ErrVersionConflict → 409/412).
// More than 1 row is NOT a conflict — it means the WHERE clause was too broad
// (a missing id predicate, a fan-out UPDATE on a versioned aggregate); that is
// a programming bug and must surface as an internal error (500), never be
// masked as a benign conflict (review finding ARCH-20).
func ExpectOneRow(tag pgconn.CommandTag, entity string) error {
	switch n := tag.RowsAffected(); {
	case n == 1:
		return nil
	case n == 0:
		return fmt.Errorf("%s: %w", entity, ErrVersionConflict)
	default:
		return fmt.Errorf("%s: expected to affect 1 row, affected %d (WHERE clause too broad)", entity, n)
	}
}
