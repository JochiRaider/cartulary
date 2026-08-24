package workbook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// MutationFailureKind is Workbook's closed, content-safe mutation failure
// vocabulary. Application adapters translate owner failures into this
// vocabulary; arbitrary upstream errors and detail maps are never accepted.
type MutationFailureKind string

const (
	MutationFailureInvalidPayload            MutationFailureKind = "invalid_payload"
	MutationFailureClientTxnConflict         MutationFailureKind = "client_txn_conflict"
	MutationFailureIncidentClosed            MutationFailureKind = "incident_closed"
	MutationFailureTargetNotFound            MutationFailureKind = "target_not_found"
	MutationFailureRecordDeleted             MutationFailureKind = "record_deleted"
	MutationFailureRowVersionConflict        MutationFailureKind = "row_version_conflict"
	MutationFailureSameFieldConflict         MutationFailureKind = "same_field_conflict"
	MutationFailureNoEffectiveChange         MutationFailureKind = "no_effective_change"
	MutationFailureIllegalTransition         MutationFailureKind = "illegal_transition"
	MutationFailureEntityMatchConflict       MutationFailureKind = "entity_match_conflict"
	MutationFailurePartyMatchConflict        MutationFailureKind = "party_match_conflict"
	MutationFailureEvidenceAttach            MutationFailureKind = "evidence_attach_rejected"
	MutationFailureObjectStoreInvalid        MutationFailureKind = "object_store_invalid"
	MutationFailureObjectStoreAccessRejected MutationFailureKind = "object_store_access_rejected"
	MutationFailureObjectStoreUnavailable    MutationFailureKind = "object_store_unavailable"
)

type mutationFailureDetail interface{ mutationFailureDetail() }

type invalidPayloadFailureDetail struct {
	field          string
	reasonCode     string
	requestedCount *int
	maxCount       *int
	fieldKey       string
}

func (invalidPayloadFailureDetail) mutationFailureDetail() {}

type clientTxnConflictFailureDetail struct{ clientTxnID string }

func (clientTxnConflictFailureDetail) mutationFailureDetail() {}

type rowVersionConflictFailureDetail struct {
	recordID          uuid.UUID
	baseRowVersion    int64
	currentRowVersion int64
}

func (rowVersionConflictFailureDetail) mutationFailureDetail() {}

// SameFieldConflictClass is the closed conflict-resolution vocabulary admitted
// by the base Workbook interaction contract. Adding another class requires a
// new typed representation rather than widening the safe failure detail.
type SameFieldConflictClass string

const (
	SameFieldConflictAtomicReplace    SameFieldConflictClass = "atomic_replace"
	SameFieldConflictTextCompareMerge SameFieldConflictClass = "text_compare_merge"
	SameFieldConflictCollectionReview SameFieldConflictClass = "collection_review"
)

// OptionalConflictValue distinguishes an omitted conflict member from an
// explicitly present JSON null value.
type OptionalConflictValue struct {
	Present bool
	Value   any
}

// SameFieldConflictInput is the closed application-adapter input for a public
// same-field conflict. Values are copied into an immutable JSON representation
// by SameFieldConflictFailure.
type SameFieldConflictInput struct {
	ConflictToken           string
	RecordID                uuid.UUID
	FieldKey                string
	ConflictResolutionClass SameFieldConflictClass
	BaseRowVersion          int64
	CurrentRowVersion       int64
	ClientValue             any
	ServerValue             any
	BaseValue               OptionalConflictValue
	ServerUpdatedBy         uuid.UUID
	ServerUpdatedAt         time.Time
	SuggestedMergedValue    OptionalConflictValue
}

type immutableConflictValue struct{ raw []byte }

type sameFieldConflictFailureDetail struct {
	conflictToken           string
	recordID                uuid.UUID
	fieldKey                string
	conflictResolutionClass SameFieldConflictClass
	baseRowVersion          int64
	currentRowVersion       int64
	clientValue             immutableConflictValue
	serverValue             immutableConflictValue
	baseValue               *immutableConflictValue
	serverUpdatedBy         uuid.UUID
	serverUpdatedAt         time.Time
	suggestedMergedValue    *immutableConflictValue
}

