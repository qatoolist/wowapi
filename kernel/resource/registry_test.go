package resource_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/resource"
)

// registry_test.go — QA G5 (framework contract): the resource-type Registry is a
// module registration primitive (validated at boot) with no direct unit test.
// These pin the registration contract — key shape, module-prefix ownership,
// duplicate rejection, and error accumulation — so a regression fails fast.

func TestRegistryAcceptsWellFormedType(t *testing.T) {
	r := resource.NewRegistry()
	r.Register("requests", resource.TypeSpec{Key: "requests.request", Description: "a request"})
	if err := r.Err(); err != nil {
		t.Fatalf("a well-formed type must register cleanly: %v", err)
	}
	if _, ok := r.Specs()["requests.request"]; !ok {
		t.Fatal("registered spec not present in Specs()")
	}
}

func TestRegistryRejectsMalformedKey(t *testing.T) {
	r := resource.NewRegistry()
	r.Register("requests", resource.TypeSpec{Key: "NotAModuleName"}) // no module.name shape
	if r.Err() == nil {
		t.Fatal("a malformed resource-type key must be an error")
	}
}

func TestRegistryRejectsForeignModulePrefix(t *testing.T) {
	r := resource.NewRegistry()
	// module "requests" may not register a "billing.*" type (ownership rule).
	r.Register("requests", resource.TypeSpec{Key: "billing.invoice"})
	if r.Err() == nil {
		t.Fatal("registering another module's type prefix must be an error")
	}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	r := resource.NewRegistry()
	r.Register("requests", resource.TypeSpec{Key: "requests.request"})
	r.Register("requests", resource.TypeSpec{Key: "requests.request"})
	err := r.Err()
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("a duplicate resource type must be a duplicate error, got %v", err)
	}
}

func TestRegistryAccumulatesAllErrors(t *testing.T) {
	r := resource.NewRegistry()
	r.Register("requests", resource.TypeSpec{Key: "BAD"})             // malformed
	r.Register("requests", resource.TypeSpec{Key: "billing.invoice"}) // foreign prefix
	err := r.Err()
	if err == nil {
		t.Fatal("expected accumulated errors")
	}
	// Both problems should be reported, not just the first.
	msg := err.Error()
	if !strings.Contains(msg, "BAD") || !strings.Contains(msg, "billing.invoice") {
		t.Fatalf("Err() must accumulate all registration failures, got: %s", msg)
	}
}

func TestValidTypeKeyAndRefIsZero(t *testing.T) {
	cases := map[string]bool{
		"requests.request": true,
		"a.b":              true,
		"NoDot":            false,
		"requests.":        false,
		".request":         false,
		"UPPER.case":       false,
		"":                 false,
	}
	for key, want := range cases {
		if got := resource.ValidTypeKey(key); got != want {
			t.Errorf("ValidTypeKey(%q) = %v, want %v", key, got, want)
		}
	}
	if !(resource.Ref{}).IsZero() {
		t.Error("zero Ref must be IsZero")
	}
	if (resource.Ref{Type: "requests.request", ID: uuid.New()}).IsZero() {
		t.Error("a populated Ref must not be IsZero")
	}
}
