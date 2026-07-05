package policy_test

import (
	"encoding/json"
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/policy"
)

// rawCond builds a condition with an exact raw JSON operand, letting a test feed
// malformed values that json.Marshal would never produce.
func rawCond(attr, op, rawValue string) authz.Condition {
	return authz.Condition{Attribute: attr, Op: op, Value: json.RawMessage(rawValue)}
}

// evalOne is a helper: match a single condition and return (ok, err).
func evalOne(c authz.Condition, attrs map[string]any) (bool, error) {
	return policy.New().Matches([]authz.Condition{c}, attrs)
}

// TestNumericScalarEqualCrossesTypes drives scalarEqual's numeric branch and
// toFloat's int / int64 / json.Number arms: an int attribute must compare equal
// to a JSON number operand (which decodes as float64).
func TestNumericScalarEqualCrossesTypes(t *testing.T) {
	cases := []struct {
		name  string
		attr  any
		value any // marshaled into the operand
		want  bool
	}{
		{"int-eq-hit", int(5), 5, true},
		{"int-eq-miss", int(5), 6, false},
		{"int64-eq-hit", int64(7), 7, true},
		{"int64-eq-miss", int64(7), 8, false},
		{"jsonnumber-eq-hit", json.Number("3"), 3, true},
		{"jsonnumber-eq-miss", json.Number("3"), 4, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := cond("n", "eq", tc.value)
			got, err := evalOne(c, map[string]any{"n": tc.attr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("eq %v vs %v: got %v, want %v", tc.attr, tc.value, got, tc.want)
			}
		})
	}
}

