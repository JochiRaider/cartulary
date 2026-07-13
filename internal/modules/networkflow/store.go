package networkflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool           postgres.DB
	limits         Limits
	incidentLocks  IncidentLockPort
	auditAppender  AdministrativeAuditPort
	indicators     IndicatorParticipationPort
	transactionRun postgres.TransactionRunner
	safeDigester   SafeDigester
}

type StoreOption func(*Store)

func WithLimits(limits Limits) StoreOption {
	return func(s *Store) {
		s.limits = limits.normalized()
	}
}

func WithOwnerParticipants(incidentLocks IncidentLockPort, auditAppender AdministrativeAuditPort, indicatorParticipant IndicatorParticipationPort) StoreOption {
	return func(s *Store) {
		s.incidentLocks = incidentLocks
		s.auditAppender = auditAppender
		s.indicators = indicatorParticipant
	}
}

func WithTransactionRunner(runner postgres.TransactionRunner) StoreOption {
	return func(s *Store) { s.transactionRun = runner }
}

func WithSafeDigester(digester SafeDigester) StoreOption {
	return func(s *Store) { s.safeDigester = digester }
}

type TableRecord struct {
	TableID                   string
	IncidentID                uuid.UUID
	DisplayName               string
	TableVersion              int64
	TableStatus               string
	SourceImportSessionID     uuid.UUID
	SourceImportUnitID        uuid.UUID
	SourceContentSHA256       string
	SourceFilenameDisplay     string
	SourceFilenameDigest      string
	SourceFilenameDigestKeyID string
	MappingFingerprint        string
	SourceProfileID           string
	ParserProfileID           string
	RowCountAccepted          int64
	RowCountRejected          int64
	DiagnosticsTruncated      bool
	CreatedByUserID           uuid.UUID
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	DeletedAt                 *time.Time
}

type FlowRow struct {
	RowID                     string
	NetworkFlowTableID        string
	IncidentID                uuid.UUID
	SourceRowNumber           int64
	SourceRowDigestSHA256     string
	NormalizedRowDigestSHA256 string
	MappingFingerprint        string
	FlowStartUTC              time.Time
	FlowEndUTC                time.Time
	SrcIP                     string
	DstIP                     string
	SrcPort                   *int32
	DstPort                   *int32
	IPProtocol                int32
	BytesCount                string
	PacketsCount              string
	ExporterID                *string
	InputInterface            *string
	OutputInterface           *string
	TCPFlags                  *int32
	ApplicationLabel          *string
	UnmappedRaw               json.RawMessage
	ObservationSourceRef      json.RawMessage
	CreatedAt                 time.Time
	CreatedByUserID           uuid.UUID
}

type RejectedRowDiagnostic struct {
	DiagnosticID        string
	SourceRowNumber     int64
	SourceColumnOrdinal *int64
	RawHeaderSHA256     *string
	FieldKey            *string
	ErrorCode           string
	ReasonCode          string
	SafeSample          *string
	RawValueSHA256      *string
	MessageKey          string
	MessageArgs         json.RawMessage
	Message             string
	LimitName           *string
	LimitValue          *int64
	ActualValue         *int64
}

type CreateTableParams struct {
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
	Rows                      []FlowRow
	Diagnostics               []RejectedRowDiagnostic
	DiagnosticsTruncated      bool
	Now                       time.Time
}

type RenameTableParams struct {
	IncidentID       uuid.UUID
	ActorUserID      uuid.UUID
	TableID          string
	BaseTableVersion int64
	DisplayName      string
	ClientTxnID      string
	RequestID        string
	SafeDigester     SafeDigester
	Now              time.Time
}

type SoftDeleteTableParams struct {
	IncidentID       uuid.UUID
	ActorUserID      uuid.UUID
	TableID          string
	BaseTableVersion int64
	ClientTxnID      string
	RequestID        string
	Now              time.Time
}

type RetainedCounts struct {
	Active   int64
	Retained int64
}

func NewStore(pool postgres.DB, options ...StoreOption) *Store {
	store := &Store{
		pool:           pool,
		limits:         DefaultLimits(),
		transactionRun: postgres.NewTransactionRunner(pool),
	}
	for _, option := range options {
		option(store)
	}
	store.limits = store.limits.normalized()
	return store
}

func (s *Store) CreateTable(ctx context.Context, params CreateTableParams) (TableRecord, error) {
	var table TableRecord
	err := s.transactionRun.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.CreateTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return TableRecord{}, err
	}
	return table, nil
}