func (sameFieldConflictFailureDetail) mutationFailureDetail() {}

type fieldFailureDetail struct {
	field      string
	reasonCode string
}

func (fieldFailureDetail) mutationFailureDetail() {}

type illegalTransitionFailureDetail struct {
	fromStatus     string
	toStatus       string
	reasonCode     string
	violatedGuards []string
}

func (illegalTransitionFailureDetail) mutationFailureDetail() {}

type entityMatchConflictFailureDetail struct {
	entityType         string
	identifierClass    string
	candidateRecordIDs []uuid.UUID
}

func (entityMatchConflictFailureDetail) mutationFailureDetail() {}

type partyMatchConflictFailureDetail struct {
	reasonCode           string
	conflictingFieldKeys []string
}

func (partyMatchConflictFailureDetail) mutationFailureDetail() {}

type reasonFailureDetail struct{ reasonCode string }

func (reasonFailureDetail) mutationFailureDetail() {}

type MutationFailure struct {
	kind   MutationFailureKind
	detail mutationFailureDetail
}

func (failure *MutationFailure) Kind() MutationFailureKind {
	if failure == nil {
		return ""
	}
	return failure.kind
}

// InvalidPayloadDetail returns the closed validation identity without exposing
// the private failure-detail hierarchy or an arbitrary details map.
func (failure *MutationFailure) InvalidPayloadDetail() (field string, reasonCode string, ok bool) {
	if failure == nil || failure.kind != MutationFailureInvalidPayload {
		return "", "", false
	}
	detail, ok := failure.detail.(invalidPayloadFailureDetail)
	if !ok {
		return "", "", false
	}
	return detail.field, detail.reasonCode, true
}

func InvalidPayloadFailure(field string, reasonCode string) *MutationFailure {
	return &MutationFailure{
		kind:   MutationFailureInvalidPayload,
		detail: invalidPayloadFailureDetail{field: field, reasonCode: reasonCode},
	}
}

func InvalidPayloadLimitFailure(field string, reasonCode string, requestedCount int, maxCount int, fieldKey string) *MutationFailure {
	return &MutationFailure{
		kind: MutationFailureInvalidPayload,
		detail: invalidPayloadFailureDetail{
			field: field, reasonCode: reasonCode, requestedCount: &requestedCount,
			maxCount: &maxCount, fieldKey: fieldKey,
		},
	}
}

func ClientTxnConflictFailure(clientTxnID string) *MutationFailure {
	return &MutationFailure{
		kind:   MutationFailureClientTxnConflict,
		detail: clientTxnConflictFailureDetail{clientTxnID: clientTxnID},
	}
}

func IncidentClosedFailure() *MutationFailure {
	return &MutationFailure{kind: MutationFailureIncidentClosed}
}

func TargetNotFoundFailure() *MutationFailure {
	return &MutationFailure{kind: MutationFailureTargetNotFound}
}

func RecordDeletedFailure() *MutationFailure {
	return &MutationFailure{kind: MutationFailureRecordDeleted}
}

func RowVersionConflictFailure(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) *MutationFailure {
	return &MutationFailure{
		kind: MutationFailureRowVersionConflict,
		detail: rowVersionConflictFailureDetail{
			recordID: recordID, baseRowVersion: baseRowVersion, currentRowVersion: currentRowVersion,
		},
	}
}

func PartyMatchConflictFailure(reasonCode string, conflictingFieldKeys []string) *MutationFailure {
	return &MutationFailure{
		kind: MutationFailurePartyMatchConflict,
		detail: partyMatchConflictFailureDetail{
			reasonCode:           reasonCode,
			conflictingFieldKeys: append([]string(nil), conflictingFieldKeys...),
		},
	}
}

