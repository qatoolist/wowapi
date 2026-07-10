// Package seeds loads a module's declarative catalog seeds (permissions, roles,
// resource types, relationship types) from embedded YAML and syncs them
// idempotently into the global catalogs at boot. Seeds touch ONLY global
// catalogs — never tenant data (blueprint 06 §2 lifecycle SeedSync). Because
// the catalogs back authorization, they are written with platform privilege
// (app_platform / owner), never as app_rt (SEC-13/D-0026).
package seeds

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Bundle is the parsed, merged seed catalog for one or more modules.
type Bundle struct {
	Permissions       []PermissionSeed       `yaml:"permissions"`
	Roles             []RoleSeed             `yaml:"roles"`
	ResourceTypes     []ResourceTypeSeed     `yaml:"resource_types"`
	RelationshipTypes []RelationshipTypeSeed `yaml:"relationship_types"`
}

// PermissionSeed declares a permission in the catalog.
type PermissionSeed struct {
	Key         string `yaml:"key"`
	Description string `yaml:"description"`
	Sensitive   bool   `yaml:"sensitive"`
	// GrantedVia declares the ReBAC rule fed into the authz registry.
	GrantedVia string `yaml:"granted_via"`
	// StepUp declares that this permission requires an elevated authentication
	// factor (MFA): an otherwise-allowed decision becomes a step-up challenge
	// when the actor's AMR carries no strong factor (roadmap S3). Propagated to
	// authz.Permission.StepUp at boot and persisted to permissions.step_up.
	StepUp bool `yaml:"step_up"`
}

// RoleSeed declares a platform-template role and the permissions it grants.
type RoleSeed struct {
	Key         string   `yaml:"key"`
	Name        string   `yaml:"name"`
	Permissions []string `yaml:"permissions"`
}

// ResourceTypeSeed declares a resource type.
type ResourceTypeSeed struct {
	Key         string `yaml:"key"`
	Description string `yaml:"description"`
}

// RelationshipTypeSeed declares a relationship type.
type RelationshipTypeSeed struct {
	Key         string `yaml:"key"`
	SubjectKind string `yaml:"subject_kind"`
	ObjectKind  string `yaml:"object_kind"`
	Cardinality string `yaml:"cardinality"`
	Description string `yaml:"description"`
}

// Load parses every *.yaml (and *.yml) file in src into one merged Bundle,
// strict-decoding so a typo (unknown key) fails the load. module is the owning
// module name: every declared key must be prefixed with "<module>." so a module
// cannot seed another module's catalog entries.
func Load(src fs.FS, module string) (Bundle, error) {
	var b Bundle
	err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || (!strings.HasSuffix(p, ".yaml") && !strings.HasSuffix(p, ".yml")) {
			return nil
		}
		data, rerr := fs.ReadFile(src, p)
		if rerr != nil {
			return rerr
		}
		var one Bundle
		dec := yaml.NewDecoder(strings.NewReader(string(data)))
		dec.KnownFields(true) // unknown keys are errors (typo defense)
		if derr := dec.Decode(&one); derr != nil && derr.Error() != "EOF" {
			return kerr.E(kerr.KindInternal, "invalid_seed", fmt.Sprintf("seed %s: %v", p, derr))
		}
		b.merge(one)
		return nil
	})
	if err != nil {
		return Bundle{}, kerr.Wrapf(err, "seeds.Load", "load seeds for %s", module)
	}
	if err := b.validate(module); err != nil {
		return Bundle{}, err
	}
	return b, nil
}

func (b *Bundle) merge(o Bundle) {
	b.Permissions = append(b.Permissions, o.Permissions...)
	b.Roles = append(b.Roles, o.Roles...)
	b.ResourceTypes = append(b.ResourceTypes, o.ResourceTypes...)
	b.RelationshipTypes = append(b.RelationshipTypes, o.RelationshipTypes...)
}

// validate enforces the module-prefix ownership rule and rejects empties. A
// module may only seed — and grant, and wire granted_via to — keys it owns.
// The role grant-list and granted_via are the sharpest edges: unchecked, a
// module could grant itself another module's permission or hijack another
// module's relationship as an auto-grant (review findings SEC-32/SEC-34).
func (b Bundle) validate(module string) error {
	prefix := module + "."
	var errs []string
	check := func(kind, key string) {
		if key == "" {
			errs = append(errs, kind+": empty key")
			return
		}
		if !strings.HasPrefix(key, prefix) {
			errs = append(errs, fmt.Sprintf("%s %q: module %q may only seed keys prefixed %q", kind, key, module, prefix))
		}
	}
	// The set of relationship types this bundle declares — granted_via must
	// reference one of them (an owned, existing type), never a dangling or
	// foreign type.
	relTypes := map[string]bool{}
	for _, rt := range b.RelationshipTypes {
		relTypes[rt.Key] = true
	}

	for _, p := range b.Permissions {
		check("permission", p.Key)
		if p.GrantedVia != "" {
			check("permission granted_via", p.GrantedVia)
			if !relTypes[p.GrantedVia] {
				errs = append(errs, fmt.Sprintf("permission %q: granted_via %q is not a relationship type declared by this module", p.Key, p.GrantedVia))
			}
		}
	}
	for _, r := range b.Roles {
		check("role", r.Key)
		for _, perm := range r.Permissions {
			// A role may only grant permissions its OWN module owns (SEC-32).
			check("role grant", perm)
		}
	}
	for _, rt := range b.ResourceTypes {
		check("resource_type", rt.Key)
	}
	for _, rt := range b.RelationshipTypes {
		check("relationship_type", rt.Key)
	}
	if len(errs) > 0 {
		sort.Strings(errs)
		return kerr.E(kerr.KindInternal, "invalid_seed",
			"seed ownership violations: "+strings.Join(errs, "; "))
	}
	return nil
}

