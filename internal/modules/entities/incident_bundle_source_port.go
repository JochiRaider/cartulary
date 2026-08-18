package entities

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "entities", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.entities", OwnerRelationIDs: []string{"entity-source"},
		Dependencies: []string{"parties"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/entity_mentions.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"entity_mention_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/hosts.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/identities.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/entity_preserved_identifiers.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"entity_preserved_identifier_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
			{LogicalPath: "data/entity_aliases.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"entity_alias_id"}, StableIdentityInvariantID: "entities.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"entities.mentions_observational", "entities.envelope_type_scope",
			"entities.resolution_merge_coherent", "entities.alias_identifier_normalized",
			"entities.alias_identifier_classified", "entities.alias_identifier_unique",
			"entities.alias_identifier_same_incident", "entities.source_identity_admitted",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(ExportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM (
        SELECT host.record_id, host.incident_id, 'host'::text AS required_type FROM hosts host
        UNION ALL
        SELECT identity.record_id, identity.incident_id, 'identity'::text FROM identities identity
    ) entity
    LEFT JOIN records record ON record.record_id = entity.record_id
    WHERE entity.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> entity.required_type)
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("entities.envelope_type_scope")
			}
			return nil
		},
	})
}
