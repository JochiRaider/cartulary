package incidentbundle

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func NewSourcePort() sourceport.Port {
	descriptor := linksIncidentBundleDescriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(ExportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			files, err := sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
			if err != nil {
				return nil, err
			}
			if err := validateTagCatalog(descriptor, files); err != nil {
				return nil, err
			}
			return files, nil
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			err := ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
			var unexpectedColumns *incidentportability.UnexpectedColumnsError
			if errors.As(err, &unexpectedColumns) {
				return descriptor.DeclaredFailure("links_tags.link_tuple_legal")
			}
			return err
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM record_links link
    LEFT JOIN records source ON source.record_id = link.src_record_id
    LEFT JOIN records target ON target.record_id = link.dst_record_id
    WHERE link.incident_id = $1
      AND (source.record_id IS NULL OR target.record_id IS NULL
           OR source.incident_id <> $1 OR target.incident_id <> $1)
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("links_tags.endpoints_same_incident")
			}
			return nil
		},
	})
}

func linksIncidentBundleDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "links_tags", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.links", OwnerRelationIDs: []string{"links-and-tags"},
		Dependencies: []string{"assessments"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/record_links.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"record_link_id"}, StableIdentityInvariantID: "links_tags.source_identity_admitted"},
			{LogicalPath: "data/tags.ndjson", ContentRole: "validation_rows", Versions: []int{3}, StableIdentity: []string{"normalized_tag_name", "tag_name"}, StableIdentityInvariantID: "links_tags.source_identity_admitted"},
			{LogicalPath: "data/record_tags.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"record_tag_id"}, StableIdentityInvariantID: "links_tags.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"links_tags.endpoints_same_incident", "links_tags.link_tuple_legal",
			"links_tags.link_unique", "links_tags.deletion_tuple_legal",
			"links_tags.tag_normalized", "links_tags.tag_catalog_exact",
			"links_tags.source_identity_admitted",
		},
	}
}

func validateTagCatalog(descriptor sourceport.Descriptor, files sourceport.PreparedFiles) error {
	catalogRows, err := incidentportability.DecodeNDJSON(files["data/tags.ndjson"])
	if err != nil {
		return err
	}
	tagRows, err := incidentportability.DecodeNDJSON(files["data/record_tags.ndjson"])
	if err != nil {
		return err
	}
	catalog := map[string]struct{}{}
	for _, row := range catalogRows {
		if len(row) != 2 {
			return descriptor.DeclaredFailure("links_tags.tag_catalog_exact")
		}
		key, err := tagCatalogKey(descriptor, row)
		if err != nil {
			return err
		}
		catalog[key] = struct{}{}
	}
	derived := map[string]struct{}{}
	for _, row := range tagRows {
		key, err := tagCatalogKey(descriptor, row)
		if err != nil {
			return err
		}
		derived[key] = struct{}{}
	}
	if len(catalog) != len(derived) {
		return descriptor.DeclaredFailure("links_tags.tag_catalog_exact")
	}
	for key := range catalog {
		if _, ok := derived[key]; !ok {
			return descriptor.DeclaredFailure("links_tags.tag_catalog_exact")
		}
	}
	return nil
}

func tagCatalogKey(descriptor sourceport.Descriptor, row map[string]any) (string, error) {
	tagName, tagOK := row["tag_name"].(string)
	normalized, normalizedOK := row["normalized_tag_name"].(string)
	if !tagOK || !normalizedOK || tagName == "" || normalized == "" {
		return "", descriptor.DeclaredFailure("links_tags.tag_normalized")
	}
	encoded, _ := json.Marshal([]string{tagName, normalized})
	return string(encoded), nil
}
