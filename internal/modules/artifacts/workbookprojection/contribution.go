package workbookprojection

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type Rows interface {
	RefreshArtifactTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadArtifactTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

type SourceReader interface {
	LoadProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionInput, bool, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

type DerivedFact struct {
	RecordID     uuid.UUID
	ArtifactType string
	FindingKind  *string
	Value        map[string]any
}

type Reader interface {
	CollectDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DerivedFact, error)
}

type Ports struct {
	Rows   Rows
	Reader Reader
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("artifacts projection source is required")
	}
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return Contribution{}, fmt.Errorf("compose Artifacts projection catalog: %w", err)
	}
	intents, err := surfaceIntents(catalog)
	if err != nil {
		return Contribution{}, err
	}
	contract, err := providercontract.NewContribution(
		"artifacts",
		[]providercontract.ProviderDescriptor{descriptor(catalog)},
		intents,
	)
	if err != nil {
		return Contribution{}, err
	}
	return Contribution{contract: contract, source: source}, nil
}

func (contribution Contribution) ProjectionContribution() providercontract.Contribution {
	return contribution.contract
}

func (contribution Contribution) Source() SourceReader {
	return contribution.source
}

func descriptor(catalog *sourcecatalog.Catalog) providercontract.ProviderDescriptor {
	surfaces := catalog.ProjectionSurfaces()
	viewSchemaIDs := make([]string, 0, len(surfaces))
	for _, surface := range surfaces {
		viewSchemaIDs = append(viewSchemaIDs, surface.ViewSchemaID)
	}
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   "artifact",
		SourceOwnerModule:            "artifacts",
		ViewSchemaIDs:                viewSchemaIDs,
		SourceRecordTypes:            []string{"artifact"},
		SourceAuthorityModules:       []string{"artifacts", "links", "parties", "records"},
		ProjectionTableIDs:           []string{"artifact_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/artifacts/workbookprojection"},
		RebuildAfter:   []string{"assessment"},
		CharacterizationRefs: []string{
			"internal/modules/workbook/coordination_surfaces_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		},
	}
}

func surfaceIntents(catalog *sourcecatalog.Catalog) ([]providercontract.SurfaceIntent, error) {
	surfaces := catalog.ProjectionSurfaces()
	intents := make([]providercontract.SurfaceIntent, 0, len(surfaces))
	for _, surface := range surfaces {
		viewSchemaID := surface.ViewSchemaID
		intent, err := surfaceIntent(viewSchemaID)
		if err != nil {
			return nil, err
		}
		intents = append(intents, intent)
	}
	return intents, nil
}

func surfaceIntent(viewSchemaID string) (providercontract.SurfaceIntent, error) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return providercontract.SurfaceIntent{}, fmt.Errorf("artifact projection surface %q has no view-schema contract", viewSchemaID)
	}
	filter, ok := schema.CanonicalSourceFilter()
	if !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value == "" {
		return providercontract.SurfaceIntent{}, fmt.Errorf("artifact projection surface %q has no canonical artifact_type filter", viewSchemaID)
	}
	fields := schema.Fields()
	fieldKeys := make([]string, 0, len(fields))
	for fieldKey := range fields {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	return providercontract.SurfaceIntent{
		ViewSchemaID: viewSchemaID,
		FieldKeys:    fieldKeys,
		CanonicalSourceFilter: &providercontract.SourceFilterIntent{
			Kind:  "artifact_type",
			Field: "artifact_type",
			Value: filter.Value,
		},
	}, nil
}
