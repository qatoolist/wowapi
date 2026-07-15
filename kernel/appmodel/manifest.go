package appmodel

import (
	"fmt"
	"net/http"
	"strings"
)

// RouteDecl represents a single declared route in a module manifest.
type RouteDecl struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Permission string `json:"permission,omitempty"`
	Public     bool   `json:"public,omitempty"`
}

// PermissionDecl represents a single declared permission in a module manifest.
type PermissionDecl struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// ResourceDecl represents a single declared resource type in a module manifest.
type ResourceDecl struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// Manifest is the authoritative single source of truth for a module's declarations.
type Manifest struct {
	ID          string           `json:"id"`
	Version     string           `json:"version"`
	DependsOn   []string         `json:"depends_on,omitempty"`
	Routes      []RouteDecl      `json:"routes,omitempty"`
	Permissions []PermissionDecl `json:"permissions,omitempty"`
	Resources   []ResourceDecl   `json:"resources,omitempty"`
}

// Projections bundles the derived text-projections derived from a Manifest.
type Projections struct {
	Routes      string
	Permissions string
	Resources   string
	Schema      string
	OpenAPI     string
	Lifecycle   string
	Profile     string
	Test        string
	Doc         string
}

// GenerateProjections derives all 9 required projections from a Manifest.
func GenerateProjections(m Manifest) Projections {
	// 1. Routes projection
	var routes []string
	for _, r := range m.Routes {
		routes = append(routes, fmt.Sprintf("%s %s (Permission: %s, Public: %t)", r.Method, r.Path, r.Permission, r.Public))
	}

	// 2. Permissions projection
	var perms []string
	for _, p := range m.Permissions {
		perms = append(perms, fmt.Sprintf("%s: %s", p.Key, p.Description))
	}

	// 3. Resources projection
	var res []string
	for _, r := range m.Resources {
		res = append(res, fmt.Sprintf("%s: %s", r.Key, r.Description))
	}

	// 4. Schema projection
	var schemas []string
	for _, r := range m.Routes {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			schemas = append(schemas, fmt.Sprintf("Schema ref for %s %s: RequestPayload", r.Method, r.Path))
		}
	}

	// 5. OpenAPI projection
	var oas []string
	oas = append(oas, "paths:")
	for _, r := range m.Routes {
		oas = append(oas, fmt.Sprintf("  %s:", r.Path))
		oas = append(oas, fmt.Sprintf("    %s:", strings.ToLower(r.Method)))
		if r.Public {
			oas = append(oas, "      security: []")
		} else {
			oas = append(oas, "      security:")
			oas = append(oas, fmt.Sprintf("        - OAuth2: [%s]", r.Permission))
		}
	}

	// 6. Lifecycle projection
	lifecycle := fmt.Sprintf("Module: %s\nDependsOn: %s", m.ID, strings.Join(m.DependsOn, ", "))

	// 7. Profile projection
	profile := fmt.Sprintf("Profiles:\n  API: routes=%d\n  Worker: jobs=0\n  Migrate: migrations=true", len(m.Routes))

	// 8. Test projection
	var tests []string
	for _, r := range m.Routes {
		tests = append(tests, fmt.Sprintf("TestRoute: %s %s (public=%t)", r.Method, r.Path, r.Public))
	}

	// 9. Doc projection
	var docs []string
	docs = append(docs, "| Method | Path | Permission | Public |")
	docs = append(docs, "|---|---|---|---|")
	for _, r := range m.Routes {
		docs = append(docs, fmt.Sprintf("| %s | %s | %s | %t |", r.Method, r.Path, r.Permission, r.Public))
	}

	return Projections{
		Routes:      strings.Join(routes, "\n"),
		Permissions: strings.Join(perms, "\n"),
		Resources:   strings.Join(res, "\n"),
		Schema:      strings.Join(schemas, "\n"),
		OpenAPI:     strings.Join(oas, "\n"),
		Lifecycle:   lifecycle,
		Profile:     profile,
		Test:        strings.Join(tests, "\n"),
		Doc:         strings.Join(docs, "\n"),
	}
}

// String combines all projections into a single stable text snapshot for golden diffing.
func (p Projections) String() string {
	var b strings.Builder
	b.WriteString("=== Routes ===\n" + p.Routes + "\n\n")
	b.WriteString("=== Permissions ===\n" + p.Permissions + "\n\n")
	b.WriteString("=== Resources ===\n" + p.Resources + "\n\n")
	b.WriteString("=== Schema ===\n" + p.Schema + "\n\n")
	b.WriteString("=== OpenAPI ===\n" + p.OpenAPI + "\n\n")
	b.WriteString("=== Lifecycle ===\n" + p.Lifecycle + "\n\n")
	b.WriteString("=== Profile ===\n" + p.Profile + "\n\n")
	b.WriteString("=== Test ===\n" + p.Test + "\n\n")
	b.WriteString("=== Doc ===\n" + p.Doc + "\n")
	return b.String()
}

// LintManifest validates a Manifest for duplicate identities or omitted projection mappings.
func LintManifest(m Manifest) []string {
	var violations []string

	// Check duplicate routes
	seenRoutes := map[string]bool{}
	for _, r := range m.Routes {
		key := r.Method + " " + r.Path
		if seenRoutes[key] {
			violations = append(violations, fmt.Sprintf("duplicate route identity: %s", key))
		}
		seenRoutes[key] = true
	}

	// Check duplicate permissions
	seenPerms := map[string]bool{}
	for _, p := range m.Permissions {
		if seenPerms[p.Key] {
			violations = append(violations, fmt.Sprintf("duplicate permission identity: %s", p.Key))
		}
		seenPerms[p.Key] = true
	}

	// Check duplicate resources
	seenRes := map[string]bool{}
	for _, r := range m.Resources {
		if seenRes[r.Key] {
			violations = append(violations, fmt.Sprintf("duplicate resource identity: %s", r.Key))
		}
		seenRes[r.Key] = true
	}

	// Check omitted projections: every permission referenced by a route must be declared in Permissions catalog
	for _, r := range m.Routes {
		if !r.Public && r.Permission != "" {
			if !seenPerms[r.Permission] {
				violations = append(violations, fmt.Sprintf("omitted projection: route %s %s references undeclared permission %q", r.Method, r.Path, r.Permission))
			}
		}
	}

	return violations
}
