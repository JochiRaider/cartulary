package hostidentity

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EligibleAlias struct {
	RecordID uuid.UUID
	RawText  string
}

type SourceFacts struct{ records *records.Store }

func NewSourceFacts() *SourceFacts {
	return &SourceFacts{records: records.NewStore()}
}

func (*SourceFacts) ListEligibleAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) ([]EligibleAlias, error) {
	var query string
	switch entityType {
	case "host":
		query = `SELECT ea.record_id, ea.raw_text FROM entity_aliases ea JOIN hosts h ON h.record_id = ea.record_id AND h.incident_id = ea.incident_id JOIN records r ON r.record_id = h.record_id AND r.deleted_at IS NULL WHERE ea.incident_id = $1 AND ea.entity_type = $2 AND ea.deleted_at IS NULL AND h.host_state IN ('stub', 'canonical') ORDER BY ea.record_id, ea.created_at, ea.entity_alias_id`
	case "identity":
		query = `SELECT ea.record_id, ea.raw_text FROM entity_aliases ea JOIN identities i ON i.record_id = ea.record_id AND i.incident_id = ea.incident_id JOIN records r ON r.record_id = i.record_id AND r.deleted_at IS NULL WHERE ea.incident_id = $1 AND ea.entity_type = $2 AND ea.deleted_at IS NULL AND i.identity_state IN ('stub', 'canonical') ORDER BY ea.record_id, ea.created_at, ea.entity_alias_id`
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

// AdmitResolvedTargetTx protects the current entity target before mutation.
// ValidateResolvedTargetTx remains available to consumers already holding the
// complete reference set, such as Assessment creation.
func (s *SourceFacts) AdmitResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	if entityType != "host" && entityType != "identity" {
		return ErrHostIdentityRecordNotFound
	}
	envelope, err := s.records.LoadEnvelopeForNoKeyUpdateTx(ctx, tx, recordID)
	if errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		return ErrHostIdentityRecordNotFound
	}
	if err != nil {
		return err
	}
	if envelope.IncidentID != incidentID || envelope.RecordType != entityType || envelope.DeletedAt != nil {
		return ErrHostIdentityRecordNotFound
	}
	return s.ValidateResolvedTargetTx(ctx, tx, incidentID, entityType, recordID)
}

// InspectResolvedTargetTx checks preview eligibility without reserving a target for a write.
func (s *SourceFacts) InspectResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	if entityType != "host" && entityType != "identity" {
		return ErrHostIdentityRecordNotFound
	}
	envelope, err := s.records.LoadEnvelopeTx(ctx, tx, recordID, false)
	if errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		return ErrHostIdentityRecordNotFound
	}
	if err != nil {
		return err
	}
	if envelope.IncidentID != incidentID || envelope.RecordType != entityType || envelope.DeletedAt != nil {
		return ErrHostIdentityRecordNotFound
	}
	return s.ValidateResolvedTargetTx(ctx, tx, incidentID, entityType, recordID)
}

func (*SourceFacts) ValidateResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	var exists bool
	var query string
	switch entityType {
	case "host":
		query = `SELECT EXISTS (SELECT 1 FROM hosts h JOIN records r ON r.record_id = h.record_id AND r.deleted_at IS NULL WHERE h.record_id = $1 AND h.incident_id = $2 AND h.host_state IN ('stub', 'canonical'))`
	case "identity":
		query = `SELECT EXISTS (SELECT 1 FROM identities i JOIN records r ON r.record_id = i.record_id AND r.deleted_at IS NULL WHERE i.record_id = $1 AND i.incident_id = $2 AND i.identity_state IN ('stub', 'canonical'))`
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
