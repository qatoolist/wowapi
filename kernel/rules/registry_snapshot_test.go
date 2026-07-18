package rules

import "testing"

// Closure-review regression (2026-07-17, F-10): Points() returns a snapshot —
// mutating it must not alter the registry's backing map.
func TestPointsReturnsSnapshot(t *testing.T) {
	r := NewRegistry()
	r.Register("owner", Point{
		Key:         "owner.area.point",
		ValueSchema: []byte(`{"type":"integer"}`),
		Default:     []byte(`0`),
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	got := r.Points()
	delete(got, "owner.area.point")
	fresh := r.Points()
	if _, ok := fresh["owner.area.point"]; !ok {
		t.Fatal("mutating the Points() snapshot deleted a point from the registry backing map")
	}
}

// Second closure-audit regression (2026-07-17, F-10): the outer-map copy is
// not enough — nested aliases (schema/default bytes, scope slices) must not be
// shared with callers in either direction.
func TestPointNestedDataIsNotAliased(t *testing.T) {
	r := NewRegistry()
	in := Point{
		Key:           "owner.area.point",
		ValueSchema:   []byte(`{"type":"integer"}`),
		Default:       []byte(`0`),
		AllowedScopes: []ScopeKind{ScopeTenant},
	}
	r.Register("owner", in)
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	// Mutate the RETAINED registration value's nested bytes/slice.
	in.ValueSchema[len(in.ValueSchema)-2] = 'X'
	in.Default[0] = '9'
	in.AllowedScopes[0] = ScopePlatform

	got, ok := r.Get("owner.area.point")
	if !ok {
		t.Fatal("point missing")
	}
	if string(got.ValueSchema) != `{"type":"integer"}` || string(got.Default) != `0` {
		t.Fatalf("retained registration value mutated the registry's schema/default: %s / %s", got.ValueSchema, got.Default)
	}
	if got.AllowedScopes[0] != ScopeTenant {
		t.Fatalf("retained registration value mutated the registry's scopes: %v", got.AllowedScopes)
	}

	// Mutate a GETTER result's nested data; the registry must be unaffected.
	got.ValueSchema[0] = 'X'
	got.AllowedScopes[0] = ScopeOrg
	pts := r.Points()
	pts["owner.area.point"].ValueSchema[0] = 'Y'
	again, _ := r.Get("owner.area.point")
	if string(again.ValueSchema) != `{"type":"integer"}` || again.AllowedScopes[0] != ScopeTenant {
		t.Fatalf("mutating getter results altered the registry: %s / %v", again.ValueSchema, again.AllowedScopes)
	}
}
