package application

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RecoveryJournalPayloadSchemaID = "cartulary.operator_recovery_journal_payload.v2"
	RecoveryAuditSummarySchemaID   = "cartulary.operator_recovery_audit_summary.v2"
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
	OperationID        uuid.UUID
	Operation          Operation
	AttemptID          *string
	StartedAt          time.Time
	CompletedAt        time.Time
	Result             ResultStatus
	BackupSetID        *uuid.UUID
	ConsistencyPointAt *time.Time
	ArtifactCounts     []ArtifactCount
	ErrorCode          *string
	ErrorReason        *string
}

type RecoveryEvidenceRepository interface {
	AppendAdmission(context.Context, RecoveryAdmissionRecord) error
	AppendCompletion(context.Context, RecoveryCompletionRecord) error
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
	return record, nil
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
