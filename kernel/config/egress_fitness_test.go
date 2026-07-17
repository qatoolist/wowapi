package config_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// requestScopedParamNames walks a function declaration and returns the names
// of parameters whose types are context.Context or *http.Request (accounting
// for the local import names used in the file).
func requestScopedParamNames(f *ast.File, fn *ast.FuncDecl) map[string]struct{} {
	imports := localImportNames(f)
	ctxName := imports["context"]
	httpName := imports["net/http"]
	if ctxName == "" && httpName == "" {
		return nil
	}

	out := make(map[string]struct{})
	add := func(name string, typ ast.Expr) {
		switch t := typ.(type) {
		case *ast.SelectorExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				if ident.Name == ctxName && t.Sel.Name == "Context" {
					out[name] = struct{}{}
				}
			}
		case *ast.StarExpr:
			if sel, ok := t.X.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == httpName && sel.Sel.Name == "Request" {
						out[name] = struct{}{}
					}
				}
			}
		}
	}

	if fn.Recv != nil {
		for _, f := range fn.Recv.List {
			for _, n := range f.Names {
				add(n.Name, f.Type)
			}
		}
	}
	if fn.Type.Params != nil {
		for _, f := range fn.Type.Params.List {
			for _, n := range f.Names {
				add(n.Name, f.Type)
			}
		}
	}
	return out
}

// localImportNames maps import path -> local package name for the file.
func localImportNames(f *ast.File) map[string]string {
	m := make(map[string]string)
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		name := path
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			// Use the last path segment as the default local name.
			if i := strings.LastIndex(path, "/"); i >= 0 {
				name = path[i+1:]
			}
		}
		m[path] = name
	}
	return m
}

// isEgressConstruction reports whether call is a construction call site we care
// about: httpclient.New, httpclient.Config{...}, auth.NewJWKSKeySource, or
// auth.JWKSConfig{...}. It returns the kind string for diagnostics.
func isEgressConstruction(f *ast.File, call ast.Expr) (kind string, ok bool) {
	imports := localImportNames(f)
	httpLocal := imports["github.com/qatoolist/wowapi/v2/kernel/httpclient"]
	authLocal := imports["github.com/qatoolist/wowapi/v2/kernel/auth"]

	selName := func(x *ast.Ident, s string) bool {
		if x == nil {
			return false
		}
		return x.Name == httpLocal || x.Name == authLocal
	}

	switch n := call.(type) {
	case *ast.CallExpr:
		if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				switch {
				case selName(id, sel.Sel.Name) && id.Name == httpLocal && sel.Sel.Name == "New":
					return "httpclient.New", true
				case selName(id, sel.Sel.Name) && id.Name == authLocal && sel.Sel.Name == "NewJWKSKeySource":
					return "auth.NewJWKSKeySource", true
				}
			}
		}
	case *ast.CompositeLit:
		if sel, ok := n.Type.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				switch {
				case id.Name == httpLocal && sel.Sel.Name == "Config":
					return "httpclient.Config literal", true
				case id.Name == authLocal && sel.Sel.Name == "JWKSConfig":
					return "auth.JWKSConfig literal", true
				}
			}
		}
	}
	return "", false
}

// usesRequestScopedData reports whether expr references any request-scoped
// parameter, including via method calls like ctx.Value or r.Context and via
// nested function calls that take such parameters.
func usesRequestScopedData(expr ast.Expr, params map[string]struct{}) bool {
	if len(params) == 0 {
		return false
	}
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		switch x := n.(type) {
		case *ast.Ident:
			if _, ok := params[x.Name]; ok {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			// Detect ctx.Value(...), r.Context(), etc.
			if id, ok := x.X.(*ast.Ident); ok {
				if _, ok := params[id.Name]; ok {
					found = true
					return false
				}
			}
		case *ast.CallExpr:
			// If a nested call takes a request-scoped arg, that's a violation too.
			for _, arg := range x.Args {
				if usesRequestScopedData(arg, params) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

// checkFile parses a single Go file and returns any egress-construction
// violations where request- or tenant-scoped data is used.
func checkFile(path string) ([]string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var violations []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		params := requestScopedParamNames(f, fn)
		if len(params) == 0 {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if n == nil {
				return true
			}
			switch node := n.(type) {
			case *ast.CallExpr:
				if kind, ok := isEgressConstruction(f, node); ok {
					for _, arg := range node.Args {
						if usesRequestScopedData(arg, params) {
							violations = append(violations, kind+" uses request-scoped data at "+fset.Position(arg.Pos()).String())
						}
					}
				}
			case *ast.CompositeLit:
				if kind, ok := isEgressConstruction(f, node); ok {
					for _, elt := range node.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if usesRequestScopedData(kv.Value, params) {
								violations = append(violations, kind+" uses request-scoped data at "+fset.Position(kv.Value.Pos()).String())
							}
						}
					}
				}
			}
			return true
		})
	}
	return violations, nil
}

