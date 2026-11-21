package incidentbundle

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

var (
	portableTaskColumns = []string{
		"record_id", "incident_id", "title", "status", "owner_user_id", "priority", "task_kind",
		"workstream", "due_at", "requester_party_text", "requester_party_id", "blocked_reason",
		"completed_at", "external_ticket_ref", "closure_summary", "decision_record_id",
		"created_at", "updated_at",
	}
	portableDecisionColumns = []string{
		"record_id", "incident_id", "summary", "status", "owner_user_id", "decision_type",
		"decided_at", "rationale", "created_at", "updated_at",
	}
)

// taskRowCodec stages one admitted task; COPY writes the importing actor as
// owner whenever a historical owner is present.
type taskRowCodec struct{ actor uuid.UUID }

func (taskRowCodec) Encode(row portableTaskRequest) ([]byte, error) { return json.Marshal(row) }
func (taskRowCodec) Decode(line []byte) (portableTaskRequest, error) {
	var row portableTaskRequest
	if err := json.Unmarshal(line, &row); err != nil {
		return portableTaskRequest{}, incidentportability.ErrPreparedStream
	}
	return row, nil
}
func (taskRowCodec) Columns() []string { return portableTaskColumns }
func (c taskRowCodec) Values(row portableTaskRequest) ([]any, error) {
	var owner *uuid.UUID
	if row.PortableOwnerUserID != nil {
		local := c.actor
		owner = &local
	}
	return []any{row.RecordID, row.IncidentID, row.Title, row.Status, owner, row.Priority,
		row.TaskKind, row.Workstream, row.DueAt, row.RequesterPartyText,
		row.RequesterPartyID, row.BlockedReason, row.CompletedAt,
		row.ExternalTicketRef, row.ClosureSummary, row.DecisionRecordID,
		row.CreatedAt, row.UpdatedAt}, nil
}

type decisionRowCodec struct{ actor uuid.UUID }

func (decisionRowCodec) Encode(row portableDecision) ([]byte, error) { return json.Marshal(row) }
func (decisionRowCodec) Decode(line []byte) (portableDecision, error) {
	var row portableDecision
	if err := json.Unmarshal(line, &row); err != nil {
		return portableDecision{}, incidentportability.ErrPreparedStream
	}
	return row, nil
}
func (decisionRowCodec) Columns() []string { return portableDecisionColumns }
func (c decisionRowCodec) Values(row portableDecision) ([]any, error) {
	return []any{row.RecordID, row.IncidentID, row.Summary, row.Status, c.actor,
		row.DecisionType, row.DecidedAt, row.Rationale, row.CreatedAt, row.UpdatedAt}, nil
}

// applyPreparedTasksDecisionsImportTx writes the staged rows with one COPY per
// family. Attribution was staged while preparing; the core flushes it in step 9.
func applyPreparedTasksDecisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedTasksDecisionsImport,
	importContext sourceport.ImportContext,
) error {
	if importContext.Attributions == nil {
		return tasksDecisionsInvariantFailure("tasks_decisions.attribution_exact")
	}
	if err := prepared.stagedFor(importContext); err != nil {
		return err
	}
	store := importContext.Preparation
	tasks, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.taskRows, taskRowCodec{actor: importContext.ActorUserID})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"task_requests"}, portableTaskColumns, tasks)
	if err := tasks.Finish(written, err); err != nil {
		return err
	}
	decisions, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.decisionRows, decisionRowCodec{actor: importContext.ActorUserID})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"decisions"}, portableDecisionColumns, decisions)
	return decisions.Finish(written, err)
}