func (s *Store) CreateTableTx(ctx context.Context, tx pgx.Tx, params CreateTableParams) (TableRecord, error) {
	if len(params.Rows) == 0 {
		return TableRecord{}, ErrNoAcceptedRows
	}
	now := normalizedNow(params.Now)
	if err := s.lockIncidentTx(ctx, tx, params.IncidentID); err != nil {
		return TableRecord{}, err
	}
	counts, err := retainedCountsTx(ctx, tx, params.IncidentID)
	if err != nil {
		return TableRecord{}, err
	}
	if counts.Active >= s.limits.MaxActiveTablesPerIncident {
		return TableRecord{}, &TableLimitError{IncidentID: params.IncidentID, LimitName: "network_flow.max_active_tables_per_incident", Limit: s.limits.MaxActiveTablesPerIncident, Current: counts.Active}
	}
	if counts.Retained >= s.limits.MaxRetainedTablesPerIncident {
		return TableRecord{}, &TableLimitError{IncidentID: params.IncidentID, LimitName: "network_flow.max_retained_tables_per_incident", Limit: s.limits.MaxRetainedTablesPerIncident, Current: counts.Retained}
	}
	existingNames, err := activeDisplayNamesTx(ctx, tx, params.IncidentID, "")
	if err != nil {
		return TableRecord{}, err
	}
	displayName, err := finalDisplayName(params, existingNames)
	if err != nil {
		return TableRecord{}, err
	}
	parserProfileID := params.ParserProfileID
	if parserProfileID == "" {
		parserProfileID = ParserProfileRFC4180HeaderedCSV
	}
	sourceFilenameDisplay := SanitizeSourceFilenameDisplay(params.OriginalFilename)
	tableID, table, err := s.insertTableWithGeneratedID(ctx, tx, insertTableParams{
		IncidentID:                params.IncidentID,
		DisplayName:               displayName,
		SourceImportSessionID:     params.ImportSessionID,
		SourceImportUnitID:        params.ImportUnitID,
		SourceContentSHA256:       params.SourceContentSHA256,
		SourceFilenameDisplay:     sourceFilenameDisplay,
		SourceFilenameDigest:      params.SourceFilenameDigest,
		SourceFilenameDigestKeyID: params.SourceFilenameDigestKeyID,
		MappingFingerprint:        params.MappingFingerprint,
		SourceProfileID:           params.SourceProfileID,
		ParserProfileID:           parserProfileID,
		RowCountAccepted:          int64(len(params.Rows)),
		RowCountRejected:          int64(len(params.Diagnostics)),
		DiagnosticsTruncated:      params.DiagnosticsTruncated,
		CreatedByUserID:           params.ActorUserID,
		CreatedAt:                 now,
	})
	if err != nil {
		return TableRecord{}, err
	}
	if err := insertRowsTx(ctx, tx, tableID, params, parserProfileID, now); err != nil {
		return TableRecord{}, err
	}
	if err := insertDiagnosticsTx(ctx, tx, tableID, params, now); err != nil {
		return TableRecord{}, err
	}
	if params.ClientTxnID != "" && params.ActorUserID != uuid.Nil {
		if err := s.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
			ActorUserID: &params.ActorUserID,
			IncidentID:  &params.IncidentID,
			EventKind:   "network_flow_table_created",
			ClientTxnID: optionalStringPtr(params.ClientTxnID),
			RequestID:   optionalStringPtr(params.RequestID),
			AfterJSON: map[string]any{
				"incident_id":                    params.IncidentID.String(),
				"actor_user_id":                  params.ActorUserID.String(),
				"network_flow_table_id":          table.TableID,
				"source_filename_digest":         table.SourceFilenameDigest,
				"source_filename_digest_key_id":  table.SourceFilenameDigestKeyID,
				"source_content_sha256":          table.SourceContentSHA256,
				"source_profile_id":              table.SourceProfileID,
				"parser_profile_id":              table.ParserProfileID,
				"mapping_fingerprint":            table.MappingFingerprint,
				"row_count_accepted":             table.RowCountAccepted,
				"row_count_rejected":             table.RowCountRejected,
				"network_flow.audit_event_code":  "network_flow_table_created",
				"network_flow.audit_resource_id": table.TableID,
			},
		}); err != nil {
			return TableRecord{}, err
		}
	}
	return table, nil
}

