package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
)

const (
	RecoveryJournalPayloadSchemaID   = "cartulary.operator_recovery_journal_payload.v4"
	RecoveryJournalPayloadV3SchemaID = "cartulary.operator_recovery_journal_payload.v3"
	RecoveryJournalPayloadV2SchemaID = "cartulary.operator_recovery_journal_payload.v2"
	RecoveryAuditSummarySchemaID     = "cartulary.operator_recovery_audit_summary.v2"
)

type ArtifactCount struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

type RecoveryAdmissionRecord struct {
	OperationID        uuid.UUID
	Operation          Operation
	AttemptID          *string
	StartedAt          time.Time
	BackupSetID        *uuid.UUID
	ConsistencyPointAt *time.Time
	ArtifactKinds      []string
}

type RecoveryCompletionRecord struct {
	OperationID               uuid.UUID
	Operation                 Operation
	AttemptID                 *string
	StartedAt                 time.Time
	CompletedAt               time.Time
	Result                    ResultStatus
	BackupSetID               *uuid.UUID
	ConsistencyPointAt        *time.Time
	ArtifactCounts            []ArtifactCount
	ErrorCode                 *string
	ErrorReason               *string
	GraphProjectionCompletion *GraphProjectionCompletionEvidence
}

type GraphProjectionCompletionEvidence = restorecontract.GraphProjectionCompletionEvidence

type RecoveryEvidenceRepository interface {
	AppendAdmission(context.Context, RecoveryAdmissionRecord) error
	AppendCompletion(context.Context, RecoveryCompletionRecord) error
}

// RecoveryEvidenceReplayReader is an optional read capability implemented by
// durable repositories. Writers that do not provide it remain valid for unit
// isolation, but production Recovery uses it to replay exact terminal success.
type RecoveryEvidenceReplayReader interface {
	FindSuccessfulCompletion(context.Context, uuid.UUID, Operation, *string, uuid.UUID) (*RecoveryCompletionRecord, error)
}

type RecoveryEvidenceRepositoryFactory func(PostgresPool) (RecoveryEvidenceRepository, error)

func NormalizeAdmissionRecord(record RecoveryAdmissionRecord) (RecoveryAdmissionRecord, error) {
	if err := validateEvidenceIdentity(record.OperationID, record.Operation, record.AttemptID); err != nil {
		return RecoveryAdmissionRecord{}, err
	}
	record.StartedAt = record.StartedAt.UTC()
	if record.StartedAt.IsZero() {
		return RecoveryAdmissionRecord{}, fmt.Errorf("recovery admission started_at is required")
	}
	record.ConsistencyPointAt = normalizedTimePointer(record.ConsistencyPointAt)
	record.ArtifactKinds = normalizedArtifactKinds(record.ArtifactKinds)
	return record, nil
}

