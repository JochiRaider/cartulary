package artifacts

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func AdmitConflictResolution(
	reader io.Reader,
	token string,
	context ConflictAdmissionContext,
) (ConflictResolveAdmission, *AdmissionError) {
	if !validConflictAdmissionContext(context) {
		return ConflictResolveAdmission{}, newAdmissionError("conflict_token", admissionInvalidValue)
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return ConflictResolveAdmission{}, newAdmissionError("", admissionRequestNotObject)
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveAdmission{}, newAdmissionError(key, admissionUnknownField)
		}
	}
	request := conflictResolveRequest{ConflictToken: token}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveAdmission{}, newAdmissionError("conflict_token", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return ConflictResolveAdmission{}, newAdmissionError("conflict_token", admissionInvalidValue)
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveAdmission{}, newAdmissionError("resolution_kind", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return ConflictResolveAdmission{}, newAdmissionError("resolution_kind", admissionInvalidValue)
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveAdmission{}, newAdmissionError("resolution_kind", admissionInvalidValue)
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveAdmission{}, newAdmissionError("client_txn_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveAdmission{}, newAdmissionError("client_txn_id", admissionMissingRequiredField)
	}
	resolvedValue, present := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if present {
			return ConflictResolveAdmission{}, newAdmissionError("resolved_value", admissionForbiddenField)
		}
	} else {
		if !present {
			return ConflictResolveAdmission{}, newAdmissionError("resolved_value", admissionMissingRequiredField)
		}
		field, ok := viewschema.LookupField(context.ViewSchemaID, context.FieldKey)
		if !ok || !field.Writable {
			return ConflictResolveAdmission{}, newAdmissionError("field_key", admissionUnsupportedFieldKey)
		}
		change := patchChange{FieldKey: context.FieldKey}
		if field.ConflictResolutionClass == "collection_review" {
			payload, admissionErr := decodeArtifactCollectionActionPayload(context.FieldKey, resolvedValue)
			if admissionErr != nil {
				return ConflictResolveAdmission{}, admissionErr
			}
			change.Collection = &payload
			change.CanonicalValue = canonicalArtifactCollectionPayload(payload)
		} else {
			value, canonical, admissionErr := decodeArtifactValue(context.FieldKey, field, resolvedValue, true)
			if admissionErr != nil {
				return ConflictResolveAdmission{}, admissionErr
			}
			change.Value = &value
			change.CanonicalValue = canonical
		}
		request.Patch = &patchRequest{
			ViewSchemaID: context.ViewSchemaID, BaseRowVersion: context.CurrentRowVersion,
			ClientTxnID: request.ClientTxnID, Changes: []patchChange{change},
		}
		request.CanonicalValue = change.CanonicalValue
	}
	hash := conflictResolutionAdmissionHash(context, request)
	return ConflictResolveAdmission{
		request: cloneConflictResolveRequest(request), context: context, hash: hash, admitted: true,
	}, nil
}

func validConflictAdmissionContext(context ConflictAdmissionContext) bool {
	if context.Version < 1 || context.RecordID == uuid.Nil || !isArtifactBackedView(context.ViewSchemaID) ||
		context.RouteKey != string(OperationConflictResolve) || context.FieldKey == "" || context.ConflictResolutionClass == "" ||
		context.BaseRowVersion < 1 || context.CurrentRowVersion < context.BaseRowVersion ||
		context.OriginalRequestHash == "" || context.IssuedAt.IsZero() ||
		context.ExpiresAt.IsZero() || !context.ExpiresAt.After(context.IssuedAt) {
		return false
	}
	field, ok := viewschema.LookupField(context.ViewSchemaID, context.FieldKey)
	return ok && field.Writable && field.ConflictResolutionClass == context.ConflictResolutionClass
}
