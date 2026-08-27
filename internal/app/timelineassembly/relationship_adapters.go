package timelineassembly

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

type linkAdapter struct {
	store *links.Store
	facts links.FactReader
}

type mentionLinkAdapter struct {
	store *links.Store
}

func (a mentionLinkAdapter) UpsertMentionLinkTx(ctx context.Context, tx pgx.Tx, command mentions.LinkCommand) (mentions.LinkCommandResult, error) {
	result, err := a.store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    links.LinkType(command.LinkType),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: command.ActorUserID,
		Now:         command.Now,
	})
	if err != nil {
		return mentions.LinkCommandResult{}, err
	}
	return mentionLinkResult(result), nil
}

func (a mentionLinkAdapter) TombstoneActiveMentionLinkTx(ctx context.Context, tx pgx.Tx, command mentions.TombstoneLinkCommand) (mentions.LinkCommandResult, bool, error) {
	result, found, err := a.store.TombstoneActiveLinkCommandTx(ctx, tx, links.TombstoneActiveLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    links.LinkType(command.LinkType),
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
		LinkType:     mentions.LinkType(result.LinkType),
	}
	if result.Mutation != nil {
		converted.Mutation = &mentions.LinkMutation{
			RecordLinkID: result.Mutation.RecordLinkID,
			Operation:    result.Mutation.Operation,
			BeforeValue:  cloneLinkMutationMap(result.Mutation.BeforeValue),
			AfterValue:   cloneLinkMutationMap(result.Mutation.AfterValue),
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
	result, err := a.store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    links.LinkType(command.LinkType),
		Provenance:  links.LinkProvenance(command.Provenance),
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
	result, err := a.store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID:         command.IncidentID,
		SourceRecordID:     command.SourceRecordID,
		ActorUserID:        command.ActorUserID,
		FieldKey:           command.FieldKey,
		LinkType:           links.LinkType(command.LinkType),
		ExpectedTargetType: command.ExpectedTargetType,
		AddRecordIDs:       command.AddRecordIDs,
		RemoveRecordIDs:    command.RemoveRecordIDs,
		Now:                command.Now,
	})
	return collectionResult(result), err
}

func timelineLinkResult(result links.RecordLinkCommandResult) timeline.RecordLinkCommandResult {
	converted := timeline.RecordLinkCommandResult{
		RecordLinkID: result.RecordLinkID,
		SrcRecordID:  result.SrcRecordID,
		DstRecordID:  result.DstRecordID,
		LinkType:     result.LinkType.String(),
	}
	if result.Mutation != nil {
		converted.Mutation = &timeline.RecordLinkMutation{
			RecordLinkID: result.Mutation.RecordLinkID,
			Operation:    result.Mutation.Operation,
			BeforeValue:  cloneLinkMutationMap(result.Mutation.BeforeValue),
			AfterValue:   cloneLinkMutationMap(result.Mutation.AfterValue),
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
	return collectionResult(result), err
}

func (a linkAdapter) LoadCollectionFieldsChangedTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, changedAt time.Time) ([]string, error) {
	facts, err := a.facts.LoadCollectionChangesTx(ctx, tx, incidentID, recordID, changedAt)
	if err != nil {
		return nil, err
	}
	fields := append([]string(nil), facts.LinkFieldKeys...)
	if facts.TagsChanged {
		fields = append(fields, "timeline.tags")
	}
	return fields, nil
}

func collectionResult(result links.CollectionMutationResult) timeline.CollectionMutationResult {
	converted := timeline.CollectionMutationResult{
		RecordLinks: make([]timeline.RecordLinkMutation, 0, len(result.RecordLinks)),
		RecordTags:  make([]timeline.RecordTagMutation, 0, len(result.RecordTags)),
	}
	for _, mutation := range result.RecordLinks {
		converted.RecordLinks = append(converted.RecordLinks, timeline.RecordLinkMutation(mutation))
	}
	for _, mutation := range result.RecordTags {
		converted.RecordTags = append(converted.RecordTags, timeline.RecordTagMutation(mutation))
	}
	return converted
}
