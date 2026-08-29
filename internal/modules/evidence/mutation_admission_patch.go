package evidence

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

// AdmitPatchJSON admits one Evidence patch and canonicalizes its ordered
// changes before binding its durability hash.
func AdmitPatchJSON(reader io.Reader) (PatchAdmission, *AdmissionFailure) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return PatchAdmission{}, newAdmissionFailure("", admissionRequestNotObject)
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchAdmission{}, newAdmissionFailure(key, admissionUnknownField)
		}
	}
	var request patchRequest
	if value, present := raw["view_schema_id"]; !present {
		return PatchAdmission{}, newAdmissionFailure("view_schema_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ViewSchemaID) != nil || request.ViewSchemaID != ViewSchemaID {
		return PatchAdmission{}, newAdmissionFailure("view_schema_id", admissionInvalidViewSchemaID)
	}
	if value, present := raw["base_row_version"]; !present {
		return PatchAdmission{}, newAdmissionFailure("base_row_version", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return PatchAdmission{}, newAdmissionFailure("base_row_version", admissionInvalidBaseRowVersion)
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchAdmission{}, newAdmissionFailure("client_txn_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchAdmission{}, newAdmissionFailure("client_txn_id", admissionMissingRequiredField)
	}
	rawChanges, failure := decodeEvidenceRawChanges(raw["changes"])
	if failure != nil {
		return PatchAdmission{}, failure
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, changeFailure := decodeEvidencePatchChange(rawChange)
		if changeFailure != nil {
			return PatchAdmission{}, changeFailure
		}
		if _, duplicate := seen[change.FieldKey]; duplicate {
			return PatchAdmission{}, newAdmissionFailure("changes", admissionDuplicateFieldKey)
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

func patchAdmissionHash(request patchRequest) [sha256.Size]byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": change.CanonicalValue})
	}
	return sha256.Sum256(canonicalHashBytes(map[string]any{
		"view_schema_id": ViewSchemaID, "base_row_version": request.BaseRowVersion, "changes": changes,
	}))
}
