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

	switch request.TargetViewSchemaID {
	case entitycontract.HostsViewSchemaID:
		record, before, operation, _, snapshot, err := s.upsertHostTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshHostTx(ctx, tx, record.RecordID); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return s.finalizeHostImportRowTx(
			ctx, tx, command, record.RecordID, record.RowVersion,
			before, snapshot, buildHostRow(record), operation,
		)
	case entitycontract.IdentitiesViewSchemaID:
		record, before, operation, _, snapshot, err := s.upsertIdentityTx(ctx, tx, actor, request.IncidentID, createRequest, now)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.ports.projections.RefreshIdentityTx(ctx, tx, record.RecordID); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return s.finalizeIdentityImportRowTx(
			ctx, tx, command, record.RecordID, record.RowVersion,
			before, snapshot, buildIdentityRow(record), operation,
		)
	default:
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("entity import surface %q not mapped", request.TargetViewSchemaID)
	}
}

func (s *importOwner) finalizeHostImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ownerfacade.ImportOwnerCreateCommand,
	recordID uuid.UUID,
	rowVersion int64,
	beforeRow map[string]any,
	beforeSnapshot *revisions.RecordSnapshot,
	afterRow map[string]any,
	operationKind string,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	afterSnapshot, err := s.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	operation := operationKind
	createdOrReused := "created"
	resultCode := "created"
	var beforeVersionID *string
	if beforeRow != nil {
		createdOrReused = "reused"
		resultCode = "reused"
		beforeVersion := rowVersion
		if reflect.DeepEqual(beforeRow, afterRow) {
			operation = "reuse"
		} else {
			resultCode = "updated"
			if rowVersion > 1 {
				beforeVersion = rowVersion - 1
			}
		}
		value := entityVersionID("host", recordID, beforeVersion)
		beforeVersionID = &value
	}
	if operation == "" {
		operation = "host_import_create"
	}
	afterVersionID := entityVersionID("host", recordID, rowVersion)
	if err := s.revisionAppender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		TargetKind:      "host",
		RecordID:        recordID,
		OperationKind:   operation,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
		changedFields := entityChangedFieldKeys(beforeRow, afterRow)
		if err := s.revisionAppender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID:    command.ChangeSetID,
			RecordID:       recordID,
			RowVersion:     rowVersion,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			ConflictFacts:  entityRevisionFacts(beforeRow, afterRow, changedFields),
		}); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		request := command.Request
		if err := s.appendRecordChangedTx(
			ctx, tx, request.IncidentID, request.ActorUserID,
			request.ClientTxnID, command.ChangeSetID, recordID, rowVersion,
			max(command.SequenceNo-1, 0), command.Now,
			entitycontract.HostsViewSchemaID, afterRow, changedFields,
		); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:             recordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      createdOrReused,
		OwnerResultCode:      resultCode,
		RowRefresh:           afterRow,
	}, nil
}

func (s *importOwner) finalizeIdentityImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ownerfacade.ImportOwnerCreateCommand,
	recordID uuid.UUID,
	rowVersion int64,
	beforeRow map[string]any,
	beforeSnapshot *revisions.RecordSnapshot,
	afterRow map[string]any,
	operationKind string,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	afterSnapshot, err := s.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	operation := operationKind
	createdOrReused := "created"
	resultCode := "created"
	var beforeVersionID *string
	if beforeRow != nil {
		createdOrReused = "reused"
		resultCode = "reused"
		beforeVersion := rowVersion
		if reflect.DeepEqual(beforeRow, afterRow) {
			operation = "reuse"
		} else {
			resultCode = "updated"
			if rowVersion > 1 {
				beforeVersion = rowVersion - 1
			}
		}
		value := entityVersionID("identity", recordID, beforeVersion)
		beforeVersionID = &value
	}
	if operation == "" {
		operation = "identity_import_create"
	}
	afterVersionID := entityVersionID("identity", recordID, rowVersion)
	if err := s.revisionAppender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		TargetKind:      "identity",
		RecordID:        recordID,
		OperationKind:   operation,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
		changedFields := entityChangedFieldKeys(beforeRow, afterRow)
		if err := s.revisionAppender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID:    command.ChangeSetID,
			RecordID:       recordID,
			RowVersion:     rowVersion,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			ConflictFacts:  entityRevisionFacts(beforeRow, afterRow, changedFields),
		}); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		request := command.Request
		if err := s.appendRecordChangedTx(
			ctx, tx, request.IncidentID, request.ActorUserID,
			request.ClientTxnID, command.ChangeSetID, recordID, rowVersion,
			max(command.SequenceNo-1, 0), command.Now,
			entitycontract.IdentitiesViewSchemaID, afterRow, changedFields,
		); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:             recordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      createdOrReused,
		OwnerResultCode:      resultCode,
		RowRefresh:           afterRow,
	}, nil
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
