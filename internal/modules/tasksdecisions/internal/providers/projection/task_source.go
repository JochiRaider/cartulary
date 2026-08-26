package projection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

type TaskRequestSource struct{}

func NewTaskRequestSource() *TaskRequestSource { return &TaskRequestSource{} }

func (*TaskRequestSource) LoadTaskRequestProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (taskprojection.TaskRequestProjectionInput, bool, error) {
	input, err := scanTaskRequestProjectionInput(tx.QueryRow(ctx, taskRequestProjectionSourceSQL+`
 WHERE t.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return taskprojection.TaskRequestProjectionInput{}, false, nil
	}
	if err != nil {
		return taskprojection.TaskRequestProjectionInput{}, false, fmt.Errorf("load Task-request projection input: %w", err)
	}
	return input, true, nil
}

func (*TaskRequestSource) ListTaskRequestProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (taskprojection.TaskRequestProjectionInputPage, error) {
	if limit < 1 || limit > 1000 {
		return taskprojection.TaskRequestProjectionInputPage{}, fmt.Errorf("task-request projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, taskRequestProjectionSourceSQL+`
 WHERE t.incident_id = $1
   AND ($2::uuid IS NULL OR t.record_id > $2)
 ORDER BY t.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return taskprojection.TaskRequestProjectionInputPage{}, fmt.Errorf("list task-request projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]taskprojection.TaskRequestProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanTaskRequestProjectionInput(rows)
		if scanErr != nil {
			return taskprojection.TaskRequestProjectionInputPage{}, fmt.Errorf("scan Task-request projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return taskprojection.TaskRequestProjectionInputPage{}, fmt.Errorf("iterate Task-request projection inputs: %w", err)
	}
	page := taskprojection.TaskRequestProjectionInputPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanTaskRequestProjectionInput(scanner interface{ Scan(...any) error }) (taskprojection.TaskRequestProjectionInput, error) {
	var (
		input              taskprojection.TaskRequestProjectionInput
		title              pgtype.Text
		ownerUserID        pgtype.UUID
		priority           pgtype.Text
		taskKind           pgtype.Text
		workstream         pgtype.Text
		dueAt              pgtype.Timestamptz
		requesterPartyText pgtype.Text
		requesterPartyID   pgtype.UUID
		blockedReason      pgtype.Text
		completedAt        pgtype.Timestamptz
		externalTicketRef  pgtype.Text
		closureSummary     pgtype.Text
		decisionRecordID   pgtype.UUID
		linkedRecordCount  int32
	)
	if err := scanner.Scan(
		&input.RecordID, &input.IncidentID, &input.RowVersion, &title,
		&input.Status, &ownerUserID, &priority, &taskKind, &workstream,
		&dueAt, &requesterPartyText, &requesterPartyID, &blockedReason,
		&completedAt, &externalTicketRef, &closureSummary, &decisionRecordID,
		&linkedRecordCount, &input.UpdatedAt, &input.NoOwner,
	); err != nil {
		return taskprojection.TaskRequestProjectionInput{}, err
	}
	input.Title = taskTextPointer(title)
	input.OwnerUserID = taskUUIDPointer(ownerUserID)
	input.Priority = taskTextPointer(priority)
	input.TaskKind = taskTextPointer(taskKind)
	input.Workstream = taskTextPointer(workstream)
	input.DueAt = taskTimePointer(dueAt)
	input.RequesterPartyText = taskTextPointer(requesterPartyText)
	input.RequesterPartyID = taskUUIDPointer(requesterPartyID)
	input.BlockedReason = taskTextPointer(blockedReason)
	input.CompletedAt = taskTimePointer(completedAt)
	input.ExternalTicketRef = taskTextPointer(externalTicketRef)
	input.ClosureSummary = taskTextPointer(closureSummary)
	input.DecisionRecordID = taskUUIDPointer(decisionRecordID)
	input.LinkedRecordCount = int(linkedRecordCount)
	input.UpdatedAt = input.UpdatedAt.UTC()
	return input, nil
}

func taskTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func taskUUIDPointer(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	result := uuid.UUID(value.Bytes)
	return &result
}

func taskTimePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

const taskRequestProjectionSourceSQL = `
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.title,
    t.status,
    t.owner_user_id,
    t.priority,
    t.task_kind,
    t.workstream,
    t.due_at,
    t.requester_party_text,
    t.requester_party_id,
    t.blocked_reason,
    t.completed_at,
    t.external_ticket_ref,
    t.closure_summary,
    t.decision_record_id,
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 rl
         WHERE rl.incident_id = t.incident_id
           AND rl.src_record_id = t.record_id
           AND rl.link_type = 'references_record'
           AND rl.field_key = 'task.linked_record_ids'
    ),
    t.updated_at,
    t.owner_user_id IS NULL
  FROM task_requests t
  JOIN records r
    ON r.incident_id = t.incident_id
   AND r.record_id = t.record_id
   AND r.deleted_at IS NULL`
