package runtime

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

func newTimelineProvider(descriptor providercontract.ProviderDescriptor, source TimelineSource) Provider {
	return Provider{
		descriptor:                      descriptor,
		queryStrategy:                   queryStrategyCompiledPlan,
		queryPlans:                      queryengine.TimelinePlans(),
		evidenceAssociationEffectFields: []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"},
		loadEvidenceAssociationStateTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
			return store.LoadRowTx(ctx, tx, descriptor.ViewSchemaIDs[0], recordID)
		},
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshTimelineTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentTimelineTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newHostProvider(descriptor providercontract.ProviderDescriptor, source entitycontract.SourceReader) Provider {
	return Provider{
		descriptor:                      descriptor,
		queryStrategy:                   queryStrategySourceOwnerHydration,
		evidenceAssociationEffectFields: []string{"host.evidence_count"},
		loadEvidenceAssociationStateTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
			return loadHostEvidenceAssociationStateTx(ctx, store, tx, recordID)
		},
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshHostTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentHostsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newIdentityProvider(descriptor providercontract.ProviderDescriptor, source entitycontract.SourceReader) Provider {
	return Provider{
		descriptor:                      descriptor,
		queryStrategy:                   queryStrategySourceOwnerHydration,
		evidenceAssociationEffectFields: []string{"identity.evidence_count"},
		loadEvidenceAssociationStateTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
			return loadIdentityEvidenceAssociationStateTx(ctx, store, tx, recordID)
		},
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshIdentityTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newIndicatorProvider(descriptor providercontract.ProviderDescriptor, source indicatorprojection.SourceReader) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.IndicatorPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshIndicatorTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newAssessmentProvider(descriptor providercontract.ProviderDescriptor, source assessmentprojection.SourceReader) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.AssessmentPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshAssessmentTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newArtifactProvider(descriptor providercontract.ProviderDescriptor, source artifactprojection.SourceReader) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.ArtifactPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshArtifactTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentArtifactsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newEvidenceProvider(descriptor providercontract.ProviderDescriptor, source evidenceprojection.SourceReader) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.EvidencePlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshEvidenceTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newPartyProvider(descriptor providercontract.ProviderDescriptor, source partyprojection.SourceReader) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.PartyPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshPartyTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newTaskRequestProvider(descriptor providercontract.ProviderDescriptor, source TaskRequestSource) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.TaskRequestPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshTaskRequestTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func newDecisionProvider(descriptor providercontract.ProviderDescriptor, source DecisionSource) Provider {
	return Provider{
		descriptor:    descriptor,
		queryStrategy: queryStrategyCompiledPlan,
		queryPlans:    queryengine.DecisionPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshDecisionTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID, source)
		},
	}
}
