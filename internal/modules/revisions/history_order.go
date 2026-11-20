package revisions

import (
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
)

// ErrHistoryOrderMissing reports a committed change set without the history
// order that the commit-time allocator assigns. It is an invariant break, not
// a caller precondition, and never stands in for a zero or inferred order.
var ErrHistoryOrderMissing = errors.New("revisions: committed change set has no history order")

func committedHistoryOrder(value pgtype.Int8) (int64, error) {
	if !value.Valid || value.Int64 <= 0 {
		return 0, ErrHistoryOrderMissing
	}
	return value.Int64, nil
}
