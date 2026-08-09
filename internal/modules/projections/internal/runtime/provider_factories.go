package runtime

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
)

func NewTimelineProvider(descriptor ProviderDescriptor, source TimelineSource) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.TimelinePlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshTimelineTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentTimelineTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewHostProvider(descriptor ProviderDescriptor, sources ...entityprojection.SourceReader) Provider {
	var source entityprojection.SourceReader
	if len(sources) == 1 {
		source = sources[0]
	}
	return Provider{
		descriptor: descriptor,
		typedQuery: source != nil,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshHostTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentHostsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewIdentityProvider(descriptor ProviderDescriptor, sources ...entityprojection.SourceReader) Provider {
	var source entityprojection.SourceReader
	if len(sources) == 1 {
		source = sources[0]
	}
	return Provider{
		descriptor: descriptor,
		typedQuery: source != nil,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshIdentityTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewIndicatorProvider(descriptor ProviderDescriptor, source indicatorprojection.SourceReader) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.IndicatorPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshIndicatorTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewAssessmentProvider(descriptor ProviderDescriptor, source assessmentprojection.SourceReader) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.AssessmentPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshAssessmentTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewArtifactProvider(descriptor ProviderDescriptor, source artifactprojection.SourceReader) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.ArtifactPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshArtifactTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentArtifactsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewEvidenceProvider(descriptor ProviderDescriptor, source evidenceprojection.SourceReader) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.EvidencePlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshEvidenceTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewPartyProvider(descriptor ProviderDescriptor, source partyprojection.SourceReader) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.PartyPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshPartyTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewTaskRequestProvider(descriptor ProviderDescriptor, source TaskRequestSource) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.TaskRequestPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshTaskRequestTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewDecisionProvider(descriptor ProviderDescriptor, source DecisionSource) Provider {
	return Provider{
		descriptor: descriptor,
		queryPlans: queryengine.DecisionPlans(),
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshDecisionTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID, source)
		},
	}
}
