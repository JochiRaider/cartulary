package incidents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/workbookpreferences"
)

type incidentBundleImportFinalizer struct {
	preferenceBootstrap PreferenceBootstrapPort
}

func NewIncidentBundleImportFinalizer() IncidentBundleImportFinalizer {
	return &incidentBundleImportFinalizer{
		preferenceBootstrap: workbookpreferences.NewBootstrap(),
	}
}

func (f *incidentBundleImportFinalizer) FinalizeIncidentBundleImportTx(
	ctx context.Context,
	tx pgx.Tx,
	params IncidentBundleImportFinalizationParams,
) error {
	publishedAt := params.PublishedAt
	if publishedAt.IsZero() {
		publishedAt = time.Now().UTC()
	}
	publishedAt = publishedAt.UTC()

	txRepository := newRepository(tx)
	initialAdmin, err := txRepository.getIncidentBundleInitialAdminForUpdate(ctx, params.SubmittedByUserID)
	if err != nil {
		return err
	}
	if !initialAdmin.IsActive || !initialAdmin.IsDeploymentAdmin {
		return ErrInitialAdminUnavailable
	}

	membership, err := txRepository.createBootstrapMembership(ctx, createBootstrapMembershipPersistenceParams{
		IncidentID:  params.IncidentID,
		UserID:      params.SubmittedByUserID,
		JoinedAt:    publishedAt,
		Role:        "admin",
		DisplayName: initialAdmin.DisplayName,
	})
	if err != nil {
		return err
	}

	if err := f.preferenceBootstrap.BootstrapIncidentPreferencesTx(
		ctx,
		tx,
		params.IncidentID,
		params.SubmittedByUserID,
		publishedAt,
	); err != nil {
		return err
	}

	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &params.SubmittedByUserID,
		TargetUserID: &params.SubmittedByUserID,
		IncidentID:   &params.IncidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_created",
		ClientTxnID:  params.ClientTxnID,
		RequestID:    params.RequestID,
		AfterJSON:    BuildMembershipResource(membership),
		PublicSource: "system",
	}); err != nil {
		return err
	}
	return nil
}

func ImportBundleRequestID(jobID uuid.UUID) string {
	return "incident_bundle_import:" + jobID.String()
}
