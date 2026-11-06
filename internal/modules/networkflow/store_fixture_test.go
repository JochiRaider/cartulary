package networkflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"sort"
	"time"
)

type createTableFixtureParams struct {
	CommittedSourceID         uuid.UUID
	IncidentID                uuid.UUID
	ActorUserID               uuid.UUID
	ImportSessionID           uuid.UUID
	ImportUnitID              uuid.UUID
	ClientTxnID               string
	RequestID                 string
	SourceContentSHA256       string
	OriginalFilename          string
	SourceFilenameDigest      string
	SourceFilenameDigestKeyID string
	MappingFingerprint        string
	SourceProfileID           string
	ParserProfileID           string
	DisplayNameOverride       *string
	Rows                      []normalizedRow
	RowCountRejected          int64
	Diagnostics               []rejectedRowDiagnostic
	DiagnosticsTruncated      bool
	Now                       time.Time
}

// Fixture-only convenience operations wrap the actual transaction-bound owner.
func (s *store) CreateTable(ctx context.Context, fixture createTableFixtureParams) (tableRecord, error) {
	ordered := append([]normalizedRow(nil), fixture.Rows...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].SourceRowNumber < ordered[j].SourceRowNumber })
	digest := sha256.New()
	for _, row := range ordered {
		digest.Write([]byte(row.SourceRowDigestSHA256 + row.NormalizedRowDigestSHA256))
	}
	for _, diagnostic := range fixture.Diagnostics {
		raw, err := canonicalJSON(diagnostic)
		if err != nil {
			return tableRecord{}, err
		}
		digest.Write(raw)
	}
	params := createTableParams{
		SourceRawCellScalarLimit: s.limits.MaxRawCellScalarLength,
		CommittedSourceID:        imports.CommittedSourceID(fixture.ImportUnitID), SourceInterpretationSHA256: hex.EncodeToString(digest.Sum(nil)),
		IncidentID:                fixture.IncidentID,
		ActorUserID:               fixture.ActorUserID,
		ImportSessionID:           fixture.ImportSessionID,
		ImportUnitID:              fixture.ImportUnitID,
		ClientTxnID:               fixture.ClientTxnID,
		RequestID:                 fixture.RequestID,
		SourceContentSHA256:       fixture.SourceContentSHA256,
		OriginalFilename:          fixture.OriginalFilename,
		SourceFilenameDigest:      fixture.SourceFilenameDigest,
		SourceFilenameDigestKeyID: fixture.SourceFilenameDigestKeyID,
		MappingFingerprint:        fixture.MappingFingerprint,
		SourceProfileID:           fixture.SourceProfileID,
		ParserProfileID:           fixture.ParserProfileID,
		DisplayNameOverride:       fixture.DisplayNameOverride,
		RowCountRejected:          fixture.RowCountRejected,
		Diagnostics:               fixture.Diagnostics,
		DiagnosticsTruncated:      fixture.DiagnosticsTruncated,
		Now:                       fixture.Now,
		RowCountAccepted:          int64(len(ordered)),
		VisitRows: func(visit func(normalizedRow) error) error {
			for _, row := range ordered {
				if err := visit(row); err != nil {
					return err
				}
			}
			return nil
		},
	}

	if fixture.CommittedSourceID != uuid.Nil {
		params.CommittedSourceID = fixture.CommittedSourceID
	}
	var table tableRecord
	err := withinTransaction(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.CreateTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return tableRecord{}, err
	}
	return table, nil
}

func (s *store) RenameTable(ctx context.Context, params renameTableParams) (tableRecord, error) {
	var table tableRecord
	err := withinTransaction(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.renameTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return tableRecord{}, err
	}
	return table, nil
}

func (s *store) SoftDeleteTable(ctx context.Context, params softDeleteTableParams) (tableRecord, error) {
	var table tableRecord
	err := withinTransaction(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.softDeleteTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return tableRecord{}, err
	}
	return table, nil
}

func (s *store) RetainedCounts(ctx context.Context, incidentID uuid.UUID) (retainedCounts, error) {
	return retainedCountsTx(ctx, s.pool, incidentID)
}
