package authn

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ReadRouteIdempotencyPage includes all retained receipts, without a job-retention
// predicate. Auth owns storage; callers own the selected route payload contracts.
func ReadRouteIdempotencyPage(ctx context.Context, reader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, routes []string, after uuid.UUID) ([]RouteIdempotencyRecord, uuid.UUID, error) {
	rows, err := reader.Query(ctx, `SELECT id, route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json FROM route_idempotency WHERE route_key = ANY($1) AND id > $2 ORDER BY id LIMIT 128`, routes, after)
	if err != nil {
		return nil, after, err
	}
	defer rows.Close()
	records := make([]RouteIdempotencyRecord, 0, 128)
	for rows.Next() {
		var record RouteIdempotencyRecord
		if err := rows.Scan(&after, &record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.ActorUserID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON); err != nil {
			return nil, after, err
		}
		records = append(records, record)
	}
	return records, after, rows.Err()
}
