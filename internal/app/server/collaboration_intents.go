package server

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// collaborationIntentTranslator is the application-owned translation boundary
// between source-owned producer contracts and Collaboration persistence.
type collaborationIntentTranslator struct {
	appender collaboration.PublicationAppender
}

func newCollaborationIntentTranslator(appender collaboration.PublicationAppender) collaborationIntentTranslator {
	return collaborationIntentTranslator{appender: appender}
}

func (a collaborationIntentTranslator) AppendProgressIntentTx(ctx context.Context, tx pgx.Tx, source jobs.ProgressIntent) error {
	return a.appender.AppendJobProgressTx(ctx, tx, collaboration.JobProgressIntentInput{
		IntentKey: source.IntentKey, IncidentID: source.IncidentID, CanonicalPayload: source.CanonicalPayload,
		SourceIdentity: source.SourceIdentity, CreatedAt: source.CreatedAt,
	})
}

func (a collaborationIntentTranslator) AppendResourceIntentTx(ctx context.Context, tx pgx.Tx, source networkflow.ResourceIntent) error {
	return a.appender.AppendExtensionResourceChangedTx(ctx, tx, collaboration.ExtensionResourceChangeIntentInput{
		IntentKey: source.IntentKey, IncidentID: source.IncidentID, CanonicalPayload: source.CanonicalPayload,
		SourceIdentity: source.SourceIdentity, CreatedAt: source.CreatedAt,
	})
}
