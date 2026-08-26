package incidentbundle

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func applyPreparedTasksDecisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedTasksDecisionsImport,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) error {
	for _, row := range prepared.tasks {
		if attributions != nil && row.PortableOwnerUserID != nil {
			if err := attributions.RecordImportedAttribution("task_requests", row.RecordID.String(), "owner_user_id", row.PortableOwnerUserID.String()); err != nil {
				return err
			}
		}
		var runtimeOwnerUserID *uuid.UUID
		if row.PortableOwnerUserID != nil {
			runtimeOwnerUserID = &actorUserID
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority, task_kind,
    workstream, due_at, requester_party_text, requester_party_id, blocked_reason,
    completed_at, external_ticket_ref, closure_summary, decision_record_id,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
`, row.RecordID, row.IncidentID, row.Title, row.Status, runtimeOwnerUserID, row.Priority,
			row.TaskKind, row.Workstream, row.DueAt, row.RequesterPartyText,
			row.RequesterPartyID, row.BlockedReason, row.CompletedAt,
			row.ExternalTicketRef, row.ClosureSummary, row.DecisionRecordID,
			row.CreatedAt, row.UpdatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
	}
	for _, row := range prepared.decisions {
		if attributions != nil {
			if err := attributions.RecordImportedAttribution("decisions", row.RecordID.String(), "owner_user_id", row.PortableOwnerUserID.String()); err != nil {
				return err
			}
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO decisions (
    record_id, incident_id, summary, status, owner_user_id, decision_type,
    decided_at, rationale, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, row.RecordID, row.IncidentID, row.Summary, row.Status, actorUserID,
			row.DecisionType, row.DecidedAt, row.Rationale, row.CreatedAt, row.UpdatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
	}
	return nil
}
