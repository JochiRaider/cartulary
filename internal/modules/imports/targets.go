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
	applyStatusSupportedWhenClaimed   = "supported_when_claimed"

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

	applyFacadeNetworkFlow = "network_flow_import_facade_v1"
)

type importTarget struct {
	TargetKind         string
	ViewSchemaID       string
	ExtensionProfileID string
	Owner              string
	RecordFamily       string
	ApplyStatus        string
	CreateFacade       string
	ApplyFacade        string
	AllowRawCapture    bool
	AllowCustomAttrs   bool
}

func (target importTarget) importable(profileClaimed func(string) bool) bool {
	switch target.ApplyStatus {
	case applyStatusSupported:
		return true
	case applyStatusSupportedWhenClaimed:
		return profileClaimed != nil && profileClaimed(target.ExtensionProfileID)
	default:
		return false
	}
}

func (target importTarget) readyCheckImportable() bool {
	return target.ApplyStatus == applyStatusSupported || target.ApplyStatus == applyStatusSupportedWhenClaimed
}

func (target importTarget) ownerCreateFacadeAvailable() bool {
	return target.CreateFacade != ""
}

func (target importTarget) ownerApplyFacadeAvailable() bool {
	return target.ApplyFacade != ""
}

func lookupImportTarget(viewSchemaID string) (importTarget, bool) {
	target, ok := importTargets[viewSchemaID]
	return target, ok
}

func lookupApprovedImportTarget(mapping ApprovedMapping) (importTarget, bool) {
	switch mapping.targetKindOrDefault() {
	case ImportTargetKindViewSchema:
		return lookupImportTarget(mapping.TargetViewSchemaID)
	case ImportTargetKindNetworkFlowTable:
		target, ok := analyticalImportTargets[analyticalImportTargetKey{
			TargetKind:         mapping.TargetKind,
			ExtensionProfileID: mapping.ExtensionProfileID,
		}]
		return target, ok
	default:
		return importTarget{}, false
	}
}

type analyticalImportTargetKey struct {
	TargetKind         string
	ExtensionProfileID string
}

var importTargets = map[string]importTarget{
	timeline.TimelineViewSchemaID: {
		TargetKind:       ImportTargetKindViewSchema,
		ViewSchemaID:     timeline.TimelineViewSchemaID,
		Owner:            "timeline",
		RecordFamily:     "timeline_event",
		ApplyStatus:      applyStatusSupported,
		CreateFacade:     createFacadeTimeline,
		AllowRawCapture:  true,
		AllowCustomAttrs: false,
	},
	hostidentity.HostsViewSchemaID: {
		TargetKind:      ImportTargetKindViewSchema,
		ViewSchemaID:    hostidentity.HostsViewSchemaID,
		Owner:           "entities",
		RecordFamily:    "host",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeHost,
		AllowRawCapture: false,
	},
	hostidentity.IdentitiesViewSchemaID: {
		TargetKind:      ImportTargetKindViewSchema,
		ViewSchemaID:    hostidentity.IdentitiesViewSchemaID,
		Owner:           "entities",
		RecordFamily:    "identity",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeIdentity,
		AllowRawCapture: false,
	},
	indicators.ViewSchemaID: {
		TargetKind:      ImportTargetKindViewSchema,
		ViewSchemaID:    indicators.ViewSchemaID,
		Owner:           "indicators",
		RecordFamily:    "indicator",
		ApplyStatus:     applyStatusSupported,
		CreateFacade:    createFacadeIndicator,
		AllowRawCapture: false,
	},
	"cartulary.view.evidence.v1": {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: "cartulary.view.evidence.v1",
		Owner:        "evidence",
		RecordFamily: "evidence",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeEvidence,
	},
	artifacts.NotesViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.NotesViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:note",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeNoteArtifact,
	},
	"cartulary.view.assessments.v1": {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: "cartulary.view.assessments.v1",
		Owner:        "assessments",
		RecordFamily: "assessment",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeAssessment,
	},
	"cartulary.view.task_requests.v1": {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: "cartulary.view.task_requests.v1",
		Owner:        "tasksdecisions/links",
		RecordFamily: "task_request",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeTask,
	},
	"cartulary.view.decisions.v1": {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: "cartulary.view.decisions.v1",
		Owner:        "tasksdecisions/links",
		RecordFamily: "decision",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeDecision,
	},
	"cartulary.view.parties.v1": {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: "cartulary.view.parties.v1",
		Owner:        "parties",
		RecordFamily: "party",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeParty,
	},
	artifacts.CommLogViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.CommLogViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:comm_log",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.HandoffViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.HandoffViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:handoff",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.StatusReviewViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.StatusReviewViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:status_review",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.LessonViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.LessonViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:lesson",
		ApplyStatus:  applyStatusSupported,
		CreateFacade: createFacadeArtifact,
	},
	artifacts.FindingsViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.FindingsViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:finding",
		ApplyStatus:  applyStatusSupportedWhenAvailable,
	},
	artifacts.InvestigativeQueriesViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.InvestigativeQueriesViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:investigative_query",
		ApplyStatus:  applyStatusSupportedWhenAvailable,
	},
	artifacts.ForensicKeywordsViewSchemaID: {
		TargetKind:   ImportTargetKindViewSchema,
		ViewSchemaID: artifacts.ForensicKeywordsViewSchemaID,
		Owner:        "artifacts/links",
		RecordFamily: "artifact:forensic_keyword",
		ApplyStatus:  applyStatusSupportedWhenAvailable,
	},
}

var analyticalImportTargets = map[analyticalImportTargetKey]importTarget{
	{
		TargetKind:         ImportTargetKindNetworkFlowTable,
		ExtensionProfileID: NetworkFlowExtensionProfileID,
	}: {
		TargetKind:         ImportTargetKindNetworkFlowTable,
		ExtensionProfileID: NetworkFlowExtensionProfileID,
		Owner:              NetworkFlowExtensionProfileID,
		RecordFamily:       ImportTargetKindNetworkFlowTable,
		ApplyStatus:        applyStatusSupportedWhenClaimed,
		ApplyFacade:        applyFacadeNetworkFlow,
	},
}
