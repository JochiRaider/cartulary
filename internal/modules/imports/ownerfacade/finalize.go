package ownerfacade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type FinalizeCommand struct {
	Request         tabularingest.ImportOwnerCreateRequest
	ChangeSetID     uuid.UUID
	SequenceNo      int
	RecordID        uuid.UUID
	Operation       string
	CreatedOrReused string
	OwnerResultCode string
	BeforeVersionID *string
	BeforeValue     map[string]any
	Row             map[string]any
}

type RevisionAppender interface {
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendMutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
}

func ValuesByField(fields []tabularingest.ImportFieldValue) map[string]tabularingest.ImportScalarValue {
	values := make(map[string]tabularingest.ImportScalarValue, len(fields))
	for _, field := range fields {
		values[field.FieldKey] = field.NormalizedValue
	}
	return values
}

func FinalizeTx(ctx context.Context, tx pgx.Tx, revisionAppender RevisionAppender, command FinalizeCommand) (tabularingest.ImportOwnerCreateResponse, error) {
	if revisionAppender == nil {
		return tabularingest.ImportOwnerCreateResponse{}, fmt.Errorf("finalize import owner row: revision appender is required")
	}
	rowVersion, err := RowVersionFromRow(command.Row)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
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
	afterVersionID := VersionID(command.RecordID, rowVersion)
	if err := revisionAppender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		TargetKind:      "record",
		TargetID:        command.RecordID.String(),
		OperationKind:   operation,
		BeforeVersionID: command.BeforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     command.BeforeValue,
		AfterValue:      command.Row,
	}); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if operation == "create" {
		if err := revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID: command.ChangeSetID,
			RecordID:    command.RecordID,
			RowVersion:  rowVersion,
			AfterValue:  command.Row,
		}); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
	} else if command.BeforeValue != nil && operation != "reuse" {
		if err := revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID: command.ChangeSetID,
			RecordID:    command.RecordID,
			RowVersion:  rowVersion,
			BeforeValue: command.BeforeValue,
			AfterValue:  command.Row,
		}); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
	}
	return tabularingest.ImportOwnerCreateResponse{
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
