// Package safety defines the duplicate-safety contract for high-impact adapter
// operations. Adapters that perform externally visible work (sending a webhook,
// delivering a notification, etc.) must declare the mechanism that keeps their
// work duplicate-safe when retried.
package safety

// Mechanism names the duplicate-safety strategy an adapter uses.
type Mechanism int

const (
	// None means the adapter has no built-in duplicate-safety mechanism. The
	// caller/framework must provide duplicate suppression (e.g. lease fencing
	// and idempotency keys) around the adapter.
	None Mechanism = iota

	// InboxEffectLedger means the adapter writes a durable inbox/effect ledger
	// and deduplicates by a unique correlation key.
	InboxEffectLedger

	// DomainCAS means the adapter relies on a domain-level compare-and-swap or
	// conditional write for idempotency.
	DomainCAS

	// ProviderIdempotencyKey means the adapter carries a provider-supplied
	// idempotency key that the remote service honors.
	ProviderIdempotencyKey
)

// String returns a human-readable name for the mechanism.
func (m Mechanism) String() string {
	switch m {
	case None:
		return "none"
	case InboxEffectLedger:
		return "inbox_effect_ledger"
	case DomainCAS:
		return "domain_cas"
	case ProviderIdempotencyKey:
		return "provider_idempotency_key"
	default:
		return "unknown"
	}
}

// Declarer is the contract an adapter satisfies to register for a high-impact
// operation. Implementations return the duplicate-safety mechanism they use.
type Declarer interface {
	DuplicateSafety() Mechanism
}
