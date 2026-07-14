// Package document preserves the v1 import path for the document foundation.
package document

import (
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/storage"
)

type (
	AccessEvent   = document.AccessEvent
	AccessHook    = document.AccessHook
	Class         = document.Class
	ConfirmInput  = document.ConfirmInput
	CreateInput   = document.CreateInput
	Download      = document.Download
	DownloadInput = document.DownloadInput
	GrantInput    = document.GrantInput
	Hooks         = document.Hooks
	Registry      = document.Registry
	Sensitivity   = document.Sensitivity
	Service       = document.Service
	UploadEvent   = document.UploadEvent
	UploadHook    = document.UploadHook
	UploadSession = document.UploadSession
)

const (
	SensitivityPublic       = document.SensitivityPublic
	SensitivityInternal     = document.SensitivityInternal
	SensitivityConfidential = document.SensitivityConfidential
	SensitivityRestricted   = document.SensitivityRestricted
	PermRead                = document.PermRead
	PermWrite               = document.PermWrite
)

func NewRegistry() *Registry { return document.NewRegistry() }
func NewHooks() *Hooks       { return document.NewHooks() }
func New(reg *Registry, store storage.Adapter, ev authz.Evaluator, ob outbox.Writer, hooks *Hooks, idgen model.IDGen) *Service {
	return document.New(reg, store, ev, ob, hooks, idgen)
}
