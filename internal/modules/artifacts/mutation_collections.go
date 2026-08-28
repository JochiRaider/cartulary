package artifacts

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func validateArtifactReferencesTx(ctx context.Context, tx pgx.Tx, members MemberReferenceCapability, linkStore LinkCapability, incidentID uuid.UUID, viewSchemaID string, values map[string]fieldValue, collections map[string]collectionActionPayload) error {
	for fieldKey, value := range values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := members.ValidateIncidentMemberUserTx(ctx, tx, incidentID, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range collections {
		if err := validateArtifactCollectionPayloadTx(ctx, tx, linkStore, incidentID, viewSchemaID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func validateArtifactPatchReferencesTx(ctx context.Context, tx pgx.Tx, members MemberReferenceCapability, linkStore LinkCapability, incidentID uuid.UUID, request patchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := members.ValidateIncidentMemberUserTx(ctx, tx, incidentID, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateArtifactCollectionPayloadTx(ctx, tx, linkStore, incidentID, request.ViewSchemaID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateArtifactCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore LinkCapability, incidentID uuid.UUID, viewSchemaID string, fieldKey string, payload collectionActionPayload) error {
	sourcePolicy, ok := lookupArtifactSourceField(fieldKey)
	if !ok || sourcePolicy.ViewSchemaID != viewSchemaID ||
		sourcePolicy.Kind != sourcecatalog.FieldKindCollection || (!sourcePolicy.View.Writable && !sourcePolicy.View.CreateWritable) {
		return collectionValidationError(fieldKey)
	}
	policy := collectionPolicyFromCatalogField(sourcePolicy)
	if policy.allowsRiskRefs() {
		return validateHandoffRiskRefPayload(riskRefPayloadFromWorkbook(payload))
	}
	switch {
	case policy.allowsRecordRefs():
		command, err := artifactRecordRefValidation(incidentID, policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidateRecordRefCollectionTx(ctx, tx, command)
	case policy.allowsPartyRefs():
		command, err := artifactPartyRefValidation(incidentID, policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidatePartyRefCollectionTx(ctx, tx, command)
	case policy.allowsTags():
		command, err := artifactTagValidation(policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidateTagCollectionTx(ctx, tx, command)
	default:
		return collectionValidationError(fieldKey)
	}
}

func (f *MutationFacade) applyCollectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, collections map[string]collectionActionPayload, now time.Time) (links.CollectionMutationResult, error) {
	fieldKeys := make([]string, 0, len(collections))
	for fieldKey := range collections {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	mutations := links.CollectionMutationResult{}
	for _, fieldKey := range fieldKeys {
		_, collectionResult, err := f.applyCollectionTx(ctx, tx, incidentID, recordID, actorID, viewSchemaID, fieldKey, collections[fieldKey], now)
		if err != nil {
			return links.CollectionMutationResult{}, err
		}
		mutations.RecordLinks = append(mutations.RecordLinks, collectionResult.RecordLinks...)
		mutations.RecordTags = append(mutations.RecordTags, collectionResult.RecordTags...)
	}
	return mutations, nil
}

func (f *MutationFacade) applyCollectionTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, fieldKey string, payload collectionActionPayload, now time.Time) (bool, links.CollectionMutationResult, error) {
	sourcePolicy, ok := lookupArtifactSourceField(fieldKey)
	if !ok || sourcePolicy.ViewSchemaID != viewSchemaID ||
		sourcePolicy.Kind != sourcecatalog.FieldKindCollection || (!sourcePolicy.View.Writable && !sourcePolicy.View.CreateWritable) {
		return false, links.CollectionMutationResult{}, collectionValidationError(fieldKey)
	}
	policy := collectionPolicyFromCatalogField(sourcePolicy)
	if policy.allowsRiskRefs() {
		changed, err := f.source.rows.applyHandoffRiskRefPayloadTx(ctx, tx, f.recordEnvelopes, incidentID, recordID, actorID, riskRefPayloadFromWorkbook(payload), now)
		return changed, links.CollectionMutationResult{}, err
	}
	switch {
	case policy.allowsRecordRefs():
		command, err := artifactRecordRefCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		result, err := f.linkStore.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, command)
		return len(result.RecordLinks) > 0, result, err
	case policy.allowsPartyRefs():
		command, err := artifactPartyRefCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		result, err := f.linkStore.ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, command)
		return len(result.RecordLinks) > 0, result, err
	case policy.allowsTags():
		command, err := artifactTagCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		result, err := f.linkStore.ApplyTagCollectionWithMutationValuesTx(ctx, tx, command)
		return len(result.RecordTags) > 0, result, err
	default:
		return false, links.CollectionMutationResult{}, collectionValidationError(fieldKey)
	}
}

func (f *MutationFacade) appendCollectionMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequence int, mutations links.CollectionMutationResult) (int, error) {
	sequence := startSequence
	for _, mutation := range mutations.RecordLinks {
		if err := f.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID: changeSetID, SequenceNo: sequence, TargetKind: "record_link",
			TargetID: mutation.RecordLinkID.String(), OperationKind: mutation.Operation,
			BeforeValue: mutation.BeforeValue, AfterValue: mutation.AfterValue,
		}); err != nil {
			return sequence, err
		}
		sequence++
	}
	for _, mutation := range mutations.RecordTags {
		if err := f.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID: changeSetID, SequenceNo: sequence, TargetKind: "record_tag",
			TargetID: links.RecordTagItemRef(mutation.RecordID, mutation.RecordTagID), OperationKind: mutation.Operation,
			BeforeValue: mutation.BeforeValue, AfterValue: mutation.AfterValue,
		}); err != nil {
			return sequence, err
		}
		sequence++
	}
	return sequence, nil
}

func artifactRecordRefValidation(incidentID uuid.UUID, policy collectionPolicy, payload collectionActionPayload) (links.RecordRefCollectionValidation, error) {
	adds, removes, err := artifactRecordRefActions(policy, payload)
	return links.RecordRefCollectionValidation{IncidentID: incidentID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes}, err
}

func artifactRecordRefCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy collectionPolicy, payload collectionActionPayload, now time.Time) (links.RecordRefCollectionCommand, error) {
	adds, removes, err := artifactRecordRefActions(policy, payload)
	return links.RecordRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes, Now: now}, err
}

