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
	taskdecisioncontract "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
	taskdecisionports "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookrestoreprobe "github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Runtime is the application-composition view of the sole Projections adapter.
// Concrete runtime, catalog, storage, and query-engine values never cross this
// boundary.
type Runtime struct {
	ports projectionadapters.Ports
}

type ImportRebuilder interface {
	RebuildImportedIncidentTx(context.Context, pgx.Tx, uuid.UUID) error
}

func buildRuntime(
	pool postgres.DB,
	timelineContribution timelineprojection.Contribution,
	entitiesContribution entitycontract.Contribution,
	indicatorsContribution indicatorcontract.Contribution,
	assessmentsContribution assessmentcontract.Contribution,
	artifactsContribution artifactcontract.Contribution,
	evidenceContribution evidencecontract.Contribution,
	partiesContribution partycontract.Contribution,
	taskDecisionContribution taskdecisioncontract.Contribution,
) (*Runtime, error) {
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
	return &Runtime{ports: ports}, nil
}

func (runtime *Runtime) DescriptorSet() providercontract.DescriptorSet {
	if runtime == nil {
		return providercontract.DescriptorSet{}
	}
	return runtime.ports.DescriptorSet()
}

func (runtime *Runtime) WorkbookQueryProvider(viewSchemaID string) (workbook.QueryProvider, bool) {
	if runtime == nil {
		return nil, false
	}
	return runtime.ports.WorkbookQueryProvider(viewSchemaID)
}

func (runtime *Runtime) RecoveryPorts() restorecontract.ProjectionPorts {
	if runtime == nil {
		return restorecontract.ProjectionPorts{}
	}
	return runtime.ports.RecoveryPorts()
}

func (runtime *Runtime) RestoreProbeQuery() workbookrestoreprobe.ProjectionQuery {
	if runtime == nil {
		return nil
	}
	return runtime.ports.RestoreProbeQuery()
}

func (runtime *Runtime) RevisionRebuilder() revisions.ProjectionRebuilder {
	if runtime == nil {
		return nil
	}
	return runtime.ports.RevisionRebuilder()
}

func (runtime *Runtime) RevisionLiveRecords() revisions.LiveRecordReader {
	if runtime == nil {
		return nil
	}
	return runtime.ports.RevisionLiveRecords()
}

func (runtime *Runtime) SourceTextRows() projectionadapters.SourceTextRows {
	if runtime == nil {
		return nil
	}
	return runtime.ports.SourceTextRows()
}

func (runtime *Runtime) ImportRebuilder() ImportRebuilder {
	if runtime == nil {
		return nil
	}
	return runtime.ports.ImportRebuilder()
}

func (runtime *Runtime) TimelinePorts() timelineprojection.Ports {
	if runtime == nil {
		return timelineprojection.Ports{}
	}
	return runtime.ports.Timeline()
}

func (runtime *Runtime) EntityPorts() entitycontract.Ports {
	if runtime == nil {
		return entitycontract.Ports{}
	}
	return runtime.ports.Entities()
}

func (runtime *Runtime) IndicatorPorts() indicatorcontract.Ports {
	if runtime == nil {
		return indicatorcontract.Ports{}
	}
	return runtime.ports.Indicators()
}

func (runtime *Runtime) AssessmentPorts() assessmentcontract.Ports {
	if runtime == nil {
		return assessmentcontract.Ports{}
	}
	return runtime.ports.Assessments()
}

func (runtime *Runtime) ArtifactPorts() artifactcontract.Ports {
	if runtime == nil {
		return artifactcontract.Ports{}
	}
	return runtime.ports.Artifacts()
}

func (runtime *Runtime) EvidencePorts() evidencecontract.Ports {
	if runtime == nil {
		return evidencecontract.Ports{}
	}
	return runtime.ports.Evidence()
}

func (runtime *Runtime) PartyPorts() partycontract.Ports {
	if runtime == nil {
		return partycontract.Ports{}
	}
	return runtime.ports.Parties()
}

func (runtime *Runtime) TaskDecisionMutationRows() taskdecisionports.MutationRows {
	if runtime == nil {
		return nil
	}
	return runtime.ports.TaskDecisionMutationRows()
}

func (runtime *Runtime) TaskDecisionReportingReader() taskdecisionports.ReportingReader {
	if runtime == nil {
		return nil
	}
	return runtime.ports.TaskDecisionReportingReader()
}

// NewEvidenceContribution keeps executable Evidence source construction at
// application composition while returning only the owner facade contract.
func newEvidenceContribution() (evidencecontract.Contribution, error) {
	return evidenceprojection.NewContribution()
}
