package aggregate

import (
	"context"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

// ResolveActor resolves the acting principal an aggregate write is attributed
// to (DATA-06 T2 — the mechanism DATA-07 T3 consumes; see
// resource.PgRegistrar.UpsertAs for the mirror-write side).
//
// Resolution rules, in order:
//
//   - A user principal (httpx.ActorFrom, Kind == authz.ActorUser) MUST be
//     attributable: the audit actor bound by the authz gate
//     (database.ActorIDFrom — the acting capacity), else the principal's own
//     CapacityID, else its UserID. A user-initiated write with none of these
//     fails fast with KindUnauthenticated — it never proceeds with a
//     placeholder.
//   - A machine principal (system/webhook) uses the bound audit actor when one
//     exists, else a deterministic system-actor id derived from its System
//     name (SystemActorID) — attributable, stable, and never a real user's id.
//   - No principal at all is a system-initiated path (job runners and the
//     outbox relay bind only the tenant): the bound audit actor when present,
//     else SystemActorID("").
//
// The returned kind is the audit_logs actor_kind value ("user", "system",
// "webhook").
func ResolveActor(ctx context.Context) (uuid.UUID, string, error) {
	boundID, bound := database.ActorIDFrom(ctx)
	hasBound := bound && boundID != uuid.Nil

	principal, ok := httpx.ActorFrom(ctx)
	if !ok {
		if hasBound {
			return boundID, string(authz.ActorSystem), nil
		}
		return SystemActorID(""), string(authz.ActorSystem), nil
	}

	if principal.Kind == authz.ActorUser {
		switch {
		case hasBound:
			return boundID, string(authz.ActorUser), nil
		case principal.CapacityID != uuid.Nil:
			return principal.CapacityID, string(authz.ActorUser), nil
		case principal.UserID != uuid.Nil:
			return principal.UserID, string(authz.ActorUser), nil
		default:
			return uuid.Nil, "", kerr.E(kerr.KindUnauthenticated, "actor_required",
				"user-initiated aggregate write has no resolvable actor")
		}
	}

	kind := string(principal.Kind)
	if kind == "" {
		kind = string(authz.ActorSystem)
	}
	if hasBound {
		return boundID, kind, nil
	}
	return SystemActorID(principal.System), kind, nil
}

// SystemActorID derives the deterministic actor id for a named system
// principal ("outbox-relay", "webhook:payments", …), so system-initiated
// writes carry stable, traceable attribution in NOT NULL created_by columns
// instead of an anonymous placeholder. The empty name is the generic
// framework system actor. UUIDv5 over a fixed namespace: the same system name
// always maps to the same id, and no derived id can collide with a v4/v7
// user or capacity id in practice.
func SystemActorID(system string) uuid.UUID {
	if system == "" {
		system = "system"
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("wowapi.system-actor:"+system))
}
