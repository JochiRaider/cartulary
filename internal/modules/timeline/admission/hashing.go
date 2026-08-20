package admission

import (
	"github.com/google/uuid"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
)

func CreateRequestHash(request timeline.CreateRequest) []byte {
	payload := map[string]any{
		"timeline.date_entered_text":      valuecodec.OptionalString(request.DateEnteredText),
		"timeline.analyst_text":           valuecodec.OptionalString(request.AnalystText),
		"timeline.mitre_stage_text":       valuecodec.OptionalString(request.MitreStageText),
		"timeline.device_object_text":     valuecodec.OptionalString(request.DeviceObjectText),
		"timeline.ip_address_text":        valuecodec.OptionalString(request.IPAddressText),
		"timeline.activity_utc_text":      valuecodec.OptionalString(request.ActivityUTCText),
		"timeline.activity_local_text":    valuecodec.OptionalString(request.ActivityLocalText),
		"timeline.raw_activity_text":      valuecodec.OptionalString(request.RawActivityText),
		"timeline.activity_synopsis_text": valuecodec.OptionalString(request.ActivitySynopsisText),
		"timeline.data_source_text":       valuecodec.OptionalString(request.DataSourceText),
		"timeline.host_refs":              request.HostRefs.CanonicalValue(),
		"timeline.identity_refs":          request.IdentityRefs.CanonicalValue(),
		"timeline.tags":                   request.Tags.CanonicalValue(),
		"timeline.attached_evidence_ids":  request.AttachedEvidence.CanonicalValue(),
		"raw_capture.import_columns":      request.RawCaptureColumns,
	}
	return valuecodec.CanonicalJSONSHA256(payload)
}

func PatchRequestHash(request timeline.PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.CanonicalChange))
	for _, change := range request.CanonicalChange {
		entry := map[string]any{"field_key": change.FieldKey}
		if change.ActionPayload != nil {
			entry["action_payload"] = change.ActionPayload.CanonicalValue()
		} else {
			entry["value"] = change.CanonicalValue()
		}
		changes = append(changes, entry)
	}
	return valuecodec.CanonicalJSONSHA256(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"changes":          changes,
	})
}

func ConflictResolveRequestHash(claims conflicttokens.ConflictTokenClaims, request timeline.ConflictResolveRequest) []byte {
	return valuecodec.CanonicalJSONSHA256(map[string]any{
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
		"reason":                valuecodec.OptionalString(reason),
		"replacement_record_id": valuecodec.OptionalUUID(replacementRecordID),
	}
	return valuecodec.CanonicalJSONSHA256(payload)
}
