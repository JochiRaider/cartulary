package imports

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

// CommittedSource is durable documentary provenance. Origin identifiers carry
// no capability and are never dereferenced to obtain the retained bytes.
type CommittedSource struct {
	SchemaID              string          `json:"schema_id"`
	CommittedSourceID     uuid.UUID       `json:"committed_source_id"`
	IncidentID            uuid.UUID       `json:"incident_id"`
	ContentSHA256         string          `json:"content_sha256"`
	ByteSize              int64           `json:"byte_size"`
	SourceFileKind        string          `json:"source_file_kind"`
	SourceMediaType       string          `json:"source_media_type"`
	OriginalFilename      string          `json:"original_filename"`
	ParserProfileID       string          `json:"parser_profile_id"`
	ParserVersion         string          `json:"parser_version"`
	LocatorKind           string          `json:"locator_kind"`
	Locator               string          `json:"locator"`
	SourceRectA1          string          `json:"source_rect_a1"`
	HeaderRowRef          int             `json:"header_row_ref"`
	DataStartRowRef       int             `json:"data_start_row_ref"`
	ApprovedMapping       json.RawMessage `json:"approved_mapping"`
	MappingFingerprint    string          `json:"mapping_fingerprint"`
	TargetKind            string          `json:"target_kind"`
	TargetViewSchemaID    string          `json:"target_view_schema_id"`
	ExtensionProfileID    string          `json:"extension_profile_id"`
	OriginImportSessionID uuid.UUID       `json:"origin_import_session_id"`
	OriginImportUnitID    uuid.UUID       `json:"origin_import_unit_id"`
	CreatedByUserID       uuid.UUID       `json:"created_by_user_id"`
	CommittedAt           time.Time       `json:"committed_at"`
}

type committedSourceSummary struct {
	CommittedSourceID  uuid.UUID `json:"committed_source_id"`
	ContentSHA256      string    `json:"content_sha256"`
	OriginalFilename   string    `json:"original_filename"`
	TargetKind         string    `json:"target_kind"`
	TargetViewSchemaID string    `json:"target_view_schema_id"`
	ExtensionProfileID string    `json:"extension_profile_id"`
	CommittedAt        time.Time `json:"committed_at"`
}

func (s *service) handleCommittedSources(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, importEditingRoles); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the list query is route handling, decided after the
	// incident's visibility and the source-access role.
	query, apiErr := decodeCommittedSourceListQuery(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	var err error
	recordID := query.recordID
	binding, cursor, reason := s.cursorCodec.ResolveListRequest(query.paging, "imports.committed_sources.list", principal.User.ID.String(), map[string]string{"incident_id": incidentID.String(), "record_id": recordID.String()})
	if reason != "" {
		writeAPIError(w, r, invalidPaginationRequest(reason))
		return
	}
	after := uuid.Nil
	if cursor != nil {
		after, err = uuid.Parse(cursor.Position["committed_source_id"])
		if err != nil || cursor.Mode != pagination.ModeKeyset || len(cursor.Position) != 1 {
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
			return
		}
	}
	rows, err := s.store.db.Query(r.Context(), `SELECT committed_source_id,content_sha256,original_filename,target_kind,target_view_schema_id,extension_profile_id,committed_at
FROM import_committed_sources s WHERE incident_id=$1 AND committed_source_id>$2
AND ($4::uuid='00000000-0000-0000-0000-000000000000' OR EXISTS (
SELECT 1 FROM import_committed_source_records b WHERE b.incident_id=s.incident_id AND b.committed_source_id=s.committed_source_id AND b.record_id=$4))
ORDER BY committed_source_id LIMIT $3`, incidentID, after, binding.Limit+1, recordID)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	defer rows.Close()
	values := make([]committedSourceSummary, 0, binding.Limit+1)
	for rows.Next() {
		var value committedSourceSummary
		if err := rows.Scan(&value.CommittedSourceID, &value.ContentSHA256, &value.OriginalFilename, &value.TargetKind, &value.TargetViewSchemaID, &value.ExtensionProfileID, &value.CommittedAt); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var next *string
	if len(values) > binding.Limit {
		values = values[:binding.Limit]
		token, err := s.cursorCodec.Encode(pagination.Cursor{Mode: pagination.ModeKeyset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope, Position: map[string]string{"committed_source_id": values[len(values)-1].CommittedSourceID.String()}})
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		next = &token
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"committed_sources": values}, httpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next})
}

