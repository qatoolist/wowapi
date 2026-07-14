package port_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/appmodel"
	"github.com/qatoolist/wowapi/kernel/port"
)

type TestService interface {
	DoSomething() string
}

type testServiceImpl struct{}

func (testServiceImpl) DoSomething() string {
	return "done"
}

func TestHappyPathDefineProvideResolve(t *testing.T) {
	c := appmodel.NewCompiler()

	// Mint registrars for mod1 (provider) and mod2 (consumer)
	reg1 := c.GetRegistrar("mod1")
	reg2 := c.GetRegistrar("mod2")

	// Create a typed key
	key := port.NewKey[TestService]("test_service")

	// 1. Define port
	err := port.Define(reg1, key)
	if err != nil {
		t.Fatalf("Define failed: %v", err)
	}

	// 2. Provide port
	var impl TestService = testServiceImpl{}
	err = port.Provide(reg1, key, impl)
	if err != nil {
		t.Fatalf("Provide failed: %v", err)
	}

	// 3. Require port
	err = port.Require(reg2, key)
	if err != nil {
		t.Fatalf("Require failed: %v", err)
	}

	// 4. Compile the model
	_, err = c.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// 5. Resolve port
	resolved, err := port.Resolve(reg2, key)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if resolved.DoSomething() != "done" {
		t.Fatalf("Expected 'done', got %q", resolved.DoSomething())
	}
}

func TestPortTypeConflictRoundTrip(t *testing.T) {
	c := appmodel.NewCompiler()
	reg1 := c.GetRegistrar("mod1")

	keyInt := port.NewKey[int]("my_port")
	keyStr := port.NewKey[string]("my_port")

	// Define as int
	err := port.Define(reg1, keyInt)
	if err != nil {
		t.Fatalf("Define failed: %v", err)
	}

	// Define as string (type conflict)
	err = port.Define(reg1, keyStr)
	if err == nil {
		t.Fatal("Expected Define with conflicting type to fail")
	}

	// Provide as string (type conflict)
	err = port.Provide(reg1, keyStr, "hello")
	if err == nil {
		t.Fatal("Expected Provide with conflicting type to fail")
	}

	// Require as string (type conflict)
	err = port.Require(reg1, keyStr)
	if err == nil {
		t.Fatal("Expected Require with conflicting type to fail")
	}
}

func TestRegistrarForgeCompileFail(t *testing.T) {
	// Programmatically execute 'go build' on our compile-fail fixture.
	// We expect the compilation to fail.
	fixturePath := filepath.Join("testdata", "registrar_forge_compile_fail_fixture")
	cmd := exec.Command("go", "build", "-o", "/dev/null", filepath.Join(fixturePath, "main.go"))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected compile-fail fixture to FAIL compilation, but it compiled successfully!")
	}

	output := stderr.String()
	t.Logf("Compile-fail output:\n%s", output)

	// Assert that we get compile errors for unexported fields and unexported method seal
	if !strings.Contains(output, "cannot refer to unexported field owner") && !strings.Contains(output, "unexported field") {
		t.Error("Expected compile error for unexported field 'owner'")
	}
	if !strings.Contains(output, "seal") && !strings.Contains(output, "cannot refer to unexported method") && !strings.Contains(output, "no field or method") {
		t.Error("Expected compile error for unexported method 'seal'")
	}
}
