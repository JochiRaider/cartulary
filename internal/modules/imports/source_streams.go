package imports

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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
	OriginalFilename string
	Reader           io.ReadCloser
}

// ExtensionSourcePort is the exact Imports-owned capability analytical owners
// receive. It exposes source bytes and the in-transaction admission check
// required by apply without exposing Imports persistence operations.
type ExtensionSourcePort interface {
	OpenSourceStream(context.Context, string) (ImportSourceStream, error)
	OpenSourceStreamTx(context.Context, pgx.Tx, string) (ImportSourceStream, error)
	ValidateExtensionApplyPreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string, string) error
}

type extensionSourcePort struct {
	db             postgres.DB
	incidentAccess *admission.Checker
}

func NewExtensionSourcePort(db postgres.DB) (ExtensionSourcePort, error) {
	if nilInterface(db) {
		return nil, errors.New("imports extension source port requires PostgreSQL")
	}
	return &extensionSourcePort{db: db, incidentAccess: admission.NewChecker(db)}, nil
}

func newImportSourceStreamRef() string {
	return sourceStreamRefPrefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func (p *extensionSourcePort) OpenSourceStream(ctx context.Context, sourceStreamRef string) (ImportSourceStream, error) {
	if p == nil || p.db == nil {
		return ImportSourceStream{}, errors.New("imports extension source port unavailable")
	}
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return ImportSourceStream{}, err
	}
	stream, err := loadSourceMetadata(ctx, tx, sourceStreamRef, false)
	if err != nil {
		_ = closeSourceSnapshot(tx)
		return ImportSourceStream{}, err
	}
	stream.Reader = newSourceReader(ctx, tx, stream.ImportSourceCapability, true)
	return stream, nil
}

// OpenSourceStreamTx borrows the caller's transaction and protects the source
// until that transaction ends. Reader cleanup never ends a borrowed transaction.
func (p *extensionSourcePort) OpenSourceStreamTx(ctx context.Context, tx pgx.Tx, sourceStreamRef string) (ImportSourceStream, error) {
	if p == nil || nilInterface(tx) {
		return ImportSourceStream{}, errors.New("imports source transaction unavailable")
	}
	stream, err := loadSourceMetadata(ctx, tx, sourceStreamRef, true)
	if err != nil {
		return ImportSourceStream{}, err
	}
	stream.Reader = newSourceReader(ctx, tx, stream.ImportSourceCapability, false)
	return stream, nil
}

func (s *store) sourceCapabilityForUnit(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (ImportSourceCapability, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ImportSourceCapability{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	capability, err := sourceCapabilityForUnitTx(ctx, tx, sessionID, unitID)
	if err != nil {
		return ImportSourceCapability{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ImportSourceCapability{}, err
	}
	return capability, nil
}

func sourceCapabilityForUnitTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID) (ImportSourceCapability, error) {
	return loadUnitSourceCapabilityTx(ctx, tx, sessionID, unitID, false)
}

// protectMappingSourceTx validates the metadata returned after acquiring source
// protection. The caller already holds the session and unit locks; this source
// lock stays with that transaction through mapping and idempotency publication.
func protectMappingSourceTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID, expected ImportSourceCapability) error {
	actual, err := loadUnitSourceCapabilityTx(ctx, tx, sessionID, unitID, true)
	if errors.Is(err, errNotFound) {
		return importApplyBlockedError("source_changed")
	}
	if err != nil {
		return err
	}
	if actual != expected {
		return importApplyBlockedError("source_changed")
	}
	return nil
}

func loadUnitSourceCapabilityTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID, protect bool) (ImportSourceCapability, error) {
	var sourceStreamRef *string
	err := tx.QueryRow(ctx, `
SELECT source_stream_ref
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID).Scan(&sourceStreamRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ImportSourceCapability{}, errNotFound
		}
		return ImportSourceCapability{}, err
	}
	if sourceStreamRef == nil || *sourceStreamRef == "" {
		return ImportSourceCapability{}, importApplyBlockedError("source_changed")
	}
	stream, err := loadSourceMetadata(ctx, tx, *sourceStreamRef, protect)
	if err != nil {
		return ImportSourceCapability{}, err
	}
	if stream.ImportSessionID != sessionID || stream.ImportUnitID != unitID {
		return ImportSourceCapability{}, importApplyBlockedError("source_changed")
	}
	return stream.ImportSourceCapability, nil
}

func (p *extensionSourcePort) ValidateExtensionApplyPreconditionsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	sessionID uuid.UUID,
	unitID uuid.UUID,
	expectedSourceStreamRef string,
	expectedSourceContentSHA256 string,
) error {
	if p == nil {
		return fmt.Errorf("import extension apply transaction unavailable")
	}
	return validateExtensionApplyPreconditionsTx(
		ctx, p.incidentAccess, tx, incidentID, sessionID, unitID,
		expectedSourceStreamRef, expectedSourceContentSHA256,
	)
}

func validateExtensionApplyPreconditionsTx(
	ctx context.Context,
	incidentAccess *admission.Checker,
	tx pgx.Tx,
	incidentID uuid.UUID,
	sessionID uuid.UUID,
	unitID uuid.UUID,
	expectedSourceStreamRef string,
	expectedSourceContentSHA256 string,
) error {
	if incidentAccess == nil || tx == nil {
		return fmt.Errorf("import extension apply transaction unavailable")
	}
	if err := incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return err
	}
	capability, err := sourceCapabilityForUnitTx(ctx, tx, sessionID, unitID)
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

func loadSourceMetadata(ctx context.Context, querier sourceStreamQuerier, sourceStreamRef string, protect bool) (ImportSourceStream, error) {
	var stream ImportSourceStream
	query := `
SELECT streams.source_stream_ref, streams.import_session_id, streams.import_unit_id,
       streams.source_content_sha256, streams.source_media_type, streams.source_byte_size,
       sessions.original_filename
  FROM import_source_streams streams
  JOIN import_sessions sessions
    ON sessions.import_session_id = streams.import_session_id
 WHERE streams.source_stream_ref = $1
`
	if protect {
		query += " FOR SHARE OF streams"
	}
	err := querier.QueryRow(ctx, query, sourceStreamRef).Scan(
		&stream.SourceStreamRef,
		&stream.ImportSessionID,
		&stream.ImportUnitID,
		&stream.SourceContentSHA256,
		&stream.SourceMediaType,
		&stream.SourceByteSize,
		&stream.OriginalFilename,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ImportSourceStream{}, errNotFound
		}
		return ImportSourceStream{}, err
	}
	return stream, nil
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
