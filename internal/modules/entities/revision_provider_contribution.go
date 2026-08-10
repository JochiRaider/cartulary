package entities

import (
	entitiesdeleterestore "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/deleterestore"
	entityrollback "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/rollbackprovider"
	mentionrollback "github.com/JochiRaider/cartulary/internal/modules/entities/mentions/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	collectionProvider := entityrollback.NewCollectionProvider()
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerEntities,
		Records: []revisions.RecordProviderContribution{
			{
				SourceOwnerModule:   revisions.SourceOwnerEntities,
				RecordType:          "host",
				SnapshotSchemaID:    "cartulary.revisions.snapshot.host.v1",
				HistoryTargetKinds:  []string{"host"},
				DeleteRestoreSource: entitiesdeleterestore.NewHostSource(),
				RowRollbackProvider: entityrollback.NewHostProvider(),
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "entities.hosts",
					ViewSchemaIDs:  []string{"cartulary.view.hosts.v1"},
				}},
			},
			{
				SourceOwnerModule:   revisions.SourceOwnerEntities,
				RecordType:          "identity",
				SnapshotSchemaID:    "cartulary.revisions.snapshot.identity.v1",
				HistoryTargetKinds:  []string{"identity"},
				DeleteRestoreSource: entitiesdeleterestore.NewIdentitySource(),
				RowRollbackProvider: entityrollback.NewIdentityProvider(),
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "entities.identities",
					ViewSchemaIDs:  []string{"cartulary.view.identities.v1"},
				}},
			},
		},
		NonRowTargets: []revisions.NonRowProviderContribution{
			{
				SourceOwnerModule: revisions.SourceOwnerEntities,
				TargetKind:        "entity_mention",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"source_record_id"}, revisions.HistorySingleEntry),
				RollbackProvider:  mentionrollback.NewMentionProvider(),
			},
			{
				SourceOwnerModule: revisions.SourceOwnerEntities,
				TargetKind:        "entity_alias",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"record_id"}, revisions.HistoryNotIndividuallyAddressable),
				RollbackProvider:  collectionProvider,
			},
			{
				SourceOwnerModule: revisions.SourceOwnerEntities,
				TargetKind:        "entity_preserved_identifier",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"record_id"}, revisions.HistoryNotIndividuallyAddressable),
				RollbackProvider:  collectionProvider,
			},
		},
	}
}