// TestBadJSONNumberIsNotNumeric exercises toFloat's json.Number error path: a
// non-numeric json.Number is not a float, so equality falls back to string form.
func TestBadJSONNumberIsNotNumeric(t *testing.T) {
	// attr is a json.Number that cannot parse as a float; operand "5" is numeric.
	// Numeric compare is skipped (attr not numeric) and string compare fails.
	got, err := evalOne(cond("n", "eq", 5), map[string]any{"n": json.Number("notanumber")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatal("non-numeric json.Number must not equal numeric operand")
	}
	// And it equals its own string form.
	got, err = evalOne(cond("n", "eq", "notanumber"), map[string]any{"n": json.Number("notanumber")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("non-numeric json.Number must equal its string form")
	}
}

// TestMalformedScalarValueErrors covers decodeScalar's error arm and the
// per-operator error propagation in matchOne for scalar-decoding operators.
func TestMalformedScalarValueErrors(t *testing.T) {
	attrs := map[string]any{"x": "y", "tags": []any{"a"}}
	for _, op := range []string{"eq", "neq", "contains"} {
		t.Run(op, func(t *testing.T) {
			// "{" is a truncated JSON object → json.Unmarshal into any fails.
			attr := "x"
			if op == "contains" {
				attr = "tags"
			}
			ok, err := evalOne(rawCond(attr, op, "{"), attrs)
			if err == nil {
				t.Fatalf("malformed scalar value must error, got ok=%v", ok)
			}
			if kerr.KindOf(err) != kerr.KindInternal {
				t.Fatalf("want KindInternal, got %v", kerr.KindOf(err))
			}
		})
	}
}

// TestMalformedListValueErrors covers decodeList's error arm and matchOne's
// list-decoding operators (in / not_in) when the operand is not a JSON list.
func TestMalformedListValueErrors(t *testing.T) {
	attrs := map[string]any{"x": "y"}
	for _, op := range []string{"in", "not_in"} {
		t.Run(op, func(t *testing.T) {
			// A JSON string is not a list → decodeList fails.
			ok, err := evalOne(rawCond("x", op, `"notalist"`), attrs)
			if err == nil {
				t.Fatalf("malformed list value must error, got ok=%v", ok)
			}
			if kerr.KindOf(err) != kerr.KindInternal {
				t.Fatalf("want KindInternal, got %v", kerr.KindOf(err))
			}
		})
	}
}

// TestContainsOnNonListAttribute drives listContains' type-assertion failure:
// a scalar attribute can never "contain" anything.
func TestContainsOnNonListAttribute(t *testing.T) {
	got, err := evalOne(cond("status", "contains", "x"), map[string]any{"status": "locked"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatal("contains on a non-list attribute must be false")
	}
}

// TestContainsAbsentAttribute: absent attribute (nil, not a list) → false.
func TestContainsAbsentAttribute(t *testing.T) {
	got, err := evalOne(cond("missing", "contains", "x"), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatal("contains on an absent attribute must be false")
	}
}

// TestNumericCompareRequiresNumericOperands covers compareNumeric's error arm
// for both a non-numeric attribute and a non-numeric operand.
func TestNumericCompareRequiresNumericOperands(t *testing.T) {
	t.Run("non-numeric-attr", func(t *testing.T) {
		ok, err := evalOne(cond("status", "gte", 5), map[string]any{"status": "locked"})
		if err == nil {
			t.Fatalf("gte on non-numeric attribute must error, got ok=%v", ok)
		}
		if kerr.KindOf(err) != kerr.KindInternal {
			t.Fatalf("want KindInternal, got %v", kerr.KindOf(err))
		}
	})
	t.Run("non-numeric-operand", func(t *testing.T) {
		ok, err := evalOne(cond("count", "lte", "abc"), map[string]any{"count": float64(5)})
		if err == nil {
			t.Fatalf("lte with non-numeric operand must error, got ok=%v", ok)
		}
		if kerr.KindOf(err) != kerr.KindInternal {
			t.Fatalf("want KindInternal, got %v", kerr.KindOf(err))
		}
	})
	t.Run("malformed-operand", func(t *testing.T) {
		ok, err := evalOne(rawCond("count", "gte", "{"), map[string]any{"count": float64(5)})
		if err == nil {
			t.Fatalf("gte with malformed operand must error, got ok=%v", ok)
		}
	})
}

// TestNumericCompareAbsentAttribute: gte/lte on an absent attribute → false, no
// error (compareNumeric's !present arm).
func TestNumericCompareAbsentAttribute(t *testing.T) {
	for _, op := range []string{"gte", "lte"} {
		got, err := evalOne(cond("missing", op, 1), map[string]any{})
		if err != nil {
			t.Fatalf("%s absent: unexpected error: %v", op, err)
		}
		if got {
			t.Fatalf("%s on absent attribute must be false", op)
		}
	}
}

// TestWithinWindowBranches exercises every non-happy arm of withinWindow.
func TestWithinWindowBranches(t *testing.T) {
	validBounds := []any{"2026-07-03T00:00:00Z", "2026-07-03T23:59:59Z"}

	t.Run("absent-attr", func(t *testing.T) {
		got, err := evalOne(cond("t", "within", validBounds), map[string]any{})
		if err != nil || got {
			t.Fatalf("absent attr: ok=%v err=%v", got, err)
		}
	})

	t.Run("wrong-arity", func(t *testing.T) {
		// Three-element bounds → len != 2 → error.
		ok, err := evalOne(cond("t", "within", []any{"a", "b", "c"}),
			map[string]any{"t": "2026-07-03T12:00:00Z"})
		if err == nil {
			t.Fatalf("within with 3 bounds must error, got ok=%v", ok)
		}
		if kerr.KindOf(err) != kerr.KindInternal {
			t.Fatalf("want KindInternal, got %v", kerr.KindOf(err))
		}
	})

	t.Run("non-list-bounds", func(t *testing.T) {
		ok, err := evalOne(rawCond("t", "within", `"notalist"`),
			map[string]any{"t": "2026-07-03T12:00:00Z"})
		if err == nil {
			t.Fatalf("within with non-list bounds must error, got ok=%v", ok)
		}
	})

	t.Run("unparsable-actual-time", func(t *testing.T) {
		// Actual value is not RFC3339 → false, no error.
		got, err := evalOne(cond("t", "within", validBounds),
			map[string]any{"t": "not-a-timestamp"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Fatal("unparsable actual time must be false")
		}
	})

	t.Run("unparsable-bounds", func(t *testing.T) {
		ok, err := evalOne(cond("t", "within", []any{"bad-from", "bad-to"}),
			map[string]any{"t": "2026-07-03T12:00:00Z"})
		if err == nil {
			t.Fatalf("within with non-RFC3339 bounds must error, got ok=%v", ok)
		}
		if kerr.KindOf(err) != kerr.KindInternal {
			t.Fatalf("want KindInternal, got %v", kerr.KindOf(err))
		}
	})

	t.Run("boundary-inclusive", func(t *testing.T) {
		// Exactly on the lower bound → inclusive → true.
		got, err := evalOne(cond("t", "within", validBounds),
			map[string]any{"t": "2026-07-03T00:00:00Z"})
		if err != nil || !got {
			t.Fatalf("lower boundary must be inclusive: ok=%v err=%v", got, err)
		}
	})
}

// TestNegativeOpsOnAbsentAttribute: neq / not_in are satisfied when the
// attribute is absent (the !present arms).
func TestNegativeOpsOnAbsentAttribute(t *testing.T) {
	t.Run("neq-absent", func(t *testing.T) {
		got, err := evalOne(cond("missing", "neq", "x"), map[string]any{})
		if err != nil || !got {
			t.Fatalf("neq on absent attr must be true: ok=%v err=%v", got, err)
		}
	})
	t.Run("not_in-absent", func(t *testing.T) {
		got, err := evalOne(cond("missing", "not_in", []any{"x"}), map[string]any{})
		if err != nil || !got {
			t.Fatalf("not_in on absent attr must be true: ok=%v err=%v", got, err)
		}
	})
	t.Run("neq-present-equal", func(t *testing.T) {
		got, err := evalOne(cond("s", "neq", "x"), map[string]any{"s": "x"})
		if err != nil || got {
			t.Fatalf("neq on equal present attr must be false: ok=%v err=%v", got, err)
		}
	})
	t.Run("not_in-present-member", func(t *testing.T) {
		got, err := evalOne(cond("s", "not_in", []any{"x", "y"}), map[string]any{"s": "x"})
		if err != nil || got {
			t.Fatalf("not_in on member must be false: ok=%v err=%v", got, err)
		}
	})
}

// TestErrorPropagatesFromMatches confirms Matches surfaces a malformed
// condition's error (Engine.Matches error arm) even when earlier conditions pass.
func TestErrorPropagatesFromMatches(t *testing.T) {
	attrs := map[string]any{"a": "1", "x": "y"}
	ok, err := policy.New().Matches([]authz.Condition{
		cond("a", "eq", "1"),    // passes
		rawCond("x", "eq", "{"), // malformed → error
	}, attrs)
	if err == nil {
		t.Fatalf("Matches must propagate a malformed condition error, got ok=%v", ok)
	}
	if ok {
		t.Fatal("ok must be false when an error occurs")
	}
}
