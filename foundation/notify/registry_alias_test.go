package notify

import "testing"

// Second closure-audit regression (2026-07-17, F-10): TemplateSpec.Vars is the
// variable allowlist template bodies are validated against; the registry must
// not alias it with callers — a retained registration value or a mutated Get
// result must never widen a validated allowlist.
func TestTemplateSpecNestedDataIsNotAliased(t *testing.T) {
	r := NewRegistry()
	in := TemplateSpec{Key: "widgets.area.welcome", Vars: []string{"Name"}, Channels: []string{"email"}}
	r.Register("widgets", in)
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	in.Vars[0] = "SecretToken"
	in.Channels[0] = "webhook"

	got, ok := r.Get("widgets.area.welcome")
	if !ok {
		t.Fatal("spec missing")
	}
	if got.Vars[0] != "Name" || got.Channels[0] != "email" {
		t.Fatalf("retained registration value mutated the spec: %+v", got)
	}
	if !got.allowsVar("Name") || got.allowsVar("SecretToken") {
		t.Fatalf("validated variable allowlist changed through a retained alias: %v", got.Vars)
	}

	got.Vars[0] = "SecretToken"
	again, _ := r.Get("widgets.area.welcome")
	if again.Vars[0] != "Name" {
		t.Fatalf("mutating a Get result altered the registry: %v", again.Vars)
	}
}
