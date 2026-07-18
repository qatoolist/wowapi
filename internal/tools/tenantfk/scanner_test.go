package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// testDSN returns a PostgreSQL DSN from the environment. Local runs may skip,
// but authoritative gates fail closed when WOWAPI_REQUIRE_DB is set.
func testDSN(t *testing.T) string {
	t.Helper()
	d := os.Getenv("DATABASE_URL")
	if d == "" {
		d = os.Getenv("WOWAPI_TEST_DSN")
	}
	if d == "" {
		if os.Getenv("WOWAPI_REQUIRE_DB") != "" {
			t.Fatal("WOWAPI_REQUIRE_DB is set but neither DATABASE_URL nor WOWAPI_TEST_DSN is available")
		}
		t.Skip("DATABASE_URL or WOWAPI_TEST_DSN not set")
	}
	return d
}

// testConnect opens a connection to the DSN.
func testConnect(t *testing.T, d string) *pgx.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)
	defer cancel()
	conn, err := pgx.Connect(ctx, d)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return conn
}

const defaultConnectTimeout = 10 * time.Second

// createFixtureDB creates a fresh database named after the test and runs the
// 8-edge DATA-01 fixture schema. The caller must close conn and drop the DB.
func createFixtureDB(t *testing.T) (dbname string, conn *pgx.Conn) {
	t.Helper()
	base := testDSN(t)
	admin := testConnect(t, base)
	defer admin.Close(context.Background())

	dbname = fmt.Sprintf("tenantfk_fixture_%s_%d", sanitize(t.Name()), os.Getpid())
	if _, err := admin.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbname)); err != nil {
		t.Fatalf("drop fixture db: %v", err)
	}
	if _, err := admin.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %q", dbname)); err != nil {
		t.Fatalf("create fixture db: %v", err)
	}

	// Replace the database name in the DSN.
	fixtureDSN := strings.Replace(base, "/wowapi?", "/"+dbname+"?", 1)
	if fixtureDSN == base {
		// Fallback: replace last path component before query or fragment.
		fixtureDSN = replaceDBName(base, dbname)
	}
	conn = testConnect(t, fixtureDSN)

	if _, err := conn.Exec(context.Background(), fixtureSchema); err != nil {
		t.Fatalf("create fixture schema: %v", err)
	}
	return dbname, conn
}

func dropFixtureDB(t *testing.T, baseDSN, dbname string) {
	t.Helper()
	admin := testConnect(t, baseDSN)
	defer admin.Close(context.Background())
	if _, err := admin.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %q WITH (FORCE)", dbname)); err != nil {
		t.Logf("drop fixture db: %v", err)
	}
}

func replaceDBName(dsn, dbname string) string {
	// Handle postgres://user:pass@host:port/dbname?...
	idx := strings.LastIndex(dsn, "/")
	if idx == -1 {
		return dsn
	}
	rest := dsn[idx+1:]
	q := strings.IndexAny(rest, "?#")
	if q == -1 {
		return dsn[:idx+1] + dbname
	}
	return dsn[:idx+1] + dbname + rest[q:]
}

func sanitize(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return strings.ToLower(b.String())
}

const fixtureSchema = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Tenants table (global, not RLS).
CREATE TABLE tenants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    display_name text NOT NULL,
    created_by uuid NOT NULL
);

-- Parent tables.
CREATE TABLE parties (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    kind text NOT NULL,
    display_name text NOT NULL,
    created_by uuid NOT NULL
);
CREATE UNIQUE INDEX parties_tenant_id_id_uidx ON parties (tenant_id, id);

CREATE TABLE organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    name text NOT NULL,
    created_by uuid NOT NULL
);
CREATE UNIQUE INDEX organizations_tenant_id_id_uidx ON organizations (tenant_id, id);

CREATE TABLE documents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    document_class text NOT NULL,
    title text NOT NULL,
    created_by uuid NOT NULL
);
CREATE UNIQUE INDEX documents_tenant_id_id_uidx ON documents (tenant_id, id);

