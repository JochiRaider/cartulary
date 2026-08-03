package projections

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func NewTimelineProvider(descriptor ProviderDescriptor, source TimelineSource) Provider {
	return Provider{
		descriptor: descriptor,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshTimelineTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentTimelineTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewHostProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshHostTxCore, (*Store).rebuildIncidentHostsTxCore)
}

func NewIdentityProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshIdentityTxCore, (*Store).rebuildIncidentIdentitiesTxCore)
}

func NewIndicatorProvider(descriptor ProviderDescriptor, source IndicatorSource) Provider {
	return Provider{
		descriptor: descriptor,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshIndicatorTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewAssessmentProvider(descriptor ProviderDescriptor, source AssessmentSource) Provider {
	return Provider{
		descriptor: descriptor,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshAssessmentTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewArtifactProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshArtifactTxCore, (*Store).rebuildIncidentArtifactsTxCore)
}

func NewEvidenceProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshEvidenceTxCore, (*Store).rebuildIncidentEvidenceTxCore)
}

func NewPartyProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshPartyTxCore, (*Store).rebuildIncidentPartiesTxCore)
}

func NewTaskRequestProvider(descriptor ProviderDescriptor, source TaskDecisionSource) Provider {
	return Provider{
		descriptor: descriptor,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshTaskRequestTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func NewDecisionProvider(descriptor ProviderDescriptor, source TaskDecisionSource) Provider {
	return Provider{
		descriptor: descriptor,
		refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return store.refreshDecisionTxCore(ctx, tx, recordID, source)
		},
		rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID, source)
		},
	}
}

func providerWithHandlers(
	descriptor ProviderDescriptor,
	refresh func(*Store, context.Context, pgx.Tx, uuid.UUID) error,
	rebuild func(*Store, context.Context, pgx.Tx, uuid.UUID) error,
) Provider {
	provider := Provider{descriptor: descriptor}
	if refresh != nil {
		provider.refreshRowTx = func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
			return refresh(store, ctx, tx, recordID)
		}
	}
	if rebuild != nil {
		provider.rebuildIncidentTx = func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
			return rebuild(store, ctx, tx, incidentID)
		}
	}
	return provider
}
