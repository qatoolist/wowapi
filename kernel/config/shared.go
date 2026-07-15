package config

import "fmt"

// shared.go — cross-process config drift detection (blueprint 12 §7 / criterion
// #27). A deployment runs several processes (api, worker, migrate) from the same
// config; the sections they SHARE must be identical, or they behave
// inconsistently (e.g. api and worker pointed at different databases). The
// process-specific sections — HTTP (api-only) and Log (may differ by process) —
// are excluded, so a difference there is expected, not drift.

// SharedSection is the config subset that must match across every process of one
// deployment. The outbound allowlist and JWKS trusted-issuer config are
// included so that a change to any egress escape hatch is reflected in the
// shared fingerprint used for cross-process drift detection (SEC-06 T1/T4).
type SharedSection struct {
	Environment   Env      `json:"environment"`
	SchemaVersion int      `json:"schema_version"`
	DB            DB       `json:"db"`
	Security      Security `json:"security"`
	Webhook       Webhook  `json:"webhook"`
}

// SharedSection extracts the cross-process-shared configuration.
func (f Framework) SharedSection() SharedSection {
	return SharedSection{
		Environment:   f.Environment,
		SchemaVersion: f.SchemaVersion,
		DB:            f.DB,
		Security:      f.Security,
		Webhook:       f.Webhook,
	}
}

// SharedFingerprint is the fingerprint of the shared section only — the value
// api/worker/migrate compare to detect drift. Like Fingerprint it is redacted
// (secret VALUES never enter it), so it is safe to log and expose.
func (f Framework) SharedFingerprint() (Fingerprint, error) {
	return FingerprintOf(f.SharedSection())
}

// CheckSharedDrift reports an error when this process's shared-config fingerprint
// differs from expected (the hex fingerprint the deployment pins, e.g. via an
// env var stamped at release). An empty expected disables the check. Wire it as
// a startup gate or a /readyz check so a mis-deployed process fails loudly rather
// than silently diverging.
func (f Framework) CheckSharedDrift(expected string) error {
	if expected == "" {
		return nil
	}
	fp, err := f.SharedFingerprint()
	if err != nil {
		return err
	}
	got := fp.String()
	if got != expected {
		return fmt.Errorf("config: shared-section drift — this process %s != expected %s (api/worker/migrate deployed with divergent shared config)",
			shortHex(got), shortHex(expected))
	}
	return nil
}

func shortHex(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
