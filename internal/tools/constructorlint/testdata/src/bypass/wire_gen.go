package bypass

import "github.com/qatoolist/wowapi/v2/kernel/authz"

func generatedBuild() {
	_ = authz.NewStore() // want `ad hoc infrastructure constructor authz.NewStore is only allowed in composition packages`
}
