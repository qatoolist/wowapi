package validation_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/i18n"
	"github.com/qatoolist/wowapi/v2/kernel/validation"
)

type reqOnly struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"min=18"`
}

func mrCatalog() *i18n.Catalog {
	cat := i18n.NewRegistry().Catalog()
	cat.Add("mr", i18n.KeyValidationMessage("required"), "हे फील्ड आवश्यक आहे")
	cat.Add("mr", i18n.KeyValidationMessage("min"), "किमान %s असणे आवश्यक")
	return cat
}

func fieldByName(fields []errors.FieldError, name string) (errors.FieldError, bool) {
	for _, f := range fields {
		if f.Field == name {
			return f, true
		}
	}
	return errors.FieldError{}, false
}

func TestStructCtxLocalizesMessageStableFieldCode(t *testing.T) {
	v := validation.New()
	ctx := i18n.WithContext(context.Background(), "mr", mrCatalog())

	err := v.StructCtx(ctx, reqOnly{Name: "", Age: 5})
	if err == nil {
		t.Fatal("expected validation error")
	}
	e, ok := errors.As(err)
	if !ok {
		t.Fatalf("not an *errors.Error: %v", err)
	}
	name, ok := fieldByName(e.Fields, "name")
	if !ok {
		t.Fatalf("no field error for name: %+v", e.Fields)
	}
	if name.Message != "हे फील्ड आवश्यक आहे" {
		t.Errorf("message not localized: %q", name.Message)
	}
	if name.Code != "required" { // stable code
		t.Errorf("code changed: %q", name.Code)
	}
	if name.Field != "name" { // stable field path
		t.Errorf("field changed: %q", name.Field)
	}
	// Parameterised tag: %s filled with the tag param.
	age, _ := fieldByName(e.Fields, "age")
	if age.Message != "किमान 18 असणे आवश्यक" {
		t.Errorf("param message not localized/filled: %q", age.Message)
	}
	if age.Code != "min" {
		t.Errorf("param code changed: %q", age.Code)
	}
}

func TestStructCtxEnglishUnchangedWithoutCatalog(t *testing.T) {
	v := validation.New()
	// No i18n context => byte-identical to the historical Struct output.
	errCtx := v.StructCtx(context.Background(), reqOnly{Name: "", Age: 5})
	errPlain := v.Struct(reqOnly{Name: "", Age: 5})

	ec, _ := errors.As(errCtx)
	ep, _ := errors.As(errPlain)
	nameCtx, _ := fieldByName(ec.Fields, "name")
	namePlain, _ := fieldByName(ep.Fields, "name")
	if nameCtx.Message != namePlain.Message || nameCtx.Message != "this field is required" {
		t.Errorf("StructCtx without catalog diverged: ctx=%q plain=%q", nameCtx.Message, namePlain.Message)
	}
	ageCtx, _ := fieldByName(ec.Fields, "age")
	if ageCtx.Message != "must be at least 18" {
		t.Errorf("english param message changed: %q", ageCtx.Message)
	}
}
