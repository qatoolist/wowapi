package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/qatoolist/wowapi/kernel/appmodel"
)

const referencePath = "docs/reference/application-model.md"

func canonicalManifest() appmodel.Manifest {
	return appmodel.Manifest{
		ID:      "requests",
		Version: "1.0.0",
		Routes: []appmodel.RouteDecl{
			{Method: "GET", Path: "/requests/healthz", Public: true},
			{Method: "POST", Path: "/requests", Permission: "requests.request.create"},
		},
		Permissions: []appmodel.PermissionDecl{
			{Key: "requests.request.create", Description: "Create a request"},
		},
		Resources: []appmodel.ResourceDecl{
			{Key: "requests.request", Description: "Request resource"},
		},
	}
}

func renderReference() []byte {
	return []byte(appmodel.GenerateProjections(canonicalManifest()).Doc + "\n")
}

func checkReference(root string) error {
	want := renderReference()
	path := filepath.Join(root, referencePath)
	got, err := os.ReadFile(path) // #nosec G304 -- documentation tool reads its fixed generated-reference path
	if err != nil {
		return fmt.Errorf("read generated reference %s: %w", referencePath, err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("generated reference %s is stale; run go run ./internal/tools/docexamples -write-reference", referencePath)
	}
	return nil
}

func writeReference(root string) error {
	path := filepath.Join(root, referencePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create generated reference directory: %w", err)
	}
	if err := os.WriteFile(path, renderReference(), 0o600); err != nil {
		return fmt.Errorf("write generated reference %s: %w", referencePath, err)
	}
	return nil
}
