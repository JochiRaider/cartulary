package entities

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := entitySourceDescriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export:     exportEntityIncidentBundleFiles,
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareEntityImport(bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedEntityImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return applyPreparedEntityImportTx(ctx, tx, prepared, importContext)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedEntityImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return validatePreparedEntityImportTx(ctx, tx, prepared, importContext)
		},
	})
}

func entitySourceDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "entities", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.entities", OwnerRelationIDs: []string{"entity-source"},
		Dependencies: []string{"parties"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/entity_mentions.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"entity_mention_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/hosts.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/identities.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/entity_preserved_identifiers.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"entity_preserved_identifier_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/entity_aliases.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"entity_alias_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"entities.mentions_observational", "entities.envelope_type_scope",
			"entities.resolution_merge_coherent", "entities.alias_identifier_normalized",
			"entities.alias_identifier_classified", "entities.alias_identifier_unique",
			"entities.alias_identifier_same_incident", "entities.source_identity_admitted",
		},
	}
}

func entitySourceFailure(invariantID string) error {
	return entitySourceDescriptor().DeclaredFailure(invariantID)
}
