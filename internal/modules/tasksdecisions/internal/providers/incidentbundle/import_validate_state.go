package incidentbundle

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

func (p preparedTasksDecisionsImport) validateTaskState(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext) error {
	rows, err := tx.Query(ctx, `SELECT record_id,incident_id,title,status,owner_user_id,priority,task_kind,workstream,due_at,requester_party_text,requester_party_id,blocked_reason,completed_at,external_ticket_ref,closure_summary,decision_record_id,created_at,updated_at FROM task_requests WHERE incident_id=$1 ORDER BY record_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.walkTasks(ctx, input, func(expected portableTaskRequest) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
		}
		var actual portableTaskRequest
		var title, priority, kind *string
		if err := rows.Scan(&actual.RecordID, &actual.IncidentID, &title, &actual.Status, &actual.PortableOwnerUserID,
			&priority, &kind, &actual.Workstream, &actual.DueAt, &actual.RequesterPartyText,
			&actual.RequesterPartyID, &actual.BlockedReason, &actual.CompletedAt, &actual.ExternalTicketRef,
			&actual.ClosureSummary, &actual.DecisionRecordID, &actual.CreatedAt, &actual.UpdatedAt); err != nil {
			return err
		}
		if title == nil || priority == nil || kind == nil {
			return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
		}
		actual.Title, actual.Priority, actual.TaskKind = *title, *priority, *kind
		actual.CreatedAt, actual.UpdatedAt = actual.CreatedAt.UTC(), actual.UpdatedAt.UTC()
		if actual.DueAt != nil {
			*actual.DueAt = actual.DueAt.UTC()
		}
		if actual.CompletedAt != nil {
			*actual.CompletedAt = actual.CompletedAt.UTC()
		}
		if expected.PortableOwnerUserID != nil {
			expected.PortableOwnerUserID = &p.actorID
		}
		if !reflect.DeepEqual(actual, expected) {
			return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
	}
	return rows.Err()
}

func (p preparedTasksDecisionsImport) validateDecisionState(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext) error {
	rows, err := tx.Query(ctx, `SELECT record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale,created_at,updated_at FROM decisions WHERE incident_id=$1 ORDER BY record_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.walkDecisions(ctx, input, func(expected portableDecision) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
		}
		var actual portableDecision
		var summary, kind, rationale *string
		var owner *uuid.UUID
		var decided *time.Time
		if err := rows.Scan(&actual.RecordID, &actual.IncidentID, &summary, &actual.Status, &owner,
			&kind, &decided, &rationale, &actual.CreatedAt, &actual.UpdatedAt); err != nil {
			return err
		}
		if summary == nil || kind == nil || rationale == nil || owner == nil || decided == nil {
			return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
		}
		actual.Summary, actual.DecisionType, actual.Rationale = *summary, *kind, *rationale
		actual.PortableOwnerUserID, actual.DecidedAt = *owner, *decided
		actual.CreatedAt, actual.UpdatedAt, actual.DecidedAt = actual.CreatedAt.UTC(), actual.UpdatedAt.UTC(), actual.DecidedAt.UTC()
		expected.PortableOwnerUserID = p.actorID
		if !reflect.DeepEqual(actual, expected) {
			return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return tasksDecisionsInvariantFailure(policy.InvariantSourceState)
	}
	return rows.Err()
}
