package workbookprojection

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const partyViewSchemaID = "cartulary.view.parties.v1"

type Rows interface {
	RefreshPartyTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadPartyTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RebuildPartiesTx(context.Context, pgx.Tx, uuid.UUID) error
}

type ProjectionInput struct {
	RecordID         uuid.UUID
	IncidentID       uuid.UUID
	RowVersion       int64
	DisplayName      *string
	PartyKind        *string
	OrganizationName *string
	RoleTitle        *string
	PrimaryEmail     *string
	TimezoneName     *string
	ExternalRef      *string
	Notes            *string
	UpdatedAt        time.Time
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

type SourceReader interface {
	LoadProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionInput, bool, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

type Rebuilder interface {
	RebuildParties(context.Context, uuid.UUID) error
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
		return Contribution{}, fmt.Errorf("parties projection contribution accepts at most one source")
	}
	contract, err := providercontract.NewContribution("parties", descriptors, intents)
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
		return Contribution{}, fmt.Errorf("parties projection source is required")
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
		ProviderID:                   "party",
		SourceOwnerModule:            "parties",
		ViewSchemaIDs:                []string{partyViewSchemaID},
		SourceRecordTypes:            []string{"party"},
		SourceAuthorityModules:       []string{"parties", "records"},
		ProjectionTableIDs:           []string{"party_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/parties/workbookprojection"},
		RebuildAfter:   []string{"evidence"},
		CharacterizationRefs: []string{
			"internal/modules/workbook/parties_integration_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		},
	}
}

func SurfaceIntent() providercontract.SurfaceIntent {
	schema, ok := viewschema.Lookup(partyViewSchemaID)
	if !ok {
		panic("Parties projection surface has no view-schema contract")
	}
	fields := schema.Fields()
	fieldKeys := make([]string, 0, len(fields))
	for fieldKey := range fields {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	return providercontract.SurfaceIntent{
		ViewSchemaID: partyViewSchemaID,
		FieldKeys:    fieldKeys,
	}
}
