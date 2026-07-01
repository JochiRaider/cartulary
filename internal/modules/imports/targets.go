package imports

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

const (
	applyStatusSupported              = "supported"
	applyStatusSupportedWhenAvailable = "supported_when_implemented"

	createFacadeTimeline  = "timeline.import_create"
	createFacadeHost      = "entities.host.import_create"
	createFacadeIdentity  = "entities.identity.import_create"
	createFacadeIndicator = "entities.indicator.import_create"
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
	entities.HostsViewSchemaID: {
		ViewSchemaID:    entities.HostsViewSchemaID,
		Owner:           "entities",
		RecordFamily:    "host",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeHost,
		AllowRawCapture: false,
	},
	entities.IdentitiesViewSchemaID: {
		ViewSchemaID:    entities.IdentitiesViewSchemaID,
		Owner:           "entities",
		RecordFamily:    "identity",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeIdentity,
		AllowRawCapture: false,
	},
	entities.IndicatorsViewSchemaID: {
		ViewSchemaID:    entities.IndicatorsViewSchemaID,
		Owner:           "entities",
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
	},
	artifacts.NotesViewSchemaID: {
		ViewSchemaID: artifacts.NotesViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:note",
		ApplyStatus:  applyStatusSupported,
	},
	"cartulary.view.assessments.v1": {
		ViewSchemaID: "cartulary.view.assessments.v1",
		Owner:        "entities",
		RecordFamily: "assessment",
		ApplyStatus:  applyStatusSupported,
	},
	"cartulary.view.task_requests.v1": {
		ViewSchemaID: "cartulary.view.task_requests.v1",
		Owner:        "tasksdecisions/links",
		RecordFamily: "task_request",
		ApplyStatus:  applyStatusSupported,
	},
	"cartulary.view.decisions.v1": {
		ViewSchemaID: "cartulary.view.decisions.v1",
		Owner:        "tasksdecisions/links",
		RecordFamily: "decision",
		ApplyStatus:  applyStatusSupported,
	},
	"cartulary.view.parties.v1": {
		ViewSchemaID: "cartulary.view.parties.v1",
		Owner:        "entities",
		RecordFamily: "party",
		ApplyStatus:  applyStatusSupported,
	},
	artifacts.CommLogViewSchemaID: {
		ViewSchemaID: artifacts.CommLogViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:comm_log",
		ApplyStatus:  applyStatusSupported,
	},
	artifacts.HandoffViewSchemaID: {
		ViewSchemaID: artifacts.HandoffViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:handoff",
		ApplyStatus:  applyStatusSupported,
	},
	artifacts.StatusReviewViewSchemaID: {
		ViewSchemaID: artifacts.StatusReviewViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:status_review",
		ApplyStatus:  applyStatusSupported,
	},
	artifacts.LessonViewSchemaID: {
		ViewSchemaID: artifacts.LessonViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:lesson",
		ApplyStatus:  applyStatusSupported,
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
