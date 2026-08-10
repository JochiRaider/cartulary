package artifacts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type artifactRecordsCapabilitySpy struct {
	wantTx       pgx.Tx
	wantRecordID uuid.UUID
	envelope     records.Envelope
	loadErr      error
	called       bool
}

func (*artifactRecordsCapabilitySpy) InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error) {
	return uuid.Nil, errors.New("unexpected InsertTx")
}

func (*artifactRecordsCapabilitySpy) AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	return 0, errors.New("unexpected AdvanceVersionTx")
}

func (spy *artifactRecordsCapabilitySpy) LoadEnvelopeTx(
	_ context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	lock bool,
) (records.Envelope, error) {
	spy.called = true
	if tx != spy.wantTx {
		return records.Envelope{}, errors.New("artifact supplied a different transaction")
	}
	if recordID != spy.wantRecordID {
		return records.Envelope{}, errors.New("artifact supplied a different record ID")
	}
	if !lock {
		return records.Envelope{}, errors.New("artifact current-envelope lookup did not request a lock")
	}
	return spy.envelope, spy.loadErr
}

type artifactRecordsCapabilityTx struct{ pgx.Tx }

func TestArtifactCurrentEnvelopeUsesRecordsCallerTransactionCapability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	recordID := uuid.New()
	incidentID := uuid.New()
	tx := pgx.Tx(&artifactRecordsCapabilityTx{})

	t.Run("locked envelope metadata", func(t *testing.T) {
		spy := &artifactRecordsCapabilitySpy{
			wantTx: tx, wantRecordID: recordID,
			envelope: records.Envelope{
				RecordID: recordID, IncidentID: incidentID,
				RecordType: "artifact", RowVersion: 7,
			},
		}
		facade := &MutationFacade{source: artifactSourceKernel{records: spy}}
		meta, err := facade.loadArtifactRecordMetaForUpdateTx(ctx, tx, recordID)
		if err != nil {
			t.Fatalf("load locked artifact envelope: %v", err)
		}
		if !spy.called || meta.IncidentID != incidentID || meta.RecordType != "artifact" || meta.RowVersion != 7 {
			t.Fatalf("locked artifact metadata = %#v, called=%v", meta, spy.called)
		}
	})

	t.Run("missing envelope remains concealed", func(t *testing.T) {
		spy := &artifactRecordsCapabilitySpy{
			wantTx: tx, wantRecordID: recordID, loadErr: records.ErrEnvelopeNotFound,
		}
		facade := &MutationFacade{source: artifactSourceKernel{records: spy}}
		if _, err := facade.loadArtifactRecordMetaForUpdateTx(ctx, tx, recordID); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("missing artifact envelope error = %v, want pgx.ErrNoRows", err)
		}
	})

	t.Run("deleted envelope requires restore", func(t *testing.T) {
		deletedAt := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
		spy := &artifactRecordsCapabilitySpy{
			wantTx: tx, wantRecordID: recordID,
			envelope: records.Envelope{
				RecordID: recordID, IncidentID: incidentID,
				RecordType: "artifact", RowVersion: 7, DeletedAt: &deletedAt,
			},
		}
		facade := &MutationFacade{source: artifactSourceKernel{records: spy}}
		if _, err := facade.loadArtifactRecordMetaForUpdateTx(ctx, tx, recordID); !errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
			t.Fatalf("deleted artifact envelope error = %v, want ErrRecordDeletedUseRestore", err)
		}
	})
}
