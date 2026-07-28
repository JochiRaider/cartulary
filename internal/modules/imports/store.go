package imports

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrNotFound = errors.New("imports: not found")
var ErrStateConflict = errors.New("imports: state conflict")
var ErrApplyBlocked = errors.New("imports: apply blocked")

type StateConflictError struct {
	ReasonCode string
}

func (e *StateConflictError) Error() string {
	return ErrStateConflict.Error()
}

func (e *StateConflictError) Unwrap() error {
	return ErrStateConflict
}

type ApplyBlockedError struct {
	ReasonCode string
	Field      string
}

func (e *ApplyBlockedError) Error() string {
	return ErrApplyBlocked.Error()
}

func (e *ApplyBlockedError) Unwrap() error {
	return ErrApplyBlocked
}

type Store struct {
	pool             *pgxpool.Pool
	incidentAccess   incidents.Access
	revisionAppender *revisions.Appender
	jobTransactions  *jobs.TransactionService
	intents          collaboration.IntentAppender
}

type DiscoveredUnit struct {
	LocatorKind         string
	Locator             string
	SourceRectA1        string
	HeaderRowRef        int
	DataStartRowRef     int
	InferredRowCount    int
	InferredColumnCount int
	WarningCodes        []string
	Columns             []map[string]any
	SourceRows          []map[string]any
	PreviewRows         []map[string]any
}

type CreateAcceptedSessionParams struct {
	ActorUserID         uuid.UUID
	Request             CreateSessionRequest
	SourceFileKind      string
	OriginalFilename    string
	SourceContentSHA256 string
	SourceMediaType     string
	SourceByteSize      int64
	SourceBytes         []byte
	Units               []DiscoveredUnit
	NormalizedRequest   []byte
	Now                 time.Time
}

type CreateAcceptedSessionResult struct {
	Job             jobs.Resource
	ImportSessionID uuid.UUID
	Replayed        bool
}

type MappingParams struct {
	ActorUserID       uuid.UUID
	SessionID         uuid.UUID
	UnitID            uuid.UUID
	Request           MappingRequest
	NormalizedRequest []byte
	Now               time.Time
}

type UnitActionParams struct {
	ActorUserID       uuid.UUID
	SessionID         uuid.UUID
	UnitID            uuid.UUID
	RouteKey          string
	Request           ActionRequest
	NormalizedRequest []byte
	Now               time.Time
}

type UnitActionResult struct {
	Payload    map[string]any
	IncidentID uuid.UUID
	Replayed   bool
}

type ApplyStartParams struct {
	ActorUserID       uuid.UUID
	SessionID         uuid.UUID
	Request           ApplyRequest
	NormalizedRequest []byte
	Now               time.Time
}

type ApplyStartResult struct {
	Job             jobs.Resource
	IncidentID      uuid.UUID
	ImportSessionID uuid.UUID
	ClientTxnID     string
	SelectedUnitIDs []uuid.UUID
	Replayed        bool
}

type discoveryJobHandlerPayload struct {
	ImportSessionID string `json:"import_session_id"`
}

type applyJobHandlerPayload struct {
	IncidentID      string   `json:"incident_id"`
	ImportSessionID string   `json:"import_session_id"`
	ActorUserID     string   `json:"actor_user_id"`
	ClientTxnID     string   `json:"client_txn_id"`
	SelectedUnitIDs []string `json:"selected_unit_ids"`
}

type ApplyUnitData struct {
	UnitID              uuid.UUID
	SourceRows          []map[string]any
	ApprovedMapping     ApprovedMapping
	MappingFingerprint  string
	SourceFileKind      string
	SourceContentSHA256 string
	SourceStreamRef     string
	ParserProfileID     string
	ParserVersion       string
	LocatorKind         string
	Locator             string
	SourceRectA1        string
}

type ApplyJournalParams struct {
	ImportSessionID      uuid.UUID
	ImportUnitID         uuid.UUID
	MappingFingerprint   string
	SourceRowRef         int
	TargetViewSchemaID   string
	OwnerCreateFacade    string
	RecordID             uuid.UUID
	RowVersion           int64
	ChangeSetID          uuid.UUID
	ChangeSetMutationRef string
	OwnerResultCode      string
	CreatedOrReused      string
	OwnerResponse        map[string]any
	RowRefresh           map[string]any
	CreatedAt            time.Time
}