func (s *Store) ListActiveTables(ctx context.Context, incidentID uuid.UUID) ([]TableRecord, error) {
	rows, err := s.pool.Query(ctx, tableSelectColumns()+`
  FROM network_flow_tables
 WHERE incident_id = $1
   AND table_status = 'active'
 ORDER BY created_at ASC, network_flow_table_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list active network flow tables: %w", err)
	}
	defer rows.Close()
	return scanTables(rows)
}

func (s *Store) GetActiveTable(ctx context.Context, incidentID uuid.UUID, tableID string) (TableRecord, error) {
	table, err := getTable(ctx, s.pool, incidentID, tableID, false)
	if err != nil {
		return TableRecord{}, err
	}
	if table.TableStatus != TableStatusActive {
		return TableRecord{}, ErrTableNotActive
	}
	return table, nil
}

func (s *Store) GetTable(ctx context.Context, incidentID uuid.UUID, tableID string) (TableRecord, error) {
	return getTable(ctx, s.pool, incidentID, tableID, false)
}

func (s *Store) RetainedCounts(ctx context.Context, incidentID uuid.UUID) (RetainedCounts, error) {
	return retainedCountsTx(ctx, s.pool, incidentID)
}

func (s *Store) RenameTable(ctx context.Context, params RenameTableParams) (TableRecord, error) {
	var table TableRecord
	err := s.transactionRun.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.renameTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return TableRecord{}, err
	}
	return table, nil
}

func (s *Store) renameTableTx(ctx context.Context, tx pgx.Tx, params RenameTableParams) (TableRecord, error) {
	now := normalizedNow(params.Now)
	if err := s.lockIncidentTx(ctx, tx, params.IncidentID); err != nil {
		return TableRecord{}, err
	}
	table, err := getTable(ctx, tx, params.IncidentID, params.TableID, true)
	if err != nil {
		return TableRecord{}, err
	}
	if table.TableStatus != TableStatusActive {
		return TableRecord{}, ErrTableNotActive
	}
	if table.TableVersion != params.BaseTableVersion {
		return TableRecord{}, &TableVersionConflictError{TableID: params.TableID, BaseTableVersion: params.BaseTableVersion, CurrentTableVersion: table.TableVersion}
	}
	displayName, err := normalizeExplicitDisplayName(params.DisplayName)
	if err != nil {
		return TableRecord{}, err
	}
	if displayName == table.DisplayName {
		return table, nil
	}
	existingNames, err := activeDisplayNamesTx(ctx, tx, params.IncidentID, params.TableID)
	if err != nil {
		return TableRecord{}, err
	}
	if _, exists := existingNames[displayName]; exists {
		return TableRecord{}, &InvalidDisplayNameError{ReasonCode: "duplicate_display_name"}
	}
	updated, err := updateTableNameTx(ctx, tx, params.IncidentID, params.TableID, displayName, now)
	if err != nil {
		return TableRecord{}, err
	}
	if params.ClientTxnID != "" && params.ActorUserID != uuid.Nil {
		digester := params.SafeDigester
		if digester == nil {
			digester = s.safeDigester
		}
		if digester == nil {
			return TableRecord{}, fmt.Errorf("network flow safe digester unavailable")
		}
		oldDigest, keyID, err := digester.Digest("table_display_name", table.DisplayName)
		if err != nil {
			return TableRecord{}, err
		}
		newDigest, newKeyID, err := digester.Digest("table_display_name", updated.DisplayName)
		if err != nil {
			return TableRecord{}, err
		}
		if newKeyID != keyID {
			return TableRecord{}, fmt.Errorf("network flow safe-digest epoch changed during table rename")
		}
		if err := s.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
			ActorUserID: &params.ActorUserID,
			IncidentID:  &params.IncidentID,
			EventKind:   "network_flow_table_renamed",
			ClientTxnID: optionalStringPtr(params.ClientTxnID),
			RequestID:   optionalStringPtr(params.RequestID),
			BeforeJSON: map[string]any{
				"incident_id":                    params.IncidentID.String(),
				"actor_user_id":                  params.ActorUserID.String(),
				"network_flow_table_id":          table.TableID,
				"display_name_digest":            oldDigest,
				"display_name_digest_key_id":     keyID,
				"table_version":                  table.TableVersion,
				"network_flow.audit_event_code":  "network_flow_table_renamed",
				"network_flow.audit_resource_id": table.TableID,
			},
			AfterJSON: map[string]any{
				"incident_id":                    params.IncidentID.String(),
				"actor_user_id":                  params.ActorUserID.String(),
				"network_flow_table_id":          updated.TableID,
				"old_display_name_digest":        oldDigest,
				"new_display_name_digest":        newDigest,
				"display_name_digest_key_id":     keyID,
				"table_version":                  updated.TableVersion,
				"network_flow.audit_event_code":  "network_flow_table_renamed",
				"network_flow.audit_resource_id": updated.TableID,
			},
		}); err != nil {
			return TableRecord{}, err
		}
	}
	return updated, nil
}

func (s *Store) SoftDeleteTable(ctx context.Context, params SoftDeleteTableParams) (TableRecord, error) {
	var table TableRecord
	err := s.transactionRun.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.softDeleteTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return TableRecord{}, err
	}
	return table, nil
}

func (s *Store) softDeleteTableTx(ctx context.Context, tx pgx.Tx, params SoftDeleteTableParams) (TableRecord, error) {
	now := normalizedNow(params.Now)
	if err := s.lockIncidentTx(ctx, tx, params.IncidentID); err != nil {
		return TableRecord{}, err
	}
	table, err := getTable(ctx, tx, params.IncidentID, params.TableID, true)
	if err != nil {
		return TableRecord{}, err
	}
	if table.TableStatus != TableStatusActive {
		return TableRecord{}, ErrTableNotActive
	}
	if table.TableVersion != params.BaseTableVersion {
		return TableRecord{}, &TableVersionConflictError{TableID: params.TableID, BaseTableVersion: params.BaseTableVersion, CurrentTableVersion: table.TableVersion}
	}
	deleted, err := updateTableSoftDeletedTx(ctx, tx, params.IncidentID, params.TableID, now)
	if err != nil {
		return TableRecord{}, err
	}
	if params.ClientTxnID != "" && params.ActorUserID != uuid.Nil {
		if err := s.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
			ActorUserID: &params.ActorUserID,
			IncidentID:  &params.IncidentID,
			EventKind:   "network_flow_table_soft_deleted",
			ClientTxnID: optionalStringPtr(params.ClientTxnID),
			RequestID:   optionalStringPtr(params.RequestID),
			BeforeJSON: map[string]any{
				"incident_id":                    params.IncidentID.String(),
				"actor_user_id":                  params.ActorUserID.String(),
				"network_flow_table_id":          table.TableID,
				"table_version":                  table.TableVersion,
				"table_status":                   table.TableStatus,
				"network_flow.audit_event_code":  "network_flow_table_soft_deleted",
				"network_flow.audit_resource_id": table.TableID,
			},
			AfterJSON: map[string]any{
				"incident_id":                    params.IncidentID.String(),
				"actor_user_id":                  params.ActorUserID.String(),
				"network_flow_table_id":          deleted.TableID,
				"table_version":                  deleted.TableVersion,
				"row_count_accepted":             deleted.RowCountAccepted,
				"row_count_rejected":             deleted.RowCountRejected,
				"network_flow.audit_event_code":  "network_flow_table_soft_deleted",
				"network_flow.audit_resource_id": deleted.TableID,
			},
		}); err != nil {
			return TableRecord{}, err
		}
	}
	return deleted, nil
}

func (s *Store) ListRows(ctx context.Context, incidentID uuid.UUID, tableID string) ([]FlowRow, error) {
	if _, err := s.GetActiveTable(ctx, incidentID, tableID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
SELECT `+flowRowColumnList()+`
  FROM network_flow_rows
 WHERE incident_id = $1
   AND network_flow_table_id = $2
 ORDER BY source_row_number ASC, network_flow_row_id ASC
`, incidentID, tableID)
	if err != nil {
		return nil, fmt.Errorf("list network flow rows: %w", err)
	}
	defer rows.Close()
	records := []FlowRow{}
	for rows.Next() {
		record, err := scanFlowRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan network flow rows: %w", err)
	}
	return records, nil
}

