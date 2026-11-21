package storage

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

// projectionRowsResult reads the results of a page of projection rows sent as
// one pipelined batch, so a rebuild costs one round trip per page rather than
// per record, and reports the first row that fails.
func projectionRowsResult(results pgx.BatchResults, rows int, label string) error {
	for index := range rows {
		if _, err := results.Exec(); err != nil {
			_ = results.Close()
			return fmt.Errorf("%s %d of %d: %w", label, index+1, rows, err)
		}
	}
	if err := results.Close(); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}
