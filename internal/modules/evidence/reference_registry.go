package evidence

import (
	"context"
	"github.com/jackc/pgx/v5"
)

// ReferencedEvidenceRegistryEntriesTx supplies the Evidence-owned usage
// projection. Evidence records currently have no registry-token assignment
// field; the required Base unknown entry remains available for rough capture.
func ReferencedEvidenceRegistryEntriesTx(context.Context, pgx.Tx) ([]string, error) {
	return []string{"unknown"}, nil
}
