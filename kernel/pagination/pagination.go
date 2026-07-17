// Package pagination provides wowapi's page/cursor response envelopes and the
// opaque keyset cursor used for feed-style listing. It is the kernel counterpart
// of the shapes documented in docs/blueprint/04 §4 (PageResponse, CursorPage)
// and the pagination half of docs/blueprint/05 §2 ("allowlist-driven; SQL
// injection impossible by construction").
//
// A Cursor is an opaque, tamper-evident-by-decode-failure encoding of a keyset
// position — the physical column values of the last row returned. httpx's
// ParsePagination builds a Request from the raw per_page + cursor query params
// via Parse; the resulting Request carries a clamped Limit and the decoded
// Cursor. Attacker-supplied cursors that do not decode yield a KindValidation
// error and never panic.
//
// The package name is pagination even though the blueprint refers to it as
// "page" in some signatures.
package pagination

import (
	"strconv"
	"strings"

	"github.com/qatoolist/wowapi/kernel/errors"
)

// PageResponse is the offset-page envelope (admin/small lists). TotalCount is
// omitted from the wire when a COUNT would be too expensive. See 04 §4.
type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalCount int64 `json:"total_count,omitempty"`
}

// CursorPage is the cursor-page envelope (default for feeds/large lists).
// NextCursor is an opaque base64url(JSON) keyset position (see Cursor); it is
// omitted when there is no further page. See 04 §4.
type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// Defaults configures per_page clamping for Parse.
//
//	PerPage    – page size used when the client omits per_page (or sends 0).
//	MaxPerPage – hard upper bound; requests above it are clamped down. A value
//	             <= 0 disables the upper bound.
type Defaults struct {
	PerPage    int
	MaxPerPage int
}

// Request is the parsed, validated pagination input for a list query: a page
// size clamped to [1, MaxPerPage] and the decoded keyset Cursor.
type Request struct {
	Limit  int
	Cursor Cursor
}

// Parse turns the raw per_page and cursor query-parameter strings into a
// Request. per_page handling (documented contract):
//
//	""        → Defaults.PerPage
//	"0"       → Defaults.PerPage
//	> Max     → Defaults.MaxPerPage (when Max > 0)
//	negative  → KindValidation error
//	non-int   → KindValidation error
//
// A malformed cursor yields a KindValidation error (see DecodeCursor); an empty
// cursor yields a zero Cursor (IsZero == true), i.e. "start from the beginning".
func Parse(perPageRaw, cursorRaw string, def Defaults) (Request, error) {
	limit, err := clampPerPage(perPageRaw, def)
	if err != nil {
		return Request{}, err
	}
	cur, err := DecodeCursor(cursorRaw)
	if err != nil {
		return Request{}, err
	}
	return Request{Limit: limit, Cursor: cur}, nil
}

// defaultPerPage validates the configured default before it becomes a limit:
// the documented Request contract is a limit clamped to [1, MaxPerPage], and a
// zero/negative Defaults.PerPage is server misconfiguration, not client input —
// it must fail loudly (adversarial review 2026-07-17, F-08), never flow into
// SQL as LIMIT 0 or a negative limit a caller may read as "unlimited".
func defaultPerPage(def Defaults) (int, error) {
	if def.PerPage < 1 {
		return 0, errors.E(errors.KindInternal, "internal",
			"pagination: Defaults.PerPage must be positive")
	}
	return clampMax(def.PerPage, def), nil
}

func clampPerPage(raw string, def Defaults) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultPerPage(def)
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.E(errors.KindValidation, "validation_failed", "per_page must be an integer")
	}
	if n < 0 {
		return 0, errors.E(errors.KindValidation, "validation_failed", "per_page must not be negative")
	}
	if n == 0 {
		return defaultPerPage(def)
	}
	return clampMax(n, def), nil
}

func clampMax(n int, def Defaults) int {
	if def.MaxPerPage > 0 && n > def.MaxPerPage {
		return def.MaxPerPage
	}
	return n
}
