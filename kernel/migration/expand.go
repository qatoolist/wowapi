package migration

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ExpandPhase carries the schema-additive, non-blocking DDL operations that
// make up the DATA-09 expand phase. Each operation produces SQL text that the
// caller executes with ExecDDL (online) or directly (maintenance).
type ExpandPhase struct{}

// AddColumnNullableDefault returns DDL to add a nullable column with a safe
// default. Old readers that do not reference the column are unaffected.
func (ExpandPhase) AddColumnNullableDefault(table, column, typ, defaultExpr string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s DEFAULT %s",
		quoteIdent(table), quoteIdent(column), typ, defaultExpr)
}

// AddNotValidCheck returns DDL to add a CHECK constraint without validating
// existing rows. The constraint is enforced for new/changed rows immediately.
func (ExpandPhase) AddNotValidCheck(table, name, check string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s CHECK (%s) NOT VALID",
		quoteIdent(table), quoteIdent(name), check)
}

// CreateIndexConcurrently returns DDL to build an index without blocking
// writes. It must be executed outside a transaction.
func (ExpandPhase) CreateIndexConcurrently(table, indexName, columns string) string {
	return fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s (%s)",
		quoteIdent(indexName), quoteIdent(table), columns)
}

// CreateCompatibilityView returns DDL to create a view that exposes the old
// column layout after an expand rename/add so N-1 binaries continue to read.
func (ExpandPhase) CreateCompatibilityView(view, table, selectExpr string) string {
	return fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT %s FROM %s",
		quoteIdent(view), selectExpr, quoteIdent(table))
}

// ExecExpandDDL runs an expand-phase DDL statement with the online lock budget
// and bounded retry ceiling. It is a thin wrapper that also resets timeouts.
func ExecExpandDDL(ctx context.Context, conn *pgx.Conn, stmt string) error {
	return ExecDDL(ctx, conn, stmt, DefaultLockBudget, DefaultRetryCeiling)
}

// quoteIdent wraps a single SQL identifier in double quotes. It panics on
// invalid identifiers because migration tooling identifiers are programmer-
// authored, not user input.
func quoteIdent(id string) string {
	if id == "" {
		panic("migration: empty identifier")
	}
	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			panic(fmt.Sprintf("migration: invalid identifier %q", id))
		}
	}
	return fmt.Sprintf("\"%s\"", id)
}
