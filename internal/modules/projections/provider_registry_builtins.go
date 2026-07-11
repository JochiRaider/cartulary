package projections

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
)

func builtInProjectionProviders() []projectionProvider {
	return []projectionProvider{
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "timeline",
				SourceOwnerKey:            "timeline",
				ViewSchemaIDs:             []string{timelineViewSchemaID},
				SourceRecordTypes:         []string{"timeline_event"},
				SourceAuthorityModules:    []string{"entities", "evidence", "links", "revisions", "timeline"},
				ProjectionTableFamilies:   []string{"timeline_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/timeline/workbookprojection"},
				CharacterizationRefs: []string{"internal/modules/timeline/phase3_projection_contract_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentTimelineTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "host",
				SourceOwnerKey:            "entities",
				ViewSchemaIDs:             []string{hostsViewSchemaID},
				SourceRecordTypes:         []string{"host"},
				SourceAuthorityModules:    []string{"entities", "evidence", "links", "revisions"},
				ProjectionTableFamilies:   []string{"host_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/entities/hostidentity"},
				RebuildAfter:         []string{"timeline"},
				CharacterizationRefs: []string{"internal/modules/entities/phase4_integration_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshHostTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentHostsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "identity",
				SourceOwnerKey:            "entities",
				ViewSchemaIDs:             []string{identitiesViewSchemaID},
				SourceRecordTypes:         []string{"identity"},
				SourceAuthorityModules:    []string{"entities", "evidence", "links", "revisions"},
				ProjectionTableFamilies:   []string{"identity_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/entities/hostidentity"},
				RebuildAfter:         []string{"host"},
				CharacterizationRefs: []string{"internal/modules/entities/phase4_integration_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshIdentityTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "indicator",
				SourceOwnerKey:            "indicators",
				ViewSchemaIDs:             []string{indicatorsViewSchemaID},
				SourceRecordTypes:         []string{"indicator"},
				SourceAuthorityModules:    []string{"indicators", "links", "revisions"},
				ProjectionTableFamilies:   []string{"indicator_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/indicators"},
				RebuildAfter:         []string{"identity"},
				CharacterizationRefs: []string{"internal/modules/indicators/phase9_indicators_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "assessment",
				SourceOwnerKey:            "assessments",
				ViewSchemaIDs:             []string{assessmentsViewSchemaID},
				SourceRecordTypes:         []string{"assessment"},
				SourceAuthorityModules:    []string{"assessments", "links", "revisions"},
				ProjectionTableFamilies:   []string{"assessment_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        assessmentprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/assessments"},
				RebuildAfter:         []string{"indicator"},
				CharacterizationRefs: []string{"internal/modules/assessments/phase9_assessment_contract_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshAssessmentTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:  projectionProviderDescriptorSchemaVersion,
				Status:         ProviderStatusActive,
				ProviderKey:    "artifact",
				SourceOwnerKey: "artifacts",
				ViewSchemaIDs: []string{
					notesViewSchemaID,
					commLogViewSchemaID,
					handoffViewSchemaID,
					statusReviewViewSchemaID,
					lessonViewSchemaID,
					findingsViewSchemaID,
					investigativeQueriesViewSchemaID,
					forensicKeywordsViewSchemaID,
				},
				SourceRecordTypes:         []string{"artifact"},
				SourceAuthorityModules:    []string{"artifacts", "links", "parties", "revisions"},
				ProjectionTableFamilies:   []string{"artifact_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        artifactprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/artifacts", "internal/modules/artifacts/linkednotes", "internal/modules/workbook"},
				RebuildAfter:         []string{"assessment"},
				CharacterizationRefs: []string{"internal/modules/workbook/phase9_coordination_surfaces_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshArtifactTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentArtifactsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "evidence",
				SourceOwnerKey:            "evidence",
				ViewSchemaIDs:             []string{evidenceViewSchemaID},
				SourceRecordTypes:         []string{"evidence"},
				SourceAuthorityModules:    []string{"evidence", "revisions"},
				ProjectionTableFamilies:   []string{"evidence_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        evidenceprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/evidence"},
				RebuildAfter:         []string{"artifact"},
				CharacterizationRefs: []string{"internal/modules/evidence/phase5_integration_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshEvidenceTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "party",
				SourceOwnerKey:            "parties",
				ViewSchemaIDs:             []string{partiesViewSchemaID},
				SourceRecordTypes:         []string{"party"},
				SourceAuthorityModules:    []string{"parties", "revisions"},
				ProjectionTableFamilies:   []string{"party_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        partyprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/parties"},
				RebuildAfter:         []string{"evidence"},
				CharacterizationRefs: []string{"internal/modules/workbook/phase9_parties_integration_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshPartyTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "task_request",
				SourceOwnerKey:            "tasksdecisions",
				ViewSchemaIDs:             []string{taskRequestsViewSchemaID},
				SourceRecordTypes:         []string{"task_request"},
				SourceAuthorityModules:    []string{"links", "revisions", "tasksdecisions"},
				ProjectionTableFamilies:   []string{"task_request_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        taskdecisionprojection.TaskRequestQuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/tasksdecisions"},
				RebuildAfter:         []string{"party"},
				CharacterizationRefs: []string{"internal/modules/tasksdecisions/phase9_task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshTaskRequestTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "decision",
				SourceOwnerKey:            "tasksdecisions",
				ViewSchemaIDs:             []string{decisionsViewSchemaID},
				SourceRecordTypes:         []string{"decision"},
				SourceAuthorityModules:    []string{"links", "revisions", "tasksdecisions"},
				ProjectionTableFamilies:   []string{"decision_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        taskdecisionprojection.DecisionQuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/tasksdecisions"},
				RebuildAfter:         []string{"task_request"},
				CharacterizationRefs: []string{"internal/modules/tasksdecisions/phase9_task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshDecisionTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID)
			},
		},
	}
}
