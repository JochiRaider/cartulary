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
			{SourceOwnerModule: revisions.SourceOwnerEntities, RecordType: "host", DeleteRestoreProvider: entitiesdeleterestore.HostProvider(), RowRollbackProvider: entityrollback.NewHostProvider()},
			{SourceOwnerModule: revisions.SourceOwnerEntities, RecordType: "identity", DeleteRestoreProvider: entitiesdeleterestore.IdentityProvider(), RowRollbackProvider: entityrollback.NewIdentityProvider()},
		},
		NonRowTargets: []revisions.NonRowProviderContribution{
			{SourceOwnerModule: revisions.SourceOwnerEntities, TargetKind: "entity_mention", RollbackProvider: mentionrollback.NewMentionProvider()},
			{SourceOwnerModule: revisions.SourceOwnerEntities, TargetKind: "entity_alias", RollbackProvider: collectionProvider},
			{SourceOwnerModule: revisions.SourceOwnerEntities, TargetKind: "entity_preserved_identifier", RollbackProvider: collectionProvider},
		},
	}
}
