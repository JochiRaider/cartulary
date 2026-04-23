package migrations

import (
	"embed"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	EmbeddedPath   = "."
	RepositoryPath = "db/migrations"
)

// Files embeds the authored SQL migrations so repo-owned callers do not depend on cwd.
//
//go:embed *.sql
var Files embed.FS

func Source() postgres.MigrationSource {
	return postgres.NewEmbeddedMigrationSource(Files, EmbeddedPath, RepositoryPath)
}
