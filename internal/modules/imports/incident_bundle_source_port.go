package imports

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

const committedSourcesPath = "data/committed_sources.ndjson"
const committedSourceRecordsPath = "data/committed_source_records.ndjson"
const committedSourceBlobPrefix = "sources/sha256/"

func committedSourceDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "committed_sources", ContractMajor: sourceport.ContractMajor, OwnerID: "module.imports",
		OwnerRelationIDs: []string{"committed-import-sources"}, Dependencies: []string{"incident", "records"},
		Paths: []sourceport.Path{
			{LogicalPath: committedSourcesPath, ContentRole: "source_rows", SchemaID: "cartulary.imports.committed_source.v1", Versions: []int{7}, StableIdentity: []string{"committed_source_id"}, StableIdentityInvariantID: "committed_sources.source_identity_admitted"},
			{LogicalPath: committedSourceRecordsPath, ContentRole: "source_rows", SchemaID: "cartulary.imports.committed_source_record.v1", Versions: []int{7}, StableIdentity: []string{"committed_source_id", "source_row_ref"}, StableIdentityInvariantID: "committed_sources.source_identity_admitted"},
		},
		InvariantIDs: []string{"committed_sources.source_identity_admitted", "committed_sources.metadata_exact", "committed_sources.mapping_exact", "committed_sources.blob_exact", "committed_sources.attribution_exact", "committed_sources.record_binding_exact"},
	}
}

