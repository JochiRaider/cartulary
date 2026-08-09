package workbookprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

const assessmentsViewSchemaID = "cartulary.view.assessments.v1"

type SourceReader interface {
	BuildProjectionMutationTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionMutation, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

type Rows interface {
	ApplyAssessmentMutationTx(context.Context, pgx.Tx, ProjectionMutation) error
	RefreshAssessmentTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadAssessmentTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RebuildAssessmentsTx(context.Context, pgx.Tx, uuid.UUID) error
}

type Rebuilder interface {
	RebuildAssessments(context.Context, uuid.UUID) error
}

type Ports struct {
	Rows      Rows
	Rebuilder Rebuilder
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("assessments projection source is required")
	}
	contract, err := providercontract.NewContribution(
		"assessments",
		[]providercontract.ProviderDescriptor{Descriptor()},
		[]providercontract.SurfaceIntent{SurfaceIntent()},
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

func Descriptor() providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   "assessment",
		SourceOwnerModule:            "assessments",
		ViewSchemaIDs:                []string{assessmentsViewSchemaID},
		SourceRecordTypes:            []string{"assessment"},
		SourceAuthorityModules:       []string{"assessments", "links", "records"},
		ProjectionTableIDs:           []string{"assessment_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/assessments/workbookprojection"},
		RebuildAfter:   []string{"indicator"},
		CharacterizationRefs: []string{
			"internal/modules/assessments/assessment_contract_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		},
	}
}

func SurfaceIntent() providercontract.SurfaceIntent {
	return providercontract.SurfaceIntent{
		ViewSchemaID: assessmentsViewSchemaID,
		FieldKeys: []string{
			"assessment.subject_ref",
			"assessment.subject_type",
			"assessment.assessment_state",
			"assessment.confidence_band",
			"assessment.confidence_score",
			"assessment.rationale",
			"assessment.assessor",
			"assessment.assessed_at",
			"assessment.support_refs",
			"assessment.supporting_link_count",
		},
	}
}
