package config

import (
	"fmt"
	"log/slog"

	"github.com/qatoolist/wowapi/kernel/secrets"
)

// Secret holds a resolved secret value with structural redaction: every
// standard rendering path (fmt verbs, JSON/text marshaling, slog) emits a
// redaction marker, never the value. The raw value is reachable only via
// Reveal, whose call sites are restricted by boundary lint to adapters and
// the app composition root.
//
// The zero Secret is empty and renders as "[redacted]".
type Secret struct {
	ref   string // canonical secretref, "" when constructed directly in tests
	value string
}

// NewSecret builds a resolved secret. ref may be empty (e.g. testkit fakes).
func NewSecret(ref, value string) Secret { return Secret{ref: ref, value: value} }

// Reveal returns the raw secret value. Do not log it. Boundary lint flags
// Reveal calls outside adapters/, app/, and _test.go files.
func (s Secret) Reveal() string { return s.value }

// Ref returns the secret reference this value was resolved from ("" if none).
// Safe to log.
func (s Secret) Ref() string { return s.ref }

// IsZero reports whether the secret is unset (no ref and no value).
func (s Secret) IsZero() bool { return s.ref == "" && s.value == "" }

func (s Secret) redacted() string {
	if s.ref != "" {
		return "[redacted:" + s.ref + "]"
	}
	return "[redacted]"
}

// String implements fmt.Stringer with redaction.
func (s Secret) String() string { return s.redacted() }

// GoString implements fmt.GoStringer so %#v cannot leak the value.
func (s Secret) GoString() string { return "config.Secret(" + s.redacted() + ")" }

// Format implements fmt.Formatter so every fmt verb (%v, %+v, %s, %q, %x, …)
// renders the redaction marker.
func (s Secret) Format(f fmt.State, verb rune) {
	switch verb {
	case 'q':
		fmt.Fprintf(f, "%q", s.redacted())
	default:
		fmt.Fprint(f, s.redacted())
	}
}

// MarshalJSON redacts. Secrets are never serialized as values.
func (s Secret) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%q", s.redacted()), nil
}

// MarshalText redacts (covers yaml/text encoders that honor TextMarshaler).
func (s Secret) MarshalText() ([]byte, error) { return []byte(s.redacted()), nil }

// UnmarshalText accepts only a secret *reference*; the value is resolved
// later, at boot, by the app composition root via a secrets.Provider.
// A raw (non-reference) value is rejected so plaintext secrets cannot enter
// through config files or environment variables.
func (s *Secret) UnmarshalText(b []byte) error {
	ref, err := secrets.ParseRef(string(b))
	if err != nil {
		return fmt.Errorf("config: secret fields accept only secretref:// references: %w", err)
	}
	*s = Secret{ref: ref.String()}
	return nil
}

// LogValue implements slog.LogValuer with redaction.
func (s Secret) LogValue() slog.Value { return slog.StringValue(s.redacted()) }
