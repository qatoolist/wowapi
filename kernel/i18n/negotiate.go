package i18n

import (
	"sort"
	"strconv"
	"strings"
)

// Negotiate picks the best locale from an HTTP Accept-Language header value
// against the supported list, falling back to def when the header is empty,
// unparseable, or names nothing supported.
//
// It implements RFC 9110 §12.5.4 content negotiation, ported from a
// battle-tested product implementation. It is deliberately narrow: it matches
// on the primary language subtag only (a supported "mr" matches an offered
// "mr-IN"), which covers the common case without a full BCP 47 tag matcher. A
// "*" wildcard is intentionally NOT treated as a match for any specific
// supported locale — it expresses no preference, so we fall back to def. An
// offer with q=0 is an explicit refusal and is skipped.
func Negotiate(acceptLanguage string, supported []string, def string) string {
	type candidate struct {
		tag string
		q   float64
	}

	var candidates []candidate
	for _, part := range strings.Split(acceptLanguage, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag, qRaw, hasQ := strings.Cut(part, ";")
		tag = strings.TrimSpace(tag)
		if tag == "" || tag == "*" {
			continue
		}
		q := 1.0
		if hasQ {
			qRaw = strings.TrimSpace(qRaw)
			qRaw = strings.TrimPrefix(qRaw, "q=")
			qRaw = strings.TrimPrefix(qRaw, "Q=")
			if v, err := strconv.ParseFloat(qRaw, 64); err == nil {
				q = v
			} else {
				continue // malformed q-value: skip this offer
			}
		}
		if q <= 0 {
			continue
		}
		candidates = append(candidates, candidate{tag: strings.ToLower(tag), q: q})
	}

	// Stable sort by descending q so equal-q offers keep header order (the
	// leftmost offer wins ties, matching Accept-Language convention).
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].q > candidates[j].q })

	supportedSet := make(map[string]bool, len(supported))
	for _, s := range supported {
		supportedSet[strings.ToLower(s)] = true
	}

	for _, c := range candidates {
		base, _, _ := strings.Cut(c.tag, "-")
		if supportedSet[c.tag] {
			return c.tag
		}
		if supportedSet[base] {
			return base
		}
	}
	return def
}
