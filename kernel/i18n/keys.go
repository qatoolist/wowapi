package i18n

import "github.com/qatoolist/wowapi/kernel/errors"

// reservedPrefix namespaces every message the framework itself ships. Modules
// may not register keys under it (see Registry.Register), so a product can never
// shadow — accidentally or otherwise — a framework problem title or validation
// message.
const reservedPrefix = "kernel."

// KeyProblemTitle is the well-known catalog key under which the framework's
// English problem-detail title for kind is stored (and translations are keyed).
// It is derived from the kind's STABLE machine code (kind.DefaultCode()), never
// from the English text, so a translation can never drift the key and the
// machine Code on the wire stays byte-stable regardless of locale.
func KeyProblemTitle(kind errors.Kind) string {
	return reservedPrefix + "problem." + kind.DefaultCode()
}

// KeyValidationMessage is the well-known catalog key for the framework's English
// message for a validator tag (e.g. "required", "email", "min"). Keyed by the
// stable tag name, independent of the translated text, so the FieldError.Code
// stays stable.
func KeyValidationMessage(tag string) string {
	return reservedPrefix + "validation." + tag
}

// DefaultLocale is the framework's default locale and ultimate fallback. English
// is always present in the framework catalog.
const DefaultLocale = "en"