// SpineInvalidator drops an in-process authorization cache after a seed
// (authorization-spine) write commits. *authz.CachingStore satisfies it via its
// InvalidateAll method — declared here as a narrow local interface so this base
// package stays free of an authz import. A seed sync rewrites GLOBAL platform
// roles and their role_permissions, which any tenant's actors may hold and which
// the cache pre-joins into ActiveAssignments, so the WHOLE cache is dropped (not
// one tenant) — see CachingStore.InvalidateAll.
type SpineInvalidator interface{ InvalidateAll() }

// Sync upserts the bundle's catalog rows idempotently. It must run on a
// platform-privileged connection (the global catalogs are not app_rt-writable).
// Running twice is a no-op diff (ON CONFLICT DO UPDATE); tenant data is never
// touched.
//
// invalidators, if any, are invoked AFTER every write succeeds so an in-process
// authz cache does not serve stale role/permission grants past the sync (CA-2).
// Pass the kernel's live cache (Kernel.AuthzCache) when it is non-nil; pass
// nothing when caching is off — the default — and Sync behaves exactly as before.
func Sync(ctx context.Context, db database.DBTX, b Bundle, invalidators ...SpineInvalidator) error {
	for _, p := range b.Permissions {
		if _, err := db.Exec(ctx,
			`INSERT INTO permissions (key, module, description, sensitive, step_up)
                  VALUES ($1, $2, $3, $4, $5)
             ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description, sensitive = EXCLUDED.sensitive,
                   step_up = EXCLUDED.step_up`,
			p.Key, moduleOf(p.Key), p.Description, p.Sensitive, p.StepUp); err != nil {
			return kerr.Wrapf(err, "seeds.Sync", "upsert permission %s", p.Key)
		}
	}
	for _, rt := range b.ResourceTypes {
		if _, err := db.Exec(ctx,
			`INSERT INTO resource_types (key, module, description)
                  VALUES ($1, $2, $3)
             ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description`,
			rt.Key, moduleOf(rt.Key), rt.Description); err != nil {
			return kerr.Wrapf(err, "seeds.Sync", "upsert resource_type %s", rt.Key)
		}
	}
	for _, rt := range b.RelationshipTypes {
		card := rt.Cardinality
		if card == "" {
			card = "many"
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO relationship_types (key, module, subject_kind, object_kind, cardinality, description)
                  VALUES ($1, $2, $3, $4, $5, $6)
             ON CONFLICT (key) DO UPDATE SET subject_kind = EXCLUDED.subject_kind,
                   object_kind = EXCLUDED.object_kind, cardinality = EXCLUDED.cardinality,
                   description = EXCLUDED.description`,
			rt.Key, moduleOf(rt.Key), rt.SubjectKind, rt.ObjectKind, card, rt.Description); err != nil {
			return kerr.Wrapf(err, "seeds.Sync", "upsert relationship_type %s", rt.Key)
		}
	}
	// Roles are platform templates (tenant_id NULL). Upsert the role, then its
	// permission grants. A stable id is derived from the key so reseeding
	// updates the same row (roles_key unique index is on (coalesce(tenant),key)).
	for _, r := range b.Roles {
		var roleID string
		if err := db.QueryRow(ctx,
			`INSERT INTO roles (id, tenant_id, key, name, is_system, created_by)
                  VALUES (gen_random_uuid(), NULL, $1, $2, true, '00000000-0000-0000-0000-000000000000')
             ON CONFLICT (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key)
             DO UPDATE SET name = EXCLUDED.name
             RETURNING id`,
			r.Key, r.Name).Scan(&roleID); err != nil {
			return kerr.Wrapf(err, "seeds.Sync", "upsert role %s", r.Key)
		}
		for _, perm := range r.Permissions {
			if _, err := db.Exec(ctx,
				`INSERT INTO role_permissions (role_id, permission_key) VALUES ($1, $2)
                 ON CONFLICT (role_id, permission_key) DO NOTHING`,
				roleID, perm); err != nil {
				return kerr.Wrapf(err, "seeds.Sync", "grant %s to role %s", perm, r.Key)
			}
		}
		// Reconcile: prune grants no longer in the seed so a demoted role
		// cannot keep stale permissions across redeploys (least-privilege,
		// review finding ARCH-47). Seeds are the source of truth for a
		// platform role's grant set.
		if _, err := db.Exec(ctx,
			`DELETE FROM role_permissions
                  WHERE role_id = $1 AND NOT (permission_key = ANY($2))`,
			roleID, r.Permissions); err != nil {
			return kerr.Wrapf(err, "seeds.Sync", "prune stale grants on role %s", r.Key)
		}
	}
	// The writes above changed the authorization spine (platform roles and their
	// permission grants). Drop any in-process authz cache broadly so a role's
	// changed permission set takes effect immediately rather than after the TTL
	// (CA-2). No-op when caching is off (no invalidator passed).
	for _, inv := range invalidators {
		if inv != nil {
			inv.InvalidateAll()
		}
	}
	return nil
}

func moduleOf(key string) string {
	if i := strings.IndexByte(key, '.'); i > 0 {
		return key[:i]
	}
	return key
}
