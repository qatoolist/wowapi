package i18n

import "github.com/qatoolist/wowapi/kernel/errors"

// frameworkProblemTitles is the framework's own English problem-detail titles,
// keyed by errors.Kind. These MUST stay byte-for-byte identical to kernel/httpx's
// `titles` map so the localized path and the zero-config path produce exactly the
// same English title for every kind (no behavior change). In particular
// KindIdempotencyExpired is intentionally ABSENT from BOTH maps: httpx renders an
// absent title as "Internal error" via its p.Title=="" fallback, and this catalog
// mirrors that by having no entry (localizeTitle then also resolves to the
// KindInternal title), keeping the two paths in lockstep.
var frameworkProblemTitles = map[errors.Kind]string{
	errors.KindValidation:          "Validation failed",
	errors.KindUnauthenticated:     "Authentication required",
	errors.KindForbidden:           "Permission denied",
	errors.KindTenantIsolation:     "Not found",
	errors.KindNotFound:            "Not found",
	errors.KindConflict:            "Conflict",
	errors.KindVersionConflict:     "Version conflict",
	errors.KindIdempotencyInFlight: "Retry later",
	errors.KindRuleViolation:       "Rule violation",
	errors.KindWorkflowState:       "Invalid transition",
	errors.KindRateLimited:         "Rate limited",
	errors.KindExternal:            "Upstream error",
	errors.KindInternal:            "Internal error",
}

// frameworkValidationMessages is the framework's own English validation-tag
// messages, keyed by validator tag. Parameterised tags (min/max/len/oneof) carry
// a Go %s placeholder filled by kernel/validation with the tag param at render
// time; a locale that omits the placeholder simply renders without the param.
// These mirror kernel/validation's historical messageForTag output verbatim.
var frameworkValidationMessages = map[string]string{
	"required": "this field is required",
	"email":    "must be a valid email address",
	"min":      "must be at least %s",
	"max":      "must be at most %s",
	"len":      "must be exactly %s characters long",
	"oneof":    "must be one of: %s",
	"uuid":     "must be a valid UUID",
	"gte":      "must be at least %s",
	"lte":      "must be at most %s",
}

// installFramework adds the framework's English catalog into cat under the
// reserved kernel.* namespace. Called once by NewRegistry so every catalog the
// framework hands out already localizes problem titles and validation messages
// in English (the default locale and ultimate fallback).
func installFramework(cat *Catalog) {
	for kind, title := range frameworkProblemTitles {
		cat.Add(DefaultLocale, KeyProblemTitle(kind), title)
	}
	for tag, msg := range frameworkValidationMessages {
		cat.Add(DefaultLocale, KeyValidationMessage(tag), msg)
	}
}