func (s *Store) ListRowsForTables(ctx context.Context, incidentID uuid.UUID, tableIDs []string) ([]FlowRow, error) {
	if len(tableIDs) == 0 {
		return []FlowRow{}, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT `+flowRowColumnList()+`
  FROM network_flow_rows
 WHERE incident_id = $1
   AND network_flow_table_id = ANY($2::text[])
 ORDER BY source_row_number ASC, network_flow_table_id ASC, network_flow_row_id ASC
`, incidentID, tableIDs)
	if err != nil {
		return nil, fmt.Errorf("list network flow rows for tables: %w", err)
	}
	defer rows.Close()
	records := []FlowRow{}
	for rows.Next() {
		record, err := scanFlowRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan network flow table rows: %w", err)
	}
	return records, nil
}

func (s *Store) ListRejectedRowDiagnostics(ctx context.Context, incidentID uuid.UUID, tableID string) ([]RejectedRowDiagnostic, error) {
	if _, err := s.GetActiveTable(ctx, incidentID, tableID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
SELECT diagnostic_id, source_row_number, source_column_ordinal, raw_header_sha256,
       field_key, error_code, reason_code, safe_sample, raw_value_sha256, message_key,
       message_args, message, limit_name, limit_value, actual_value
  FROM network_flow_rejected_row_diagnostics
 WHERE incident_id = $1
   AND network_flow_table_id = $2
 ORDER BY source_row_number ASC, source_column_ordinal ASC NULLS LAST,
          field_key ASC NULLS LAST, error_code ASC, diagnostic_id ASC
`, incidentID, tableID)
	if err != nil {
		return nil, fmt.Errorf("list network flow rejected-row diagnostics: %w", err)
	}
	defer rows.Close()
	diagnostics := []RejectedRowDiagnostic{}
	for rows.Next() {
		diagnostic, err := scanRejectedRowDiagnostic(rows)
		if err != nil {
			return nil, err
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan network flow rejected-row diagnostics: %w", err)
	}
	return diagnostics, nil
}

type insertTableParams struct {
	IncidentID                uuid.UUID
	DisplayName               string
	SourceImportSessionID     uuid.UUID
	SourceImportUnitID        uuid.UUID
	SourceContentSHA256       string
	SourceFilenameDisplay     string
	SourceFilenameDigest      string
	SourceFilenameDigestKeyID string
	MappingFingerprint        string
	SourceProfileID           string
	ParserProfileID           string
	RowCountAccepted          int64
	RowCountRejected          int64
	DiagnosticsTruncated      bool
	CreatedByUserID           uuid.UUID
	CreatedAt                 time.Time
}

func (s *Store) insertTableWithGeneratedID(ctx context.Context, tx pgx.Tx, params insertTableParams) (string, TableRecord, error) {
	for attempt := 0; attempt < 8; attempt++ {
		tableID, err := newTableID()
		if err != nil {
			return "", TableRecord{}, err
		}
		table, err := insertTableTx(ctx, tx, tableID, params)
		if isUniqueViolationOnConstraint(err, "network_flow_tables_pkey") {
			continue
		}
		if err != nil {
			return "", TableRecord{}, err
		}
		return tableID, table, nil
	}
	return "", TableRecord{}, ErrIDGenerationFailed
}

func insertTableTx(ctx context.Context, tx pgx.Tx, tableID string, params insertTableParams) (TableRecord, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO network_flow_tables (
    network_flow_table_id, incident_id, display_name, table_version, table_status,
    source_import_session_id, source_import_unit_id, source_content_sha256,
    source_filename_display, source_filename_digest, source_filename_digest_key_id,
    mapping_fingerprint, source_profile_id, parser_profile_id, row_count_accepted,
    row_count_rejected, diagnostics_truncated, created_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, 1, 'active', $4, $5, $6, $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17, $17
)
RETURNING `+tableColumnList(), tableID, params.IncidentID, params.DisplayName, params.SourceImportSessionID, params.SourceImportUnitID, params.SourceContentSHA256, params.SourceFilenameDisplay, params.SourceFilenameDigest, params.SourceFilenameDigestKeyID, params.MappingFingerprint, params.SourceProfileID, params.ParserProfileID, params.RowCountAccepted, params.RowCountRejected, params.DiagnosticsTruncated, params.CreatedByUserID, params.CreatedAt)
	table, err := scanTable(row)
	if err != nil {
		return TableRecord{}, fmt.Errorf("insert network flow table: %w", err)
	}
	return table, nil
}

func updateTableNameTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, tableID string, displayName string, now time.Time) (TableRecord, error) {
	row := tx.QueryRow(ctx, `
UPDATE network_flow_tables
   SET display_name = $3,
       table_version = table_version + 1,
       updated_at = $4
 WHERE incident_id = $1
   AND network_flow_table_id = $2
RETURNING `+tableColumnList(), incidentID, tableID, displayName, now)
	table, err := scanTable(row)
	if err != nil {
		return TableRecord{}, fmt.Errorf("rename network flow table: %w", err)
	}
	return table, nil
}

func updateTableSoftDeletedTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, tableID string, now time.Time) (TableRecord, error) {
	row := tx.QueryRow(ctx, `
UPDATE network_flow_tables
   SET table_status = 'soft_deleted',
       table_version = table_version + 1,
       updated_at = $3,
       deleted_at = $3
 WHERE incident_id = $1
   AND network_flow_table_id = $2
RETURNING `+tableColumnList(), incidentID, tableID, now)
	table, err := scanTable(row)
	if err != nil {
		return TableRecord{}, fmt.Errorf("soft delete network flow table: %w", err)
	}
	return table, nil
}

func insertRowsTx(ctx context.Context, tx pgx.Tx, tableID string, params CreateTableParams, parserProfileID string, now time.Time) error {
	ordered := append([]FlowRow(nil), params.Rows...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].SourceRowNumber == ordered[j].SourceRowNumber {
			return ordered[i].RowID < ordered[j].RowID
		}
		return ordered[i].SourceRowNumber < ordered[j].SourceRowNumber
	})
	for _, source := range ordered {
		row := materializeRow(tableID, params, parserProfileID, source, now)
		if _, err := tx.Exec(ctx, `
INSERT INTO network_flow_rows (
    network_flow_row_id, network_flow_table_id, incident_id, source_row_number,
    source_row_digest_sha256, normalized_row_digest_sha256, mapping_fingerprint,
    flow_start_utc, flow_end_utc, src_ip, dst_ip, src_port, dst_port, ip_protocol,
    bytes_count, packets_count, exporter_id, input_interface, output_interface,
    tcp_flags, application_label, unmapped_raw, observation_source_ref, created_at, created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
    $18, $19, $20, $21, $22::jsonb, $23::jsonb, $24, $25
)
`, row.RowID, row.NetworkFlowTableID, row.IncidentID, row.SourceRowNumber, row.SourceRowDigestSHA256, row.NormalizedRowDigestSHA256, row.MappingFingerprint, row.FlowStartUTC, row.FlowEndUTC, row.SrcIP, row.DstIP, nullableInt32(row.SrcPort), nullableInt32(row.DstPort), row.IPProtocol, row.BytesCount, row.PacketsCount, row.ExporterID, row.InputInterface, row.OutputInterface, nullableInt32(row.TCPFlags), row.ApplicationLabel, string(row.UnmappedRaw), string(row.ObservationSourceRef), row.CreatedAt, row.CreatedByUserID); err != nil {
			return fmt.Errorf("insert network flow row %s: %w", row.RowID, err)
		}
	}
	return nil
}

func insertDiagnosticsTx(ctx context.Context, tx pgx.Tx, tableID string, params CreateTableParams, now time.Time) error {
	ordered := append([]RejectedRowDiagnostic(nil), params.Diagnostics...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return compareDiagnostics(ordered[i], ordered[j]) < 0
	})
	for _, diagnostic := range ordered {
		messageArgs := diagnostic.MessageArgs
		if len(messageArgs) == 0 {
			messageArgs = json.RawMessage(`{}`)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO network_flow_rejected_row_diagnostics (
    diagnostic_id, network_flow_table_id, incident_id, source_row_number,
    source_column_ordinal, raw_header_sha256, field_key, error_code, reason_code,
    safe_sample, raw_value_sha256, message_key, message_args, message,
    limit_name, limit_value, actual_value, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14, $15, $16, $17, $18
)
`, diagnostic.DiagnosticID, tableID, params.IncidentID, diagnostic.SourceRowNumber, diagnostic.SourceColumnOrdinal, diagnostic.RawHeaderSHA256, diagnostic.FieldKey, diagnostic.ErrorCode, diagnostic.ReasonCode, diagnostic.SafeSample, diagnostic.RawValueSHA256, diagnostic.MessageKey, string(messageArgs), diagnostic.Message, diagnostic.LimitName, diagnostic.LimitValue, diagnostic.ActualValue, now); err != nil {
			return fmt.Errorf("insert network flow rejected-row diagnostic %s: %w", diagnostic.DiagnosticID, err)
		}
	}
	return nil
}

