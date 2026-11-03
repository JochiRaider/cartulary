package reporting

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func getPreviewPayloadTx(ctx context.Context, jobReader jobs.RetainedJobReader, tx pgx.Tx, jobID uuid.UUID) (compositionPreviewJobPayload, error) {
	a, input, err := materializeRetainedRenderTx(ctx, jobReader, tx, jobID)
	if err != nil {
		return compositionPreviewJobPayload{}, err
	}
	return previewPayloadFromPrepared(a, input)
}
