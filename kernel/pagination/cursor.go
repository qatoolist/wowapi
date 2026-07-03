package pagination

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/kernel/errors"
)

// maxCursorLen bounds the encoded cursor string an attacker can submit. A keyset
// tuple is a handful of scalars; anything larger is rejected before decoding so
// we never allocate an unbounded payload from untrusted input.
const maxCursorLen = 4096

// Cursor is an opaque keyset position: the physical column values of the last
// row returned, so the next query can resume with a WHERE (cols) > (values)
// comparison. It is encoded as base64url(JSON) and is deliberately not
// human-meaningful — clients round-trip it verbatim via CursorPage.NextCursor.
//
// Supported scalar value types (encode → decode round-trip):
//
//	string          → string
//	bool            → bool
//	int/…/int64     → int64   (integer JSON numbers decode back to int64)
//	uint/…/uint64   → int64
//	float32/float64 → float64 (fractional/exponent JSON numbers)
//	uuid.UUID       → string  (canonical RFC 4122 form)
//	time.Time       → string  (RFC 3339, nanosecond precision, UTC)
//
// uuid.UUID and time.Time are normalised to their string forms on encode, so
// Values reports them as strings; that is sufficient to rebuild a keyset WHERE
// clause where the column type drives the parameter binding.
type Cursor struct {
	values map[string]any
}

// EncodeCursor encodes a keyset tuple (last-row column values) into an opaque
// cursor string. An empty/nil map encodes to "" (the zero cursor). An
// unsupported value type is a server-side programming error (the caller controls
// the keyset columns), so it is returned as a plain error, not a KindValidation.
func EncodeCursor(values map[string]any) (string, error) {
	if len(values) == 0 {
		return "", nil
	}
	norm := make(map[string]any, len(values))
	for k, v := range values {
		nv, err := normalize(v)
		if err != nil {
			return "", fmt.Errorf("pagination: cursor key %q: %w", k, err)
		}
		norm[k] = nv
	}
	b, err := json.Marshal(norm)
	if err != nil {
		return "", fmt.Errorf("pagination: encode cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeCursor parses an opaque cursor produced by EncodeCursor. An empty string
// decodes to the zero Cursor. Any malformed input — bad base64, non-object JSON,
// trailing data, or an oversized payload — yields a KindValidation error and
// never panics (this is attacker-reachable input).
func DecodeCursor(s string) (Cursor, error) {
	if s == "" {
		return Cursor{}, nil
	}
	if len(s) > maxCursorLen {
		return Cursor{}, badCursor()
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, badCursor()
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // preserve int64 vs float64 distinction
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		return Cursor{}, badCursor()
	}
	if dec.More() {
		return Cursor{}, badCursor()
	}
	if m == nil {
		return Cursor{}, badCursor()
	}
	for k, v := range m {
		if n, ok := v.(json.Number); ok {
			m[k] = convertNumber(n)
		}
	}
	return Cursor{values: m}, nil
}

// Values returns a copy of the decoded keyset tuple. Mutating the result does
// not affect the Cursor.
func (c Cursor) Values() map[string]any {
	if len(c.values) == 0 {
		return nil
	}
	out := make(map[string]any, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

// IsZero reports whether the cursor carries no position (start from the
// beginning).
func (c Cursor) IsZero() bool { return len(c.values) == 0 }

func badCursor() error {
	return errors.E(errors.KindValidation, "validation_failed", "invalid pagination cursor")
}

// normalize coerces a supported keyset value into a JSON-safe scalar. See the
// Cursor doc for the supported set.
func normalize(v any) (any, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case bool:
		return x, nil
	case int:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case uint:
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		return int64(x), nil
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	case uuid.UUID:
		return x.String(), nil
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano), nil
	default:
		return nil, fmt.Errorf("unsupported cursor value type %T", v)
	}
}

// convertNumber turns a json.Number back into int64 when it is integral, else
// float64, so round-trips preserve the encoded scalar kind.
func convertNumber(n json.Number) any {
	s := n.String()
	if !strings.ContainsAny(s, ".eE") {
		if i, err := n.Int64(); err == nil {
			return i
		}
	}
	if f, err := n.Float64(); err == nil {
		return f
	}
	return s
}
