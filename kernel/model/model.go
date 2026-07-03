// Package model defines wowapi's base model primitives: embeddable structs
// for identity, tenancy, audit, versioning, temporal validity, and status;
// plus kernel-wide value objects for money, references, and time ranges.
//
// Composition rules and anti-patterns are specified in
// docs/blueprint/04-project-and-primitives.md §3. The key principle is
// composition over a god BaseModel — each entity embeds only the structs
// whose corresponding columns it actually carries.
//
// Import boundary: stdlib + github.com/google/uuid + github.com/shopspring/decimal only.
// This package is at the base of the dependency graph; adding anything else
// here pulls it into every consumer.
package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ---------------------------------------------------------------------------
// Base embeddable structs
// ---------------------------------------------------------------------------

// BaseFields carries identity only. Embed in every persisted entity.
type BaseFields struct {
	ID uuid.UUID `db:"id"`
}

// TenantScoped marks an entity as tenant-owned. Repo helpers key on the
// presence of this struct — an entity without it cannot use TenantDB write helpers.
type TenantScoped struct {
	TenantID uuid.UUID `db:"tenant_id"`
}

// Auditable records who created and last modified a row. Embed in mutable
// entities. NOT for append-only rows — use CreatedOnly there.
type Auditable struct {
	CreatedAt time.Time  `db:"created_at"`
	CreatedBy uuid.UUID  `db:"created_by"`
	UpdatedAt *time.Time `db:"updated_at"`
	UpdatedBy *uuid.UUID `db:"updated_by"`
}

// CreatedOnly is the append-only variant of Auditable. Once a row is
// inserted its authorship is immutable; there is no UpdatedAt/UpdatedBy column.
type CreatedOnly struct {
	CreatedAt time.Time `db:"created_at"`
	CreatedBy uuid.UUID `db:"created_by"`
}

// Versioned supports optimistic locking. Embed in user-editable aggregates.
// Anti-pattern: embedding on append-only tables or high-frequency counters.
type Versioned struct {
	Version int `db:"version"`
}

// Temporal carries a validity window for history-aware rows (assignments,
// relationships, grants). Only embed where "as-of" queries are a real
// requirement — every temporal table pays query complexity forever.
type Temporal struct {
	ValidFrom time.Time  `db:"valid_from"`
	ValidTo   *time.Time `db:"valid_to"`
}

// ActiveAt reports whether the temporal row is active at the given instant.
//
// Boundary semantics: at == ValidFrom is active; at == ValidTo is NOT active
// (half-open interval [ValidFrom, ValidTo)).
func (t Temporal) ActiveAt(at time.Time) bool {
	if at.Before(t.ValidFrom) {
		return false
	}
	if t.ValidTo != nil && !at.Before(*t.ValidTo) {
		return false
	}
	return true
}

// Statused carries a lifecycle status using a typed string constant.
// Embed instead of soft-delete booleans; status vocabulary is per-entity.
type Statused[S ~string] struct {
	Status S `db:"status"`
}

// ---------------------------------------------------------------------------
// Value objects
// ---------------------------------------------------------------------------

// ResourceRef is a kernel-wide pointer to any domain object. It is used
// wherever code must reference an entity without importing that entity's package.
type ResourceRef struct {
	Type string
	ID   uuid.UUID
}

// ActorKind identifies the kind of principal performing an action.
type ActorKind string

const (
	// KindUser represents a human user principal.
	KindUser ActorKind = "user"
	// KindSystem represents an automated system principal.
	KindSystem ActorKind = "system"
)

// ActorRef identifies who performed an action. Only one of UserID, CapacityID,
// or System is meaningful for a given Kind; the others carry zero values.
type ActorRef struct {
	Kind       ActorKind
	UserID     uuid.UUID
	CapacityID uuid.UUID
	System     string
}

// Money is an exact monetary amount. Amount uses shopspring/decimal to avoid
// floating-point error. The DB representation is numeric + char(3).
type Money struct {
	Amount   decimal.Decimal
	Currency string
}

// TimeRange is a half-open time interval [From, To). To == nil means open-ended.
type TimeRange struct {
	From time.Time
	To   *time.Time
}

// Metadata is a schema-free extension bag for module-declared display extras.
// It must never drive core logic — the moment code branches on a key, promote
// that key to a typed column.
type Metadata map[string]any

// ExternalRef is a pointer into an external system (e.g. Stripe, Salesforce).
type ExternalRef struct {
	System string
	ID     string
}

// ---------------------------------------------------------------------------
// IDGen port + default
// ---------------------------------------------------------------------------

// IDGen produces primary keys. Injected everywhere IDs are minted so tests
// can run deterministic sequences (docs/blueprint/03 §1: UUIDv7, app-generated).
type IDGen interface {
	New() uuid.UUID
}

// uuidV7Gen is the production IDGen backed by time-ordered UUIDv7.
type uuidV7Gen struct{}

// New returns a new UUIDv7. uuid.Must panics on rand failure, which is
// unrecoverable in production — no silent fallback to v4.
func (uuidV7Gen) New() uuid.UUID { return uuid.Must(uuid.NewV7()) }

// UUIDv7 returns the production generator (time-ordered UUIDv7).
func UUIDv7() IDGen { return uuidV7Gen{} }
