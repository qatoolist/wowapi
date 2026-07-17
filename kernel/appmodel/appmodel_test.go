package appmodel_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/appmodel"
)

func TestStateTransitions(t *testing.T) {
	c := appmodel.NewCompiler()
	reg := c.GetRegistrar("mod1")
	if reg.Owner() != "mod1" {
		t.Fatal("Registrar identity check failed")
	}

	// Verify initial state is collecting.
	// We don't have a model pointer before compilation, just the compiler state.
	// (StateCollecting is internal to the compiler).

	// 1. Success transition path
	// Define a port, provider, and requirement
	typeInt := reflect.TypeOf(0)
	err := reg.DefinePort("port1", typeInt)
	if err != nil {
		t.Fatalf("DefinePort failed: %v", err)
	}

	err = reg.ProvidePort("port1", 42, typeInt)
	if err != nil {
		t.Fatalf("ProvidePort failed: %v", err)
	}

	reg2 := c.GetRegistrar("mod2")
	err = reg2.RequirePort("port1", typeInt)
	if err != nil {
		t.Fatalf("RequirePort failed: %v", err)
	}

	// Compile should validate and transition to sealed
	sealedModel, err := c.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if sealedModel.State() != appmodel.StateSealed {
		t.Fatalf("Expected state to be sealed, got %s", sealedModel.State())
	}

	// Double compile should fail with invalid transition error
	_, err = c.Compile()
	if !errors.Is(err, appmodel.ErrInvalidTransition) {
		t.Fatalf("Expected ErrInvalidTransition on double compile, got %v", err)
	}
}

func TestValidationRules(t *testing.T) {
	typeInt := reflect.TypeOf(0)

	// Case 1: Providing an undefined port must fail at registration (collection) time
	{
		c := appmodel.NewCompiler()
		reg := c.GetRegistrar("mod1")
		err := reg.ProvidePort("port1", 42, typeInt)
		if err == nil {
			t.Fatal("Expected ProvidePort for undefined port to fail immediately")
		}
	}

	// Case 2: Requiring an undefined port must fail at registration (collection) time
	{
		c := appmodel.NewCompiler()
		reg := c.GetRegistrar("mod2")
		err := reg.RequirePort("port1", typeInt)
		if err == nil {
			t.Fatal("Expected RequirePort for undefined port to fail immediately")
		}
	}

	// Case 3: Required port is defined but has no provider (fails during Compile validation)
	{
		c := appmodel.NewCompiler()
		reg := c.GetRegistrar("mod1")
		err := reg.DefinePort("port1", typeInt)
		if err != nil {
			t.Fatalf("DefinePort failed: %v", err)
		}

		reg2 := c.GetRegistrar("mod2")
		err = reg2.RequirePort("port1", typeInt)
		if err != nil {
			t.Fatalf("RequirePort failed: %v", err)
		}

		_, err = c.Compile()
		if err == nil {
			t.Fatal("Expected compile to fail due to missing provider")
		}
	}
}

func TestPostSealMutation(t *testing.T) {
	c := appmodel.NewCompiler()
	reg := c.GetRegistrar("mod1")
	typeInt := reflect.TypeOf(0)

	err := reg.DefinePort("port1", typeInt)
	if err != nil {
		t.Fatalf("DefinePort failed: %v", err)
	}

	_, err = c.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Helper to recover and verify if a panic or error happened
	verifyPostSeal := func(fn func() error) {
		panicked := false
		var panicVal any
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					panicVal = r
				}
			}()
			err = fn()
		}()

		if appmodel.IsDevBuild {
			if !panicked {
				t.Fatal("Expected post-seal mutation to panic in dev/test build tag, but it returned an error instead")
			}
			errVal, ok := panicVal.(error)
			if !ok || !errors.Is(errVal, appmodel.ErrPostSealMutation) {
				t.Fatalf("Expected ErrPostSealMutation panic, got %v", panicVal)
			}
		} else {
			if panicked {
				t.Fatalf("Expected post-seal mutation to return an error in production build tag, but it panicked instead: %v", panicVal)
			}
			if err == nil || !errors.Is(err, appmodel.ErrPostSealMutation) {
				t.Fatalf("Expected ErrPostSealMutation, got %v", err)
			}
		}
	}

	verifyPostSeal(func() error {
		return reg.DefinePort("port2", typeInt)
	})

	verifyPostSeal(func() error {
		return reg.ProvidePort("port1", 100, typeInt)
	})

	verifyPostSeal(func() error {
		return reg.RequirePort("port1", typeInt)
	})
}

func TestTypeConflictEnforcement(t *testing.T) {
	c := appmodel.NewCompiler()
	reg := c.GetRegistrar("mod1")
	typeInt := reflect.TypeOf(0)
	typeStr := reflect.TypeOf("")

	// Define as int
	err := reg.DefinePort("port1", typeInt)
	if err != nil {
		t.Fatalf("DefinePort failed: %v", err)
	}

	// Defining again with different type (string) must fail with type conflict
	err = reg.DefinePort("port1", typeStr)
	if err == nil {
		t.Fatal("Expected DefinePort with conflicting type to fail")
	}

	// Providing with different type must fail with type conflict
	err = reg.ProvidePort("port1", "not-an-int", typeStr)
	if err == nil {
		t.Fatal("Expected ProvidePort with conflicting type to fail")
	}

	// Requiring with different type must fail with type conflict
	err = reg.RequirePort("port1", typeStr)
	if err == nil {
		t.Fatal("Expected RequirePort with conflicting type to fail")
	}
}

func TestZeroValueRegistrar(t *testing.T) {
	var zeroReg appmodel.Registrar[any]

	err := zeroReg.DefinePort("port1", reflect.TypeOf(0))
	if !errors.Is(err, appmodel.ErrInvalidRegistrar) {
		t.Fatalf("Expected ErrInvalidRegistrar, got %v", err)
	}

	err = zeroReg.ProvidePort("port1", 42, reflect.TypeOf(0))
	if !errors.Is(err, appmodel.ErrInvalidRegistrar) {
		t.Fatalf("Expected ErrInvalidRegistrar, got %v", err)
	}

	err = zeroReg.RequirePort("port1", reflect.TypeOf(0))
	if !errors.Is(err, appmodel.ErrInvalidRegistrar) {
		t.Fatalf("Expected ErrInvalidRegistrar, got %v", err)
	}

	_, err = zeroReg.ResolvePort("port1")
	if !errors.Is(err, appmodel.ErrInvalidRegistrar) {
		t.Fatalf("Expected ErrInvalidRegistrar, got %v", err)
	}
}
