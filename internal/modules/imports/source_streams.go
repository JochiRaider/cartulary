package imports

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const sourceStreamRefPrefix = "impsrc_"

type ImportSourceCapability struct {
	SourceStreamRef     string
	ImportSessionID     uuid.UUID
	ImportUnitID        uuid.UUID
	SourceContentSHA256 string
	SourceMediaType     string
	SourceByteSize      int64
}

type ImportSourceStream struct {
	ImportSourceCapability
	Reader io.ReadCloser
}

func newImportSourceStreamRef() string {
	return sourceStreamRefPrefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func (s *Store) OpenSourceStream(ctx context.Context, sourceStreamRef string) (ImportSourceStream, error) {
	capability, sourceBytes, err := s.loadSourceStream(ctx, s.pool, sourceStreamRef)
	if err != nil {
		return ImportSourceStream{}, err
	}
	return ImportSourceStream{
		ImportSourceCapability: capability,
		Reader:                 io.NopCloser(bytes.NewReader(sourceBytes)),
	}, nil
}

func (s *Store) SourceCapabilityForUnit(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (ImportSourceCapability, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ImportSourceCapability{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	capability, err := s.sourceCapabilityForUnitTx(ctx, tx, sessionID, unitID)
	if err != nil {
		return ImportSourceCapability{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ImportSourceCapability{}, err
	}
	return capability, nil
}

func (s *Store) sourceCapabilityForUnitTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID) (ImportSourceCapability, error) {
	var sourceStreamRef *string
	err := tx.QueryRow(ctx, `
SELECT source_stream_ref
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID).Scan(&sourceStreamRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ImportSourceCapability{}, ErrNotFound
		}
		return ImportSourceCapability{}, err
	}
	if sourceStreamRef == nil || *sourceStreamRef == "" {
		return ImportSourceCapability{}, importApplyBlockedError("source_changed")
	}
	capability, _, err := s.loadSourceStream(ctx, tx, *sourceStreamRef)
	return capability, err
}

// ValidateExtensionApplyPreconditionsTx is the Import owner's physical
// application-composition adapter for a shared final transaction. The returned
// capability is a fresh in-transaction read of the admitted unit; callers
// cannot query Import storage directly.
func (s *Store) ValidateExtensionApplyPreconditionsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	sessionID uuid.UUID,
	unitID uuid.UUID,
	expectedSourceStreamRef string,
	expectedSourceContentSHA256 string,
) error {
	if s == nil || tx == nil {
		return fmt.Errorf("import extension apply transaction unavailable")
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return err
	}
	capability, err := s.sourceCapabilityForUnitTx(ctx, tx, sessionID, unitID)
	if err != nil {
		return err
	}
	if capability.ImportSessionID != sessionID ||
		capability.ImportUnitID != unitID ||
		capability.SourceStreamRef != expectedSourceStreamRef ||
		capability.SourceContentSHA256 != expectedSourceContentSHA256 {
		return importApplyBlockedError("source_changed")
	}
	return nil
}

type sourceStreamQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (s *Store) loadSourceStream(ctx context.Context, querier sourceStreamQuerier, sourceStreamRef string) (ImportSourceCapability, []byte, error) {
	var capability ImportSourceCapability
	var sourceBytes []byte
	err := querier.QueryRow(ctx, `
SELECT source_stream_ref, import_session_id, import_unit_id, source_content_sha256,
       source_media_type, source_byte_size, source_bytes
  FROM import_source_streams
 WHERE source_stream_ref = $1
`, sourceStreamRef).Scan(
		&capability.SourceStreamRef,
		&capability.ImportSessionID,
		&capability.ImportUnitID,
		&capability.SourceContentSHA256,
		&capability.SourceMediaType,
		&capability.SourceByteSize,
		&sourceBytes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ImportSourceCapability{}, nil, ErrNotFound
		}
		return ImportSourceCapability{}, nil, err
	}
	sum := sha256.Sum256(sourceBytes)
	if hex.EncodeToString(sum[:]) != capability.SourceContentSHA256 {
		return ImportSourceCapability{}, nil, fmt.Errorf("import source stream digest mismatch")
	}
	return capability, sourceBytes, nil
}
