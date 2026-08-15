package recoveryassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresrestore"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// NewGraphProjectionRestoreParticipant composes the current code-backed empty
// source registry with the narrow borrowed-Postgres restore writer. It does not
// activate the retained Store or any ordinary runtime producer.
func NewGraphProjectionRestoreParticipant(db postgres.DB) (restorecontract.GraphProjectionParticipant, error) {
	writer, err := postgresrestore.New(db)
	if err != nil {
		return nil, err
	}
	return graphprojection.NewRestoreService(
		writer,
		graphprojection.CurrentRestoreSourceRegistry(),
		graphprojection.RestoreServiceOptions{},
	)
}
