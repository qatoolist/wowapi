// Package httpx is wowapi's HTTP toolbox: response envelopes, the RFC 9457
// problem-details error writer, strict JSON decoding, metadata-enforced route
// registration, and the request helpers module handlers compose. The kernel
// provides helpers; modules keep the request flow visible (blueprint 05 §1,
// 04 §4–5).
package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// APIResponse is the success envelope. Data is the resource DTO; Meta is
// optional (request id, audit info).
type APIResponse[T any] struct {
	Data T     `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// Meta carries response metadata.
type Meta struct {
	RequestID string     `json:"request_id"`
	Audit     *AuditMeta `json:"audit,omitempty"`
}

// AuditMeta echoes the persisted audit columns; clients return Version via
// If-Match for optimistic concurrency.
type AuditMeta struct {
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Version   int        `json:"version"`
}

// OK wraps a DTO in the success envelope with no meta.
func OK[T any](data T) APIResponse[T] { return APIResponse[T]{Data: data} }

// OKWithMeta wraps a DTO with meta.
func OKWithMeta[T any](data T, meta *Meta) APIResponse[T] {
	return APIResponse[T]{Data: data, Meta: meta}
}

// WriteJSON serializes body as JSON with the given status. It sets the content
// type before writing the status line. A marshal failure (a programming bug —
// DTOs must be marshalable) degrades to a 500 problem body.
func WriteJSON[T any](w http.ResponseWriter, status int, body T) {
	buf, err := json.Marshal(body)
	if err != nil {
		writeInternal(context.Background(), w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf)
}
