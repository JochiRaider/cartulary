package incidents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
)

var ErrInitialAdminUnavailable = errors.New("incidents: initial admin unavailable")

func (s *Store) FinalizeIncidentBundleImportTx(ctx context.Context, tx pgx.Tx, params IncidentBundleImportFinalizationParams) error {
	publishedAt := params.PublishedAt
	if publishedAt.IsZero() {
		publishedAt = time.Now().UTC()
	}
	publishedAt = publishedAt.UTC()

	var displayName string
	var isActive bool
	var isDeploymentAdmin bool
	err := tx.QueryRow(ctx, `
SELECT display_name, is_active, is_deployment_admin
  FROM users
 WHERE id = $1
 FOR UPDATE
`, params.SubmittedByUserID).Scan(&displayName, &isActive, &isDeploymentAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInitialAdminUnavailable
	}
	if err != nil {
		return fmt.Errorf("read incident bundle initial admin: %w", err)
	}
	if !isActive || !isDeploymentAdmin {
		return ErrInitialAdminUnavailable
	}

	membershipRow, err := sqlc.New(tx).CreateBootstrapIncidentMembership(ctx, sqlc.CreateBootstrapIncidentMembershipParams{
		IncidentID: pgUUID(params.IncidentID),
		UserID:     pgUUID(params.SubmittedByUserID),
		JoinedAt:   pgTimestamptz(publishedAt),
		Role:       "admin",
		Column5:    displayName,
	})
	if err != nil {
		return fmt.Errorf("insert imported incident bootstrap membership: %w", err)
	}
	membership, err := membershipRecordFromSQL(membershipRow)
	if err != nil {
		return err
	}

	if s.workbookBootstrap == nil {
		return errors.New("incidents: workbook bootstrap port is required for incident bundle import finalization")
	}
	if err := s.workbookBootstrap.BootstrapIncidentCreatePreferencesTx(ctx, tx, params.IncidentID, params.SubmittedByUserID, publishedAt); err != nil {
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