func NewStore(
	pool *pgxpool.Pool,
	appender *revisions.Appender,
	jobTransactions *jobs.TransactionService,
	intents collaboration.IntentAppender,
) *Store {
	return &Store{
		pool:             pool,
		incidentAccess:   incidents.NewAccess(pool),
		revisionAppender: appender,
		jobTransactions:  jobTransactions,
		intents:          intents,
	}
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Store) CreateAcceptedSession(ctx context.Context, params CreateAcceptedSessionParams) (CreateAcceptedSessionResult, error) {
	sum := sha256.Sum256(params.NormalizedRequest)
	requestHash := sum[:]
	key := authn.RouteIdempotencyKey{
		RouteKey:    "imports.sessions.create",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.Request.IncidentID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return CreateAcceptedSessionResult{}, authn.ErrClientTxnConflict
		}
		var resource jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &resource); err != nil {
			return CreateAcceptedSessionResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateAcceptedSessionResult{}, err
		}
		return CreateAcceptedSessionResult{Job: resource, Replayed: true}, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return CreateAcceptedSessionResult{}, err
	}

	sessionID := uuid.New()
	if len(params.Units) == 0 {
		return CreateAcceptedSessionResult{}, fmt.Errorf("import session requires at least one discovered unit")
	}
	handlerPayload, err := json.Marshal(discoveryJobHandlerPayload{ImportSessionID: sessionID.String()})
	if err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &params.Request.IncidentID}
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		"import.discovery_v1",
		key,
		scope,
		params.NormalizedRequest,
	)
	if err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	job, err := s.jobTransactions.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             scope,
		SubmittedByUserID: params.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerName:       importDiscoveryJobHandlerName,
		HandlerPayload:    handlerPayload,
		Extension:         admission,
	}, params.Now.UTC())
	if err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	if _, err := tx.Exec(ctx, `
INSERT INTO import_sessions (
    import_session_id, incident_id, created_by_user_id, client_txn_id, assistant_profile,
    source_file_kind, original_filename, source_content_sha256, source_media_type, source_byte_size,
    parser_profile_id, parser_version,
    session_status, discovery_job_id, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'created', $13, $14, $14)
`, sessionID, params.Request.IncidentID, params.ActorUserID, params.Request.ClientTxnID, params.Request.AssistantProfile, params.SourceFileKind, params.OriginalFilename, params.SourceContentSHA256, params.SourceMediaType, params.SourceByteSize, ParserProfileWorkbookImport, ParserVersionWorkbookImport, jobID, params.Now.UTC()); err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	for index, unit := range params.Units {
		unitID := uuid.New()
		sourceStreamRef := newImportSourceStreamRef()
		previewRows, err := json.Marshal(unit.PreviewRows)
		if err != nil {
			return CreateAcceptedSessionResult{}, err
		}
		sourceRows, err := json.Marshal(unit.SourceRows)
		if err != nil {
			return CreateAcceptedSessionResult{}, err
		}
		columns, err := json.Marshal(unit.Columns)
		if err != nil {
			return CreateAcceptedSessionResult{}, err
		}
		if _, err := tx.Exec(ctx, `
	INSERT INTO import_units (
	    import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
	    header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
	    warning_codes, columns_json, source_rows_json, preview_rows_json, source_stream_ref,
	    discovery_sequence, created_at, updated_at
	)
	VALUES ($1, $2, 'discovered', $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16)
	`, unitID, sessionID, unit.LocatorKind, unit.Locator, unit.SourceRectA1,
			unit.HeaderRowRef, unit.DataStartRowRef, unit.InferredRowCount, unit.InferredColumnCount,
			unit.WarningCodes, columns, sourceRows, previewRows, sourceStreamRef, index+1, params.Now.UTC()); err != nil {
			return CreateAcceptedSessionResult{}, err
		}
		if _, err := tx.Exec(ctx, `
	INSERT INTO import_source_streams (
	    source_stream_ref, import_session_id, import_unit_id, source_content_sha256,
	    source_media_type, source_byte_size, source_bytes, created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, sourceStreamRef, sessionID, unitID, params.SourceContentSHA256, params.SourceMediaType, params.SourceByteSize, params.SourceBytes, params.Now.UTC()); err != nil {
			return CreateAcceptedSessionResult{}, err
		}
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateAcceptedSessionResult{}, err
	}
	return CreateAcceptedSessionResult{Job: job, ImportSessionID: sessionID}, nil
}

func (s *Store) MarkDiscovered(ctx context.Context, sessionID uuid.UUID, now time.Time) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE import_sessions
   SET session_status = 'discovered',
       updated_at = $2
 WHERE import_session_id = $1
   AND session_status = 'created'
`, sessionID, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkDiscoveredTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = 'discovered',
       updated_at = $2
 WHERE import_session_id = $1
   AND session_status = 'created'
`, sessionID, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, sessionID uuid.UUID) (map[string]any, uuid.UUID, error) {
	row := s.pool.QueryRow(ctx, `
SELECT import_session_id, incident_id, created_by_user_id, client_txn_id, assistant_profile,
       source_file_kind, original_filename, source_content_sha256, parser_profile_id, parser_version,
       session_status, discovery_job_id, apply_job_id, to_jsonb(selected_unit_ids),
       blocking_diagnostics_json, to_jsonb(nonblocking_warning_codes), created_at, updated_at
  FROM import_sessions
 WHERE import_session_id = $1
`, sessionID)
	resource, incidentID, err := scanSessionResource(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, uuid.UUID{}, ErrNotFound
	}
	return resource, incidentID, err
}

func (s *Store) ListUnits(ctx context.Context, sessionID uuid.UUID) ([]map[string]any, uuid.UUID, error) {
	session, incidentID, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	_ = session
	rows, err := s.pool.Query(ctx, `
SELECT import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
       header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
       to_jsonb(warning_codes), mapping_fingerprint, approved_mapping_json, created_at, updated_at
 FROM import_units
 WHERE import_session_id = $1
 ORDER BY discovery_sequence ASC
`, sessionID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	defer rows.Close()
	units := make([]map[string]any, 0)
	for rows.Next() {
		unit, err := scanUnitResource(rows)
		if err != nil {
			return nil, uuid.UUID{}, err
		}
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, uuid.UUID{}, err
	}
	return units, incidentID, nil
}

func (s *Store) GetUnit(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (map[string]any, uuid.UUID, error) {
	_, incidentID, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	row := s.pool.QueryRow(ctx, `
SELECT import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
       header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
       to_jsonb(warning_codes), mapping_fingerprint, approved_mapping_json, created_at, updated_at
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID)
	unit, err := scanUnitResource(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, uuid.UUID{}, ErrNotFound
	}
	return unit, incidentID, err
}

