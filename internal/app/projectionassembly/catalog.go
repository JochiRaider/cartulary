package projectionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Bundle struct {
	Catalog       *projections.Catalog
	Query         *projections.QueryService
	Rebuild       *projections.RebuildService
	Coordinator   *projections.Coordinator
	Timeline      *projections.TimelineRows
	Entities      *projections.EntityRows
	Assessments   *projections.AssessmentRows
	Artifacts     *projections.ArtifactRows
	Evidence      *projections.EvidenceRows
	Parties       *projections.PartyRows
	TaskDecisions *projections.TaskDecisionRows
}

func NewBundle(pool postgres.DB, timelineSource projections.TimelineSource) (*Bundle, error) {
	catalog, err := NewCatalog(timelineSource)
	if err != nil {
		return nil, err
	}
	return &Bundle{
		Catalog:       catalog,
		Query:         projections.NewQueryService(pool, catalog),
		Rebuild:       projections.NewRebuildService(pool, catalog),
		Coordinator:   projections.NewCoordinator(pool, catalog),
		Timeline:      projections.NewTimelineRows(pool),
		Entities:      projections.NewEntityRows(pool),
		Assessments:   projections.NewAssessmentRows(pool, assessmentprojection.QuerySurfaces()...),
		Artifacts:     projections.NewArtifactRows(pool, artifactprojection.QuerySurfaces()...),
		Evidence:      projections.NewEvidenceRows(pool, evidenceprojection.QuerySurfaces()...),
		Parties:       projections.NewPartyRows(pool, partyprojection.QuerySurfaces()...),
		TaskDecisions: projections.NewTaskDecisionRows(pool, taskDecisionQuerySurfaces()...),
	}, nil
}

func taskDecisionQuerySurfaces() []providercontract.QuerySurface {
	surfaces := taskdecisionprojection.TaskRequestQuerySurfaces()
	return append(surfaces, taskdecisionprojection.DecisionQuerySurfaces()...)
}

