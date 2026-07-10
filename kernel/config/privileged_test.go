package config

import (
	"strings"
	"testing"
)

// Tests for the Privileged allow-list config section (backlog B10). Mirrors
// the "collect ALL errors, joined" convention used by Framework.Validate.

func TestPrivilegedValidate_EmptyIsValid(t *testing.T) {
	var p Privileged
	if err := p.Validate(); err != nil {
		t.Fatalf("empty/absent Privileged config must validate: %v", err)
	}
	f := Defaults()
	if err := f.Validate(); err != nil {
		t.Fatalf("Defaults() with zero-value Privileged must validate: %v", err)
	}
}

func TestPrivilegedValidate_ExplicitEntriesAccepted(t *testing.T) {
	p := Privileged{
		"committee": PrivilegedGrant{
			AllowRelTypes: []string{"core.owner_of", "core.member_of"},
			AllowRuleKeys: []string{"policy.retention.days"},
		},
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("concrete explicit entries must validate: %v", err)
	}
}

// TestPrivilegedValidate_RejectsWildcards proves the SEC requirement: "*" in
// either list is rejected, whatever module it's declared under.
func TestPrivilegedValidate_RejectsWildcards(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    Privileged
	}{
		{"rel type wildcard", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"*"}}}},
		{"rule key wildcard", Privileged{"m": PrivilegedGrant{AllowRuleKeys: []string{"*"}}}},
		{"rel type prefix glob", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"core.*"}}}},
		{"rule key prefix glob", Privileged{"m": PrivilegedGrant{AllowRuleKeys: []string{"policy.*"}}}},
		{"rel type suffix glob", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"*.owner_of"}}}},
		{"rel type question mark glob", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"core.owner_?f"}}}},
		{"rel type bracket glob", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"core.[abc]"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if err == nil {
				t.Fatalf("%s: wildcard/glob entry must fail validation", tc.name)
			}
			if !strings.Contains(err.Error(), "wildcard") && !strings.Contains(err.Error(), "glob") {
				t.Errorf("%s: error should mention wildcard/glob, got: %v", tc.name, err)
			}
		})
	}
}

// TestPrivilegedValidate_RejectsEmptyEntries proves an empty-string entry
// (which would behave like "match everything" in a naive prefix/glob checker)
// is explicitly rejected too, not silently ignored.
func TestPrivilegedValidate_RejectsEmptyEntries(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    Privileged
	}{
		{"empty rel type", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{""}}}},
		{"empty rule key", Privileged{"m": PrivilegedGrant{AllowRuleKeys: []string{""}}}},
		{"whitespace rel type", Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"  "}}}},
		{"empty module name", Privileged{"": PrivilegedGrant{AllowRelTypes: []string{"core.owner_of"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.p.Validate(); err == nil {
				t.Fatalf("%s: must fail validation", tc.name)
			}
		})
	}
}

// TestPrivilegedValidate_CollectsAllErrors matches Framework.Validate's
// "report everything wrong, not just the first" convention.
func TestPrivilegedValidate_CollectsAllErrors(t *testing.T) {
	p := Privileged{
		"a": PrivilegedGrant{AllowRelTypes: []string{"*"}},
		"b": PrivilegedGrant{AllowRuleKeys: []string{"*"}},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected errors")
	}
	msg := err.Error()
	if !strings.Contains(msg, "\"a\"") || !strings.Contains(msg, "\"b\"") {
		t.Errorf("joined error must name both offending modules: %v", err)
	}
}

// TestFrameworkValidate_IncludesPrivileged proves Framework.Validate wires in
// the new section (the ONE new field/line added to the Framework struct).
func TestFrameworkValidate_IncludesPrivileged(t *testing.T) {
	f := Defaults()
	f.Privileged = Privileged{"m": PrivilegedGrant{AllowRelTypes: []string{"*"}}}
	err := f.Validate()
	if err == nil {
		t.Fatal("Framework.Validate must surface Privileged validation errors")
	}
	if !strings.Contains(err.Error(), "privileged") {
		t.Errorf("error should be scoped under privileged.*, got: %v", err)
	}
}
