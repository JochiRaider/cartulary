package evidence

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type conflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	Patch          *patchRequest
	CanonicalValue any
}

// AdmitConflictResolveJSON admits a conflict resolution together with the
// complete validated conflict-token context.
func AdmitConflictResolveJSON(
	reader io.Reader,
	token string,
	context ConflictAdmissionContext,
) (ConflictResolveAdmission, *AdmissionFailure) {
	if !validConflictAdmissionContext(context) {
		return ConflictResolveAdmission{}, newAdmissionFailure("conflict_token", admissionInvalidValue)
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return ConflictResolveAdmission{}, newAdmissionFailure("", admissionRequestNotObject)
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveAdmission{}, newAdmissionFailure(key, admissionUnknownField)
		}
	}
	request := conflictResolveRequest{ConflictToken: token}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveAdmission{}, newAdmissionFailure("conflict_token", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return ConflictResolveAdmission{}, newAdmissionFailure("conflict_token", admissionInvalidValue)
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveAdmission{}, newAdmissionFailure("resolution_kind", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return ConflictResolveAdmission{}, newAdmissionFailure("resolution_kind", admissionInvalidValue)
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveAdmission{}, newAdmissionFailure("resolution_kind", admissionInvalidValue)
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveAdmission{}, newAdmissionFailure("client_txn_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveAdmission{}, newAdmissionFailure("client_txn_id", admissionMissingRequiredField)
	}
	resolvedValue, present := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if present {
			return ConflictResolveAdmission{}, newAdmissionFailure("resolved_value", admissionForbiddenField)
		}
	} else {
		if !present {
			return ConflictResolveAdmission{}, newAdmissionFailure("resolved_value", admissionMissingRequiredField)
		}
		field, ok := viewschema.LookupField(context.ViewSchemaID, context.FieldKey)
		if !ok || !field.Writable {
			return ConflictResolveAdmission{}, newAdmissionFailure("field_key", admissionUnsupportedFieldKey)
		}
		value, canonical, valueFailure := decodeEvidenceValue(context.FieldKey, field, resolvedValue, true)
		if valueFailure != nil {
			return ConflictResolveAdmission{}, valueFailure
		}
		request.Patch = &patchRequest{
			ViewSchemaID: context.ViewSchemaID, BaseRowVersion: context.CurrentRowVersion,
			ClientTxnID: request.ClientTxnID,
			Changes:     []patchChange{{FieldKey: context.FieldKey, Value: &value, CanonicalValue: canonical}},
		}
		request.CanonicalValue = canonical
	}
	hash := conflictResolveAdmissionHash(context, request)
	return ConflictResolveAdmission{
		request: cloneConflictResolveRequest(request), context: context, hash: hash, admitted: true,
	}, nil
}

func validConflictAdmissionContext(context ConflictAdmissionContext) bool {
	if context.Version < 1 || context.RecordID == uuid.Nil || context.ViewSchemaID != ViewSchemaID ||
		context.RouteKey != string(OperationConflictResolve) || context.FieldKey == "" ||
		context.ConflictResolutionClass == "" || context.BaseRowVersion < 1 ||
		context.CurrentRowVersion < context.BaseRowVersion || context.OriginalRequestHash == "" ||
		context.IssuedAt.IsZero() || context.ExpiresAt.IsZero() || !context.ExpiresAt.After(context.IssuedAt) {
		return false
	}
	field, ok := viewschema.LookupField(context.ViewSchemaID, context.FieldKey)
	return ok && field.Writable && field.ConflictResolutionClass == context.ConflictResolutionClass
}

func conflictResolveAdmissionHash(context ConflictAdmissionContext, request conflictResolveRequest) [sha256.Size]byte {
	return sha256.Sum256(canonicalHashBytes(map[string]any{
		"conflict_token": request.ConflictToken, "resolution_kind": request.ResolutionKind,
		"record_id": context.RecordID, "view_schema_id": context.ViewSchemaID,
		"field_key": context.FieldKey, "current_row_version": context.CurrentRowVersion,
		"resolved_value": request.CanonicalValue,
	}))
}