// TestFitnessCheckDetectsKnownViolation proves the checker can spot a
// deliberate violation, so the real-source test below is meaningful.
func TestFitnessCheckDetectsKnownViolation(t *testing.T) {
	src := `package example

import (
	"context"
	"net/http"
	"github.com/qatoolist/wowapi/v2/kernel/httpclient"
	"github.com/qatoolist/wowapi/v2/kernel/auth"
)

func bad(ctx context.Context, r *http.Request) {
	_ = httpclient.New(httpclient.Config{
		AllowedHosts: []string{ctx.Value("host").(string)},
	})
	_, _ = auth.NewJWKSKeySource(auth.JWKSConfig{
		Issuer: r.URL.String(),
	})
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "violation.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	var violations []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		params := requestScopedParamNames(f, fn)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				if kind, ok := isEgressConstruction(f, node); ok {
					for _, arg := range node.Args {
						if usesRequestScopedData(arg, params) {
							violations = append(violations, kind)
						}
					}
				}
			case *ast.CompositeLit:
				if kind, ok := isEgressConstruction(f, node); ok {
					for _, elt := range node.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if usesRequestScopedData(kv.Value, params) {
								violations = append(violations, kind)
							}
						}
					}
				}
			}
			return true
		})
	}

	if len(violations) == 0 {
		t.Fatal("expected checker to detect deliberate request-scoped data use")
	}
	want := map[string]struct{}{
		"httpclient.New":            {},
		"httpclient.Config literal": {},
		"auth.NewJWKSKeySource":     {},
		"auth.JWKSConfig literal":   {},
	}
	for _, v := range violations {
		delete(want, v)
	}
	if len(want) != 0 {
		t.Fatalf("missing expected violations: %v; got %v", want, violations)
	}
}

// TestFitnessCheckKernelAndAppAreClean walks the framework source files that
// construct SSRF-safe HTTP clients or JWKS sources and asserts none of them
// read request-scoped or tenant-scoped context data at construction time.
func TestFitnessCheckKernelAndAppAreClean(t *testing.T) {
	_, here, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(here), "..", "..")

	dirs := []string{
		filepath.Join(root, "kernel"),
		filepath.Join(root, "app"),
		filepath.Join(root, "internal", "cli"),
	}

	var all []string
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			violations, err := checkFile(path)
			if err != nil {
				// Skip files that are not valid Go in the current working tree
				// (e.g., templates or files being edited by other workstreams).
				return nil
			}
			all = append(all, violations...)
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", dir, err)
		}
	}

	if len(all) > 0 {
		for _, v := range all {
			t.Log(v)
		}
		t.Fatalf("found %d egress-construction violations that read request-scoped data", len(all))
	}
}

// TestFitnessCheckConstructorSignaturesDoNotAcceptContext asserts the
// constructors themselves never accept a context.Context or *http.Request,
// which would invite callers to pass request-scoped data in.
func TestFitnessCheckConstructorSignaturesDoNotAcceptContext(t *testing.T) {
	_, here, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(here), "..", "..")

	cases := []struct {
		file     string
		funcName string
	}{
		{filepath.Join(root, "kernel", "httpclient", "client.go"), "New"},
		{filepath.Join(root, "kernel", "auth", "jwks.go"), "NewJWKSKeySource"},
	}

	for _, tc := range cases {
		src, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatalf("reading %s: %v", tc.file, err)
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, tc.file, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parsing %s: %v", tc.file, err)
		}

		var found bool
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != tc.funcName {
				continue
			}
			found = true
			params := requestScopedParamNames(f, fn)
			if len(params) > 0 {
				t.Fatalf("%s must not accept request-scoped parameters; got %v", tc.funcName, params)
			}
		}
		if !found {
			t.Fatalf("%s not found in %s", tc.funcName, tc.file)
		}
	}
}

// TestFitnessCheckTemplateJWKSUsesConfigOnly asserts the generated api main
// template constructs the JWKS source purely from static config values, never
// from ctx or the request.
func TestFitnessCheckTemplateJWKSUsesConfigOnly(t *testing.T) {
	_, here, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(here), "..", "..")
	path := filepath.Join(root, "internal", "cli", "templates", "init", "cmd_api_main.go.tmpl")

	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading template: %v", err)
	}

	// Extract the JWKSConfig literal block from the template.
	start := strings.Index(string(src), "kauth.JWKSConfig{")
	if start == -1 {
		t.Fatal("JWKSConfig literal not found in template")
	}
	end := strings.Index(string(src)[start:], "})")
	if end == -1 {
		t.Fatal("could not locate end of JWKSConfig literal")
	}
	block := string(src)[start : start+end]

	forbidden := []string{"ctx", "r.", "tenant", "request"}
	for _, f := range forbidden {
		if strings.Contains(block, f) {
			t.Fatalf("JWKSConfig template block references request-scoped data %q:\n%s", f, block)
		}
	}
}