func materializeRow(tableID string, params CreateTableParams, parserProfileID string, source FlowRow, now time.Time) FlowRow {
	row := source
	row.NetworkFlowTableID = tableID
	row.IncidentID = params.IncidentID
	row.CreatedAt = now
	row.CreatedByUserID = params.ActorUserID
	if row.MappingFingerprint == "" {
		row.MappingFingerprint = params.MappingFingerprint
	}
	if len(row.UnmappedRaw) == 0 {
		row.UnmappedRaw = json.RawMessage(`{}`)
	}
	if len(row.ObservationSourceRef) == 0 {
		row.ObservationSourceRef = observationSourceRef(params, parserProfileID, row)
	}
	if row.RowID == "" {
		row.RowID = RowID(params.IncidentID, tableID, row.SourceRowNumber, row.SourceRowDigestSHA256, row.NormalizedRowDigestSHA256)
	}
	return row
}

func observationSourceRef(params CreateTableParams, parserProfileID string, row FlowRow) json.RawMessage {
	payload := map[string]any{
		"import_session_id":        params.ImportSessionID.String(),
		"import_unit_id":           params.ImportUnitID.String(),
		"source_content_sha256":    params.SourceContentSHA256,
		"source_profile_id":        params.SourceProfileID,
		"parser_profile_id":        parserProfileID,
		"mapping_fingerprint":      params.MappingFingerprint,
		"source_row_number":        row.SourceRowNumber,
		"source_row_digest_sha256": row.SourceRowDigestSHA256,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

func finalDisplayName(params CreateTableParams, existingNames map[string]struct{}) (string, error) {
	if params.DisplayNameOverride == nil {
		return DeriveTableDisplayName(params.OriginalFilename, existingNames)
	}
	displayName, err := normalizeExplicitDisplayName(*params.DisplayNameOverride)
	if err != nil {
		return "", err
	}
	if _, exists := existingNames[displayName]; exists {
		return "", &InvalidDisplayNameError{ReasonCode: "duplicate_display_name"}
	}
	return displayName, nil
}

type tableQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type retainedCountQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func getTable(ctx context.Context, querier tableQuerier, incidentID uuid.UUID, tableID string, forUpdate bool) (TableRecord, error) {
	query := tableSelectColumns() + `
  FROM network_flow_tables
 WHERE incident_id = $1
   AND network_flow_table_id = $2`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	table, err := scanTable(querier.QueryRow(ctx, query, incidentID, tableID))
	if errors.Is(err, pgx.ErrNoRows) {
		return TableRecord{}, ErrTableNotFound
	}
	if err != nil {
		return TableRecord{}, fmt.Errorf("get network flow table: %w", err)
	}
	return table, nil
}

func (s *Store) lockIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if s.incidentLocks == nil {
		return fmt.Errorf("network flow incident lock participant unavailable")
	}
	found, err := s.incidentLocks.LockIncidentTx(ctx, tx, incidentID)
	if err != nil {
		return err
	}
	if !found {
		return ErrIncidentNotFound
	}
	return nil
}

func activeDisplayNamesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, excludedTableID string) (map[string]struct{}, error) {
	rows, err := tx.Query(ctx, `
SELECT display_name
  FROM network_flow_tables
 WHERE incident_id = $1
   AND table_status = 'active'
   AND ($2 = '' OR network_flow_table_id <> $2)
 ORDER BY display_name ASC
`, incidentID, excludedTableID)
	if err != nil {
		return nil, fmt.Errorf("load active network flow display names: %w", err)
	}
	defer rows.Close()
	names := map[string]struct{}{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan active network flow display name: %w", err)
		}
		names[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan active network flow display names: %w", err)
	}
	return names, nil
}