// NewIncidentBundleSourcePort retains successful applications without retaining
// jobs, sessions, mapping approvals, receipts or any source authentication state.
func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := committedSourceDescriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: exportCommittedSources,
		Prepare: func(ctx context.Context, members sourceport.Bundle, input sourceport.ImportContext) (any, error) {
			return PrepareCommittedSources(ctx, members, input.IncidentID, input.Actors)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, input sourceport.ImportContext) error {
			prepared, ok := value.(*PreparedCommittedSources)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return prepared.apply(ctx, tx, input)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, input sourceport.ImportContext) error {
			prepared, ok := value.(*PreparedCommittedSources)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			var count int64
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM import_committed_sources WHERE incident_id=$1`, input.IncidentID).Scan(&count); err != nil {
				return err
			}
			if count != prepared.count {
				return descriptor.DeclaredFailure("committed_sources.source_identity_admitted")
			}
			if err := prepared.validateRecordBindings(ctx, tx); err != nil {
				return err
			}
			if err := incidentportability.ValidateTopLevelAttributions(ctx, prepared.members, committedSourcesPath, "import_committed_sources", "committed_source_id", input.Attributions); err != nil {
				return descriptor.DeclaredFailure("committed_sources.attribution_exact")
			}
			return nil
		},
	})
}

// The transport is one bounded row per association, never a source-sized array.
func exportCommittedSourceRecords(ctx context.Context, input sourceport.ExportContext) error {
	return incidentportability.WriteMemberFrom(ctx, input.Members, committedSourceRecordsPath, func(ctx context.Context, out io.Writer) error {
		afterSource, afterRow := uuid.Nil, 0
		for {
			value := committedSourceRecord{SchemaID: "cartulary.imports.committed_source_record.v1"}
			err := input.Query.QueryRow(ctx, `SELECT committed_source_id,source_row_ref,record_id
FROM import_committed_source_records WHERE incident_id=$1 AND (committed_source_id,source_row_ref)>($2,$3)
ORDER BY committed_source_id,source_row_ref LIMIT 1`, input.IncidentID, afterSource, afterRow).Scan(&value.CommittedSourceID, &value.SourceRowRef, &value.RecordID)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			raw, err := json.Marshal(value)
			if err != nil {
				return err
			}
			canonical, err := incidentportability.CanonicalRawJSON(raw)
			if err != nil {
				return err
			}
			if n, err := out.Write(canonical); err != nil {
				return err
			} else if n != len(canonical) {
				return io.ErrShortWrite
			}
			afterSource, afterRow = value.CommittedSourceID, value.SourceRowRef
		}
	})
}

func exportCommittedSources(ctx context.Context, input sourceport.ExportContext) error {
	// Fetch one identity at a time. No collection-sized metadata or source buffer is
	// held, and each source/attribution read borrows the caller's exact snapshot.
	err := incidentportability.WriteMemberFrom(ctx, input.Members, committedSourcesPath, func(ctx context.Context, out io.Writer) error {
		after := uuid.Nil
		for {
			var id uuid.UUID
			err := input.Query.QueryRow(ctx, `SELECT committed_source_id FROM import_committed_sources WHERE incident_id=$1 AND committed_source_id>$2 ORDER BY committed_source_id LIMIT 1`, input.IncidentID, after).Scan(&id)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			source, err := ReadCommittedSourceTx(ctx, input.Query, input.IncidentID, id)
			if err != nil {
				return err
			}
			if input.PortableAttributions == nil {
				return fmt.Errorf("committed source export requires historical attribution")
			}
			actors, err := input.PortableAttributions.ResolvePortableSourceActors(ctx, input.Query, input.IncidentID, "import_committed_sources", "created_by_user_id", []string{id.String()})
			if err != nil {
				return err
			}
			if actor, ok := actors[id.String()]; ok {
				source.CreatedByUserID, err = uuid.Parse(actor)
				if err != nil {
					return err
				}
			}
			if input.Actors != nil {
				if err := input.Actors.RecordActorReference(source.CreatedByUserID); err != nil {
					return err
				}
			}
			if err := validateCommittedSource(source, input.IncidentID); err != nil {
				return err
			}
			raw, err := json.Marshal(source)
			if err != nil {
				return err
			}
			canonical, err := incidentportability.CanonicalRawJSON(raw)
			if err != nil {
				return err
			}
			if len(canonical)-1 > incidentportability.MaxNDJSONLogicalLineBytes {
				return committedSourceDescriptor().DeclaredFailure("committed_sources.metadata_exact")
			}
			if n, err := out.Write(canonical); err != nil {
				return err
			} else if n != len(canonical) {
				return io.ErrShortWrite
			}
			after = id
		}
	})
	if err != nil {
		return err
	}
	if err := exportCommittedSourceRecords(ctx, input); err != nil {
		return err
	}
	after := ""
	for {
		var digest string
		var sourceID uuid.UUID
		err := input.Query.QueryRow(ctx, `SELECT content_sha256,committed_source_id FROM import_committed_sources WHERE incident_id=$1 AND content_sha256>$2 ORDER BY content_sha256,committed_source_id LIMIT 1`, input.IncidentID, after).Scan(&digest, &sourceID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		source, reader, err := OpenCommittedSourceTx(ctx, input.Query, input.IncidentID, sourceID)
		if err != nil {
			return err
		}
		member, writeErr := input.Members.WriteMember(ctx, committedSourceBlobPrefix+digest, reader)
		if err := errors.Join(writeErr, reader.Close()); err != nil {
			return err
		}
		if member.SHA256 != digest || member.ByteSize != source.ByteSize {
			return committedSourceDescriptor().DeclaredFailure("committed_sources.blob_exact")
		}
		after = digest
	}
}

// PreparedCommittedSources is a read-only dependency capability over admitted,
// sealed inputs. Consumers can request sources only within the admitted incident.
// It never exposes an archive, SQL handle, storage key or live Import workflow.
type PreparedCommittedSources struct {
	incidentID  uuid.UUID
	members     *incidentportability.BoundMembers
	count       int64
	recordCount int64
}

func PrepareCommittedSources(ctx context.Context, members incidentportability.MemberReader, incidentID uuid.UUID, actors sourceport.ActorCatalog) (*PreparedCommittedSources, error) {
	descriptor := committedSourceDescriptor()
	sourceRows, err := sourceport.PrepareMembers(ctx, descriptor, members, 7)
	if err != nil {
		return nil, err
	}
	catalog := map[string]incidentportability.Member{}
	for _, member := range members.Members() {
		if strings.HasPrefix(member.Path, committedSourceBlobPrefix) {
			digest := strings.TrimPrefix(member.Path, committedSourceBlobPrefix)
			if !sourceDigestValid(digest) || member.SHA256 != digest || member.ByteSize < 0 {
				return nil, descriptor.DeclaredFailure("committed_sources.blob_exact")
			}
			if _, exists := catalog[member.Path]; exists {
				return nil, descriptor.DeclaredFailure("committed_sources.blob_exact")
			}
			catalog[member.Path] = member
		}
	}
	used := map[string]bool{}
	origins := map[uuid.UUID]bool{}
	firstRows := map[uuid.UUID]int{}
	var count int64
	paths := []string{committedSourcesPath, committedSourceRecordsPath}
	err = walkCommittedSources(ctx, sourceRows, func(source CommittedSource) error {
		if err := validateCommittedSource(source, incidentID); err != nil {
			return err
		}
		if _, ok := actors.Lookup(source.CreatedByUserID.String()); !ok {
			return descriptor.DeclaredFailure("committed_sources.attribution_exact")
		}
		if origins[source.OriginImportUnitID] {
			return descriptor.DeclaredFailure("committed_sources.source_identity_admitted")
		}
		// Origin identities are a bounded validation index, separate from bulk bytes.
		if len(origins) >= 64*1024*1024/128 {
			return descriptor.DeclaredFailure("committed_sources.metadata_exact")
		}
		origins[source.OriginImportUnitID] = true
		if source.TargetKind == ImportTargetKindViewSchema {
			firstRows[source.CommittedSourceID] = source.DataStartRowRef
		}
		path := committedSourceBlobPrefix + source.ContentSHA256
		member, ok := catalog[path]
		if !ok || member.ByteSize != source.ByteSize {
			return descriptor.DeclaredFailure("committed_sources.blob_exact")
		}
		if !used[path] {
			reader, err := incidentportability.OpenVerifiedMember(ctx, members, member)
			if err != nil {
				return err
			}
			_, readErr := io.Copy(io.Discard, reader)
			if err := errors.Join(readErr, reader.Close()); err != nil {
				return err
			}
			used[path] = true
			paths = append(paths, path)
		}
		count++
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(used) != len(catalog) {
		return nil, descriptor.DeclaredFailure("committed_sources.blob_exact")
	}
	bound, err := incidentportability.BindMembers(members, paths)
	if err != nil {
		return nil, err
	}
	prepared := &PreparedCommittedSources{incidentID: incidentID, members: bound, count: count}
	if err := walkCommittedSourceRecords(ctx, bound, func(record committedSourceRecord) error {
		first, ok := firstRows[record.CommittedSourceID]
		if !ok || record.SourceRowRef < first {
			return recordBindingFailure()
		}
		prepared.recordCount++
		return nil
	}); err != nil {
		return nil, err
	}
	return prepared, nil
}

func (p *PreparedCommittedSources) ReadSource(ctx context.Context, id uuid.UUID) (CommittedSource, error) {
	if p == nil || p.members == nil || id == uuid.Nil {
		return CommittedSource{}, pgx.ErrNoRows
	}
	var found CommittedSource
	err := walkCommittedSources(ctx, p.members, func(source CommittedSource) error {
		if source.CommittedSourceID == id {
			found = source
		}
		return nil
	})
	if err != nil {
		return CommittedSource{}, err
	}
	if found.CommittedSourceID == uuid.Nil {
		return CommittedSource{}, pgx.ErrNoRows
	}
	return found, nil
}
func (p *PreparedCommittedSources) OpenSource(ctx context.Context, id uuid.UUID) (CommittedSource, io.ReadCloser, error) {
	source, err := p.ReadSource(ctx, id)
	if err != nil {
		return CommittedSource{}, nil, err
	}
	reader, _, err := p.members.OpenMember(ctx, committedSourceBlobPrefix+source.ContentSHA256)
	return source, reader, err
}

func walkCommittedSources(ctx context.Context, members incidentportability.MemberReader, visit func(CommittedSource) error) error {
	return incidentportability.WalkPath(ctx, members, committedSourcesPath, func(row map[string]any) error {
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		var source CommittedSource
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&source); err != nil {
			return committedSourceDescriptor().DeclaredFailure("committed_sources.metadata_exact")
		}
		// Closed non-null fields are required even when their legal value is empty.
		encoded, err := json.Marshal(source)
		if err != nil {
			return err
		}
		var shape map[string]json.RawMessage
		if err := json.Unmarshal(encoded, &shape); err != nil {
			return err
		}
		if len(shape) != len(row) {
			return committedSourceDescriptor().DeclaredFailure("committed_sources.metadata_exact")
		}
		for field := range shape {
			if value, exists := row[field]; !exists || value == nil {
				return committedSourceDescriptor().DeclaredFailure("committed_sources.metadata_exact")
			}
		}
		return visit(source)
	})
}

func sourceDigestValid(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && value == strings.ToLower(value)
}

func validateCommittedSource(source CommittedSource, incidentID uuid.UUID) error {
	descriptor := committedSourceDescriptor()
	bad := func() error { return descriptor.DeclaredFailure("committed_sources.metadata_exact") }
	if source.SchemaID != "cartulary.imports.committed_source.v1" || source.IncidentID != incidentID || incidentID == uuid.Nil || source.CommittedSourceID == uuid.Nil || source.OriginImportUnitID == uuid.Nil || source.OriginImportSessionID == uuid.Nil || source.CreatedByUserID == uuid.Nil || source.CommittedSourceID != CommittedSourceID(source.OriginImportUnitID) || source.CommittedAt.IsZero() || !sourceDigestValid(source.ContentSHA256) || source.ByteSize < 0 || !sourceDigestValid(source.MappingFingerprint) {
		return bad()
	}
	if source.ParserProfileID != parserProfileWorkbookImport || source.ParserVersion != parserVersionWorkbookImport || source.OriginalFilename == "" || source.HeaderRowRef < 1 || source.DataStartRowRef <= source.HeaderRowRef {
		return bad()
	}
	switch source.SourceFileKind {
	case sourceFileKindCSV:
		if source.SourceMediaType != mediaTypeCSV && source.SourceMediaType != mediaTypeOctetStream {
			return bad()
		}
		if source.LocatorKind != "csv_file" || source.Locator != "file" {
			return bad()
		}
	case sourceFileKindXLSX:
		if source.SourceMediaType != mediaTypeXLSX && source.SourceMediaType != mediaTypeOctetStream {
			return bad()
		}
		var locator map[string]json.RawMessage
		if json.Unmarshal([]byte(source.Locator), &locator) != nil || len(locator) == 0 || source.SourceRectA1 == "" {
			return bad()
		}
	default:
		return bad()
	}
	var mapping approvedMapping
	decoder := json.NewDecoder(bytes.NewReader(source.ApprovedMapping))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&mapping); err != nil {
		return descriptor.DeclaredFailure("committed_sources.mapping_exact")
	}
	payload := normalizedMappingPayload(mappingRequest{HeaderRowRef: source.HeaderRowRef, DataStartRowRef: source.DataStartRowRef, approvedMapping: mapping})
	payload["client_txn_id"] = "committed-source-validation"
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, apiErr := decodeMappingRequest(bytes.NewReader(raw), nil)
	if apiErr != nil || len(mapping.SourceColumns) == 0 || mapping.targetKindOrDefault() != source.TargetKind || mapping.TargetViewSchemaID != source.TargetViewSchemaID || mapping.ExtensionProfileID != source.ExtensionProfileID {
		return descriptor.DeclaredFailure("committed_sources.mapping_exact")
	}
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok {
		return descriptor.DeclaredFailure("committed_sources.mapping_exact")
	}
	if source.TargetKind == ImportTargetKindViewSchema && (request.Fingerprint != source.MappingFingerprint || validateViewMapping(mapping, target) != nil) {
		return descriptor.DeclaredFailure("committed_sources.mapping_exact")
	}
	return nil
}

func (p *PreparedCommittedSources) apply(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext) error {
	if p == nil || p.incidentID != input.IncidentID || input.Attributions == nil {
		return sourceport.ErrPreparedBinding
	}
	for _, member := range p.members.Members() {
		if !strings.HasPrefix(member.Path, committedSourceBlobPrefix) {
			continue
		}
		digest := strings.TrimPrefix(member.Path, committedSourceBlobPrefix)
		result, err := tx.Exec(ctx, `INSERT INTO import_committed_source_blobs(content_sha256,byte_size) VALUES($1,$2) ON CONFLICT(content_sha256) DO NOTHING`, digest, member.ByteSize)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			var size int64
			if err := tx.QueryRow(ctx, `SELECT byte_size FROM import_committed_source_blobs WHERE content_sha256=$1`, digest).Scan(&size); err != nil {
				return err
			}
			if size != member.ByteSize {
				return committedSourceDescriptor().DeclaredFailure("committed_sources.blob_exact")
			}
			reader := newCommittedBlobReader(ctx, tx, digest, size)
			_, err = io.Copy(io.Discard, reader)
			if err := errors.Join(err, reader.Close()); err != nil {
				return err
			}
			continue
		}
		if err := p.insertBlob(ctx, tx, member, digest); err != nil {
			return err
		}
	}
	err := walkCommittedSources(ctx, p.members, func(source CommittedSource) error {
		if err := input.Attributions.RecordImportedAttribution("import_committed_sources", source.CommittedSourceID.String(), "created_by_user_id", source.CreatedByUserID.String()); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO import_committed_sources
(committed_source_id,incident_id,content_sha256,source_file_kind,source_media_type,original_filename,parser_profile_id,parser_version,locator_kind,locator,source_rect_a1,header_row_ref,data_start_row_ref,approved_mapping_json,mapping_fingerprint,target_kind,target_view_schema_id,extension_profile_id,origin_import_session_id,origin_import_unit_id,created_by_user_id,committed_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`, source.CommittedSourceID, source.IncidentID, source.ContentSHA256, source.SourceFileKind, source.SourceMediaType, source.OriginalFilename, source.ParserProfileID, source.ParserVersion, source.LocatorKind, source.Locator, source.SourceRectA1, source.HeaderRowRef, source.DataStartRowRef, source.ApprovedMapping, source.MappingFingerprint, source.TargetKind, source.TargetViewSchemaID, source.ExtensionProfileID, source.OriginImportSessionID, source.OriginImportUnitID, input.ActorUserID, source.CommittedAt)
		return err
	})
	if err != nil {
		return err
	}
	return p.applyRecordBindings(ctx, tx)
}
func (p *PreparedCommittedSources) insertBlob(ctx context.Context, tx pgx.Tx, member incidentportability.Member, digest string) error {
	reader, err := incidentportability.OpenVerifiedMember(ctx, p.members, member)
	if err != nil {
		return err
	}
	defer reader.Close()
	buffer := make([]byte, sourceReadChunkBytes)
	var ordinal int64
	for {
		n, readErr := io.ReadFull(reader, buffer)
		if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			return readErr
		}
		if n > 0 {
			if _, err := tx.Exec(ctx, `INSERT INTO import_committed_source_chunks(content_sha256,chunk_ordinal,source_bytes) VALUES($1,$2,$3)`, digest, ordinal, buffer[:n]); err != nil {
				return err
			}
			ordinal++
		}
		if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
			return reader.Close()
		}
	}
}
