package entities

import (
	"context"
	"github.com/jackc/pgx/v5"
)

// ReferencedHostRegistryEntriesTx supplies the Entities-owned usage projection.
// Host records currently have no registry-token assignment field. Their rough
// capture descriptor is the required Base unknown entry.
func ReferencedHostRegistryEntriesTx(context.Context, pgx.Tx) ([]string, error) {
	return []string{"unknown"}, nil
}
