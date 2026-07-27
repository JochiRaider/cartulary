package links

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "links_tags", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.links", OwnerRelationIDs: []string{"links-and-tags"},
		Dependencies: []string{"assessments"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/record_links.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_link_id"}},
			{LogicalPath: "data/tags.ndjson", ContentRole: "validation_rows", Versions: []int{1, 2}, StableIdentity: []string{"normalized_tag_name", "tag_name"}},
			{LogicalPath: "data/record_tags.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_tag_id"}},
		},
		InvariantIDs: []string{
			"links_tags.endpoints_same_incident", "links_tags.link_tuple_legal",
			"links_tags.link_unique", "links_tags.deletion_tuple_legal",
			"links_tags.tag_normalized", "links_tags.tag_catalog_exact",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: ExportIncidentBundleFiles,
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			files, err := sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
			if err != nil {
				return nil, err
			}
			if err := validateTagCatalog(files); err != nil {
				return nil, err
			}
			return files, nil
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, importContext sourceport.ImportContext) error {
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
				return &sourceport.Failure{FamilyID: "links_tags", InvariantID: "links_tags.endpoints_same_incident"}
			}
			return nil
		},
	})
}

func validateTagCatalog(files sourceport.PreparedFiles) error {
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
			return &sourceport.Failure{FamilyID: "links_tags", InvariantID: "links_tags.tag_catalog_exact"}
		}
		key, err := tagCatalogKey(row)
		if err != nil {
			return err
		}
		catalog[key] = struct{}{}
	}
	derived := map[string]struct{}{}
	for _, row := range tagRows {
		key, err := tagCatalogKey(row)
		if err != nil {
			return err
		}
		derived[key] = struct{}{}
	}
	if len(catalog) != len(derived) {
		return &sourceport.Failure{FamilyID: "links_tags", InvariantID: "links_tags.tag_catalog_exact"}
	}
	for key := range catalog {
		if _, ok := derived[key]; !ok {
			return &sourceport.Failure{FamilyID: "links_tags", InvariantID: "links_tags.tag_catalog_exact"}
		}
	}
	return nil
}

func tagCatalogKey(row map[string]any) (string, error) {
	tagName, tagOK := row["tag_name"].(string)
	normalized, normalizedOK := row["normalized_tag_name"].(string)
	if !tagOK || !normalizedOK || tagName == "" || normalized == "" {
		return "", &sourceport.Failure{FamilyID: "links_tags", InvariantID: "links_tags.tag_normalized"}
	}
	encoded, _ := json.Marshal([]string{tagName, normalized})
	return string(encoded), nil
}