func NewCatalog(timelineSource projections.TimelineSource) (*projections.Catalog, error) {
	providers := []projections.Provider{
		projections.NewTimelineProvider(descriptor(
			"timeline",
			"timeline",
			[]string{timeline.TimelineViewSchemaID},
			[]string{"timeline_event"},
			[]string{"entities", "evidence", "links", "revisions", "timeline"},
			[]string{"timeline_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			timelineprojection.QuerySurfaces(),
			nil,
			[]string{"internal/modules/timeline"},
			[]string{"internal/modules/timeline/projection_contract_test.go"},
		), timelineSource),
		projections.NewHostProvider(descriptor(
			"host",
			"entities",
			[]string{hostidentity.HostsViewSchemaID},
			[]string{"host"},
			[]string{"entities", "evidence", "links", "revisions"},
			[]string{"host_grid_projection"},
			projections.ProviderCapabilities{RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			nil,
			[]string{"timeline"},
			[]string{"internal/modules/entities/hostidentity"},
			[]string{"internal/modules/entities/resolution_integration_test.go"},
		)),
		projections.NewIdentityProvider(descriptor(
			"identity",
			"entities",
			[]string{hostidentity.IdentitiesViewSchemaID},
			[]string{"identity"},
			[]string{"entities", "evidence", "links", "revisions"},
			[]string{"identity_grid_projection"},
			projections.ProviderCapabilities{RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			nil,
			[]string{"host"},
			[]string{"internal/modules/entities/hostidentity"},
			[]string{"internal/modules/entities/resolution_integration_test.go"},
		)),
		projections.NewIndicatorProvider(descriptor(
			"indicator",
			"indicators",
			[]string{indicators.ViewSchemaID},
			[]string{"indicator"},
			[]string{"indicators", "links", "revisions"},
			[]string{"indicator_grid_projection"},
			projections.ProviderCapabilities{Query: true, RestoreRebuild: true, IncidentRebuild: true},
			indicatorprojection.QuerySurfaces(),
			[]string{"identity"},
			[]string{"internal/modules/indicators"},
			[]string{"internal/modules/indicators/indicators_test.go"},
		)),
		projections.NewAssessmentProvider(descriptor(
			"assessment",
			"assessments",
			[]string{assessments.AssessmentsViewSchemaID},
			[]string{"assessment"},
			[]string{"assessments", "links", "revisions"},
			[]string{"assessment_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			assessmentprojection.QuerySurfaces(),
			[]string{"indicator"},
			[]string{"internal/modules/assessments"},
			[]string{"internal/modules/assessments/assessment_contract_test.go", "internal/modules/projections/query_test.go"},
		)),
		projections.NewArtifactProvider(descriptor(
			"artifact",
			"artifacts",
			[]string{
				artifacts.NotesViewSchemaID,
				artifacts.CommLogViewSchemaID,
				artifacts.HandoffViewSchemaID,
				artifacts.StatusReviewViewSchemaID,
				artifacts.LessonViewSchemaID,
				artifacts.FindingsViewSchemaID,
				artifacts.InvestigativeQueriesViewSchemaID,
				artifacts.ForensicKeywordsViewSchemaID,
			},
			[]string{"artifact"},
			[]string{"artifacts", "links", "parties", "revisions"},
			[]string{"artifact_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			artifactprojection.QuerySurfaces(),
			[]string{"assessment"},
			[]string{"internal/modules/artifacts", "internal/modules/artifacts/linkednotes", "internal/modules/workbook"},
			[]string{"internal/modules/workbook/coordination_surfaces_test.go", "internal/modules/projections/query_test.go"},
		)),
		projections.NewEvidenceProvider(descriptor(
			"evidence",
			"evidence",
			[]string{evidence.ViewSchemaID},
			[]string{"evidence"},
			[]string{"evidence", "links", "revisions"},
			[]string{"evidence_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			evidenceprojection.QuerySurfaces(),
			[]string{"artifact"},
			[]string{"internal/modules/evidence"},
			[]string{"internal/modules/evidence/integration_test.go", "internal/modules/projections/query_test.go"},
		)),
		projections.NewPartyProvider(descriptor(
			"party",
			"parties",
			[]string{parties.ViewSchemaID},
			[]string{"party"},
			[]string{"parties", "revisions"},
			[]string{"party_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			partyprojection.QuerySurfaces(),
			[]string{"evidence"},
			[]string{"internal/modules/parties"},
			[]string{"internal/modules/workbook/parties_integration_test.go", "internal/modules/projections/query_test.go"},
		)),
		projections.NewTaskRequestProvider(descriptor(
			"task_request",
			"tasksdecisions",
			[]string{tasksdecisions.TaskRequestsViewSchemaID},
			[]string{"task_request"},
			[]string{"links", "revisions", "tasksdecisions"},
			[]string{"task_request_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			taskdecisionprojection.TaskRequestQuerySurfaces(),
			[]string{"party"},
			[]string{"internal/modules/tasksdecisions"},
			[]string{"internal/modules/tasksdecisions/task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
		)),
		projections.NewDecisionProvider(descriptor(
			"decision",
			"tasksdecisions",
			[]string{tasksdecisions.DecisionsViewSchemaID},
			[]string{"decision"},
			[]string{"links", "revisions", "tasksdecisions"},
			[]string{"decision_grid_projection"},
			projections.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
			taskdecisionprojection.DecisionQuerySurfaces(),
			[]string{"task_request"},
			[]string{"internal/modules/tasksdecisions"},
			[]string{"internal/modules/tasksdecisions/task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
		)),
	}
	catalog, err := projections.NewCatalog(providers)
	if err != nil {
		return nil, fmt.Errorf("assemble projection catalog: %w", err)
	}
	return catalog, nil
}

func descriptor(
	providerKey string,
	sourceOwner string,
	viewSchemaIDs []string,
	recordTypes []string,
	sourceAuthorities []string,
	tableFamilies []string,
	capabilities projections.ProviderCapabilities,
	querySurfaces []providercontract.QuerySurface,
	rebuildAfter []string,
	facadePackages []string,
	characterizationRefs []string,
) projections.ProviderDescriptor {
	return projections.ProviderDescriptor{
		SchemaVersion:             providercontract.DescriptorSchemaVersion,
		Status:                    projections.ProviderStatusActive,
		ProviderKey:               providerKey,
		SourceOwnerKey:            sourceOwner,
		ViewSchemaIDs:             viewSchemaIDs,
		SourceRecordTypes:         recordTypes,
		SourceAuthorityModules:    sourceAuthorities,
		ProjectionTableFamilies:   tableFamilies,
		ProjectionStorageOwnerKey: "projections",
		Capabilities:              capabilities,
		QuerySurfaces:             querySurfaces,
		RestoreRebuild:            projections.RestoreRebuildRequired,
		FacadePackages:            facadePackages,
		RebuildAfter:              rebuildAfter,
		CharacterizationRefs:      characterizationRefs,
	}
}
