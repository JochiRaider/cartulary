package tasksdecisions

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func AdmitConflictResolveJSON(reader io.Reader, token string, claims ConflictClaims) (ConflictResolveRequest, *AdmissionFailure) {
	if claims.RecordID == uuid.Nil || !isMutationView(claims.ViewSchemaID) || claims.CurrentRowVersion < 1 {
		return ConflictResolveRequest{}, invalidAdmission("conflict_token", "invalid_value")
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return ConflictResolveRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	request := ConflictResolveRequest{ConflictToken: token}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveRequest{}, invalidAdmission("conflict_token", "missing_required_field")
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return ConflictResolveRequest{}, invalidAdmission("conflict_token", "invalid_value")
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveRequest{}, invalidAdmission("resolution_kind", "missing_required_field")
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return ConflictResolveRequest{}, invalidAdmission("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveRequest{}, invalidAdmission("resolution_kind", "invalid_value")
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	resolved, present := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if present {
			return ConflictResolveRequest{}, invalidAdmission("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !present {
		return ConflictResolveRequest{}, invalidAdmission("resolved_value", "missing_required_field")
	}
	field, ok := viewschema.LookupField(claims.ViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return ConflictResolveRequest{}, invalidAdmission("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		payload, apiErr := decodeMutationCollectionPayload(claims.FieldKey, resolved)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Collection = &payload
		change.CanonicalValue = canonicalMutationCollection(payload)
	} else {
		value, canonical, apiErr := decodeMutationValue(claims.FieldKey, field, resolved, true)
		if apiErr != nil {
			return ConflictResolveRequest{}, apiErr
		}
		change.Value, change.CanonicalValue = &value, canonical
	}
	request.Patch = &PatchRequest{
		ViewSchemaID: claims.ViewSchemaID, BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID: request.ClientTxnID, Changes: []PatchChange{change},
	}
	request.CanonicalValue = change.CanonicalValue
	return request, nil
}
