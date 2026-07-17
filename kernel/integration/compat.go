// Package integration preserves the v1 import path for the integration foundation.
package integration

import (
	"github.com/qatoolist/wowapi/v2/foundation/integration"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/secrets"
)

type (
	Config   = integration.Config
	Provider = integration.Provider
	Registry = integration.Registry
	Store    = integration.Store
	UpsertIn = integration.UpsertIn
)

func NewRegistry() *Registry { return integration.NewRegistry() }
func NewStore(reg *Registry, sec secrets.Provider, idgen model.IDGen) *Store {
	return integration.NewStore(reg, sec, idgen)
}
