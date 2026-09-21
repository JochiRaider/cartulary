package hostidentity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ReadRecordLabelTx supplies source-owned labels without taking mutation locks.
func ReadRecordLabelTx(ctx context.Context, tx pgx.Tx, recordType string, recordID uuid.UUID) (string, error) {
	var label string
	var query string
	switch recordType {
	case "host":
		query = `SELECT display_name FROM hosts WHERE record_id=$1`
	case "identity":
		query = `SELECT display_name FROM identities WHERE record_id=$1`
	default:
		return "", fmt.Errorf("entities: unsupported label record type")
	}
	err := tx.QueryRow(ctx, query, recordID).Scan(&label)
	return label, err
}
