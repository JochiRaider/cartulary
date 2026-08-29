package projectionassembly

import (
	artifactcontract "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentcontract "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	evidenceports "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	indicatorcontract "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partycontract "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	taskdecisionports "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookrestoreprobe "github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
)

// Runtime is the application-composition view of the sole Projections adapter.
// Concrete runtime, catalog, storage, and query-engine values never cross this
// boundary.
type Runtime struct {
	ports projectionadapters.Ports
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

func (runtime *Runtime) MaintenanceRebuilder() projectionadapters.MaintenanceRebuilder {
	if runtime == nil {
		return nil
	}
	return runtime.ports.MaintenanceRebuilder()
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

func (runtime *Runtime) ImportRebuilder() incidentbundles.ImportProjectionRebuilder {
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

func (runtime *Runtime) EntityMutationRows() entityports.MutationRows {
	if runtime == nil {
		return nil
	}
	return runtime.ports.EntityMutationRows()
}

func (runtime *Runtime) EntityQueryReader() entityports.QueryReader {
	if runtime == nil {
		return nil
	}
	return runtime.ports.EntityQueryReader()
}

func (runtime *Runtime) EntityReportingReader() entityports.ReportingReader {
	if runtime == nil {
		return nil
	}
	return runtime.ports.EntityReportingReader()
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

func (runtime *Runtime) EvidenceMutationRows() evidenceports.MutationRows {
	if runtime == nil {
		return nil
	}
	return runtime.ports.EvidenceMutationRows()
}

func (runtime *Runtime) EvidenceAssociationEffects() evidenceports.AssociationEffects {
	if runtime == nil {
		return nil
	}
	return runtime.ports.EvidenceAssociationEffects()
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
