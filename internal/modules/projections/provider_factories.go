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

func NewIndicatorProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, nil, (*Store).rebuildIncidentIndicatorsTxCore)
}

func NewAssessmentProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshAssessmentTxCore, (*Store).rebuildIncidentAssessmentsTxCore)
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

func NewTaskRequestProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshTaskRequestTxCore, (*Store).rebuildIncidentTaskRequestsTxCore)
}

func NewDecisionProvider(descriptor ProviderDescriptor) Provider {
	return providerWithHandlers(descriptor, (*Store).refreshDecisionTxCore, (*Store).rebuildIncidentDecisionsTxCore)
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