// committedSourceListQuery is the admitted query of the committed-source
// collection: the optional record_id filter and the common paging members.
type committedSourceListQuery struct {
	recordID uuid.UUID
	paging   url.Values
}

// decodeImportListQuery admits an Imports GET collection's raw query under
// Core 01 REQ-01-583 in this order: raw-query integrity, duplicate member
// names, pagination aliases, members outside limit, cursor_token and the
// route's declared list-query members, then an empty limit or cursor_token
// (§3.3.6.2). Paging values are otherwise left to the cursor codec.
func decodeImportListQuery(raw string, declared ...string) (url.Values, *httpapi.APIError) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, invalidListQuery("malformed_query")
	}
	for key, entries := range values {
		if !utf8.ValidString(key) {
			return nil, invalidListQuery("malformed_query")
		}
		for _, value := range entries {
			if !utf8.ValidString(value) {
				return nil, invalidListQuery("malformed_query")
			}
		}
	}
	for _, entries := range values {
		if len(entries) != 1 {
			return nil, invalidListQuery("duplicate_query_member")
		}
	}
	for _, alias := range []string{"page", "offset", "page_size", "block_size"} {
		if values.Has(alias) {
			return nil, invalidPaginationRequest(pagination.ReasonInvalidLimit)
		}
	}
	for key := range values {
		if key != "limit" && key != "cursor_token" && !slices.Contains(declared, key) {
			return nil, invalidListQuery("unknown_query_member")
		}
	}
	if values.Has("limit") && values.Get("limit") == "" {
		return nil, invalidPaginationRequest(pagination.ReasonInvalidLimit)
	}
	if values.Has("cursor_token") && values.Get("cursor_token") == "" {
		return nil, invalidPaginationRequest(pagination.ReasonInvalidCursorToken)
	}
	return values, nil
}

// decodeCommittedSourceListQuery admits the collection's list-query contract
// (Core 01 REQ-01-685 under REQ-01-583): the common list-query checks with
// record_id declared, then the record_id value.
func decodeCommittedSourceListQuery(raw string) (committedSourceListQuery, *httpapi.APIError) {
	values, apiErr := decodeImportListQuery(raw, "record_id")
	if apiErr != nil {
		return committedSourceListQuery{}, apiErr
	}
	query := committedSourceListQuery{recordID: uuid.Nil, paging: values}
	if values.Has("record_id") {
		recordID, err := uuid.Parse(values.Get("record_id"))
		if err != nil || recordID == uuid.Nil {
			return committedSourceListQuery{}, invalidListQuery("invalid_filter_value")
		}
		query.recordID = recordID
		values.Del("record_id")
	}
	return query, nil
}

func invalidListQuery(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_list_query",
		Message: "invalid list query",
		Details: map[string]any{"reason_code": reasonCode},
	}
}

// CommittedSourceRead is the read-only physical adapter admitted by Imports.
// Durable callers borrow a snapshot without acquiring workflow or write access.
type CommittedSourceRead interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// ReadCommittedSourceTx borrows the caller's snapshot and reads only Imports
// state. Authorization and incident scope are supplied by the calling owner.
func ReadCommittedSourceTx(ctx context.Context, tx CommittedSourceRead, incidentID, sourceID uuid.UUID) (CommittedSource, error) {
	value := CommittedSource{SchemaID: "cartulary.imports.committed_source.v1"}
	err := tx.QueryRow(ctx, `SELECT s.committed_source_id,s.incident_id,s.content_sha256,b.byte_size,
s.source_file_kind,s.source_media_type,s.original_filename,s.parser_profile_id,s.parser_version,
s.locator_kind,s.locator,s.source_rect_a1,s.header_row_ref,s.data_start_row_ref,s.approved_mapping_json,s.mapping_fingerprint,
s.target_kind,s.target_view_schema_id,s.extension_profile_id,s.origin_import_session_id,
s.origin_import_unit_id,s.created_by_user_id,s.committed_at
FROM import_committed_sources s JOIN import_committed_source_blobs b USING(content_sha256)
WHERE s.incident_id=$1 AND s.committed_source_id=$2`, incidentID, sourceID).Scan(
		&value.CommittedSourceID, &value.IncidentID, &value.ContentSHA256, &value.ByteSize,
		&value.SourceFileKind, &value.SourceMediaType, &value.OriginalFilename, &value.ParserProfileID, &value.ParserVersion,
		&value.LocatorKind, &value.Locator, &value.SourceRectA1, &value.HeaderRowRef, &value.DataStartRowRef, &value.ApprovedMapping, &value.MappingFingerprint,
		&value.TargetKind, &value.TargetViewSchemaID, &value.ExtensionProfileID, &value.OriginImportSessionID,
		&value.OriginImportUnitID, &value.CreatedByUserID, &value.CommittedAt)
	return value, err
}

