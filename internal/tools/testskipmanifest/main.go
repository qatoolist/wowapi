package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const manifestVersion = 1

type Manifest struct {
	Version int        `json:"version"`
	Skips   []Approval `json:"skips"`
}

type Approval struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	Function       string `json:"function"`
	Method         string `json:"method"`
	Message        string `json:"message"`
	Ordinal        int    `json:"ordinal"`
	Owner          string `json:"owner"`
	Classification string `json:"classification"`
	Rationale      string `json:"rationale"`
	Guard          string `json:"guard,omitempty"`
}

type site struct {
	Path     string
	Function string
	Method   string
	Message  string
	Ordinal  int
	Line     int
}

func main() {
	root := flag.String("root", ".", "repository root")
	manifestPath := flag.String("manifest", "miscellaneous/test-skip-manifest.json", "approved test-skip manifest")
	flag.Parse()

	manifest, err := loadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "test-skip manifest:", err)
		os.Exit(1)
	}
	if err := validate(*root, manifest); err != nil {
		fmt.Fprintln(os.Stderr, "test-skip manifest validation failed:")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("test-skip manifest validated: %d approved site(s)\n", len(manifest.Skips))
}

func loadManifest(path string) (Manifest, error) {
	f, err := os.Open(path) // #nosec G304 -- validation tool intentionally reads the selected manifest
	if err != nil {
		return Manifest{}, err
	}
	defer func() { _ = f.Close() }()

	var manifest Manifest
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if err := ensureEOF(decoder); err != nil {
		return Manifest{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return manifest, nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); errors.Is(err, io.EOF) {
		return nil
	} else if err != nil {
		return err
	}
	return errors.New("multiple JSON values")
}

func validate(root string, manifest Manifest) error {
	var problems []string
	if manifest.Version != manifestVersion {
		problems = append(problems, fmt.Sprintf("manifest version = %d, want %d", manifest.Version, manifestVersion))
	}

	approved := make(map[string]Approval, len(manifest.Skips))
	ids := make(map[string]struct{}, len(manifest.Skips))
	for i, approval := range manifest.Skips {
		prefix := fmt.Sprintf("skips[%d]", i)
		var missing []string
		if strings.TrimSpace(approval.ID) == "" {
			missing = append(missing, "id")
		}
		if strings.TrimSpace(approval.Owner) == "" || approval.Owner == "unassigned" {
			missing = append(missing, "owner")
		}
		if strings.TrimSpace(approval.Rationale) == "" {
			missing = append(missing, "rationale")
		}
		if approval.Ordinal < 1 {
			missing = append(missing, "ordinal")
		}
		if len(missing) > 0 {
			problems = append(problems, fmt.Sprintf("%s missing valid %s", prefix, strings.Join(missing, ", ")))
		}
		switch approval.Classification {
		case "optional":
		case "required-fail-closed":
			if strings.TrimSpace(approval.Guard) == "" {
				problems = append(problems, fmt.Sprintf("%s required-fail-closed approval missing guard", prefix))
			}
		default:
			problems = append(problems, fmt.Sprintf("%s classification %q must be optional or required-fail-closed", prefix, approval.Classification))
		}
		if _, exists := ids[approval.ID]; approval.ID != "" && exists {
			problems = append(problems, fmt.Sprintf("duplicate approval id %q", approval.ID))
		}
		ids[approval.ID] = struct{}{}
		key := approvalKey(approval.Path, approval.Function, approval.Method, approval.Message, approval.Ordinal)
		if _, exists := approved[key]; exists {
			problems = append(problems, fmt.Sprintf("duplicate approval for %s", key))
		}
		approved[key] = approval
	}

	sites, err := scan(root)
	if err != nil {
		problems = append(problems, err.Error())
	} else {
		observed := make(map[string]site, len(sites))
		for _, found := range sites {
			key := approvalKey(found.Path, found.Function, found.Method, found.Message, found.Ordinal)
			observed[key] = found
			if _, ok := approved[key]; !ok {
				problems = append(problems, fmt.Sprintf("unapproved t.%s at %s:%d in %s: %q", found.Method, found.Path, found.Line, found.Function, found.Message))
			}
		}
		for key, approval := range approved {
			if _, ok := observed[key]; !ok {
				problems = append(problems, fmt.Sprintf("stale approval %s (%s): no matching t.%s site", approval.ID, key, approval.Method))
			}
		}
	}

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return errors.New(strings.Join(problems, "\n"))
}

func scan(root string) ([]site, error) {
	var sites []site
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "vendor", "testdata", ".cicache", ".fuzzcache":
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		return scanFile(root, path, &sites)
	})
	if err != nil {
		return nil, fmt.Errorf("scan test files: %w", err)
	}
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].Path != sites[j].Path {
			return sites[i].Path < sites[j].Path
		}
		return sites[i].Line < sites[j].Line
	})
	return sites, nil
}

func scanFile(root, path string, sites *[]site) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		ordinals := make(map[string]int)
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			receiver, ok := selector.X.(*ast.Ident)
			if !ok || receiver.Name != "t" || !isSkipMethod(selector.Sel.Name) {
				return true
			}
			message := firstMessage(call)
			ordinalKey := selector.Sel.Name + "\x00" + message
			ordinals[ordinalKey]++
			*sites = append(*sites, site{
				Path:     rel,
				Function: fn.Name.Name,
				Method:   selector.Sel.Name,
				Message:  message,
				Ordinal:  ordinals[ordinalKey],
				Line:     fset.Position(call.Pos()).Line,
			})
			return true
		})
	}
	return nil
}

func isSkipMethod(name string) bool {
	return name == "Skip" || name == "Skipf" || name == "SkipNow"
}

func firstMessage(call *ast.CallExpr) string {
	if len(call.Args) == 0 {
		return ""
	}
	literal, ok := call.Args[0].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "<non-literal>"
	}
	message, err := strconv.Unquote(literal.Value)
	if err != nil {
		return literal.Value
	}
	return message
}

func approvalKey(path, function, method, message string, ordinal int) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d", filepath.ToSlash(path), function, method, message, ordinal)
}
