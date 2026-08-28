package artifacts

import (
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func AdmitPatch(reader io.Reader) (PatchAdmission, *AdmissionError) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return PatchAdmission{}, newAdmissionError("", admissionRequestNotObject)
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchAdmission{}, newAdmissionError(key, admissionUnknownField)
		}
	}
	var request patchRequest
	if value, present := raw["view_schema_id"]; !present {
		return PatchAdmission{}, newAdmissionError("view_schema_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ViewSchemaID) != nil || !isArtifactBackedView(request.ViewSchemaID) {
		return PatchAdmission{}, newAdmissionError("view_schema_id", admissionInvalidViewSchemaID)
	}
	if value, present := raw["base_row_version"]; !present {
		return PatchAdmission{}, newAdmissionError("base_row_version", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return PatchAdmission{}, newAdmissionError("base_row_version", admissionInvalidBaseRowVersion)
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchAdmission{}, newAdmissionError("client_txn_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchAdmission{}, newAdmissionError("client_txn_id", admissionMissingRequiredField)
	}
	rawChanges, admissionErr := decodeArtifactRawChanges(raw["changes"])
	if admissionErr != nil {
		return PatchAdmission{}, admissionErr
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, admissionErr := decodeArtifactPatchChange(request.ViewSchemaID, rawChange)
		if admissionErr != nil {
			return PatchAdmission{}, admissionErr
		}
		if _, duplicate := seen[change.FieldKey]; duplicate {
			return PatchAdmission{}, newAdmissionError("changes", admissionDuplicateFieldKey)
		}
		seen[change.FieldKey] = struct{}{}
		request.Changes = append(request.Changes, change)
	}
	slices.SortFunc(request.Changes, func(left patchChange, right patchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	hash := patchAdmissionHash(request)
	return PatchAdmission{request: clonePatchRequest(request), hash: hash, admitted: true}, nil
}

func decodeArtifactRawChanges(raw json.RawMessage) ([]json.RawMessage, *AdmissionError) {
	if raw == nil {
		return nil, newAdmissionError("changes", admissionMissingRequiredField)
	}
	var changes []json.RawMessage
	if json.Unmarshal(raw, &changes) != nil {
		return nil, newAdmissionError("changes", admissionInvalidValue)
	}
	if len(changes) == 0 {
		return nil, newAdmissionError("changes", admissionEmptyChanges)
	}
	if len(changes) > maxMutationPatchChanges {
		return nil, newAdmissionLimitError("changes", "", admissionChangeCountExceeded, len(changes), maxMutationPatchChanges)
	}
	return changes, nil
}

func decodeArtifactPatchChange(viewSchemaID string, raw json.RawMessage) (patchChange, *AdmissionError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return patchChange{}, newAdmissionError("changes", admissionInvalidChange)
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return patchChange{}, newAdmissionError("changes", admissionUnknownField)
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return patchChange{}, newAdmissionError("changes", admissionMissingFieldKey)
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return patchChange{}, newAdmissionError("field_key", admissionInvalidValue)
	}
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return patchChange{}, newAdmissionError(fieldKey, admissionUnsupportedFieldKey)
	}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return patchChange{}, newAdmissionError("changes", admissionInvalidChange)
	}
	change := patchChange{FieldKey: fieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return patchChange{}, newAdmissionError("action_payload", admissionMissingRequiredField)
		}
		payload, admissionErr := decodeArtifactCollectionActionPayload(fieldKey, actionPayload)
		if admissionErr != nil {
			return patchChange{}, admissionErr
		}
		change.Collection = &payload
		change.CanonicalValue = canonicalArtifactCollectionPayload(payload)
		return change, nil
	}
	if !hasValue {
		return patchChange{}, newAdmissionError("value", admissionMissingRequiredField)
	}
	direct, canonical, admissionErr := decodeArtifactValue(fieldKey, field, value, true)
	if admissionErr != nil {
		return patchChange{}, admissionErr
	}
	change.Value = &direct
	change.CanonicalValue = canonical
	return change, nil
}
