package server

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// collaborationIntentAdapters is the application-owned translation boundary
// between source-owned producer contracts and Collaboration persistence.
type collaborationIntentAdapters struct {
	appender collaboration.IntentAppender
}

func (a collaborationIntentAdapters) AppendProgressIntentTx(ctx context.Context, tx pgx.Tx, source jobs.ProgressIntent) error {
	intent, err := collaboration.NewEventIntent(
		source.IntentKey,
		source.IncidentID,
		collaboration.EventFamilyJobProgress,
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

func (a collaborationIntentAdapters) AppendResourceIntentTx(ctx context.Context, tx pgx.Tx, source networkflow.ResourceIntent) error {
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
