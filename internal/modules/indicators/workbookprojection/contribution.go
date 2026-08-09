package workbookprojection

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

const indicatorViewSchemaID = "cartulary.view.indicators.v1"

type SourceReader interface {
	LoadProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionInput, bool, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

// ProjectionInput is the source-owned, typed materialization input for one
// Indicator workbook row. It intentionally contains no projection table or
// executable query-plan details.
type ProjectionInput struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	RowVersion          int64
	IndicatorType       string
	ValueKind           string
	DisplayValue        string
	NormalizedValue     *string
	DedupeKey           string
	DefangedValue       *string
	HashAlgorithm       *string
	HashValue           *string
	STIXPattern         *string
	FirstObservedAt     *time.Time
	LastObservedAt      *time.Time
	ObservationCount    int
	LifecycleSummary    *string
	SupportingLinkCount int
	EditedAt            time.Time
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

type Rows interface {
	RefreshIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadIndicatorTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	DeleteIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error
	RebuildIndicatorsTx(context.Context, pgx.Tx, uuid.UUID) error
}

type Rebuilder interface {
	RebuildIndicators(context.Context, uuid.UUID) error
}

type Ports struct {
	Rows      Rows
	Rebuilder Rebuilder
}

func (ports Ports) Ready() bool {
	return ports.Rows != nil && ports.Rebuilder != nil
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(
	descriptors []providercontract.ProviderDescriptor,
	intents []providercontract.SurfaceIntent,
	sources ...SourceReader,
) (Contribution, error) {
	if len(sources) > 1 {
		return Contribution{}, fmt.Errorf("indicators projection contribution accepts at most one source")
	}
	contract, err := providercontract.NewContribution("indicators", descriptors, intents)
	if err != nil {
		return Contribution{}, err
	}
	var source SourceReader
	if len(sources) == 1 {
		source = sources[0]
	}
	return Contribution{contract: contract, source: source}, nil
}

func NewRuntimeContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("indicators projection source is required")
	}
	return NewContribution(
		[]providercontract.ProviderDescriptor{Descriptor()},
		[]providercontract.SurfaceIntent{SurfaceIntent()},
		source,
	)
}

func (contribution Contribution) ProjectionContribution() providercontract.Contribution {
	return contribution.contract
}

func (contribution Contribution) Source() SourceReader {
	return contribution.source
}

func Descriptor() providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   "indicator",
		SourceOwnerModule:            "indicators",
		ViewSchemaIDs:                []string{indicatorViewSchemaID},
		SourceRecordTypes:            []string{"indicator"},
		SourceAuthorityModules:       []string{"indicators", "links", "records"},
		ProjectionTableIDs:           []string{"indicator_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild:       providercontract.RestoreRebuildRequired,
		FacadePackages:       []string{"internal/modules/indicators/workbookprojection"},
		RebuildAfter:         []string{"identity"},
		CharacterizationRefs: []string{"internal/modules/indicators/indicators_test.go"},
	}
}

func SurfaceIntent() providercontract.SurfaceIntent {
	return providercontract.SurfaceIntent{
		ViewSchemaID: indicatorViewSchemaID,
		FieldKeys: []string{
			"indicator.indicator_type",
			"indicator.value_kind",
			"indicator.display_value",
			"indicator.normalized_value",
			"indicator.defanged_value",
			"indicator.hash_algorithm",
			"indicator.hash_value",
			"indicator.stix_pattern",
			"indicator.first_observed_at",
			"indicator.last_observed_at",
			"indicator.observation_count",
			"indicator.lifecycle_summary",
			"indicator.supporting_link_count",
		},
	}
}
