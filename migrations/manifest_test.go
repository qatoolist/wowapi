package migrations_test

import (
	"io/fs"
	"testing"

	"github.com/qatoolist/wowapi/kernel/migration"
	"github.com/qatoolist/wowapi/migrations"
)

// TestKernelMigrationsHaveManifests verifies that every kernel migration at or
// above the DATA-09 manifest boundary carries a valid +wowapi:manifest block.
// Pre-boundary migrations are exempt; if they contain a block it is still
// validated.
func TestKernelMigrationsHaveManifests(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isMigrationFile(path) {
			return nil
		}
		body, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		ver, err := migration.MigrationVersion(path)
		if err != nil {
			return err
		}
		m, parseErr := migration.ParseManifest(path, string(body))
		if ver >= migration.ManifestRequiredVersion {
			if parseErr != nil {
				t.Errorf("%s: %v", path, parseErr)
				return nil
			}
			if m == nil {
				t.Errorf("%s: missing +wowapi:manifest block", path)
				return nil
			}
			if err := m.Validate(); err != nil {
				t.Errorf("%s: %v", path, err)
			}
		} else if parseErr == nil && m != nil {
			// Opt-in manifest on older migrations is still validated.
			if err := m.Validate(); err != nil {
				t.Errorf("%s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
}

func isMigrationFile(path string) bool {
	return len(path) > 4 && path[len(path)-4:] == ".sql"
}
