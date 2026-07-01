package projections

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type projectionProviderManifest struct {
	SchemaID                        string                            `json:"schema_id"`
	ManifestVersion                 int                               `json:"manifest_version"`
	Authority                       string                            `json:"authority"`
	SourceRegistry                  string                            `json:"source_registry"`
	ApprovedProductionFacadeImports []string                          `json:"approved_production_facade_imports"`
	Providers                       []projectionProviderManifestEntry `json:"providers"`
}

type projectionProviderManifestEntry struct {
	ProviderID                   string                               `json:"provider_id"`
	SchemaVersion                string                               `json:"schema_version"`
	SourceOwnerModule            string                               `json:"source_owner_module"`
	ProjectionStorageOwnerModule string                               `json:"projection_storage_owner_module"`
	ViewSchemaIDs                []string                             `json:"view_schema_ids"`
	ProjectionTableIDs           []string                             `json:"projection_table_ids"`
	SourceAuthorities            []string                             `json:"source_authorities"`
	Capabilities                 projectionProviderManifestCapability `json:"capabilities"`
	RestoreRebuild               string                               `json:"restore_rebuild"`
	Status                       string                               `json:"status"`
	FacadePackages               []string                             `json:"facade_packages"`
	RebuildAfter                 []string                             `json:"rebuild_after"`
	CharacterizationRefs         []string                             `json:"characterization_refs"`
}

type projectionProviderManifestCapability struct {
	Query           bool `json:"query"`
	RefreshRow      bool `json:"refresh_row"`
	RestoreRebuild  bool `json:"restore_rebuild"`
	IncidentRebuild bool `json:"incident_rebuild"`
}

func TestProjectionProviderManifestMirrorsCodeBackedRegistry(t *testing.T) {
	manifestFile := filepath.Join("..", "..", "..", "contracts", "projection-providers", "index.json")
	data, err := os.ReadFile(manifestFile)
	if err != nil {
		t.Fatalf("read projection provider manifest: %v", err)
	}

	var manifest projectionProviderManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode projection provider manifest: %v", err)
	}

	expected := expectedProjectionProviderManifest(t)
	if !reflect.DeepEqual(manifest, expected) {
		t.Fatalf("projection provider manifest drift\nexpected:\n%s\nactual:\n%s", prettyProjectionManifest(expected), prettyProjectionManifest(manifest))
	}
}

func expectedProjectionProviderManifest(t *testing.T) projectionProviderManifest {
	t.Helper()

	providers := builtInProjectionProviders()
	entries := make([]projectionProviderManifestEntry, 0, len(providers))
	for _, provider := range providers {
		descriptor := provider.descriptor
		entries = append(entries, projectionProviderManifestEntry{
			ProviderID:                   descriptor.ProviderKey,
			SchemaVersion:                descriptor.SchemaVersion,
			SourceOwnerModule:            descriptor.SourceOwnerKey,
			ProjectionStorageOwnerModule: descriptor.ProjectionStorageOwnerKey,
			ViewSchemaIDs:                manifestStrings(descriptor.ViewSchemaIDs),
			ProjectionTableIDs:           manifestStrings(descriptor.ProjectionTableFamilies),
			SourceAuthorities:            manifestStrings(descriptor.SourceRecordTypes),
			Capabilities: projectionProviderManifestCapability{
				Query:           descriptor.Capabilities.Query,
				RefreshRow:      descriptor.Capabilities.RefreshRow,
				RestoreRebuild:  descriptor.Capabilities.RestoreRebuild,
				IncidentRebuild: descriptor.Capabilities.IncidentRebuild,
			},
			RestoreRebuild:       string(descriptor.RestoreRebuild),
			Status:               string(descriptor.Status),
			FacadePackages:       manifestStrings(descriptor.FacadePackages),
			RebuildAfter:         manifestStrings(descriptor.RebuildAfter),
			CharacterizationRefs: manifestStrings(descriptor.CharacterizationRefs),
		})
	}

	return projectionProviderManifest{
		SchemaID:                        "cartulary.projection_provider_manifest.v2",
		ManifestVersion:                 2,
		Authority:                       "validation_only_code_backed_registry_authoritative",
		SourceRegistry:                  "internal/modules/projections/provider_registry.go",
		ApprovedProductionFacadeImports: approvedProductionProjectionImporterPaths(),
		Providers:                       entries,
	}
}

func manifestStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return append([]string(nil), values...)
}

func prettyProjectionManifest(manifest projectionProviderManifest) string {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
