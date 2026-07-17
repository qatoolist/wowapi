// Package validation wraps go-playground/validator/v10 and translates its
// FieldError slice into kernel/errors.FieldError values, producing a
// *errors.Error with Kind=KindValidation. It is the "shape validation" half
// of the two-layer validation strategy described in docs/blueprint/04 §5:
// struct-tag checks live here; domain / cross-field / rule-engine logic lives
// in module domain/validation.go and returns errors.E(KindValidation|KindRuleViolation…).
//
// Import boundary: stdlib + kernel/errors + validator lib. Never module, app,
// adapters, or testkit.
package validation

import (
	"context"
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/i18n"
)

// tagToCode maps common go-playground/validator tag names to the stable
// machine codes used in errors.FieldError.Code. Tags absent from this table
// fall back to the raw tag name, keeping the set extensible without a code
// change to this package.
var tagToCode = map[string]string{
	"required": "required",
	"email":    "invalid_format",
	"min":      "min",
	"max":      "max",
	"len":      "length",
	"oneof":    "invalid_value",
	"uuid":     "invalid_format",
	"gte":      "min",
	"lte":      "max",
}

// Validator wraps a *validator.Validate instance. It is safe for concurrent
// use: go-playground/validator guarantees the *Validate value is read-only
// after construction, making concurrent Struct calls safe without additional
// synchronisation.
type Validator struct {
	v *validator.Validate
}

// New constructs a Validator ready for use. It registers a TagNameFunc so that
// field paths in errors.FieldError values use the json tag name (e.g.
// "email_address") rather than the Go struct field name ("EmailAddress").
// Fields without a json tag fall back to their Go field name.
func New() *Validator {
	v := validator.New()
	// RegisterTagNameFunc is called once during construction, before any
	// concurrent Struct calls, so no locking is required.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "" {
			return fld.Name
		}
		// Strip options such as ",omitempty".
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		// json:"-" means excluded from serialisation; keep the Go name so field
		// paths still make sense in logs / errors.
		if name == "-" {
			return fld.Name
		}
		return name
	})
	return &Validator{v: v}
}

// Struct validates the exported fields of s using struct tags. On success it
// returns nil. On failure it returns a *errors.Error with Kind=KindValidation
// carrying one errors.FieldError per invalid field — all violations are
// collected (not short-circuited) so the caller receives the full picture in
// one round-trip.
//
// If s is not a struct (or pointer-to-struct), Struct returns a KindInternal
// error: passing a non-struct is a programming mistake in the caller, not a
// user-input problem.
func (vl *Validator) Struct(s any) error {
	return vl.StructCtx(context.Background(), s)
}

// StructCtx validates s exactly like Struct but localizes each field message
// against the i18n catalog bound to ctx (kernel/i18n; set by the httpx.Locale
// middleware). The Field path and machine Code stay byte-stable regardless of
// locale — only the human Message is translated. With no catalog in ctx
// (zero-config) the output is byte-identical to Struct's English messages.
//
// httpx.BindAndValidate uses this so API validation errors localize
// automatically; direct Struct callers keep the English behavior.
func (vl *Validator) StructCtx(ctx context.Context, s any) error {
	err := vl.v.Struct(s)
	if err == nil {
		return nil
	}

	// InvalidValidationError means the caller passed a non-struct value; that
	// is a bug in the calling code, not a user error.
	var ive *validator.InvalidValidationError
	if stderrors.As(err, &ive) {
		return errors.E(errors.KindInternal, "internal", "validator misuse", ive)
	}

	var ve validator.ValidationErrors
	if !stderrors.As(err, &ve) {
		// Should never happen: go-playground/validator only returns the two
		// types handled above. Guard anyway so surprises become 500s, not panics.
		return errors.E(errors.KindInternal, "internal", "unexpected validator error", err)
	}

	cat := i18n.CatalogFrom(ctx)
	locale := i18n.LocaleFrom(ctx)
	fields := make([]errors.FieldError, 0, len(ve))
	for _, fe := range ve {
		fields = append(fields, errors.FieldError{
			Field:   fieldPath(fe),
			Code:    codeForTag(fe.Tag()),
			Message: localizedMessage(cat, locale, fe),
		})
	}
	return errors.Validation("validation failed", fields...)
}

// fieldPath converts a validator.FieldError Namespace into a dotted JSON path
// by stripping the root struct type name. go-playground/validator prefixes the
// full namespace with the Go struct type name (e.g. "CreateReq.address.city");
// we drop that first segment to produce "address.city".
//
// The json-tag RegisterTagNameFunc ensures all segments after the root already
// carry json names, so no further transformation is needed.
func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	if idx := strings.Index(ns, "."); idx != -1 {
		return ns[idx+1:]
	}
	// Fallback: namespace had no dot (shouldn't happen in practice).
	return fe.Field()
}

// codeForTag returns the stable machine code for a validator tag. Unknown tags
// return the tag name itself so the mapping stays extensible without changes here.
func codeForTag(tag string) string {
	if code, ok := tagToCode[tag]; ok {
		return code
	}
	return tag
}

// paramTags are the validator tags whose message carries a single %s filled
// with fe.Param(). Kept in sync with messageForTag and the framework catalog's
// parameterised entries so a localized template gets the same substitution.
var paramTags = map[string]bool{
	"min": true, "max": true, "len": true, "oneof": true, "gte": true, "lte": true,
}

// localizedMessage resolves the human message for fe. With no catalog (cat is
// nil / zero-config) it returns the historical English messageForTag output
// verbatim. With a catalog it looks up the framework key for the tag; if the
// catalog served a locale-specific translation, it fills the %s param for
// parameterised tags. A missing translation falls back through the catalog to
// the English framework entry (deterministic i18n fallback), and an unknown tag
// (no framework entry) falls back to messageForTag so novel tags still render.
func localizedMessage(cat *i18n.Catalog, locale string, fe validator.FieldError) string {
	if cat == nil {
		return messageForTag(fe)
	}
	tag := fe.Tag()
	tmpl, _ := cat.Lookup(locale, i18n.KeyValidationMessage(tag))
	// tmpl == the key itself means the framework has no entry for this tag (e.g.
	// a custom tag registered by a product): defer to messageForTag.
	if tmpl == i18n.KeyValidationMessage(tag) {
		return messageForTag(fe)
	}
	if paramTags[tag] {
		param := fe.Param()
		if tag == "oneof" {
			param = strings.ReplaceAll(param, " ", ", ")
		}
		return fmt.Sprintf(tmpl, param)
	}
	return tmpl
}

// messageForTag builds a short human-readable message for a validation
// failure. It deliberately excludes fe.Value() (the actual input) to prevent
// accidental leakage of secrets or PII into error messages (redaction
// discipline from blueprint 04 §5).
func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "uuid":
		return "must be a valid UUID"
	case "gte":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be at most %s", fe.Param())
	default:
		return fmt.Sprintf("failed %q validation", fe.Tag())
	}
}
