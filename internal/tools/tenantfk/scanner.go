// tenantfk is the DATA-01 tenant-FK catalog scanner and CI gate.
//
// It provides two related operations against a PostgreSQL database:
//
//  1. Enumerate every foreign key on a tenant-scoped (RLS + tenant_id) table
//     and report whether the FK is composite on (tenant_id, ...).
//  2. Check migration SQL files for new non-composite tenant FKs, keyed off
//     the live RLS-tagged tenant-table matrix.
//
// The scanner keys off the same live catalog definition as testkit's RLS
// census: public tables with ROW LEVEL SECURITY enabled that carry a tenant_id
// column.
package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Edge is one foreign key on a tenant-scoped table.
type Edge struct {
	Constraint    string
	ChildTable    string
	ParentTable   string
	ChildColumns  []string
	ParentColumns []string
	Composite     bool // true when child columns include tenant_id
}

// Scanner inspects the PostgreSQL catalog for tenant-scoped FKs.
type Scanner struct {
	DB *pgx.Conn
}

// TenantScopedTables returns the set of public tables that have RLS enabled
// and carry a tenant_id column. This is the authoritative DATA-01 matrix.
func (s *Scanner) TenantScopedTables(ctx context.Context) (map[string]struct{}, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.relname
		  FROM pg_class c
		  JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = 'public' AND c.relrowsecurity
		   AND EXISTS (
		       SELECT 1 FROM information_schema.columns col
		        WHERE col.table_schema = 'public' AND col.table_name = c.relname
		          AND col.column_name = 'tenant_id')`)
	if err != nil {
		return nil, fmt.Errorf("tenantfk: query tenant-scoped tables: %w", err)
	}
	defer rows.Close()

	set := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("tenantfk: scan table name: %w", err)
		}
		set[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tenantfk: iterate tenant-scoped tables: %w", err)
	}
	return set, nil
}

// Enumerate returns every FK on a tenant-scoped table, marking whether it is
// composite on (tenant_id, ...).
func (s *Scanner) Enumerate(ctx context.Context) ([]Edge, error) {
	tables, err := s.TenantScopedTables(ctx)
	if err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, nil
	}

	rows, err := s.DB.Query(ctx, `
		SELECT con.conname,
		       child.relname AS child_table,
		       parent.relname AS parent_table,
		       array_agg(a.attname ORDER BY u.ord) AS child_cols
		  FROM pg_constraint con
		  JOIN pg_class child ON child.oid = con.conrelid
		  JOIN pg_namespace n ON n.oid = child.relnamespace
		  JOIN pg_class parent ON parent.oid = con.confrelid
		  JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS u(attnum, ord) ON true
		  JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = u.attnum
		 WHERE con.contype = 'f'
		   AND n.nspname = 'public'
		   AND child.relname = ANY($1)
		 GROUP BY con.conname, child.relname, parent.relname, con.conkey
		 ORDER BY child.relname, con.conname`, mapKeys(tables))
	if err != nil {
		return nil, fmt.Errorf("tenantfk: query FKs: %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.Constraint, &e.ChildTable, &e.ParentTable, &e.ChildColumns); err != nil {
			return nil, fmt.Errorf("tenantfk: scan FK row: %w", err)
		}
		for _, col := range e.ChildColumns {
			if col == "tenant_id" {
				e.Composite = true
				break
			}
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tenantfk: iterate FKs: %w", err)
	}
	return edges, nil
}

func mapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
