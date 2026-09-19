package postgresresult

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/jackc/pgx/v5"
)

// ExactLeaseKey identifies a stored protection independently of its generated ID.
type ExactLeaseKey struct {
	ProjectionResultID   string
	LeaseOwnerID         string
	LeaseOwnerResourceID string
	LeasePurpose         string
}

// LookupUnexpiredLease observes the caller's exact scope without renewal or locks.
func (reader *Reader) LookupUnexpiredLease(ctx context.Context, key ExactLeaseKey, observedAt time.Time) (string, error) {
	if ctx == nil || reader == nil || reader.db == nil {
		return "", graphprojection.ErrResultV2Invalid
	}

	var leaseID string
	err := reader.db.QueryRow(ctx, `SELECT lease_id FROM graph_projection_result_leases
 WHERE projection_result_id = $1 AND lease_owner_id = $2
 AND lease_owner_resource_id = $3 AND lease_purpose = $4 AND leased_until > $5`,
		key.ProjectionResultID, key.LeaseOwnerID, key.LeaseOwnerResourceID, key.LeasePurpose, observedAt.UTC()).Scan(&leaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", graphprojection.ErrResultV2LeaseNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup Graph Projection lease: %w", err)
	}
	return leaseID, nil
}

// LeaseReleaseScope cannot omit any scope dimension or imply a wildcard.
type LeaseReleaseScope struct {
	SourceOwnerID        string
	LeaseOwnerID         string
	LeaseOwnerResourceID string
	LeasePurpose         string
}

type LeaseReleaser struct{ tx pgx.Tx }

func NewLeaseReleaser(tx pgx.Tx) (*LeaseReleaser, error) {
	if tx == nil {
		return nil, fmt.Errorf("graph projection lease release transaction is required")
	}
	return &LeaseReleaser{tx: tx}, nil
}

// ReleaseScopedLeasesTx includes expired and live leases across all results in
// the exact source scope. It never commits or closes the caller's transaction.
func (releaser *LeaseReleaser) ReleaseScopedLeasesTx(ctx context.Context, scope LeaseReleaseScope) error {
	if ctx == nil || releaser == nil || releaser.tx == nil || scope.SourceOwnerID == "" ||
		!validLeaseScope(scope.LeaseOwnerID, scope.LeaseOwnerResourceID, scope.LeasePurpose) {
		return graphprojection.ErrResultV2Invalid
	}
	_, err := releaser.tx.Exec(ctx, `DELETE FROM graph_projection_result_leases lease
 USING graph_projection_results result
 WHERE lease.projection_result_id = result.projection_result_id
 AND result.source_owner_id = $1 AND lease.lease_owner_id = $2
 AND lease.lease_owner_resource_id = $3 AND lease.lease_purpose = $4`,
		scope.SourceOwnerID, scope.LeaseOwnerID, scope.LeaseOwnerResourceID, scope.LeasePurpose)
	if err != nil {
		return fmt.Errorf("release scoped Graph Projection leases: %w", err)
	}
	return nil
}

func validLeaseScope(owner, resource, purpose string) bool {
	return owner != "" && resource != "" && purpose != ""
}
