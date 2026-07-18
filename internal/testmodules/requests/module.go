package requests

import (
	"context"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/resource/aggregate"
	"github.com/qatoolist/wowapi/module"
)

// Config is the module's typed configuration namespace (modules.requests.*).
type Config struct {
	SLAHours int `json:"sla_hours"`
}

// Module is the "requests" domain-neutral fixture module. It satisfies
// module.Module and serves as the canonical neutral example for the wowapi
// module-contract suite (blueprint 08 §2, 11 §4).
type Module struct{}

var _ module.Module = (*Module)(nil)

// Name returns the module identifier. All seeds, permissions, and resource
// types are prefixed "requests.".
func (m *Module) Name() string { return "requests" }

// DependsOn declares no dependencies: requests is a leaf fixture module.
func (m *Module) DependsOn() []string { return nil }

// Register wires the module into the framework. It only wires — no I/O.
func (m *Module) Register(mc module.Context) error {
	// Default before decode; overlay from config namespace if present.
	cfg := Config{SLAHours: 48}
	if err := mc.Config().Decode(&cfg); err != nil {
		return err
	}

	mc.Migrations(migrationsFS)
	mc.Seeds(seedsFS)
	mc.OpenAPI(openapiFragment)

	h := &Handlers{
		tx:     mc.Tx(),
		writer: aggregate.New(mc.Tx(), resource.NewRegistrar(), mc.Audit(), mc.Outbox()),
		authz:  mc.Authz(),
		val:    mc.Validator(),
		idgen:  mc.IDGen(),
	}

	r := mc.Routes()
	// Public first: net/http 1.22 resolves /requests/healthz before /{id}.
	r.Handle("GET", "/requests/healthz", httpx.RouteMeta{Public: true}, h.Healthz)
	r.Handle("POST", "/requests", httpx.RouteMeta{
		Permission: "requests.request.create",
		Request:    CreateRequest{},
	}, httpx.ValidatedHandler[CreateRequest](h.val, 64*1024, h.Create))
	r.Handle("GET", "/requests/{id}", httpx.RouteMeta{Permission: "requests.request.read"}, h.Read)
	r.Handle("GET", "/requests", httpx.RouteMeta{Permission: "requests.request.list"}, h.List)

	mc.Health("db", func(_ context.Context) error { return nil })
	mc.ProvidePort("requests.Lookup", &lookupImpl{tx: mc.Tx()})

	return nil
}

// Lookup is the inter-module port other modules consume via
// mc.Port("requests.Lookup"). Exercises the ProvidePort/Port wiring path.
type Lookup interface {
	ByID(ctx context.Context, id uuid.UUID) (*RequestDTO, error)
}

// lookupImpl is the in-process implementation of Lookup backed by TxManager.
type lookupImpl struct{ tx database.TxManager }

// ByID fetches a single request by id within the caller's tenant context.
// Full implementation (authz, proper error mapping) lands in Phase 5.
func (l *lookupImpl) ByID(ctx context.Context, id uuid.UUID) (*RequestDTO, error) {
	var dto RequestDTO
	if err := l.tx.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		row := db.QueryRow(ctx,
			`SELECT id, title, status FROM requests_request WHERE id = $1`, id)
		return row.Scan(&dto.ID, &dto.Title, &dto.Status)
	}); err != nil {
		return nil, err
	}
	return &dto, nil
}
