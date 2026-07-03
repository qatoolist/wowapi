package app

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/module"
)

type fakeModule struct {
	name string
	deps []string
}

func (m fakeModule) Name() string                    { return m.name }
func (m fakeModule) DependsOn() []string             { return m.deps }
func (m fakeModule) Register(_ module.Context) error { return nil }

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modules []module.Module
		wantErr []string // substrings that must all appear; empty = must pass
	}{
		{
			name:    "empty app is valid",
			modules: nil,
		},
		{
			name: "valid dependency graph",
			modules: []module.Module{
				fakeModule{name: "requests", deps: []string{"assets"}},
				fakeModule{name: "assets"},
			},
		},
		{
			name: "duplicate name",
			modules: []module.Module{
				fakeModule{name: "requests"},
				fakeModule{name: "requests"},
			},
			wantErr: []string{`"requests": registered more than once`},
		},
		{
			name:    "invalid name",
			modules: []module.Module{fakeModule{name: "Requests-2"}},
			wantErr: []string{"invalid name"},
		},
		{
			name:    "unknown dependency",
			modules: []module.Module{fakeModule{name: "requests", deps: []string{"ghosts"}}},
			wantErr: []string{`depends on unknown module "ghosts"`},
		},
		{
			name:    "self dependency",
			modules: []module.Module{fakeModule{name: "requests", deps: []string{"requests"}}},
			wantErr: []string{"depends on itself"},
		},
		{
			name: "cycle detected",
			modules: []module.Module{
				fakeModule{name: "a", deps: []string{"b"}},
				fakeModule{name: "b", deps: []string{"c"}},
				fakeModule{name: "c", deps: []string{"a"}},
			},
			wantErr: []string{"dependency cycle"},
		},
		{
			name: "multiple problems reported together",
			modules: []module.Module{
				fakeModule{name: "BAD NAME"},
				fakeModule{name: "requests", deps: []string{"ghosts"}},
			},
			wantErr: []string{"invalid name", `unknown module "ghosts"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := New()
			a.Register(tc.modules...)
			err := a.Validate()
			if len(tc.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %v", tc.wantErr)
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error missing %q:\n%v", want, err)
				}
			}
		})
	}
}

func TestOrderedIsDeterministicAndDepsFirst(t *testing.T) {
	a := New()
	a.Register(
		fakeModule{name: "billing", deps: []string{"parties", "catalog"}},
		fakeModule{name: "catalog"},
		fakeModule{name: "parties", deps: []string{"catalog"}},
	)
	got, err := a.Ordered()
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(got))
	for i, m := range got {
		names[i] = m.Name()
	}
	want := "catalog,parties,billing"
	if strings.Join(names, ",") != want {
		t.Errorf("Ordered() = %v, want %s", names, want)
	}

	// Determinism: repeated calls yield the same order.
	again, err := a.Ordered()
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		if got[i].Name() != again[i].Name() {
			t.Fatalf("non-deterministic order: %v vs %v", got, again)
		}
	}
}
