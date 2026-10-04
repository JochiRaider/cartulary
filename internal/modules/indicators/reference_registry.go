package indicators

import (
	"context"
	"github.com/jackc/pgx/v5"
)

// ReferencedIndicatorRegistryEntriesTx includes deleted retained records and
// unresolved observations: neither loses its exact historical type token.
func ReferencedIndicatorRegistryEntriesTx(ctx context.Context, tx pgx.Tx) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT indicator_type COLLATE "C" FROM indicators UNION SELECT parsed_indicator_type COLLATE "C" FROM indicator_observations WHERE parsed_indicator_type IS NOT NULL ORDER BY 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		result = append(result, token)
	}
	return result, rows.Err()
}
