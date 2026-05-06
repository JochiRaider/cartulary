package timeline

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	TimelineViewSchemaID    = "cartulary.view.timeline.v1"
	timelineQueryRouteKey   = "timeline.query"
	createRouteKey          = "timeline.rows.create"
	patchRouteKey           = "timeline.records.patch"
	conflictResolveRouteKey = "timeline.records.conflicts.resolve"
	reviewRouteKey          = "timeline.records.mark_reviewed"
	supersedeRouteKey       = "timeline.records.supersede"
	maxPatchChanges         = 32
	maxCollectionActions    = 64
)

var directWritableFieldKeys = map[string]struct{}{
	"timeline.occurred_at": {},
	"timeline.summary":     {},
	"timeline.details":     {},
	"timeline.source_text": {},
}

type CreateRequest struct {
	ClientTxnID      string
	OccurredAt       *time.Time
	Summary          *string
	Details          *string
	SourceText       *string
	HostRefs         *CollectionActionPayload
	IdentityRefs     *CollectionActionPayload
	Tags             *CollectionActionPayload
	AttachedEvidence *CollectionActionPayload
}

type PatchRequest struct {
	ViewSchemaID    string
	BaseRowVersion  int64
	ClientTxnID     string
	CanonicalChange []PatchChange
}

type ConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	ResolvedChange *PatchChange
	CanonicalAny   any
}

type PatchChange struct {
	FieldKey      string
	OccurredAt    *time.Time
	TextValue     *string
	ActionPayload *CollectionActionPayload
	CanonicalAny  any
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	NormalizedText string
	ItemRef        string
	ResolvedRecord *uuid.UUID
	LinkedRecordID *uuid.UUID
}

type ActionRequest struct {
	BaseRowVersion int64
	ClientTxnID    string
	Reason         *string
}

type SupersedeRequest struct {
	BaseRowVersion      int64
	ClientTxnID         string
	Reason              string
	ReplacementRecordID *uuid.UUID
}

func DecodeViewQueryRequest(reader io.Reader, viewSchemaID string) (viewschema.QueryMeta, *auth.APIError) {
	query, err := viewquery.Decode(reader, viewSchemaID)
	if err != nil {
		return viewschema.QueryMeta{}, invalidViewQueryValidation(err)
	}
	return query.Meta, nil
}

