package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	compileMarker      = "<!-- doc-example: compile -->"
	illustrativeMarker = "<!-- doc-example: illustrative -->"
)

type example struct {
	path   string
	line   int
	source []byte
}

func extractExamples(path string, data []byte) ([]example, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: read: %w", path, err)
	}

	var examples []example
	for i := 0; i < len(lines); i++ {
		marker := strings.TrimSpace(lines[i])
		if marker != compileMarker && marker != illustrativeMarker {
			if marker == "```go" {
				return nil, fmt.Errorf("%s:%d: Go fence must be classified with %s or %s", path, i+1, compileMarker, illustrativeMarker)
			}
			continue
		}
		if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "```go" {
			return nil, fmt.Errorf("%s:%d: %s must be immediately followed by ```go", path, i+1, marker)
		}
		start := i + 2
		end := start
		for end < len(lines) && strings.TrimSpace(lines[end]) != "```" {
			end++
		}
		if end == len(lines) {
			return nil, fmt.Errorf("%s:%d: tagged Go fence is not closed", path, i+2)
		}
		if marker == compileMarker {
			source := strings.Join(lines[start:end], "\n") + "\n"
			examples = append(examples, example{path: filepath.ToSlash(path), line: start + 1, source: []byte(source)})
		}
		i = end
	}
	return examples, nil
}

func compileExamples(ctx context.Context, root string, examples []example) error {
	if len(examples) == 0 {
		return fmt.Errorf("no %s examples found", compileMarker)
	}

	temp, err := os.MkdirTemp(root, ".docexamples-")
	if err != nil {
		return fmt.Errorf("create throwaway packages: %w", err)
	}
	defer func() { _ = os.RemoveAll(temp) }()

	for i, ex := range examples {
		pkgDir := filepath.Join(temp, fmt.Sprintf("example-%03d", i+1))
		if err := os.Mkdir(pkgDir, 0o750); err != nil {
			return fmt.Errorf("create package for %s:%d: %w", ex.path, ex.line, err)
		}
		generated := fmt.Sprintf("//line %s:%d\n%s", ex.path, ex.line, ex.source)
		if err := os.WriteFile(filepath.Join(pkgDir, "example.go"), []byte(generated), 0o600); err != nil {
			return fmt.Errorf("write package for %s:%d: %w", ex.path, ex.line, err)
		}

		rel, err := filepath.Rel(root, pkgDir)
		if err != nil {
			return fmt.Errorf("locate package for %s:%d: %w", ex.path, ex.line, err)
		}
		cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(pkgDir, "example"), "./"+filepath.ToSlash(rel)) // #nosec G204 -- fixed go build invocation targets a generated package under the tool-created directory
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GOWORK=off")
		output, err := cmd.CombinedOutput()
		if err != nil {
			output = bytes.ReplaceAll(output, []byte(filepath.Base(temp)), []byte(".docexamples"))
			return fmt.Errorf("%s:%d: example does not compile: %w\n%s", ex.path, ex.line, err, bytes.TrimSpace(output))
		}
	}
	return nil
}
