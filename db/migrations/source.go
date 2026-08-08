package migrations

import (
	"embed"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

const (
	EmbeddedPath   = "."
	RepositoryPath = "db/migrations"

	LineageID       = "cartulary.prod_ddl_rebaseline.v1"
	LineageBoundary = "prod_ddl_rebaseline_v1"
)

// Files embeds the authored SQL migrations so repo-owned callers do not depend on cwd.
//
//go:embed *.sql
var Files embed.FS

func Source() database_migrations.MigrationSource {
	source := database_migrations.NewEmbeddedMigrationSource(Files, EmbeddedPath, RepositoryPath)
	source.ExpectedLineageID = LineageID
	source.ExpectedLineageBoundary = LineageBoundary
	return source
}
