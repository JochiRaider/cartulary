package entities

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var ErrInvalidCreateRequest = errors.New("entities: invalid create request")

type Store struct {
	pool            *pgxpool.Pool
	authStore       *authn.Store
	revisionsStore  *revisions.Store
	projectionStore *projections.Store
	linkStore       *links.Store
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:            pool,
		authStore:       authn.NewStore(pool),
		revisionsStore:  revisions.NewStore(),
		projectionStore: projections.NewStore(pool),
		linkStore:       links.NewStore(),
	}
}

func (s *Store) QueryHostRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("query host rows: store pool is nil")
	}

	sqlText, args, err := buildHostQuerySQL(incidentID, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query host rows: %w", err)
	}
	defer rows.Close()

	aliasesByRecord, err := loadEntityAliasesByRecord(ctx, s.pool, incidentID, "host")
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanHostQueryRecord(rows)
		if err != nil {
			return nil, err
		}
		record.SuggestionOnlyAliases = aliasesByRecord[record.RecordID]
		result = append(result, BuildHostRow(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate host rows: %w", err)
	}
	return result, nil
}

func (s *Store) QueryIdentityRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("query identity rows: store pool is nil")
	}

	sqlText, args, err := buildIdentityQuerySQL(incidentID, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query identity rows: %w", err)
	}
	defer rows.Close()

	aliasesByRecord, err := loadEntityAliasesByRecord(ctx, s.pool, incidentID, "identity")
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanIdentityQueryRecord(rows)
		if err != nil {
			return nil, err
		}
		record.SuggestionOnlyAliases = aliasesByRecord[record.RecordID]
		result = append(result, BuildIdentityRow(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity rows: %w", err)
	}
	return result, nil
}

func (s *Store) QueryIndicatorRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("query indicator rows: store pool is nil")
	}

	sqlText, args, err := buildIndicatorQuerySQL(incidentID, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query indicator rows: %w", err)
	}
	defer rows.Close()

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanIndicatorProjectionRecord(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, BuildIndicatorRow(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate indicator rows: %w", err)
	}
	return result, nil
}

var hostSortExpressions = map[string]string{
	"record_id":               "h.record_id",
	"host.display_name":       "p.display_name",
	"host.hostname":           "p.hostname",
	"host.host_state":         "p.host_state",
	"host.linked_event_count": "p.linked_event_count",
	"host.evidence_count":     "p.evidence_count",
	"host.location":           "p.location",
	"host.os_platform":        "p.os_platform",
	"host.business_owner":     "p.business_owner",
	"host.criticality":        "p.criticality",
	"host.containment_status": "p.containment_status",
	"host.edited_at":          "p.edited_at",
}

var identitySortExpressions = map[string]string{
	"record_id":                   "i.record_id",
	"identity.display_name":       "p.display_name",
	"identity.upn":                "p.upn",
	"identity.email":              "p.email",
	"identity.sam_account_name":   "p.sam_account_name",
	"identity.identity_state":     "p.identity_state",
	"identity.linked_event_count": "p.linked_event_count",
	"identity.evidence_count":     "p.evidence_count",
	"identity.privilege_level":    "p.privilege_level",
	"identity.mfa_state":          "p.mfa_state",
	"identity.reset_status":       "p.reset_status",
	"identity.edited_at":          "p.edited_at",
}

var indicatorSortExpressions = map[string]string{
	"record_id":                       "i.record_id",
	"indicator.indicator_type":        "i.indicator_type",
	"indicator.value_kind":            "i.value_kind",
	"indicator.display_value":         "i.display_value",
	"indicator.normalized_value":      "i.normalized_value",
	"indicator.defanged_value":        "i.defanged_value",
	"indicator.hash_algorithm":        "i.hash_algorithm",
	"indicator.hash_value":            "i.hash_value",
	"indicator.stix_pattern":          "i.stix_pattern",
	"indicator.first_observed_at":     "i.first_observed_at",
	"indicator.last_observed_at":      "i.last_observed_at",
	"indicator.observation_count":     "i.observation_count",
	"indicator.lifecycle_summary":     "i.lifecycle_summary",
	"indicator.supporting_link_count": "i.supporting_link_count",
}

