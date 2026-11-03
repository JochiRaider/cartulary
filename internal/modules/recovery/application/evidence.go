package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const (
	RecoveryJournalPayloadSchemaID = recovery.RecoveryJournalPayloadSchemaID
	RecoveryAuditSummarySchemaID   = "cartulary.operator_recovery_audit_summary.v3"
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
	RestoreGeneration  *recovery.RestoreGenerationEvidence
	TargetBindings     *recovery.TargetBindingDigests
}

type RecoveryEvidenceRepository interface {
	AppendAdmission(context.Context, RecoveryAdmissionRecord) error
	AppendCompletion(context.Context, RecoveryCompletionRecord) error
	FindCompletion(context.Context, uuid.UUID, Operation, *string) (*RecoveryCompletionRecord, error)
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
	if record.TargetBindings != nil {
		b := *record.TargetBindings
		if !isLowerSHA256(b.DatabaseSHA256) || !isLowerSHA256(b.ObjectStoreSHA256) || !isLowerSHA256(b.ReferencePackStorageSHA256) || !isLowerSHA256(b.ExportOutputsSHA256) {
			return RecoveryCompletionRecord{}, fmt.Errorf("restore target binding digests are invalid")
		}
		record.TargetBindings = &b
	} else if record.RestoreGeneration != nil {
		return RecoveryCompletionRecord{}, fmt.Errorf("restore completion target binding digests are required")
	}
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
	perTarget := record.Result != ResultNoOp && (record.Operation == OperationRestoreLatest || record.Operation == OperationRestoreBundle || record.Operation == OperationRestoreVerifyLatest ||
		(record.Operation == OperationRestoreVerifyDue && record.AttemptID != nil))
	if perTarget && record.RestoreGeneration != nil {
		generation := record.RestoreGeneration
		if generation == nil || generation.TargetGenerationID == uuid.Nil || !isLowerSHA256(generation.RecoveryStateCatalogSHA256) || !isLowerSHA256(generation.CodecRegistrySHA256) || record.TargetBindings == nil || record.BackupSetID == nil || *record.BackupSetID == uuid.Nil || record.ConsistencyPointAt == nil {
			return RecoveryCompletionRecord{}, fmt.Errorf("admitted restore completion generation and identity are required")
		}
		copied := *generation
		record.RestoreGeneration = &copied
	} else if record.RestoreGeneration != nil || record.TargetBindings != nil {
		return RecoveryCompletionRecord{}, fmt.Errorf("restore generation and bindings must be paired on a per-target completion")
	} else if perTarget && record.Result == ResultSucceeded {
		return RecoveryCompletionRecord{}, fmt.Errorf("successful restore completion requires admitted generation and bindings")
	}
	return record, nil
}

type DecodedRecoveryJournalPayload struct {
	SchemaID          string
	RecordKind        string
	RestoreGeneration *recovery.RestoreGenerationEvidence
	TargetBindings    *recovery.TargetBindingDigests
}

