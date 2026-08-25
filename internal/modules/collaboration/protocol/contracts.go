// Package protocol owns Collaboration's semantic WebSocket transport and
// message contracts. It intentionally contains no route, hub, persistence, or
// process implementation.
package protocol

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	HeartbeatInterval       = 15 * time.Second
	HeartbeatTimeout        = 45 * time.Second
	PresenceTTL             = 45 * time.Second
	ResumeWindow            = 5 * time.Minute
	ResumeStatusReplayed    = "replayed"
	ResumeStatusResetNeeded = "reset_required"
	IncidentTerminalClosed  = "incident_closed"
)

type Message struct {
	Type       string          `json:"type"`
	IncidentID string          `json:"incident_id,omitempty"`
	EventID    string          `json:"event_id,omitempty"`
	EmittedAt  string          `json:"emitted_at,omitempty"`
	StreamSeq  *int64          `json:"stream_seq,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

func RawPayload(payload any) json.RawMessage {
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return data
}

const (
	ExtensionResourceChangeKindInvalidate = "invalidate"
	ExtensionResourceChangeKindRemove     = "remove"

	ExtensionResourceReasonRenamed           = "renamed"
	ExtensionResourceReasonSoftDeleted       = "soft_deleted"
	ExtensionResourceReasonAuthorizationLost = "authorization_lost"
)

type ExtensionWorkspaceRef struct {
	Kind               string `json:"kind"`
	ExtensionProfileID string `json:"extension_profile_id"`
	WorkspaceKey       string `json:"workspace_key"`
}

type ExtensionResourceChangePayload struct {
	ExtensionProfileID string                  `json:"extension_profile_id"`
	ResourceKind       string                  `json:"resource_kind"`
	ResourceID         string                  `json:"resource_id"`
	ChangeKind         string                  `json:"change_kind"`
	ReasonCode         string                  `json:"reason_code"`
	WorkspaceRefs      []ExtensionWorkspaceRef `json:"workspace_refs,omitempty"`
}

const (
	JobScopeKindIncident   = "incident"
	JobScopeKindDeployment = "deployment"

	JobStatusQueued          = "queued"
	JobStatusRunning         = "running"
	JobStatusCancelRequested = "cancel_requested"
	JobStatusSucceeded       = "succeeded"
	JobStatusFailed          = "failed"
	JobStatusCanceled        = "canceled"
)

type JobScope struct {
	Kind       string `json:"kind"`
	IncidentID string `json:"incident_id,omitempty"`
}

type JobProgress struct {
	Completed int64  `json:"completed"`
	Total     *int64 `json:"total"`
}

type JobProgressPayload struct {
	JobID         string      `json:"job_id"`
	Scope         JobScope    `json:"scope"`
	Status        string      `json:"status"`
	Progress      JobProgress `json:"progress"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Cancelable    *bool       `json:"cancelable,omitempty"`
	Message       string      `json:"message,omitempty"`
	ResultSummary any         `json:"result_summary,omitempty"`
	ErrorSummary  any         `json:"error_summary,omitempty"`
	RetainedUntil *time.Time  `json:"retained_until,omitempty"`
}

type PresenceInput struct {
	SheetRef map[string]string `json:"sheet_ref"`
	RecordID *string           `json:"record_id,omitempty"`
	FieldKey *string           `json:"field_key,omitempty"`
	Mode     string            `json:"mode"`
}

type PresenceRecord struct {
	ConnectionID string            `json:"connection_id"`
	UserID       string            `json:"user_id"`
	DisplayName  string            `json:"display_name"`
	SheetRef     map[string]string `json:"sheet_ref"`
	RecordID     *string           `json:"record_id,omitempty"`
	FieldKey     *string           `json:"field_key,omitempty"`
	Mode         string            `json:"mode"`
	ObservedAt   string            `json:"observed_at"`
	ExpiresAt    string            `json:"expires_at"`
}