func NormalizeCompletionRecord(record RecoveryCompletionRecord) (RecoveryCompletionRecord, error) {
	if err := validateEvidenceIdentity(record.OperationID, record.Operation, record.AttemptID); err != nil {
		return RecoveryCompletionRecord{}, err
	}
	record.StartedAt = record.StartedAt.UTC()
	record.CompletedAt = record.CompletedAt.UTC()
	if record.StartedAt.IsZero() || record.CompletedAt.IsZero() {
		return RecoveryCompletionRecord{}, fmt.Errorf("recovery completion timestamps are required")
	}
	if record.CompletedAt.Before(record.StartedAt) {
		return RecoveryCompletionRecord{}, fmt.Errorf("recovery completion completed_at precedes started_at")
	}
	switch record.Result {
	case ResultSucceeded, ResultNoOp:
		if record.Result == ResultNoOp && record.Operation != OperationRestoreVerifyDue {
			return RecoveryCompletionRecord{}, fmt.Errorf("recovery completion no_op is only valid for restore_verify_due")
		}
		if record.ErrorCode != nil || record.ErrorReason != nil {
			return RecoveryCompletionRecord{}, fmt.Errorf("successful recovery completion must not contain error fields")
		}
	default:
		if record.Result != ResultFailed {
			return RecoveryCompletionRecord{}, fmt.Errorf("recovery completion result %q is invalid", record.Result)
		}
		if normalizedOptionalString(record.ErrorCode) == nil || normalizedOptionalString(record.ErrorReason) == nil {
			return RecoveryCompletionRecord{}, fmt.Errorf("failed recovery completion requires error code and reason")
		}
		record.ErrorCode = normalizedOptionalString(record.ErrorCode)
		record.ErrorReason = normalizedOptionalString(record.ErrorReason)
	}
	record.ConsistencyPointAt = normalizedTimePointer(record.ConsistencyPointAt)
	counts, err := normalizedArtifactCounts(record.ArtifactCounts)
	if err != nil {
		return RecoveryCompletionRecord{}, err
	}
	record.ArtifactCounts = counts
	if record.GraphProjectionCompletion != nil {
		completion := *record.GraphProjectionCompletion
		completion.ConsistencyPointAt = completion.ConsistencyPointAt.UTC()
		if record.Operation != OperationRestoreLatest && record.Operation != OperationRestoreVerifyLatest && record.Operation != OperationRestoreVerifyDue {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion is valid only for restore operations")
		}
		if completion.TargetGenerationID == uuid.Nil || completion.RestoreOperationID != record.OperationID {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion operation or generation identity is inconsistent")
		}
		if completion.BackupSetID == uuid.Nil || record.BackupSetID == nil || completion.BackupSetID != *record.BackupSetID {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion backup identity is inconsistent")
		}
		if completion.ConsistencyPointAt.IsZero() || record.ConsistencyPointAt == nil || !completion.ConsistencyPointAt.Equal(*record.ConsistencyPointAt) {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion consistency point is inconsistent")
		}
		if completion.RecoveryStateCatalogSHA256 == "" {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion recovery catalog digest is missing")
		}
		if completion.SourceRegistrySHA256 == "" {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion source registry digest is missing")
		}
		if completion.ImplementationBindingSHA256 == "" {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion implementation binding digest is missing")
		}
		if completion.PostconditionSHA256 == "" {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection completion postcondition digest is missing")
		}
		if completion.ParticipantResult.RestoreOperationID != record.OperationID.String() ||
			completion.ParticipantResult.TargetGenerationID != completion.TargetGenerationID.String() {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection participant result identity is inconsistent")
		}
		if completion.ParticipantResult.SourceRegistrySHA256 != completion.SourceRegistrySHA256 ||
			completion.ParticipantResult.ImplementationBindingSHA256 != completion.ImplementationBindingSHA256 ||
			completion.ParticipantResult.PostconditionSHA256 == nil || *completion.ParticipantResult.PostconditionSHA256 != completion.PostconditionSHA256 {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection participant result digest tuple is inconsistent")
		}
		if !completion.ParticipantResult.ReadinessSatisfied() {
			return RecoveryCompletionRecord{}, fmt.Errorf("graph-projection participant result does not prove readiness")
		}
		record.GraphProjectionCompletion = &completion
	}
	return record, nil
}

type DecodedRecoveryJournalPayload struct {
	SchemaID                  string
	RecordKind                string
	GraphProjectionCompletion *GraphProjectionCompletionEvidence
}

