package incidentbundle

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/textpolicy"
)

type preparedTasksDecisionsImport struct {
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
}

func prepareTasksDecisionsImport(ctx context.Context, bundle incidentportability.MemberReader, importContext sourceport.ImportContext) (preparedTasksDecisionsImport, error) {
	members, err := incidentportability.BindMembers(bundle, []string{taskRequestsBundlePath, decisionsBundlePath})
	if err != nil {
		return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	prepared := preparedTasksDecisionsImport{members: members, incidentID: importContext.IncidentID, actorID: importContext.ActorUserID}
	var selected error
	selectedRank := 0
	remember := func(err error) error {
		if err == nil {
			return nil
		}
		var failure *sourceport.Failure
		if !errors.As(err, &failure) || failure.FamilyID() != "tasks_decisions" {
			return err
		}
		rank := 0
		switch failure.InvariantID() {
		case policy.InvariantEnvelope:
			rank = 1
		case policy.InvariantLifecycle:
			rank = 4
		case policy.InvariantDependent:
			rank = 5
		case policy.InvariantReferences:
			rank = 6
		default:
			return err
		}
		if selected == nil || rank < selectedRank {
			selected, selectedRank = err, rank
		}
		return nil
	}
	for _, path := range []string{taskRequestsBundlePath, decisionsBundlePath} {
		previous := ""
		if err := prepared.walk(ctx, importContext, path, func(raw map[string]any) error {
			// Ordering is independent of lower-priority semantic admission.
			if id, ok := canonicalPortableUUID(raw["record_id"]); ok {
				if id.String() <= previous {
					if err := remember(tasksDecisionsInvariantFailure(policy.InvariantEnvelope)); err != nil {
						return err
					}
				}
				previous = id.String()
			}
			var err error
			if path == taskRequestsBundlePath {
				_, err = preparePortableTaskRequest(raw, importContext)
			} else {
				_, err = preparePortableDecision(raw, importContext)
			}
			return remember(err)
		}); err != nil {
			return preparedTasksDecisionsImport{}, err
		}
	}
	if selected != nil {
		return preparedTasksDecisionsImport{}, selected
	}
	return prepared, nil
}
func (p preparedTasksDecisionsImport) walk(ctx context.Context, importContext sourceport.ImportContext, path string, visit func(map[string]any) error) error {
	if p.members == nil || p.incidentID != importContext.IncidentID || p.actorID != importContext.ActorUserID {
		return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	err := incidentportability.WalkPath(ctx, p.members, path, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	return err
}
func (p preparedTasksDecisionsImport) walkTasks(ctx context.Context, importContext sourceport.ImportContext, visit func(portableTaskRequest) error) error {
	return p.walk(ctx, importContext, taskRequestsBundlePath, func(raw map[string]any) error {
		row, err := preparePortableTaskRequest(raw, importContext)
		if err != nil {
			return err
		}
		return visit(row)
	})
}
func (p preparedTasksDecisionsImport) walkDecisions(ctx context.Context, importContext sourceport.ImportContext, visit func(portableDecision) error) error {
	return p.walk(ctx, importContext, decisionsBundlePath, func(raw map[string]any) error {
		row, err := preparePortableDecision(raw, importContext)
		if err != nil {
			return err
		}
		return visit(row)
	})
}

func preparePortableTaskRequest(raw map[string]any, importContext sourceport.ImportContext) (portableTaskRequest, error) {
	required := []string{
		"record_id", "incident_id", "title", "status", "owner_user_id", "priority",
		"task_kind", "workstream", "due_at", "requester_party_text", "requester_party_id",
		"blocked_reason", "completed_at", "external_ticket_ref", "closure_summary",
		"decision_record_id", "created_at", "updated_at",
	}
	if !exactPortableMembers(raw, required) {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	recordID, ok := canonicalPortableUUID(raw["record_id"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	incidentID, ok := canonicalPortableUUID(raw["incident_id"])
	if !ok || incidentID != importContext.IncidentID {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	status, statusOK := portableString(raw["status"])
	if statusOK && !policy.ValidTaskStatus(status) {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
	}
	if err := textpolicy.ValidateStoredSource("task_request", raw); err != nil {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	title, titleOK := portableString(raw["title"])
	owner, ownerOK := nullableAdmittedPortableActor(raw["owner_user_id"], importContext)
	priority, priorityOK := portableString(raw["priority"])
	taskKind, taskKindOK := portableString(raw["task_kind"])
	createdAt, createdOK := canonicalPortableTime(raw["created_at"])
	updatedAt, updatedOK := canonicalPortableTime(raw["updated_at"])
	if !titleOK || strings.TrimSpace(title) == "" || !statusOK || !ownerOK ||
		!priorityOK || !taskKindOK || !createdOK || !updatedOK || updatedAt.Before(createdAt) {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	if !policy.ValidTaskPriority(priority) || !policy.ValidTaskKind(taskKind) {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	workstream, ok := nullablePortableString(raw["workstream"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	dueAt, ok := nullableCanonicalPortableTime(raw["due_at"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	requesterPartyText, ok := nullablePortableString(raw["requester_party_text"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	blockedReason, ok := nullablePortableString(raw["blocked_reason"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	completedAt, ok := nullableCanonicalPortableTime(raw["completed_at"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	externalTicketRef, ok := nullablePortableString(raw["external_ticket_ref"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	closureSummary, ok := nullablePortableString(raw["closure_summary"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	state := policy.TaskLifecycleState{Status: status, CreatedAt: createdAt}
	if owner != nil {
		state.OwnerUserID = sql.NullString{String: owner.String(), Valid: true}
	}
	if blockedReason != nil {
		state.BlockedReason = sql.NullString{String: *blockedReason, Valid: true}
	}
	if completedAt != nil {
		state.CompletedAt = sql.NullTime{Time: *completedAt, Valid: true}
	}
	if err := policy.ValidateTaskState(state); err != nil {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	requesterPartyID, ok := nullableCanonicalPortableUUID(raw["requester_party_id"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
	}
	decisionRecordID, ok := nullableCanonicalPortableUUID(raw["decision_record_id"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
	}
	return portableTaskRequest{
		RecordID: recordID, IncidentID: incidentID, Title: title, Status: status,
		PortableOwnerUserID: owner, Priority: priority, TaskKind: taskKind,
		Workstream: workstream, DueAt: dueAt, RequesterPartyText: requesterPartyText,
		RequesterPartyID: requesterPartyID, BlockedReason: blockedReason,
		CompletedAt: completedAt, ExternalTicketRef: externalTicketRef,
		ClosureSummary: closureSummary, DecisionRecordID: decisionRecordID,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func preparePortableDecision(raw map[string]any, importContext sourceport.ImportContext) (portableDecision, error) {
	required := []string{
		"record_id", "incident_id", "summary", "status", "owner_user_id",
		"decision_type", "decided_at", "rationale", "created_at", "updated_at",
	}
	if !exactPortableMembers(raw, required) {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	recordID, ok := canonicalPortableUUID(raw["record_id"])
	if !ok {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	incidentID, ok := canonicalPortableUUID(raw["incident_id"])
	if !ok || incidentID != importContext.IncidentID {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	status, statusOK := portableString(raw["status"])
	if statusOK && !policy.ValidDecisionStatus(status) {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
	}
	if err := textpolicy.ValidateStoredSource("decision", raw); err != nil {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	summary, summaryOK := portableString(raw["summary"])
	owner, ownerOK := admittedPortableActor(raw["owner_user_id"], importContext)
	decisionType, typeOK := portableString(raw["decision_type"])
	decidedAt, decidedOK := canonicalPortableTime(raw["decided_at"])
	rationale, rationaleOK := portableString(raw["rationale"])
	createdAt, createdOK := canonicalPortableTime(raw["created_at"])
	updatedAt, updatedOK := canonicalPortableTime(raw["updated_at"])
	if !summaryOK || strings.TrimSpace(summary) == "" || !statusOK || !ownerOK ||
		!typeOK || !decidedOK || !rationaleOK || strings.TrimSpace(rationale) == "" ||
		!createdOK || !updatedOK || updatedAt.Before(createdAt) {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	if !policy.ValidDecisionType(decisionType) {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	return portableDecision{
		RecordID: recordID, IncidentID: incidentID, Summary: summary, Status: status,
		PortableOwnerUserID: owner, DecisionType: decisionType, DecidedAt: decidedAt,
		Rationale: rationale, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func tasksDecisionsInvariantFailure(invariantID string) error {
	return tasksDecisionsSourceDescriptor().DeclaredFailure(invariantID)
}
