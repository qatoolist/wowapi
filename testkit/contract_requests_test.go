package testkit_test

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/internal/testmodules/requests"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationRequestsModuleContract runs the framework's own module
// conformance suite against the neutral internal fixture module: it must boot,
// migrate + seed idempotently, enforce RLS on its tables, and reject an
// invalid config namespace (blueprint 08 §2 / 11).
func TestIntegrationRequestsModuleContract(t *testing.T) {
	testkit.RunModuleContract(t, &requests.Module{})
}