func buildHostQuerySQL(incidentID uuid.UUID, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    h.record_id,
    h.incident_id,
    h.display_name,
    h.aad_device_id,
    h.fqdn,
    h.hostname,
    p.host_state,
    p.linked_event_count,
    p.evidence_count,
    p.location,
    p.os_platform,
    p.business_owner,
    p.criticality,
    p.containment_status,
    h.merged_into_record_id,
    h.entity_origin,
    h.seed_entity_mention_id,
    h.row_version,
    h.created_at,
    h.updated_at,
    h.created_by_user_id,
    h.updated_by_user_id
  FROM hosts h
  JOIN host_grid_projection p
    ON p.record_id = h.record_id
 WHERE h.incident_id = $1
   AND p.host_state IN ('stub', 'canonical')`)
	args := []any{incidentID}

	for _, filter := range query.Filters {
		switch filter.FieldKey {
		case "host.host_state":
			if err := appendQueryEqualityClause(&builder, &args, "p.host_state", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "host.location":
			if err := appendQueryEqualityClause(&builder, &args, "p.location", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "host.os_platform":
			if err := appendQueryEqualityClause(&builder, &args, "p.os_platform", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "host.business_owner":
			if err := appendQueryEqualityClause(&builder, &args, "p.business_owner", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "host.criticality":
			if err := appendQueryEqualityClause(&builder, &args, "p.criticality", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "host.containment_status":
			if err := appendQueryEqualityClause(&builder, &args, "p.containment_status", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		default:
			return "", nil, fmt.Errorf("host query filter field %q not mapped", filter.FieldKey)
		}
	}

	if err := appendOrderBy(&builder, query.Sort, hostSortExpressions); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func buildIdentityQuerySQL(incidentID uuid.UUID, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    i.record_id,
    i.incident_id,
    i.display_name,
    i.aad_object_id,
    i.sid,
    i.upn,
    i.email::text,
    i.sam_account_name,
    p.identity_state,
    p.linked_event_count,
    p.evidence_count,
    p.privilege_level,
    p.mfa_state,
    p.reset_status,
    i.merged_into_record_id,
    i.entity_origin,
    i.seed_entity_mention_id,
    i.row_version,
    i.created_at,
    i.updated_at,
    i.created_by_user_id,
    i.updated_by_user_id
  FROM identities i
  JOIN identity_grid_projection p
    ON p.record_id = i.record_id
 WHERE i.incident_id = $1
   AND p.identity_state IN ('stub', 'canonical')`)
	args := []any{incidentID}

	for _, filter := range query.Filters {
		switch filter.FieldKey {
		case "identity.identity_state":
			if err := appendQueryEqualityClause(&builder, &args, "p.identity_state", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "identity.privilege_level":
			if err := appendQueryEqualityClause(&builder, &args, "p.privilege_level", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "identity.mfa_state":
			if err := appendQueryEqualityClause(&builder, &args, "p.mfa_state", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "identity.reset_status":
			if err := appendQueryEqualityClause(&builder, &args, "p.reset_status", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		default:
			return "", nil, fmt.Errorf("identity query filter field %q not mapped", filter.FieldKey)
		}
	}

	if err := appendOrderBy(&builder, query.Sort, identitySortExpressions); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func buildIndicatorQuerySQL(incidentID uuid.UUID, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    i.record_id::text,
    i.incident_id::text,
    i.row_version,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    i.first_observed_at,
    i.last_observed_at,
    i.observation_count,
    i.lifecycle_summary,
    i.supporting_link_count,
    i.edited_at
  FROM indicator_grid_projection i
 WHERE i.incident_id = $1`)
	args := []any{incidentID}

	for _, filter := range query.Filters {
		switch filter.FieldKey {
		case "indicator.indicator_type":
			if err := appendQueryEqualityClause(&builder, &args, "i.indicator_type", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "indicator.value_kind":
			if err := appendQueryEqualityClause(&builder, &args, "i.value_kind", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "indicator.hash_algorithm":
			if err := appendQueryEqualityClause(&builder, &args, "i.hash_algorithm", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "indicator.lifecycle_summary":
			if err := appendQueryEqualityClause(&builder, &args, "i.lifecycle_summary", filter.Arg, ""); err != nil {
				return "", nil, err
			}
		case "indicator.first_observed_at":
			if err := appendQueryDateTimeClause(&builder, &args, "i.first_observed_at", filter); err != nil {
				return "", nil, err
			}
		case "indicator.last_observed_at":
			if err := appendQueryDateTimeClause(&builder, &args, "i.last_observed_at", filter); err != nil {
				return "", nil, err
			}
		default:
			return "", nil, fmt.Errorf("indicator query filter field %q not mapped", filter.FieldKey)
		}
	}

	if err := appendOrderBy(&builder, query.Sort, indicatorSortExpressions); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func appendOrderBy(builder *strings.Builder, sort []viewschema.SortEntry, expressions map[string]string) error {
	builder.WriteString(" ORDER BY ")
	for index, sortEntry := range sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		expr, ok := expressions[sortEntry.FieldKey]
		if !ok {
			return fmt.Errorf("query sort field %q not mapped", sortEntry.FieldKey)
		}
		builder.WriteString(expr)
		if sortEntry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
	}
	return nil
}

func appendQueryEqualityClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any, cast string) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindQueryValue(args, value, cast))
		return nil
	}
	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", expr)
	}
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindQueryValue(args, value, cast))
	}
	builder.WriteString(")")
	return nil
}

func appendQueryDateTimeClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendQueryEqualityClause(builder, args, expr, filter.Arg, "timestamptz")
	case "range":
		for _, bound := range []struct {
			Key string
			Op  string
		}{
			{Key: "gt", Op: ">"},
			{Key: "gte", Op: ">="},
			{Key: "lt", Op: "<"},
			{Key: "lte", Op: "<="},
		} {
			value, ok := filter.Arg[bound.Key]
			if !ok {
				continue
			}
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteByte(' ')
			builder.WriteString(bound.Op)
			builder.WriteByte(' ')
			builder.WriteString(bindQueryValue(args, value, "timestamptz"))
		}
		return nil
	default:
		return fmt.Errorf("timestamp filter operator %q not mapped", filter.Op)
	}
}

func bindQueryValue(args *[]any, value any, cast string) string {
	*args = append(*args, value)
	placeholder := fmt.Sprintf("$%d", len(*args))
	if cast == "" {
		return placeholder
	}
	return placeholder + "::" + cast
}

func scanHostQueryRecord(scanner interface {
	Scan(dest ...any) error
}) (HostRecord, error) {
	var (
		record           HostRecord
		rawAADDeviceID   pgtype.Text
		rawFQDN          pgtype.Text
		rawHostname      pgtype.Text
		rawLocation      pgtype.Text
		rawOSPlatform    pgtype.Text
		rawBusinessOwner pgtype.Text
		rawCriticality   pgtype.Text
		rawContainment   pgtype.Text
		rawMergedInto    pgtype.UUID
		rawSeedMention   pgtype.UUID
		linkedEventCount int32
		evidenceCount    int32
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&rawAADDeviceID,
		&rawFQDN,
		&rawHostname,
		&record.HostState,
		&linkedEventCount,
		&evidenceCount,
		&rawLocation,
		&rawOSPlatform,
		&rawBusinessOwner,
		&rawCriticality,
		&rawContainment,
		&rawMergedInto,
		&record.EntityOrigin,
		&rawSeedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return HostRecord{}, fmt.Errorf("scan host query record: %w", err)
	}
	record.AADDeviceID = textPointer(rawAADDeviceID)
	record.FQDN = textPointer(rawFQDN)
	record.Hostname = textPointer(rawHostname)
	record.Location = textPointer(rawLocation)
	record.OSPlatform = textPointer(rawOSPlatform)
	record.BusinessOwner = textPointer(rawBusinessOwner)
	record.Criticality = textPointer(rawCriticality)
	record.ContainmentStatus = textPointer(rawContainment)
	record.MergedIntoRecordID = uuidPointerFromPG(rawMergedInto)
	record.SeedMentionID = uuidPointerFromPG(rawSeedMention)
	record.LinkedEventCount = int(linkedEventCount)
	record.EvidenceCount = int(evidenceCount)
	return record, nil
}

func scanIdentityQueryRecord(scanner interface {
	Scan(dest ...any) error
}) (IdentityRecord, error) {
	var (
		record            IdentityRecord
		rawAADObjectID    pgtype.Text
		rawSID            pgtype.Text
		rawUPN            pgtype.Text
		rawEmail          pgtype.Text
		rawSamAccountName pgtype.Text
		rawPrivilegeLevel pgtype.Text
		rawMFAState       pgtype.Text
		rawResetStatus    pgtype.Text
		rawMergedInto     pgtype.UUID
		rawSeedMention    pgtype.UUID
		linkedEventCount  int32
		evidenceCount     int32
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&rawAADObjectID,
		&rawSID,
		&rawUPN,
		&rawEmail,
		&rawSamAccountName,
		&record.IdentityState,
		&linkedEventCount,
		&evidenceCount,
		&rawPrivilegeLevel,
		&rawMFAState,
		&rawResetStatus,
		&rawMergedInto,
		&record.EntityOrigin,
		&rawSeedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return IdentityRecord{}, fmt.Errorf("scan identity query record: %w", err)
	}
	record.AADObjectID = textPointer(rawAADObjectID)
	record.SID = textPointer(rawSID)
	record.UPN = textPointer(rawUPN)
	record.Email = textPointer(rawEmail)
	record.SamAccountName = textPointer(rawSamAccountName)
	record.PrivilegeLevel = textPointer(rawPrivilegeLevel)
	record.MFAState = textPointer(rawMFAState)
	record.ResetStatus = textPointer(rawResetStatus)
	record.MergedIntoRecordID = uuidPointerFromPG(rawMergedInto)
	record.SeedMentionID = uuidPointerFromPG(rawSeedMention)
	record.LinkedEventCount = int(linkedEventCount)
	record.EvidenceCount = int(evidenceCount)
	return record, nil
}

func (s *Store) CreateHostRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + HostsViewSchemaID
	if existing, err := s.authStore.GetRouteIdempotency(ctx, hostCreateRouteKey, scopeKey, request.ClientTxnID); err == nil {
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

	record, beforeRow, operationKind, statusCode, err := s.upsertHostTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := upsertHostProjectionTx(ctx, tx, record); err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
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
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
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

	payload := BuildMutationPayload(HostsViewSchemaID, changeSetID, afterRow)
	if err := insertRouteIdempotency(ctx, tx, hostCreateRouteKey, scopeKey, request.ClientTxnID, actor.ID, requestHash, statusCode, payload); err != nil {
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
	if existing, err := s.authStore.GetRouteIdempotency(ctx, identityCreateRouteKey, scopeKey, request.ClientTxnID); err == nil {
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

	record, beforeRow, operationKind, statusCode, err := s.upsertIdentityTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := upsertIdentityProjectionTx(ctx, tx, record); err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
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
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
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

	payload := BuildMutationPayload(IdentitiesViewSchemaID, changeSetID, afterRow)
	if err := insertRouteIdempotency(ctx, tx, identityCreateRouteKey, scopeKey, request.ClientTxnID, actor.ID, requestHash, statusCode, payload); err != nil {
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

func loadHostByHostnameTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, hostname string) (HostRecord, error) {
	record, err := scanHostRecord(tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    host_state,
    merged_into_record_id,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM hosts
 WHERE incident_id = $1
   AND hostname = $2
   AND host_state IN ('stub', 'canonical')
 ORDER BY updated_at DESC, record_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, hostname))
	return record, err
}

func insertHostTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO hosts (
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    host_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $11, $11)
RETURNING record_id
`, record.IncidentID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.HostState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE hosts
   SET display_name = $2,
       aad_device_id = $3,
       fqdn = $4,
       hostname = $5,
       host_state = $6,
       merged_into_record_id = $7,
       row_version = $8,
       updated_at = $9,
       updated_by_user_id = $10
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.HostState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update host: %w", err)
	}
	return nil
}

func upsertHostProjectionTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
INSERT INTO host_grid_projection (
    record_id,
    incident_id,
    row_version,
    display_name,
    hostname,
    host_state,
    linked_event_count,
    evidence_count,
    location,
    os_platform,
    business_owner,
    criticality,
    containment_status,
    edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, 0, 0, NULL, NULL, NULL, NULL, NULL, $7)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    hostname = EXCLUDED.hostname,
    host_state = EXCLUDED.host_state,
    edited_at = EXCLUDED.edited_at
`, record.RecordID, record.IncidentID, record.RowVersion, record.DisplayName, record.Hostname, record.HostState, record.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("upsert host projection: %w", err)
	}
	return nil
}

func loadIdentityByEmailTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, email string) (IdentityRecord, error) {
	record, err := scanIdentityRecord(tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email::text,
    sam_account_name,
    identity_state,
    merged_into_record_id,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM identities
 WHERE incident_id = $1
   AND email = $2
   AND identity_state IN ('stub', 'canonical')
 ORDER BY updated_at DESC, record_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, email))
	return record, err
}

func insertIdentityTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO identities (
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email,
    sam_account_name,
    identity_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12, $13, $13)
RETURNING record_id
`, record.IncidentID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
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
       identity_state = $8,
       merged_into_record_id = $9,
       row_version = $10,
       updated_at = $11,
       updated_by_user_id = $12
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update identity: %w", err)
	}
	return nil
}

func upsertIdentityProjectionTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	_, err := tx.Exec(ctx, `
INSERT INTO identity_grid_projection (
    record_id,
    incident_id,
    row_version,
    display_name,
    upn,
    email,
    sam_account_name,
    identity_state,
    linked_event_count,
    evidence_count,
    privilege_level,
    mfa_state,
    reset_status,
    edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 0, NULL, NULL, NULL, $9)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    upn = EXCLUDED.upn,
    email = EXCLUDED.email,
    sam_account_name = EXCLUDED.sam_account_name,
    identity_state = EXCLUDED.identity_state,
    edited_at = EXCLUDED.edited_at
`, record.RecordID, record.IncidentID, record.RowVersion, record.DisplayName, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("upsert identity projection: %w", err)
	}
	return nil
}

