package parties

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/partypolicy"
	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const (
	maxPatchChanges                  = 32
	workbookCreateOperation          = "workbook.rows.create"
	workbookPatchOperation           = "workbook.records.patch"
	workbookConflictResolveOperation = "workbook.records.conflicts.resolve"
)

type ConflictClaims struct {
	RecordID          uuid.UUID
	ViewSchemaID      string
	FieldKey          string
	CurrentRowVersion int64
}

// AdmissionError is Party's transport-neutral closed validation failure.
// Workbook alone selects the public status, code, message, and detail shape.
type AdmissionError struct {
	Field          string
	ReasonCode     string
	requestedCount int
	maxCount       int
	hasLimit       bool
}

func (e *AdmissionError) Error() string { return "parties: invalid mutation admission" }

func (e *AdmissionError) Limit() (requestedCount int, maxCount int, ok bool) {
	if e == nil || !e.hasLimit {
		return 0, 0, false
	}
	return e.requestedCount, e.maxCount, true
}

// CreateAdmission is an immutable admitted Party create request. Its field
// values and owner-derived request hash are intentionally inaccessible to
// callers outside Parties.
type CreateAdmission struct {
	clientTxnID string
	values      map[string]policy.Value
	requestHash [sha256.Size]byte
}

func (a CreateAdmission) ClientTransactionID() string { return a.clientTxnID }

// PatchAdmission is an immutable field-key-sorted Party patch request.
type PatchAdmission struct {
	operationID    string
	baseRowVersion int64
	clientTxnID    string
	changes        []patchChange
	requestHash    [sha256.Size]byte
}

func (a PatchAdmission) ClientTransactionID() string   { return a.clientTxnID }
func (a PatchAdmission) AdmittedBaseRowVersion() int64 { return a.baseRowVersion }

type patchChange struct {
	fieldKey string
	value    policy.Value
}

// ConflictResolveAdmission is an immutable admitted Party resolution request.
type ConflictResolveAdmission struct {
	conflictToken  string
	resolutionKind string
	clientTxnID    string
	claims         ConflictClaims
	patch          *PatchAdmission
	resolvedValue  any
	requestHash    [sha256.Size]byte
}

func (a ConflictResolveAdmission) ClientTransactionID() string { return a.clientTxnID }

