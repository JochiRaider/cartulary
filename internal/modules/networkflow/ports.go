package networkflow

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
)

type IncidentLockPort interface {
	LockIncidentTx(context.Context, pgx.Tx, uuid.UUID) (bool, error)
}

type AdministrativeAuditPort interface {
	AppendNetworkFlowEventTx(context.Context, pgx.Tx, *uuid.UUID, *uuid.UUID, string, *string, *string, any, any) error
}

type IndicatorParticipationPort interface {
	GetActiveIndicatorParticipant(context.Context, uuid.UUID, uuid.UUID) (indicators.IndicatorReference, error)
	GetActiveIndicatorParticipantTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (indicators.IndicatorReference, error)
	FindOrCreateIndicatorParticipantTx(context.Context, pgx.Tx, indicators.IndicatorFindOrCreateParticipantCommand) (indicators.IndicatorFindOrCreateParticipantResult, error)
}
