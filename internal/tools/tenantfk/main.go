// Command tenantfk is the DATA-01 tenant-FK catalog scanner and CI gate.
//
// Usage:
//
//	enantfk enumerate --dsn=$DATABASE_URL
//	enantfk gate --dsn=$DATABASE_URL --migrations=migrations/...
//
// The gate exits non-zero when a migration adds a non-composite tenant FK.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "enumerate":
		os.Exit(runEnumerate(os.Args[2:]))
	case "gate":
		os.Exit(runGate(os.Args[2:]))
	case "help", "-h", "--help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "tenantfk: unknown command %q\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `tenantfk — DATA-01 tenant-FK catalog scanner and CI gate

Usage:
  tenantfk enumerate --dsn=DATABASE_URL
  tenantfk gate --dsn=DATABASE_URL --migrations=PATH [--migrations=PATH...]

Commands:
  enumerate   List all FKs on tenant-scoped tables and flag non-composite ones.
  gate        Check migration SQL files for new non-composite tenant FKs.

The scanner keys off the live catalog: public tables with RLS enabled that
contain a tenant_id column. A tenant FK is composite-compliant when its child
columns include tenant_id.
`)
}

func connect(dsn string) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return pgx.Connect(ctx, dsn)
}

func runEnumerate(args []string) int {
	fs := flag.NewFlagSet("enumerate", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk enumerate: %v\n", err)
		return 2
	}
	if *dsn == "" {
		fmt.Fprintln(os.Stderr, "tenantfk enumerate: --dsn or DATABASE_URL required")
		return 2
	}

	conn, err := connect(*dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk enumerate: connect: %v\n", err)
		return 1
	}
	defer func() { _ = conn.Close(context.Background()) }()

	scan := &Scanner{DB: conn}
	edges, err := scan.Enumerate(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk enumerate: %v\n", err)
		return 1
	}

	var bad int
	fmt.Println("CONSTRAINT\tCHILD_TABLE\tPARENT_TABLE\tCOMPOSITE")
	for _, e := range edges {
		ok := "ok"
		if !e.Composite {
			ok = "NON-COMPOSITE"
			bad++
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", e.Constraint, e.ChildTable, e.ParentTable, ok)
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "tenantfk enumerate: found %d non-composite tenant FK(s)\n", bad)
		return 1
	}
	return 0
}

func runGate(args []string) int {
	fs := flag.NewFlagSet("gate", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	since := fs.Int("since", 0, "Only check migrations with a version number greater than this")
	var paths multiFlag
	fs.Var(&paths, "migrations", "migration file or directory (may be repeated)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk gate: %v\n", err)
		return 2
	}
	if *dsn == "" {
		fmt.Fprintln(os.Stderr, "tenantfk gate: --dsn or DATABASE_URL required")
		return 2
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "tenantfk gate: at least one --migrations path required")
		return 2
	}

	conn, err := connect(*dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk gate: connect: %v\n", err)
		return 1
	}
	defer func() { _ = conn.Close(context.Background()) }()

	paths, err = filterMigrationPaths(paths, *since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk gate: %v\n", err)
		return 1
	}
	if len(paths) == 0 {
		fmt.Println("tenantfk gate: no migration files to check")
		return 0
	}

	scan := &Scanner{DB: conn}
	violations, err := scan.CheckMigrations(context.Background(), paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenantfk gate: %v\n", err)
		return 1
	}
	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "tenantfk gate: non-composite tenant FK detected:")
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  %s:%s: %s\n", v.FK.SourceFile, v.FK.Constraint, v.Reason)
		}
		return 1
	}
	fmt.Println("tenantfk gate: no non-composite tenant FKs detected")
	return 0
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

// filterMigrationPaths expands directories into *.sql files and drops files
// whose leading NNNNN version is <= since. Files without a leading version
// number are kept (e.g. test fixtures).
func filterMigrationPaths(paths []string, since int) ([]string, error) {
	var out []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if migrationVersionGT(p, since) {
				out = append(out, p)
			}
			continue
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
				continue
			}
			path := filepath.Join(p, e.Name())
			if migrationVersionGT(path, since) {
				out = append(out, path)
			}
		}
	}
	return out, nil
}

func migrationVersionGT(path string, since int) bool {
	if since == 0 {
		return true
	}
	name := filepath.Base(path)
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return true // no version; keep
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return true // not a version; keep
	}
	return v > since
}
