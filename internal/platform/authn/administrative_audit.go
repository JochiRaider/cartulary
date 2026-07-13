package authn

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	beforeJSON, err := administrativeAuditJSONOrNil(before)
	if err != nil {
		return err
	}
	afterJSON, err := administrativeAuditJSONOrNil(after)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (
    actor_user_id, incident_id, event_source, event_kind,
    client_txn_id, request_id, before_json, after_json
)
VALUES ($1, $2, 'network_flow', $3, $4, $5, $6, $7)
`, actorUserID, incidentID, eventKind, clientTxnID, requestID, beforeJSON, afterJSON); err != nil {
		return fmt.Errorf("append Network Flow administrative audit event: %w", err)
	}
	return nil
}

func administrativeAuditJSONOrNil(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal administrative audit payload: %w", err)
	}
	return encoded, nil
}
