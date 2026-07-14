package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRejectsUnapprovedSkip(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, `package fixture
import "testing"
func TestUnapproved(t *testing.T) { t.Skip("not approved") }
`)
	manifest := Manifest{Version: 1}

	err := validate(root, manifest)
	if err == nil || !strings.Contains(err.Error(), "unapproved t.Skip") {
		t.Fatalf("validate error = %v, want unapproved t.Skip diagnosis", err)
	}
}

func TestValidateAcceptsApprovedSkipWithOwnerAndRationale(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, `package fixture
import "testing"
func TestOptional(t *testing.T) { t.Skip("requires rootless execution") }
`)
	manifest := Manifest{Version: 1, Skips: []Approval{{
		ID:             "SKIP-001",
		Path:           "fixture_test.go",
		Function:       "TestOptional",
		Method:         "Skip",
		Message:        "requires rootless execution",
		Ordinal:        1,
		Owner:          "release-engineering",
		Classification: "optional",
		Rationale:      "The permission-error branch cannot be exercised when the process is root.",
	}}}

	if err := validate(root, manifest); err != nil {
		t.Fatalf("validate approved skip: %v", err)
	}
}

func TestValidateRejectsIncompleteApproval(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, `package fixture
import "testing"
func TestOptional(t *testing.T) { t.Skip("requires rootless execution") }
`)
	manifest := Manifest{Version: 1, Skips: []Approval{{
		ID:             "SKIP-001",
		Path:           "fixture_test.go",
		Function:       "TestOptional",
		Method:         "Skip",
		Message:        "requires rootless execution",
		Ordinal:        1,
		Classification: "optional",
	}}}

	err := validate(root, manifest)
	if err == nil || !strings.Contains(err.Error(), "owner") || !strings.Contains(err.Error(), "rationale") {
		t.Fatalf("validate error = %v, want owner and rationale diagnostics", err)
	}
}

func writeFixture(t *testing.T, root, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