func DecodeRecoveryJournalPayload(body []byte) (DecodedRecoveryJournalPayload, error) {
	var selector struct {
		SchemaID   string `json:"schema_id"`
		RecordKind string `json:"record_kind"`
	}
	if err := json.Unmarshal(body, &selector); err != nil {
		return DecodedRecoveryJournalPayload{}, fmt.Errorf("decode Recovery journal selector: %w", err)
	}
	var destination any
	switch selector.SchemaID {
	case RecoveryJournalPayloadV2SchemaID:
		if selector.RecordKind == "admission" {
			destination = &recoveryJournalAdmissionPayloadV2{}
		} else if selector.RecordKind == "completion" {
			destination = &recoveryJournalCompletionPayloadV2{}
		}
	case RecoveryJournalPayloadV3SchemaID:
		if selector.RecordKind == "admission" {
			destination = &recoveryJournalAdmissionPayloadV3{}
		} else if selector.RecordKind == "completion" {
			destination = &recoveryJournalCompletionPayloadV3{}
		}
	case RecoveryJournalPayloadSchemaID:
		if selector.RecordKind == "admission" {
			destination = &recoveryJournalAdmissionPayloadV4{}
		} else if selector.RecordKind == "completion" {
			destination = &recoveryJournalCompletionPayloadV4{}
		}
	}
	if destination == nil {
		return DecodedRecoveryJournalPayload{}, fmt.Errorf("unsupported Recovery journal payload schema or record kind")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return DecodedRecoveryJournalPayload{}, fmt.Errorf("decode Recovery journal payload: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return DecodedRecoveryJournalPayload{}, fmt.Errorf("recovery journal payload has trailing content")
	}
	decoded := DecodedRecoveryJournalPayload{SchemaID: selector.SchemaID, RecordKind: selector.RecordKind}
	if completion, ok := destination.(*recoveryJournalCompletionPayloadV3); ok {
		decoded.GraphProjectionCompletion = completion.GraphProjectionCompletion
	}
	if completion, ok := destination.(*recoveryJournalCompletionPayloadV4); ok {
		decoded.GraphProjectionCompletion = completion.GraphProjectionCompletion
	}
	return decoded, nil
}

type recoveryJournalAdmissionPayloadV2 struct {
	SchemaID           string     `json:"schema_id"`
	RecordKind         string     `json:"record_kind"`
	OperationID        uuid.UUID  `json:"operation_id"`
	Operation          Operation  `json:"operation"`
	AttemptID          *string    `json:"attempt_id"`
	StartedAt          time.Time  `json:"started_at"`
	BackupSetID        *uuid.UUID `json:"backup_set_id"`
	ConsistencyPointAt *time.Time `json:"consistency_point_at"`
	ArtifactKinds      []string   `json:"artifact_kinds"`
}

type recoveryJournalCompletionPayloadV2 struct {
	SchemaID           string          `json:"schema_id"`
	RecordKind         string          `json:"record_kind"`
	OperationID        uuid.UUID       `json:"operation_id"`
	Operation          Operation       `json:"operation"`
	AttemptID          *string         `json:"attempt_id"`
	StartedAt          time.Time       `json:"started_at"`
	CompletedAt        time.Time       `json:"completed_at"`
	Result             ResultStatus    `json:"result"`
	BackupSetID        *uuid.UUID      `json:"backup_set_id"`
	ConsistencyPointAt *time.Time      `json:"consistency_point_at"`
	ArtifactCounts     []ArtifactCount `json:"artifact_counts"`
	ErrorCode          *string         `json:"error_code"`
	ErrorReason        *string         `json:"error_reason"`
}

type recoveryJournalAdmissionPayloadV3 = recoveryJournalAdmissionPayloadV2

type recoveryJournalCompletionPayloadV3 struct {
	recoveryJournalCompletionPayloadV2
	GraphProjectionCompletion *GraphProjectionCompletionEvidence `json:"graph_projection_completion"`
}

type recoveryJournalAdmissionPayloadV4 = recoveryJournalAdmissionPayloadV2

type recoveryJournalCompletionPayloadV4 struct {
	recoveryJournalCompletionPayloadV2
	GraphProjectionCompletion *GraphProjectionCompletionEvidence `json:"graph_projection_completion"`
}

func ArtifactCountsFor(refs []ArtifactRef) []ArtifactCount {
	counts := make(map[string]int)
	for _, ref := range refs {
		kind := strings.TrimSpace(ref.Kind)
		if kind != "" {
			counts[kind]++
		}
	}
	out := make([]ArtifactCount, 0, len(counts))
	for kind, count := range counts {
		out = append(out, ArtifactCount{Kind: kind, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Kind < out[j].Kind
	})
	return out
}

func validateEvidenceIdentity(operationID uuid.UUID, operation Operation, attemptID *string) error {
	if operationID == uuid.Nil {
		return fmt.Errorf("recovery evidence operation_id is required")
	}
	switch operation {
	case OperationBackupCreate, OperationRestoreLatest, OperationRestoreVerifyLatest, OperationRestoreVerifyDue:
	default:
		return fmt.Errorf("recovery evidence operation %q is not mutating", operation)
	}
	if attemptID != nil && strings.TrimSpace(*attemptID) == "" {
		return fmt.Errorf("recovery evidence attempt_id must be null or non-empty")
	}
	return nil
}

func normalizedArtifactKinds(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func normalizedArtifactCounts(values []ArtifactCount) ([]ArtifactCount, error) {
	counts := make(map[string]int, len(values))
	for _, value := range values {
		kind := strings.TrimSpace(value.Kind)
		if kind == "" || value.Count < 0 {
			return nil, fmt.Errorf("recovery artifact count is invalid")
		}
		if _, duplicate := counts[kind]; duplicate {
			return nil, fmt.Errorf("recovery artifact count kind %q is duplicated", kind)
		}
		counts[kind] = value.Count
	}
	out := make([]ArtifactCount, 0, len(counts))
	for kind, count := range counts {
		out = append(out, ArtifactCount{Kind: kind, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Kind < out[j].Kind
	})
	return out, nil
}

func normalizedOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func normalizedTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	if normalized.IsZero() {
		return nil
	}
	return &normalized
}
