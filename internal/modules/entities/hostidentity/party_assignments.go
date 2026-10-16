package hostidentity

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	partyreferences "github.com/JochiRaider/cartulary/internal/modules/parties/references"
)

// collectionMutation keeps each source owner's actual history target identity.
type collectionMutation struct {
	targetKind, targetID, operationKind string
	beforeValue, afterValue             any
}

type PartyReference struct {
	PartyID     uuid.UUID
	DisplayText string
}

func applyPartyAssignmentActionsTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID, field string, actions []CollectionAction, actorID uuid.UUID, now time.Time) ([]collectionMutation, error) {
	command := links.PartyRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", Now: now}
	for _, action := range actions {
		switch action.Op {
		case "add_party_ref":
			if action.PartyID == uuid.Nil {
				return nil, invalidMutationPayload(field, "invalid_value")
			}
			command.AddPartyIDs = append(command.AddPartyIDs, action.PartyID)
		case "remove_party_ref":
			id, err := links.ParsePartyRefItemRef(action.ItemRef)
			if err != nil || id == uuid.Nil {
				return nil, invalidMutationPayload(field, "invalid_value")
			}
			command.RemovePartyIDs = append(command.RemovePartyIDs, id)
		default:
			return nil, invalidMutationPayload(field, "invalid_value")
		}
	}
	result, err := links.NewStore().ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, command)
	if err != nil {
		var invalid *links.CollectionValidationError
		if errors.As(err, &invalid) {
			return nil, invalidMutationPayload(field, "invalid_value")
		}
		return nil, err
	}
	mutations := make([]collectionMutation, 0, len(result.Mutations()))
	for _, m := range result.Mutations() {
		mutations = append(mutations, collectionMutation{targetKind: m.TargetKind(), targetID: m.TargetID(), operationKind: m.OperationKind(), beforeValue: m.BeforeValue(), afterValue: m.AfterValue()})
	}
	return mutations, nil
}

func applyPartyAssignmentAddsTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID, adds map[string][]CollectionAction, actorID uuid.UUID, now time.Time) ([]collectionMutation, error) {
	keys := make([]string, 0, len(adds))
	for key := range adds {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	var mutations []collectionMutation
	for _, key := range keys {
		for _, action := range adds[key] {
			if action.Op != "add_party_ref" {
				return nil, invalidMutationPayload(key, "invalid_value")
			}
		}
		applied, err := applyPartyAssignmentActionsTx(ctx, tx, incidentID, recordID, key, adds[key], actorID, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func loadPartyAssignmentReferences(ctx context.Context, query entityAliasQueryer, incidentID uuid.UUID, recordIDs []uuid.UUID) (map[uuid.UUID]map[string][]PartyReference, error) {
	facts, err := (links.FactReader{}).LoadPartyAssignmentsForRecords(ctx, query, incidentID, recordIDs)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(facts))
	for _, fact := range facts {
		ids = append(ids, fact.PartyID)
	}
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	ids = slices.Compact(ids)
	labels, err := partyreferences.LoadLabels(ctx, query, incidentID, ids)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]map[string][]PartyReference)
	for _, fact := range facts {
		label, exists := labels[fact.PartyID]
		if !exists {
			return nil, errors.New("party assignment label unavailable")
		}
		fields := result[fact.SubjectRecordID]
		if fields == nil {
			fields = make(map[string][]PartyReference)
			result[fact.SubjectRecordID] = fields
		}
		field := fact.SubjectRecordType + "." + fact.AssignmentKind + "_ids"
		fields[field] = append(fields[field], PartyReference{PartyID: fact.PartyID, DisplayText: label})
	}
	for _, fields := range result {
		for key := range fields {
			slices.SortFunc(fields[key], func(a, b PartyReference) int {
				if order := strings.Compare(a.DisplayText, b.DisplayText); order != 0 {
					return order
				}
				return strings.Compare(a.PartyID.String(), b.PartyID.String())
			})
		}
	}
	return result, nil
}

func partyAssignmentCollectionValue(refs []PartyReference) any {
	items := make([]map[string]any, 0, len(refs))
	for _, ref := range refs {
		items = append(items, map[string]any{"item_ref": links.PartyRefItemRef(ref.PartyID), "item_kind": "party_ref", "party_id": ref.PartyID.String(), "display_text": ref.DisplayText})
	}
	return collectionValue(false, items)
}

func (s *mutationCore) appendCollectionMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, sequence int, mutations []collectionMutation) error {
	for index, m := range mutations {
		if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{ChangeSetID: changeSetID, SequenceNo: sequence + index, TargetKind: m.targetKind, TargetID: m.targetID, OperationKind: m.operationKind, BeforeValue: m.beforeValue, AfterValue: m.afterValue}); err != nil {
			return err
		}
	}
	return nil
}
