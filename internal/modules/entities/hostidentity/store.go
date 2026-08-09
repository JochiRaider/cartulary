package hostidentity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrInvalidCreateRequest       = errors.New("entities: invalid create request")
	ErrHostIdentityRecordNotFound = errors.New("entities: host/identity record not found")
)

type Store struct {
	pool             postgres.DB
	authStore        *authn.Store
	incidentAccess   incidents.Access
	revisionAppender *revisions.Appender
	keepSaved        conflicts.IdempotencyPort
	ports            entityStorePorts
	projectionReader workbookprojection.Reader
}

type StoreOption func(*Store)

func WithProjectionReader(reader workbookprojection.Reader) StoreOption {
	return func(store *Store) {
		store.projectionReader = reader
	}
}

func NewStore(
	pool postgres.DB,
	appender *revisions.Appender,
	keepSaved conflicts.IdempotencyPort,
	projectionWriter workbookprojection.Writer,
	options ...StoreOption,
) *Store {
	if projectionWriter == nil {
		panic("compose entity store: workbook projection writer is required")
	}
	store := &Store{
		pool:             pool,
		authStore:        authn.NewStore(pool),
		incidentAccess:   incidents.NewAccess(pool),
		revisionAppender: appender,
		keepSaved:        keepSaved,
		ports:            newEntityStorePorts(pool, appender, projectionWriter),
	}
	for _, option := range options {
		if option != nil {
			option(store)
		}
	}
	return store
}

func (s *Store) QueryHostRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	page, err := s.QueryHostRowsPage(ctx, incidentID, query, querypage.Window{Limit: int(^uint(0)>>1) - 1})
	return page.Rows, err
}

func (s *Store) QueryHostRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	if s.pool == nil {
		return querypage.Result{}, fmt.Errorf("query host rows: store pool is nil")
	}
	if s.projectionReader == nil {
		return querypage.Result{}, fmt.Errorf("query host rows: projection reader is required")
	}
	projections, err := s.projectionReader.SelectHostQueryProjections(ctx, incidentID, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	recordIDs := make([]uuid.UUID, 0, len(projections))
	for _, projection := range projections {
		recordIDs = append(recordIDs, projection.RecordID)
	}
	recordsByID, err := loadHostSourceRecordsByID(ctx, s.pool, incidentID, recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	hydration, err := loadEntityRowHydrationByRecord(ctx, s.pool, incidentID, "host", recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	result := make([]map[string]any, 0, len(projections))
	for _, projection := range projections {
		record, ok := recordsByID[projection.RecordID]
		if !ok {
			return querypage.Result{}, fmt.Errorf("hydrate host projection %s: authoritative source row is missing", projection.RecordID)
		}
		applyHostProjection(&record, projection)
		applyHostRowHydration(&record, hydration)
		result = append(result, BuildHostRow(record))
	}
	return querypage.Finish(result, window.Limit), nil
}

func loadHostSourceRecordsByID(
	ctx context.Context,
	querier entityAliasQueryer,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]HostRecord, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID]HostRecord{}, nil
	}
	rows, err := querier.Query(ctx, `
SELECT
    h.record_id,
    h.incident_id,
    h.display_name,
    h.aad_device_id,
    h.fqdn,
    h.hostname,
    h.merged_into_record_id,
    h.entity_origin,
    h.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM hosts h
  JOIN records r ON r.record_id = h.record_id
 WHERE h.incident_id = $1
   AND h.record_id = ANY($2::uuid[])
   AND r.deleted_at IS NULL
`, incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load bounded host source rows: %w", err)
	}
	defer rows.Close()

	records := make(map[uuid.UUID]HostRecord, len(recordIDs))
	for rows.Next() {
		record, scanErr := scanHostSourceRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records[record.RecordID] = record
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bounded host source rows: %w", err)
	}
	return records, nil
}

func scanHostSourceRecord(scanner interface{ Scan(...any) error }) (HostRecord, error) {
	var (
		record      HostRecord
		aadDeviceID pgtype.Text
		fqdn        pgtype.Text
		hostname    pgtype.Text
		mergedInto  pgtype.UUID
		seedMention pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&aadDeviceID,
		&fqdn,
		&hostname,
		&mergedInto,
		&record.EntityOrigin,
		&seedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return HostRecord{}, fmt.Errorf("scan bounded host source row: %w", err)
	}
	record.AADDeviceID = textPointer(aadDeviceID)
	record.FQDN = textPointer(fqdn)
	record.Hostname = textPointer(hostname)
	record.MergedIntoRecordID = uuidPointerFromPG(mergedInto)
	record.SeedMentionID = uuidPointerFromPG(seedMention)
	return record, nil
}