func (s *Store) GetPreview(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (map[string]any, uuid.UUID, error) {
	unit, incidentID, err := s.GetUnit(ctx, sessionID, unitID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	row := s.pool.QueryRow(ctx, `
SELECT columns_json, preview_rows_json
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID)
	var columnsJSON []byte
	var previewJSON []byte
	if err := row.Scan(&columnsJSON, &previewJSON); err != nil {
		return nil, uuid.UUID{}, err
	}
	var columns any
	if err := json.Unmarshal(columnsJSON, &columns); err != nil {
		return nil, uuid.UUID{}, err
	}
	var previewRows any
	if err := json.Unmarshal(previewJSON, &previewRows); err != nil {
		return nil, uuid.UUID{}, err
	}
	preview := map[string]any{
		"import_session_id":     sessionID.String(),
		"import_unit_id":        unitID.String(),
		"locator_kind":          unit["locator_kind"],
		"locator":               unit["locator"],
		"source_rect_a1":        unit["source_rect_a1"],
		"header_row_ref":        unit["header_row_ref"],
		"data_start_row_ref":    unit["data_start_row_ref"],
		"inferred_row_count":    unit["inferred_row_count"],
		"inferred_column_count": unit["inferred_column_count"],
		"warning_codes":         unit["warning_codes"],
		"unit_status":           unit["unit_status"],
		"columns":               columns,
		"preview_rows":          previewRows,
		"truncated":             previewTruncated(unit, previewRows),
	}
	if unit["mapping_fingerprint"] != nil {
		preview["mapping_fingerprint"] = unit["mapping_fingerprint"]
	}
	return preview, incidentID, nil
}

func (s *Store) GetUnitColumns(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) ([]map[string]any, uuid.UUID, error) {
	_, incidentID, err := s.GetUnit(ctx, sessionID, unitID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	row := s.pool.QueryRow(ctx, `
SELECT columns_json
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID)
	var columnsJSON []byte
	if err := row.Scan(&columnsJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, uuid.UUID{}, ErrNotFound
		}
		return nil, uuid.UUID{}, err
	}
	var columns []map[string]any
	if err := json.Unmarshal(columnsJSON, &columns); err != nil {
		return nil, uuid.UUID{}, err
	}
	if columns == nil {
		columns = []map[string]any{}
	}
	return columns, incidentID, nil
}

func (s *Store) SaveMapping(ctx context.Context, params MappingParams) (map[string]any, uuid.UUID, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "imports.units.mapping",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.SessionID.String() + ":" + params.UnitID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	return s.withUnitMutation(ctx, key, params.NormalizedRequest, params.SessionID, params.UnitID, params.Now, func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string]any, error) {
		status, err := unitStatusTx(ctx, tx, params.SessionID, params.UnitID)
		if err != nil {
			return nil, err
		}
		switch status {
		case "applying":
			return nil, importConflictError("unit_applying")
		case "applied", "skipped", "rejected", "failed":
			return nil, importConflictError("unit_terminal")
		}
		nextStatus := "mapped"
		if status == "selected" || status == "ready" {
			nextStatus = "ready"
		}
		mappingJSON, err := json.Marshal(params.Request.ApprovedMapping)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
	UPDATE import_units
	   SET unit_status = $4,
	       header_row_ref = $5,
	       data_start_row_ref = $6,
	       mapping_fingerprint = $7,
	       approved_mapping_json = $8,
	       approved_target_kind = $9,
	       approved_extension_profile_id = $10,
	       approved_target_view_schema_id = $11,
	       updated_at = $12
	 WHERE import_session_id = $1
	   AND import_unit_id = $2
	   AND unit_status = $3
	`, params.SessionID, params.UnitID, status, nextStatus, params.Request.HeaderRowRef, params.Request.DataStartRowRef, params.Request.Fingerprint, mappingJSON, params.Request.ApprovedMapping.targetKindOrDefault(), nullableString(params.Request.ApprovedMapping.ExtensionProfileID), nullableString(params.Request.ApprovedMapping.TargetViewSchemaID), params.Now.UTC()); err != nil {
			return nil, err
		}
		if err := refreshSessionStatusTx(ctx, tx, params.SessionID, params.Now); err != nil {
			return nil, err
		}
		unit, err := scanUnitResource(tx.QueryRow(ctx, unitResourceSQL()+` WHERE import_session_id = $1 AND import_unit_id = $2`, params.SessionID, params.UnitID))
		if err != nil {
			return nil, err
		}
		return unit, nil
	})
}

func (s *Store) SelectUnit(ctx context.Context, params UnitActionParams) (UnitActionResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    params.RouteKey,
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.SessionID.String() + ":" + params.UnitID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	payload, incidentID, err := s.withUnitMutation(ctx, key, params.NormalizedRequest, params.SessionID, params.UnitID, params.Now, func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string]any, error) {
		status, err := unitStatusTx(ctx, tx, params.SessionID, params.UnitID)
		if err != nil {
			return nil, err
		}
		switch status {
		case "applying":
			return nil, importConflictError("unit_applying")
		case "applied", "skipped", "rejected", "failed":
			return nil, importConflictError("unit_terminal")
		}
		nextStatus := "selected"
		if status == "mapped" || status == "ready" {
			nextStatus = "ready"
		}
		if _, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = $4,
       updated_at = $5
 WHERE import_session_id = $1
   AND import_unit_id = $2
   AND unit_status = $3
`, params.SessionID, params.UnitID, status, nextStatus, params.Now.UTC()); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET selected_unit_ids = (
       SELECT ARRAY(
           SELECT DISTINCT unit_id
             FROM unnest(selected_unit_ids || ARRAY[$2]::uuid[]) AS unit_id
            ORDER BY unit_id
       )
   ),
       updated_at = $3
 WHERE import_session_id = $1
`, params.SessionID, params.UnitID, params.Now.UTC()); err != nil {
			return nil, err
		}
		if err := refreshSessionStatusTx(ctx, tx, params.SessionID, params.Now); err != nil {
			return nil, err
		}
		return unitActionPayloadTx(ctx, tx, params.SessionID, params.UnitID)
	})
	return UnitActionResult{Payload: payload, IncidentID: incidentID}, err
}

func (s *Store) SkipUnit(ctx context.Context, params UnitActionParams) (UnitActionResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    params.RouteKey,
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.SessionID.String() + ":" + params.UnitID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	payload, incidentID, err := s.withUnitMutation(ctx, key, params.NormalizedRequest, params.SessionID, params.UnitID, params.Now, func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string]any, error) {
		status, err := unitStatusTx(ctx, tx, params.SessionID, params.UnitID)
		if err != nil {
			return nil, err
		}
		switch status {
		case "applying":
			return nil, importConflictError("unit_applying")
		case "applied", "rejected", "failed":
			return nil, importConflictError("unit_terminal")
		}
		if status != "skipped" {
			if _, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = 'skipped',
       updated_at = $3
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, params.SessionID, params.UnitID, params.Now.UTC()); err != nil {
				return nil, err
			}
		}
		if _, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET selected_unit_ids = array_remove(selected_unit_ids, $2),
       updated_at = $3
 WHERE import_session_id = $1
`, params.SessionID, params.UnitID, params.Now.UTC()); err != nil {
			return nil, err
		}
		if err := refreshSessionStatusTx(ctx, tx, params.SessionID, params.Now); err != nil {
			return nil, err
		}
		return unitActionPayloadTx(ctx, tx, params.SessionID, params.UnitID)
	})
	return UnitActionResult{Payload: payload, IncidentID: incidentID}, err
}

