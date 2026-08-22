package hostidentity

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EligibleAlias struct {
	RecordID uuid.UUID
	RawText  string
}

type SourceFacts struct{}

func NewSourceFacts() *SourceFacts {
	return &SourceFacts{}
}

func (*SourceFacts) ListEligibleAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) ([]EligibleAlias, error) {
	var query string
	switch entityType {
	case "host":
		query = `SELECT ea.record_id, ea.raw_text FROM entity_aliases ea JOIN hosts h ON h.record_id = ea.record_id AND h.incident_id = ea.incident_id WHERE ea.incident_id = $1 AND ea.entity_type = $2 AND ea.deleted_at IS NULL AND h.host_state IN ('stub', 'canonical') ORDER BY ea.record_id, ea.created_at, ea.entity_alias_id`
	case "identity":
		query = `SELECT ea.record_id, ea.raw_text FROM entity_aliases ea JOIN identities i ON i.record_id = ea.record_id AND i.incident_id = ea.incident_id WHERE ea.incident_id = $1 AND ea.entity_type = $2 AND ea.deleted_at IS NULL AND i.identity_state IN ('stub', 'canonical') ORDER BY ea.record_id, ea.created_at, ea.entity_alias_id`
	default:
		return []EligibleAlias{}, nil
	}
	rows, err := tx.Query(ctx, query, incidentID, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]EligibleAlias, 0)
	for rows.Next() {
		var alias EligibleAlias
		if err := rows.Scan(&alias.RecordID, &alias.RawText); err != nil {
			return nil, err
		}
		result = append(result, alias)
	}
	return result, rows.Err()
}

func (*SourceFacts) ValidateResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	var exists bool
	var query string
	switch entityType {
	case "host":
		query = `SELECT EXISTS (SELECT 1 FROM hosts WHERE record_id = $1 AND incident_id = $2 AND host_state IN ('stub', 'canonical'))`
	case "identity":
		query = `SELECT EXISTS (SELECT 1 FROM identities WHERE record_id = $1 AND incident_id = $2 AND identity_state IN ('stub', 'canonical'))`
	default:
		return ErrHostIdentityRecordNotFound
	}
	if err := tx.QueryRow(ctx, query, recordID, incidentID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrHostIdentityRecordNotFound
	}
	return nil
}
