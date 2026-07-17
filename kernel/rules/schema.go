package rules

import (
	"bytes"
	"encoding/json"
	"regexp"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// ruleValueSchema is the closed grammar validateAgainstSchema enforces. This
// is NOT JSON Schema — it is a small, framework-specific, formally limited
// subset named RuleValueSchema (ratified Decision 2: a strict limited grammar,
// no JSON-Schema library). Every field here is the WHOLE contract; there is no
// nested "properties"/"items" sub-schema evaluator and never will be.
// See the package-level docs (rules.go) and the rules user-guide section for
// the normative field list.
//
// DisallowUnknownFields makes the decode itself the enforcement mechanism for
// "no keyword outside this list": a schema naming any keyword not declared as
// a field here (e.g. "multipleOf", "additionalProperties", "items") is a
// decode error, not a silently-ignored no-op (B3 defect 2).
type ruleValueSchema struct {
	Type             string            `json:"type"`
	Enum             []json.RawMessage `json:"enum"`
	Minimum          *float64          `json:"minimum"`
	Maximum          *float64          `json:"maximum"`
	ExclusiveMinimum *float64          `json:"exclusiveMinimum"`
	ExclusiveMaximum *float64          `json:"exclusiveMaximum"`
	MinLength        *int              `json:"minLength"`
	MaxLength        *int              `json:"maxLength"`
	Pattern          string            `json:"pattern"`
	MinItems         *int              `json:"minItems"`
	MaxItems         *int              `json:"maxItems"`
	Required         []string          `json:"required"`
}

// decodeRuleValueSchema strict-decodes schema into the closed RuleValueSchema
// grammar: an empty schema is the universal "anything goes" schema (nil, nil);
// any keyword the grammar does not declare is rejected outright (B3 defect 2)
// rather than silently dropped by a lax json.Unmarshal.
func decodeRuleValueSchema(schema json.RawMessage) (*ruleValueSchema, error) {
	if len(schema) == 0 {
		return nil, nil
	}
	var s ruleValueSchema
	dec := json.NewDecoder(bytes.NewReader(schema))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return nil, kerr.E(kerr.KindInternal, "invalid_rule",
			"rule point has a malformed or unrecognized value_schema (RuleValueSchema keywords only): "+err.Error())
	}
	if s.Type != "" && !knownType(s.Type) {
		return nil, kerr.E(kerr.KindInternal, "invalid_rule",
			"rule point value_schema names an unknown type "+s.Type)
	}
	return &s, nil
}

// knownType reports whether t is a type RuleValueSchema recognizes. Anything
// else is a malformed schema (B3 defect 1: the old typeMatches defaulted to
// `true`/accept-anything for an unrecognized type, which is fail-open).
func knownType(t string) bool {
	switch t {
	case "integer", "number", "string", "boolean", "object", "array", "null":
		return true
	}
	return false
}

// validateAgainstSchema checks a rule value against its point's
// RuleValueSchema (see ruleValueSchema/decodeRuleValueSchema) — a FOCUSED,
// closed grammar, not a full JSON Schema implementation. It enforces:
//   - the top-level "type" (integer/number/string/boolean/object/array/null);
//   - "enum" membership;
//   - numeric bounds: "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum";
//   - string bounds: "minLength", "maxLength", "pattern" (RE2, via regexp);
//   - array bounds: "minItems", "maxItems";
//   - object "required" property PRESENCE (key exists in the value) — NOT
//     recursive per-property schema validation (see below).
//
// Explicitly OUT OF SCOPE (narrowing the contract, GAP-007): nested
// "properties" schemas, "additionalProperties", "items" sub-schemas, and any
// other JSON Schema keyword not listed above. A rule point that needs
// per-property typing should declare separate top-level rule points rather
// than one object-shaped point with nested constraints — the framework
// deliberately does not carry a recursive JSON Schema evaluator. "required" is
// supported as a shallow presence check (common case: "this object must set
// these keys") without pulling in recursive validation. Any keyword outside
// this list, or an unrecognized "type", is REJECTED at decode time (fail
// closed, B3) rather than silently ignored or accepted.
//
// A rule value that a broken schema can't type-check fails loud (KindValidation)
// at write time — defense in depth over the read-path Decode (SEC-40). A future
// full validator can replace this without changing call sites.
func validateAgainstSchema(schema, value json.RawMessage) error {
	s, err := decodeRuleValueSchema(schema)
	if err != nil {
		return err
	}
	if s == nil {
		return nil
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

	if f, ok := v.(float64); ok {
		if s.Minimum != nil && f < *s.Minimum {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value is below the schema minimum")
		}
		if s.Maximum != nil && f > *s.Maximum {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value is above the schema maximum")
		}
		if s.ExclusiveMinimum != nil && f <= *s.ExclusiveMinimum {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value must be strictly greater than the schema exclusiveMinimum")
		}
		if s.ExclusiveMaximum != nil && f >= *s.ExclusiveMaximum {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value must be strictly less than the schema exclusiveMaximum")
		}
	}

	if str, ok := v.(string); ok {
		n := len([]rune(str))
		if s.MinLength != nil && n < *s.MinLength {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value is shorter than the schema minLength")
		}
		if s.MaxLength != nil && n > *s.MaxLength {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value is longer than the schema maxLength")
		}
		if s.Pattern != "" {
			re, err := regexp.Compile(s.Pattern)
			if err != nil {
				return kerr.E(kerr.KindInternal, "invalid_rule", "rule point has a malformed value_schema pattern")
			}
			if !re.MatchString(str) {
				return kerr.E(kerr.KindValidation, "rule_violation", "rule value does not match the schema pattern")
			}
		}
	}

	if arr, ok := v.([]any); ok {
		if s.MinItems != nil && len(arr) < *s.MinItems {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value has fewer items than the schema minItems")
		}
		if s.MaxItems != nil && len(arr) > *s.MaxItems {
			return kerr.E(kerr.KindValidation, "rule_violation", "rule value has more items than the schema maxItems")
		}
	}

	if obj, ok := v.(map[string]any); ok && len(s.Required) > 0 {
		for _, key := range s.Required {
			if _, present := obj[key]; !present {
				return kerr.E(kerr.KindValidation, "rule_violation", "rule value is missing required property "+key)
			}
		}
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
	return false // unknown type keyword → fail closed (B3 defect 1); decodeRuleValueSchema
	// already rejects an unrecognized "type" at decode time, so this branch is
	// unreachable from a schema that passed validation — it exists only to keep
	// typeMatches itself safe-by-default if ever called directly.
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
