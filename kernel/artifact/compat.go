// Package artifact preserves the v1 import path for the artifact foundation.
package artifact

import (
	"github.com/qatoolist/wowapi/foundation/artifact"
	"github.com/qatoolist/wowapi/kernel/model"
)

type (
	Artifact        = artifact.Artifact
	Input           = artifact.Input
	Pipeline        = artifact.Pipeline
	TemplateVersion = artifact.TemplateVersion
	Templates       = artifact.Templates
)

func New(idgen model.IDGen) *Pipeline { return artifact.New(idgen) }
func NewTemplates() *Templates        { return artifact.NewTemplates() }
