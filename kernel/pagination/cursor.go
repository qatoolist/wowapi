package pagination

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
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
	sig    string // sort-spec signature; "" for a legacy flat cursor
}

// Reserved envelope keys for a signed cursor. A signed cursor encodes as the
// two-key object {"__s": <sig>, "__v": {<values>}}; a legacy cursor encodes the
// values map flat. DB column names never take these double-underscore forms, so
// the two encodings are unambiguous on decode.
const (
	keySig  = "__s"
	keyVals = "__v"
)

// Sig returns the sort-spec signature the cursor was minted under, or "" if it
// carries none (a legacy flat cursor). Callers that know the current sort should
// reject a cursor whose Sig does not match — see filtering.KeysetClause.
func (c Cursor) Sig() string { return c.sig }

// EncodeCursor encodes a keyset tuple (last-row column values) into an opaque
// cursor string. An empty/nil map encodes to "" (the zero cursor). An
// unsupported value type is a server-side programming error (the caller controls
// the keyset columns), so it is returned as a plain error, not a KindValidation.
func EncodeCursor(values map[string]any) (string, error) {
	if len(values) == 0 {
		return "", nil
	}
	norm, err := normalizeMap(values)
	if err != nil {
		return "", err
	}
	return encode(norm)
}

// EncodeCursorWithSig encodes a keyset tuple together with the signature of the
// sort it was minted under, so a later request can detect that the sort order
// changed (a direction flip or column reorder that the column-set check alone
// would miss — roadmap R7). An empty sig produces the legacy flat encoding, so
// this is a drop-in for EncodeCursor when no sort binding is desired.
func EncodeCursorWithSig(sig string, values map[string]any) (string, error) {
	if sig == "" {
		return EncodeCursor(values)
	}
	if len(values) == 0 {
		return "", nil
	}
	norm, err := normalizeMap(values)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{keySig: sig, keyVals: norm})
}

func normalizeMap(values map[string]any) (map[string]any, error) {
	norm := make(map[string]any, len(values))
	for k, v := range values {
		nv, err := normalize(v)
		if err != nil {
			return nil, fmt.Errorf("pagination: cursor key %q: %w", k, err)
		}
		norm[k] = nv
	}
	return norm, nil
}

func encode(payload map[string]any) (string, error) {
	b, err := json.Marshal(payload)
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
	// Signed envelope: exactly {"__s": string, "__v": object}. Anything else is a
	// legacy flat cursor whose keys are the value columns directly.
	if len(m) == 2 {
		if sv, ok := m[keySig].(string); ok {
			if vv, ok := m[keyVals].(map[string]any); ok {
				convertNumbers(vv)
				return Cursor{values: vv, sig: sv}, nil
			}
		}
	}
	convertNumbers(m)
	return Cursor{values: m}, nil
}

// convertNumbers rewrites json.Number values in place back to int64/float64 so
// round-trips preserve the encoded scalar kind.
func convertNumbers(m map[string]any) {
	for k, v := range m {
		if n, ok := v.(json.Number); ok {
			m[k] = convertNumber(n)
		}
	}
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
		// uint→int64 can wrap on 64-bit platforms; no earlier validation bounds
		// application-supplied cursor key values, so fail closed (W01-E01-S002
		// gosec G115 triage) instead of silently corrupting cursor ordering.
		if uint64(x) > math.MaxInt64 {
			return nil, fmt.Errorf("cursor value %d overflows int64", x)
		}
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		// Same fail-closed bound as uint above (gosec G115 triage).
		if x > math.MaxInt64 {
			return nil, fmt.Errorf("cursor value %d overflows int64", x)
		}
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
