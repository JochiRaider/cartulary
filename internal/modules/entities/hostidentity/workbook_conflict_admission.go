package hostidentity

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type WorkbookConflictClaims struct {
	RecordID          uuid.UUID
	ViewSchemaID      string
	FieldKey          string
	CurrentRowVersion int64
}

type WorkbookConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	Patch          *PatchRequest
	CanonicalValue any
}

func DecodeWorkbookConflictResolveRequest(
	reader io.Reader,
	token string,
	claims WorkbookConflictClaims,
) (WorkbookConflictResolveRequest, *httpapi.APIError) {
	if claims.RecordID == uuid.Nil || !isEntityPatchSurface(claims.ViewSchemaID) ||
		claims.CurrentRowVersion < 1 {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for field := range raw {
		if _, ok := allowed[field]; !ok {
			return WorkbookConflictResolveRequest{}, invalidMutationPayload(field, "unknown_field")
		}
	}

	request := WorkbookConflictResolveRequest{ConflictToken: token}
	if value, ok := raw["conflict_token"]; !ok {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, ok := raw["resolution_kind"]; !ok {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	resolvedValue, hasResolvedValue := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if hasResolvedValue {
			return WorkbookConflictResolveRequest{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !hasResolvedValue {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("resolved_value", "missing_required_field")
	}

	field, ok := viewschema.LookupField(claims.ViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable ||
		(!isEntityDirectPatchField(claims.ViewSchemaID, claims.FieldKey) && !IsAliasCollectionField(claims.FieldKey)) {
		return WorkbookConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	change := PatchChange{FieldKey: claims.FieldKey}
	if IsAliasCollectionField(claims.FieldKey) {
		actions, ok := decodeAliasPatchActionPayload(claims.FieldKey, resolvedValue)
		if !ok {
			return WorkbookConflictResolveRequest{}, invalidMutationPayload(claims.FieldKey, "invalid_value")
		}
		change.CollectionActions = actions
		request.CanonicalValue = canonicalAliasActions(actions)
	} else {
		value, apiErr := decodeEntityPatchValue(claims.FieldKey, field, resolvedValue)
		if apiErr != nil {
			return WorkbookConflictResolveRequest{}, apiErr
		}
		change.Value = value
		request.CanonicalValue = canonicalPatchValue(value)
	}
	request.Patch = &PatchRequest{
		ViewSchemaID: claims.ViewSchemaID, BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID: request.ClientTxnID, Changes: []PatchChange{change},
	}
	return request, nil
}

func WorkbookConflictResolveRequestHash(
	claims WorkbookConflictClaims,
	request WorkbookConflictResolveRequest,
) []byte {
	data, _ := json.Marshal(map[string]any{
		"conflict_token": request.ConflictToken, "resolution_kind": request.ResolutionKind,
		"record_id": claims.RecordID, "view_schema_id": claims.ViewSchemaID,
		"field_key": claims.FieldKey, "current_row_version": claims.CurrentRowVersion,
		"resolved_value": request.CanonicalValue,
	})
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}
