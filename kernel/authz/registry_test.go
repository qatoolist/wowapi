package authz_test

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
)

func TestRegistryValidatesKeys(t *testing.T) {
	r := authz.NewRegistry()
	r.Register(authz.Permission{Key: "requests.request.create"})
	r.Register(authz.Permission{Key: "requests.request.read"})
	if err := r.Err(); err != nil {
		t.Fatalf("valid permissions should register: %v", err)
	}
	if !r.Has("requests.request.create") {
		t.Error("registered permission should be present")
	}
	if len(r.Keys()) != 2 {
		t.Errorf("keys = %v", r.Keys())
	}
}

func TestRegistryRejectsBadKeys(t *testing.T) {
	cases := []string{
		"requests.request",            // missing action
		"Requests.Request.Read",       // uppercase
		"requests.request.frobnicate", // action not in closed verb set
		"requests..read",              // empty segment
	}
	for _, key := range cases {
		r := authz.NewRegistry()
		r.Register(authz.Permission{Key: key})
		if r.Err() == nil {
			t.Errorf("bad key %q should fail registration", key)
		}
	}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	r := authz.NewRegistry()
	r.Register(authz.Permission{Key: "requests.request.read"})
	r.Register(authz.Permission{Key: "requests.request.read"})
	if r.Err() == nil {
		t.Fatal("duplicate permission must fail registration")
	}
}