func SameFieldConflictFailure(input SameFieldConflictInput) (*MutationFailure, error) {
	if strings.TrimSpace(input.ConflictToken) == "" {
		return nil, fmt.Errorf("workbook same-field conflict has an empty conflict token")
	}
	if input.RecordID == uuid.Nil || strings.TrimSpace(input.FieldKey) == "" {
		return nil, fmt.Errorf("workbook same-field conflict has an invalid record or field identity")
	}
	if input.BaseRowVersion < 1 || input.CurrentRowVersion <= input.BaseRowVersion {
		return nil, fmt.Errorf("workbook same-field conflict has an invalid revision window")
	}
	if input.ServerUpdatedBy == uuid.Nil || input.ServerUpdatedAt.IsZero() {
		return nil, fmt.Errorf("workbook same-field conflict has invalid server attribution")
	}
	input.ServerUpdatedAt = input.ServerUpdatedAt.UTC()
	if input.ConflictResolutionClass != SameFieldConflictAtomicReplace &&
		input.ConflictResolutionClass != SameFieldConflictTextCompareMerge &&
		input.ConflictResolutionClass != SameFieldConflictCollectionReview {
		return nil, fmt.Errorf("workbook same-field conflict has unsupported resolution class %q", input.ConflictResolutionClass)
	}
	if input.ConflictResolutionClass != SameFieldConflictAtomicReplace && !input.BaseValue.Present {
		return nil, fmt.Errorf("workbook merge-capable same-field conflict omits its base value")
	}
	if input.ConflictResolutionClass != SameFieldConflictTextCompareMerge && input.SuggestedMergedValue.Present {
		return nil, fmt.Errorf("workbook non-text same-field conflict contains a merge suggestion")
	}
	if input.ConflictResolutionClass == SameFieldConflictTextCompareMerge {
		if !isTextConflictValue(input.ClientValue) || !isTextConflictValue(input.ServerValue) ||
			!isTextConflictValue(input.BaseValue.Value) ||
			(input.SuggestedMergedValue.Present && !isTextConflictValue(input.SuggestedMergedValue.Value)) {
			return nil, fmt.Errorf("workbook text same-field conflict contains a non-text value")
		}
	}
	if input.ConflictResolutionClass == SameFieldConflictCollectionReview {
		if !isCollectionConflictValue(input.ClientValue) || !isCollectionConflictValue(input.ServerValue) ||
			!isCollectionConflictValue(input.BaseValue.Value) {
			return nil, fmt.Errorf("workbook collection same-field conflict contains an invalid collection value")
		}
	}
	clientValue, err := newImmutableConflictValue(input.ClientValue)
	if err != nil {
		return nil, fmt.Errorf("workbook same-field client value: %w", err)
	}
	serverValue, err := newImmutableConflictValue(input.ServerValue)
	if err != nil {
		return nil, fmt.Errorf("workbook same-field server value: %w", err)
	}
	detail := sameFieldConflictFailureDetail{
		conflictToken: input.ConflictToken, recordID: input.RecordID, fieldKey: input.FieldKey,
		conflictResolutionClass: input.ConflictResolutionClass,
		baseRowVersion:          input.BaseRowVersion, currentRowVersion: input.CurrentRowVersion,
		clientValue: clientValue, serverValue: serverValue, serverUpdatedBy: input.ServerUpdatedBy,
		serverUpdatedAt: input.ServerUpdatedAt,
	}
	if input.BaseValue.Present {
		value, valueErr := newImmutableConflictValue(input.BaseValue.Value)
		if valueErr != nil {
			return nil, fmt.Errorf("workbook same-field base value: %w", valueErr)
		}
		detail.baseValue = &value
	}
	if input.SuggestedMergedValue.Present {
		value, valueErr := newImmutableConflictValue(input.SuggestedMergedValue.Value)
		if valueErr != nil {
			return nil, fmt.Errorf("workbook same-field suggested merge value: %w", valueErr)
		}
		detail.suggestedMergedValue = &value
	}
	return &MutationFailure{kind: MutationFailureSameFieldConflict, detail: detail}, nil
}

