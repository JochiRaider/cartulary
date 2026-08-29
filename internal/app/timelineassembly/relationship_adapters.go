package timelineassembly

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

type linkAdapter struct {
	store *links.Store
}

type mentionLinkAdapter struct {
	store *links.Store
}

func (a mentionLinkAdapter) UpsertMentionLinkTx(ctx context.Context, tx pgx.Tx, command mentions.LinkCommand) (mentions.LinkCommandResult, error) {
	linkType, err := links.ParseLinkType(string(command.LinkType))
	if err != nil {
		return mentions.LinkCommandResult{}, err
	}
	result, err := a.store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    linkType,
		Provenance:  links.LinkProvenanceManual,
		OwnerUserID: command.ActorUserID,
		Now:         command.Now,
	})
	if err != nil {
		return mentions.LinkCommandResult{}, err
	}
	return mentionLinkResult(result), nil
}

func (a mentionLinkAdapter) TombstoneActiveMentionLinkTx(ctx context.Context, tx pgx.Tx, command mentions.LinkCommand) (mentions.LinkCommandResult, bool, error) {
	linkType, err := links.ParseLinkType(string(command.LinkType))
	if err != nil {
		return mentions.LinkCommandResult{}, false, err
	}
	result, found, err := a.store.TombstoneActiveLinkCommandTx(ctx, tx, links.TombstoneActiveLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    linkType,
		ActorUserID: command.ActorUserID,
		Now:         command.Now,
	})
	if err != nil {
		return mentions.LinkCommandResult{}, false, err
	}
	if !found {
		return mentions.LinkCommandResult{}, false, nil
	}
	return mentionLinkResult(result), true, nil
}

func mentionLinkResult(result links.RecordLinkCommandResult) mentions.LinkCommandResult {
	converted := mentions.LinkCommandResult{
		RecordLinkID: result.RecordLinkID,
		SrcRecordID:  result.SrcRecordID,
		DstRecordID:  result.DstRecordID,
		LinkType:     mentions.LinkType(result.LinkType.String()),
	}
	if mutation, ok := result.Mutation(); ok {
		beforeValue, _ := mutation.BeforeValue().(map[string]any)
		afterValue, _ := mutation.AfterValue().(map[string]any)
		converted.Mutation = &mentions.LinkMutation{
			RecordLinkID: result.RecordLinkID,
			Operation:    mutation.OperationKind(),
			BeforeValue:  cloneLinkMutationMap(beforeValue),
			AfterValue:   cloneLinkMutationMap(afterValue),
		}
	}
	return converted
}

func cloneLinkMutationMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

func (a linkAdapter) InsertSupersedesCommandTx(ctx context.Context, tx pgx.Tx, command timeline.InsertSupersedesCommand) (timeline.RecordLinkCommandResult, error) {
	link, err := a.store.InsertSupersedesCommandTx(ctx, tx, links.InsertSupersedesCommand(command))
	if err != nil {
		return timeline.RecordLinkCommandResult{}, err
	}
	return timelineLinkResult(link), nil
}

func (a linkAdapter) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command timeline.UpsertLinkCommand) (timeline.RecordLinkCommandResult, error) {
	linkType, err := links.ParseLinkType(command.LinkType)
	if err != nil {
		return timeline.RecordLinkCommandResult{}, err
	}
	provenance, err := links.ParseLinkProvenance(command.Provenance)
	if err != nil {
		return timeline.RecordLinkCommandResult{}, err
	}
	result, err := a.store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    linkType,
		Provenance:  provenance,
		Confidence:  command.Confidence,
		OwnerUserID: command.OwnerUserID,
		Now:         command.Now,
	})
	if err != nil {
		return timeline.RecordLinkCommandResult{}, err
	}
	return timelineLinkResult(result), nil
}

func (a linkAdapter) HasActiveIncomingSupersedesLinkForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (bool, error) {
	return a.store.HasActiveIncomingSupersedesLinkForUpdateTx(ctx, tx, incidentID, recordID)
}

