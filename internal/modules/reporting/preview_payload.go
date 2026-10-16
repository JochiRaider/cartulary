package reporting

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func getPreviewPayloadTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (compositionPreviewJobPayload, error) {
	a, input, err := materializeRetainedRenderTx(ctx, tx, jobID)
	if err != nil {
		return compositionPreviewJobPayload{}, err
	}
	return previewPayloadFromPrepared(a, input)
}
