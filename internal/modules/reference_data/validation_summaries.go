package reference_data

import (
	"context"
	"errors"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (c *Coordinator) GetValidationSummary(ctx context.Context, id string) (*packformat.ValidationSummary, error) {
	if len(id) != 69 || !strings.HasPrefix(id, "rpvs_") {
		return nil, ErrNotFound
	}
	for _, ch := range id[5:] {
		if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f') {
			return nil, ErrNotFound
		}
	}
	var data []byte
	err := c.pool.QueryRow(ctx, `SELECT m.canonical_validation_summary FROM reference_pack_attempt_members m
JOIN reference_pack_attempts a USING(attempt_id) WHERE m.validation_summary_id=$1 AND a.completed_at IS NOT NULL
AND a.outcome='content_rejected' ORDER BY m.attempt_id,m.ordinal LIMIT 1`, id).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if "rpvs_"+packformat.Digest(data) != id {
		return nil, errHistoricalIntegrity
	}
	return packformat.DecodeValidationSummary(data)
}

func validateHistoricalDiagnostics(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}) error {
	after, ordinal := uuid.Nil, int64(0)
	for {
		var id uuid.UUID
		var next int64
		var code, check, summaryID string
		var data []byte
		err := db.QueryRow(ctx, `SELECT attempt_id,ordinal,failure_code,check_id,validation_summary_id,canonical_validation_summary
FROM reference_pack_attempt_members WHERE (attempt_id,ordinal)>($1::uuid,$2::bigint) AND verdict='content_rejected'
ORDER BY attempt_id,ordinal LIMIT 1`, after, ordinal).Scan(&id, &next, &code, &check, &summaryID, &data)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		summary, err := packformat.DecodeValidationSummary(data)
		if err != nil || summary.Result != "failed" || summary.Issues[0].Code != code || summary.Issues[0].CheckID != check || summaryID != "rpvs_"+packformat.Digest(data) {
			return errHistoricalIntegrity
		}
		after, ordinal = id, next
	}
	var broken bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_events e
WHERE convert_from(e.canonical_attestation,'UTF8')::jsonb->>'validation_summary_ref' IS NOT NULL
AND NOT EXISTS(SELECT 1 FROM reference_pack_attempts a JOIN reference_pack_attempt_members m USING(attempt_id)
WHERE a.operation_id=e.operation_id AND m.pack_key=e.pack_key AND m.pack_version=e.pack_version
AND a.outcome='content_rejected' AND m.validation_summary_id=convert_from(e.canonical_attestation,'UTF8')::jsonb->>'validation_summary_ref'))`).Scan(&broken); err != nil {
		return err
	}
	if broken {
		return errHistoricalIntegrity
	}
	return nil
}
