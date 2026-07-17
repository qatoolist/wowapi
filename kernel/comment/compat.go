// Package comment preserves the v1 import path for the comment foundation.
package comment

import (
	"github.com/qatoolist/wowapi/foundation/comment"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
)

type (
	Comment  = comment.Comment
	CreateIn = comment.CreateIn
	Service  = comment.Service
)

func New(idgen model.IDGen, ob outbox.Writer) *Service { return comment.New(idgen, ob) }
