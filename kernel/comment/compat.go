// Package comment preserves the v1 import path for the comment foundation.
package comment

import (
	"github.com/qatoolist/wowapi/v2/foundation/comment"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
)

type (
	Comment  = comment.Comment
	CreateIn = comment.CreateIn
	Service  = comment.Service
)

func New(idgen model.IDGen, ob outbox.Writer) *Service { return comment.New(idgen, ob) }
