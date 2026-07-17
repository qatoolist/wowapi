package kernel

import "github.com/qatoolist/wowapi/kernel/authz"

func build() {
	_ = authz.NewStore()
}
