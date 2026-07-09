package migrations

import (
	"embed"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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

func Source() postgres.MigrationSource {
	source := postgres.NewEmbeddedMigrationSource(Files, EmbeddedPath, RepositoryPath)
	source.ExpectedLineageID = LineageID
	source.ExpectedLineageBoundary = LineageBoundary
	return source
}
