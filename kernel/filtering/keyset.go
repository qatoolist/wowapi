package filtering

import (
	"strings"

	"github.com/qatoolist/wowapi/kernel/pagination"
)

// KeysetClause builds the "rows strictly after the cursor" predicate for
// keyset pagination, matching the given Sort's columns and directions exactly
// (blueprint 05 §2). It is injection-proof by the same construction as the
// filter/sort builders: column names come only from the Sort's allowlisted
// terms, never from the cursor; the cursor supplies only VALUES, bound as $N
// placeholders (review findings ARCH-31, SEC-22).
//
// For a sort (c1 d1, c2 d2, …) the predicate is the standard lexicographic
// expansion, correct for mixed directions:
//
//	(c1 OP1 v1)
//	 OR (c1 = v1 AND c2 OP2 v2)
//	 OR (c1 = v1 AND c2 = v2 AND c3 OP3 v3) …
//
// where OPi is ">" for ascending and "<" for descending. It returns "" (no
// predicate) when the sort or cursor is empty. Placeholders start at startArg;
// nextArg is the next free placeholder index.
//
// The cursor MUST carry a value for every sort column (it was minted from a row
// under this sort); a missing value is a KindValidation error, which also
// guards against a forged cursor whose keys do not match the sort (SEC-22).
func KeysetClause(s Sort, cur pagination.Cursor, startArg int) (sql string, args []any, nextArg int, err error) {
	terms := s.Terms()
	if len(terms) == 0 || cur.IsZero() {
		return "", nil, startArg, nil
	}
	values := cur.Values()

	// If the cursor was minted with a sort-spec signature (NextCursor), reject it
	// loudly when the current sort differs — this catches a direction flip or
	// column reorder that the column-set check below cannot see (roadmap R7). A
	// legacy cursor without a signature falls back to the column-set check only.
	if sig := cur.Sig(); sig != "" && sig != s.Signature() {
		return "", nil, startArg, validationErr("cursor was minted for a different sort order")
	}

	// Validate that the cursor provides exactly the sort columns — no more, no
	// less. Extra keys mean a forged/mismatched cursor.
	if len(values) != len(terms) {
		return "", nil, startArg, validationErr("cursor does not match the current sort")
	}
	colValue := make([]any, len(terms))
	for i, t := range terms {
		v, ok := values[t.Col]
		if !ok {
			return "", nil, startArg, validationErr("cursor is missing a value for the sort field")
		}
		colValue[i] = v
	}

	arg := startArg
	// Assign one placeholder per column value up front; each column value is
	// reused across the OR terms (equalities and the strict comparison).
	ph := make([]string, len(terms))
	for i := range terms {
		ph[i] = "$" + itoa(arg)
		args = append(args, colValue[i])
		arg++
	}

	var ors []string
	for i, t := range terms {
		var conj []string
		for j := 0; j < i; j++ {
			conj = append(conj, terms[j].Col+" = "+ph[j])
		}
		op := ">"
		if t.Dir == DirDesc {
			op = "<"
		}
		conj = append(conj, t.Col+" "+op+" "+ph[i])
		ors = append(ors, "("+strings.Join(conj, " AND ")+")")
	}
	return "(" + strings.Join(ors, " OR ") + ")", args, arg, nil
}

// NextCursor mints the opaque keyset cursor for the last row returned under sort
// s, binding s's signature so KeysetClause rejects it if a later request changes
// the sort order (roadmap R7). values must carry exactly one entry per sort
// column. An empty sort yields a signatureless cursor.
func NextCursor(s Sort, values map[string]any) (string, error) {
	return pagination.EncodeCursorWithSig(s.Signature(), values)
}

// itoa is a tiny local int→string to avoid pulling strconv for one call.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
