package parties

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const ViewSchemaID = "cartulary.view.parties.v1"

type Store struct {
	pool          postgres.DB
	recordStore   *records.Store
	revisionStore *revisions.Store
	rowProjector  *projectionadapters.RowProjector
}

type FieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type CreateParams struct {
	Values map[string]FieldValue
}

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return "parties: invalid mutation request"
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:          pool,
		recordStore:   records.NewStore(),
		revisionStore: revisions.NewStore(pool),
		rowProjector:  projectionadapters.NewRowProjector(pool),
	}
}

func (s *Store) FindReusablePartyTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, params CreateParams) (uuid.UUID, bool, error) {
	if value := normalizedOptionalText(params.Values, "party.primary_email"); value != "" {
		recordID, found, err := findUniqueReusablePartyByFieldTx(ctx, tx, incidentID, "primary_email", value)
		if err != nil || found {
			return recordID, found, err
		}
	}
	if value := normalizedOptionalText(params.Values, "party.external_ref"); value != "" {
		return findUniqueReusablePartyByFieldTx(ctx, tx, incidentID, "external_ref", value)
	}
	return uuid.UUID{}, false, nil
}

func (s *Store) InsertPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, params CreateParams, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name, role_title,
    primary_email, timezone_name, external_ref, notes, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
`, recordID, incidentID,
		textValue(params.Values, "party.display_name"),
		textValue(params.Values, "party.party_kind"),
		nullableTextValue(params.Values, "party.organization_name"),
		nullableTextValue(params.Values, "party.role_title"),
		nullableTextValue(params.Values, "party.primary_email"),
		nullableTextValue(params.Values, "party.timezone_name"),
		nullableTextValue(params.Values, "party.external_ref"),
		nullableTextValue(params.Values, "party.notes"),
		now)
	if err != nil {
		return fmt.Errorf("insert party: %w", err)
	}
	return nil
}

func (s *Store) ApplyDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string, value FieldValue, now time.Time) (bool, error) {
	if !strings.HasPrefix(fieldKey, "party.") {
		return false, fmt.Errorf("parties: unsupported party field key %q", fieldKey)
	}
	column := strings.TrimPrefix(fieldKey, "party.")
	dbValue := directDBValue(value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE parties SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, column, column), recordID, dbValue, now)
	if err != nil {
		return false, fmt.Errorf("apply party direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) TouchPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE parties SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch party row: %w", err)
	}
	return nil
}

func ValidateCreateParams(params CreateParams) error {
	if !hasText(params.Values, "party.display_name") {
		return &ValidationError{Field: "party.display_name", ReasonCode: "missing_required_field"}
	}
	if !validText(params.Values, "party.party_kind", ValidKind) {
		return &ValidationError{Field: "party.party_kind", ReasonCode: "missing_required_field"}
	}
	return nil
}

func ValidKind(value string) bool {
	switch value {
	case "person", "team", "organization", "distribution_list", "other":
		return true
	default:
		return false
	}
}

func (s *Store) records() *records.Store {
	if s != nil && s.recordStore != nil {
		return s.recordStore
	}
	return records.NewStore()
}

func (s *Store) revisions() *revisions.Store {
	if s != nil && s.revisionStore != nil {
		return s.revisionStore
	}
	return revisions.NewStore()
}

func (s *Store) rowProjections() *projectionadapters.RowProjector {
	if s != nil && s.rowProjector != nil {
		return s.rowProjector
	}
	return projectionadapters.NewRowProjector(nil)
}

func findUniqueReusablePartyByFieldTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, column string, normalizedValue string) (uuid.UUID, bool, error) {
	if column != "primary_email" && column != "external_ref" {
		return uuid.UUID{}, false, fmt.Errorf("unsupported party reuse column %q", column)
	}
	rows, err := tx.Query(ctx, `
SELECT p.record_id
  FROM parties p
  JOIN records r
    ON r.incident_id = p.incident_id
   AND r.record_id = p.record_id
   AND r.record_type = 'party'
   AND r.deleted_at IS NULL
 WHERE p.incident_id = $1
   AND lower(trim(p.`+column+`)) = $2
 ORDER BY p.record_id ASC
 LIMIT 2
`, incidentID, normalizedValue)
	if err != nil {
		return uuid.UUID{}, false, fmt.Errorf("find reusable party: %w", err)
	}
	defer rows.Close()
	var matches []uuid.UUID
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("scan reusable party: %w", err)
		}
		matches = append(matches, recordID)
	}
	if err := rows.Err(); err != nil {
		return uuid.UUID{}, false, fmt.Errorf("iterate reusable parties: %w", err)
	}
	if len(matches) != 1 {
		return uuid.UUID{}, false, nil
	}
	return matches[0], true, nil
}

func normalizedOptionalText(values map[string]FieldValue, fieldKey string) string {
	value, ok := values[fieldKey]
	if !ok || value.Text == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*value.Text))
}

func directDBValue(value FieldValue) any {
	switch {
	case value.Text != nil:
		return *value.Text
	case value.Timestamp != nil:
		return value.Timestamp.UTC()
	case value.UUID != nil:
		return *value.UUID
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func textValue(values map[string]FieldValue, field string) string {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return ""
}

func nullableTextValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return nil
}

func hasText(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func validText(values map[string]FieldValue, field string, predicate func(string) bool) bool {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return false
	}
	return predicate(*value.Text)
}
