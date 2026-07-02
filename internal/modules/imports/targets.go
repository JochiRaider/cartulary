package imports

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

const (
	applyStatusSupported              = "supported"
	applyStatusSupportedWhenAvailable = "supported_when_implemented"

	createFacadeTimeline     = "timeline.import_create"
	createFacadeHost         = "entities.host.import_create"
	createFacadeIdentity     = "entities.identity.import_create"
	createFacadeIndicator    = "indicators.import_create"
	createFacadeEvidence     = "evidence.import_create"
	createFacadeNoteArtifact = "artifacts.note.import_create"
	createFacadeArtifact     = "artifacts.import_create"
	createFacadeAssessment   = "assessments.import_create"
	createFacadeTask         = "tasksdecisions.task_request.import_create"
	createFacadeDecision     = "tasksdecisions.decision.import_create"
	createFacadeParty        = "parties.import_create"
)

type importTarget struct {
	ViewSchemaID     string
	Owner            string
	RecordFamily     string
	ApplyStatus      string
	CreateFacade     string
	AllowRawCapture  bool
	AllowCustomAttrs bool
}

func (target importTarget) importable() bool {
	return target.ApplyStatus == applyStatusSupported
}

func (target importTarget) ownerCreateFacadeAvailable() bool {
	return target.CreateFacade != ""
}

func lookupImportTarget(viewSchemaID string) (importTarget, bool) {
	target, ok := importTargets[viewSchemaID]
	return target, ok
}

var importTargets = map[string]importTarget{
	timeline.TimelineViewSchemaID: {
		ViewSchemaID:     timeline.TimelineViewSchemaID,
		Owner:            "timeline",
		RecordFamily:     "timeline_event",
		ApplyStatus:      applyStatusSupported,
		CreateFacade:     createFacadeTimeline,
		AllowRawCapture:  true,
		AllowCustomAttrs: false,
	},
	hostidentity.HostsViewSchemaID: {
		ViewSchemaID:    hostidentity.HostsViewSchemaID,
		Owner:           "entities",
		RecordFamily:    "host",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeHost,
		AllowRawCapture: false,
	},
	hostidentity.IdentitiesViewSchemaID: {
		ViewSchemaID:    hostidentity.IdentitiesViewSchemaID,
		Owner:           "entities",
		RecordFamily:    "identity",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeIdentity,
		AllowRawCapture: false,
	},
	indicators.ViewSchemaID: {
		ViewSchemaID:    indicators.ViewSchemaID,
		Owner:           "indicators",
		RecordFamily:    "indicator",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeIndicator,
		AllowRawCapture: false,
	},
	"cartulary.view.evidence.v1": {
		ViewSchemaID: "cartulary.view.evidence.v1",
		Owner:        "evidence",
		RecordFamily: "evidence",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeEvidence,
	},
	artifacts.NotesViewSchemaID: {
		ViewSchemaID: artifacts.NotesViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:note",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeNoteArtifact,
	},
	"cartulary.view.assessments.v1": {
		ViewSchemaID: "cartulary.view.assessments.v1",
		Owner:        "assessments",
		RecordFamily: "assessment",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeAssessment,
	},
	"cartulary.view.task_requests.v1": {
		ViewSchemaID: "cartulary.view.task_requests.v1",
		Owner:        "tasksdecisions/links",
		RecordFamily: "task_request",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeTask,
	},
	"cartulary.view.decisions.v1": {
		ViewSchemaID: "cartulary.view.decisions.v1",
		Owner:        "tasksdecisions/links",
		RecordFamily: "decision",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeDecision,
	},
	"cartulary.view.parties.v1": {
		ViewSchemaID: "cartulary.view.parties.v1",
		Owner:        "parties",
		RecordFamily: "party",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeParty,
	},
	artifacts.CommLogViewSchemaID: {
		ViewSchemaID: artifacts.CommLogViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:comm_log",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.HandoffViewSchemaID: {
		ViewSchemaID: artifacts.HandoffViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:handoff",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.StatusReviewViewSchemaID: {
		ViewSchemaID: artifacts.StatusReviewViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:status_review",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.LessonViewSchemaID: {
		ViewSchemaID: artifacts.LessonViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:lesson",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.FindingsViewSchemaID: {
		ViewSchemaID: artifacts.FindingsViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:finding",
		ApplyStatus:  applyStatusSupportedWhenAvailable,
	},
	artifacts.InvestigativeQueriesViewSchemaID: {
		ViewSchemaID: artifacts.InvestigativeQueriesViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:investigative_query",
		ApplyStatus:  applyStatusSupportedWhenAvailable,
	},
	artifacts.ForensicKeywordsViewSchemaID: {
		ViewSchemaID: artifacts.ForensicKeywordsViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:forensic_keyword",
		ApplyStatus:  applyStatusSupportedWhenAvailable,
	},
}
