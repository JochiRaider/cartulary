package tasksdecisionassembly

import (
	"context"
	"errors"
	partyreferences "github.com/JochiRaider/cartulary/internal/modules/parties/references"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type referenceEnvelopeLocker interface {
	LoadEnvelopeForShareTx(context.Context, pgx.Tx, uuid.UUID) (enveloperead.Envelope, error)
	LoadEnvelopeForShareNowaitTx(context.Context, pgx.Tx, uuid.UUID) (enveloperead.Envelope, error)
	LoadEnvelopeForNoKeyUpdateTx(context.Context, pgx.Tx, uuid.UUID) (enveloperead.Envelope, error)
	LoadEnvelopeForNoKeyUpdateNowaitTx(context.Context, pgx.Tx, uuid.UUID) (enveloperead.Envelope, error)
}

type partySourceLocker interface {
	LockSourceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, bool) (bool, error)
}

type referenceTargetAdapter struct {
	envelopes referenceEnvelopeLocker
	parties   partySourceLocker
}

func NewReferenceTargetCapability() tasksdecisions.ReferenceTargetCapability {
	return referenceTargetAdapter{envelopes: records.NewStore(), parties: partyreferences.SourceLocker{}}
}

func (a referenceTargetAdapter) LockEnvelopeFactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, expectedType string, nowait bool) (tasksdecisions.EnvelopeFact, bool, error) {
	var envelope enveloperead.Envelope
	var err error
	switch {
	case expectedType != "party" && expectedType != "decision":
		return tasksdecisions.EnvelopeFact{}, false, errors.New("Task reference envelope kind is unsupported")
	case expectedType == "decision" && nowait:
		envelope, err = a.envelopes.LoadEnvelopeForNoKeyUpdateNowaitTx(ctx, tx, recordID)
	case expectedType == "decision":
		envelope, err = a.envelopes.LoadEnvelopeForNoKeyUpdateTx(ctx, tx, recordID)
	case nowait:
		envelope, err = a.envelopes.LoadEnvelopeForShareNowaitTx(ctx, tx, recordID)
	default:
		envelope, err = a.envelopes.LoadEnvelopeForShareTx(ctx, tx, recordID)
	}
	if errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		return tasksdecisions.EnvelopeFact{}, false, nil
	}
	if err != nil {
		return tasksdecisions.EnvelopeFact{}, false, err
	}
	return tasksdecisions.EnvelopeFact{IncidentID: envelope.IncidentID, RecordType: envelope.RecordType, Deleted: envelope.DeletedAt != nil}, true, nil
}

func (a referenceTargetAdapter) LockPartySourceTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID, nowait bool) (bool, error) {
	found, err := a.parties.LockSourceTx(ctx, tx, incidentID, recordID, nowait)
	return found && err == nil, err
}
