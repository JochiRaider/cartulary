package timeline

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"example.com/todo/cartulary/internal/modules/auth"
)

const (
	TimelineViewSchemaID  = "cartulary.view.timeline.v1"
	timelineQueryRouteKey = "timeline.query"
	createRouteKey        = "timeline.rows.create"
	patchRouteKey         = "timeline.records.patch"
	reviewRouteKey        = "timeline.records.mark_reviewed"
	supersedeRouteKey     = "timeline.records.supersede"
	maxPatchChanges       = 32
)

var writableFieldKeys = map[string]struct{}{
	"timeline.occurred_at": {},
	"timeline.summary":     {},
	"timeline.details":     {},
	"timeline.source_text": {},
}

type CreateRequest struct {
	ClientTxnID string
	OccurredAt  *time.Time
	Summary     *string
	Details     *string
	SourceText  *string
}

type PatchRequest struct {
	ViewSchemaID    string
	BaseRowVersion  int64
	ClientTxnID     string
	CanonicalChange []PatchChange
}

type PatchChange struct {
	FieldKey   string
	OccurredAt *time.Time
	TextValue  *string
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

func DecodeTimelineQueryRequest(reader io.Reader) (map[string]json.RawMessage, *auth.APIError) {
	if reader == nil {
		return map[string]json.RawMessage{}, nil
	}
	decoder := json.NewDecoder(reader)
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		if err == io.EOF {
			return map[string]json.RawMessage{}, nil
		}
		return nil, invalidViewQuery("", "request_not_object")
	}
	if len(raw) != 0 {
		for key := range raw {
			return nil, invalidViewQuery(key, "unknown_field")
		}
	}
	return raw, nil
}

func DecodeTimelineCreateRequest(reader io.Reader) (CreateRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"client_txn_id":        {},
		"timeline.occurred_at": {},
		"timeline.summary":     {},
		"timeline.details":     {},
		"timeline.source_text": {},
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
		return PatchRequest{}, invalidMutationPayload("changes", "change_count_exceeded")
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
		"client_txn_id":        request.ClientTxnID,
		"timeline.occurred_at": formatTimestampPointer(request.OccurredAt),
		"timeline.summary":     derefString(request.Summary),
		"timeline.details":     derefString(request.Details),
		"timeline.source_text": derefString(request.SourceText),
	}
	return hashRequestPayload(payload)
}

func TimelinePatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.CanonicalChange))
	for _, change := range request.CanonicalChange {
		changes = append(changes, map[string]any{
			"field_key": change.FieldKey,
			"value":     canonicalChangeValue(change),
		})
	}
	return hashRequestPayload(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"client_txn_id":    request.ClientTxnID,
		"changes":          changes,
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
		"timeline.host_refs":               map[string]any{"value": collectionValue(true)},
		"timeline.identity_refs":           map[string]any{"value": collectionValue(true)},
		"timeline.evidence_count":          map[string]any{"value": record.EvidenceCount},
		"timeline.tags":                    map[string]any{"value": collectionValue(false)},
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

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
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

func rowVersionConflictError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: map[string]any{}}
}

func illegalTransitionError(reasonCode string) *auth.APIError {
	details := map[string]any{}
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
		"field_key": {},
		"value":     {},
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
	if _, ok := writableFieldKeys[fieldKey]; !ok {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}

	value, ok := object["value"]
	if !ok {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}

	change := PatchChange{FieldKey: fieldKey}
	switch fieldKey {
	case "timeline.occurred_at":
		timestamp, ok := normalizeNullableTimestampValue(value)
		if !ok {
			return PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		change.OccurredAt = timestamp
	default:
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

func collectionValue(ordered bool) map[string]any {
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   []any{},
	}
}

func canonicalChangeValue(change PatchChange) any {
	if change.FieldKey == "timeline.occurred_at" {
		return formatTimestampPointer(change.OccurredAt)
	}
	return derefString(change.TextValue)
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
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	if normalized == "" {
		return "", false
	}
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return "", false
		}
	}
	return normalized, true
}

func normalizeNote(raw string) (string, bool) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	if normalized == "" {
		return "", false
	}
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r):
			return "", false
		}
	}
	return normalized, true
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

func encodeCursor(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}