CREATE TABLE document_versions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    document_id uuid NOT NULL,
    version_no int NOT NULL,
    storage_key text NOT NULL,
    mime_type text NOT NULL,
    size_bytes bigint NOT NULL,
    checksum_sha256 text NOT NULL,
    uploaded_by uuid NOT NULL,
    UNIQUE (document_id, version_no)
);
CREATE UNIQUE INDEX document_versions_tenant_id_id_uidx ON document_versions (tenant_id, id);

-- Child tables.
CREATE TABLE persons (
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    party_id uuid NOT NULL,
    given_name text NOT NULL,
    PRIMARY KEY (tenant_id, party_id)
);
ALTER TABLE persons ADD CONSTRAINT persons_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id);

CREATE TABLE legal_entities (
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    party_id uuid NOT NULL,
    legal_name text NOT NULL,
    PRIMARY KEY (tenant_id, party_id)
);
ALTER TABLE legal_entities ADD CONSTRAINT legal_entities_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id);

CREATE TABLE party_contacts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    party_id uuid NOT NULL,
    kind text NOT NULL,
    value text NOT NULL,
    created_by uuid NOT NULL
);
ALTER TABLE party_contacts ADD CONSTRAINT party_contacts_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id);

CREATE TABLE acting_capacities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    user_id uuid NOT NULL,
    party_id uuid NOT NULL,
    label text NOT NULL,
    created_by uuid NOT NULL
);
ALTER TABLE acting_capacities ADD CONSTRAINT acting_capacities_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id);

CREATE TABLE resources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    resource_type text NOT NULL,
    org_id uuid NOT NULL,
    label text NOT NULL,
    created_by uuid NOT NULL
);
ALTER TABLE resources ADD CONSTRAINT resources_org_id_tenant_fkey FOREIGN KEY (tenant_id, org_id) REFERENCES organizations (tenant_id, id);

CREATE TABLE document_versions_child (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    document_id uuid NOT NULL,
    version_no int NOT NULL,
    storage_key text NOT NULL,
    mime_type text NOT NULL,
    size_bytes bigint NOT NULL,
    checksum_sha256 text NOT NULL,
    uploaded_by uuid NOT NULL,
    UNIQUE (document_id, version_no)
);
ALTER TABLE document_versions_child ADD CONSTRAINT document_versions_document_id_tenant_fkey FOREIGN KEY (tenant_id, document_id) REFERENCES documents (tenant_id, id);

CREATE TABLE document_access_grants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    document_id uuid NOT NULL,
    grantee_kind text NOT NULL,
    grantee_ref text NOT NULL,
    access text NOT NULL,
    created_by uuid NOT NULL
);
ALTER TABLE document_access_grants ADD CONSTRAINT document_access_grants_document_id_tenant_fkey FOREIGN KEY (tenant_id, document_id) REFERENCES documents (tenant_id, id);

CREATE TABLE attachments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    resource_type text NOT NULL,
    resource_id uuid NOT NULL,
    document_version_id uuid NOT NULL,
    created_by uuid NOT NULL
);
ALTER TABLE attachments ADD CONSTRAINT attachments_document_version_id_tenant_fkey FOREIGN KEY (tenant_id, document_version_id) REFERENCES document_versions (tenant_id, id);

-- Enable RLS on all tenant tables so the scanner keys off the live matrix.
ALTER TABLE parties ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;
ALTER TABLE legal_entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE party_contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE acting_capacities ENABLE ROW LEVEL SECURITY;
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_versions_child ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_access_grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE attachments ENABLE ROW LEVEL SECURITY;
`

func TestParseMigrationSQLFixture(t *testing.T) {
	const src = "fixture.sql"
	body := `
-- +goose Up
ALTER TABLE persons
  ADD CONSTRAINT persons_party_id_tenant_fkey
  FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id) NOT VALID;

CREATE TABLE widgets (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    owner_id uuid REFERENCES owners (id)
);

