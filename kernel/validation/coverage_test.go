package validation_test

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// tagReq exercises the remaining validator tags whose messages/codes were not
// covered by the base suite: max, len, uuid, gte, lte, and an unmapped tag.
type tagReq struct {
	// max: string longer than the bound violates.
	Bio string `json:"bio" validate:"max=5"`
	// len: exact-length requirement.
	Code string `json:"code" validate:"len=4"`
	// uuid: must be a valid UUID; also maps to invalid_format.
	ID string `json:"id" validate:"uuid"`
	// gte / lte: numeric bounds mapping to min / max codes.
	Age   int `json:"age"   validate:"gte=18"`
	Score int `json:"score" validate:"lte=100"`
	// alphanum is intentionally NOT in tagToCode, so codeForTag falls back to
	// the raw tag name and messageForTag hits its default branch.
	Handle string `json:"handle" validate:"alphanum"`
}

func TestValidator_MaxViolation(t *testing.T) {
	v := validation.New()
	req := tagReq{
		Bio:    "toolongvalue", // > 5 chars
		Code:   "abcd",
		ID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Age:    20,
		Score:  50,
		Handle: "abc123",
	}
	appErr := mustKindValidation(t, v.Struct(req))
	fe := mustField(t, appErr.Fields, "bio")
	if fe.Code != "max" {
		t.Errorf("expected code %q, got %q", "max", fe.Code)
	}
	if fe.Message != "must be at most 5" {
		t.Errorf("unexpected message: %q", fe.Message)
	}
}

func TestValidator_LenViolation(t *testing.T) {
	v := validation.New()
	req := tagReq{
		Bio:    "ok",
		Code:   "ab", // len != 4
		ID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Age:    20,
		Score:  50,
		Handle: "abc123",
	}
	appErr := mustKindValidation(t, v.Struct(req))
	fe := mustField(t, appErr.Fields, "code")
	if fe.Code != "length" {
		t.Errorf("expected code %q, got %q", "length", fe.Code)
	}
	if fe.Message != "must be exactly 4 characters long" {
		t.Errorf("unexpected message: %q", fe.Message)
	}
}

func TestValidator_UUIDViolation(t *testing.T) {
	v := validation.New()
	req := tagReq{
		Bio:    "ok",
		Code:   "abcd",
		ID:     "not-a-uuid",
		Age:    20,
		Score:  50,
		Handle: "abc123",
	}
	appErr := mustKindValidation(t, v.Struct(req))
	fe := mustField(t, appErr.Fields, "id")
	if fe.Code != "invalid_format" {
		t.Errorf("expected code %q, got %q", "invalid_format", fe.Code)
	}
	if fe.Message != "must be a valid UUID" {
		t.Errorf("unexpected message: %q", fe.Message)
	}
}

func TestValidator_GTEViolation(t *testing.T) {
	v := validation.New()
	req := tagReq{
		Bio:    "ok",
		Code:   "abcd",
		ID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Age:    17, // < 18
		Score:  50,
		Handle: "abc123",
	}
	appErr := mustKindValidation(t, v.Struct(req))
	fe := mustField(t, appErr.Fields, "age")
	if fe.Code != "min" {
		t.Errorf("expected code %q, got %q", "min", fe.Code)
	}
	if fe.Message != "must be at least 18" {
		t.Errorf("unexpected message: %q", fe.Message)
	}
}

func TestValidator_LTEViolation(t *testing.T) {
	v := validation.New()
	req := tagReq{
		Bio:    "ok",
		Code:   "abcd",
		ID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Age:    20,
		Score:  101, // > 100
		Handle: "abc123",
	}
	appErr := mustKindValidation(t, v.Struct(req))
	fe := mustField(t, appErr.Fields, "score")
	if fe.Code != "max" {
		t.Errorf("expected code %q, got %q", "max", fe.Code)
	}
	if fe.Message != "must be at most 100" {
		t.Errorf("unexpected message: %q", fe.Message)
	}
}

// TestValidator_UnmappedTag covers codeForTag's fall-through (raw tag name) and
// messageForTag's default branch for a tag absent from tagToCode.
func TestValidator_UnmappedTag(t *testing.T) {
	v := validation.New()
	req := tagReq{
		Bio:    "ok",
		Code:   "abcd",
		ID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Age:    20,
		Score:  50,
		Handle: "not valid!", // spaces/punct fail alphanum
	}
	appErr := mustKindValidation(t, v.Struct(req))
	fe := mustField(t, appErr.Fields, "handle")
	if fe.Code != "alphanum" {
		t.Errorf("expected raw tag as code %q, got %q", "alphanum", fe.Code)
	}
	if fe.Message != `failed "alphanum" validation` {
		t.Errorf("unexpected default message: %q", fe.Message)
	}
}

// noJSONTagReq has a field with no json tag: the TagNameFunc must fall back to
// the Go field name ("Untagged").
type noJSONTagReq struct {
	Untagged string `validate:"required"`
}

func TestValidator_NoJSONTagFallsBackToFieldName(t *testing.T) {
	v := validation.New()
	appErr := mustKindValidation(t, v.Struct(noJSONTagReq{}))
	if findField(appErr.Fields, "Untagged") == nil {
		t.Errorf("expected field name %q from Go field, got: %+v", "Untagged", appErr.Fields)
	}
}

// dashJSONTagReq has json:"-" (excluded from serialisation): the TagNameFunc
// must keep the Go field name ("Hidden") rather than emitting "-".
type dashJSONTagReq struct {
	Hidden string `json:"-" validate:"required"`
}

func TestValidator_DashJSONTagKeepsFieldName(t *testing.T) {
	v := validation.New()
	appErr := mustKindValidation(t, v.Struct(dashJSONTagReq{}))
	if findField(appErr.Fields, "Hidden") == nil {
		t.Errorf("expected field name %q for json:\"-\", got: %+v", "Hidden", appErr.Fields)
	}
	if findField(appErr.Fields, "-") != nil {
		t.Errorf(`field name "-" must not be emitted for json:"-"`)
	}
}

// omitemptyJSONTagReq exercises the ",omitempty" stripping branch in New's
// TagNameFunc: the emitted field name must be "nickname", not "nickname,omitempty".
type omitemptyJSONTagReq struct {
	Nickname string `json:"nickname,omitempty" validate:"required"`
}

func TestValidator_OmitemptyOptionStripped(t *testing.T) {
	v := validation.New()
	appErr := mustKindValidation(t, v.Struct(omitemptyJSONTagReq{}))
	if findField(appErr.Fields, "nickname") == nil {
		t.Errorf("expected stripped field name %q, got: %+v", "nickname", appErr.Fields)
	}
}

// TestValidator_NonStructPointerNilIsInternal ensures a typed-nil pointer (a
// caller misuse) is reported as KindInternal via the InvalidValidationError path.
func TestValidator_NonStructIntIsInternal(t *testing.T) {
	v := validation.New()
	err := v.Struct(42) // not a struct
	if err == nil {
		t.Fatal("expected error for non-struct int, got nil")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Errorf("expected KindInternal, got %v", errors.KindOf(err))
	}
}