func (s *Store) StartApply(ctx context.Context, params ApplyStartParams) (ApplyStartResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "imports.sessions.apply",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.SessionID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ApplyStartResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		incidentID, selected, err := sessionApplySelectionTx(ctx, tx, params.SessionID, params.Request.SelectedUnitIDs)
		if err != nil {
			return ApplyStartResult{}, err
		}
		requestHash, err := applyRequestHash(params.Request, selected)
		if err != nil {
			return ApplyStartResult{}, err
		}
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return ApplyStartResult{}, authn.ErrClientTxnConflict
		}
		var resource jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &resource); err != nil {
			return ApplyStartResult{}, err
		}
		return ApplyStartResult{Job: resource, IncidentID: incidentID, ImportSessionID: params.SessionID, ClientTxnID: params.Request.ClientTxnID, SelectedUnitIDs: selected, Replayed: true}, tx.Commit(ctx)
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return ApplyStartResult{}, err
	}

	status, incidentID, selected, err := sessionStatusAndSelectionTx(ctx, tx, params.SessionID, params.Request.SelectedUnitIDs)
	if err != nil {
		return ApplyStartResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return ApplyStartResult{}, err
	}
	switch status {
	case "applying":
		return ApplyStartResult{}, importConflictError("session_applying")
	case "applied", "partially_applied", "failed", "canceled":
		return ApplyStartResult{}, importApplyBlockedError("duplicate_apply_blocked")
	}
	if len(selected) == 0 {
		return ApplyStartResult{}, importApplyBlockedError("unit_not_ready")
	}
	if err := requireSelectedUnitsReadyTx(ctx, tx, params.SessionID, selected); err != nil {
		return ApplyStartResult{}, err
	}
	requestHash, err := applyRequestHash(params.Request, selected)
	if err != nil {
		return ApplyStartResult{}, err
	}
	normalizedRequest, err := normalizedApplyRequest(params.Request, selected)
	if err != nil {
		return ApplyStartResult{}, err
	}
	handlerPayload, err := json.Marshal(applyJobHandlerPayload{
		IncidentID:      incidentID.String(),
		ImportSessionID: params.SessionID.String(),
		ActorUserID:     params.ActorUserID.String(),
		ClientTxnID:     params.Request.ClientTxnID,
		SelectedUnitIDs: uuidStrings(selected),
	})
	if err != nil {
		return ApplyStartResult{}, err
	}
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		"import.apply_v1",
		key,
		scope,
		normalizedRequest,
	)
	if err != nil {
		return ApplyStartResult{}, err
	}
	job, err := s.jobTransactions.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             scope,
		SubmittedByUserID: params.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0, Total: intPtr(len(selected))},
		HandlerName:       importApplyJobHandlerName,
		HandlerPayload:    handlerPayload,
		Extension:         admission,
	}, params.Now.UTC())
	if err != nil {
		return ApplyStartResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	if _, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = 'applying',
       apply_job_id = $2,
       selected_unit_ids = $3,
       updated_at = $4
 WHERE import_session_id = $1
`, params.SessionID, jobID, selected, params.Now.UTC()); err != nil {
		return ApplyStartResult{}, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = 'applying',
       updated_at = $3
 WHERE import_session_id = $1
   AND import_unit_id = ANY($2)
`, params.SessionID, selected, params.Now.UTC()); err != nil {
		return ApplyStartResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return ApplyStartResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ApplyStartResult{}, err
	}
	return ApplyStartResult{Job: job, IncidentID: incidentID, ImportSessionID: params.SessionID, ClientTxnID: params.Request.ClientTxnID, SelectedUnitIDs: selected}, nil
}