func applyHostProjection(record *HostRecord, projection workbookprojection.HostQueryProjection) {
	record.HostState = projection.HostState
	record.LinkedEventCount = projection.LinkedEventCount
	record.EvidenceCount = projection.EvidenceCount
	record.Location = projection.Location
	record.OSPlatform = projection.OSPlatform
	record.BusinessOwner = projection.BusinessOwner
	record.Criticality = projection.Criticality
	record.ContainmentStatus = projection.ContainmentStatus
}

func (s *Store) QueryIdentityRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	page, err := s.QueryIdentityRowsPage(ctx, incidentID, query, querypage.Window{Limit: int(^uint(0)>>1) - 1})
	return page.Rows, err
}

func (s *Store) QueryIdentityRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	if s.pool == nil {
		return querypage.Result{}, fmt.Errorf("query identity rows: store pool is nil")
	}
	if s.projectionReader == nil {
		return querypage.Result{}, fmt.Errorf("query identity rows: projection reader is required")
	}
	projections, err := s.projectionReader.SelectIdentityQueryProjections(ctx, incidentID, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	recordIDs := make([]uuid.UUID, 0, len(projections))
	for _, projection := range projections {
		recordIDs = append(recordIDs, projection.RecordID)
	}
	recordsByID, err := loadIdentitySourceRecordsByID(ctx, s.pool, incidentID, recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	hydration, err := loadEntityRowHydrationByRecord(ctx, s.pool, incidentID, "identity", recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	result := make([]map[string]any, 0, len(projections))
	for _, projection := range projections {
		record, ok := recordsByID[projection.RecordID]
		if !ok {
			return querypage.Result{}, fmt.Errorf("hydrate identity projection %s: authoritative source row is missing", projection.RecordID)
		}
		applyIdentityProjection(&record, projection)
		applyIdentityRowHydration(&record, hydration)
		result = append(result, BuildIdentityRow(record))
	}
	return querypage.Finish(result, window.Limit), nil
}

func loadIdentitySourceRecordsByID(
	ctx context.Context,
	querier entityAliasQueryer,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]IdentityRecord, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID]IdentityRecord{}, nil
	}
	rows, err := querier.Query(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.display_name,
    i.aad_object_id,
    i.sid,
    i.upn,
    i.email::text,
    i.sam_account_name,
    i.merged_into_record_id,
    i.entity_origin,
    i.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM identities i
  JOIN records r ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND i.record_id = ANY($2::uuid[])
   AND r.deleted_at IS NULL
`, incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load bounded identity source rows: %w", err)
	}
	defer rows.Close()

	records := make(map[uuid.UUID]IdentityRecord, len(recordIDs))
	for rows.Next() {
		record, scanErr := scanIdentitySourceRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records[record.RecordID] = record
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bounded identity source rows: %w", err)
	}
	return records, nil
}

func scanIdentitySourceRecord(scanner interface{ Scan(...any) error }) (IdentityRecord, error) {
	var (
		record         IdentityRecord
		aadObjectID    pgtype.Text
		sid            pgtype.Text
		upn            pgtype.Text
		email          pgtype.Text
		samAccountName pgtype.Text
		mergedInto     pgtype.UUID
		seedMention    pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&aadObjectID,
		&sid,
		&upn,
		&email,
		&samAccountName,
		&mergedInto,
		&record.EntityOrigin,
		&seedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return IdentityRecord{}, fmt.Errorf("scan bounded identity source row: %w", err)
	}
	record.AADObjectID = textPointer(aadObjectID)
	record.SID = textPointer(sid)
	record.UPN = textPointer(upn)
	record.Email = textPointer(email)
	record.SamAccountName = textPointer(samAccountName)
	record.MergedIntoRecordID = uuidPointerFromPG(mergedInto)
	record.SeedMentionID = uuidPointerFromPG(seedMention)
	return record, nil
}

func applyIdentityProjection(record *IdentityRecord, projection workbookprojection.IdentityQueryProjection) {
	record.IdentityState = projection.IdentityState
	record.LinkedEventCount = projection.LinkedEventCount
	record.EvidenceCount = projection.EvidenceCount
	record.PrivilegeLevel = projection.PrivilegeLevel
	record.MFAState = projection.MFAState
	record.ResetStatus = projection.ResetStatus
}

func (s *Store) CreateHostRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + HostsViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    hostCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed host create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query host create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin host create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	record, beforeRow, operationKind, statusCode, err := s.upsertHostTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.ports.projections.RefreshHostTx(ctx, tx, record.RecordID); err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      hostCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	afterRow := BuildHostRow(record)
	var beforeVersionID *string
	if beforeRow != nil {
		beforeVersion := record.RowVersion
		if !reflect.DeepEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("host", record.RecordID, beforeVersion)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("host", record.RecordID, record.RowVersion)
	if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "host",
		TargetID:        record.RecordID.String(),
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.appendAliasCreateMutationsTx(ctx, tx, changeSetID, 2, record.AliasMutations); err != nil {
		return MutationResult{}, err
	}
	if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
		if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, entityRecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    record.RecordID,
			RowVersion:  record.RowVersion,
			BeforeValue: beforeRow,
			AfterValue:  afterRow,
		}); err != nil {
			return MutationResult{}, err
		}
	}

	payload := BuildMutationPayload(HostsViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, statusCode, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit host create transaction: %w", err)
	}

	return MutationResult{
		Payload:     payload,
		StatusCode:  statusCode,
		RecordID:    record.RecordID,
		ChangeSetID: changeSetID,
		RowVersion:  record.RowVersion,
	}, nil
}

func (s *Store) CreateIdentityRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + IdentitiesViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    identityCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed identity create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query identity create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin identity create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	record, beforeRow, operationKind, statusCode, err := s.upsertIdentityTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.ports.projections.RefreshIdentityTx(ctx, tx, record.RecordID); err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      identityCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	afterRow := BuildIdentityRow(record)
	var beforeVersionID *string
	if beforeRow != nil {
		beforeVersion := record.RowVersion
		if !reflect.DeepEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("identity", record.RecordID, beforeVersion)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("identity", record.RecordID, record.RowVersion)
	if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "identity",
		TargetID:        record.RecordID.String(),
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.appendAliasCreateMutationsTx(ctx, tx, changeSetID, 2, record.AliasMutations); err != nil {
		return MutationResult{}, err
	}
	if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
		if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, entityRecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    record.RecordID,
			RowVersion:  record.RowVersion,
			BeforeValue: beforeRow,
			AfterValue:  afterRow,
		}); err != nil {
			return MutationResult{}, err
		}
	}

	payload := BuildMutationPayload(IdentitiesViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, statusCode, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit identity create transaction: %w", err)
	}

	return MutationResult{
		Payload:     payload,
		StatusCode:  statusCode,
		RecordID:    record.RecordID,
		ChangeSetID: changeSetID,
		RowVersion:  record.RowVersion,
	}, nil
}

func (s *Store) appendAliasCreateMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequence int, aliases []AliasMutationValue) error {
	for index, alias := range aliases {
		if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    startSequence + index,
			TargetKind:    "entity_alias",
			TargetID:      "entity_alias:" + alias.EntityAliasID.String(),
			OperationKind: "create",
			AfterValue:    alias.MutationValue(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func insertHostTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO hosts (
    record_id,
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    location,
    os_platform,
    business_owner,
    criticality,
    containment_status,
    host_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16, $17, $17)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.Location, record.OSPlatform, record.BusinessOwner, record.Criticality, record.ContainmentStatus, record.HostState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE hosts
   SET display_name = $2,
       aad_device_id = $3,
       fqdn = $4,
       hostname = $5,
       location = $6,
       os_platform = $7,
       business_owner = $8,
       criticality = $9,
       containment_status = $10,
       host_state = $11,
       merged_into_record_id = $12,
       row_version = $13,
       updated_at = $14,
       updated_by_user_id = $15
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.Location, record.OSPlatform, record.BusinessOwner, record.Criticality, record.ContainmentStatus, record.HostState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update host: %w", err)
	}
	return nil
}

func UpdateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	return updateHostTx(ctx, tx, record)
}

func insertIdentityTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO identities (
    record_id,
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email,
    sam_account_name,
    privilege_level,
    mfa_state,
    reset_status,
    identity_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16, $17, $17)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.PrivilegeLevel, record.MFAState, record.ResetStatus, record.IdentityState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateIdentityTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE identities
   SET display_name = $2,
       aad_object_id = $3,
       sid = $4,
       upn = $5,
       email = $6,
       sam_account_name = $7,
       privilege_level = $8,
       mfa_state = $9,
       reset_status = $10,
       identity_state = $11,
       merged_into_record_id = $12,
       row_version = $13,
       updated_at = $14,
       updated_by_user_id = $15
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.PrivilegeLevel, record.MFAState, record.ResetStatus, record.IdentityState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update identity: %w", err)
	}
	return nil
}

func UpdateIdentityTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	return updateIdentityTx(ctx, tx, record)
}

func LoadHostByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (HostRecord, error) {
	record, err := scanHostRecord(tx.QueryRow(ctx, `
SELECT
    h.record_id,
    h.incident_id,
    h.display_name,
    h.aad_device_id,
    h.fqdn,
    h.hostname,
    h.location,
    h.os_platform,
    h.business_owner,
    h.criticality,
    h.containment_status,
    h.host_state,
    h.merged_into_record_id,
    h.entity_origin,
    h.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM hosts h
  JOIN records r
    ON r.record_id = h.record_id
 WHERE h.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return HostRecord{}, ErrHostIdentityRecordNotFound
	}
	if err != nil {
		return HostRecord{}, fmt.Errorf("load host by record id: %w", err)
	}
	return record, nil
}

func LoadIdentityByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IdentityRecord, error) {
	record, err := scanIdentityRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.display_name,
    i.aad_object_id,
    i.sid,
    i.upn,
    i.email::text,
    i.sam_account_name,
    i.privilege_level,
    i.mfa_state,
    i.reset_status,
    i.identity_state,
    i.merged_into_record_id,
    i.entity_origin,
    i.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM identities i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IdentityRecord{}, ErrHostIdentityRecordNotFound
	}
	if err != nil {
		return IdentityRecord{}, fmt.Errorf("load identity by record id: %w", err)
	}
	return record, nil
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func entityVersionID(prefix string, recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("%s:%s:%d", prefix, recordID.String(), rowVersion)
}

func optionalValue(values map[string]string, key string) *string {
	value, ok := values[key]
	if !ok {
		return nil
	}
	cloned := value
	return &cloned
}

type entityAliasQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type entityRowHydration struct {
	AliasesByRecord             map[uuid.UUID][]AliasValue
	ReusableIdentifiersByRecord map[uuid.UUID][]ReusableIdentifier
}

func loadEntityRowHydrationByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string, recordIDs []uuid.UUID) (entityRowHydration, error) {
	aliasesByRecord, err := loadEntityAliasesByRecord(ctx, querier, incidentID, entityType, recordIDs)
	if err != nil {
		return entityRowHydration{}, err
	}
	reusableIdentifiersByRecord, err := loadEntityReusableIdentifiersByRecord(ctx, querier, incidentID, entityType, recordIDs)
	if err != nil {
		return entityRowHydration{}, err
	}
	return entityRowHydration{
		AliasesByRecord:             aliasesByRecord,
		ReusableIdentifiersByRecord: reusableIdentifiersByRecord,
	}, nil
}

