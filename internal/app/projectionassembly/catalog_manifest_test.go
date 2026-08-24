package projectionassembly

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	artifactcontract "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentcontract "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	evidencecontract "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorowner "github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprovider "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type projectionManifestDB struct {
	postgres.DB
}

type projectionManifestTimelineSource struct {
	timelineprojection.SourceReader
}

type projectionManifestEntitySource struct {
	entityprojection.SourceReader
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

	bundle, err := buildRuntime(
		&projectionManifestDB{},
		projectionManifestTimelineContribution(t),
		projectionManifestEntitiesContribution(t),
		projectionManifestIndicatorsContribution(t),
		projectionManifestAssessmentsContribution(t),
		projectionManifestArtifactsContribution(t),
		projectionManifestEvidenceContribution(t),
		projectionManifestPartiesContribution(t),
		projectionManifestTaskDecisionContribution(t),
	)
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
	bundle, err := buildRuntime(
		&projectionManifestDB{},
		projectionManifestTimelineContribution(t),
		projectionManifestEntitiesContribution(t),
		projectionManifestIndicatorsContribution(t),
		projectionManifestAssessmentsContribution(t),
		projectionManifestArtifactsContribution(t),
		projectionManifestEvidenceContribution(t),
		projectionManifestPartiesContribution(t),
		projectionManifestTaskDecisionContribution(t),
	)
	if err != nil {
		t.Fatalf("assemble projection adapter: %v", err)
	}
	if bundle.DescriptorSet().Len() != 10 {
		t.Fatalf("projection adapter descriptors are incomplete: %d", bundle.DescriptorSet().Len())
	}
	if !bundle.RecoveryPorts().Ready() ||
		bundle.TimelinePorts().Writer == nil || bundle.TimelinePorts().Rebuilder == nil ||
		bundle.EntityPorts().Writer == nil || bundle.EntityPorts().Rebuilder == nil || bundle.EntityPorts().Reader == nil ||
		bundle.IndicatorPorts().Rows == nil || bundle.IndicatorPorts().Rebuilder == nil ||
		bundle.AssessmentPorts().Rows == nil || bundle.AssessmentPorts().Rebuilder == nil ||
		bundle.ArtifactPorts().Rows == nil || bundle.ArtifactPorts().Rebuilder == nil || bundle.ArtifactPorts().Reader == nil ||
		bundle.EvidencePorts().Rows == nil || bundle.EvidencePorts().Rebuilder == nil ||
		bundle.PartyPorts().Rows == nil ||
		bundle.TaskDecisionPorts().Rows == nil || bundle.TaskDecisionPorts().Rebuilder == nil || bundle.TaskDecisionPorts().Reader == nil ||
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

func projectionManifestTimelineContribution(t testing.TB) timelineprojection.Contribution {
	t.Helper()
	contribution, err := timelineprojection.NewContribution(&projectionManifestTimelineSource{})
	if err != nil {
		t.Fatalf("construct Timeline projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestEntitiesContribution(t testing.TB) entityprojection.Contribution {
	t.Helper()
	contribution, err := entityprojection.NewContribution(&projectionManifestEntitySource{})
	if err != nil {
		t.Fatalf("construct Entities projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestIndicatorsContribution(t testing.TB) indicatorprojection.Contribution {
	t.Helper()
	contribution, err := indicatorowner.NewProjectionContribution()
	if err != nil {
		t.Fatalf("construct Indicators projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestAssessmentsContribution(t testing.TB) assessmentcontract.Contribution {
	t.Helper()
	contribution, err := assessmentassembly.NewProjectionContribution()
	if err != nil {
		t.Fatalf("construct Assessments projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestArtifactsContribution(t testing.TB) artifactcontract.Contribution {
	t.Helper()
	contribution, err := artifacts.NewProjectionContribution()
	if err != nil {
		t.Fatalf("construct Artifacts projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestEvidenceContribution(t testing.TB) evidencecontract.Contribution {
	t.Helper()
	contribution, err := evidenceprojection.NewContribution()
	if err != nil {
		t.Fatalf("construct Evidence projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestPartiesContribution(t testing.TB) partyprojection.Contribution {
	t.Helper()
	contribution, err := parties.NewProjectionContribution()
	if err != nil {
		t.Fatalf("construct Parties projection contribution: %v", err)
	}
	return contribution
}

func projectionManifestTaskDecisionContribution(t testing.TB) taskdecisionprojection.Contribution {
	t.Helper()
	contribution, err := taskdecisionprovider.NewContribution()
	if err != nil {
		t.Fatalf("construct Tasks/Decisions projection contribution: %v", err)
	}
	return contribution
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
