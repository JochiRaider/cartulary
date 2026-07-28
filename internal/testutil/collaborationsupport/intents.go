package collaborationsupport

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// IntentAdapters supplies the same narrow source-to-Collaboration translation
// used by application composition to service-backed tests.
type IntentAdapters struct {
	appender collaboration.IntentAppender
}

func NewIntentAdapters() IntentAdapters {
	return IntentAdapters{appender: collaboration.NewIntentAppender()}
}

func NewJobTransactions() *jobs.TransactionService {
	service, err := jobs.NewTransactionService(NewIntentAdapters())
	if err != nil {
		panic(err)
	}
	return service
}

func (a IntentAdapters) AppendProgressIntentTx(ctx context.Context, tx pgx.Tx, source jobs.ProgressIntent) error {
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
