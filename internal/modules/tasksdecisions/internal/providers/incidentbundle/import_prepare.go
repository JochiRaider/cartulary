package incidentbundle

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
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
	previous := ""
	admit := func(id uuid.UUID) error {
		if id.String() <= previous {
			return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
		previous = id.String()
		return nil
	}
	if err := prepared.walkTasks(ctx, importContext, func(row portableTaskRequest) error { return admit(row.RecordID) }); err != nil {
		return preparedTasksDecisionsImport{}, err
	}
	previous = ""
	if err := prepared.walkDecisions(ctx, importContext, func(row portableDecision) error { return admit(row.RecordID) }); err != nil {
		return preparedTasksDecisionsImport{}, err
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
	title, titleOK := portableString(raw["title"])
	status, statusOK := portableString(raw["status"])
	owner, ownerOK := nullableAdmittedPortableActor(raw["owner_user_id"], importContext)
	priority, priorityOK := portableString(raw["priority"])
	taskKind, taskKindOK := portableString(raw["task_kind"])
	createdAt, createdOK := canonicalPortableTime(raw["created_at"])
	updatedAt, updatedOK := canonicalPortableTime(raw["updated_at"])
	if !titleOK || strings.TrimSpace(title) == "" || !statusOK || !ownerOK ||
		!priorityOK || !taskKindOK || !createdOK || !updatedOK || updatedAt.Before(createdAt) {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
	}
	if !policy.ValidTaskStatus(status) {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
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
	requesterPartyID, ok := nullableCanonicalPortableUUID(raw["requester_party_id"])
	if !ok {
		return portableTaskRequest{}, tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
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
	summary, summaryOK := portableString(raw["summary"])
	status, statusOK := portableString(raw["status"])
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
	if !policy.ValidDecisionStatus(status) {
		return portableDecision{}, tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
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
