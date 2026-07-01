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
	Row             map[string]any
}

func ValuesByField(fields []tabularingest.ImportFieldValue) map[string]tabularingest.ImportScalarValue {
	values := make(map[string]tabularingest.ImportScalarValue, len(fields))
	for _, field := range fields {
		values[field.FieldKey] = field.NormalizedValue
	}
	return values
}

func FinalizeTx(ctx context.Context, tx pgx.Tx, revisionStore *revisions.Store, command FinalizeCommand) (tabularingest.ImportOwnerCreateResponse, error) {
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
	if revisionStore == nil {
		revisionStore = revisions.NewStore()
	}
	if err := revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    command.ChangeSetID,
		SequenceNo:     command.SequenceNo,
		TargetKind:     "record",
		TargetID:       command.RecordID.String(),
		OperationKind:  operation,
		AfterVersionID: &afterVersionID,
		AfterValue:     command.Row,
	}); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if operation == "create" {
		if err := revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
			ChangeSetID: command.ChangeSetID,
			RecordID:    command.RecordID,
			RowVersion:  rowVersion,
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
