package imports

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// Staged committed-source streams (Core 01 REQ-01-641). Paths are relative to
// the port's preparation scope. Source blobs are opaque payloads and are not
// staged.
const (
	committedSourcePreparedRowsPath    = "import_committed_sources.rows"
	committedSourcePreparedBindingPath = "import_committed_source_records.rows"
	committedSourcePreparedAttribution = "attribution.rows"
	committedSourceBindingBatch        = 512
)

type committedSourceStaging interface {
	incidentportability.MemberReader
	incidentportability.MemberWriter
}

// stagedCommittedBlob is one source blob admitted as an opaque payload: its
// sealed bundle member descriptor.
type stagedCommittedBlob struct {
	Digest   string
	ByteSize int64
	Member   incidentportability.Member
}

// committedSourceView is the binding scope of one view-schema source: its
// first data row and the record types its public view admits.
type committedSourceView struct {
	firstRow int
	types    []string
	public   bool
}

var committedSourceColumns = []string{
	"committed_source_id", "incident_id", "content_sha256", "source_file_kind", "source_media_type", "original_filename",
	"parser_profile_id", "parser_version", "locator_kind", "locator", "source_rect_a1", "header_row_ref", "data_start_row_ref",
	"approved_mapping_json", "mapping_fingerprint", "target_kind", "target_view_schema_id", "extension_profile_id",
	"origin_import_session_id", "origin_import_unit_id", "created_by_user_id", "committed_at",
}

// committedSourceRowCodec stages the admitted source with its historical
// creator; COPY writes the importing actor.
type committedSourceRowCodec struct{ actor uuid.UUID }

func (committedSourceRowCodec) Encode(source CommittedSource) ([]byte, error) {
	return json.Marshal(source)
}
func (committedSourceRowCodec) Decode(line []byte) (CommittedSource, error) {
	var source CommittedSource
	if err := json.Unmarshal(line, &source); err != nil {
		return CommittedSource{}, incidentportability.ErrPreparedStream
	}
	source.CommittedAt = source.CommittedAt.UTC()
	return source, nil
}
func (committedSourceRowCodec) Columns() []string { return committedSourceColumns }
func (c committedSourceRowCodec) Values(source CommittedSource) ([]any, error) {
	return []any{source.CommittedSourceID, source.IncidentID, source.ContentSHA256, source.SourceFileKind, source.SourceMediaType,
		source.OriginalFilename, source.ParserProfileID, source.ParserVersion, source.LocatorKind, source.Locator, source.SourceRectA1,
		source.HeaderRowRef, source.DataStartRowRef, source.ApprovedMapping, source.MappingFingerprint, source.TargetKind,
		source.TargetViewSchemaID, source.ExtensionProfileID, source.OriginImportSessionID, source.OriginImportUnitID, c.actor,
		source.CommittedAt}, nil
}

var committedSourceBindingColumns = []string{"incident_id", "committed_source_id", "source_row_ref", "record_id"}

type committedSourceBindingCodec struct{ incidentID uuid.UUID }

func (committedSourceBindingCodec) Encode(record committedSourceRecord) ([]byte, error) {
	return json.Marshal(record)
}
func (committedSourceBindingCodec) Decode(line []byte) (committedSourceRecord, error) {
	var record committedSourceRecord
	if err := json.Unmarshal(line, &record); err != nil {
		return committedSourceRecord{}, incidentportability.ErrPreparedStream
	}
	return record, nil
}
func (committedSourceBindingCodec) Columns() []string { return committedSourceBindingColumns }
func (c committedSourceBindingCodec) Values(record committedSourceRecord) ([]any, error) {
	return []any{c.incidentID, record.CommittedSourceID, record.SourceRowRef, record.RecordID}, nil
}

// stage writes the admitted sources, record bindings and historical
// attribution, and resolves each view-schema source's binding scope.
func (p *PreparedCommittedSources) stage(ctx context.Context, store committedSourceStaging, actors sourceport.ActorCatalog) error {
	p.views = map[uuid.UUID]committedSourceView{}
	sources, err := incidentportability.StagePreparedRows(ctx, store, committedSourcePreparedRowsPath, committedSourceRowCodec{}, func(emit func(CommittedSource) error) error {
		return walkCommittedSources(ctx, p.members, func(source CommittedSource) error {
			if source.TargetKind == ImportTargetKindViewSchema {
				view := committedSourceView{firstRow: source.DataStartRowRef}
				if resource, ok := viewschema.LookupPublicResource(source.TargetViewSchemaID); ok {
					view.public, view.types = true, append([]string(nil), resource.SourceRecordTypes...)
				}
				p.views[source.CommittedSourceID] = view
			}
			return emit(source)
		})
	})
	if err != nil {
		return err
	}
	bindings, err := incidentportability.StagePreparedRows(ctx, store, committedSourcePreparedBindingPath, committedSourceBindingCodec{}, func(emit func(committedSourceRecord) error) error {
		return walkCommittedSourceRecords(ctx, p.members, emit)
	})
	if err != nil {
		return err
	}
	if sources.Rows != p.count || bindings.Rows != p.recordCount {
		return incidentportability.ErrPreparedStream
	}
	admitted := func(actorID string) bool { _, ok := actors.Lookup(actorID); return ok }
	attribution, err := incidentportability.StageAttributions(ctx, store, committedSourcePreparedAttribution, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		return incidentportability.ReadPreparedRows(ctx, store, sources, committedSourceRowCodec{}, func(source CommittedSource) error {
			return emit(incidentportability.AttributionRow{SourceTable: "import_committed_sources", SourceRowID: source.CommittedSourceID.String(), SourceColumn: "created_by_user_id", SourceActorID: source.CreatedByUserID.String()})
		})
	})
	if err != nil {
		return err
	}
	p.staged, p.sources, p.bindings, p.attribution = true, sources, bindings, attribution
	return nil
}

// insertStagedBlob streams one new blob's bytes from its admitted payload in
// fixed chunks with one COPY. Reaching EOF completes the seal check, the only
// hash of these bytes.
func insertStagedBlob(ctx context.Context, tx pgx.Tx, payloads incidentportability.PayloadMembers, blob stagedCommittedBlob) error {
	reader, err := payloads.OpenPayload(ctx, blob.Member)
	if err != nil {
		return err
	}
	chunks := &stagedBlobChunks{reader: reader, digest: blob.Digest, buffer: make([]byte, sourceReadChunkBytes)}
	_, copyErr := tx.CopyFrom(ctx, pgx.Identifier{"import_committed_source_chunks"}, []string{"content_sha256", "chunk_ordinal", "source_bytes"}, chunks)
	if err := errors.Join(copyErr, chunks.err, reader.Close()); err != nil {
		return err
	}
	if chunks.bytes != blob.ByteSize {
		return incidentportability.ErrPreparedStream
	}
	return nil
}

type stagedBlobChunks struct {
	reader  io.Reader
	digest  string
	buffer  []byte
	chunk   []byte
	ordinal int64
	bytes   int64
	done    bool
	err     error
}

func (c *stagedBlobChunks) Next() bool {
	if c.done || c.err != nil {
		return false
	}
	n, err := io.ReadFull(c.reader, c.buffer)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		c.done = true
	} else if err != nil {
		c.err = err
		return false
	}
	if n == 0 {
		return false
	}
	c.chunk = c.buffer[:n]
	c.bytes += int64(n)
	return true
}

func (c *stagedBlobChunks) Values() ([]any, error) {
	ordinal := c.ordinal
	c.ordinal++
	return []any{c.digest, ordinal, c.chunk}, nil
}

func (c *stagedBlobChunks) Err() error { return c.err }
