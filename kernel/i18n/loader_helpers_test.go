package i18n_test

import "github.com/qatoolist/wowapi/kernel/errors"

// kindNotFound returns errors.KindNotFound, whose KeyProblemTitle key
// ("kernel.problem.not_found") is a stable framework key present in the embedded
// defaults — the anchor for the override/precedence tests.
func kindNotFound() errors.Kind { return errors.KindNotFound }
