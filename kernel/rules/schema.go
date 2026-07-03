package rules

import (
	"encoding/json"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// validateAgainstSchema checks a rule value against its point's JSON Schema.
// This is a FOCUSED validator, not a full JSON Schema implementation: it
// enforces the top-level "type" (the common case for rule points —
// integer/number/string/boolean/object/array) and, for enums, membership.
// A rule value that a broken schema can't type-check fails loud (KindValidation)
// at write time — defense in depth over the read-path Decode (SEC-40). A future
// full validator can replace this without changing call sites.
func validateAgainstSchema(schema, value json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}
	var s struct {
		Type string            `json:"type"`
		Enum []json.RawMessage `json:"enum"`
	}
	if err := json.Unmarshal(schema, &s); err != nil {
		return kerr.E(kerr.KindInternal, "invalid_rule", "rule point has a malformed value_schema")
	}

	if len(s.Enum) > 0 {
		want := string(compact(value))
		for _, e := range s.Enum {
			if string(compact(e)) == want {
				return nil
			}
		}
		return kerr.E(kerr.KindValidation, "rule_violation", "rule value is not one of the allowed enum values")
	}
	if s.Type == "" {
		return nil
	}

	var v any
	if err := json.Unmarshal(value, &v); err != nil {
		return kerr.E(kerr.KindValidation, "rule_violation", "rule value is not valid JSON")
	}
	if !typeMatches(s.Type, v) {
		return kerr.E(kerr.KindValidation, "rule_violation", "rule value does not match the schema type "+s.Type)
	}
	return nil
}

func typeMatches(t string, v any) bool {
	switch t {
	case "integer":
		f, ok := v.(float64)
		return ok && f == float64(int64(f))
	case "number":
		_, ok := v.(float64)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "null":
		return v == nil
	}
	return true // unknown type keyword → don't block (focused validator)
}

func compact(raw json.RawMessage) json.RawMessage {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return raw
	}
	out, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	return out
}
