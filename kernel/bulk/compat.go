// Package bulk preserves the v1 import path for the bulk foundation.
package bulk

import (
	"github.com/qatoolist/wowapi/v2/foundation/bulk"
	"github.com/qatoolist/wowapi/v2/kernel/model"
)

type (
	ItemFunc = bulk.ItemFunc
	Progress = bulk.Progress
	Service  = bulk.Service
)

func New(idgen model.IDGen) *Service { return bulk.New(idgen) }
