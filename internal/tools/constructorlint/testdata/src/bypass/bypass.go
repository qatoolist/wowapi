package bypass

import wiring "github.com/qatoolist/wowapi/kernel/authz"

func build() {
	_ = wiring.NewStore()    // want `ad hoc infrastructure constructor authz.NewStore is only allowed in composition packages`
	_ = wiring.NewSQLStore() // want `ad hoc infrastructure constructor authz.NewSQLStore is only allowed in composition packages`
}
