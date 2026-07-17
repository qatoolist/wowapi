package httpx

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"

	"github.com/google/uuid"
)

// ETagFrom renders a strong ETag from an aggregate's optimistic-lock version.
// Clients echo it back via If-Match on mutating requests.
func ETagFrom(version int) string { return fmt.Sprintf("%q", "v"+strconv.Itoa(version)) }

// RequireIfMatch parses the mandatory If-Match header into the version the
// client last saw. A missing header is KindValidation (the client must opt into
// optimistic concurrency); a malformed one is likewise a client error. The
// caller compares the returned version against the row's version and returns
// KindVersionConflict (412) on mismatch.
//
// The wildcard "If-Match: *" is intentionally NOT accepted: wowapi's optimistic
// concurrency requires the client to assert the concrete version it observed,
// so "match any" would defeat the guard it opts into (review finding ARCH-34).
func RequireIfMatch(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.Header.Get("If-Match"))
	if raw == "" {
		return 0, kerr.E(kerr.KindValidation, "validation_failed", "If-Match header is required for this operation")
	}
	if raw == "*" {
		return 0, kerr.E(kerr.KindValidation, "validation_failed", "If-Match must carry a concrete version tag, not \"*\"")
	}
	raw = strings.TrimPrefix(raw, "W/") // tolerate weak validators
	raw = strings.Trim(raw, `"`)
	raw = strings.TrimPrefix(raw, "v")
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, kerr.E(kerr.KindValidation, "validation_failed", "If-Match header is not a valid version tag")
	}
	return v, nil
}

// ParseResourceID reads a path parameter (net/http 1.22 pattern wildcard) and
// parses it as a UUID. A missing or malformed id is KindValidation.
func ParseResourceID(r *http.Request, param string) (uuid.UUID, error) {
	raw := r.PathValue(param)
	if raw == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "validation_failed", "missing path parameter: "+param)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "validation_failed", "invalid id in path parameter: "+param)
	}
	return id, nil
}

// AuditMetaFrom builds response AuditMeta from a model.Auditable and version.
func AuditMetaFrom(a model.Auditable, version int) *AuditMeta {
	return &AuditMeta{
		CreatedAt: a.CreatedAt,
		CreatedBy: a.CreatedBy,
		UpdatedAt: a.UpdatedAt,
		Version:   version,
	}
}
