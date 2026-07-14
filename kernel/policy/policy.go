// Package policy is wowapi's ABAC condition engine: it evaluates a policy's
// conditions against an attribute bag using a closed operator set. It is the
// PolicyEngine kernel/authz delegates to, kept separate so the matching logic
// is small, pure, and independently testable (blueprint 01 §3 step 5).
//
// Operators never execute arbitrary expressions; each is a fixed comparison, so
// a policy row can constrain but never inject behavior.
package policy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qatoolist/wowapi/kernel/authz"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Engine implements authz.PolicyEngine. It is stateless and safe for concurrent
// use.
type Engine struct{}

// New returns a policy engine.
func New() *Engine { return &Engine{} }

var _ authz.PolicyEngine = (*Engine)(nil)

// Matches reports whether every condition holds against attrs (logical AND —
// an empty condition set matches). A malformed condition (unknown operator,
// unparsable value) is an error, not a silent false: a broken policy must fail
// loud, never quietly stop denying.
func (Engine) Matches(conds []authz.Condition, attrs map[string]any) (bool, error) {
	for _, c := range conds {
		ok, err := matchOne(c, attrs)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func matchOne(c authz.Condition, attrs map[string]any) (bool, error) {
	actual, present := attrs[c.Attribute]

	switch c.Op {
	case "eq":
		want, err := decodeScalar(c.Value)
		if err != nil {
			return false, err
		}
		return present && scalarEqual(actual, want), nil
	case "neq":
		want, err := decodeScalar(c.Value)
		if err != nil {
			return false, err
		}
		return !present || !scalarEqual(actual, want), nil
	case "in":
		list, err := decodeList(c.Value)
		if err != nil {
			return false, err
		}
		return present && containsScalar(list, actual), nil
	case "not_in":
		list, err := decodeList(c.Value)
		if err != nil {
			return false, err
		}
		return !present || !containsScalar(list, actual), nil
	case "contains":
		// actual (a list attribute) contains the wanted scalar.
		want, err := decodeScalar(c.Value)
		if err != nil {
			return false, err
		}
		return listContains(actual, want), nil
	case "gte", "lte":
		return compareNumeric(c.Op, actual, present, c.Value)
	case "within":
		// env.time within [from,to] RFC3339 window: value = ["from","to"].
		return withinWindow(actual, present, c.Value)
	default:
		return false, kerr.E(kerr.KindInternal, "invalid_policy",
			"unknown policy operator: "+c.Op)
	}
}

func decodeScalar(raw json.RawMessage) (any, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, kerr.E(kerr.KindInternal, "invalid_policy", "unparsable policy value")
	}
	return v, nil
}

func decodeList(raw json.RawMessage) ([]any, error) {
	var v []any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, kerr.E(kerr.KindInternal, "invalid_policy", "policy value must be a list")
	}
	return v, nil
}

// scalarEqual compares attribute values by normalized form: JSON numbers decode
// as float64, so numeric attributes are compared numerically and everything
// else by string form.
func scalarEqual(a, b any) bool {
	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			return af == bf
		}
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func containsScalar(list []any, v any) bool {
	for _, e := range list {
		if scalarEqual(e, v) {
			return true
		}
	}
	return false
}

func listContains(actual any, want any) bool {
	list, ok := actual.([]any)
	if !ok {
		return false
	}
	return containsScalar(list, want)
}

func compareNumeric(op string, actual any, present bool, raw json.RawMessage) (bool, error) {
	if !present {
		return false, nil
	}
	want, err := decodeScalar(raw)
	if err != nil {
		return false, err
	}
	af, aok := toFloat(actual)
	bf, bok := toFloat(want)
	if !aok || !bok {
		return false, kerr.E(kerr.KindInternal, "invalid_policy",
			"gte/lte require numeric operands for attribute")
	}
	if op == "gte" {
		return af >= bf, nil
	}
	return af <= bf, nil
}

func withinWindow(actual any, present bool, raw json.RawMessage) (bool, error) {
	if !present {
		return false, nil
	}
	bounds, err := decodeList(raw)
	if err != nil || len(bounds) != 2 {
		return false, kerr.E(kerr.KindInternal, "invalid_policy", "within requires [from,to]")
	}
	at, err := time.Parse(time.RFC3339, fmt.Sprint(actual))
	if err != nil {
		// Deliberate fail-closed non-finding (nilerr; adjudicated in MATRIX
		// CS-23 / W01-E01-S002): an unparseable RUNTIME value makes this
		// condition evaluate false (deny) — it is not an internal error, unlike
		// a malformed POLICY, which line 161 above rejects loudly. Returning
		// the parse error here would turn every odd runtime attribute into a
		// policy-engine failure instead of a denied condition.
		return false, nil //nolint:nilerr // fail-closed by design: bad runtime value → condition false, not error
	}
	from, err1 := time.Parse(time.RFC3339, fmt.Sprint(bounds[0]))
	to, err2 := time.Parse(time.RFC3339, fmt.Sprint(bounds[1]))
	if err1 != nil || err2 != nil {
		return false, kerr.E(kerr.KindInternal, "invalid_policy", "within bounds must be RFC3339")
	}
	return !at.Before(from) && !at.After(to), nil
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}
