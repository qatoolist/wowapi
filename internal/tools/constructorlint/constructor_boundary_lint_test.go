package constructorlint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzerRejectsAdHocInfrastructureConstructor(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "bypass")
	analysistest.Run(t, analysistest.TestData(), Analyzer, "github.com/qatoolist/wowapi/kernel")
}
