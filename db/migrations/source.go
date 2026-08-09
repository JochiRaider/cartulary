package migrations

import (
	"embed"
	"sync"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

const (
	EmbeddedPath = "."

	LineageID       = "cartulary.prod_ddl_rebaseline.v1"
	LineageBoundary = "prod_ddl_rebaseline_v1"
)

// Files embeds the authored SQL migrations so repo-owned callers do not depend on cwd.
//
//go:embed *.sql
var Files embed.FS

var (
	sourceOnce sync.Once
	source     database_migrations.Source
	sourceErr  error
)

func Source() (database_migrations.Source, error) {
	sourceOnce.Do(func() {
		source, sourceErr = database_migrations.NewSource(Files, EmbeddedPath, LineageID, LineageBoundary)
	})
	return source, sourceErr
}
