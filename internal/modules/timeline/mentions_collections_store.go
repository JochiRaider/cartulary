package timeline

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *store) hydrateProjectedCollections(ctx context.Context, tx pgx.Tx, record *workbookprojection.DerivedRecord) error {
	facts, err := s.collectionReader.LoadTimelineCollectionFactsTx(ctx, tx, record.IncidentID, record.RecordID)
	if err != nil {
		return err
	}
	workbookprojection.ApplyCollectionFacts(record, facts)
	return nil
}

type mentionInsertOptions struct {
	allowInteractiveAutoResolution bool
	originKind                     string
}

type mentionProjectionRefresh struct {
	Hosts      map[uuid.UUID]struct{}
	Identities map[uuid.UUID]struct{}
}

type mentionApplicationResult struct {
	Projection    mentionProjectionRefresh
	LinkMutations []attachedEvidenceMutation
}

func (r *mentionApplicationResult) merge(other mentionApplicationResult) {
	r.Projection.merge(other.Projection)
	r.LinkMutations = append(r.LinkMutations, other.LinkMutations...)
}

func (s *store) refreshMentionEntityProjectionsTx(ctx context.Context, tx pgx.Tx, refresh mentionProjectionRefresh) error {
	hostIDs := mapKeys(refresh.Hosts)
	identityIDs := mapKeys(refresh.Identities)
	slices.SortFunc(hostIDs, compareUUIDs)
	slices.SortFunc(identityIDs, compareUUIDs)
	for _, recordID := range hostIDs {
		if err := s.entityProjections.RefreshHostTx(ctx, tx, recordID); err != nil {
			return err
		}
	}
	for _, recordID := range identityIDs {
		if err := s.entityProjections.RefreshIdentityTx(ctx, tx, recordID); err != nil {
			return err
		}
	}
	return nil
}

func (r *mentionProjectionRefresh) include(fieldKey string, recordID uuid.UUID) {
	switch collectionEntityType(fieldKey) {
	case "host":
		if r.Hosts == nil {
			r.Hosts = make(map[uuid.UUID]struct{})
		}
		r.Hosts[recordID] = struct{}{}
	case "identity":
		if r.Identities == nil {
			r.Identities = make(map[uuid.UUID]struct{})
		}
		r.Identities[recordID] = struct{}{}
	}
}

func (r *mentionProjectionRefresh) merge(other mentionProjectionRefresh) {
	for recordID := range other.Hosts {
		r.include("timeline.host_refs", recordID)
	}
	for recordID := range other.Identities {
		r.include("timeline.identity_refs", recordID)
	}
}

func mapKeys(values map[uuid.UUID]struct{}) []uuid.UUID {
	keys := make([]uuid.UUID, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func compareUUIDs(left uuid.UUID, right uuid.UUID) int {
	return strings.Compare(left.String(), right.String())
}

func (s *store) applyCreateMentionActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, hostRefs *CollectionActionPayload, identityRefs *CollectionActionPayload, options createRowOptions, now time.Time) (mentionApplicationResult, error) {
	var result mentionApplicationResult
	insertOptions := mentionInsertOptions{
		allowInteractiveAutoResolution: options.allowInteractiveAutoResolution,
	}
	hostResult, err := s.insertMentionActionsTx(ctx, tx, s.linkStore, actorUserID, incidentID, recordID, "timeline.host_refs", collectionEntityType("timeline.host_refs"), hostRefs, insertOptions, now)
	if err != nil {
		return mentionApplicationResult{}, err
	}
	result.merge(hostResult)
	identityResult, err := s.insertMentionActionsTx(ctx, tx, s.linkStore, actorUserID, incidentID, recordID, "timeline.identity_refs", collectionEntityType("timeline.identity_refs"), identityRefs, insertOptions, now)
	if err != nil {
		return mentionApplicationResult{}, err
	}
	result.merge(identityResult)
	return result, nil
}

func (s *store) applyCreateTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, tags *CollectionActionPayload, now time.Time) ([]recordTagMutation, error) {
	return s.applyTimelineTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, tags, now)
}

func (s *store) applyPatchMentionActionsTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) (mentionApplicationResult, error) {
	var result mentionApplicationResult
	for _, change := range changes {
		if change.ActionPayload == nil {
			continue
		}
		entityType := collectionEntityType(change.FieldKey)
		for _, action := range change.ActionPayload.Actions {
			switch action.Op {
			case "add_token", "add_resolved_ref":
				applied, err := s.insertMentionActionsTx(ctx, tx, s.linkStore, actor.ID, incidentID, recordID, change.FieldKey, entityType, &CollectionActionPayload{Actions: []CollectionAction{action}}, mentionInsertOptions{
					allowInteractiveAutoResolution: true,
				}, now)
				if err != nil {
					return mentionApplicationResult{}, err
				}
				result.merge(applied)
			case "resolve_item":
				mentionID, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return mentionApplicationResult{}, err
				}
				mutations, err := s.mentionStore.ResolveExistingFromMentionTx(ctx, tx, actor, recordID, change.FieldKey, mentionID, action.ResolvedRecord, now)
				if err != nil {
					return mentionApplicationResult{}, err
				}
				for _, mutation := range mutations {
					result.LinkMutations = append(result.LinkMutations, attachedEvidenceMutation(mutation))
				}
			case "dismiss_item", "revert_to_unresolved":
				mentionID, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return mentionApplicationResult{}, err
				}
				mutations, err := s.mentionStore.ApplyMentionLifecycleTx(ctx, tx, actor, recordID, change.FieldKey, mentionID, action.Op, nil, now)
				if err != nil {
					return mentionApplicationResult{}, err
				}
				for _, mutation := range mutations {
					result.LinkMutations = append(result.LinkMutations, attachedEvidenceMutation(mutation))
				}
			default:
				return mentionApplicationResult{}, fmt.Errorf("unsupported mention action: %s", action.Op)
			}
		}
	}
	return result, nil
}

func (s *store) applyPatchTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) ([]recordTagMutation, error) {
	mutations := make([]recordTagMutation, 0)
	for _, change := range changes {
		if !isTimelineTagCollection(change.FieldKey) || change.ActionPayload == nil {
			continue
		}
		applied, err := s.applyTimelineTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func (s *store) applyPatchAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) ([]attachedEvidenceMutation, error) {
	var mutations []attachedEvidenceMutation
	for _, change := range changes {
		if !isTimelineAttachedEvidenceCollection(change.FieldKey) || change.ActionPayload == nil {
			continue
		}
		applied, err := s.applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, incidentID, recordID, change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func (s *store) applyAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, payload *CollectionActionPayload, now time.Time) ([]attachedEvidenceMutation, error) {
	if payload == nil || len(payload.Actions) == 0 {
		return nil, nil
	}
	policy, ok := LookupCollectionPolicy("timeline.attached_evidence_ids")
	if !ok || !policy.AllowsLinksCollectionMutation() {
		return nil, fmt.Errorf("missing collection policy: timeline.attached_evidence_ids")
	}
	command, err := timelineRecordRefCollectionCommand(incidentID, recordID, actorUserID, policy, payload, now)
	if err != nil {
		return nil, err
	}
	if err := s.evidenceStore.ValidateTimelineAttachmentsTx(ctx, tx, incidentID, command.AddRecordIDs); err != nil {
		return nil, err
	}
	result, err := s.linkStore.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, command)
	if err != nil {
		return nil, err
	}
	affectedRecordIDs := append([]uuid.UUID{}, command.AddRecordIDs...)
	affectedRecordIDs = append(affectedRecordIDs, command.RemoveRecordIDs...)
	if err := s.evidenceStore.RefreshTimelineAttachmentProjectionsTx(ctx, tx, affectedRecordIDs); err != nil {
		return nil, err
	}
	mutations := make([]attachedEvidenceMutation, 0, len(result.RecordLinks))
	for _, mutation := range result.RecordLinks {
		mutations = append(mutations, attachedEvidenceMutation(mutation))
	}
	return mutations, nil
}

func (s *store) insertAttachedEvidenceMutationEntriesTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequenceNo int, mutations []attachedEvidenceMutation) error {
	sequenceNo := startSequenceNo
	for _, mutation := range mutations {
		if mutation.RecordLinkID == uuid.Nil {
			continue
		}
		if err := s.revisionsStore.AppendMutationTx(ctx, tx, MutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    sequenceNo,
			TargetKind:    "record_link",
			TargetID:      mutation.RecordLinkID.String(),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return err
		}
		sequenceNo++
	}
	return nil
}

