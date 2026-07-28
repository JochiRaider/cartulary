package networkflowsupport

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
)

type ResourceIntentAppender struct {
	appender collaboration.IntentAppender
}

func NewResourceIntentAppender() ResourceIntentAppender {
	return ResourceIntentAppender{appender: collaboration.NewIntentAppender()}
}

func (a ResourceIntentAppender) AppendResourceIntentTx(ctx context.Context, tx pgx.Tx, source networkflow.ResourceIntent) error {
	intent, err := collaboration.NewEventIntent(
		source.IntentKey,
		source.IncidentID,
		collaboration.EventFamilyExtensionResourceChange,
		source.CanonicalPayload,
		source.SourceIdentity,
		0,
		source.CreatedAt,
	)
	if err != nil {
		return err
	}
	return a.appender.AppendIntentTx(ctx, tx, intent)
}
