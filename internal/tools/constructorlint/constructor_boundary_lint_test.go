package constructorlint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzerRejectsAdHocInfrastructureConstructor(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "bypass")
	analysistest.Run(t, analysistest.TestData(), Analyzer, "github.com/qatoolist/wowapi/v2/kernel")
}

func TestLegacyCompatibilityShimExceptionIsFileAndPackageScoped(t *testing.T) {
	t.Parallel()

	if !isLegacyCompatibilityShim(modulePath+"/kernel/document", "/repo/kernel/document/compat.go") {
		t.Fatal("document compatibility shim should be exempt")
	}
	for _, tc := range []struct {
		pkg  string
		file string
	}{
		{modulePath + "/kernel/document", "/repo/kernel/document/registry.go"},
		{modulePath + "/kernel/authz", "/repo/kernel/authz/compat.go"},
		{modulePath + "/foundation/document", "/repo/foundation/document/compat.go"},
	} {
		if isLegacyCompatibilityShim(tc.pkg, tc.file) {
			t.Fatalf("unexpected compatibility exemption for %s %s", tc.pkg, tc.file)
		}
	}
}
