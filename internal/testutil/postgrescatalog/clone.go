package postgrescatalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CloneDatabase publishes a WAL-backed copy after PostgreSQL's synchronous
// CREATE commit. A separate checkpoint would flush unrelated cluster data and
// is unnecessary for WAL durability. Unlike FILE_COPY's file-sync traversal,
// this uses PostgreSQL's ordinary interrupt-aware block-copy path.
// created reports successful CREATE even if lock release fails. The caller
// owns subsequent provisioning and cleanup of that proven-created target.
// Failed CREATE never grants ownership of an existing database.
func CloneDatabase(ctx context.Context, adminDSN, target, template string, admissionLimit time.Duration) (created bool, err error) {
	if !cloneIdentifier(target) || !cloneIdentifier(template) || target == template {
		return false, errors.New("postgres template clone requires distinct safe database identities")
	}
	err = WithMutation(ctx, adminDSN, target, admissionLimit, func(admin *sql.Conn) error {
		statement := fmt.Sprintf(`CREATE DATABASE "%s" TEMPLATE "%s" STRATEGY WAL_LOG`, target, template)
		if _, err := admin.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("create postgres template clone: %w", err)
		}
		created = true
		return nil
	})
	return created, err
}

func cloneIdentifier(value string) bool {
	if value == "" || len(value) > 63 {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '_' {
			continue
		}
		return false
	}
	return true
}
