package tasksdecisions

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func (f *MutationFacade) applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) ([]links.RecordLinkMutation, error) {
	fieldKeys := make([]string, 0, len(collections))
	for fieldKey := range collections {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	mutations := make([]links.RecordLinkMutation, 0)
	for _, fieldKey := range fieldKeys {
		_, applied, err := f.applyCollectionPayloadTx(ctx, tx, viewSchemaID, incidentID, recordID, actorID, fieldKey, collections[fieldKey], now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func validateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore LinkCapability, viewSchemaID string, incidentID uuid.UUID, fieldKey string, payload CollectionActionPayload) error {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok || descriptor.ViewSchemaID != viewSchemaID {
		return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	command, err := recordRefValidation(incidentID, descriptor, payload)
	if err != nil {
		return err
	}
	return linkStore.ValidateRecordRefCollectionTx(ctx, tx, command)
}

func (f *MutationFacade) applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload CollectionActionPayload, now time.Time) (bool, []links.RecordLinkMutation, error) {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok || descriptor.ViewSchemaID != viewSchemaID {
		return false, nil, &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	adds, removes, err := recordRefActions(descriptor, payload)
	if err != nil {
		return false, nil, err
	}
	result, err := f.linkStore.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID:         incidentID,
		SourceRecordID:     recordID,
		ActorUserID:        actorID,
		FieldKey:           descriptor.FieldKey,
		LinkType:           links.LinkType(descriptor.LinkType),
		ExpectedTargetType: descriptor.ExpectedTargetType,
		AddRecordIDs:       adds,
		RemoveRecordIDs:    removes,
		Now:                now,
	})
	if err != nil {
		return false, nil, err
	}
	return len(result.RecordLinks) > 0, result.RecordLinks, nil
}

func (f *MutationFacade) appendRecordLinkMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequence int, mutations []links.RecordLinkMutation) error {
	for index, mutation := range mutations {
		if err := f.revisions.AppendMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    startSequence + index,
			TargetKind:    "record_link",
			TargetID:      mutation.RecordLinkID.String(),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return err
		}
	}
	return nil
}

type collectionDescriptor struct {
	ViewSchemaID       string
	FieldKey           string
	LinkType           string
	ExpectedTargetType string
	AllowedOperations  []string
}

func isRecordRefCollectionField(fieldKey string) bool {
	_, ok := lookupCollectionDescriptor(fieldKey)
	return ok
}

func allowsCollectionOp(fieldKey string, op string) bool {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok {
		return false
	}
	return slices.Contains(descriptor.AllowedOperations, op)
}

func lookupCollectionDescriptor(fieldKey string) (collectionDescriptor, bool) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return collectionDescriptor{}, false
	}
	field, ok := catalog.Field(fieldKey)
	if !ok || field.Kind != sourcecatalog.FieldKindCollection || field.Collection.Family != "record_ref" {
		return collectionDescriptor{}, false
	}
	return collectionDescriptor{
		ViewSchemaID: field.ViewSchemaID, FieldKey: field.FieldKey, LinkType: field.Collection.LinkType,
		ExpectedTargetType: field.Collection.ExpectedTargetRecordType,
		AllowedOperations:  slices.Clone(field.Collection.AllowedOperations),
	}, true
}

func recordRefValidation(incidentID uuid.UUID, descriptor collectionDescriptor, payload CollectionActionPayload) (links.RecordRefCollectionValidation, error) {
	adds, removes, err := recordRefActions(descriptor, payload)
	return links.RecordRefCollectionValidation{IncidentID: incidentID, FieldKey: descriptor.FieldKey, LinkType: links.LinkType(descriptor.LinkType), ExpectedTargetType: descriptor.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes}, err
}

func recordRefActions(descriptor collectionDescriptor, payload CollectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, nil, &links.CollectionValidationError{Field: descriptor.FieldKey, ReasonCode: "invalid_value"}
			}
			adds = append(adds, *action.LinkedRecordID)
		case "remove_record_ref":
			recordID, err := links.ParseRecordRefItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, &links.CollectionValidationError{Field: descriptor.FieldKey, ReasonCode: "invalid_value"}
			}
			removes = append(removes, recordID)
		default:
			return nil, nil, &links.CollectionValidationError{Field: descriptor.FieldKey, ReasonCode: "invalid_value"}
		}
	}
	return adds, removes, nil
}