func (a linkAdapter) ApplyRecordRefCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command timeline.RecordRefCollectionCommand) (timeline.CollectionMutationResult, error) {
	linkType, err := links.ParseLinkType(command.LinkType)
	if err != nil {
		return timeline.CollectionMutationResult{}, err
	}
	result, err := a.store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID:         command.IncidentID,
		SourceRecordID:     command.SourceRecordID,
		ActorUserID:        command.ActorUserID,
		FieldKey:           command.FieldKey,
		LinkType:           linkType,
		ExpectedTargetType: command.ExpectedTargetType,
		AddRecordIDs:       command.AddRecordIDs,
		RemoveRecordIDs:    command.RemoveRecordIDs,
		Now:                command.Now,
	})
	if err != nil {
		return timeline.CollectionMutationResult{}, err
	}
	return collectionResult(result)
}

func timelineLinkResult(result links.RecordLinkCommandResult) timeline.RecordLinkCommandResult {
	converted := timeline.RecordLinkCommandResult{
		RecordLinkID: result.RecordLinkID,
		SrcRecordID:  result.SrcRecordID,
		DstRecordID:  result.DstRecordID,
		LinkType:     result.LinkType.String(),
	}
	if mutation, ok := result.Mutation(); ok {
		beforeValue, _ := mutation.BeforeValue().(map[string]any)
		afterValue, _ := mutation.AfterValue().(map[string]any)
		converted.Mutation = &timeline.RecordLinkMutation{
			RecordLinkID: result.RecordLinkID,
			Operation:    mutation.OperationKind(),
			BeforeValue:  cloneLinkMutationMap(beforeValue),
			AfterValue:   cloneLinkMutationMap(afterValue),
		}
	}
	return converted
}

func (a linkAdapter) ApplyTagCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command timeline.TagCollectionCommand) (timeline.CollectionMutationResult, error) {
	adds := make([]links.TagCollectionAdd, 0, len(command.AddTags))
	for _, add := range command.AddTags {
		adds = append(adds, links.TagCollectionAdd(add))
	}
	removes := make([]links.RecordTagRef, 0, len(command.RemoveTags))
	for _, remove := range command.RemoveTags {
		removes = append(removes, links.RecordTagRef(remove))
	}
	result, err := a.store.ApplyTagCollectionWithMutationValuesTx(ctx, tx, links.TagCollectionCommand{
		IncidentID:  command.IncidentID,
		RecordID:    command.RecordID,
		ActorUserID: command.ActorUserID,
		FieldKey:    command.FieldKey,
		AddTags:     adds,
		RemoveTags:  removes,
		Now:         command.Now,
	})
	if err != nil {
		return timeline.CollectionMutationResult{}, err
	}
	return collectionResult(result)
}

func collectionResult(result links.CollectionMutationResult) (timeline.CollectionMutationResult, error) {
	converted := timeline.CollectionMutationResult{
		RecordLinks: make([]timeline.RecordLinkMutation, 0),
		RecordTags:  make([]timeline.RecordTagMutation, 0),
	}
	for _, mutation := range result.Mutations() {
		beforeValue, _ := mutation.BeforeValue().(map[string]any)
		afterValue, _ := mutation.AfterValue().(map[string]any)
		switch mutation.TargetKind() {
		case "record_link":
			recordLinkID, err := uuid.Parse(mutation.TargetID())
			if err != nil {
				return timeline.CollectionMutationResult{}, err
			}
			converted.RecordLinks = append(converted.RecordLinks, timeline.RecordLinkMutation{
				RecordLinkID: recordLinkID,
				Operation:    mutation.OperationKind(),
				BeforeValue:  beforeValue,
				AfterValue:   afterValue,
			})
		case "record_tag":
			recordID, recordTagID, err := links.ParseRecordTagItemRef(mutation.TargetID())
			if err != nil {
				return timeline.CollectionMutationResult{}, err
			}
			converted.RecordTags = append(converted.RecordTags, timeline.RecordTagMutation{
				RecordTagID: recordTagID,
				RecordID:    recordID,
				Operation:   mutation.OperationKind(),
				BeforeValue: beforeValue,
				AfterValue:  afterValue,
			})
		default:
			return timeline.CollectionMutationResult{}, fmt.Errorf("timeline assembly: unsupported Links mutation target %q", mutation.TargetKind())
		}
	}
	return converted, nil
}
