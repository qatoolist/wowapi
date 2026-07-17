package kernel

import "github.com/qatoolist/wowapi/v2/kernel/authz"

func build() {
	_ = authz.NewStore()
}