func applyRequestHash(request ApplyRequest, resolvedSelected []uuid.UUID) ([]byte, error) {
	normalized, err := normalizedApplyRequest(request, resolvedSelected)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(normalized)
	return sum[:], nil
}

func normalizedApplyRequest(request ApplyRequest, resolvedSelected []uuid.UUID) ([]byte, error) {
	normalized := request.Normalized
	if request.SelectedUnitIDs == nil {
		var err error
		normalized, err = json.Marshal(map[string]any{
			"client_txn_id":     request.ClientTxnID,
			"selected_unit_ids": uuidStrings(resolvedSelected),
		})
		if err != nil {
			return nil, err
		}
	}
	return normalized, nil
}

func (s *Store) GetApplyUnits(ctx context.Context, sessionID uuid.UUID, unitIDs []uuid.UUID) ([]ApplyUnitData, error) {
	rows, err := s.pool.Query(ctx, `
SELECT u.import_unit_id,
       u.source_rows_json,
       u.approved_mapping_json,
       COALESCE(u.mapping_fingerprint, ''),
	       s.source_file_kind,
	       s.source_content_sha256,
	       COALESCE(u.source_stream_ref, ''),
	       s.parser_profile_id,
       s.parser_version,
       u.locator_kind,
       u.locator,
       u.source_rect_a1
  FROM import_units u
  JOIN import_sessions s
    ON s.import_session_id = u.import_session_id
 WHERE u.import_session_id = $1
   AND u.import_unit_id = ANY($2)
 ORDER BY u.discovery_sequence ASC
`, sessionID, unitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	units := make([]ApplyUnitData, 0)
	for rows.Next() {
		var unit ApplyUnitData
		var sourceRowsJSON []byte
		var mappingJSON []byte
		if err := rows.Scan(&unit.UnitID, &sourceRowsJSON, &mappingJSON, &unit.MappingFingerprint, &unit.SourceFileKind, &unit.SourceContentSHA256, &unit.SourceStreamRef, &unit.ParserProfileID, &unit.ParserVersion, &unit.LocatorKind, &unit.Locator, &unit.SourceRectA1); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(sourceRowsJSON, &unit.SourceRows); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(mappingJSON, &unit.ApprovedMapping); err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, rows.Err()
}

func (s *Store) CompleteApply(ctx context.Context, sessionID uuid.UUID, unitIDs []uuid.UUID, status string, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.CompleteApplyTx(ctx, tx, sessionID, unitIDs, status, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CompleteApplyTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitIDs []uuid.UUID, status string, now time.Time) error {
	command, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = $2,
       updated_at = $3
 WHERE import_session_id = $1
   AND session_status = 'applying'
   AND NOT EXISTS (
       SELECT 1
         FROM import_units
        WHERE import_session_id = $1
          AND import_unit_id = ANY($4)
          AND unit_status NOT IN ('applied', 'failed')
   )
`, sessionID, status, now.UTC(), unitIDs)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("import apply terminal state precondition failed")
	}
	return nil
}

func (s *Store) MarkApplyUnitStatus(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID, status string, now time.Time) error {
	if status != "applied" && status != "failed" {
		return fmt.Errorf("invalid terminal import unit status %q", status)
	}
	command, err := s.pool.Exec(ctx, `
UPDATE import_units
   SET unit_status = $3,
       updated_at = $4
 WHERE import_session_id = $1
   AND import_unit_id = $2
   AND unit_status = 'applying'
`, sessionID, unitID, status, now.UTC())
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		var existing string
		if scanErr := s.pool.QueryRow(ctx, `
SELECT unit_status
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID).Scan(&existing); scanErr == nil && existing == status {
			return nil
		}
		return fmt.Errorf("import unit terminal state precondition failed")
	}
	return nil
}

func (s *Store) FailApply(ctx context.Context, sessionID uuid.UUID, unitIDs []uuid.UUID, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = 'failed',
       updated_at = $3
 WHERE import_session_id = $1
   AND import_unit_id = ANY($2)
   AND unit_status = 'applying'
`, sessionID, unitIDs, now.UTC()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = 'failed',
       updated_at = $2
 WHERE import_session_id = $1
`, sessionID, now.UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CancelApply(ctx context.Context, sessionID uuid.UUID, unitIDs []uuid.UUID, now time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = 'failed',
       updated_at = $3
 WHERE import_session_id = $1
   AND import_unit_id = ANY($2)
   AND unit_status = 'applying'
`, sessionID, unitIDs, now.UTC()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = 'canceled',
       updated_at = $2
 WHERE import_session_id = $1
   AND session_status = 'applying'
`, sessionID, now.UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CancelDiscovery(ctx context.Context, sessionID uuid.UUID, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
UPDATE import_sessions
   SET session_status = 'canceled',
       updated_at = $2
 WHERE import_session_id = $1
   AND session_status = 'created'
`, sessionID, now.UTC())
	return err
}

func (s *Store) InsertApplyJournalTx(ctx context.Context, tx pgx.Tx, params ApplyJournalParams) error {
	ownerResponse, err := json.Marshal(params.OwnerResponse)
	if err != nil {
		return fmt.Errorf("encode import owner response: %w", err)
	}
	rowRefresh, err := json.Marshal(params.RowRefresh)
	if err != nil {
		return fmt.Errorf("encode import row refresh: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO import_apply_journal (
    import_session_id,
    import_unit_id,
    mapping_fingerprint,
    source_row_ref,
    target_view_schema_id,
    owner_create_facade,
    record_id,
    row_version,
    change_set_id,
    change_set_mutation_ref,
    owner_result_code,
    created_or_reused,
    owner_response_json,
    row_refresh_json,
    created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15
)
`, params.ImportSessionID, params.ImportUnitID, params.MappingFingerprint, params.SourceRowRef, params.TargetViewSchemaID,
		params.OwnerCreateFacade, params.RecordID, params.RowVersion, params.ChangeSetID, params.ChangeSetMutationRef,
		params.OwnerResultCode, params.CreatedOrReused, ownerResponse, rowRefresh, params.CreatedAt.UTC()); err != nil {
		return fmt.Errorf("insert import apply journal: %w", err)
	}
	return nil
}

func lookupRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID)
	var record authn.RouteIdempotencyRecord
	if err := row.Scan(&record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.ActorUserID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
		}
		return authn.RouteIdempotencyRecord{}, err
	}
	return record, nil
}

