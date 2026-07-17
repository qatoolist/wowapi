package bulk

import (
	"context"

	"github.com/qatoolist/wowapi/kernel/database"
)

// SetCancelInterceptor wires the test-only fault-injection seam invoked between
// Cancel's aggregate transition and its item cleanup (same tenant tx). Compiled
// only into test binaries; the production API stays free of test seams.
func SetCancelInterceptor(s *Service, fn func(ctx context.Context, db database.TenantDB) error) {
	s.cancelInterceptor = fn
}