func insertRouteIdempotency(ctx context.Context, tx pgx.Tx, routeKey string, scopeKey string, clientTxnID string, actorUserID uuid.UUID, requestHash []byte, statusCode int, payload map[string]any) error {
	responseJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal idempotency payload: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key,
    scope_key,
    client_txn_id,
    actor_user_id,
    target_user_id,
    request_hash,
    status_code,
    response_json
)
VALUES ($1, $2, $3, $4, NULL, $5, $6, $7)
`, routeKey, scopeKey, clientTxnID, actorUserID, requestHash, statusCode, responseJSON); err != nil {
		return fmt.Errorf("insert route idempotency: %w", err)
	}
	return nil
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

func loadEntityAliasesByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string) (map[uuid.UUID][]string, error) {
	rows, err := querier.Query(ctx, `
SELECT record_id, raw_text
  FROM entity_aliases
 WHERE incident_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY record_id ASC, normalized_text ASC, created_at ASC, entity_alias_id ASC
`, incidentID, entityType)
	if err != nil {
		return nil, fmt.Errorf("query entity aliases by record: %w", err)
	}
	defer rows.Close()

	aliasesByRecord := make(map[uuid.UUID][]string)
	for rows.Next() {
		var (
			recordID uuid.UUID
			rawText  string
		)
		if err := rows.Scan(&recordID, &rawText); err != nil {
			return nil, fmt.Errorf("scan entity alias by record: %w", err)
		}
		aliasesByRecord[recordID] = append(aliasesByRecord[recordID], rawText)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity aliases by record: %w", err)
	}

	for recordID, aliases := range aliasesByRecord {
		slices.Sort(aliases)
		aliasesByRecord[recordID] = aliases
	}
	return aliasesByRecord, nil
}

func scanIndicatorProjectionRecord(scanner interface {
	Scan(dest ...any) error
}) (IndicatorProjectionRecord, error) {
	var (
		record          IndicatorProjectionRecord
		recordIDRaw     string
		incidentIDRaw   string
		normalizedValue sql.NullString
		defangedValue   sql.NullString
		hashAlgorithm   sql.NullString
		hashValue       sql.NullString
		stixPattern     sql.NullString
		firstObserved   sql.NullTime
		lastObserved    sql.NullTime
		lifecycle       sql.NullString
		editedAt        sql.NullTime
	)
	if err := scanner.Scan(
		&recordIDRaw,
		&incidentIDRaw,
		&record.RowVersion,
		&record.IndicatorType,
		&record.ValueKind,
		&record.DisplayValue,
		&normalizedValue,
		&record.DedupeKey,
		&defangedValue,
		&hashAlgorithm,
		&hashValue,
		&stixPattern,
		&firstObserved,
		&lastObserved,
		&record.ObservationCount,
		&lifecycle,
		&record.SupportingLinkCnt,
		&editedAt,
	); err != nil {
		return IndicatorProjectionRecord{}, fmt.Errorf("scan indicator projection record: %w", err)
	}

	record.RecordID = uuid.MustParse(recordIDRaw)
	record.IncidentID = uuid.MustParse(incidentIDRaw)
	if normalizedValue.Valid {
		record.NormalizedValue = stringPointer(normalizedValue.String)
	}
	if defangedValue.Valid {
		record.DefangedValue = stringPointer(defangedValue.String)
	}
	if hashAlgorithm.Valid {
		record.HashAlgorithm = stringPointer(hashAlgorithm.String)
	}
	if hashValue.Valid {
		record.HashValue = stringPointer(hashValue.String)
	}
	if stixPattern.Valid {
		record.STIXPattern = stringPointer(stixPattern.String)
	}
	if firstObserved.Valid {
		value := firstObserved.Time.UTC()
		record.FirstObservedAt = &value
	}
	if lastObserved.Valid {
		value := lastObserved.Time.UTC()
		record.LastObservedAt = &value
	}
	if lifecycle.Valid {
		record.LifecycleSummary = stringPointer(lifecycle.String)
	}
	if editedAt.Valid {
		record.UpdatedAt = editedAt.Time.UTC()
	}
	return record, nil
}
