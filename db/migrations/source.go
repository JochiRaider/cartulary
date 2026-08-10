package migrations

import (
	"embed"
	"sync"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

const embeddedPath = "."

//go:embed *.sql
var files embed.FS

var (
	sourceOnce sync.Once
	source     *database_migrations.Source
	sourceErr  error
)

func Source() (*database_migrations.Source, error) {
	sourceOnce.Do(func() {
		source, sourceErr = database_migrations.BuildCanonicalEmbedded(files, embeddedPath)
	})
	return source, sourceErr
}
