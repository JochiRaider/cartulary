package tasksdecisions

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

const (
	taskRequestsBundlePath = "data/task_requests.ndjson"
	decisionsBundlePath    = "data/decisions.ndjson"
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	specs := []struct {
		path  string
		query string
	}{
		{taskRequestsBundlePath, `SELECT to_jsonb(t) FROM task_requests t WHERE incident_id = $1 ORDER BY record_id`},
		{decisionsBundlePath, `SELECT to_jsonb(t) FROM decisions t WHERE incident_id = $1 ORDER BY record_id`},
	}
	files := make([]incidentportability.File, 0, len(specs))
	for _, spec := range specs {
		file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, spec.path, spec.query)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

type preparedTasksDecisionsImport struct {
	tasks     []portableTaskRequest
	decisions []portableDecision
}

type portableTaskRequest struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	Title               string
	Status              string
	PortableOwnerUserID *uuid.UUID
	Priority            string
	TaskKind            string
	Workstream          *string
	DueAt               *time.Time
	RequesterPartyText  *string
	RequesterPartyID    *uuid.UUID
	BlockedReason       *string
	CompletedAt         *time.Time
	ExternalTicketRef   *string
	ClosureSummary      *string
	DecisionRecordID    *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type portableDecision struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	Summary             string
	Status              string
	PortableOwnerUserID uuid.UUID
	DecisionType        string
	DecidedAt           time.Time
	Rationale           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func prepareTasksDecisionsImport(
	bundle interface{ File(string) ([]byte, bool) },
	importContext sourceport.ImportContext,
) (preparedTasksDecisionsImport, error) {
	taskPayload, ok := bundle.File(taskRequestsBundlePath)
	if !ok {
		return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	decisionPayload, ok := bundle.File(decisionsBundlePath)
	if !ok {
		return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	taskRows, err := incidentportability.DecodeStrictNDJSONObjects(taskPayload, taskRequestsBundlePath)
	if err != nil {
		return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	decisionRows, err := incidentportability.DecodeStrictNDJSONObjects(decisionPayload, decisionsBundlePath)
	if err != nil {
		return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
	}
	prepared := preparedTasksDecisionsImport{
		tasks:     make([]portableTaskRequest, 0, len(taskRows)),
		decisions: make([]portableDecision, 0, len(decisionRows)),
	}
	seen := make(map[uuid.UUID]struct{}, len(taskRows)+len(decisionRows))
	for _, raw := range taskRows {
		row, rowErr := preparePortableTaskRequest(raw, importContext)
		if rowErr != nil {
			return preparedTasksDecisionsImport{}, rowErr
		}
		if _, duplicate := seen[row.RecordID]; duplicate {
			return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
		seen[row.RecordID] = struct{}{}
		prepared.tasks = append(prepared.tasks, row)
	}
	for _, raw := range decisionRows {
		row, rowErr := preparePortableDecision(raw, importContext)
		if rowErr != nil {
			return preparedTasksDecisionsImport{}, rowErr
		}
		if _, duplicate := seen[row.RecordID]; duplicate {
			return preparedTasksDecisionsImport{}, tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
		seen[row.RecordID] = struct{}{}
		prepared.decisions = append(prepared.decisions, row)
	}
	return prepared, nil
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

func exactPortableMembers(row map[string]any, required []string) bool {
	if len(row) != len(required) {
		return false
	}
	for _, member := range required {
		if _, ok := row[member]; !ok {
			return false
		}
	}
	return true
}

func canonicalPortableUUID(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func nullableCanonicalPortableUUID(value any) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := canonicalPortableUUID(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func admittedPortableActor(value any, importContext sourceport.ImportContext) (uuid.UUID, bool) {
	actorID, ok := canonicalPortableUUID(value)
	if !ok {
		return uuid.Nil, false
	}
	_, admitted := importContext.Actors.Lookup(actorID.String())
	return actorID, admitted
}

func nullableAdmittedPortableActor(value any, importContext sourceport.ImportContext) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	actorID, ok := admittedPortableActor(value, importContext)
	if !ok {
		return nil, false
	}
	return &actorID, true
}

func portableString(value any) (string, bool) {
	text, ok := value.(string)
	return text, ok
}

func nullablePortableString(value any) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := portableString(value)
	if !ok {
		return nil, false
	}
	return &text, true
}

func canonicalPortableTime(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, false
	}
	_, offset := parsed.Zone()
	canonicalText := text
	if strings.HasSuffix(canonicalText, "+00:00") {
		canonicalText = strings.TrimSuffix(canonicalText, "+00:00") + "Z"
	}
	if offset != 0 || parsed.UTC().Format(time.RFC3339Nano) != canonicalText {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func nullableCanonicalPortableTime(value any) (*time.Time, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := canonicalPortableTime(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func tasksDecisionsInvariantFailure(invariantID string) error {
	if !policy.IsPortabilityInvariant(invariantID) {
		invariantID = policy.InvariantEnvelope
	}
	return &sourceport.Failure{FamilyID: "tasks_decisions", InvariantID: invariantID}
}

func ImportIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files map[string][]byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	specs := []incidentportability.FixedImportSpec{
		{
			LogicalBundlePath: "data/task_requests.ndjson",
			AttributionTable:  "task_requests",
			StableIdentity:    []string{"record_id"},
			RequiredColumns:   []string{"record_id", "incident_id"},
			InsertSQL:         `INSERT INTO task_requests SELECT * FROM jsonb_populate_record(NULL::task_requests, $1::jsonb)`,
		},
		{
			LogicalBundlePath: "data/decisions.ndjson",
			AttributionTable:  "decisions",
			StableIdentity:    []string{"record_id"},
			RequiredColumns:   []string{"record_id", "incident_id"},
			InsertSQL:         `INSERT INTO decisions SELECT * FROM jsonb_populate_record(NULL::decisions, $1::jsonb)`,
		},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}

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
