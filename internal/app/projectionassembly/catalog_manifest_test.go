package projectionassembly

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type projectionManifestDB struct {
	postgres.DB
}

type projectionProviderManifest struct {
	SchemaID        string                            `json:"schema_id"`
	ManifestVersion int                               `json:"manifest_version"`
	Authority       string                            `json:"authority"`
	SourceRegistry  string                            `json:"source_registry"`
	ImportPolicy    projectionProviderImportPolicy    `json:"import_policy"`
	Providers       []projectionProviderManifestEntry `json:"providers"`
}

type projectionProviderImportPolicy struct {
	ApprovedRootImporters    []string `json:"approved_root_importers"`
	ApprovedAdapterPackages  []string `json:"approved_adapter_packages"`
	ApprovedContractPackages []string `json:"approved_contract_packages"`
}

type projectionProviderManifestEntry struct {
	ProviderID                   string                               `json:"provider_id"`
	SchemaVersion                string                               `json:"schema_version"`
	SourceOwnerModule            string                               `json:"source_owner_module"`
	ProjectionStorageOwnerModule string                               `json:"projection_storage_owner_module"`
	ViewSchemaIDs                []string                             `json:"view_schema_ids"`
	ProjectionTableIDs           []string                             `json:"projection_table_ids"`
	SourceRecordTypes            []string                             `json:"source_record_types"`
	SourceAuthorityModules       []string                             `json:"source_authority_modules"`
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

	bundle, err := Build(&projectionManifestDB{})
	if err != nil {
		t.Fatalf("assemble projection adapter: %v", err)
	}
	descriptors := bundle.DescriptorSet().All()
	entries := make([]projectionProviderManifestEntry, 0, len(descriptors))
	for _, descriptor := range descriptors {
		entries = append(entries, projectionProviderManifestEntry{
			ProviderID:                   descriptor.ProviderID,
			SchemaVersion:                descriptor.SchemaVersion,
			SourceOwnerModule:            descriptor.SourceOwnerModule,
			ProjectionStorageOwnerModule: descriptor.ProjectionStorageOwnerModule,
			ViewSchemaIDs:                manifestStrings(descriptor.ViewSchemaIDs),
			ProjectionTableIDs:           manifestStrings(descriptor.ProjectionTableIDs),
			SourceRecordTypes:            manifestSortedStrings(descriptor.SourceRecordTypes),
			SourceAuthorityModules:       manifestSortedStrings(descriptor.SourceAuthorityModules),
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
		SchemaID:        "cartulary.projection_provider_manifest.v4",
		ManifestVersion: 4,
		Authority:       "validation_only_code_backed_registry_authoritative",
		SourceRegistry:  "internal/app/projectionassembly/catalog.go",
		ImportPolicy: projectionProviderImportPolicy{
			ApprovedRootImporters:    []string{},
			ApprovedAdapterPackages:  []string{"internal/modules/projections/adapters"},
			ApprovedContractPackages: []string{"internal/modules/projections/providercontract"},
		},
		Providers: entries,
	}
}

func TestProjectionAssemblyPortsAreCompleteAndDescriptorsImmutable(t *testing.T) {
	bundle, err := Build(&projectionManifestDB{})
	if err != nil {
		t.Fatalf("assemble projection adapter: %v", err)
	}
	if bundle.DescriptorSet().Len() != 10 {
		t.Fatalf("projection adapter descriptors are incomplete: %d", bundle.DescriptorSet().Len())
	}
	if !bundle.RecoveryPorts().Ready() ||
		bundle.MaintenanceRebuilder() == nil ||
		bundle.TimelinePorts().Writer == nil ||
		bundle.EntityMutationRows() == nil || bundle.EntityQueryReader() == nil ||
		bundle.EntityReportingReader() == nil ||
		bundle.IndicatorPorts().Rows == nil ||
		bundle.AssessmentPorts().Rows == nil ||
		bundle.ArtifactPorts().Rows == nil || bundle.ArtifactPorts().Reader == nil ||
		bundle.EvidenceMutationRows() == nil || bundle.EvidenceAssociationEffects() == nil ||
		bundle.PartyPorts().Rows == nil ||
		bundle.TaskDecisionMutationRows() == nil || bundle.TaskDecisionReportingReader() == nil ||
		bundle.RestoreProbeQuery() == nil || bundle.RevisionRebuilder() == nil || bundle.RevisionLiveRecords() == nil || bundle.SourceTextRows() == nil {
		t.Fatalf("projection assembly consumer ports are incomplete")
	}
	for _, descriptor := range bundle.DescriptorSet().All() {
		if descriptor.Capabilities.Query {
			for _, viewSchemaID := range descriptor.ViewSchemaIDs {
				provider, ok := bundle.WorkbookQueryProvider(viewSchemaID)
				if descriptor.ProviderID == "host" || descriptor.ProviderID == "identity" {
					if ok || provider != nil {
						t.Fatalf("typed entity query provider %q leaked through the generic adapter", viewSchemaID)
					}
					continue
				}
				if !ok || provider == nil {
					t.Fatalf("projection query provider %q is unavailable", viewSchemaID)
				}
			}
		}
	}

	descriptors := bundle.DescriptorSet().All()
	descriptors[0].ViewSchemaIDs[0] = "mutated.snapshot"
	again, _ := bundle.DescriptorSet().Lookup("timeline")
	if again.SchemaVersion != providercontract.DescriptorSchemaVersion || reflect.DeepEqual(again.ViewSchemaIDs, descriptors[0].ViewSchemaIDs) {
		t.Fatalf("projection assembly exposed mutable descriptors: %#v", again)
	}
}

func manifestSortedStrings(values []string) []string {
	result := manifestStrings(values)
	sort.Strings(result)
	return result
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
