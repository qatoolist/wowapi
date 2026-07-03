package requests

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// Handlers holds injected collaborators for every route in this module.
//
// TODO(phase-5): authz.Evaluator is stored but not called. Actor derivation
// from request context requires the auth middleware wired in Phase 5. The
// contract suite exercises registration, migration, seed, and RLS — not live
// authz through HTTP — so Evaluate is intentionally deferred.
type Handlers struct {
	tx    database.TxManager
	authz authz.Evaluator //nolint:unused // wired; called once actor-from-ctx lands
	val   *validation.Validator
	idgen model.IDGen
}

// Create handles POST /requests.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := httpx.BindAndValidate[CreateRequest](r, h.val, 64*1024)
	if err != nil {
		httpx.WriteError(ctx, w, err)
		return
	}
	id := h.idgen.New()
	var dto RequestDTO
	if err := h.tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, err := db.Exec(ctx,
			`INSERT INTO requests_request (id, tenant_id, title, status, version, created_at, created_by)
             VALUES ($1, app_tenant_id(), $2, 'open', 1, now(), $3)`,
			id, req.Title, uuid.Nil); err != nil { // TODO(phase-5): created_by from actor ctx
			return err
		}
		return resource.NewRegistrar().Bind(db).Upsert(ctx,
			resource.Ref{Type: "requests.request", ID: id}, nil, req.Title, "open")
	}); err != nil {
		httpx.WriteError(ctx, w, err)
		return
	}
	dto = RequestDTO{ID: id, Title: req.Title, Status: "open"}
	httpx.WriteJSON(w, http.StatusCreated, httpx.OK(dto))
}

// Read handles GET /requests/{id}.
func (h *Handlers) Read(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := httpx.ParseResourceID(r, "id")
	if err != nil {
		httpx.WriteError(ctx, w, err)
		return
	}
	var dto RequestDTO
	if err := h.tx.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		row := db.QueryRow(ctx,
			`SELECT id, title, status FROM requests_request WHERE id = $1`, id)
		return row.Scan(&dto.ID, &dto.Title, &dto.Status)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(ctx, w, kerr.E(kerr.KindNotFound, "not_found", "request not found"))
			return
		}
		httpx.WriteError(ctx, w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.OK(dto))
}

// List handles GET /requests.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dtos := make([]RequestDTO, 0)
	if err := h.tx.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx,
			`SELECT id, title, status FROM requests_request ORDER BY created_at DESC`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			var title, status string
			if err := rows.Scan(&id, &title, &status); err != nil {
				return err
			}
			dtos = append(dtos, toDTO(id, title, status))
		}
		return rows.Err()
	}); err != nil {
		httpx.WriteError(ctx, w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.OK(dtos))
}

// Healthz handles GET /requests/healthz (public, no auth required).
func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, httpx.OK(map[string]string{"status": "ok"}))
}
