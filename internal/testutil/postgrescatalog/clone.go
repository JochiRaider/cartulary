package postgrescatalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CloneDatabase completes the physical template copy before returning it to
// its owner. FILE_COPY checkpoints during preparation instead of leaving bulk
// copy synchronization to an unrelated fixture's later DROP checkpoint.
// Callers retain template authorization, target ownership and failed-acquisition
// cleanup; this operation never replaces an existing database.
func CloneDatabase(ctx context.Context, adminDSN, target, template string, admissionLimit time.Duration) error {
	if !cloneIdentifier(target) || !cloneIdentifier(template) || target == template {
		return errors.New("postgres template clone requires distinct safe database identities")
	}
	return WithMutation(ctx, adminDSN, target, admissionLimit, func(admin *sql.DB) error {
		statement := fmt.Sprintf(`CREATE DATABASE "%s" TEMPLATE "%s" STRATEGY FILE_COPY`, target, template)
		if _, err := admin.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("create postgres template clone: %w", err)
		}
		return nil
	})
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
