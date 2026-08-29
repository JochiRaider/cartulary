package server

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// collaborationJobProgressTranslator is the application-owned translation
// boundary between the Jobs producer contract and Collaboration persistence.
type collaborationJobProgressTranslator struct {
	appender collaboration.JobProgressAppender
}

func newCollaborationJobProgressTranslator(appender collaboration.JobProgressAppender) collaborationJobProgressTranslator {
	return collaborationJobProgressTranslator{appender: appender}
}

func (a collaborationJobProgressTranslator) AppendProgressIntentTx(ctx context.Context, tx pgx.Tx, source jobs.ProgressIntent) error {
	return a.appender.AppendJobProgressTx(ctx, tx, collaboration.JobProgressIntentInput{
		IntentKey: source.IntentKey, IncidentID: source.IncidentID, CanonicalPayload: source.CanonicalPayload,
		SourceIdentity: source.SourceIdentity, CreatedAt: source.CreatedAt,
	})
}

// collaborationExtensionResourceChangeTranslator is the application-owned
// translation boundary between Network Flow and Collaboration persistence.
type collaborationExtensionResourceChangeTranslator struct {
	appender collaboration.ExtensionResourceChangedAppender
}

func newCollaborationExtensionResourceChangeTranslator(appender collaboration.ExtensionResourceChangedAppender) collaborationExtensionResourceChangeTranslator {
	return collaborationExtensionResourceChangeTranslator{appender: appender}
}

func (a collaborationExtensionResourceChangeTranslator) AppendResourceIntentTx(ctx context.Context, tx pgx.Tx, source networkflow.ResourceIntent) error {
	return a.appender.AppendExtensionResourceChangedTx(ctx, tx, collaboration.ExtensionResourceChangeIntentInput{
		IntentKey: source.IntentKey, IncidentID: source.IncidentID, CanonicalPayload: source.CanonicalPayload,
		SourceIdentity: source.SourceIdentity, CreatedAt: source.CreatedAt,
	})
}