func IsReplayableMessageType(messageType string) bool {
	switch messageType {
	case "record_changed", "extension_resource_changed", "job_progress":
		return true
	default:
		return false
	}
}

func IsClientMessageType(messageType string) bool {
	switch messageType {
	case "hello", "resume", "pong", "presence_update":
		return true
	default:
		return false
	}
}

func IsServerMessageType(messageType string) bool {
	switch messageType {
	case "hello_ack", "resume_ack", "presence_snapshot", "presence_delta",
		"record_changed", "extension_resource_changed", "job_progress", "ping",
		"error", "session_revoked":
		return true
	default:
		return false
	}
}

func EphemeralMessage(incidentID uuid.UUID, messageType string, payload any, now time.Time) Message {
	if payload == nil {
		payload = map[string]any{}
	}
	return Message{
		Type:       messageType,
		IncidentID: incidentID.String(),
		EventID:    uuid.New().String(),
		EmittedAt:  now.UTC().Format(time.RFC3339Nano),
		Payload:    RawPayload(payload),
	}
}

func PresenceSnapshotMessage(incidentID uuid.UUID, presences []PresenceRecord, now time.Time) Message {
	if presences == nil {
		presences = []PresenceRecord{}
	}
	return EphemeralMessage(incidentID, "presence_snapshot", map[string]any{"presences": presences}, now)
}

func NewIncidentJobProgressPayload(jobID string, incidentID uuid.UUID, status string, progress JobProgress, updatedAt time.Time) JobProgressPayload {
	return JobProgressPayload{
		JobID:  jobID,
		Scope:  JobScope{Kind: JobScopeKindIncident, IncidentID: incidentID.String()},
		Status: status, Progress: progress, UpdatedAt: updatedAt.UTC(),
	}
}

func ValidateIncidentJobProgressPayload(incidentID uuid.UUID, payload JobProgressPayload) error {
	if strings.TrimSpace(payload.JobID) == "" {
		return fmt.Errorf("job_progress.job_id is required")
	}
	if payload.Scope.Kind != JobScopeKindIncident {
		return fmt.Errorf("job_progress.scope.kind must be incident")
	}
	scopeIncidentID, err := uuid.Parse(payload.Scope.IncidentID)
	if err != nil || scopeIncidentID != incidentID {
		return fmt.Errorf("job_progress.scope.incident_id must match envelope incident_id")
	}
	switch payload.Status {
	case JobStatusQueued, JobStatusRunning, JobStatusCancelRequested, JobStatusSucceeded, JobStatusFailed, JobStatusCanceled:
	default:
		return fmt.Errorf("job_progress.status is invalid")
	}
	if payload.Progress.Completed < 0 {
		return fmt.Errorf("job_progress.progress.completed must be non-negative")
	}
	if payload.Progress.Total != nil && (*payload.Progress.Total <= 0 || payload.Progress.Completed > *payload.Progress.Total) {
		return fmt.Errorf("job_progress.progress total is invalid")
	}
	if payload.UpdatedAt.IsZero() {
		return fmt.Errorf("job_progress.updated_at is required")
	}
	return nil
}

