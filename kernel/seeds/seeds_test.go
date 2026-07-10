package seeds_test

import (
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/seeds"
)

func fsys(files map[string]string) fstest.MapFS {
	m := fstest.MapFS{}
	for name, body := range files {
		m[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return m
}

func TestLoadMergesAndValidates(t *testing.T) {
	src := fsys(map[string]string{
		"permissions.yaml": `
permissions:
  - key: requests.request.create
    description: create a request
  - key: requests.request.read
    granted_via: requests.assigned_to
resource_types:
  - key: requests.request
    description: a request
relationship_types:
  - key: requests.assigned_to
    subject_kind: capacity
    object_kind: resource
roles:
  - key: requests.org.member
    name: Member
    permissions: [requests.request.create, requests.request.read]
`,
	})
	b, err := seeds.Load(src, "requests")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Permissions) != 2 || len(b.ResourceTypes) != 1 || len(b.RelationshipTypes) != 1 || len(b.Roles) != 1 {
		t.Fatalf("unexpected bundle: %+v", b)
	}
	if b.Permissions[1].GrantedVia != "requests.assigned_to" {
		t.Errorf("granted_via not parsed: %+v", b.Permissions[1])
	}
}

func TestLoadRejectsForeignKeys(t *testing.T) {
	// A module may only seed keys prefixed with its own name.
	src := fsys(map[string]string{
		"p.yaml": "permissions:\n  - key: other.thing.read\n",
	})
	_, err := seeds.Load(src, "requests")
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("foreign key should be rejected: %v", err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	// A typo (unknown key) must fail the strict decode.
	src := fsys(map[string]string{
		"p.yaml": "permissions:\n  - key: requests.request.read\n    sensitivve: true\n",
	})
	if _, err := seeds.Load(src, "requests"); err == nil {
		t.Fatal("unknown field must fail strict decode")
	}
}

// TestLoadParsesStepUp pins step_up: PermissionSeed.StepUp decodes from the
// step_up YAML key, defaults to false when absent, and a typo'd field name
// still fails strict decoding (no silent no-op).
func TestLoadParsesStepUp(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": `
permissions:
  - key: requests.request.read
  - key: requests.request.approve
    step_up: true
`,
	})
	b, err := seeds.Load(src, "requests")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Permissions) != 2 {
		t.Fatalf("unexpected bundle: %+v", b)
	}
	if b.Permissions[0].StepUp {
		t.Errorf("step_up should default to false when absent: %+v", b.Permissions[0])
	}
	if !b.Permissions[1].StepUp {
		t.Errorf("step_up: true not parsed: %+v", b.Permissions[1])
	}
}

func TestLoadRejectsStepUpTypo(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": "permissions:\n  - key: requests.request.read\n    step_upp: true\n",
	})
	if _, err := seeds.Load(src, "requests"); err == nil {
		t.Fatal("step_upp typo must fail strict decode")
	}
}

// SEC-32: a role may not grant a permission its module does not own.
func TestLoadRejectsForeignRoleGrant(t *testing.T) {
	src := fsys(map[string]string{
		"r.yaml": `
roles:
  - key: requests.admin
    name: Admin
    permissions: [billing.invoice.export]
`,
	})
	_, err := seeds.Load(src, "requests")
	if err == nil {
		t.Fatal("a role granting a foreign module's permission must be rejected (SEC-32)")
	}
}

// SEC-34: granted_via must be an owned, declared relationship type.
func TestLoadRejectsForeignGrantedVia(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": `
permissions:
  - key: requests.request.read
    granted_via: billing.owns
`,
	})
	if _, err := seeds.Load(src, "requests"); err == nil {
		t.Fatal("granted_via referencing a foreign relationship type must be rejected (SEC-34)")
	}
}

func TestLoadRejectsDanglingGrantedVia(t *testing.T) {
	// granted_via prefixed correctly but not declared in the bundle.
	src := fsys(map[string]string{
		"p.yaml": `
permissions:
  - key: requests.request.read
    granted_via: requests.nonexistent
`,
	})
	if _, err := seeds.Load(src, "requests"); err == nil {
		t.Fatal("granted_via referencing an undeclared relationship type must be rejected (SEC-34)")
	}
}

func TestLoadAcceptsOwnedGrantedVia(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": `
permissions:
  - key: requests.request.read
    granted_via: requests.assigned_to
relationship_types:
  - key: requests.assigned_to
    subject_kind: capacity
    object_kind: resource
`,
	})
	if _, err := seeds.Load(src, "requests"); err != nil {
		t.Fatalf("an owned+declared granted_via should load: %v", err)
	}
}

func TestLoadEmptyIsEmpty(t *testing.T) {
	b, err := seeds.Load(fsys(map[string]string{}), "requests")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Permissions) != 0 {
		t.Errorf("empty fs should yield empty bundle: %+v", b)
	}
}