func applyHostRowHydration(record *HostRecord, hydration entityRowHydration) {
	record.SuggestionOnlyAliases = hydration.AliasesByRecord[record.RecordID]
	record.ReusableIdentifiers = hydration.ReusableIdentifiersByRecord[record.RecordID]
}

func applyIdentityRowHydration(record *IdentityRecord, hydration entityRowHydration) {
	record.SuggestionOnlyAliases = hydration.AliasesByRecord[record.RecordID]
	record.ReusableIdentifiers = hydration.ReusableIdentifiersByRecord[record.RecordID]
}

func hydrateHostRecordTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	aliases, err := loadEntityAliasesTx(ctx, tx, record.RecordID, "host")
	if err != nil {
		return err
	}
	reusableIdentifiers, err := loadEntityReusableIdentifiersTx(ctx, tx, record.RecordID, "host")
	if err != nil {
		return err
	}
	record.SuggestionOnlyAliases = aliases
	record.ReusableIdentifiers = reusableIdentifiers
	return nil
}

func hydrateIdentityRecordTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	aliases, err := loadEntityAliasesTx(ctx, tx, record.RecordID, "identity")
	if err != nil {
		return err
	}
	reusableIdentifiers, err := loadEntityReusableIdentifiersTx(ctx, tx, record.RecordID, "identity")
	if err != nil {
		return err
	}
	record.SuggestionOnlyAliases = aliases
	record.ReusableIdentifiers = reusableIdentifiers
	return nil
}

func loadEntityAliasesByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string, recordIDs []uuid.UUID) (map[uuid.UUID][]AliasValue, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID][]AliasValue{}, nil
	}
	args := []any{incidentID, entityType}
	rows, err := querier.Query(ctx, `
SELECT record_id, entity_alias_id, normalized_text::text
  FROM entity_aliases
 WHERE incident_id = $1
   AND entity_type = $2
   AND record_id IN (`+bindUUIDList(&args, recordIDs)+`)
   AND deleted_at IS NULL
 ORDER BY record_id ASC, normalized_text ASC, created_at ASC, entity_alias_id ASC
`, args...)
	if err != nil {
		return nil, fmt.Errorf("query entity aliases by record: %w", err)
	}
	defer rows.Close()

	aliasesByRecord := make(map[uuid.UUID][]AliasValue)
	for rows.Next() {
		var (
			recordID uuid.UUID
			alias    AliasValue
		)
		if err := rows.Scan(&recordID, &alias.EntityAliasID, &alias.AliasText); err != nil {
			return nil, fmt.Errorf("scan entity alias by record: %w", err)
		}
		aliasesByRecord[recordID] = append(aliasesByRecord[recordID], alias)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity aliases by record: %w", err)
	}

	return aliasesByRecord, nil
}

func loadEntityReusableIdentifiersByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string, recordIDs []uuid.UUID) (map[uuid.UUID][]ReusableIdentifier, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID][]ReusableIdentifier{}, nil
	}
	args := []any{incidentID, entityType}
	rows, err := querier.Query(ctx, `
SELECT
    record_id,
    entity_preserved_identifier_id,
    identifier_type,
    raw_value,
    normalized_value
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
   AND entity_type = $2
   AND record_id IN (`+bindUUIDList(&args, recordIDs)+`)
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
 ORDER BY record_id ASC, identifier_type ASC, normalized_value ASC, created_at ASC, entity_preserved_identifier_id ASC
`, args...)
	if err != nil {
		return nil, fmt.Errorf("query entity reusable identifiers by record: %w", err)
	}
	defer rows.Close()

	identifiersByRecord := make(map[uuid.UUID][]ReusableIdentifier)
	for rows.Next() {
		var (
			recordID uuid.UUID
			item     ReusableIdentifier
		)
		if err := rows.Scan(&recordID, &item.EntityPreservedIdentifierID, &item.IdentifierClass, &item.RawValue, &item.NormalizedValue); err != nil {
			return nil, fmt.Errorf("scan entity reusable identifier by record: %w", err)
		}
		identifiersByRecord[recordID] = append(identifiersByRecord[recordID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity reusable identifiers by record: %w", err)
	}
	return identifiersByRecord, nil
}

func bindUUIDList(args *[]any, recordIDs []uuid.UUID) string {
	placeholders := make([]string, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		*args = append(*args, recordID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(*args)))
	}
	return strings.Join(placeholders, ", ")
}

func loadEntityReusableIdentifiersTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]ReusableIdentifier, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_preserved_identifier_id,
    identifier_type,
    raw_value,
    normalized_value
  FROM entity_preserved_identifiers
 WHERE record_id = $1
   AND entity_type = $2
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
 ORDER BY identifier_type ASC, normalized_value ASC, created_at ASC, entity_preserved_identifier_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load entity reusable identifiers: %w", err)
	}
	defer rows.Close()

	identifiers := make([]ReusableIdentifier, 0)
	for rows.Next() {
		var item ReusableIdentifier
		if err := rows.Scan(&item.EntityPreservedIdentifierID, &item.IdentifierClass, &item.RawValue, &item.NormalizedValue); err != nil {
			return nil, fmt.Errorf("scan entity reusable identifier: %w", err)
		}
		identifiers = append(identifiers, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity reusable identifiers: %w", err)
	}
	return identifiers, nil
}
