package main

import (
	"os"

	"github.com/qatoolist/wowapi/internal/compatcli"
)

func main() {
	os.Exit(compatcli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
