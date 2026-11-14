package links

import (
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/revisionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/google/uuid"
)

type historyValidator struct{}

func historyRecordReferences(mutation revisions.StoredMutation, _ uuid.UUID) ([]revisions.HistoryRecordReference, error) {
	owned, err := valuecodec.HistoryRecordReferences(mutation.TargetKind, mutation.BeforeValue, mutation.AfterValue)
	if err != nil {
		return nil, err
	}
	references := make([]revisions.HistoryRecordReference, 0, len(owned))
	for _, reference := range owned {
		references = append(references, revisions.HistoryRecordReference{IncidentID: reference.IncidentID, RecordID: reference.RecordID})
	}
	return references, nil
}

func (historyValidator) ValidateHistoryMutation(mutation revisions.StoredMutation) error {
	return valuecodec.ValidateHistoryMutation(
		mutation.TargetKind,
		mutation.TargetID,
		mutation.OperationKind,
		mutation.BeforeValue,
		mutation.AfterValue,
	)
}

func RevisionProviderContribution() revisions.ProviderContribution {
	provider := revisionprovider.NewProvider()
	validator := historyValidator{}
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerLinks,
		NonRowTargets: []revisions.NonRowProviderContribution{
			{
				SourceOwnerModule: revisions.SourceOwnerLinks,
				TargetKind:        "record_link",
				HistoryProjector:  projectLinkHistory,
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"src_record_id", "dst_record_id"}, revisions.HistorySingleEntry),
				HistoryValidator:  validator,
				HistoryReferences: historyRecordReferences,
				RollbackProvider:  provider,
			},
			{
				SourceOwnerModule: revisions.SourceOwnerLinks,
				TargetKind:        "record_tag",
				HistoryProjector:  projectTagHistory,
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"record_id"}, revisions.HistorySingleEntry),
				HistoryValidator:  validator,
				HistoryReferences: historyRecordReferences,
				RollbackProvider:  provider,
			},
		},
	}
}
