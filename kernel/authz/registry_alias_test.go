package authz

import "testing"

// Second closure-audit regression (2026-07-17, F-10): Permission carries
// nested mutable data (AllowedSchemes, StepUpPolicy.RequiredAMR). The registry
// must not alias it with callers in either direction — a retained registration
// value or a mutated Get result must never alter validated authz behavior.
func TestPermissionNestedDataIsNotAliased(t *testing.T) {
	r := NewRegistry()
	in := Permission{
		Key:            "widgets.thing.read",
		AllowedSchemes: []CredentialScheme{CredentialUser},
		StepUp:         true,
		StepUpPolicy:   &StepUpPolicy{RequiredAMR: []string{"hwk"}, Challenge: "hwk"},
	}
	r.Register(in)
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	// Mutate the RETAINED registration value's nested data.
	in.AllowedSchemes[0] = CredentialScheme("apikey")
	in.StepUpPolicy.RequiredAMR[0] = "pwd"
	in.StepUpPolicy.Challenge = "weak"

	got, ok := r.Get("widgets.thing.read")
	if !ok {
		t.Fatal("permission missing")
	}
	if got.AllowedSchemes[0] != CredentialUser {
		t.Fatalf("retained registration value mutated AllowedSchemes: %v", got.AllowedSchemes)
	}
	if got.StepUpPolicy.RequiredAMR[0] != "hwk" || got.StepUpPolicy.Challenge != "hwk" {
		t.Fatalf("retained registration value mutated the step-up policy: %+v", got.StepUpPolicy)
	}

	// Mutate the GET result's nested data; a second Get must be unaffected.
	got.AllowedSchemes[0] = CredentialScheme("apikey")
	got.StepUpPolicy.RequiredAMR[0] = "pwd"
	again, _ := r.Get("widgets.thing.read")
	if again.AllowedSchemes[0] != CredentialUser || again.StepUpPolicy.RequiredAMR[0] != "hwk" {
		t.Fatalf("mutating a Get result altered the registry: %v / %+v", again.AllowedSchemes, again.StepUpPolicy)
	}
}
