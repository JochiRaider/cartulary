package projection

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
)

type Source struct{}

func NewSource() *Source { return &Source{} }

func (*Source) LoadProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (partyprojection.ProjectionInput, bool, error) {
	input, err := scanProjectionInput(tx.QueryRow(ctx, partyProjectionSourceSQL+`
 WHERE p.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return partyprojection.ProjectionInput{}, false, nil
	}
	if err != nil {
		return partyprojection.ProjectionInput{}, false, fmt.Errorf("load Party projection input: %w", err)
	}
	return input, true, nil
}

func (*Source) ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (partyprojection.ProjectionInputPage, error) {
	if limit < 1 || limit > 1000 {
		return partyprojection.ProjectionInputPage{}, fmt.Errorf("party projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, partyProjectionSourceSQL+`
 WHERE p.incident_id = $1
   AND ($2::uuid IS NULL OR p.record_id > $2)
 ORDER BY p.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return partyprojection.ProjectionInputPage{}, fmt.Errorf("list Party projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]partyprojection.ProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanProjectionInput(rows)
		if scanErr != nil {
			return partyprojection.ProjectionInputPage{}, fmt.Errorf("scan Party projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return partyprojection.ProjectionInputPage{}, fmt.Errorf("iterate Party projection inputs: %w", err)
	}
	page := partyprojection.ProjectionInputPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanProjectionInput(scanner interface{ Scan(...any) error }) (partyprojection.ProjectionInput, error) {
	var (
		input            partyprojection.ProjectionInput
		displayName      pgtype.Text
		partyKind        pgtype.Text
		organizationName pgtype.Text
		roleTitle        pgtype.Text
		primaryEmail     pgtype.Text
		timezoneName     pgtype.Text
		externalRef      pgtype.Text
		notes            pgtype.Text
	)
	if err := scanner.Scan(
		&input.RecordID, &input.IncidentID, &input.RowVersion, &displayName,
		&partyKind, &organizationName, &roleTitle, &primaryEmail,
		&timezoneName, &externalRef, &notes, &input.UpdatedAt,
	); err != nil {
		return partyprojection.ProjectionInput{}, err
	}
	input.DisplayName = partyTextPointer(displayName)
	input.PartyKind = partyTextPointer(partyKind)
	input.OrganizationName = partyTextPointer(organizationName)
	input.RoleTitle = partyTextPointer(roleTitle)
	input.PrimaryEmail = partyTextPointer(primaryEmail)
	input.TimezoneName = partyTextPointer(timezoneName)
	input.ExternalRef = partyTextPointer(externalRef)
	input.Notes = partyTextPointer(notes)
	input.UpdatedAt = input.UpdatedAt.UTC()
	return input, nil
}

func partyTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

const partyProjectionSourceSQL = `
SELECT
    p.record_id,
    p.incident_id,
    r.row_version,
    p.display_name,
    p.party_kind,
    p.organization_name,
    p.role_title,
    p.primary_email,
    p.timezone_name,
    p.external_ref,
    p.notes,
    p.updated_at
  FROM parties p
  JOIN records r
    ON r.incident_id = p.incident_id
   AND r.record_id = p.record_id
   AND r.deleted_at IS NULL`
