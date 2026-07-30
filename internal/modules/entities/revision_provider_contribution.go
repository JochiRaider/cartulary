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
				SourceOwnerModule:      revisions.SourceOwnerEntities,
				RecordType:             "host",
				DeleteRestoreSource:    entitiesdeleterestore.NewHostSource(),
				RowRollbackProvider:    entityrollback.NewHostProvider(),
				LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "entities.hosts",
					ViewSchemaIDs:  []string{"cartulary.view.hosts.v1"},
				}},
			},
			{
				SourceOwnerModule:      revisions.SourceOwnerEntities,
				RecordType:             "identity",
				DeleteRestoreSource:    entitiesdeleterestore.NewIdentitySource(),
				RowRollbackProvider:    entityrollback.NewIdentityProvider(),
				LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "entities.identities",
					ViewSchemaIDs:  []string{"cartulary.view.identities.v1"},
				}},
			},
		},
		NonRowTargets: []revisions.NonRowProviderContribution{
			{SourceOwnerModule: revisions.SourceOwnerEntities, TargetKind: "entity_mention", RollbackProvider: mentionrollback.NewMentionProvider()},
			{SourceOwnerModule: revisions.SourceOwnerEntities, TargetKind: "entity_alias", RollbackProvider: collectionProvider},
			{SourceOwnerModule: revisions.SourceOwnerEntities, TargetKind: "entity_preserved_identifier", RollbackProvider: collectionProvider},
		},
	}
}
