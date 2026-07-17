// Command wowapi is the installable framework CLI:
//
//	go install github.com/qatoolist/wowapi/v2/cmd/wowapi@latest
package main

import (
	"os"

	"github.com/qatoolist/wowapi/v2/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
