package facts

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

// TimelineTagLabelsAreCanonicalTx checks both active and tombstoned bindings
// in the caller's snapshot. A logical restore bypasses normal writer admission.
func TimelineTagLabelsAreCanonicalTx(ctx context.Context, tx pgx.Tx) (bool, error) {
	rows, err := tx.Query(ctx, `
SELECT tag_name, normalized_tag_name FROM public.links_timeline_tag_labels_v1`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var label, key string
		if err := rows.Scan(&label, &key); err != nil {
			return false, err
		}
		if !fieldnorm.IsCanonicalTagLabel(label, key) {
			return false, nil
		}
	}
	return true, rows.Err()
}
