package networkflowsupport

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

type ResourceIntentAppender struct {
	appender collaboration.PublicationAppender
}

func NewResourceIntentAppender() ResourceIntentAppender {
	return ResourceIntentAppender{appender: collaborationsupport.NewPublicationAppender()}
}

func (a ResourceIntentAppender) AppendResourceIntentTx(ctx context.Context, tx pgx.Tx, source networkflow.ResourceIntent) error {
	return a.appender.AppendExtensionResourceChangedTx(ctx, tx, collaboration.ExtensionResourceChangeIntentInput{
		IntentKey: source.IntentKey, IncidentID: source.IncidentID,
		CanonicalPayload: source.CanonicalPayload, SourceIdentity: source.SourceIdentity,
		CreatedAt: source.CreatedAt,
	})
}
