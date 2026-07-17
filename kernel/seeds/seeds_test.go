package seeds_test

import (
	"strings"
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

// TestLoadParsesStepUpAMR pins the richer step-up form (B8): a permission can
// require a SPECIFIC AMR subset and a specific challenge hint, not just the
// deployment default set. step_up_amr/step_up_challenge decode alongside
// step_up: true.
func TestLoadParsesStepUpAMR(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": `
permissions:
  - key: requests.request.approve
    step_up: true
    step_up_amr: [hwk]
    step_up_challenge: hwk
`,
	})
	b, err := seeds.Load(src, "requests")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Permissions) != 1 {
		t.Fatalf("unexpected bundle: %+v", b)
	}
	p := b.Permissions[0]
	if !p.StepUp {
		t.Error("step_up: true not parsed")
	}
	if len(p.StepUpAMR) != 1 || p.StepUpAMR[0] != "hwk" {
		t.Errorf("step_up_amr = %v, want [hwk]", p.StepUpAMR)
	}
	if p.StepUpChallenge != "hwk" {
		t.Errorf("step_up_challenge = %q, want %q", p.StepUpChallenge, "hwk")
	}
}

// TestLoadRejectsStepUpAMRTypo: step_up_amr must strict-decode like every
// other seed field — a typo'd key fails the load, never a silent no-op.
func TestLoadRejectsStepUpAMRTypo(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": "permissions:\n  - key: requests.request.read\n    step_up_amrs: [hwk]\n",
	})
	if _, err := seeds.Load(src, "requests"); err == nil {
		t.Fatal("step_up_amrs typo must fail strict decode")
	}
}

// TestLoadRejectsStepUpAMRWithoutStepUp: step_up_amr/step_up_challenge only
// make sense alongside step_up: true — declaring them without it is very
// likely a seed-author mistake (the AMR subset would silently never gate
// anything), so validate rejects it rather than silently ignoring the fields.
func TestLoadRejectsStepUpAMRWithoutStepUp(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": "permissions:\n  - key: requests.request.read\n    step_up_amr: [hwk]\n",
	})
	if _, err := seeds.Load(src, "requests"); err == nil {
		t.Fatal("step_up_amr without step_up: true must be rejected")
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

// FBL-02: the optional top-level version field is parsed and propagated.
func TestLoadParsesVersion(t *testing.T) {
	b, err := seeds.Load(fsys(map[string]string{
		"seed.yaml": "version: v1.2.3\npermissions:\n  - key: requests.request.read\n",
	}), "requests")
	if err != nil {
		t.Fatal(err)
	}
	if b.Version != "v1.2.3" {
		t.Fatalf("version = %q, want v1.2.3", b.Version)
	}
}

// FBL-02: two files (or two modules) declaring different non-empty versions
// is a load-time error — the version label is not mergeable.
func TestLoadRejectsConflictingVersion(t *testing.T) {
	_, err := seeds.Load(fsys(map[string]string{
		"a.yaml": "version: v1\npermissions:\n  - key: requests.request.read\n",
		"b.yaml": "version: v2\npermissions:\n  - key: requests.request.write\n",
	}), "requests")
	if err == nil {
		t.Fatal("conflicting versions must fail load")
	}
	if !strings.Contains(err.Error(), "version conflict") {
		t.Fatalf("error should mention version conflict: %v", err)
	}
}
