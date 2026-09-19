package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const MaximumRestoredNonterminalPageSize = 256

type RestoredNonterminalScope struct {
	JobKind                 string
	ExtensionOwnerProfileID string
}

type RestoredNonterminalJob struct {
	JobID              uuid.UUID
	IncidentID         *uuid.UUID
	HandlerPayloadJSON json.RawMessage
}

// ListRestoredNonterminalPageTx reads a complete page through the caller's
// quiescent restore transaction. Rows are closed before returning and payloads
// own their bytes. An error never returns a successful partial page.
func ListRestoredNonterminalPageTx(ctx context.Context, tx pgx.Tx, scope RestoredNonterminalScope, afterJobID *uuid.UUID, limit int) ([]RestoredNonterminalJob, error) {
	if ctx == nil || tx == nil || scope.JobKind == "" || !safeJobToken(scope.JobKind) || scope.ExtensionOwnerProfileID == "" ||
		len(scope.JobKind) > 191 || limit < 1 || limit > MaximumRestoredNonterminalPageSize || (afterJobID != nil && *afterJobID == uuid.Nil) {
		return nil, ErrInvalidJobDefinition
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	query := `SELECT job_id, incident_id, handler_payload_json
 FROM jobs WHERE extension_owner_profile_id = $1 AND job_kind = $2
 AND status IN ('queued', 'running', 'cancel_requested')`
	args := []any{scope.ExtensionOwnerProfileID, scope.JobKind, limit}
	if afterJobID != nil {
		query += ` AND job_id > $4::uuid`
		args = append(args, *afterJobID)
	}
	query += ` ORDER BY job_id ASC LIMIT $3`
	rows, err := tx.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("enumerate restored Common Jobs: %w", err)
	}
	defer rows.Close()
	page := make([]RestoredNonterminalJob, 0, limit)
	for rows.Next() {
		var job RestoredNonterminalJob
		if err := rows.Scan(&job.JobID, &job.IncidentID, &job.HandlerPayloadJSON); err != nil {
			return nil, fmt.Errorf("scan restored Common Job: %w", err)
		}
		job.HandlerPayloadJSON = bytes.Clone(job.HandlerPayloadJSON)
		page = append(page, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restored Common Jobs: %w", err)
	}
	return page, nil
}
