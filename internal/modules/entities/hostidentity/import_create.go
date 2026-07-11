package hostidentity

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ImportCreateCommand struct {
	Request     ownerfacade.ImportOwnerCreateRequest
	ChangeSetID uuid.UUID
	SequenceNo  int
	Now         time.Time
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	createRequest := entityCreateRequestFromImport(request.ClientTxnID, request.FieldValues)
	now := command.Now.UTC()
	actor := authn.UserRecord{ID: request.ActorUserID}

	var (
		recordID      uuid.UUID
		rowVersion    int64
		beforeRow     map[string]any
		afterRow      map[string]any
		operationKind string
		entityType    string
	)
	switch request.TargetViewSchemaID {
	case HostsViewSchemaID:
		record, before, operation, _, err := s.upsertHostTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshEntityRowTx(ctx, tx, record.RecordID, "host"); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID = record.RecordID
		rowVersion = record.RowVersion
		beforeRow = before
		afterRow = BuildHostRow(record)
		operationKind = operation
		entityType = "host"
	case IdentitiesViewSchemaID:
		record, before, operation, _, err := s.upsertIdentityTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshEntityRowTx(ctx, tx, record.RecordID, "identity"); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID = record.RecordID
		rowVersion = record.RowVersion
		beforeRow = before
		afterRow = BuildIdentityRow(record)
		operationKind = operation
		entityType = "identity"
	default:
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("entity import surface %q not mapped", request.TargetViewSchemaID)
	}

	createdOrReused := "created"
	resultCode := "created"
	operation := operationKind
	var beforeVersionID *string
	var beforeValue map[string]any
	if beforeRow != nil {
		createdOrReused = "reused"
		resultCode = "reused"
		if reflect.DeepEqual(beforeRow, afterRow) {
			operation = "reuse"
		} else {
			resultCode = "updated"
			beforeValue = beforeRow
			if rowVersion > 1 {
				value := ownerfacade.VersionID(recordID, rowVersion-1)
				beforeVersionID = &value
			}
		}
	}
	if operation == "" {
		operation = entityType + "_import_create"
	}
	return ownerfacade.FinalizeTx(ctx, tx, revisions.NewAppender(), ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        recordID,
		Operation:       operation,
		CreatedOrReused: createdOrReused,
		OwnerResultCode: resultCode,
		BeforeVersionID: beforeVersionID,
		BeforeValue:     beforeValue,
		Row:             afterRow,
	})
}

func entityCreateRequestFromImport(clientTxnID string, fields []ownerfacade.ImportFieldValue) CreateRequest {
	request := CreateRequest{
		ClientTxnID: clientTxnID,
		Values:      make(map[string]string, len(fields)),
		AliasAdds:   make(map[string][]CollectionAction),
	}
	for _, field := range fields {
		value := field.NormalizedValue.Text
		if value == nil {
			continue
		}
		switch field.FieldKey {
		case "host.aliases", "identity.aliases":
			request.AliasAdds[field.FieldKey] = append(request.AliasAdds[field.FieldKey], CollectionAction{
				Op:             "add_alias",
				RawText:        *value,
				NormalizedText: *value,
			})
		default:
			request.Values[field.FieldKey] = *value
		}
	}
	return request
}
