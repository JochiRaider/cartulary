package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type commandStore struct {
	transactions           TransactionRunner
	appender               *Appender
	envelopes              RecordEnvelopePort
	authorization          CommandAuthorizer
	idempotency            IdempotencyPort
	projectionDependencies []ProjectionDependencyReader
	projections            ProjectionRefresher
	liveRecords            LiveRecordReader
	deleteRestoreSources   *DeleteRestoreSourceCatalog
	targetSemantics        *TargetSemanticsCatalog
	recordPublications     RecordPublicationPort
	importedAttributions   ImportedAttributionResolver
}

type ImportedAttributionResolver interface {
	ResolveImportedSourceActorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceTable string, sourceColumn string, sourceRowIDs []string) (map[string]string, error)
	// A nil map identifies native history. Imported history resolves every requested actor.
	ResolveImportedHistoryActorsTx(ctx context.Context, tx pgx.Tx, incidentID, changeSetID uuid.UUID, sourceActorIDs []string) (map[string]string, error)
}
