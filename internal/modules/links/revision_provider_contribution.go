package links

import (
	"github.com/JochiRaider/cartulary/internal/modules/links/revisionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	provider := revisionprovider.NewProvider()
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerLinks,
		NonRowTargets: []revisions.NonRowProviderContribution{
			{
				SourceOwnerModule: revisions.SourceOwnerLinks,
				TargetKind:        "record_link",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"src_record_id", "dst_record_id"}, revisions.HistorySingleEntry),
				RollbackProvider:  provider,
			},
			{
				SourceOwnerModule: revisions.SourceOwnerLinks,
				TargetKind:        "record_tag",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"record_id"}, revisions.HistorySingleEntry),
				RollbackProvider:  provider,
			},
		},
	}
}