func (s *store) applyTimelineTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, payload *CollectionActionPayload, now time.Time) ([]recordTagMutation, error) {
	if payload == nil || len(payload.Actions) == 0 {
		return nil, nil
	}
	policy, ok := LookupCollectionPolicy("timeline.tags")
	if !ok || !policy.AllowsLinksCollectionMutation() {
		return nil, fmt.Errorf("missing collection policy: timeline.tags")
	}
	command, err := timelineTagCollectionCommand(incidentID, recordID, actorUserID, policy, payload, now)
	if err != nil {
		return nil, err
	}
	result, err := s.linkStore.ApplyTagCollectionWithMutationValuesTx(ctx, tx, command)
	if err != nil {
		return nil, err
	}
	mutations := make([]recordTagMutation, 0, len(result.RecordTags))
	for _, mutation := range result.RecordTags {
		mutations = append(mutations, recordTagMutation(mutation))
	}
	return mutations, nil
}

func (s *store) insertRecordTagMutationEntriesTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequenceNo int, mutations []recordTagMutation) error {
	sequenceNo := startSequenceNo
	for _, mutation := range mutations {
		if mutation.RecordTagID == uuid.Nil || mutation.RecordID == uuid.Nil {
			continue
		}
		if err := s.revisionsStore.AppendMutationTx(ctx, tx, MutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    sequenceNo,
			TargetKind:    "record_tag",
			TargetID:      linkRecordTagItemRef(mutation.RecordID, mutation.RecordTagID),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return err
		}
		sequenceNo++
	}
	return nil
}

func timelineRecordRefCollectionCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionPolicy, payload *CollectionActionPayload, now time.Time) (RecordRefCollectionCommand, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		if !policy.AllowsOp(action.Op) {
			return RecordRefCollectionCommand{}, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return RecordRefCollectionCommand{}, fmt.Errorf("missing linked record id")
			}
			adds = append(adds, *action.LinkedRecordID)
		case "remove_record_ref":
			recordID, err := parseRecordRefItemRef(action.ItemRef)
			if err != nil {
				return RecordRefCollectionCommand{}, err
			}
			removes = append(removes, recordID)
		default:
			return RecordRefCollectionCommand{}, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	return RecordRefCollectionCommand{
		IncidentID:         incidentID,
		SourceRecordID:     recordID,
		ActorUserID:        actorID,
		FieldKey:           policy.FieldKey,
		LinkType:           policy.LinkType,
		ExpectedTargetType: policy.ExpectedTargetType,
		AddRecordIDs:       adds,
		RemoveRecordIDs:    removes,
		Now:                now,
	}, nil
}

func timelineTagCollectionCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionPolicy, payload *CollectionActionPayload, now time.Time) (TagCollectionCommand, error) {
	adds := make([]TagCollectionAdd, 0)
	removes := make([]RecordTagRef, 0)
	for _, action := range payload.Actions {
		if !policy.AllowsOp(action.Op) {
			return TagCollectionCommand{}, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
		switch action.Op {
		case "add_tag":
			adds = append(adds, TagCollectionAdd{RawText: action.RawText, NormalizedText: action.NormalizedText})
		case "remove_tag":
			itemRecordID, tagID, err := parseRecordTagItemRef(action.ItemRef)
			if err != nil {
				return TagCollectionCommand{}, err
			}
			removes = append(removes, RecordTagRef{RecordID: itemRecordID, RecordTagID: tagID})
		default:
			return TagCollectionCommand{}, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	return TagCollectionCommand{
		IncidentID:  incidentID,
		RecordID:    recordID,
		ActorUserID: actorID,
		FieldKey:    policy.FieldKey,
		AddTags:     adds,
		RemoveTags:  removes,
		Now:         now,
	}, nil
}

func (s *store) insertMentionActionsTx(ctx context.Context, tx pgx.Tx, linkStore LinkPort, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, fieldKey string, entityType string, payload *CollectionActionPayload, options mentionInsertOptions, now time.Time) (mentionApplicationResult, error) {
	if payload == nil || len(payload.Actions) == 0 {
		return mentionApplicationResult{}, nil
	}
	originKind := options.originKind
	if strings.TrimSpace(originKind) == "" {
		originKind = "interactive_cell"
	}
	nextOrdinal, err := s.mentionStore.NextOrdinalTx(ctx, tx, recordID, fieldKey)
	if err != nil {
		return mentionApplicationResult{}, err
	}
	var result mentionApplicationResult
	for _, action := range payload.Actions {
		if action.Op != "add_token" && action.Op != "add_resolved_ref" {
			return mentionApplicationResult{}, fmt.Errorf("unsupported mention action: %s", action.Op)
		}
		resolutionStatus := "unresolved"
		var resolvedRecordID *uuid.UUID
		var resolvedByUserID *uuid.UUID
		var resolvedAt *time.Time
		var resolutionMethod *string
		relationshipProvenance := "manual"
		var linkConfidence *int
		if action.Op == "add_resolved_ref" {
			if action.ResolvedRecord == nil {
				return mentionApplicationResult{}, fmt.Errorf("missing resolved record for action: %s", action.Op)
			}
			if err := s.entityStore.ValidateResolvedTargetTx(ctx, tx, incidentID, entityType, *action.ResolvedRecord); err != nil {
				return mentionApplicationResult{}, err
			}
			resolutionStatus = "resolved"
			resolvedRecordID = action.ResolvedRecord
			resolvedBy := actorUserID
			resolvedByUserID = &resolvedBy
			resolvedAtValue := now.UTC()
			resolvedAt = &resolvedAtValue
			method := action.Op
			resolutionMethod = &method
		} else if options.allowInteractiveAutoResolution {
			match, err := s.lookupInteractiveAutoResolutionMatchTx(ctx, tx, incidentID, fieldKey, action.RawText)
			if err != nil {
				return mentionApplicationResult{}, err
			}
			if match != nil {
				resolutionStatus = "resolved"
				resolvedRecordID = &match.RecordID
				resolvedBy := actorUserID
				resolvedByUserID = &resolvedBy
				resolvedAtValue := now.UTC()
				resolvedAt = &resolvedAtValue
				method := autoResolutionMethod
				resolutionMethod = &method
				relationshipProvenance = autoResolutionMethod
				confidence := 100
				linkConfidence = &confidence
			}
		}
		if err := s.mentionStore.InsertTx(ctx, tx, MentionCreateParams{
			SourceRecordID:   recordID,
			EntityType:       entityType,
			SourceFieldKey:   fieldKey,
			OriginKind:       originKind,
			OriginLocator:    mentionOriginLocator(recordID, fieldKey, nextOrdinal),
			RawText:          action.RawText,
			NormalizedText:   action.NormalizedText,
			ResolutionStatus: resolutionStatus,
			Ordinal:          nextOrdinal,
			CreatedByUserID:  actorUserID,
			CreatedAt:        now.UTC(),
			ResolvedRecordID: resolvedRecordID,
			ResolvedByUserID: resolvedByUserID,
			ResolvedAt:       resolvedAt,
			ResolutionMethod: resolutionMethod,
		}); err != nil {
			return mentionApplicationResult{}, fmt.Errorf("insert entity mention: %w", err)
		}
		if resolvedRecordID != nil {
			relationshipType, ok := timelineRelationshipLinkType(fieldKey)
			if !ok {
				return mentionApplicationResult{}, fmt.Errorf("unsupported link field: %s", fieldKey)
			}
			linkResult, err := linkStore.UpsertLinkCommandTx(ctx, tx, UpsertLinkCommand{
				IncidentID:  incidentID,
				SrcRecordID: recordID,
				DstRecordID: *resolvedRecordID,
				LinkType:    relationshipType,
				Provenance:  relationshipProvenance,
				Confidence:  linkConfidence,
				OwnerUserID: actorUserID,
				Now:         now.UTC(),
			})
			if err != nil {
				return mentionApplicationResult{}, fmt.Errorf("upsert record link: %w", err)
			}
			if linkResult.Mutation != nil {
				result.LinkMutations = append(result.LinkMutations, attachedEvidenceMutation(*linkResult.Mutation))
			}
			result.Projection.include(fieldKey, *resolvedRecordID)
		}
		nextOrdinal++
	}
	return result, nil
}

func timelineRelationshipLinkType(fieldKey string) (string, bool) {
	policy, ok := LookupCollectionPolicy(fieldKey)
	if !ok || policy.Family != CollectionFamilyMentionOrigin || policy.LinkType == "" {
		return "", false
	}
	return policy.LinkType, true
}

func mentionOriginLocator(recordID uuid.UUID, fieldKey string, ordinal int) string {
	return fmt.Sprintf("view:%s/record:%s/cell:%s/item:%d", TimelineViewSchemaID, recordID.String(), fieldKey, ordinal)
}

func mentionIDFromItemRef(itemRef string) (uuid.UUID, error) {
	const prefix = "entity_mention:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid mention item_ref: %s", itemRef)
	}
	mentionID, err := uuid.Parse(strings.TrimPrefix(itemRef, prefix))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse mention item_ref: %w", err)
	}
	return mentionID, nil
}
