package projectionprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
)

func (*source) LoadIdentityProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (projectioncontract.IdentityProjectionInput, bool, error) {
	input, err := scanIdentityProjectionInput(tx.QueryRow(ctx, identityProjectionSourceSQL+`
 WHERE i.record_id = $1
   AND r.deleted_at IS NULL
   AND i.identity_state IN ('stub', 'canonical')
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return projectioncontract.IdentityProjectionInput{}, false, nil
	}
	if err != nil {
		return projectioncontract.IdentityProjectionInput{}, false, fmt.Errorf("load identity projection input: %w", err)
	}
	return input, true, nil
}

func (*source) ListIdentityProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (projectioncontract.IdentityProjectionPage, error) {
	if limit < 1 || limit > 1000 {
		return projectioncontract.IdentityProjectionPage{}, fmt.Errorf("identity projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, identityProjectionSourceSQL+`
 WHERE i.incident_id = $1
   AND ($2::uuid IS NULL OR i.record_id > $2)
   AND r.deleted_at IS NULL
   AND i.identity_state IN ('stub', 'canonical')
 ORDER BY i.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return projectioncontract.IdentityProjectionPage{}, fmt.Errorf("list identity projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]projectioncontract.IdentityProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanIdentityProjectionInput(rows)
		if scanErr != nil {
			return projectioncontract.IdentityProjectionPage{}, fmt.Errorf("scan identity projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return projectioncontract.IdentityProjectionPage{}, fmt.Errorf("iterate identity projection inputs: %w", err)
	}

	page := projectioncontract.IdentityProjectionPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanIdentityProjectionInput(scanner interface{ Scan(...any) error }) (projectioncontract.IdentityProjectionInput, error) {
	var (
		input            projectioncontract.IdentityProjectionInput
		upn              pgtype.Text
		email            pgtype.Text
		samAccountName   pgtype.Text
		privilegeLevel   pgtype.Text
		mfaState         pgtype.Text
		resetStatus      pgtype.Text
		linkedEventCount int32
		evidenceCount    int32
	)
	if err := scanner.Scan(
		&input.RecordID,
		&input.IncidentID,
		&input.RowVersion,
		&input.DisplayName,
		&upn,
		&email,
		&samAccountName,
		&input.IdentityState,
		&linkedEventCount,
		&evidenceCount,
		&privilegeLevel,
		&mfaState,
		&resetStatus,
		&input.EditedAt,
	); err != nil {
		return projectioncontract.IdentityProjectionInput{}, err
	}
	input.UPN = textPointer(upn)
	input.Email = textPointer(email)
	input.SamAccountName = textPointer(samAccountName)
	input.LinkedEventCount = int(linkedEventCount)
	input.EvidenceCount = int(evidenceCount)
	input.PrivilegeLevel = textPointer(privilegeLevel)
	input.MFAState = textPointer(mfaState)
	input.ResetStatus = textPointer(resetStatus)
	return input, nil
}

const identityProjectionSourceSQL = `
SELECT
    i.record_id,
    i.incident_id,
    r.row_version,
    i.display_name,
    i.upn,
    i.email::text,
    i.sam_account_name,
    i.identity_state,
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN records source_record
            ON source_record.record_id = l.src_record_id
           AND source_record.record_type = 'timeline_event'
           AND source_record.deleted_at IS NULL
         WHERE l.incident_id = i.incident_id
           AND l.dst_record_id = i.record_id
           AND l.link_type = 'observed_as_identity'
    ),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = i.incident_id
           AND l.src_record_id = i.record_id
           AND l.link_type = 'attached_evidence'
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    i.privilege_level,
    i.mfa_state,
    i.reset_status,
    r.updated_at
  FROM identities i
  JOIN records r
    ON r.record_id = i.record_id`
