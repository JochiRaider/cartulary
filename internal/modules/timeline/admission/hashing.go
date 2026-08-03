package admission

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/google/uuid"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func CreateRequestHash(request timeline.CreateRequest) []byte {
	payload := map[string]any{
		"timeline.date_entered_text":      derefString(request.DateEnteredText),
		"timeline.analyst_text":           derefString(request.AnalystText),
		"timeline.mitre_stage_text":       derefString(request.MitreStageText),
		"timeline.device_object_text":     derefString(request.DeviceObjectText),
		"timeline.ip_address_text":        derefString(request.IPAddressText),
		"timeline.activity_utc_text":      derefString(request.ActivityUTCText),
		"timeline.activity_local_text":    derefString(request.ActivityLocalText),
		"timeline.raw_activity_text":      derefString(request.RawActivityText),
		"timeline.activity_synopsis_text": derefString(request.ActivitySynopsisText),
		"timeline.data_source_text":       derefString(request.DataSourceText),
		"timeline.host_refs":              canonicalCollectionActionPayload(request.HostRefs),
		"timeline.identity_refs":          canonicalCollectionActionPayload(request.IdentityRefs),
		"timeline.tags":                   canonicalCollectionActionPayload(request.Tags),
		"timeline.attached_evidence_ids":  canonicalCollectionActionPayload(request.AttachedEvidence),
		"raw_capture.import_columns":      request.RawCaptureColumns,
	}
	return hashRequestPayload(payload)
}

func PatchRequestHash(request timeline.PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.CanonicalChange))
	for _, change := range request.CanonicalChange {
		entry := map[string]any{"field_key": change.FieldKey}
		if change.ActionPayload != nil {
			entry["action_payload"] = canonicalCollectionActionPayload(change.ActionPayload)
		} else {
			entry["value"] = derefString(change.TextValue)
		}
		changes = append(changes, entry)
	}
	return hashRequestPayload(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"changes":          changes,
	})
}

func ConflictResolveRequestHash(claims conflicttokens.ConflictTokenClaims, request timeline.ConflictResolveRequest) []byte {
	return hashRequestPayload(map[string]any{
		"conflict_token":      request.ConflictToken,
		"resolution_kind":     request.ResolutionKind,
		"record_id":           claims.RecordID,
		"view_schema_id":      claims.ViewSchemaID,
		"field_key":           claims.FieldKey,
		"current_row_version": claims.CurrentRowVersion,
		"resolved_value":      request.CanonicalAny,
	})
}

func ActionRequestHash(baseRowVersion int64, clientTxnID string, reason *string, replacementRecordID *uuid.UUID) []byte {
	payload := map[string]any{
		"base_row_version":      baseRowVersion,
		"reason":                derefString(reason),
		"replacement_record_id": nil,
	}
	if replacementRecordID != nil {
		payload["replacement_record_id"] = replacementRecordID.String()
	}
	return hashRequestPayload(payload)
}

func canonicalCollectionActionPayload(payload *timeline.CollectionActionPayload) any {
	if payload == nil {
		return nil
	}
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.Op == "add_tag" && action.RawText != "" {
			entry["tag_name"] = action.RawText
		} else if action.RawText != "" {
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
	return append([]byte(nil), sum[:]...)
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
