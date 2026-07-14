package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type runReport struct {
	examples   int
	futureDocs int
}

func main() {
	root := flag.String("root", ".", "repository root")
	write := flag.Bool("write-reference", false, "regenerate the ApplicationModel reference table before checking")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *write {
		if err := writeReference(absRoot); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	report, err := run(context.Background(), absRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("docs-check: compiled %d tagged examples; linted %d future-state documents; generated reference byte-match ok\n", report.examples, report.futureDocs)
}

func run(ctx context.Context, root string) (runReport, error) {
	docs, err := normativeDocs(root)
	if err != nil {
		return runReport{}, err
	}
	var examples []example
	for _, rel := range docs {
		data, err := os.ReadFile(filepath.Join(root, rel)) // #nosec G304 -- documentation tool reads repository-relative paths from its own fixed discovery set
		if err != nil {
			return runReport{}, fmt.Errorf("read %s: %w", rel, err)
		}
		found, err := extractExamples(rel, data)
		if err != nil {
			return runReport{}, err
		}
		examples = append(examples, found...)
	}
	if err := compileExamples(ctx, root, examples); err != nil {
		return runReport{}, err
	}

	futureDocs, err := futureStateDocs(root)
	if err != nil {
		return runReport{}, err
	}
	for _, rel := range futureDocs {
		data, err := os.ReadFile(filepath.Join(root, rel)) // #nosec G304 -- documentation tool reads repository-relative paths from its own fixed discovery set
		if err != nil {
			return runReport{}, fmt.Errorf("read %s: %w", rel, err)
		}
		if err := lintFutureState(rel, data); err != nil {
			return runReport{}, err
		}
	}
	if err := checkReference(root); err != nil {
		return runReport{}, err
	}
	return runReport{examples: len(examples), futureDocs: len(futureDocs)}, nil
}

func normativeDocs(root string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "docs", "blueprint", "*.md"))
	if err != nil {
		return nil, fmt.Errorf("find blueprint docs: %w", err)
	}
	paths := make([]string, 0, len(matches)+1)
	paths = append(paths, "README.md")
	for _, path := range matches {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil, fmt.Errorf("locate %s: %w", path, err)
		}
		paths = append(paths, filepath.ToSlash(rel))
	}
	sort.Strings(paths)
	return paths, nil
}

func futureStateDocs(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(filepath.Join(root, "docs"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "docs/blueprint/") || strings.HasSuffix(rel, "-target-design.md") {
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find future-state docs: %w", err)
	}
	sort.Strings(paths)
	return paths, nil
}
