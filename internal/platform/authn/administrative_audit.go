package authn

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
)

// AdministrativeAuditAppender is the provider-owned transaction participant
// for the deployment administrative audit stream.
type AdministrativeAuditAppender struct{}

func NewAdministrativeAuditAppender() *AdministrativeAuditAppender {
	return &AdministrativeAuditAppender{}
}

func (*AdministrativeAuditAppender) AppendNetworkFlowEventTx(
	ctx context.Context,
	tx pgx.Tx,
	actorUserID *uuid.UUID,
	incidentID *uuid.UUID,
	eventKind string,
	clientTxnID *string,
	requestID *string,
	before any,
	after any,
) error {
	_, err := administrativeaudit.AppendRawTx(ctx, tx, administrativeaudit.RawEvent{
		ActorUserID: actorUserID,
		IncidentID:  incidentID,
		EventSource: "network_flow",
		EventKind:   eventKind,
		ClientTxnID: clientTxnID,
		RequestID:   requestID,
		Before:      before,
		After:       after,
		OccurredAt:  time.Now().UTC(),
	})
	return err
}
