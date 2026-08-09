package projectionprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
)

type Source struct{}

func NewSource() *Source { return &Source{} }

func (*Source) LoadHostProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (workbookprojection.HostProjectionInput, bool, error) {
	input, err := scanHostProjectionInput(tx.QueryRow(ctx, hostProjectionSourceSQL+`
 WHERE h.record_id = $1
   AND r.deleted_at IS NULL
   AND h.host_state IN ('stub', 'canonical')
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return workbookprojection.HostProjectionInput{}, false, nil
	}
	if err != nil {
		return workbookprojection.HostProjectionInput{}, false, fmt.Errorf("load host projection input: %w", err)
	}
	return input, true, nil
}

func (*Source) ListHostProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (workbookprojection.HostProjectionPage, error) {
	if limit < 1 || limit > 1000 {
		return workbookprojection.HostProjectionPage{}, fmt.Errorf("host projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, hostProjectionSourceSQL+`
 WHERE h.incident_id = $1
   AND ($2::uuid IS NULL OR h.record_id > $2)
   AND r.deleted_at IS NULL
   AND h.host_state IN ('stub', 'canonical')
 ORDER BY h.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return workbookprojection.HostProjectionPage{}, fmt.Errorf("list host projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]workbookprojection.HostProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanHostProjectionInput(rows)
		if scanErr != nil {
			return workbookprojection.HostProjectionPage{}, fmt.Errorf("scan host projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return workbookprojection.HostProjectionPage{}, fmt.Errorf("iterate host projection inputs: %w", err)
	}

	page := workbookprojection.HostProjectionPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanHostProjectionInput(scanner interface{ Scan(...any) error }) (workbookprojection.HostProjectionInput, error) {
	var (
		input             workbookprojection.HostProjectionInput
		hostname          pgtype.Text
		location          pgtype.Text
		osPlatform        pgtype.Text
		businessOwner     pgtype.Text
		criticality       pgtype.Text
		containmentStatus pgtype.Text
		linkedEventCount  int32
		evidenceCount     int32
	)
	if err := scanner.Scan(
		&input.RecordID,
		&input.IncidentID,
		&input.RowVersion,
		&input.DisplayName,
		&hostname,
		&input.HostState,
		&linkedEventCount,
		&evidenceCount,
		&location,
		&osPlatform,
		&businessOwner,
		&criticality,
		&containmentStatus,
		&input.EditedAt,
	); err != nil {
		return workbookprojection.HostProjectionInput{}, err
	}
	input.Hostname = textPointer(hostname)
	input.LinkedEventCount = int(linkedEventCount)
	input.EvidenceCount = int(evidenceCount)
	input.Location = textPointer(location)
	input.OSPlatform = textPointer(osPlatform)
	input.BusinessOwner = textPointer(businessOwner)
	input.Criticality = textPointer(criticality)
	input.ContainmentStatus = textPointer(containmentStatus)
	return input, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

const hostProjectionSourceSQL = `
SELECT
    h.record_id,
    h.incident_id,
    r.row_version,
    h.display_name,
    h.hostname,
    h.host_state,
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN records source_record
            ON source_record.record_id = l.src_record_id
           AND source_record.record_type = 'timeline_event'
           AND source_record.deleted_at IS NULL
         WHERE l.incident_id = h.incident_id
           AND l.dst_record_id = h.record_id
           AND l.link_type = 'observed_on_host'
    ),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = h.incident_id
           AND l.src_record_id = h.record_id
           AND l.link_type = 'attached_evidence'
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    h.location,
    h.os_platform,
    h.business_owner,
    h.criticality,
    h.containment_status,
    r.updated_at
  FROM hosts h
  JOIN records r
    ON r.record_id = h.record_id`