func DecodeRecoveryJournalPayload(body []byte) (DecodedRecoveryJournalPayload, error) {
	if _, err := canonicaljson.DecodeStrict(body); err != nil {
		return DecodedRecoveryJournalPayload{}, fmt.Errorf("invalid Recovery journal JSON")
	}
	var selector struct {
		SchemaID   string `json:"schema_id"`
		RecordKind string `json:"record_kind"`
	}
	if err := json.Unmarshal(body, &selector); err != nil {
		return DecodedRecoveryJournalPayload{}, fmt.Errorf("decode Recovery journal selector: %w", err)
	}
	var destination any
	switch selector.SchemaID {
	case RecoveryJournalPayloadSchemaID:
		if selector.RecordKind == "admission" {
			destination = &recoveryJournalAdmissionPayloadV7{}
		} else if selector.RecordKind == "completion" {
			destination = &recoveryJournalCompletionPayloadV7{}
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
	if selector.SchemaID == RecoveryJournalPayloadSchemaID {
		// All current members are required, including nullable members. Comparing
		// the lossless admitted object with its typed projection rejects omitted
		// fields and null-to-zero decoding at every nested struct boundary.
		admitted, err := canonicaljson.Canonicalize(body)
		projected, projectionErr := canonicaljson.Marshal(destination)
		if err != nil || projectionErr != nil || !bytes.Equal(admitted, projected) {
			return DecodedRecoveryJournalPayload{}, fmt.Errorf("recovery journal payload has an incomplete or invalid shape")
		}
		if err := validateCurrentJournalPayload(destination); err != nil {
			return DecodedRecoveryJournalPayload{}, fmt.Errorf("recovery journal payload violates its current contract")
		}
	}
	decoded := DecodedRecoveryJournalPayload{SchemaID: selector.SchemaID, RecordKind: selector.RecordKind}
	if completion, ok := destination.(*recoveryJournalCompletionPayloadV7); ok {
		decoded.RestoreGeneration = completion.RestoreGeneration
		decoded.TargetBindings = completion.TargetBindings
	}
	return decoded, nil
}

func validateCurrentJournalPayload(destination any) error {
	switch value := destination.(type) {
	case *recoveryJournalAdmissionPayloadV7:
		if value.ArtifactKinds == nil || len(value.ArtifactKinds) > 128 {
			return fmt.Errorf("invalid artifact kinds")
		}
		normalized, err := NormalizeAdmissionRecord(RecoveryAdmissionRecord{OperationID: value.OperationID, Operation: value.Operation, AttemptID: value.AttemptID, StartedAt: value.StartedAt, BackupSetID: value.BackupSetID, ConsistencyPointAt: value.ConsistencyPointAt, ArtifactKinds: value.ArtifactKinds})
		if err != nil || len(normalized.ArtifactKinds) != len(value.ArtifactKinds) {
			return fmt.Errorf("invalid admission")
		}
		return nil
	case *recoveryJournalCompletionPayloadV7:
		if value.ArtifactCounts == nil || len(value.ArtifactCounts) > 128 {
			return fmt.Errorf("invalid artifact counts")
		}
		_, err := NormalizeCompletionRecord(RecoveryCompletionRecord{OperationID: value.OperationID, Operation: value.Operation, AttemptID: value.AttemptID, StartedAt: value.StartedAt, CompletedAt: value.CompletedAt, Result: value.Result, BackupSetID: value.BackupSetID, ConsistencyPointAt: value.ConsistencyPointAt, ArtifactCounts: value.ArtifactCounts, ErrorCode: value.ErrorCode, ErrorReason: value.ErrorReason, RestoreGeneration: value.RestoreGeneration, TargetBindings: value.TargetBindings})
		return err
	default:
		return fmt.Errorf("invalid record kind")
	}
}

type recoveryJournalAdmissionPayloadV7 struct {
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

type recoveryJournalCompletionFields struct {
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

type recoveryJournalCompletionPayloadV7 struct {
	recoveryJournalCompletionFields
	RestoreGeneration *recovery.RestoreGenerationEvidence `json:"restore_generation"`
	TargetBindings    *recovery.TargetBindingDigests      `json:"target_binding_digests"`
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
	case OperationBackupCreate, OperationBackupExportLatest, OperationRestoreBundle, OperationRestoreLatest, OperationRestoreVerifyLatest, OperationRestoreVerifyDue:
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

// EncodeRecoveryCompletion shares the current closed journal shape across live
// and portable evidence stores.
func EncodeRecoveryCompletion(record RecoveryCompletionRecord) ([]byte, error) {
	record, err := NormalizeCompletionRecord(record)
	if err != nil {
		return nil, err
	}
	value := recoveryJournalCompletionPayloadV7{
		recoveryJournalCompletionFields: recoveryJournalCompletionFields{SchemaID: RecoveryJournalPayloadSchemaID, RecordKind: "completion", OperationID: record.OperationID, Operation: record.Operation, AttemptID: record.AttemptID, StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, Result: record.Result, BackupSetID: record.BackupSetID, ConsistencyPointAt: record.ConsistencyPointAt, ArtifactCounts: record.ArtifactCounts, ErrorCode: record.ErrorCode, ErrorReason: record.ErrorReason},
		RestoreGeneration:               record.RestoreGeneration, TargetBindings: record.TargetBindings,
	}
	return canonicaljson.Marshal(value)
}
func DecodeRecoveryCompletion(body []byte) (RecoveryCompletionRecord, error) {
	decoded, err := DecodeRecoveryJournalPayload(body)
	if err != nil || decoded.RecordKind != "completion" {
		return RecoveryCompletionRecord{}, fmt.Errorf("invalid completion payload")
	}
	var value recoveryJournalCompletionPayloadV7
	if err := json.Unmarshal(body, &value); err != nil {
		return RecoveryCompletionRecord{}, err
	}
	record, err := NormalizeCompletionRecord(RecoveryCompletionRecord{OperationID: value.OperationID, Operation: value.Operation, AttemptID: value.AttemptID, StartedAt: value.StartedAt, CompletedAt: value.CompletedAt, Result: value.Result, BackupSetID: value.BackupSetID, ConsistencyPointAt: value.ConsistencyPointAt, ArtifactCounts: value.ArtifactCounts, ErrorCode: value.ErrorCode, ErrorReason: value.ErrorReason, RestoreGeneration: value.RestoreGeneration, TargetBindings: value.TargetBindings})
	if err != nil {
		return RecoveryCompletionRecord{}, err
	}
	canonical, err := EncodeRecoveryCompletion(record)
	if err != nil || !bytes.Equal(body, canonical) {
		return RecoveryCompletionRecord{}, fmt.Errorf("completion payload is not the exact current canonical encoding")
	}
	return record, nil
}
