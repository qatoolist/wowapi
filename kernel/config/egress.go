package config

import "slices"

// EgressExceptionKind identifies a class of configured outbound-security escape
// hatch surfaced by the boot-time egress-exception report.
type EgressExceptionKind string

const (
	// EgressWebhookAllowedHosts is the exact-match hostname allowlist for
	// outbound webhook delivery.
	EgressWebhookAllowedHosts EgressExceptionKind = "webhook_allowed_hosts"
	// EgressWebhookAllowedCIDRs is the CIDR allowlist for resolved outbound
	// webhook delivery addresses.
	EgressWebhookAllowedCIDRs EgressExceptionKind = "webhook_allowed_cidrs"
	// EgressJWKSTrustedIssuers is the declared trusted-issuer allowlist that
	// governs custom JWKS HTTP-client injection (D-07).
	EgressJWKSTrustedIssuers EgressExceptionKind = "jwks_trusted_issuers"
)

// EgressException is one configured egress escape hatch, in a redacted form
// safe to log and expose in readiness output. No secret values are included.
type EgressException struct {
	Kind   EgressExceptionKind `json:"kind"`
	Values []string            `json:"values"`
}

// EgressExceptions enumerates every configured egress escape hatch in a
// redacted, credential-free form suitable for boot-time logging and readiness
// details. Empty allowlists are omitted so the report only calls attention to
// exceptions that are actually enabled.
func (f Framework) EgressExceptions() []EgressException {
	var out []EgressException
	if hosts := f.Webhook.Outbound.AllowedHosts; len(hosts) > 0 {
		out = append(out, EgressException{Kind: EgressWebhookAllowedHosts, Values: slices.Clone(hosts)})
	}
	if cidrs := f.Webhook.Outbound.AllowedCIDRs; len(cidrs) > 0 {
		out = append(out, EgressException{Kind: EgressWebhookAllowedCIDRs, Values: slices.Clone(cidrs)})
	}
	if issuers := f.Security.TrustedIssuers; len(issuers) > 0 {
		out = append(out, EgressException{Kind: EgressJWKSTrustedIssuers, Values: slices.Clone(issuers)})
	}
	return out
}

// AllowlistChange describes a mutation of the outbound webhook allowlist. It is
// the payload written by RecordAllowlistChange.
type AllowlistChange struct {
	Action   string   `json:"action"`
	OldHosts []string `json:"old_hosts"`
	NewHosts []string `json:"new_hosts"`
	OldCIDRs []string `json:"old_cidrs"`
	NewCIDRs []string `json:"new_cidrs"`
}

// AllowlistChangeRecorder receives an audit-visible record when the allowlist
// changes. Implementations may write to the durable audit store, structured
// logs, or an in-memory sink for tests. The record contains no secrets.
type AllowlistChangeRecorder func(AllowlistChange)

// RecordAllowlistChange compares before and after allowlist configurations and,
// if they differ, emits a redacted change record via rec. It is safe to call
// with a nil recorder (the change is simply not recorded). The record contains
// no credential or secret values — only the host/CIDR allowlists themselves.
func RecordAllowlistChange(before, after WebhookOutbound, rec AllowlistChangeRecorder) {
	if rec == nil {
		return
	}
	if slices.Equal(before.AllowedHosts, after.AllowedHosts) &&
		slices.Equal(before.AllowedCIDRs, after.AllowedCIDRs) {
		return
	}
	rec(AllowlistChange{
		Action:   "webhook.outbound.allowlist_changed",
		OldHosts: slices.Clone(before.AllowedHosts),
		NewHosts: slices.Clone(after.AllowedHosts),
		OldCIDRs: slices.Clone(before.AllowedCIDRs),
		NewCIDRs: slices.Clone(after.AllowedCIDRs),
	})
}