func newImmutableConflictValue(value any) (immutableConflictValue, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return immutableConflictValue{}, err
	}
	return immutableConflictValue{raw: append([]byte(nil), raw...)}, nil
}

func (value immutableConflictValue) decode() (any, error) {
	var decoded any
	if err := json.Unmarshal(value.raw, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func isTextConflictValue(value any) bool {
	if value == nil {
		return true
	}
	_, ok := value.(string)
	return ok
}

func isCollectionConflictValue(value any) bool {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 3 || object["kind"] != "collection_value_v1" {
		return false
	}
	if _, ok := object["ordered"].(bool); !ok {
		return false
	}
	switch items := object["items"].(type) {
	case []any:
		for _, item := range items {
			if _, ok := item.(map[string]any); !ok {
				return false
			}
		}
		return true
	case []map[string]any:
		return true
	default:
		return false
	}
}

func NoEffectiveChangeFailure(field string) *MutationFailure {
	return &MutationFailure{
		kind:   MutationFailureNoEffectiveChange,
		detail: fieldFailureDetail{field: field, reasonCode: "no_effective_change"},
	}
}

func IllegalTransitionFailure(fromStatus string, toStatus string, reasonCode string, violatedGuards []string) *MutationFailure {
	return &MutationFailure{
		kind: MutationFailureIllegalTransition,
		detail: illegalTransitionFailureDetail{
			fromStatus: fromStatus, toStatus: toStatus, reasonCode: reasonCode,
			violatedGuards: append([]string(nil), violatedGuards...),
		},
	}
}

func EntityMatchConflictFailure(entityType string, identifierClass string, candidateRecordIDs []uuid.UUID) *MutationFailure {
	return &MutationFailure{
		kind: MutationFailureEntityMatchConflict,
		detail: entityMatchConflictFailureDetail{
			entityType: entityType, identifierClass: identifierClass,
			candidateRecordIDs: append([]uuid.UUID(nil), candidateRecordIDs...),
		},
	}
}

func EvidenceAttachFailure(reasonCode string) *MutationFailure {
	return &MutationFailure{kind: MutationFailureEvidenceAttach, detail: reasonFailureDetail{reasonCode: reasonCode}}
}

func ObjectStoreInvalidFailure(reasonCode string) *MutationFailure {
	return &MutationFailure{kind: MutationFailureObjectStoreInvalid, detail: reasonFailureDetail{reasonCode: reasonCode}}
}

func ObjectStoreAccessRejectedFailure(reasonCode string) *MutationFailure {
	return &MutationFailure{kind: MutationFailureObjectStoreAccessRejected, detail: reasonFailureDetail{reasonCode: reasonCode}}
}

func ObjectStoreUnavailableFailure(reasonCode string) *MutationFailure {
	return &MutationFailure{kind: MutationFailureObjectStoreUnavailable, detail: reasonFailureDetail{reasonCode: reasonCode}}
}

func mutationFailureAPIError(failure *MutationFailure) *httpapi.APIError {
	if failure == nil {
		return internalAPIError(nil)
	}
	switch failure.Kind() {
	case MutationFailureInvalidPayload:
		detail, ok := failure.detail.(invalidPayloadFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		extra := map[string]any{}
		if detail.requestedCount != nil {
			extra["requested_count"] = *detail.requestedCount
		}
		if detail.maxCount != nil {
			extra["max_count"] = *detail.maxCount
		}
		if detail.fieldKey != "" {
			extra["field_key"] = detail.fieldKey
		}
		return invalidMutationPayloadWithDetails(detail.field, detail.reasonCode, extra)
	case MutationFailureClientTxnConflict:
		detail, ok := failure.detail.(clientTxnConflictFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return httpapi.ClientTxnConflictError(detail.clientTxnID)
	case MutationFailureIncidentClosed:
		return incidentClosedError()
	case MutationFailureTargetNotFound:
		return incidentNotFoundError()
	case MutationFailureRecordDeleted:
		return &httpapi.APIError{
			Status: http.StatusConflict, Code: "record_deleted_use_restore",
			Message: "record deleted use restore", Details: map[string]any{},
		}
	case MutationFailureRowVersionConflict:
		detail, ok := failure.detail.(rowVersionConflictFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return rowVersionConflictError(map[string]any{
			"record_id": detail.recordID.String(), "base_row_version": detail.baseRowVersion,
			"current_row_version": detail.currentRowVersion,
		})
	case MutationFailureSameFieldConflict:
		detail, ok := failure.detail.(sameFieldConflictFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		clientValue, err := detail.clientValue.decode()
		if err != nil {
			return internalAPIError(nil)
		}
		serverValue, err := detail.serverValue.decode()
		if err != nil {
			return internalAPIError(nil)
		}
		conflict := map[string]any{
			"conflict_token": detail.conflictToken, "record_id": detail.recordID.String(),
			"field_key": detail.fieldKey, "conflict_resolution_class": string(detail.conflictResolutionClass),
			"base_row_version": detail.baseRowVersion, "current_row_version": detail.currentRowVersion,
			"client_value": clientValue, "server_value": serverValue,
			"server_updated_by": detail.serverUpdatedBy.String(),
			"server_updated_at": detail.serverUpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if detail.baseValue != nil {
			value, decodeErr := detail.baseValue.decode()
			if decodeErr != nil {
				return internalAPIError(nil)
			}
			conflict["base_value"] = value
		}
		if detail.suggestedMergedValue != nil {
			value, decodeErr := detail.suggestedMergedValue.decode()
			if decodeErr != nil {
				return internalAPIError(nil)
			}
			conflict["suggested_merged_value"] = value
		}
		return &httpapi.APIError{
			Status: http.StatusConflict, Code: "same_field_conflict", Message: "same field conflict",
			Details: map[string]any{}, Conflict: conflict,
		}
	case MutationFailureNoEffectiveChange:
		detail, ok := failure.detail.(fieldFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return invalidMutationPayload(detail.field, detail.reasonCode)
	case MutationFailureIllegalTransition:
		detail, ok := failure.detail.(illegalTransitionFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		details := map[string]any{
			"from_status": detail.fromStatus, "to_status": detail.toStatus,
			"violated_guards": append([]string(nil), detail.violatedGuards...),
		}
		if detail.reasonCode != "" {
			details["reason_code"] = detail.reasonCode
		}
		return &httpapi.APIError{
			Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: details,
		}
	case MutationFailureEntityMatchConflict:
		detail, ok := failure.detail.(entityMatchConflictFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return entityMatchConflictError(detail.entityType, detail.identifierClass, detail.candidateRecordIDs)
	case MutationFailurePartyMatchConflict:
		detail, ok := failure.detail.(partyMatchConflictFailureDetail)
		if !ok || !validPartyMatchFailureDetail(detail) {
			return internalAPIError(nil)
		}
		return &httpapi.APIError{
			Status: http.StatusConflict, Code: "party_match_conflict", Message: "party match conflict",
			Retryable: false,
			Details: map[string]any{
				"reason_code":            detail.reasonCode,
				"conflicting_field_keys": append([]string(nil), detail.conflictingFieldKeys...),
			},
		}
	case MutationFailureEvidenceAttach:
		detail, ok := failure.detail.(reasonFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return &httpapi.APIError{
			Status: http.StatusConflict, Code: "evidence_attach_rejected", Message: "evidence attach rejected",
			Details: map[string]any{"reason_code": detail.reasonCode},
		}
	case MutationFailureObjectStoreInvalid:
		detail, ok := failure.detail.(reasonFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return &httpapi.APIError{
			Status: http.StatusInternalServerError, Code: "object_store_invalid_request",
			Details: map[string]any{"reason_code": detail.reasonCode},
		}
	case MutationFailureObjectStoreAccessRejected:
		detail, ok := failure.detail.(reasonFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return &httpapi.APIError{
			Status: http.StatusServiceUnavailable, Code: "object_store_access_rejected", Retryable: false,
			Details: map[string]any{"reason_code": detail.reasonCode},
		}
	case MutationFailureObjectStoreUnavailable:
		detail, ok := failure.detail.(reasonFailureDetail)
		if !ok {
			return internalAPIError(nil)
		}
		return &httpapi.APIError{
			Status: http.StatusServiceUnavailable, Code: "object_store_unavailable", Retryable: true,
			Details: map[string]any{"reason_code": detail.reasonCode},
		}
	default:
		return internalAPIError(nil)
	}
}

func validPartyMatchFailureDetail(detail partyMatchConflictFailureDetail) bool {
	switch detail.reasonCode {
	case "cross_key_exact_match", "exact_match_key_claimed":
	default:
		return false
	}
	if len(detail.conflictingFieldKeys) == 0 || len(detail.conflictingFieldKeys) > 2 || !slices.IsSorted(detail.conflictingFieldKeys) {
		return false
	}
	previous := ""
	for _, field := range detail.conflictingFieldKeys {
		if (field != "party.external_ref" && field != "party.primary_email") || field == previous {
			return false
		}
		previous = field
	}
	return true
}

func mutationFailureFromAPIError(apiErr *httpapi.APIError) (*MutationFailure, error) {
	if apiErr == nil {
		return nil, nil
	}
	if apiErr.Code != "invalid_mutation_payload" {
		return nil, fmt.Errorf("workbook provider returned unsupported decoder error code %q", apiErr.Code)
	}
	field, _ := apiErr.Details["field"].(string)
	reasonCode, _ := apiErr.Details["reason_code"].(string)
	requestedCount, requestedOK := apiErr.Details["requested_count"].(int)
	maxCount, maxOK := apiErr.Details["max_count"].(int)
	fieldKey, _ := apiErr.Details["field_key"].(string)
	if requestedOK || maxOK || fieldKey != "" {
		return InvalidPayloadLimitFailure(field, reasonCode, requestedCount, maxCount, fieldKey), nil
	}
	return InvalidPayloadFailure(field, reasonCode), nil
}

// DecodeMutationFailure converts only Workbook-owned decoder errors into the
// neutral failure vocabulary. Source adapters must classify owner errors
// directly and must not pass arbitrary upstream API errors here.
func DecodeMutationFailure(apiErr *httpapi.APIError) (*MutationFailure, error) {
	return mutationFailureFromAPIError(apiErr)
}

type MutationOutcome struct {
	result     *MutationResult
	resultKind mutationResultKind
	failure    *MutationFailure
}

type mutationResultKind string

const (
	mutationResultRow        mutationResultKind = "row"
	mutationResultBatch      mutationResultKind = "batch"
	mutationResultLinkedNote mutationResultKind = "linked_note"
	mutationResultSupersede  mutationResultKind = "supersede"
)

func SuccessfulRowMutation(result MutationResult) MutationOutcome {
	return successfulMutation(result, mutationResultRow)
}

func SuccessfulBatchMutation(result MutationResult) MutationOutcome {
	return successfulMutation(result, mutationResultBatch)
}

func SuccessfulLinkedNoteMutation(result MutationResult) MutationOutcome {
	return successfulMutation(result, mutationResultLinkedNote)
}

func SuccessfulSupersedeMutation(result MutationResult) MutationOutcome {
	return successfulMutation(result, mutationResultSupersede)
}

func successfulMutation(result MutationResult, kind mutationResultKind) MutationOutcome {
	cloned := cloneMutationResult(result)
	return MutationOutcome{result: &cloned, resultKind: kind}
}

func RejectedMutation(failure *MutationFailure) MutationOutcome {
	return MutationOutcome{failure: failure}
}

func (outcome MutationOutcome) Validate() error {
	if (outcome.result == nil) == (outcome.failure == nil) {
		return fmt.Errorf("workbook mutation outcome must contain exactly one result or failure")
	}
	if outcome.failure != nil && outcome.failure.kind == "" {
		return fmt.Errorf("workbook mutation outcome has an empty failure kind")
	}
	if outcome.result != nil {
		if outcome.result.StatusCode < http.StatusOK || outcome.result.StatusCode > 299 {
			return fmt.Errorf("workbook mutation outcome has invalid success status %d", outcome.result.StatusCode)
		}
		if outcome.result.Payload == nil {
			return fmt.Errorf("workbook mutation outcome has nil response payload")
		}
		if err := validateMutationResultShape(outcome.resultKind, outcome.result.Payload); err != nil {
			return err
		}
	}
	return nil
}

func validateMutationResultShape(kind mutationResultKind, payload map[string]any) error {
	var allowed map[string]struct{}
	var requiredAlternatives [][]string
	switch kind {
	case mutationResultRow:
		allowed = resultKeys("view_schema_id", "change_set_id", "row")
		requiredAlternatives = [][]string{{"row"}}
	case mutationResultBatch:
		allowed = resultKeys("view_schema_id", "change_set_id", "rows", "conflicts")
		requiredAlternatives = [][]string{{"view_schema_id", "rows"}}
	case mutationResultLinkedNote:
		allowed = resultKeys("view_schema_id", "change_set_id", "row", "source_record_id", "link_type")
		requiredAlternatives = [][]string{{"row", "source_record_id", "link_type"}}
	case mutationResultSupersede:
		allowed = resultKeys(
			"view_schema_id", "change_set_id", "row", "target_record_id", "superseding_record_id",
			"target_row_version", "superseding_row_version", "target_status", "reason",
			"record_id", "incident_id", "row_version", "capture_state", "replacement_record_id",
		)
		requiredAlternatives = [][]string{
			{"row"},
			{"target_record_id", "superseding_record_id"},
			{"record_id", "incident_id", "row_version", "capture_state", "reason", "replacement_record_id"},
		}
	default:
		return fmt.Errorf("workbook mutation outcome has unknown result kind %q", kind)
	}
	for key := range payload {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("workbook %s mutation result contains unsupported outer key %q", kind, key)
		}
	}
	for _, alternative := range requiredAlternatives {
		complete := true
		for _, key := range alternative {
			if _, ok := payload[key]; !ok {
				complete = false
				break
			}
		}
		if complete {
			return nil
		}
	}
	return fmt.Errorf("workbook %s mutation result is missing required response fields", kind)
}

func resultKeys(keys ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		result[key] = struct{}{}
	}
	return result
}

func (outcome MutationOutcome) Result() (MutationResult, bool) {
	if outcome.result == nil {
		return MutationResult{}, false
	}
	return cloneMutationResult(*outcome.result), true
}

func (outcome MutationOutcome) Failure() (*MutationFailure, bool) {
	return outcome.failure, outcome.failure != nil
}

func cloneMutationResult(result MutationResult) MutationResult {
	result.Payload = cloneMutationMap(result.Payload)
	result.ChangedFieldKeys = append([]string(nil), result.ChangedFieldKeys...)
	result.AdditionalRecordChanges = append([]MutationResult(nil), result.AdditionalRecordChanges...)
	return result
}

func cloneMutationMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

// RecordLifecycleState is the only lifecycle fact generic Workbook handlers
// need from the Records owner.
type RecordLifecycleState string

const (
	RecordLifecycleActive  RecordLifecycleState = "active"
	RecordLifecycleDeleted RecordLifecycleState = "deleted"
)

var ErrRecordTargetNotFound = errors.New("workbook record target not found")

type RecordTarget struct {
	RecordID       uuid.UUID
	IncidentID     uuid.UUID
	RecordType     string
	LifecycleState RecordLifecycleState
}

type RecordTargetResolver interface {
	ResolveRecordTarget(context.Context, uuid.UUID) (RecordTarget, error)
}

type ConflictClaims struct {
	Version                 int64
	RecordID                uuid.UUID
	ViewSchemaID            string
	RouteKey                string
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	RequestHash             string
	IssuedAt                time.Time
	ExpiresAt               time.Time
}

type ConflictTokenDecoder interface {
	DecodeConflictToken(token string) (ConflictClaims, bool)
}
