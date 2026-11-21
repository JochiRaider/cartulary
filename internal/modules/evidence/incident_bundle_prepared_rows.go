package evidence

import (
	"context"
	"errors"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// Staged Evidence streams (Core 01 REQ-01-641). Preparation admits every
// source row; the final transaction reads these streams and opens no bundle
// member. Paths are relative to the port's preparation scope.
const (
	preparedEvidenceRecordsPath = "evidence_records.rows"
	preparedEvidenceCustodyPath = "evidence_custody_events.rows"
	preparedEvidenceBlobsPath   = "object_blobs.rows"
	preparedEvidenceAttrPath    = "attribution.rows"
)

// evidenceRecordCodec and evidenceBlobCodec restore the UTC clocks the
// admitted rows carry, so staged rows compare exactly with parsed rows. The
// staged streams are read back only; none is written by COPY.
type evidenceRecordCodec struct{}

func (evidenceRecordCodec) Encode(row sourcevalue.Row) ([]byte, error) {
	return incidentportability.EncodeJSONRow(row)
}

func (evidenceRecordCodec) Decode(line []byte) (sourcevalue.Row, error) {
	var row sourcevalue.Row
	err := incidentportability.DecodeJSONRow(line, &row)
	row.UTCClocks()
	return row, err
}

func (evidenceRecordCodec) Columns() []string { return nil }

func (evidenceRecordCodec) Values(sourcevalue.Row) ([]any, error) {
	return nil, incidentportability.ErrPreparedStream
}

type evidenceBlobCodec struct{}

func (evidenceBlobCodec) Encode(row evidenceblob.Row) ([]byte, error) {
	return incidentportability.EncodeJSONRow(row)
}

func (evidenceBlobCodec) Decode(line []byte) (evidenceblob.Row, error) {
	var row evidenceblob.Row
	err := incidentportability.DecodeJSONRow(line, &row)
	row.UTC()
	return row, err
}

func (evidenceBlobCodec) Columns() []string { return nil }

func (evidenceBlobCodec) Values(evidenceblob.Row) ([]any, error) {
	return nil, incidentportability.ErrPreparedStream
}

type evidenceCustodyCodec struct{}

func (evidenceCustodyCodec) Encode(row portableCustodyRow) ([]byte, error) {
	return incidentportability.EncodeJSONRow(row)
}

func (evidenceCustodyCodec) Decode(line []byte) (portableCustodyRow, error) {
	var row portableCustodyRow
	return row, incidentportability.DecodeJSONRow(line, &row)
}

func (evidenceCustodyCodec) Columns() []string { return nil }

func (evidenceCustodyCodec) Values(portableCustodyRow) ([]any, error) {
	return nil, incidentportability.ErrPreparedStream
}

// stage writes the admitted rows and their historical attribution, then
// retires the bundle from the prepared value.
func (p *preparedEvidenceImport) stage(ctx context.Context, input sourceport.ImportContext, d sourceport.Descriptor) error {
	store := input.Preparation
	if store == nil {
		return errors.New("evidence: incident bundle preparation staging is required")
	}
	var err error
	if p.recordRows, err = incidentportability.StagePreparedRows(ctx, store, preparedEvidenceRecordsPath, evidenceRecordCodec{}, func(emit func(sourcevalue.Row) error) error {
		return p.bundleTyped(ctx, d, emit)
	}); err != nil {
		return err
	}
	if p.custodyRows, err = incidentportability.StagePreparedRows(ctx, store, preparedEvidenceCustodyPath, evidenceCustodyCodec{}, func(emit func(portableCustodyRow) error) error {
		return p.bundleCustody(ctx, d, emit)
	}); err != nil {
		return err
	}
	if p.blobRows, err = incidentportability.StagePreparedRows(ctx, store, preparedEvidenceBlobsPath, evidenceBlobCodec{}, func(emit func(evidenceblob.Row) error) error {
		return p.bundleBlobs(ctx, d, emit)
	}); err != nil {
		return err
	}
	admitted := func(actorID string) bool { _, ok := input.Actors.Lookup(actorID); return ok }
	p.attribution, err = incidentportability.StageAttributions(ctx, store, preparedEvidenceAttrPath, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		if err := incidentportability.ReadPreparedRows(ctx, store, p.recordRows, evidenceRecordCodec{}, func(row sourcevalue.Row) error {
			if row.OwnerUserID == nil {
				return nil
			}
			return emit(incidentportability.AttributionRow{SourceTable: "evidence", SourceRowID: row.RecordID.String(), SourceColumn: "owner_user_id", SourceActorID: row.OwnerUserID.String()})
		}); err != nil {
			return err
		}
		if err := incidentportability.ReadPreparedRows(ctx, store, p.blobRows, evidenceBlobCodec{}, func(row evidenceblob.Row) error {
			return emit(incidentportability.AttributionRow{SourceTable: "object_blobs", SourceRowID: row.ObjectBlobID.String(), SourceColumn: "created_by_user_id", SourceActorID: row.CreatedByUserID.String()})
		}); err != nil {
			return err
		}
		return incidentportability.ReadPreparedRows(ctx, store, p.custodyRows, evidenceCustodyCodec{}, func(row portableCustodyRow) error {
			if row.ActorUserID == nil {
				return nil
			}
			return emit(incidentportability.AttributionRow{SourceTable: "evidence_custody_events", SourceRowID: row.CustodyEventID.String(), SourceColumn: "actor_user_id", SourceActorID: row.ActorUserID.String()})
		})
	})
	if err != nil {
		return err
	}
	p.members, p.staged = nil, true
	return nil
}

// typed, custody and blobs visit the staged rows in source order.
func (p preparedEvidenceImport) typed(ctx context.Context, _ sourceport.Descriptor, visit func(sourcevalue.Row) error) error {
	if !p.staged || p.store == nil {
		return sourceport.ErrPreparedBinding
	}
	return incidentportability.ReadPreparedRows(ctx, p.store, p.recordRows, evidenceRecordCodec{}, visit)
}

func (p preparedEvidenceImport) custody(ctx context.Context, _ sourceport.Descriptor, visit func(portableCustodyRow) error) error {
	if !p.staged || p.store == nil {
		return sourceport.ErrPreparedBinding
	}
	return incidentportability.ReadPreparedRows(ctx, p.store, p.custodyRows, evidenceCustodyCodec{}, visit)
}

func (p preparedEvidenceImport) blobs(ctx context.Context, _ sourceport.Descriptor, visit func(evidenceblob.Row) error) error {
	if !p.staged || p.store == nil {
		return sourceport.ErrPreparedBinding
	}
	return incidentportability.ReadPreparedRows(ctx, p.store, p.blobRows, evidenceBlobCodec{}, visit)
}
