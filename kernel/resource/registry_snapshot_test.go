package resource

import "testing"

// Closure-review regression (2026-07-17, F-10): Specs() returns a snapshot —
// mutating it must not alter the registry's backing map.
func TestSpecsReturnsSnapshot(t *testing.T) {
	r := NewRegistry()
	r.Register("owner", TypeSpec{Key: "owner.thing"})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	got := r.Specs()
	delete(got, "owner.thing")
	got["forged.thing"] = TypeSpec{Key: "forged.thing"}
	fresh := r.Specs()
	if _, ok := fresh["owner.thing"]; !ok {
		t.Fatal("mutating the Specs() snapshot deleted a spec from the registry backing map")
	}
	if _, ok := fresh["forged.thing"]; ok {
		t.Fatal("mutating the Specs() snapshot injected a spec into the registry backing map")
	}
}
