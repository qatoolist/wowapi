package envprovider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/adapters/secrets/envprovider"
	"github.com/qatoolist/wowapi/v2/kernel/secrets"
)

func TestResolve_HappyPath(t *testing.T) {
	const want = "super_secret_value"
	p := envprovider.NewWithLookup(func(key string) (string, bool) {
		if key == "DB_DSN" {
			return want, true
		}
		return "", false
	})
	ref := secrets.Ref{Provider: "env", Path: "DB_DSN"}
	got, err := p.Resolve(context.Background(), ref)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestResolve_WrongProvider(t *testing.T) {
	p := envprovider.New()
	ref := secrets.Ref{Provider: "aws", Path: "DB_DSN"}
	_, err := p.Resolve(context.Background(), ref)
	if err == nil {
		t.Fatal("Resolve() expected error for wrong provider, got nil")
	}
	if !strings.Contains(err.Error(), "aws") {
		t.Errorf("error should mention the received provider %q: %v", "aws", err)
	}
	if !strings.Contains(err.Error(), "env") {
		t.Errorf("error should mention the expected provider %q: %v", "env", err)
	}
}

func TestResolve_MissingVar(t *testing.T) {
	p := envprovider.NewWithLookup(func(string) (string, bool) { return "", false })
	ref := secrets.Ref{Provider: "env", Path: "MISSING_VAR"}
	_, err := p.Resolve(context.Background(), ref)
	if err == nil {
		t.Fatal("Resolve() expected error for missing variable, got nil")
	}
	if !strings.Contains(err.Error(), "MISSING_VAR") {
		t.Errorf("error must name the variable, got: %v", err)
	}
}

func TestResolve_EmptyVar(t *testing.T) {
	p := envprovider.NewWithLookup(func(key string) (string, bool) {
		if key == "EMPTY_VAR" {
			return "", true // present but empty
		}
		return "", false
	})
	ref := secrets.Ref{Provider: "env", Path: "EMPTY_VAR"}
	_, err := p.Resolve(context.Background(), ref)
	if err == nil {
		t.Fatal("Resolve() expected error for empty variable, got nil")
	}
	if !strings.Contains(err.Error(), "EMPTY_VAR") {
		t.Errorf("error must name the variable, got: %v", err)
	}
}

// TestResolve_ErrorMessagesContainNoSecretValue verifies that error paths
// never leak a resolved secret value into the error string (SEC-2).
func TestResolve_ErrorMessagesContainNoSecretValue(t *testing.T) {
	const secretValue = "REDACTED_BY_PROGRAMME_AUDIT"

	// Wrong-provider path: lookup is never reached, but if it were,
	// the value must not appear in the error.
	p := envprovider.NewWithLookup(func(string) (string, bool) {
		return secretValue, true
	})
	ref := secrets.Ref{Provider: "aws", Path: "MY_SECRET"}
	_, err := p.Resolve(context.Background(), ref)
	if err != nil && strings.Contains(err.Error(), secretValue) {
		t.Errorf("wrong-provider error contains secret value: %v", err)
	}
}
