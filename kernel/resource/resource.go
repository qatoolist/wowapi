// Package resource is the kernel resource registry and the thin tenant mirror
// that lets kernel services (authz record-scope, comments, documents, workflow,
// relationships) address any module row uniformly. A module owns its business
// table; on write it upserts a matching row into the kernel `resources` table
// (same id) via a Registrar, declaring the resource type.
//
// The preferred write path is the aggregate.Writer in
// github.com/qatoolist/wowapi/kernel/resource/aggregate: a single call performs
// the business-row write, the resources-mirror upsert, an audit row, and an
// outbox event in one tenant transaction, so a module cannot commit its
// business row without also committing the mirror. The low-level Registrar
// remains available for legacy callers. See blueprint 01 §3, 03 §2, 04 §2.
package resource

import (
	"context"
	"regexp"

	"github.com/google/uuid"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Ref is a kernel-wide pointer to any domain object: its registered type key
// plus id. It is the currency of authz targets, relationships, comments, and
// attachments.
type Ref struct {
	Type string    `json:"type"` // registered resource_types.key, e.g. "requests.request"
	ID   uuid.UUID `json:"id"`
}

// IsZero reports whether the ref is unset.
func (r Ref) IsZero() bool { return r.Type == "" && r.ID == uuid.Nil }

// typeKeyRE constrains resource type keys to "module.name" lower-snake-dot.
var typeKeyRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// ValidTypeKey reports whether key is a legal resource type key.
func ValidTypeKey(key string) bool { return typeKeyRE.MatchString(key) }

// TypeSpec describes a resource type a module registers. Module is derived from
// the key prefix and validated against the registering module.
type TypeSpec struct {
	Key         string // "requests.request"
	Description string
}

// Registry accumulates resource type declarations during module registration
// and is synced to the resource_types table at boot. Duplicate keys and keys
// whose module prefix does not match the registering module are errors.
type Registry struct {
	specs  map[string]TypeSpec
	errs   []error
	sealed bool
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{specs: map[string]TypeSpec{}} }

// Seal freezes the registry once boot validation completes: any later Register
// panics rather than silently adding a resource type the boot gates never saw
// (closure review 2026-07-17, F-10).
func (r *Registry) Seal() { r.sealed = true }

// Register adds a resource type for the given module. The key must be
// "<module>.<name>"; a mismatch or duplicate records an error surfaced by Err.
func (r *Registry) Register(module string, spec TypeSpec) {
	if r.sealed {
		panic("resource: resource-type registration after boot: the extension model is sealed")
	}
	if !ValidTypeKey(spec.Key) {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_resource_type",
			"resource type key must be <module>.<name>: "+spec.Key))
		return
	}
	if prefix := module + "."; len(spec.Key) <= len(prefix) || spec.Key[:len(prefix)] != prefix {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_resource_type",
			"module "+module+" may not register resource type "+spec.Key))
		return
	}
	if _, dup := r.specs[spec.Key]; dup {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "duplicate_resource_type",
			"resource type registered more than once: "+spec.Key))
		return
	}
	r.specs[spec.Key] = spec
}

// Specs returns a COPY of the registered specs keyed by type key. Callers get
// a snapshot they can range and mutate freely without aliasing the registry's
// backing map (closure review 2026-07-17, F-10: exported backing maps let
// post-boot callers mutate sealed extension state).
func (r *Registry) Specs() map[string]TypeSpec {
	out := make(map[string]TypeSpec, len(r.specs))
	for k, v := range r.specs {
		out[k] = v
	}
	return out
}

// Err returns accumulated registration errors joined, or nil.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	return kerr.E(kerr.KindInternal, "resource_registration_failed", "resource type registration failed",
		joinErrs(r.errs))
}

// Registrar upserts the kernel resources mirror row for a module aggregate,
// inside the caller's tenant transaction so the mirror and the business write
// commit atomically. The concrete implementation lives beside the DB layer and
// receives the TenantDB; this port keeps module code off the raw pool.
type Registrar interface {
	// Upsert writes (or updates) the resources row: same id as the module row,
	// its resource type, optional org, a human label, and status.
	Upsert(ctx context.Context, ref Ref, orgID *uuid.UUID, label, status string) error
}

func joinErrs(errs []error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	msg := ""
	for i, e := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += e.Error()
	}
	return kerr.E(kerr.KindInternal, "internal", msg)
}
