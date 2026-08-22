package links

import (
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/revisionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type historyValidator struct{}

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
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"src_record_id", "dst_record_id"}, revisions.HistorySingleEntry),
				HistoryValidator:  validator,
				RollbackProvider:  provider,
			},
			{
				SourceOwnerModule: revisions.SourceOwnerLinks,
				TargetKind:        "record_tag",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"record_id"}, revisions.HistorySingleEntry),
				HistoryValidator:  validator,
				RollbackProvider:  provider,
			},
		},
	}
}
