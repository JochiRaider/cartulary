package incidents

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
)

type incidentBundleImportFinalizer struct {
	preferenceBootstrap bootstrapport.Writer
}

func NewIncidentBundleImportFinalizer(preferenceBootstrap bootstrapport.Writer) (importfinalizerport.Finalizer, error) {
	if isNilIncidentBundleFinalizerDependency(preferenceBootstrap) {
		return nil, errors.New("incidents: workbook preference bootstrap dependency is required")
	}
	return &incidentBundleImportFinalizer{
		preferenceBootstrap: preferenceBootstrap,
	}, nil
}

func (f *incidentBundleImportFinalizer) FinalizeIncidentBundleImportTx(
	ctx context.Context,
	tx pgx.Tx,
	params importfinalizerport.Params,
) error {
	if f == nil || isNilIncidentBundleFinalizerDependency(f.preferenceBootstrap) {
		return errors.New("incidents: workbook preference bootstrap dependency is required")
	}
	if isNilIncidentBundleFinalizerDependency(tx) {
		return errors.New("incidents: incident bundle finalization transaction is required")
	}
	if params.IncidentID == uuid.Nil {
		return errors.New("incidents: incident bundle finalization incident ID is required")
	}
	if params.SubmittedByUserID == uuid.Nil {
		return errors.New("incidents: incident bundle finalization submitter ID is required")
	}
	if params.PublishedAt.IsZero() {
		return errors.New("incidents: incident bundle finalization publication time is required")
	}
	publishedAt := params.PublishedAt.UTC()

	txRepository := newRepository(tx)
	initialAdmin, err := txRepository.getIncidentBundleInitialAdminForUpdate(ctx, params.SubmittedByUserID)
	if err != nil {
		return err
	}
	if !initialAdmin.IsActive || !initialAdmin.IsDeploymentAdmin {
		return importfinalizerport.ErrInitialAdminUnavailable
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

	if err := f.preferenceBootstrap.InsertInitialTx(
		ctx,
		tx,
		bootstrapport.InitialPreferenceInput{
			IncidentID:      params.IncidentID,
			UserID:          params.SubmittedByUserID,
			CommitTimestamp: publishedAt,
		},
	); err != nil {
		return err
	}

	if _, err := insertAuditEvent(ctx, tx, auditEvent{
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

func isNilIncidentBundleFinalizerDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
