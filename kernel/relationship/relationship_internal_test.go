package relationship

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// fakeRow implements pgx.Row for tests.
type fakeRow struct{ err error }

func (r fakeRow) Scan(dest ...any) error { return r.err }

// fakeDBTX implements database.DBTX for unit tests. resolveSubject only calls
// QueryRow; the other methods are stubbed.
type fakeDBTX struct {
	queryRow func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (f fakeDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f fakeDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (f fakeDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return f.queryRow(ctx, sql, args...)
}

var _ database.DBTX = fakeDBTX{}

// TestUnitResolveSubjectUnsupportedKind proves DATA-07 T2's fail-closed default:
// an unenumerated subject_kind returns KindForbidden, not a silent deny or
// infrastructure error.
func TestUnitResolveSubjectUnsupportedKind(t *testing.T) {
	checker := NewChecker()
	db := fakeDBTX{queryRow: func(ctx context.Context, sql string, args ...any) pgx.Row {
		return fakeRow{err: errors.New("unexpected query")}
	}}
	_, err := checker.resolveSubject(context.Background(), db, authz.Actor{Kind: authz.ActorUser}, "unknown_kind")
	if err == nil {
		t.Fatal("unenumerated subject_kind must fail closed")
	}
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("want KindForbidden, got %v", err)
	}
}

// TestUnitResolveSubjectCapacityNil returns uuid.Nil for a system actor.
func TestUnitResolveSubjectCapacityNil(t *testing.T) {
	checker := NewChecker()
	db := fakeDBTX{queryRow: func(ctx context.Context, sql string, args ...any) pgx.Row {
		return fakeRow{err: errors.New("unexpected query")}
	}}
	id, err := checker.resolveSubject(context.Background(), db, authz.Actor{Kind: authz.ActorSystem}, KindCapacity)
	if err != nil {
		t.Fatalf("capacity resolution for system actor: %v", err)
	}
	if id != uuid.Nil {
		t.Fatalf("want nil capacity id for system actor, got %v", id)
	}
}