func DecodeTimelineCreateRequest(reader io.Reader) (CreateRequest, *auth.APIError) {
	schema, found := viewschema.Lookup(TimelineViewSchemaID)
	if !found {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request CreateRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	var ok bool
	if request.OccurredAt, ok = normalizeNullableTimestamp(raw, "timeline.occurred_at"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.occurred_at", "invalid_value")
	}
	if request.Summary, ok = normalizeNullableLineField(raw, "timeline.summary"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.summary", "invalid_value")
	}
	if request.Details, ok = normalizeNullableNoteField(raw, "timeline.details"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.details", "invalid_value")
	}
	if request.SourceText, ok = normalizeNullableNoteField(raw, "timeline.source_text"); !ok {
		return CreateRequest{}, invalidMutationPayload("timeline.source_text", "invalid_value")
	}
	if request.HostRefs, apiErr = decodeCreateCollectionActionField(raw, "timeline.host_refs"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if request.IdentityRefs, apiErr = decodeCreateCollectionActionField(raw, "timeline.identity_refs"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if request.Tags, apiErr = decodeCreateCollectionActionField(raw, "timeline.tags"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if request.AttachedEvidence, apiErr = decodeCreateCollectionActionField(raw, "timeline.attached_evidence_ids"); apiErr != nil {
		return CreateRequest{}, apiErr
	}
	if !schema.PermitsZeroFieldCreate && !CreateRequestHasUserValue(request) {
		return CreateRequest{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}

	return request, nil
}

func DecodeTimelinePatchRequest(reader io.Reader) (PatchRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return PatchRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"view_schema_id":   {},
		"base_row_version": {},
		"client_txn_id":    {},
		"changes":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request PatchRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != TimelineViewSchemaID {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, ok := raw["base_row_version"]; !ok {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	value, ok := raw["changes"]
	if !ok {
		return PatchRequest{}, invalidMutationPayload("changes", "missing_required_field")
	}
	var rawChanges []json.RawMessage
	if err := json.Unmarshal(value, &rawChanges); err != nil {
		return PatchRequest{}, invalidMutationPayload("changes", "invalid_value")
	}
	if len(rawChanges) == 0 {
		return PatchRequest{}, invalidMutationPayload("changes", "empty_changes")
	}
	if len(rawChanges) > maxPatchChanges {
		return PatchRequest{}, invalidMutationPayloadWithDetails("changes", "change_count_exceeded", map[string]any{
			"requested_count": len(rawChanges),
			"max_count":       maxPatchChanges,
		})
	}

	seen := make(map[string]struct{}, len(rawChanges))
	request.CanonicalChange = make([]PatchChange, 0, len(rawChanges))
	for index, rawChange := range rawChanges {
		change, apiErr := decodePatchChange(rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, ok := seen[change.FieldKey]; ok {
			return PatchRequest{}, invalidMutationPayload("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.CanonicalChange = append(request.CanonicalChange, change)
		_ = index
	}
	slices.SortFunc(request.CanonicalChange, func(left PatchChange, right PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func DecodeTimelineConflictResolveRequest(reader io.Reader, token string, claims TimelineConflictTokenClaims) (ConflictResolveRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return ConflictResolveRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"conflict_token":  {},
		"resolution_kind": {},
		"client_txn_id":   {},
		"resolved_value":  {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ConflictResolveRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	request := ConflictResolveRequest{ConflictToken: token}
	if value, ok := raw["conflict_token"]; !ok {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ConflictToken); err != nil || request.ConflictToken != token {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, ok := raw["resolution_kind"]; !ok {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ResolutionKind); err != nil {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	resolvedValue, hasResolvedValue := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if hasResolvedValue {
			return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !hasResolvedValue {
		return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "missing_required_field")
	}

	field, ok := viewschema.LookupField(TimelineViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeCollectionActionPayload(claims.FieldKey, resolvedValue, claims.FieldKey, "resolved_value.actions")
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.ActionPayload = payload
		change.CanonicalAny = canonicalCollectionActionPayload(payload)
	} else {
		switch claims.FieldKey {
		case "timeline.occurred_at":
			timestamp, ok := normalizeNullableTimestampValue(resolvedValue)
			if !ok {
				return ConflictResolveRequest{}, invalidMutationPayload(claims.FieldKey, "invalid_value")
			}
			change.OccurredAt = timestamp
		default:
			if _, ok := directWritableFieldKeys[claims.FieldKey]; !ok {
				return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
			}
			textValue, ok := normalizeFieldTextValue(claims.FieldKey, resolvedValue)
			if !ok {
				return ConflictResolveRequest{}, invalidMutationPayload(claims.FieldKey, "invalid_value")
			}
			change.TextValue = textValue
		}
		change.CanonicalAny = canonicalChangeValue(change)
	}
	request.ResolvedChange = &change
	request.CanonicalAny = change.CanonicalAny
	return request, nil
}

func DecodeTimelineActionRequest(reader io.Reader) (ActionRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ActionRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request ActionRequest
	if value, ok := raw["base_row_version"]; !ok {
		return ActionRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return ActionRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return ActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	var ok bool
	if request.Reason, ok = normalizeNullableNoteField(raw, "reason"); !ok {
		return ActionRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	return request, nil
}

func DecodeTimelineSupersedeRequest(reader io.Reader) (SupersedeRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return SupersedeRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_row_version":      {},
		"client_txn_id":         {},
		"reason":                {},
		"replacement_record_id": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return SupersedeRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request SupersedeRequest
	if value, ok := raw["base_row_version"]; !ok {
		return SupersedeRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return SupersedeRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return SupersedeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return SupersedeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	value, ok := raw["reason"]
	if !ok {
		return SupersedeRequest{}, invalidMutationPayload("reason", "missing_required_field")
	}
	reason, ok := normalizeNoteValue(value)
	if !ok {
		return SupersedeRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	request.Reason = reason

	if replacementValue, ok := raw["replacement_record_id"]; ok {
		if string(replacementValue) == "null" {
			request.ReplacementRecordID = nil
		} else {
			var rawID string
			if err := json.Unmarshal(replacementValue, &rawID); err != nil {
				return SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "invalid_value")
			}
			replacementID, err := uuid.Parse(rawID)
			if err != nil {
				return SupersedeRequest{}, invalidMutationPayload("replacement_record_id", "invalid_value")
			}
			request.ReplacementRecordID = &replacementID
		}
	}

	return request, nil
}

func TimelineCreateRequestHash(request CreateRequest) []byte {
	payload := map[string]any{
		"client_txn_id":                  request.ClientTxnID,
		"timeline.occurred_at":           formatTimestampPointer(request.OccurredAt),
		"timeline.summary":               derefString(request.Summary),
		"timeline.details":               derefString(request.Details),
		"timeline.source_text":           derefString(request.SourceText),
		"timeline.host_refs":             canonicalCollectionActionPayload(request.HostRefs),
		"timeline.identity_refs":         canonicalCollectionActionPayload(request.IdentityRefs),
		"timeline.tags":                  canonicalCollectionActionPayload(request.Tags),
		"timeline.attached_evidence_ids": canonicalCollectionActionPayload(request.AttachedEvidence),
	}
	return hashRequestPayload(payload)
}

func TimelinePatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.CanonicalChange))
	for _, change := range request.CanonicalChange {
		entry := map[string]any{"field_key": change.FieldKey}
		if change.ActionPayload != nil {
			entry["action_payload"] = canonicalCollectionActionPayload(change.ActionPayload)
		} else {
			entry["value"] = canonicalChangeValue(change)
		}
		changes = append(changes, entry)
	}
	return hashRequestPayload(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"client_txn_id":    request.ClientTxnID,
		"changes":          changes,
	})
}

func TimelineConflictResolveRequestHash(claims TimelineConflictTokenClaims, request ConflictResolveRequest) []byte {
	return hashRequestPayload(map[string]any{
		"conflict_token":      request.ConflictToken,
		"resolution_kind":     request.ResolutionKind,
		"client_txn_id":       request.ClientTxnID,
		"record_id":           claims.RecordID,
		"view_schema_id":      claims.ViewSchemaID,
		"field_key":           claims.FieldKey,
		"current_row_version": claims.CurrentRowVersion,
		"resolved_value":      request.CanonicalAny,
	})
}

func TimelineActionRequestHash(baseRowVersion int64, clientTxnID string, reason *string, replacementRecordID *uuid.UUID) []byte {
	payload := map[string]any{
		"base_row_version":      baseRowVersion,
		"client_txn_id":         clientTxnID,
		"reason":                derefString(reason),
		"replacement_record_id": nil,
	}
	if replacementRecordID != nil {
		payload["replacement_record_id"] = replacementRecordID.String()
	}
	return hashRequestPayload(payload)
}

func BuildRow(record projectedRecord) map[string]any {
	cells := map[string]any{
		"timeline.occurred_at":             map[string]any{"value": formatTimestampPointer(record.OccurredAt)},
		"timeline.summary":                 map[string]any{"value": derefString(record.Summary)},
		"timeline.details":                 map[string]any{"value": derefString(record.Details)},
		"timeline.source_text":             map[string]any{"value": derefString(record.SourceText)},
		"timeline.host_refs":               map[string]any{"value": collectionValue(true, record.HostRefs)},
		"timeline.identity_refs":           map[string]any{"value": collectionValue(true, record.IdentityRefs)},
		"timeline.attached_evidence_ids":   map[string]any{"value": collectionValue(false, record.AttachedEvidence)},
		"timeline.evidence_count":          map[string]any{"value": record.EvidenceCount},
		"timeline.tags":                    map[string]any{"value": collectionValue(false, record.Tags)},
		"timeline.edited_at":               map[string]any{"value": formatTimestamp(record.EditedAt)},
		"timeline.recorded_at":             map[string]any{"value": formatTimestamp(record.RecordedAt)},
		"timeline.sort_ts":                 map[string]any{"value": formatTimestamp(record.SortTs)},
		"timeline.capture_state":           map[string]any{"value": record.CaptureState},
		"timeline.replacement_record_id":   map[string]any{"value": formatUUIDPointer(record.ReplacementRecordID)},
		"timeline.occurred_day":            map[string]any{"value": formatDatePointer(record.OccurredDay)},
		"timeline.recorded_day":            map[string]any{"value": formatDate(record.RecordedDay)},
		"timeline.has_evidence":            map[string]any{"value": record.HasEvidence},
		"timeline.has_unresolved_mentions": map[string]any{"value": record.HasUnresolvedMentions},
	}

	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells":       cells,
	}
	row["group_values"] = map[string]any{
		"timeline.occurred_day":            formatDatePointer(record.OccurredDay),
		"timeline.recorded_day":            formatDate(record.RecordedDay),
		"timeline.capture_state":           record.CaptureState,
		"timeline.has_evidence":            record.HasEvidence,
		"timeline.has_unresolved_mentions": record.HasUnresolvedMentions,
	}
	return row
}

func BuildActionPayload(record projectedRecord, changeSetID uuid.UUID, reason *string) map[string]any {
	payload := map[string]any{
		"record_id":             record.RecordID.String(),
		"incident_id":           record.IncidentID.String(),
		"row_version":           record.RowVersion,
		"capture_state":         record.CaptureState,
		"change_set_id":         changeSetID.String(),
		"reason":                derefString(reason),
		"replacement_record_id": formatUUIDPointer(record.ReplacementRecordID),
	}
	return payload
}

func BuildMutationPayload(record projectedRecord, changeSetID uuid.UUID) map[string]any {
	return map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            BuildRow(record),
	}
}

func ComputeChangedFieldKeys(before *projectedRecord, after projectedRecord) []string {
	beforeCells := map[string]any{}
	if before != nil {
		beforeRow := BuildRow(*before)
		beforeCells, _ = beforeRow["cells"].(map[string]any)
	}

	afterRow := BuildRow(after)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0, len(afterCells))
	for fieldKey, afterValue := range afterCells {
		beforeValue, ok := beforeCells[fieldKey]
		if !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			changed = append(changed, fieldKey)
		}
	}
	slices.Sort(changed)
	return changed
}

func invalidViewQuery(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidViewQueryValidation(err *viewquery.ValidationError) *auth.APIError {
	if err == nil {
		return invalidViewQuery("", "")
	}
	details := map[string]any{}
	if err.Field != "" {
		details["field"] = err.Field
	}
	if err.FieldKey != "" {
		details["field_key"] = err.FieldKey
	}
	if err.FilterIndex != nil {
		details["filter_index"] = *err.FilterIndex
	}
	if err.ReasonCode != "" {
		details["reason_code"] = err.ReasonCode
	}
	if err.RequestedCount != nil {
		details["requested_count"] = *err.RequestedCount
	}
	if err.MaxCount != nil {
		details["max_count"] = *err.MaxCount
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	for key, value := range extra {
		details[key] = value
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func incidentNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func rowVersionConflictError(details ...map[string]any) *auth.APIError {
	payload := map[string]any{}
	if len(details) > 0 && details[0] != nil {
		payload = details[0]
	}
	return &auth.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: payload}
}

func illegalTransitionError(reasonCode string, sourceErr ...error) *auth.APIError {
	details := map[string]any{}
	var transitionErr *IllegalTransitionError
	for _, err := range sourceErr {
		if errors.As(err, &transitionErr) {
			break
		}
	}
	if transitionErr != nil {
		if transitionErr.ReasonCode != "" {
			reasonCode = transitionErr.ReasonCode
		}
		details["from_status"] = transitionErr.FromStatus
		details["to_status"] = transitionErr.ToStatus
		details["violated_guards"] = append([]string{}, transitionErr.ViolatedGuards...)
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusConflict,
		Code:    "illegal_transition",
		Message: "illegal transition",
		Details: details,
	}
}

func authorizationDeniedError(requiredRole string) *auth.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &auth.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: details,
	}
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func requiredRoleDescription(roles ...string) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return roles[0]
	}
	return strings.Join(roles, "|")
}

func decodePatchChange(raw json.RawMessage) (PatchChange, *auth.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}

	allowed := map[string]struct{}{
		"field_key":      {},
		"value":          {},
		"action_payload": {},
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return PatchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}

	fieldValue, ok := object["field_key"]
	if !ok {
		return PatchChange{}, invalidMutationPayload("changes", "missing_field_key")
	}
	var fieldKey string
	if err := json.Unmarshal(fieldValue, &fieldKey); err != nil {
		return PatchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}

	change := PatchChange{FieldKey: fieldKey}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}

	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return PatchChange{}, invalidMutationPayload("action_payload", "missing_required_field")
		}
		payload, apiErr := decodeCollectionActionPayload(fieldKey, actionPayload, fieldKey, "changes.action_payload.actions")
		if apiErr != nil {
			return PatchChange{}, apiErr
		}
		change.ActionPayload = payload
		return change, nil
	}

	if !hasValue {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	switch fieldKey {
	case "timeline.occurred_at":
		timestamp, ok := normalizeNullableTimestampValue(value)
		if !ok {
			return PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		change.OccurredAt = timestamp
	default:
		if _, ok := directWritableFieldKeys[fieldKey]; !ok {
			return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
		}
		textValue, ok := normalizeFieldTextValue(fieldKey, value)
		if !ok {
			return PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		change.TextValue = textValue
	}
	return change, nil
}

func normalizeFieldTextValue(fieldKey string, value json.RawMessage) (*string, bool) {
	switch fieldKey {
	case "timeline.summary":
		return normalizeNullableLineValue(value)
	case "timeline.details", "timeline.source_text":
		return normalizeNullableNoteValuePointer(value)
	default:
		return nil, false
	}
}

func decodeCreateCollectionActionField(raw map[string]json.RawMessage, fieldKey string) (*CollectionActionPayload, *auth.APIError) {
	value, ok := raw[fieldKey]
	if !ok {
		return nil, nil
	}
	payload, apiErr := decodeCollectionActionPayload(fieldKey, value, fieldKey, fieldKey+".actions")
	if apiErr != nil {
		return nil, apiErr
	}
	for _, action := range payload.Actions {
		if fieldKey == "timeline.attached_evidence_ids" {
			if action.Op != "add_record_ref" {
				return nil, invalidMutationPayload(fieldKey, "invalid_value")
			}
			continue
		}
		if fieldKey == "timeline.tags" {
			if action.Op != "add_token" {
				return nil, invalidMutationPayload(fieldKey, "invalid_value")
			}
			continue
		}
		if action.Op != "add_token" && action.Op != "add_resolved_ref" {
			return nil, invalidMutationPayload(fieldKey, "invalid_value")
		}
	}
	return payload, nil
}

func decodeCollectionActionPayload(fieldKey string, raw json.RawMessage, invalidField string, actionsField string) (*CollectionActionPayload, *auth.APIError) {
	if fieldKey != "timeline.host_refs" &&
		fieldKey != "timeline.identity_refs" &&
		fieldKey != "timeline.tags" &&
		fieldKey != "timeline.attached_evidence_ids" {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}

	var payloadObject map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payloadObject); err != nil {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	if !objectHasOnlyFields(payloadObject, "kind", "actions") {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}

	var kind string
	if err := json.Unmarshal(payloadObject["kind"], &kind); err != nil {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	var rawActions []json.RawMessage
	if err := json.Unmarshal(payloadObject["actions"], &rawActions); err != nil {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	if kind != "collection_actions_v1" {
		return nil, invalidMutationPayload(invalidField, "invalid_value")
	}
	if len(rawActions) == 0 {
		return nil, invalidMutationPayloadWithDetails(actionsField, "empty_collection_actions", map[string]any{
			"field_key": fieldKey,
		})
	}
	if len(rawActions) > maxCollectionActions {
		return nil, invalidMutationPayloadWithDetails(actionsField, "collection_action_count_exceeded", map[string]any{
			"field_key":       fieldKey,
			"requested_count": len(rawActions),
			"max_count":       maxCollectionActions,
		})
	}

	actions := make([]CollectionAction, 0, len(rawActions))
	for _, rawActionData := range rawActions {
		var rawAction map[string]json.RawMessage
		if err := json.Unmarshal(rawActionData, &rawAction); err != nil {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
		opValue, ok := rawAction["op"]
		if !ok {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}

		var op string
		if err := json.Unmarshal(opValue, &op); err != nil {
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
		switch op {
		case "add_token":
			if fieldKey == "timeline.attached_evidence_ids" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "raw_text"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			rawTextValue, ok := rawAction["raw_text"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var rawText string
			if err := json.Unmarshal(rawTextValue, &rawText); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			normalized, ok := normalizeCollectionToken(fieldKey, rawText)
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        rawText,
				NormalizedText: normalized,
			})
		case "add_resolved_ref":
			if fieldKey == "timeline.tags" || fieldKey == "timeline.attached_evidence_ids" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "raw_text", "resolved_record_id"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			rawTextValue, ok := rawAction["raw_text"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			resolvedRecordValue, ok := rawAction["resolved_record_id"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var rawText string
			if err := json.Unmarshal(rawTextValue, &rawText); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			normalized, ok := normalizeCollectionToken(fieldKey, rawText)
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var resolvedRecordID string
			if err := json.Unmarshal(resolvedRecordValue, &resolvedRecordID); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			parsed, err := uuid.Parse(resolvedRecordID)
			if err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{
				Op:             op,
				RawText:        rawText,
				NormalizedText: normalized,
				ResolvedRecord: &parsed,
			})
		case "add_record_ref":
			if fieldKey != "timeline.attached_evidence_ids" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "linked_record_id"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			linkedRecordValue, ok := rawAction["linked_record_id"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var linkedRecordID string
			if err := json.Unmarshal(linkedRecordValue, &linkedRecordID); err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			parsed, err := uuid.Parse(linkedRecordID)
			if err != nil {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, LinkedRecordID: &parsed})
		case "resolve_item":
			if fieldKey == "timeline.tags" || fieldKey == "timeline.attached_evidence_ids" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref"}, []string{"resolved_record_id"}) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			action := CollectionAction{Op: op, ItemRef: itemRef}
			if resolvedRecordValue, ok := rawAction["resolved_record_id"]; ok {
				var resolvedRecordID string
				if err := json.Unmarshal(resolvedRecordValue, &resolvedRecordID); err != nil {
					return nil, invalidMutationPayload(invalidField, "invalid_value")
				}
				parsed, err := uuid.Parse(resolvedRecordID)
				if err != nil {
					return nil, invalidMutationPayload(invalidField, "invalid_value")
				}
				action.ResolvedRecord = &parsed
			}
			actions = append(actions, action)
		case "dismiss_item", "revert_to_unresolved":
			if fieldKey == "timeline.tags" || fieldKey == "timeline.attached_evidence_ids" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, ItemRef: itemRef})
		case "remove_record_ref":
			if fieldKey != "timeline.attached_evidence_ids" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			if !actionHasOnlyFields(rawAction, []string{"op", "item_ref"}, nil) {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			itemRefValue, ok := rawAction["item_ref"]
			if !ok {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			var itemRef string
			if err := json.Unmarshal(itemRefValue, &itemRef); err != nil || strings.TrimSpace(itemRef) == "" {
				return nil, invalidMutationPayload(invalidField, "invalid_value")
			}
			actions = append(actions, CollectionAction{Op: op, ItemRef: itemRef})
		default:
			return nil, invalidMutationPayload(invalidField, "invalid_value")
		}
	}
	return &CollectionActionPayload{Actions: actions}, nil
}

func normalizeCollectionToken(fieldKey string, rawText string) (string, bool) {
	if fieldKey == "timeline.tags" {
		return fieldnorm.NormalizeLine(rawText)
	}
	return fieldnorm.NormalizeMentionToken(rawText)
}

func objectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
	required := make([]string, 0, len(fields))
	required = append(required, fields...)
	return actionHasOnlyFields(object, required, nil)
}

func actionHasOnlyFields(action map[string]json.RawMessage, required []string, optional []string) bool {
	allowed := make(map[string]struct{}, len(required)+len(optional))
	for _, key := range required {
		allowed[key] = struct{}{}
		if _, ok := action[key]; !ok {
			return false
		}
	}
	for _, key := range optional {
		allowed[key] = struct{}{}
	}
	for key := range action {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}

func collectionValue(ordered bool, items []map[string]any) map[string]any {
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   items,
	}
}

func canonicalChangeValue(change PatchChange) any {
	if change.FieldKey == "timeline.occurred_at" {
		return formatTimestampPointer(change.OccurredAt)
	}
	return derefString(change.TextValue)
}

func canonicalCollectionActionPayload(payload *CollectionActionPayload) any {
	if payload == nil {
		return nil
	}
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.RawText != "" {
			entry["raw_text"] = action.NormalizedText
		}
		if action.ItemRef != "" {
			entry["item_ref"] = action.ItemRef
		}
		if action.ResolvedRecord != nil {
			entry["resolved_record_id"] = action.ResolvedRecord.String()
		}
		if action.LinkedRecordID != nil {
			entry["linked_record_id"] = action.LinkedRecordID.String()
		}
		actions = append(actions, entry)
	}
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": actions,
	}
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func hashesEqual(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func decodeObject(reader io.Reader, invalid func(string, string) *auth.APIError) (map[string]json.RawMessage, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalid("", "request_not_object")
	}
	return raw, nil
}

func normalizeNullableTimestamp(raw map[string]json.RawMessage, field string) (*time.Time, bool) {
	value, ok := raw[field]
	if !ok {
		return nil, true
	}
	return normalizeNullableTimestampValue(value)
}

func normalizeNullableTimestampValue(value json.RawMessage) (*time.Time, bool) {
	if string(value) == "null" {
		return nil, true
	}
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil {
		return nil, false
	}
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return nil, false
	}
	parsed, err := time.Parse(time.RFC3339, rawValue)
	if err != nil {
		return nil, false
	}
	utc := parsed.UTC()
	return &utc, true
}

func normalizeNullableLineField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok {
		return nil, true
	}
	return normalizeNullableLineValue(value)
}

func normalizeNullableLineValue(value json.RawMessage) (*string, bool) {
	if string(value) == "null" {
		return nil, true
	}
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil {
		return nil, false
	}
	normalized, ok := normalizeLine(rawValue)
	if !ok {
		return nil, false
	}
	return &normalized, true
}

func normalizeNullableNoteField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok {
		return nil, true
	}
	return normalizeNullableNoteValuePointer(value)
}

func normalizeNullableNoteValuePointer(value json.RawMessage) (*string, bool) {
	if string(value) == "null" {
		return nil, true
	}
	note, ok := normalizeNoteValue(value)
	if !ok {
		return nil, false
	}
	return &note, true
}

func normalizeNoteValue(value json.RawMessage) (string, bool) {
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil {
		return "", false
	}
	return normalizeNote(rawValue)
}

func normalizeLine(raw string) (string, bool) {
	return fieldnorm.NormalizeLine(raw)
}

func normalizeNote(raw string) (string, bool) {
	return fieldnorm.NormalizeNote(raw)
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatDate(value time.Time) string {
	return value.UTC().Format("2006-01-02")
}

func formatDatePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format("2006-01-02")
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
