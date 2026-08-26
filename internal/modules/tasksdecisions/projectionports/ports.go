package projectionports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskDerivedFact struct {
	RecordID uuid.UUID
	Value    map[string]any
}

type DecisionDerivedFact struct {
	RecordID uuid.UUID
	Value    map[string]any
}

type MutationRows interface {
	RefreshTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RefreshDecisionTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadDecisionTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type ReportingReader interface {
	CollectTaskDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]TaskDerivedFact, error)
	CollectDecisionDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DecisionDerivedFact, error)
}
