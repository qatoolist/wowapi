// Package requests is a domain-neutral private fixture module used by the
// wowapi module-contract test suite (blueprint 08 §2, 11 §4). A "request" has
// a title and a lifecycle status — no housing-society or product-domain terms.
package requests

import "github.com/google/uuid"

// CreateRequest is the wire DTO decoded from POST /requests.
type CreateRequest struct {
	Title string `json:"title" validate:"required"`
}

// RequestDTO is the success-response representation of a request aggregate.
// Domain types never serialize directly — this DTO is the only wire shape.
type RequestDTO struct {
	ID     uuid.UUID `json:"id"`
	Title  string    `json:"title"`
	Status string    `json:"status"`
}

// toDTO maps from scalar columns to RequestDTO.
func toDTO(id uuid.UUID, title, status string) RequestDTO {
	return RequestDTO{ID: id, Title: title, Status: status}
}