// OpenCommittedSourceTx verifies incident ownership from retained metadata;
// callers cannot turn an arbitrary digest into a readable blob capability.
func OpenCommittedSourceTx(ctx context.Context, tx CommittedSourceRead, incidentID, sourceID uuid.UUID) (CommittedSource, io.ReadCloser, error) {
	value, err := ReadCommittedSourceTx(ctx, tx, incidentID, sourceID)
	if err != nil {
		return CommittedSource{}, nil, err
	}
	return value, newCommittedBlobReader(ctx, tx, value.ContentSHA256, value.ByteSize), nil
}

func newCommittedBlobReader(ctx context.Context, tx CommittedSourceRead, digest string, size int64) *sourceReader {
	reader := newSourceReader(ctx, nil, ImportSourceCapability{SourceStreamRef: digest, SourceContentSHA256: digest, SourceByteSize: size}, false)
	reader.readChunk = func(offset, length int64) ([]byte, error) {
		var data []byte
		err := tx.QueryRow(ctx, committedBlobChunkQuery, digest, offset, length).Scan(&data)
		return data, err
	}
	return reader
}

const committedBlobChunkQuery = `SELECT COALESCE((SELECT substring(source_bytes FROM (($2::bigint-1)%65536+1)::integer FOR $3)
FROM import_committed_source_chunks WHERE content_sha256=$1 AND chunk_ordinal=($2::bigint-1)/65536),'\x'::bytea)`

func (s *service) handleGetCommittedSource(w http.ResponseWriter, r *http.Request) {
	s.serveCommittedSource(w, r, false)
}

func (s *service) handleGetCommittedSourceContent(w http.ResponseWriter, r *http.Request) {
	s.serveCommittedSource(w, r, true)
}

// serveCommittedSource reads one committed source's metadata, or streams its
// exact bytes. Central admission has authenticated the caller and admitted
// incident_id and committed_source_id (Core 01 REQ-01-707).
func (s *service) serveCommittedSource(w http.ResponseWriter, r *http.Request, content bool) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	sourceID, apiErr := admittedPathUUID(r, "committed_source_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, importEditingRoles); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton-read query check is route handling,
	// after the incident's visibility and role gate.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	tx, err := s.store.db.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	defer tx.Rollback(context.WithoutCancel(r.Context()))
	value, reader, err := OpenCommittedSourceTx(r.Context(), tx, incidentID, sourceID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeAPIError(w, r, httpapi.NotFound("committed_source_not_found"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	defer reader.Close()
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if content {
		w.Header().Set("Content-Type", value.SourceMediaType)
		w.Header().Set("Content-Length", strconv.FormatInt(value.ByteSize, 10))
		w.Header().Set("Content-Disposition", `attachment; filename="source.`+value.SourceFileKind+`"`)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		if _, err := io.Copy(w, reader); err != nil {
			panic(http.ErrAbortHandler)
		}
		return
	}
	actors, err := s.importedAttributions.ResolveImportedSourceActorsTx(r.Context(), tx, incidentID, "import_committed_sources", "created_by_user_id", []string{sourceID.String()})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if actor, ok := actors[sourceID.String()]; ok {
		value.CreatedByUserID, err = uuid.Parse(actor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, value)
}
