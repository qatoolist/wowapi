package document

import "testing"

// Second closure-audit regression (2026-07-17, F-10): Class.AllowedMIME is
// read on the live confirmation path (allowsMIME); the registry must not
// alias it with callers — a retained registration value or a mutated Get
// result must never change which MIME types a validated class accepts.
func TestClassAllowedMIMEIsNotAliased(t *testing.T) {
	r := NewRegistry()
	in := Class{Key: "widgets.doc", AllowedMIME: []string{"text/plain"}}
	r.Register("widgets", in)
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	// Mutate the RETAINED registration value's slice: widening the allowlist
	// this way must not take effect.
	in.AllowedMIME[0] = "application/x-msdownload"

	got, ok := r.Get("widgets.doc")
	if !ok {
		t.Fatal("class missing")
	}
	if got.AllowedMIME[0] != "text/plain" {
		t.Fatalf("retained registration value mutated AllowedMIME: %v", got.AllowedMIME)
	}
	if !got.allowsMIME("text/plain") || got.allowsMIME("application/x-msdownload") {
		t.Fatalf("validated MIME policy changed through a retained alias: %v", got.AllowedMIME)
	}

	// Mutate the GET result's slice; a second Get must be unaffected.
	got.AllowedMIME[0] = "application/x-msdownload"
	again, _ := r.Get("widgets.doc")
	if again.AllowedMIME[0] != "text/plain" {
		t.Fatalf("mutating a Get result altered the registry: %v", again.AllowedMIME)
	}
}
