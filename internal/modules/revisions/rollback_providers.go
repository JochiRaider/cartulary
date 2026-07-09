package revisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	timelinerollback "github.com/JochiRaider/cartulary/internal/modules/timeline/rollbackprovider"
)

type rollbackSourceOwnerProvider interface {
	SourceForRollbackValue(value map[string]any) (map[string]any, bool, error)
	UpdateSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, source map[string]any) (bool, error)
	TouchSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64) (bool, error)
}

var rollbackSourceOwnerProviders = map[string]rollbackSourceOwnerProvider{
	"timeline_event": timelinerollback.NewTimelineProvider(),
}

func rollbackProviderForRecordType(recordType string) (rollbackSourceOwnerProvider, bool) {
	provider, ok := rollbackSourceOwnerProviders[recordType]
	return provider, ok
}
