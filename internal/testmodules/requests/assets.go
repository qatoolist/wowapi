package requests

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationsEmbed embed.FS

//go:embed seeds/*.yaml
var seedsEmbed embed.FS

//go:embed openapi.json
var openapiFragment []byte

// migrationsFS is rooted at the .sql files (goose expects the migration files
// at the FS root, as the kernel's own migrations package does). Sub strips the
// embed's "migrations/" directory prefix.
var migrationsFS = mustSub(migrationsEmbed, "migrations")

// seedsFS holds the seed YAML; the seed loader walks recursively so the subdir
// is fine, but Sub keeps it symmetric with migrations.
var seedsFS = mustSub(seedsEmbed, "seeds")

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
