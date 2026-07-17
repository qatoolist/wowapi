// Package attachment preserves the v1 import path for the attachment foundation.
package attachment

import (
	"github.com/qatoolist/wowapi/v2/foundation/attachment"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
)

type (
	AttachIn   = attachment.AttachIn
	Attachment = attachment.Attachment
	Service    = attachment.Service
)

func New(idgen model.IDGen, ob outbox.Writer) *Service { return attachment.New(idgen, ob) }
