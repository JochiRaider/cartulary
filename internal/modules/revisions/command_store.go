package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type commandStore struct {
	transactions            TransactionRunner
	appender                *Appender
	envelopes               RecordEnvelopePort
	authorization           CommandAuthorizer
	idempotency             IdempotencyPort
	projections             ProjectionServices
	deleteRestoreSources    *DeleteRestoreSourceCatalog
	rowRollbackProviders    *RowProviderCatalog
	nonRowRollbackProviders *NonRowProviderCatalog
}

type ImportedAttributionResolver interface {
	ResolveImportedSourceActorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceTable string, sourceColumn string, sourceRowIDs []string) (map[string]string, error)
}
