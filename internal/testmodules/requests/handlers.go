package requests

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/resource/aggregate"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// Handlers holds injected collaborators for every route in this module.
//
// Route-level authorization (the RouteMeta permission) is now enforced by the
// framework's httpx.SecureHandler gate BEFORE a handler runs (AuthN → bind
// tenant/actor → AuthZ(permission)), so these handlers already run only for an
// authorized actor. The stored evaluator is for FINE-GRAINED, resource-scoped
// checks a handler makes against a concrete target (e.g. "read THIS request");
// this neutral fixture module has none, so it is unused here.
type Handlers struct {
	tx     database.TxManager
	writer *aggregate.Writer
	authz  authz.Evaluator //nolint:unused // for resource-scoped checks; unused in this fixture
	val    *validation.Validator
	idgen  model.IDGen
}

// Create handles POST /requests.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request, req CreateRequest) {
	ctx := r.Context()
	id := h.idgen.New()
	if err := h.writer.Write(ctx, aggregate.Write{
		Resource: resource.Ref{Type: "requests.request", ID: id},
		Label:    req.Title,
		Status:   "open",
		Audit:    audit.Entry{Action: "requests.request.create"},
		Event:    outbox.Event{Type: "requests.request.created"},
		Apply: func(ctx context.Context, db database.TenantDB, actorID uuid.UUID) error {
			_, err := db.Exec(ctx,
				`INSERT INTO requests_request (id, tenant_id, title, status, version, created_at, created_by)
			     VALUES ($1, app_tenant_id(), $2, 'open', 1, now(), $3)`,
				id, req.Title, actorID)
			return err
		},
	}); err != nil {
		httpx.WriteError(ctx, w, err)
		return
	}
	dto := RequestDTO{ID: id, Title: req.Title, Status: "open"}
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