func (s *Store) withUnitMutation(
	ctx context.Context,
	key authn.RouteIdempotencyKey,
	normalizedRequest []byte,
	sessionID uuid.UUID,
	unitID uuid.UUID,
	now time.Time,
	mutate func(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error),
) (map[string]any, uuid.UUID, error) {
	sum := sha256.Sum256(normalizedRequest)
	requestHash := sum[:]
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return nil, uuid.UUID{}, authn.ErrClientTxnConflict
		}
		var payload map[string]any
		if err := json.Unmarshal(existing.ResponseJSON, &payload); err != nil {
			return nil, uuid.UUID{}, err
		}
		_, incidentID, err := sessionStatusTx(ctx, tx, sessionID)
		if err != nil {
			return nil, uuid.UUID{}, err
		}
		return payload, incidentID, tx.Commit(ctx)
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return nil, uuid.UUID{}, err
	}
	_, incidentID, err := sessionStatusTx(ctx, tx, sessionID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	payload, err := mutate(ctx, tx, incidentID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		return nil, uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, uuid.UUID{}, err
	}
	_ = unitID
	_ = now
	return payload, incidentID, nil
}

func unitResourceSQL() string {
	return `
SELECT import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
       header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
       to_jsonb(warning_codes), mapping_fingerprint, approved_mapping_json, created_at, updated_at
  FROM import_units`
}

func unitStatusTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID) (string, error) {
	var status string
	if err := tx.QueryRow(ctx, `
SELECT unit_status
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
 FOR UPDATE
`, sessionID, unitID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return status, nil
}

func sessionStatusTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID) (string, uuid.UUID, error) {
	var status string
	var incidentID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT session_status, incident_id
  FROM import_sessions
 WHERE import_session_id = $1
 FOR UPDATE
`, sessionID).Scan(&status, &incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", uuid.UUID{}, ErrNotFound
		}
		return "", uuid.UUID{}, err
	}
	return status, incidentID, nil
}

func unitActionPayloadTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID) (map[string]any, error) {
	session, _, err := scanSessionResource(tx.QueryRow(ctx, `
SELECT import_session_id, incident_id, created_by_user_id, client_txn_id, assistant_profile,
       source_file_kind, original_filename, source_content_sha256, parser_profile_id, parser_version,
       session_status, discovery_job_id, apply_job_id, to_jsonb(selected_unit_ids),
       blocking_diagnostics_json, to_jsonb(nonblocking_warning_codes), created_at, updated_at
  FROM import_sessions
 WHERE import_session_id = $1
`, sessionID))
	if err != nil {
		return nil, err
	}
	unit, err := scanUnitResource(tx.QueryRow(ctx, unitResourceSQL()+` WHERE import_session_id = $1 AND import_unit_id = $2`, sessionID, unitID))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"import_session_id": sessionID.String(),
		"session_status":    session["session_status"],
		"selected_unit_ids": session["selected_unit_ids"],
		"unit":              unit,
	}, nil
}

func refreshSessionStatusTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, now time.Time) error {
	var readyCount int
	var mappedCount int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*) FILTER (WHERE unit_status = 'ready'),
       COUNT(*) FILTER (WHERE unit_status IN ('mapped', 'selected'))
  FROM import_units
 WHERE import_session_id = $1
`, sessionID).Scan(&readyCount, &mappedCount); err != nil {
		return err
	}
	status := "discovered"
	if mappedCount > 0 {
		status = "mapped"
	}
	if readyCount > 0 {
		status = "ready_to_apply"
	}
	_, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = $2,
       updated_at = $3
 WHERE import_session_id = $1
   AND session_status NOT IN ('applying', 'applied', 'partially_applied', 'failed', 'canceled')
`, sessionID, status, now.UTC())
	return err
}

func sessionStatusAndSelectionTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, explicit *[]uuid.UUID) (string, uuid.UUID, []uuid.UUID, error) {
	status, incidentID, err := sessionStatusTx(ctx, tx, sessionID)
	if err != nil {
		return "", uuid.UUID{}, nil, err
	}
	_, selected, err := sessionApplySelectionTx(ctx, tx, sessionID, explicit)
	return status, incidentID, selected, err
}

func sessionApplySelectionTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, explicit *[]uuid.UUID) (uuid.UUID, []uuid.UUID, error) {
	var incidentID uuid.UUID
	var selectedJSON []byte
	if err := tx.QueryRow(ctx, `
SELECT incident_id, to_jsonb(selected_unit_ids)
  FROM import_sessions
 WHERE import_session_id = $1
`, sessionID).Scan(&incidentID, &selectedJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, nil, ErrNotFound
		}
		return uuid.UUID{}, nil, err
	}
	if explicit != nil {
		return incidentID, append([]uuid.UUID(nil), (*explicit)...), nil
	}
	var selectedText []string
	if err := json.Unmarshal(selectedJSON, &selectedText); err != nil {
		return uuid.UUID{}, nil, err
	}
	selected := make([]uuid.UUID, 0, len(selectedText))
	for _, value := range selectedText {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return uuid.UUID{}, nil, err
		}
		selected = append(selected, parsed)
	}
	return incidentID, selected, nil
}

func requireSelectedUnitsReadyTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, selected []uuid.UUID) error {
	rows, err := tx.Query(ctx, `