func retainedCountsTx(ctx context.Context, querier retainedCountQuerier, incidentID uuid.UUID) (RetainedCounts, error) {
	var counts RetainedCounts
	err := querier.QueryRow(ctx, `
SELECT COUNT(*) FILTER (WHERE table_status = 'active') AS active_count,
       COUNT(*) FILTER (WHERE table_status IN ('active', 'soft_deleted')) AS retained_count
  FROM network_flow_tables
 WHERE incident_id = $1
`, incidentID).Scan(&counts.Active, &counts.Retained)
	if err != nil {
		return RetainedCounts{}, fmt.Errorf("count network flow retained tables: %w", err)
	}
	return counts, nil
}

func newTableID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate network flow table id: %w", err)
	}
	return "nft_" + hex.EncodeToString(bytes[:]), nil
}

func normalizedNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func tableSelectColumns() string {
	return `SELECT ` + tableColumnList()
}

func tableColumnList() string {
	return `network_flow_table_id, incident_id, display_name, table_version, table_status,
       source_import_session_id, source_import_unit_id, source_content_sha256,
       source_filename_display, source_filename_digest, source_filename_digest_key_id,
       mapping_fingerprint, source_profile_id, parser_profile_id, row_count_accepted,
       row_count_rejected, diagnostics_truncated, created_by_user_id, created_at, updated_at, deleted_at`
}

