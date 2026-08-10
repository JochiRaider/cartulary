package hostidentity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	pool postgres.DB,
	appender *revisions.Appender,
	projectionWriter workbookprojection.Writer,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != HostsViewSchemaID && targetViewSchemaID != IdentitiesViewSchemaID {
		return nil, fmt.Errorf("entity import surface %q not mapped", targetViewSchemaID)
	}
	store := NewStore(pool, appender, nil, projectionWriter)
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		store.CreateImportRowTx,
	)
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	createRequest := entityCreateRequestFromImport(request.ClientTxnID, request.FieldValues)
	now := command.Now.UTC()
	actor := authn.UserRecord{ID: request.ActorUserID}

	var (
		recordID       uuid.UUID
		rowVersion     int64
		beforeRow      map[string]any
		afterRow       map[string]any
		operationKind  string
		entityType     string
		beforeSnapshot *revisions.CapturedRecordSnapshot
		err            error
	)
	switch request.TargetViewSchemaID {
	case HostsViewSchemaID:
		beforeSnapshot, err = s.captureHostSnapshotBeforeUpsertTx(ctx, tx, request.IncidentID, createRequest)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		record, before, operation, _, err := s.upsertHostTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshHostTx(ctx, tx, record.RecordID); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID = record.RecordID
		rowVersion = record.RowVersion
		beforeRow = before
		afterRow = BuildHostRow(record)
		operationKind = operation
		entityType = "host"
	case IdentitiesViewSchemaID:
		beforeSnapshot, err = s.captureIdentitySnapshotBeforeUpsertTx(ctx, tx, request.IncidentID, createRequest)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		record, before, operation, _, err := s.upsertIdentityTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshIdentityTx(ctx, tx, record.RecordID); err != nil {
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
	return ownerfacade.FinalizeCapturedTx(ctx, tx, s.revisionAppender, ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        recordID,
		Operation:       operation,
		CreatedOrReused: createdOrReused,
		OwnerResultCode: resultCode,
		BeforeVersionID: beforeVersionID,
		BeforeValue:     beforeValue,
		BeforeSnapshot:  beforeSnapshot,
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
