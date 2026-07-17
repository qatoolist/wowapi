package httpx

import (
	"net/http"

	"github.com/qatoolist/wowapi/kernel/filtering"
	"github.com/qatoolist/wowapi/kernel/pagination"
)

// The list helpers read query parameters and hand back the allowlist-driven,
// injection-proof structures from kernel/pagination and kernel/filtering. The
// wire formats:
//   ?per_page=50&cursor=<opaque>
//   ?filter.<field>=<op>:<value>        (repeatable; e.g. filter.status=eq:active)
//   ?sort=<field>:asc,<field>:desc

// ParsePagination reads per_page + cursor and returns the clamped page request.
func ParsePagination(r *http.Request, def pagination.Defaults) (pagination.Request, error) {
	q := r.URL.Query()
	return pagination.Parse(q.Get("per_page"), q.Get("cursor"), def)
}

// ParseFilters collects filter.<field>=<op>:<value> query params and parses
// them against the allowlist. Unknown fields or disallowed operators are
// KindValidation errors — the client can only ever reference allowlisted
// columns, and values always become bound parameters.
func ParseFilters(r *http.Request, allow filtering.Allowlist) (filtering.Set, error) {
	const prefix = "filter."
	raw := map[string][]string{}
	for key, vals := range r.URL.Query() {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			raw[key[len(prefix):]] = vals
		}
	}
	return filtering.Parse(raw, allow)
}

// ParseSort reads the sort query parameter against the sort allowlist.
func ParseSort(r *http.Request, allow filtering.SortAllowlist) (filtering.Sort, error) {
	return filtering.ParseSort(r.URL.Query().Get("sort"), allow)
}