func artifactRecordRefActions(policy collectionPolicy, payload collectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		if !policy.allowsOp(action.Op) {
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			adds = append(adds, *action.LinkedRecordID)
		case "remove_record_ref":
			recordID, err := links.ParseRecordRefItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			removes = append(removes, recordID)
		default:
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
	}
	return adds, removes, nil
}

func artifactPartyRefValidation(incidentID uuid.UUID, policy collectionPolicy, payload collectionActionPayload) (links.PartyRefCollectionValidation, error) {
	adds, removes, err := artifactPartyRefActions(policy, payload)
	return links.PartyRefCollectionValidation{IncidentID: incidentID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddPartyIDs: adds, RemovePartyIDs: removes}, err
}

func artifactPartyRefCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy collectionPolicy, payload collectionActionPayload, now time.Time) (links.PartyRefCollectionCommand, error) {
	adds, removes, err := artifactPartyRefActions(policy, payload)
	return links.PartyRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddPartyIDs: adds, RemovePartyIDs: removes, Now: now}, err
}

func artifactPartyRefActions(policy collectionPolicy, payload collectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		if !policy.allowsOp(action.Op) {
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
		switch action.Op {
		case "add_party_ref":
			if action.PartyID == nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			adds = append(adds, *action.PartyID)
		case "remove_party_ref":
			partyID, err := links.ParsePartyRefItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			removes = append(removes, partyID)
		default:
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
	}
	return adds, removes, nil
}

func artifactTagValidation(policy collectionPolicy, payload collectionActionPayload) (links.TagCollectionValidation, error) {
	adds, removes, err := artifactTagActions(policy, payload)
	return links.TagCollectionValidation{FieldKey: policy.FieldKey, AddTags: adds, RemoveTags: removes}, err
}

func artifactTagCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy collectionPolicy, payload collectionActionPayload, now time.Time) (links.TagCollectionCommand, error) {
	adds, removes, err := artifactTagActions(policy, payload)
	return links.TagCollectionCommand{IncidentID: incidentID, RecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, AddTags: adds, RemoveTags: removes, Now: now}, err
}

func artifactTagActions(policy collectionPolicy, payload collectionActionPayload) ([]links.TagCollectionAdd, []links.RecordTagRef, error) {
	adds := make([]links.TagCollectionAdd, 0)
	removes := make([]links.RecordTagRef, 0)
	for _, action := range payload.Actions {
		if !policy.allowsOp(action.Op) {
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
		switch action.Op {
		case "add_tag":
			adds = append(adds, links.TagCollectionAdd{RawText: action.RawText, NormalizedText: action.NormalizedText})
		case "remove_tag":
			recordID, tagID, err := links.ParseRecordTagItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			removes = append(removes, links.RecordTagRef{RecordID: recordID, RecordTagID: tagID})
		default:
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
	}
	return adds, removes, nil
}

func riskRefPayloadFromWorkbook(payload collectionActionPayload) riskRefActionPayload {
	actions := make([]riskRefAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, riskRefAction{
			Op:             action.Op,
			ItemRef:        action.ItemRef,
			RiskRefText:    action.RiskRefText,
			NormalizedText: action.NormalizedText,
		})
	}
	return riskRefActionPayload{Actions: actions}
}

func collectionValidationError(fieldKey string) *links.CollectionValidationError {
	return &links.CollectionValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
}
