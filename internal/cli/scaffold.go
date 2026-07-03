// scaffold.go — shared helpers for wowapi scaffold commands (init, new-module, gen).
package cli

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates
var scaffoldFS embed.FS

// identRE matches the allowed pattern for module and resource names.
var identRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// renderTemplate parses and executes the named template from scaffoldFS,
// writing the result to w. name is the full path within the embed.FS
// (e.g. "templates/init/go.mod.tmpl").
func renderTemplate(name string, data any, w io.Writer) error {
	src, err := fs.ReadFile(scaffoldFS, name)
	if err != nil {
		return fmt.Errorf("template %s: %w", name, err)
	}
	t, err := template.New(filepath.Base(name)).Parse(string(src))
	if err != nil {
		return fmt.Errorf("template %s parse: %w", name, err)
	}
	return t.Execute(w, data)
}

// renderToFile renders tmplName with data to destPath, creating parent dirs.
// If force is false and the file already exists, it returns an error.
func renderToFile(destPath, tmplName string, data any, force bool) error {
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := renderTemplate(tmplName, data, &buf); err != nil {
		return err
	}
	out := buf.Bytes()
	// A codegen tool must emit gofmt-clean Go. Formatting also fails loudly on a
	// template that produced invalid Go, catching template bugs at generation time.
	if strings.HasSuffix(destPath, ".go") {
		formatted, err := format.Source(out)
		if err != nil {
			return fmt.Errorf("generated %s is not valid Go: %w", destPath, err)
		}
		out = formatted
	}
	return os.WriteFile(destPath, out, 0o644)
}

// writeEmpty writes an empty file to destPath, creating parent dirs.
// If force is false and the file already exists, it is silently skipped.
func writeEmpty(destPath string, force bool) error {
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			return nil // already exists, skip
		}
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, nil, 0o644)
}

// toCamel converts a snake_case identifier to CamelCase.
// "title" → "Title", "first_name" → "FirstName".
func toCamel(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p != "" {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}