// AdmitCreateJSON admits the closed Party create surface, calculates all four
// normalized representations once, and binds the named owner request hash.
func AdmitCreateJSON(reader io.Reader) (CreateAdmission, *AdmissionError) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return CreateAdmission{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for _, fieldKey := range policy.FieldKeys() {
		allowed[fieldKey] = struct{}{}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateAdmission{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	admission := CreateAdmission{}
	if value, present := raw["client_txn_id"]; !present {
		return CreateAdmission{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &admission.clientTxnID) != nil || strings.TrimSpace(admission.clientTxnID) == "" {
		return CreateAdmission{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	inputs := make(map[string]createValueInput, len(raw)-1)
	for _, fieldKey := range policy.FieldKeys() {
		rawValue, present := raw[fieldKey]
		if !present {
			continue
		}
		var text *string
		if json.Unmarshal(rawValue, &text) != nil {
			return CreateAdmission{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		inputs[fieldKey] = createValueInput{present: true, text: text}
	}
	values, valueErr := admitCreateValues(inputs)
	if valueErr != nil {
		return CreateAdmission{}, invalidMutationPayload(valueErr.field, valueErr.reasonCode)
	}
	admission.values = values
	hash, hashErr := createMutationRequestHash(admission.values)
	if hashErr != nil {
		return CreateAdmission{}, invalidMutationPayload("payload", "invalid_value")
	}
	admission.requestHash = hash
	return admission, nil
}

// AdmitPatchJSON admits the closed Party patch surface and binds the named
// owner request hash. A clear remains present as a change with value null.
func AdmitPatchJSON(reader io.Reader) (PatchAdmission, *AdmissionError) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return PatchAdmission{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchAdmission{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var viewSchemaID string
	if value, present := raw["view_schema_id"]; !present {
		return PatchAdmission{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if json.Unmarshal(value, &viewSchemaID) != nil || viewSchemaID != ViewSchemaID {
		return PatchAdmission{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	admission := PatchAdmission{operationID: workbookPatchOperation}
	if value, present := raw["base_row_version"]; !present {
		return PatchAdmission{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &admission.baseRowVersion) != nil || admission.baseRowVersion < 1 {
		return PatchAdmission{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchAdmission{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &admission.clientTxnID) != nil || strings.TrimSpace(admission.clientTxnID) == "" {
		return PatchAdmission{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	rawChanges, apiErr := decodeRawChanges(raw["changes"])
	if apiErr != nil {
		return PatchAdmission{}, apiErr
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, apiErr := decodePartyPatchChange(rawChange)
		if apiErr != nil {
			return PatchAdmission{}, apiErr
		}
		if _, duplicate := seen[change.fieldKey]; duplicate {
			return PatchAdmission{}, invalidMutationPayload("changes", "duplicate_field_key")
		}
		seen[change.fieldKey] = struct{}{}
		admission.changes = append(admission.changes, change)
	}
	slices.SortFunc(admission.changes, func(left, right patchChange) int {
		return strings.Compare(left.fieldKey, right.fieldKey)
	})
	hash, hashErr := patchMutationRequestHash(admission.baseRowVersion, admission.changes)
	if hashErr != nil {
		return PatchAdmission{}, invalidMutationPayload("payload", "invalid_value")
	}
	admission.requestHash = hash
	return admission, nil
}

func AdmitConflictResolveJSON(
	reader io.Reader,
	token string,
	claims ConflictClaims,
) (ConflictResolveAdmission, *AdmissionError) {
	if claims.RecordID == uuid.Nil || claims.ViewSchemaID != ViewSchemaID || claims.CurrentRowVersion < 1 {
		return ConflictResolveAdmission{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return ConflictResolveAdmission{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveAdmission{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	admission := ConflictResolveAdmission{conflictToken: token, claims: claims}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveAdmission{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if json.Unmarshal(value, &admission.conflictToken) != nil || admission.conflictToken != token {
		return ConflictResolveAdmission{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveAdmission{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if json.Unmarshal(value, &admission.resolutionKind) != nil {
		return ConflictResolveAdmission{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch admission.resolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveAdmission{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveAdmission{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &admission.clientTxnID) != nil || strings.TrimSpace(admission.clientTxnID) == "" {
		return ConflictResolveAdmission{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	resolvedValue, present := raw["resolved_value"]
	if admission.resolutionKind == "keep_saved" {
		if present {
			return ConflictResolveAdmission{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
	} else {
		if !present {
			return ConflictResolveAdmission{}, invalidMutationPayload("resolved_value", "missing_required_field")
		}
		if _, ok := policy.LookupField(claims.FieldKey); !ok {
			return ConflictResolveAdmission{}, invalidMutationPayload("field_key", "unsupported_field_key")
		}
		value, apiErr := admitPartyJSONValue(claims.FieldKey, resolvedValue)
		if apiErr != nil {
			return ConflictResolveAdmission{}, apiErr
		}
		admission.resolvedValue = value.CanonicalHashValue()
		admission.patch = &PatchAdmission{
			operationID:    workbookConflictResolveOperation,
			baseRowVersion: claims.CurrentRowVersion,
			clientTxnID:    admission.clientTxnID,
			changes:        []patchChange{{fieldKey: claims.FieldKey, value: value}},
		}
	}
	hash, hashErr := conflictMutationRequestHash(admission)
	if hashErr != nil {
		return ConflictResolveAdmission{}, invalidMutationPayload("payload", "invalid_value")
	}
	admission.requestHash = hash
	if admission.patch != nil {
		admission.patch.requestHash = hash
	}
	return admission, nil
}

func decodeRawChanges(raw json.RawMessage) ([]json.RawMessage, *AdmissionError) {
	if raw == nil {
		return nil, invalidMutationPayload("changes", "missing_required_field")
	}
	var changes []json.RawMessage
	if json.Unmarshal(raw, &changes) != nil {
		return nil, invalidMutationPayload("changes", "invalid_value")
	}
	if len(changes) == 0 {
		return nil, invalidMutationPayload("changes", "empty_changes")
	}
	if len(changes) > maxPatchChanges {
		return nil, invalidMutationLimit("changes", "change_count_exceeded", len(changes), maxPatchChanges)
	}
	return changes, nil
}

func decodePartyPatchChange(raw json.RawMessage) (patchChange, *AdmissionError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return patchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return patchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return patchChange{}, invalidMutationPayload("changes", "missing_field_key")
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return patchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	if _, ok := policy.LookupField(fieldKey); !ok {
		return patchChange{}, invalidMutationPayload(fieldKey, "unsupported_field_key")
	}
	value, hasValue := object["value"]
	if !hasValue {
		return patchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	normalized, apiErr := admitPartyJSONValue(fieldKey, value)
	if apiErr != nil {
		return patchChange{}, apiErr
	}
	return patchChange{fieldKey: fieldKey, value: normalized}, nil
}

func admitPartyJSONValue(fieldKey string, raw json.RawMessage) (policy.Value, *AdmissionError) {
	var text *string
	if json.Unmarshal(raw, &text) != nil {
		return policy.Value{}, invalidMutationPayload(fieldKey, "invalid_value")
	}
	value, admissionErr := policy.Admit(fieldKey, text)
	if admissionErr != nil {
		return policy.Value{}, invalidMutationPayload(fieldKey, admissionErr.ReasonCode)
	}
	return value, nil
}

func createMutationRequestHash(values map[string]policy.Value) ([sha256.Size]byte, error) {
	fields := make(map[string]any, len(policy.FieldKeys()))
	for _, fieldKey := range policy.FieldKeys() {
		fields[fieldKey] = values[fieldKey].CanonicalHashValue()
	}
	return mutationRequestHash(map[string]any{
		"algorithm_id":   partypolicy.MutationRequestHashAlgorithmID,
		"operation_id":   workbookCreateOperation,
		"view_schema_id": ViewSchemaID,
		"fields":         fields,
	})
}

func patchMutationRequestHash(baseRowVersion int64, changes []patchChange) ([sha256.Size]byte, error) {
	preimageChanges := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		preimageChanges = append(preimageChanges, map[string]any{
			"field_key": change.fieldKey,
			"value":     change.value.CanonicalHashValue(),
		})
	}
	return mutationRequestHash(map[string]any{
		"algorithm_id":     partypolicy.MutationRequestHashAlgorithmID,
		"operation_id":     workbookPatchOperation,
		"view_schema_id":   ViewSchemaID,
		"base_row_version": baseRowVersion,
		"changes":          preimageChanges,
	})
}

func conflictMutationRequestHash(admission ConflictResolveAdmission) ([sha256.Size]byte, error) {
	return mutationRequestHash(map[string]any{
		"algorithm_id":        partypolicy.MutationRequestHashAlgorithmID,
		"operation_id":        workbookConflictResolveOperation,
		"record_id":           admission.claims.RecordID.String(),
		"view_schema_id":      admission.claims.ViewSchemaID,
		"field_key":           admission.claims.FieldKey,
		"current_row_version": admission.claims.CurrentRowVersion,
		"conflict_token":      admission.conflictToken,
		"resolution_kind":     admission.resolutionKind,
		"resolved_value":      admission.resolvedValue,
	})
}

func mutationRequestHash(preimage any) ([sha256.Size]byte, error) {
	canonical, err := canonicaljson.Marshal(preimage)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

func invalidMutationPayload(field string, reasonCode string) *AdmissionError {
	return &AdmissionError{Field: field, ReasonCode: reasonCode}
}

func invalidMutationLimit(field string, reasonCode string, requestedCount int, maxCount int) *AdmissionError {
	return &AdmissionError{
		Field: field, ReasonCode: reasonCode,
		requestedCount: requestedCount, maxCount: maxCount, hasLimit: true,
	}
}