func scanTable(row pgx.Row) (TableRecord, error) {
	var record TableRecord
	var deletedAt pgtype.Timestamptz
	if err := row.Scan(
		&record.TableID,
		&record.IncidentID,
		&record.DisplayName,
		&record.TableVersion,
		&record.TableStatus,
		&record.SourceImportSessionID,
		&record.SourceImportUnitID,
		&record.SourceContentSHA256,
		&record.SourceFilenameDisplay,
		&record.SourceFilenameDigest,
		&record.SourceFilenameDigestKeyID,
		&record.MappingFingerprint,
		&record.SourceProfileID,
		&record.ParserProfileID,
		&record.RowCountAccepted,
		&record.RowCountRejected,
		&record.DiagnosticsTruncated,
		&record.CreatedByUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&deletedAt,
	); err != nil {
		return TableRecord{}, err
	}
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	return record, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanFlowRow(row rowScanner) (FlowRow, error) {
	var record FlowRow
	var srcPort *int32
	var dstPort *int32
	var tcpFlags *int32
	if err := row.Scan(
		&record.RowID,
		&record.NetworkFlowTableID,
		&record.IncidentID,
		&record.SourceRowNumber,
		&record.SourceRowDigestSHA256,
		&record.NormalizedRowDigestSHA256,
		&record.MappingFingerprint,
		&record.FlowStartUTC,
		&record.FlowEndUTC,
		&record.SrcIP,
		&record.DstIP,
		&srcPort,
		&dstPort,
		&record.IPProtocol,
		&record.BytesCount,
		&record.PacketsCount,
		&record.ExporterID,
		&record.InputInterface,
		&record.OutputInterface,
		&tcpFlags,
		&record.ApplicationLabel,
		&record.UnmappedRaw,
		&record.ObservationSourceRef,
		&record.CreatedAt,
		&record.CreatedByUserID,
	); err != nil {
		return FlowRow{}, err
	}
	record.SrcPort = srcPort
	record.DstPort = dstPort
	record.TCPFlags = tcpFlags
	record.FlowStartUTC = record.FlowStartUTC.UTC()
	record.FlowEndUTC = record.FlowEndUTC.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}

func scanTables(rows pgx.Rows) ([]TableRecord, error) {
	records := []TableRecord{}
	for rows.Next() {
		record, err := scanTable(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan network flow tables: %w", err)
	}
	return records, nil
}

func scanRejectedRowDiagnostic(row rowScanner) (RejectedRowDiagnostic, error) {
	var record RejectedRowDiagnostic
	if err := row.Scan(
		&record.DiagnosticID,
		&record.SourceRowNumber,
		&record.SourceColumnOrdinal,
		&record.RawHeaderSHA256,
		&record.FieldKey,
		&record.ErrorCode,
		&record.ReasonCode,
		&record.SafeSample,
		&record.RawValueSHA256,
		&record.MessageKey,
		&record.MessageArgs,
		&record.Message,
		&record.LimitName,
		&record.LimitValue,
		&record.ActualValue,
	); err != nil {
		return RejectedRowDiagnostic{}, err
	}
	if len(record.MessageArgs) == 0 {
		record.MessageArgs = json.RawMessage(`{}`)
	}
	return record, nil
}

func nullableInt32(value *int32) any {
	if value == nil {
		return nil
	}
	return *value
}

type networkFlowAuditEvent struct {
	ActorUserID *uuid.UUID
	IncidentID  *uuid.UUID
	EventKind   string
	ClientTxnID *string
	RequestID   *string
	BeforeJSON  any
	AfterJSON   any
}

func (s *Store) appendAuditEventTx(ctx context.Context, tx pgx.Tx, event networkFlowAuditEvent) error {
	if s.auditAppender == nil {
		return fmt.Errorf("network flow administrative audit participant unavailable")
	}
	return s.auditAppender.AppendNetworkFlowEventTx(ctx, tx, event.ActorUserID, event.IncidentID, event.EventKind, event.ClientTxnID, event.RequestID, event.BeforeJSON, event.AfterJSON)
}

func optionalStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func isUniqueViolationOnConstraint(err error, constraintName string) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
}

func compareDiagnostics(a, b RejectedRowDiagnostic) int {
	if a.SourceRowNumber != b.SourceRowNumber {
		if a.SourceRowNumber < b.SourceRowNumber {
			return -1
		}
		return 1
	}
	if cmp := compareNullableInt64NullsLast(a.SourceColumnOrdinal, b.SourceColumnOrdinal); cmp != 0 {
		return cmp
	}
	if cmp := compareNullableStringNullsLast(a.FieldKey, b.FieldKey); cmp != 0 {
		return cmp
	}
	if a.ErrorCode != b.ErrorCode {
		if a.ErrorCode < b.ErrorCode {
			return -1
		}
		return 1
	}
	if a.ReasonCode != b.ReasonCode {
		if a.ReasonCode < b.ReasonCode {
			return -1
		}
		return 1
	}
	if a.DiagnosticID < b.DiagnosticID {
		return -1
	}
	if a.DiagnosticID > b.DiagnosticID {
		return 1
	}
	return 0
}

func compareNullableInt64NullsLast(a, b *int64) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	case *a < *b:
		return -1
	case *a > *b:
		return 1
	default:
		return 0
	}
}

func compareNullableStringNullsLast(a, b *string) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	case *a < *b:
		return -1
	case *a > *b:
		return 1
	default:
		return 0
	}
}
