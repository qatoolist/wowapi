package validation_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/validation"
)

// addressReq is a nested struct used to test dotted field-path generation.
type addressReq struct {
	City string `json:"city" validate:"required"`
}

// createReq exercises the four tag families required by the blueprint and
// includes a nested struct with its own required field.
//
// Note: the Email field deliberately uses json:"email_address" (not
// json:"email") so the test for "json tag drives field name" is unambiguous.
type createReq struct {
	Username string     `json:"username"      validate:"required"`
	Email    string     `json:"email_address" validate:"required,email"`
	Name     string     `json:"name"          validate:"min=3"`
	Role     string     `json:"role"          validate:"oneof=a b c"`
	Address  addressReq `json:"address"`
}

// valid returns a fully-populated createReq that should pass all validation.
func valid() createReq {
	return createReq{
		Username: "alice",
		Email:    "alice@example.com",
		Name:     "Alice",
		Role:     "a",
		Address:  addressReq{City: "New York"},
	}
}

func TestValidator_ValidInput(t *testing.T) {
	v := validation.New()
	if err := v.Struct(valid()); err != nil {
		t.Fatalf("expected nil for valid input, got %v", err)
	}
}

func TestValidator_RequiredViolation(t *testing.T) {
	v := validation.New()
	req := valid()
	req.Username = "" // violates required

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error for missing Username, got nil")
	}

	appErr := mustKindValidation(t, err)
	fe := mustField(t, appErr.Fields, "username")
	if fe.Code != "required" {
		t.Errorf("expected code %q, got %q", "required", fe.Code)
	}
}

func TestValidator_EmailViolation(t *testing.T) {
	v := validation.New()
	req := valid()
	req.Email = "not-an-email" // violates email

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error for bad email, got nil")
	}

	appErr := mustKindValidation(t, err)
	// json tag is "email_address", not "Email"
	fe := mustField(t, appErr.Fields, "email_address")
	if fe.Code != "invalid_format" {
		t.Errorf("expected code %q, got %q", "invalid_format", fe.Code)
	}
}

func TestValidator_MinViolation(t *testing.T) {
	v := validation.New()
	req := valid()
	req.Name = "Al" // 2 chars, violates min=3

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error for short Name, got nil")
	}

	appErr := mustKindValidation(t, err)
	fe := mustField(t, appErr.Fields, "name")
	if fe.Code != "min" {
		t.Errorf("expected code %q, got %q", "min", fe.Code)
	}
}

func TestValidator_OneofViolation(t *testing.T) {
	v := validation.New()
	req := valid()
	req.Role = "superuser" // not in oneof=a b c

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error for invalid Role, got nil")
	}

	appErr := mustKindValidation(t, err)
	fe := mustField(t, appErr.Fields, "role")
	if fe.Code != "invalid_value" {
		t.Errorf("expected code %q, got %q", "invalid_value", fe.Code)
	}
}

func TestValidator_NestedDottedPath(t *testing.T) {
	v := validation.New()
	req := valid()
	req.Address.City = "" // violates nested required

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error for missing Address.City, got nil")
	}

	appErr := mustKindValidation(t, err)
	// Must be json-path dotted: "address.city", not "Address.City".
	fe := mustField(t, appErr.Fields, "address.city")
	if fe.Code != "required" {
		t.Errorf("expected code %q for address.city, got %q", "required", fe.Code)
	}
}

func TestValidator_MultipleViolationsAllReturned(t *testing.T) {
	v := validation.New()
	// Zero-value struct: username missing (required), email missing (required),
	// address.city missing (required).
	req := createReq{}

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected errors for zero-value struct, got nil")
	}

	appErr := mustKindValidation(t, err)
	if len(appErr.Fields) < 2 {
		t.Errorf("expected at least 2 field errors, got %d: %+v", len(appErr.Fields), appErr.Fields)
	}
}

func TestValidator_KindIsValidation(t *testing.T) {
	v := validation.New()
	req := createReq{} // several missing fields

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.KindOf(err) != errors.KindValidation {
		t.Errorf("expected KindValidation, got %v", errors.KindOf(err))
	}
}

func TestValidator_ValueAbsentFromMessages(t *testing.T) {
	v := validation.New()
	// A secret-looking value that would be catastrophic if it leaked into
	// an error body. We make it fail the email validator.
	secret := "super-secret-api-key-xK9$mR2!pL7"
	req := valid()
	req.Email = secret // not a valid email address

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}

	// The secret must not appear anywhere in the error string or field messages.
	if strings.Contains(err.Error(), secret) {
		t.Errorf("secret leaked into err.Error(): %s", err.Error())
	}
	appErr, ok := errors.As(err)
	if ok {
		for _, fe := range appErr.Fields {
			if strings.Contains(fe.Message, secret) {
				t.Errorf("secret leaked into field message for %q: %s", fe.Field, fe.Message)
			}
		}
	}
}

func TestValidator_JSONTagDrivesFieldName(t *testing.T) {
	v := validation.New()
	req := valid()
	req.Email = "bad-email" // Email field has json:"email_address"

	err := v.Struct(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	appErr := mustKindValidation(t, err)
	// Must use the json tag ("email_address"), NOT the Go field name ("Email").
	if findField(appErr.Fields, "email_address") == nil {
		t.Errorf("expected Field=%q (json tag), got fields: %+v", "email_address", appErr.Fields)
	}
	if findField(appErr.Fields, "Email") != nil {
		t.Errorf("Go field name %q must not appear; only json tag names should be used", "Email")
	}
}

func TestValidator_NonStructReturnsInternal(t *testing.T) {
	v := validation.New()
	err := v.Struct("this is not a struct")
	if err == nil {
		t.Fatal("expected error for non-struct input, got nil")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Errorf("expected KindInternal for non-struct, got %v", errors.KindOf(err))
	}
}

// ---- helpers ---------------------------------------------------------------

func mustKindValidation(t *testing.T, err error) *errors.Error {
	t.Helper()
	appErr, ok := errors.As(err)
	if !ok {
		t.Fatalf("expected *errors.Error, got %T: %v", err, err)
	}
	if appErr.Kind != errors.KindValidation {
		t.Fatalf("expected KindValidation, got %v", appErr.Kind)
	}
	return appErr
}

func mustField(t *testing.T, fields []errors.FieldError, path string) *errors.FieldError {
	t.Helper()
	fe := findField(fields, path)
	if fe == nil {
		t.Fatalf("expected FieldError with Field=%q, got: %+v", path, fields)
	}
	return fe
}

func findField(fields []errors.FieldError, path string) *errors.FieldError {
	for i := range fields {
		if fields[i].Field == path {
			return &fields[i]
		}
	}
	return nil
}