-- +goose Down
ALTER TABLE persons DROP CONSTRAINT IF EXISTS persons_party_id_tenant_fkey;
`
	fks := ParseMigrationSQL(src, body)
	if len(fks) != 2 {
		t.Fatalf("got %d FKs, want 2: %+v", len(fks), fks)
	}
	if got := fks[0]; got.Constraint != "persons_party_id_tenant_fkey" || !got.IsComposite() {
		t.Errorf("first FK = %+v, want composite persons_party_id_tenant_fkey", got)
	}
	if got := fks[1]; got.Constraint != "" || got.ChildTable != "widgets" || got.IsComposite() {
		t.Errorf("second FK = %+v, want non-composite widgets.owner_id", got)
	}
}

func TestScannerEnumerateFixture(t *testing.T) {
	base := testDSN(t)
	dbname, conn := createFixtureDB(t)
	defer conn.Close(context.Background())
	defer dropFixtureDB(t, base, dbname)

	scan := &Scanner{DB: conn}
	edges, err := scan.Enumerate(context.Background())
	if err != nil {
		t.Fatalf("enumerate: %v", err)
	}

	want := map[string]bool{
		"persons_party_id_tenant_fkey":                   true,
		"legal_entities_party_id_tenant_fkey":            true,
		"party_contacts_party_id_tenant_fkey":            true,
		"acting_capacities_party_id_tenant_fkey":         true,
		"resources_org_id_tenant_fkey":                   true,
		"document_versions_document_id_tenant_fkey":      true,
		"document_access_grants_document_id_tenant_fkey": true,
		"attachments_document_version_id_tenant_fkey":    true,
	}
	got := make(map[string]bool)
	var nonComposite []string
	for _, e := range edges {
		// The fixture seeds tenant_id -> tenants(id) FKs on every tenant table.
		// Those are real but not part of the DATA-01 8-edge matrix.
		if e.ParentTable == "tenants" {
			continue
		}
		got[e.Constraint] = true
		if !e.Composite {
			nonComposite = append(nonComposite, e.Constraint)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d edges, want %d; missing=%v, extra=%v", len(got), len(want), missing(want, got), missing(got, want))
	}
	for name := range want {
		if !got[name] {
			t.Errorf("missing expected edge %q", name)
		}
	}
	if len(nonComposite) > 0 {
		t.Errorf("found non-composite edges: %v", nonComposite)
	}
}

func TestScannerGateNegativeFixture(t *testing.T) {
	base := testDSN(t)
	dbname, conn := createFixtureDB(t)
	defer conn.Close(context.Background())
	defer dropFixtureDB(t, base, dbname)

	scan := &Scanner{DB: conn}
	violations, err := scan.CheckMigrations(context.Background(), []string{"testdata/bad_fk_migration.sql"})
	if err != nil {
		t.Fatalf("check migrations: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("got %d violations, want 1: %+v", len(violations), violations)
	}
	v := violations[0]
	if v.FK.Constraint != "persons_party_id_bad_fkey" {
		t.Errorf("constraint = %q, want persons_party_id_bad_fkey", v.FK.Constraint)
	}
}

func TestMigrationVersionGT(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		since  int
		wantGT bool
	}{
		{"new migration", "migrations/00037_foo.sql", 36, true},
		{"cleanup migration", "migrations/00036_cleanup.sql", 36, false},
		{"pre-cleanup migration", "migrations/00010_documents.sql", 36, false},
		{"fixture without version", "testdata/bad_fk_migration.sql", 36, true},
		{"since zero keeps all", "migrations/00010_documents.sql", 0, true},
		{"clean baseline included at zero", "migrations/00001_baseline.sql", 0, true},
		{"clean baseline excluded only after one", "migrations/00001_baseline.sql", 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := migrationVersionGT(tc.path, tc.since); got != tc.wantGT {
				t.Errorf("migrationVersionGT(%q, %d) = %v, want %v", tc.path, tc.since, got, tc.wantGT)
			}
		})
	}
}

func missing(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if !b[k] {
			out = append(out, k)
		}
	}
	return out
}