SELECT import_unit_id, unit_status, approved_mapping_json
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = ANY($2)
`, sessionID, selected)
	if err != nil {
		return err
	}
	defer rows.Close()
	seen := map[uuid.UUID]struct{}{}
	for rows.Next() {
		var id uuid.UUID
		var status string
		var mapping []byte
		if err := rows.Scan(&id, &status, &mapping); err != nil {
			return err
		}
		seen[id] = struct{}{}
		if status != "ready" || len(mapping) == 0 {
			return importApplyBlockedError("unit_not_ready")
		}
		var approved ApprovedMapping
		if err := json.Unmarshal(mapping, &approved); err != nil {
			return err
		}
		target, ok := lookupApprovedImportTarget(approved)
		if !ok || !target.readyCheckImportable() {
			if approved.targetKindOrDefault() == ImportTargetKindViewSchema {
				return importApplyBlockedError("target_view_schema_not_importable")
			}
			return importApplyBlockedError("target_kind_not_importable")
		}
		if approved.targetKindOrDefault() == ImportTargetKindViewSchema && !target.ownerCreateFacadeAvailable() {
			return importApplyBlockedError("owner_create_contract_unavailable")
		}
		if approved.targetKindOrDefault() != ImportTargetKindViewSchema && !target.ownerApplyFacadeAvailable() {
			return importApplyBlockedError("owner_apply_contract_unavailable")
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(seen) != len(selected) {
		return importApplyBlockedError("unit_not_ready")
	}
	return nil
}

func importConflictError(reason string) error {
	return &StateConflictError{ReasonCode: reason}
}

func importApplyBlockedError(reason string) error {
	return &ApplyBlockedError{ReasonCode: reason}
}

func intPtr(value int) *int {
	return &value
}

func scanSessionResource(row pgx.Row) (map[string]any, uuid.UUID, error) {
	var sessionID uuid.UUID
	var incidentID uuid.UUID
	var createdBy uuid.UUID
	var clientTxnID string
	var assistantProfile string
	var sourceFileKind string
	var originalFilename string
	var sourceSHA string
	var parserProfileID string
	var parserVersion string
	var status string
	var discoveryJobID *uuid.UUID
	var applyJobID *uuid.UUID
	var selectedJSON []byte
	var blockingDiagnosticsJSON []byte
	var warningJSON []byte
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&sessionID, &incidentID, &createdBy, &clientTxnID, &assistantProfile, &sourceFileKind, &originalFilename, &sourceSHA, &parserProfileID, &parserVersion, &status, &discoveryJobID, &applyJobID, &selectedJSON, &blockingDiagnosticsJSON, &warningJSON, &createdAt, &updatedAt); err != nil {
		return nil, uuid.UUID{}, err
	}
	var selected []string
	if len(selectedJSON) > 0 {
		_ = json.Unmarshal(selectedJSON, &selected)
	}
	if selected == nil {
		selected = []string{}
	}
	var blockingDiagnostics []map[string]any
	if len(blockingDiagnosticsJSON) > 0 {
		_ = json.Unmarshal(blockingDiagnosticsJSON, &blockingDiagnostics)
	}
	if blockingDiagnostics == nil {
		blockingDiagnostics = []map[string]any{}
	}
	var warnings []string
	if len(warningJSON) > 0 {
		_ = json.Unmarshal(warningJSON, &warnings)
	}
	if warnings == nil {
		warnings = []string{}
	}
	resource := map[string]any{
		"import_session_id":         sessionID.String(),
		"incident_id":               incidentID.String(),
		"created_by_user_id":        createdBy.String(),
		"created_at":                createdAt,
		"source_file_kind":          sourceFileKind,
		"original_filename":         originalFilename,
		"source_content_sha256":     sourceSHA,
		"parser_profile_id":         parserProfileID,
		"parser_version":            parserVersion,
		"assistant_profile":         assistantProfile,
		"session_status":            status,
		"selected_unit_ids":         selected,
		"blocking_diagnostics":      blockingDiagnostics,
		"nonblocking_warning_codes": warnings,
	}
	_ = clientTxnID
	_ = discoveryJobID
	_ = applyJobID
	return resource, incidentID, nil
}

func scanUnitResource(row pgx.Row) (map[string]any, error) {
	var unitID uuid.UUID
	var sessionID uuid.UUID
	var status string
	var locatorKind string
	var locator string
	var sourceRect string
	var headerRow int
	var dataStart int
	var rowCount int
	var columnCount int
	var warningJSON []byte
	var mappingFingerprint *string
	var approvedMapping []byte
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&unitID, &sessionID, &status, &locatorKind, &locator, &sourceRect, &headerRow, &dataStart, &rowCount, &columnCount, &warningJSON, &mappingFingerprint, &approvedMapping, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var warnings []string
	if len(warningJSON) > 0 {
		_ = json.Unmarshal(warningJSON, &warnings)
	}
	if warnings == nil {
		warnings = []string{}
	}
	var locatorResource any = locator
	if locatorKind == "csv_file" {
		locatorResource = map[string]any{"file": "source"}
	} else if strings.HasPrefix(locator, "{") {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(locator), &decoded); err == nil {
			locatorResource = decoded
		}
	}
	resource := map[string]any{
		"import_unit_id":        unitID.String(),
		"import_session_id":     sessionID.String(),
		"unit_status":           status,
		"locator_kind":          locatorKind,
		"locator":               locatorResource,
		"source_rect_a1":        sourceRect,
		"header_row_ref":        headerRow,
		"data_start_row_ref":    dataStart,
		"inferred_row_count":    rowCount,
		"inferred_column_count": columnCount,
		"warning_codes":         warnings,
	}
	_ = createdAt
	_ = updatedAt
	if mappingFingerprint != nil {
		resource["mapping_fingerprint"] = *mappingFingerprint
	}
	if len(approvedMapping) > 0 {
		resource["approved_mapping"] = json.RawMessage(approvedMapping)
	}
	return resource, nil
}

func SourceRect(rowCount int, columnCount int) string {
	if rowCount < 1 {
		rowCount = 1
	}
	if columnCount < 1 {
		columnCount = 1
	}
	return fmt.Sprintf("A1:%s%d", columnName(columnCount), rowCount)
}

func previewTruncated(unit map[string]any, previewRows any) bool {
	rowCount, ok := intFromAny(unit["inferred_row_count"])
	if !ok {
		return false
	}
	switch rows := previewRows.(type) {
	case []any:
		return rowCount > len(rows)
	case []map[string]any:
		return rowCount > len(rows)
	default:
		return false
	}
}

func intFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float64:
		if typed == float64(int(typed)) {
			return int(typed), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func columnName(index int) string {
	name := ""
	for index > 0 {
		index--
		name = string(rune('A'+(index%26))) + name
		index /= 26
	}
	return name
}
