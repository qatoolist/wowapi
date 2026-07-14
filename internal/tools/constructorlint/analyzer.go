// Package constructorlint enforces the application composition boundary.
package constructorlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const modulePath = "github.com/qatoolist/wowapi"

var infrastructureConstructorSuffixes = [...]string{
	"Client",
	"Manager",
	"Pool",
	"Registry",
	"Repository",
	"Resolver",
	"Runtime",
	"Sender",
	"Store",
	"Writer",
}

// Analyzer rejects infrastructure construction outside application composition packages.
var Analyzer = &analysis.Analyzer{
	Name: "constructorboundary",
	Doc:  "rejects ad hoc infrastructure constructors outside composition packages",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if isCompositionPackage(pass.Pkg.Path()) {
		return nil, nil
	}

	for _, file := range pass.Files {
		filename := pass.Fset.PositionFor(file.Pos(), false).Filename
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !isInfrastructureConstructor(selector.Sel.Name) {
				return true
			}
			fn, ok := pass.TypesInfo.Uses[selector.Sel].(*types.Func)
			if !ok || fn.Pkg() == nil {
				return true
			}
			if !strings.HasPrefix(fn.Pkg().Path(), modulePath+"/") {
				return true
			}
			pass.Reportf(call.Pos(),
				"ad hoc infrastructure constructor %s.%s is only allowed in composition packages",
				fn.Pkg().Name(), fn.Name())
			return true
		})
	}
	return nil, nil
}

func isInfrastructureConstructor(name string) bool {
	if !strings.HasPrefix(name, "New") {
		return false
	}
	for _, suffix := range infrastructureConstructorSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func isCompositionPackage(path string) bool {
	// kernel/mfa is the time-bounded FBL-01 forwarding shim; its constructors
	// delegate to foundation/mfa and do not create independent infrastructure.
	if path == modulePath+"/kernel" || path == modulePath+"/kernel/mfa" {
		return true
	}
	for _, root := range []string{
		modulePath + "/app",
		modulePath + "/cmd",
		modulePath + "/internal/cli",
		modulePath + "/internal/tools",
		modulePath + "/testkit",
	} {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}
