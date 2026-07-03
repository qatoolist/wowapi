package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Fingerprint identifies an effective configuration: the SHA-256 of its
// canonical *redacted* JSON rendering. Secret values never enter the hash
// (Secret marshals as its redaction marker), so the fingerprint is safe to
// log, expose as a metric label, and include in /readyz output — and two
// processes sharing config sections can be compared for drift (12 §7).
//
// Note the redaction consequence: rotating a secret's VALUE (same ref) does
// not change the fingerprint; changing the reference does.
type Fingerprint [sha256.Size]byte

// String returns the full lowercase hex digest.
func (f Fingerprint) String() string { return hex.EncodeToString(f[:]) }

// Short returns the first 12 hex chars — enough for log correlation.
func (f Fingerprint) Short() string { return f.String()[:12] }

// FingerprintOf hashes the canonical redacted JSON rendering of v.
// v is normally a bound config struct; json.Marshal is deterministic for
// structs (field order) and maps (sorted keys), making the hash canonical.
func FingerprintOf(v any) (Fingerprint, error) {
	js, err := json.Marshal(v)
	if err != nil {
		return Fingerprint{}, fmt.Errorf("config: fingerprint: %w", err)
	}
	return sha256.Sum256(js), nil
}
