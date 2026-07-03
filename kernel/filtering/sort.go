package filtering

import "strings"

// Dir is a sort direction. The set is closed.
type Dir string

const (
	DirAsc  Dir = "asc"
	DirDesc Dir = "desc"
)

// SortSpec maps a client-facing sort key to a physical column. Col is
// framework-controlled and is the only text that reaches the SQL column position.
type SortSpec struct {
	Col string
}

// SortAllowlist maps client-facing sort keys to their SortSpec.
type SortAllowlist map[string]SortSpec

// sortKey is one resolved, ordered sort term.
type sortKey struct {
	col string
	dir Dir
}

// Sort is an ordered list of validated sort terms.
type Sort struct {
	keys []sortKey
}

// ParseSort validates a raw sort string against the allowlist. Wire format
// (documented contract): a comma-separated list of "key[:dir]" terms, e.g.
// "created_at:desc,id:asc". A term without ":dir" defaults to ascending. An
// empty raw string yields an empty Sort (SQL == "").
//
// Errors (all KindValidation): unknown sort key, or a direction other than
// "asc"/"desc". Client text only ever selects an allowlisted key and one of the
// two direction constants — it never reaches the SQL text.
func ParseSort(raw string, allow SortAllowlist) (Sort, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Sort{}, nil
	}
	var s Sort
	for _, term := range strings.Split(raw, ",") {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		key, dirRaw, hasDir := strings.Cut(term, ":")
		key = strings.TrimSpace(key)
		spec, ok := allow[key]
		if !ok {
			return Sort{}, validationErr("unknown sort field %q", key)
		}
		dir := DirAsc
		if hasDir {
			switch Dir(strings.ToLower(strings.TrimSpace(dirRaw))) {
			case DirAsc:
				dir = DirAsc
			case DirDesc:
				dir = DirDesc
			default:
				return Sort{}, validationErr("invalid sort direction %q on field %q", dirRaw, key)
			}
		}
		s.keys = append(s.keys, sortKey{col: spec.Col, dir: dir})
	}
	return s, nil
}

// SQL renders the sort as an "ORDER BY <col> <DIR>, ..." clause using only
// allowlisted physical columns and validated direction keywords. An empty Sort
// returns "". The output contains no client-supplied text and needs no args.
func (s Sort) SQL() string {
	if len(s.keys) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("ORDER BY ")
	for i, k := range s.keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k.col)
		b.WriteByte(' ')
		if k.dir == DirDesc {
			b.WriteString("DESC")
		} else {
			b.WriteString("ASC")
		}
	}
	return b.String()
}

// IsEmpty reports whether the sort carries no terms.
func (s Sort) IsEmpty() bool { return len(s.keys) == 0 }

// Term is one resolved sort term exposed for keyset pagination: the physical
// column and its direction. Terms come only from allowlisted SortSpecs.
type Term struct {
	Col string
	Dir Dir
}

// Terms returns the ordered, resolved sort terms so a keyset predicate can be
// built matching the ORDER BY exactly (review finding ARCH-31).
func (s Sort) Terms() []Term {
	out := make([]Term, len(s.keys))
	for i, k := range s.keys {
		out[i] = Term{Col: k.col, Dir: k.dir}
	}
	return out
}
