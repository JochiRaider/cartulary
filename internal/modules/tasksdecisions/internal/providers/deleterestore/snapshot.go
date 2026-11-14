package deleterestore

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/textpolicy"
)

func ValidateTaskRequestSnapshotSource(record, source map[string]any) error {
	if err := shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "title status priority task_kind", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "workstream requester_party_text blocked_reason external_ticket_ref closure_summary", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "owner_user_id requester_party_id decision_record_id", Kind: shape.SourceUUID, Nullable: true},
		shape.SourceFieldGroup{Names: "due_at completed_at", Kind: shape.SourceTimestamp, Nullable: true},
		shape.SourceFieldGroup{Names: "created_at updated_at", Kind: shape.SourceTimestamp, Nullable: false},
	); err != nil {
		return err
	}
	if err := textpolicy.ValidateStoredSource("task_request", source); err != nil {
		return err
	}
	status := source["status"].(string)
	if !policy.ValidTaskStatus(status) || !policy.ValidTaskPriority(source["priority"].(string)) || !policy.ValidTaskKind(source["task_kind"].(string)) {
		return fmt.Errorf("invalid historical Task Request vocabulary")
	}
	if strings.TrimSpace(source["title"].(string)) == "" {
		return fmt.Errorf("missing historical Task Request title")
	}
	for _, value := range source {
		if text, ok := value.(string); ok && strings.ContainsRune(text, 0) {
			return fmt.Errorf("invalid historical Task Request text")
		}
	}
	created, _ := time.Parse(time.RFC3339Nano, source["created_at"].(string))
	state := policy.TaskLifecycleState{Status: status, CreatedAt: created}
	if owner, ok := source["owner_user_id"].(string); ok {
		state.OwnerUserID = sql.NullString{String: owner, Valid: true}
	}
	if blocker, ok := source["blocked_reason"].(string); ok {
		if status == "blocked" && strings.TrimSpace(blocker) == "" {
			return fmt.Errorf("missing historical Task Request blocker")
		}
		state.BlockedReason = sql.NullString{String: blocker, Valid: blocker != ""}
	}
	if completed, ok := source["completed_at"].(string); ok {
		instant, _ := time.Parse(time.RFC3339Nano, completed)
		state.CompletedAt = sql.NullTime{Time: instant, Valid: true}
	}
	return policy.ValidateTaskState(state)
}

func ValidateDecisionSnapshotSource(record, source map[string]any) error {
	if err := shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "status summary decision_type rationale", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "owner_user_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "decided_at created_at updated_at", Kind: shape.SourceTimestamp, Nullable: false},
	); err != nil {
		return err
	}
	if err := textpolicy.ValidateStoredSource("decision", source); err != nil {
		return err
	}
	if !policy.ValidDecisionStatus(source["status"].(string)) || !policy.ValidDecisionType(source["decision_type"].(string)) {
		return fmt.Errorf("invalid historical Decision vocabulary")
	}
	for _, field := range []string{"summary", "rationale"} {
		text := source[field].(string)
		if strings.TrimSpace(text) == "" || strings.ContainsRune(text, 0) {
			return fmt.Errorf("invalid historical Decision %s", field)
		}
	}
	return nil
}
