package projectionassembly

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactcontract "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentcontract "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	evidencecontract "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorcontract "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partycontract "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	taskdecisioncontract "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookrestoreprobe "github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Bundle is the application-composition view of the sole Projections adapter.
// Concrete runtime, catalog, storage, and query-engine values never cross this
// boundary.
type Bundle struct {
	ports projectionadapters.Ports
}

type ImportRebuilder interface {
	RebuildImportedIncidentTx(context.Context, pgx.Tx, uuid.UUID) error
}

func NewBundle(
	pool postgres.DB,
	timelineContribution timelineprojection.Contribution,
	entitiesContribution entitycontract.Contribution,
	indicatorsContribution indicatorcontract.Contribution,
	assessmentsContribution assessmentcontract.Contribution,
	artifactsContribution artifactcontract.Contribution,
	evidenceContribution evidencecontract.Contribution,
	partiesContribution partycontract.Contribution,
	taskDecisionContribution taskdecisioncontract.Contribution,
) (*Bundle, error) {
	ports, err := projectionadapters.New(projectionadapters.Dependencies{
		Postgres:       pool,
		Timeline:       timelineContribution,
		Entities:       entitiesContribution,
		Indicators:     indicatorsContribution,
		Assessments:    assessmentsContribution,
		Artifacts:      artifactsContribution,
		Evidence:       evidenceContribution,
		Parties:        partiesContribution,
		TasksDecisions: taskDecisionContribution,
	})
	if err != nil {
		return nil, fmt.Errorf("assemble projection adapter: %w", err)
	}
	return &Bundle{ports: ports}, nil
}

func (bundle *Bundle) DescriptorSet() providercontract.DescriptorSet {
	if bundle == nil {
		return providercontract.DescriptorSet{}
	}
	return bundle.ports.DescriptorSet()
}

func (bundle *Bundle) WorkbookQueryProvider(viewSchemaID string) (workbook.QueryProvider, bool) {
	if bundle == nil {
		return nil, false
	}
	return bundle.ports.WorkbookQueryProvider(viewSchemaID)
}

func (bundle *Bundle) RecoveryPorts() restorecontract.ProjectionPorts {
	if bundle == nil {
		return restorecontract.ProjectionPorts{}
	}
	return bundle.ports.RecoveryPorts()
}

func (bundle *Bundle) RestoreProbeQuery() workbookrestoreprobe.ProjectionQuery {
	if bundle == nil {
		return nil
	}
	return bundle.ports.RestoreProbeQuery()
}

func (bundle *Bundle) RevisionServices() revisions.ProjectionServices {
	if bundle == nil {
		return nil
	}
	return bundle.ports.RevisionServices()
}

func (bundle *Bundle) SourceTextRows() projectionadapters.SourceTextRows {
	if bundle == nil {
		return nil
	}
	return bundle.ports.SourceTextRows()
}

func (bundle *Bundle) ImportRebuilder() ImportRebuilder {
	if bundle == nil {
		return nil
	}
	return bundle.ports.ImportRebuilder()
}

func (bundle *Bundle) TimelinePorts() timelineprojection.Ports {
	if bundle == nil {
		return timelineprojection.Ports{}
	}
	return bundle.ports.Timeline()
}

func (bundle *Bundle) EntityPorts() entitycontract.Ports {
	if bundle == nil {
		return entitycontract.Ports{}
	}
	return bundle.ports.Entities()
}

func (bundle *Bundle) IndicatorPorts() indicatorcontract.Ports {
	if bundle == nil {
		return indicatorcontract.Ports{}
	}
	return bundle.ports.Indicators()
}

func (bundle *Bundle) AssessmentPorts() assessmentcontract.Ports {
	if bundle == nil {
		return assessmentcontract.Ports{}
	}
	return bundle.ports.Assessments()
}

func (bundle *Bundle) ArtifactPorts() artifactcontract.Ports {
	if bundle == nil {
		return artifactcontract.Ports{}
	}
	return bundle.ports.Artifacts()
}

func (bundle *Bundle) EvidencePorts() evidencecontract.Ports {
	if bundle == nil {
		return evidencecontract.Ports{}
	}
	return bundle.ports.Evidence()
}

func (bundle *Bundle) PartyPorts() partycontract.Ports {
	if bundle == nil {
		return partycontract.Ports{}
	}
	return bundle.ports.Parties()
}

func (bundle *Bundle) TaskDecisionPorts() taskdecisioncontract.Ports {
	if bundle == nil {
		return taskdecisioncontract.Ports{}
	}
	return bundle.ports.TasksDecisions()
}

// NewEvidenceContribution keeps executable Evidence source construction at
// application composition while returning only the owner facade contract.
func NewEvidenceContribution() (evidencecontract.Contribution, error) {
	return evidenceprojection.NewContribution()
}
