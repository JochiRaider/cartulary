package hostidentity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ImportDependencies struct {
	Revisions              *revisions.Appender
	ProjectionMutationRows projectionports.MutationRows
	Collaboration          collaboration.RecordChangedAppender
}

type importOwner struct {
	*mutationCore
}

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	dependencies ImportDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != entitycontract.HostsViewSchemaID && targetViewSchemaID != entitycontract.IdentitiesViewSchemaID {
		return nil, fmt.Errorf("entity import surface %q not mapped", targetViewSchemaID)
	}
	for _, dependency := range []struct {
		name  string
		value any
	}{
		{name: "Revisions", value: dependencies.Revisions},
		{name: "ProjectionMutationRows", value: dependencies.ProjectionMutationRows},
		{name: "Collaboration", value: dependencies.Collaboration},
	} {
		if isNilStoreDependency(dependency.value) {
			return nil, fmt.Errorf("compose Host/Identity import create facade: %s is required", dependency.name)
		}
	}
	owner := &importOwner{mutationCore: newMutationCore(dependencies.Revisions, dependencies.ProjectionMutationRows, dependencies.Collaboration)}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		owner.createImportRowTx,
	)
}

func (s *importOwner) createImportRowTx(ctx context.Context, tx pgx.Tx, command ownerfacade.ImportOwnerCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
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
		beforeSnapshot *revisions.RecordSnapshot
	)
	switch request.TargetViewSchemaID {
	case entitycontract.HostsViewSchemaID:
		record, before, operation, _, snapshot, err := s.upsertHostTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshHostTx(ctx, tx, record.RecordID); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID = record.RecordID
		rowVersion = record.RowVersion
		beforeRow = before
		beforeSnapshot = snapshot
		afterRow = buildHostRow(record)
		operationKind = operation
		entityType = "host"
	case entitycontract.IdentitiesViewSchemaID:
		record, before, operation, _, snapshot, err := s.upsertIdentityTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshIdentityTx(ctx, tx, record.RecordID); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID = record.RecordID
		rowVersion = record.RowVersion
		beforeRow = before
		beforeSnapshot = snapshot
		afterRow = buildIdentityRow(record)
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
	return ownerfacade.FinalizeLiveRecordTx(ctx, tx, s.revisionAppender, s.publications, ownerfacade.FinalizeCommand{
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
		CreatedAt:       now,
	})
}

func entityCreateRequestFromImport(clientTxnID string, fields []ownerfacade.ImportFieldValue) CreateRequest {
	request := CreateRequest{
		ClientTxnID: clientTxnID,
		Values:      make(map[string]string, len(fields)),
		AliasAdds:   make(map[string][]CollectionAction),
	}
	for _, field := range fields {
		value, ok := field.NormalizedValue.Text()
		if !ok {
			continue
		}
		switch field.FieldKey {
		case "host.aliases", "identity.aliases":
			request.AliasAdds[field.FieldKey] = append(request.AliasAdds[field.FieldKey], CollectionAction{
				Op:             "add_alias",
				RawText:        value,
				NormalizedText: value,
			})
		default:
			request.Values[field.FieldKey] = value
		}
	}
	return request
}
