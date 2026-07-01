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
	ProviderID           string                               `json:"provider_id"`
	SchemaVersion        string                               `json:"schema_version"`
	OwnerModule          string                               `json:"owner_module"`
	ViewSchemaIDs        []string                             `json:"view_schema_ids"`
	ProjectionTableIDs   []string                             `json:"projection_table_ids"`
	SourceAuthorities    []string                             `json:"source_authorities"`
	Capabilities         projectionProviderManifestCapability `json:"capabilities"`
	RestoreRebuild       string                               `json:"restore_rebuild"`
	Status               string                               `json:"status"`
	FacadePackages       []string                             `json:"facade_packages"`
	RebuildAfter         []string                             `json:"rebuild_after"`
	CharacterizationRefs []string                             `json:"characterization_refs"`
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
		restoreRebuild := "unsupported"
		if descriptor.RebuildIncidentSupported {
			restoreRebuild = "required"
		}
		entries = append(entries, projectionProviderManifestEntry{
			ProviderID:         descriptor.ProviderKey,
			SchemaVersion:      "projection_provider_descriptor.v1",
			OwnerModule:        descriptor.SourceOwnerKey,
			ViewSchemaIDs:      manifestStrings(descriptor.ViewSchemaIDs),
			ProjectionTableIDs: manifestStrings(descriptor.ProjectionTableFamilies),
			SourceAuthorities:  manifestStrings(descriptor.SourceRecordTypes),
			Capabilities: projectionProviderManifestCapability{
				Query:           providerSupportsQuery(descriptor.ViewSchemaIDs),
				RefreshRow:      descriptor.RefreshRowSupported,
				RestoreRebuild:  descriptor.RebuildIncidentSupported,
				IncidentRebuild: descriptor.RebuildIncidentSupported,
			},
			RestoreRebuild:       restoreRebuild,
			Status:               "active",
			FacadePackages:       facadePackagesForProvider(t, descriptor.ProviderKey),
			RebuildAfter:         manifestStrings(descriptor.RebuildAfter),
			CharacterizationRefs: manifestStrings(descriptor.CharacterizationRefs),
		})
	}

	return projectionProviderManifest{
		SchemaID:        "cartulary.projection_provider_manifest.v1",
		ManifestVersion: 1,
		Authority:       "validation_only_code_backed_registry_authoritative",
		SourceRegistry:  "internal/modules/projections/provider_registry.go",
		ApprovedProductionFacadeImports: []string{
			"internal/app/operator.go",
			"internal/modules/artifacts/import_projection.go",
			"internal/modules/artifacts/linkednotes/facade.go",
			"internal/modules/assessments/store.go",
			"internal/modules/entities/store.go",
			"internal/modules/evidence/import_projection.go",
			"internal/modules/evidence/store.go",
			"internal/modules/incidentbundles/source.go",
			"internal/modules/parties/store.go",
			"internal/modules/revisions/delete_restore_store.go",
			"internal/modules/revisions/rollback_store.go",
			"internal/modules/tasksdecisions/import_projection.go",
			"internal/modules/tasksdecisions/supersede_facade.go",
			"internal/modules/timeline/ports.go",
			"internal/modules/workbook/mutation_store.go",
			"internal/modules/workbook/store.go",
		},
		Providers: entries,
	}
}

func providerSupportsQuery(viewSchemaIDs []string) bool {
	for _, viewSchemaID := range viewSchemaIDs {
		if SupportsQuerySurface(viewSchemaID) {
			return true
		}
	}
	return false
}

func manifestStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return append([]string(nil), values...)
}

func facadePackagesForProvider(t *testing.T, providerKey string) []string {
	t.Helper()

	facadesByProvider := map[string][]string{
		"timeline":     {"internal/modules/timeline"},
		"host":         {"internal/modules/entities"},
		"identity":     {"internal/modules/entities"},
		"indicator":    {"internal/modules/indicators"},
		"assessment":   {"internal/modules/assessments"},
		"artifact":     {"internal/modules/artifacts", "internal/modules/artifacts/linkednotes", "internal/modules/workbook"},
		"evidence":     {"internal/modules/evidence"},
		"party":        {"internal/modules/parties"},
		"task_request": {"internal/modules/tasksdecisions"},
		"decision":     {"internal/modules/tasksdecisions"},
	}
	facades, ok := facadesByProvider[providerKey]
	if !ok {
		t.Fatalf("provider %q has no manifest facade package mapping", providerKey)
	}
	return append([]string(nil), facades...)
}

func prettyProjectionManifest(manifest projectionProviderManifest) string {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
