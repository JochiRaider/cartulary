package ownerfacade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type FinalizeCommand struct {
	Request         ImportOwnerCreateRequest
	ChangeSetID     uuid.UUID
	SequenceNo      int
	RecordID        uuid.UUID
	Operation       string
	CreatedOrReused string
	OwnerResultCode string
	BeforeVersionID *string
	BeforeValue     map[string]any
	BeforeSnapshot  *revisions.CapturedRecordSnapshot
	Row             map[string]any
}

type CapturedRevisionAppender interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.CapturedRecordSnapshot, error)
	AppendCapturedRecordMutationTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordMutationParams) error
	AppendCapturedRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordRevisionParams) error
}

func ValuesByField(fields []ImportFieldValue) map[string]ImportScalarValue {
	values := make(map[string]ImportScalarValue, len(fields))
	for _, field := range fields {
		values[field.FieldKey] = field.NormalizedValue
	}
	return values
}

func FinalizeCapturedTx(ctx context.Context, tx pgx.Tx, revisionAppender CapturedRevisionAppender, command FinalizeCommand) (ImportOwnerCreateResponse, error) {
	if revisionAppender == nil {
		return ImportOwnerCreateResponse{}, fmt.Errorf("finalize captured import owner row: revision appender is required")
	}
	rowVersion, err := RowVersionFromRow(command.Row)
	if err != nil {
		return ImportOwnerCreateResponse{}, err
	}
	operation := command.Operation
	if operation == "" {
		operation = "create"
	}
	createdOrReused := command.CreatedOrReused
	if createdOrReused == "" {
		createdOrReused = "created"
	}
	resultCode := command.OwnerResultCode
	if resultCode == "" {
		resultCode = createdOrReused
	}
	afterSnapshot, err := revisionAppender.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return ImportOwnerCreateResponse{}, err
	}
	afterVersionID := VersionID(command.RecordID, rowVersion)
	if err := revisionAppender.AppendCapturedRecordMutationTx(ctx, tx, revisions.AppendCapturedRecordMutationParams{
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		TargetKind:      "record",
		RecordID:        command.RecordID,
		OperationKind:   operation,
		BeforeVersionID: command.BeforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  command.BeforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return ImportOwnerCreateResponse{}, err
	}
	if operation == "create" || (command.BeforeSnapshot != nil && operation != "reuse") {
		if err := revisionAppender.AppendCapturedRecordRevisionTx(ctx, tx, revisions.AppendCapturedRecordRevisionParams{
			ChangeSetID:    command.ChangeSetID,
			RecordID:       command.RecordID,
			RowVersion:     rowVersion,
			BeforeSnapshot: command.BeforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			LiveChange: revisions.LiveRecordChange{
				BeforeValue: command.BeforeValue,
				AfterValue:  command.Row,
			},
		}); err != nil {
			return ImportOwnerCreateResponse{}, err
		}
	}
	return ImportOwnerCreateResponse{
		RecordID:             command.RecordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      createdOrReused,
		OwnerResultCode:      resultCode,
		RowRefresh:           command.Row,
	}, nil
}

func RowVersionFromRow(row map[string]any) (int64, error) {
	switch value := row["row_version"].(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("import row has unexpected row_version type %T", value)
	}
}

func VersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}
