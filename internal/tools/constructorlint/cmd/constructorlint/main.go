package main

import (
	"github.com/qatoolist/wowapi/v2/internal/tools/constructorlint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(constructorlint.Analyzer)
}
