// Package filtering is wowapi's allowlist-driven filter/sort builder — the
// mechanism behind docs/blueprint/05 §2, "Pagination / filtering / sorting
// (allowlist-driven; SQL injection impossible by construction)".
//
// Security invariant (enforced by construction, exercised by the tests):
//
//   - Column names in emitted SQL come ONLY from FieldSpec.Col / SortSpec.Col,
//     which are framework-controlled. A client picks a *key* into an Allowlist;
//     an unknown key is rejected with a KindValidation error. Client text never
//     becomes a column or an operator.
//   - Operators are validated against a fixed internal set AND the per-field
//     FieldSpec.Ops permit-list, then rendered from an internal map. Client text
//     never becomes an operator token.
//   - Client VALUES are never concatenated into SQL. Every value — including each
//     element of an "in" list — is emitted as a $N placeholder and appended to
//     the args slice. The literal value therefore cannot appear in the SQL text.
//
// The result: a caller can only ever produce SQL fragments that reference
// allowlisted physical columns with parameter placeholders, so injection is
// impossible regardless of what a client sends.
package filtering

import (
	"fmt"
	"sort"
	"strings"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Op is a filter comparison operator. The set is closed.
type Op string

const (
	OpEq   Op = "eq"
	OpNeq  Op = "neq"
	OpIn   Op = "in"
	OpGt   Op = "gt"
	OpGte  Op = "gte"
	OpLt   Op = "lt"
	OpLte  Op = "lte"
	OpLike Op = "like"
)

// sqlOps is the ONLY source of operator tokens rendered into SQL. A key absent
// here is not a valid operator.
var sqlOps = map[Op]string{
	OpEq:   "=",
	OpNeq:  "<>",
	OpIn:   "IN",
	OpGt:   ">",
	OpGte:  ">=",
	OpLt:   "<",
	OpLte:  "<=",
	OpLike: "LIKE",
}

// FieldSpec maps a client-facing field to a physical column and the operators
// permitted on it. Col is framework-controlled and is the only text that can
// reach the SQL column position.
type FieldSpec struct {
	Col string
	Ops []Op
}

// Allowlist maps client-facing field names to their FieldSpec. A field absent
// from the allowlist cannot be filtered.
type Allowlist map[string]FieldSpec

// Condition is one parsed, validated filter predicate. Field is the client-facing
// name (kept for introspection); the resolved physical column is held privately
// so only Parse can set it — external code cannot forge a column.
type Condition struct {
	Field  string
	Op     Op
	Values []any

	col string // resolved physical column (FieldSpec.Col); set by Parse only
}

// Set is an AND-combined collection of validated Conditions.
type Set struct {
	conds []Condition
}

// Parse validates raw client filter input against the allowlist and returns a
// Set. Wire format (documented contract): raw maps a client field name to one or
// more "op:value" entries.
//
//	{"status": {"eq:active"}}          → status = $n
//	{"age":    {"gte:18"}}             → age >= $n
//	{"status": {"in:active,pending"}}  → status IN ($n, $n+1)   (comma-separated)
//	{"name":   {"like:ac%"}}           → name LIKE $n
//
// Multiple entries for one field, and multiple fields, are AND-combined. Fields
// are processed in sorted order so placeholder numbering is deterministic.
//
// Errors (all KindValidation): unknown field, missing "op:" prefix, unknown
// operator, an operator not permitted by the field's FieldSpec.Ops, or an empty
// "in" list.
func Parse(raw map[string][]string, allow Allowlist) (Set, error) {
	fields := make([]string, 0, len(raw))
	for f := range raw {
		fields = append(fields, f)
	}
	sort.Strings(fields)

	var set Set
	for _, field := range fields {
		spec, ok := allow[field]
		if !ok {
			return Set{}, validationErr("unknown filter field %q", field)
		}
		for _, entry := range raw[field] {
			cond, err := parseEntry(field, spec, entry)
			if err != nil {
				return Set{}, err
			}
			set.conds = append(set.conds, cond)
		}
	}
	return set, nil
}

func parseEntry(field string, spec FieldSpec, entry string) (Condition, error) {
	rawOp, rawVal, ok := strings.Cut(entry, ":")
	if !ok {
		return Condition{}, validationErr("filter %q must be in \"op:value\" form", field)
	}
	op := Op(rawOp)
	if _, known := sqlOps[op]; !known {
		return Condition{}, validationErr("unknown filter operator %q on field %q", rawOp, field)
	}
	if !opAllowed(op, spec.Ops) {
		return Condition{}, validationErr("operator %q not permitted on field %q", op, field)
	}

	var values []any
	if op == OpIn {
		parts := strings.Split(rawVal, ",")
		for _, p := range parts {
			if p == "" {
				return Condition{}, validationErr("filter %q: empty value in \"in\" list", field)
			}
			values = append(values, p)
		}
		if len(values) == 0 {
			return Condition{}, validationErr("filter %q: \"in\" requires at least one value", field)
		}
	} else {
		values = []any{rawVal}
	}

	return Condition{Field: field, Op: op, Values: values, col: spec.Col}, nil
}

func opAllowed(op Op, allowed []Op) bool {
	for _, a := range allowed {
		if a == op {
			return true
		}
	}
	return false
}

// SQL renders the conditions into a boolean SQL fragment (no leading WHERE),
// appending each value to args as a $N placeholder. startArg is the next
// placeholder number (1-based); the returned nextArg is the following free
// number. An empty Set returns "", nil, startArg.
//
// Columns are taken from the resolved FieldSpec.Col; values only ever appear as
// $N. See the package doc for the injection-proof invariant.
func (s Set) SQL(startArg int) (sql string, args []any, nextArg int) {
	if len(s.conds) == 0 {
		return "", nil, startArg
	}
	var b strings.Builder
	n := startArg
	for i, c := range s.conds {
		if i > 0 {
			b.WriteString(" AND ")
		}
		token := sqlOps[c.Op]
		if c.Op == OpIn {
			b.WriteString(c.col)
			b.WriteString(" IN (")
			for j, v := range c.Values {
				if j > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "$%d", n)
				args = append(args, v)
				n++
			}
			b.WriteByte(')')
			continue
		}
		fmt.Fprintf(&b, "%s %s $%d", c.col, token, n)
		args = append(args, c.Values[0])
		n++
	}
	return b.String(), args, n
}

// Where wraps SQL with a leading "WHERE ". An empty Set returns "", nil,
// startArg so callers can append it unconditionally. Conditions are AND-combined.
func (s Set) Where(startArg int) (sql string, args []any, nextArg int) {
	frag, args, next := s.SQL(startArg)
	if frag == "" {
		return "", nil, startArg
	}
	return "WHERE " + frag, args, next
}

// Conditions returns a copy of the parsed conditions (introspection only).
func (s Set) Conditions() []Condition {
	if len(s.conds) == 0 {
		return nil
	}
	out := make([]Condition, len(s.conds))
	copy(out, s.conds)
	return out
}

// IsEmpty reports whether the set carries no conditions.
func (s Set) IsEmpty() bool { return len(s.conds) == 0 }

func validationErr(format string, a ...any) error {
	return errors.E(errors.KindValidation, "validation_failed", fmt.Sprintf(format, a...))
}