func ValidateExtensionResourceChangePayload(payload ExtensionResourceChangePayload) error {
	if strings.TrimSpace(payload.ExtensionProfileID) == "" || strings.TrimSpace(payload.ResourceKind) == "" || strings.TrimSpace(payload.ResourceID) == "" {
		return fmt.Errorf("extension_resource_changed identity is required")
	}
	switch payload.ChangeKind {
	case ExtensionResourceChangeKindInvalidate, ExtensionResourceChangeKindRemove:
	default:
		return fmt.Errorf("extension_resource_changed.change_kind is invalid")
	}
	switch payload.ReasonCode {
	case ExtensionResourceReasonRenamed:
		if payload.ChangeKind != ExtensionResourceChangeKindInvalidate {
			return fmt.Errorf("extension_resource_changed.renamed requires invalidate")
		}
	case ExtensionResourceReasonSoftDeleted, ExtensionResourceReasonAuthorizationLost:
		if payload.ChangeKind != ExtensionResourceChangeKindRemove {
			return fmt.Errorf("extension_resource_changed.%s requires remove", payload.ReasonCode)
		}
	default:
		return fmt.Errorf("extension_resource_changed.reason_code is invalid")
	}
	lastWorkspaceKey := ""
	seenWorkspaceKeys := map[string]struct{}{}
	for _, ref := range payload.WorkspaceRefs {
		if ref.Kind != "extension_workspace" || ref.ExtensionProfileID != payload.ExtensionProfileID || strings.TrimSpace(ref.WorkspaceKey) == "" {
			return fmt.Errorf("extension_resource_changed.workspace_refs is invalid")
		}
		if _, exists := seenWorkspaceKeys[ref.WorkspaceKey]; exists {
			return fmt.Errorf("extension_resource_changed.workspace_refs duplicate workspace_key")
		}
		seenWorkspaceKeys[ref.WorkspaceKey] = struct{}{}
		if lastWorkspaceKey != "" && ref.WorkspaceKey < lastWorkspaceKey {
			return fmt.Errorf("extension_resource_changed.workspace_refs must be sorted by workspace_key")
		}
		lastWorkspaceKey = ref.WorkspaceKey
	}
	return nil
}

func CanonicalExtensionResourceChangePayload(payload ExtensionResourceChangePayload) ExtensionResourceChangePayload {
	payload.ExtensionProfileID = strings.TrimSpace(payload.ExtensionProfileID)
	payload.ResourceKind = strings.TrimSpace(payload.ResourceKind)
	payload.ResourceID = strings.TrimSpace(payload.ResourceID)
	payload.ChangeKind = strings.TrimSpace(payload.ChangeKind)
	payload.ReasonCode = strings.TrimSpace(payload.ReasonCode)
	refs := append([]ExtensionWorkspaceRef(nil), payload.WorkspaceRefs...)
	for index := range refs {
		refs[index].Kind = strings.TrimSpace(refs[index].Kind)
		refs[index].ExtensionProfileID = strings.TrimSpace(refs[index].ExtensionProfileID)
		refs[index].WorkspaceKey = strings.TrimSpace(refs[index].WorkspaceKey)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].WorkspaceKey < refs[j].WorkspaceKey })
	if len(refs) > 0 {
		payload.WorkspaceRefs = refs
	}
	return payload
}

func ValidatePresenceInput(input PresenceInput) error {
	if input.SheetRef == nil || input.SheetRef["kind"] == "" {
		return fmt.Errorf("presence.sheet_ref is required")
	}
	switch input.SheetRef["kind"] {
	case "view_schema", "saved_view":
		if input.SheetRef["id"] == "" || !sheetRefHasOnlyKeys(input.SheetRef, "kind", "id") {
			return fmt.Errorf("presence.sheet_ref is invalid")
		}
	case "extension_workspace":
		if input.SheetRef["extension_profile_id"] == "" || input.SheetRef["workspace_key"] == "" || !sheetRefHasOnlyKeys(input.SheetRef, "kind", "extension_profile_id", "workspace_key") || input.RecordID != nil || input.FieldKey != nil {
			return fmt.Errorf("presence extension workspace is invalid")
		}
	default:
		return fmt.Errorf("presence.sheet_ref.kind is invalid")
	}
	switch input.Mode {
	case "viewing", "editing", "idle":
	default:
		return fmt.Errorf("presence.mode is invalid")
	}
	if input.FieldKey != nil && input.Mode != "editing" {
		return fmt.Errorf("presence.field_key requires editing mode")
	}
	return nil
}

func sheetRefHasOnlyKeys(sheetRef map[string]string, keys ...string) bool {
	allowed := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		allowed[key] = struct{}{}
	}
	for key := range sheetRef {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}
