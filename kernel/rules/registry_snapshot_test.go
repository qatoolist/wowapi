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
