package migrations

import (
	"embed"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	EmbeddedPath   = "."
	RepositoryPath = "db/migrations"

	// PreRecordEnvelopeVersion is the last schema version before records became
	// the shared first-class record envelope.
	PreRecordEnvelopeVersion = "5"

	// PreAssessmentsCore02Version is the last schema version before legacy
	// compromise assessments were migrated to the Core 02 assessment vocabulary.
	PreAssessmentsCore02Version = "7"
)

// Files embeds the authored SQL migrations so repo-owned callers do not depend on cwd.
//
//go:embed *.sql
var Files embed.FS

func Source() postgres.MigrationSource {
	return postgres.NewEmbeddedMigrationSource(Files, EmbeddedPath, RepositoryPath)
}
