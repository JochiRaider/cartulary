package evidence_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type phase5ErrorRegistry struct {
	Errors []struct {
		Code       string `json:"code"`
		HTTPStatus int    `json:"http_status"`
		Summary    string `json:"summary"`
	} `json:"errors"`
	ReasonRegistries []struct {
		ErrorCode   string `json:"error_code"`
		ReasonCodes []struct {
			Code    string `json:"code"`
			Summary string `json:"summary"`
		} `json:"reason_codes"`
	} `json:"reason_registries"`
}

func TestPhase5_ReasonCodeRegistryConformance_U_5_10(t *testing.T) {
	registry := loadPhase5ErrorRegistry(t)
	errorCodes := map[string]int{}
	for _, entry := range registry.Errors {
		errorCodes[entry.Code] = entry.HTTPStatus
	}
	if errorCodes["evidence_attach_rejected"] != 409 {
		t.Fatalf("evidence_attach_rejected missing or wrong status: %#v", errorCodes["evidence_attach_rejected"])
	}
	if errorCodes["evidence_access_unavailable"] != 409 {
		t.Fatalf("evidence_access_unavailable missing or wrong status: %#v", errorCodes["evidence_access_unavailable"])
	}
	if errorCodes["object_store_invalid_request"] != 500 {
		t.Fatalf("object_store_invalid_request missing or wrong status: %#v", errorCodes["object_store_invalid_request"])
	}

	attachReasons := phase5ReasonCodes(registry, "evidence_attach_rejected")
	wantAttachReasons := []string{
		evidence.AttachReasonBlobNotVisible,
		evidence.AttachReasonBlobPending,
		evidence.AttachReasonBlobFailed,
		evidence.AttachReasonBlobQuarantined,
		evidence.AttachReasonAcceptedContractMismatch,
		evidence.AttachReasonEvidenceQuarantined,
		evidence.AttachReasonEvidenceInconsistent,
	}
	if !reflect.DeepEqual(attachReasons, wantAttachReasons) {
		t.Fatalf("attach reason registry got %v want %v", attachReasons, wantAttachReasons)
	}
	for _, forbidden := range []string{"blob_not_attachable", "incident_mismatch"} {
		if containsString(attachReasons, forbidden) {
			t.Fatalf("attach reason registry retained legacy reason %q: %v", forbidden, attachReasons)
		}
	}

	accessReasons := phase5ReasonCodes(registry, "evidence_access_unavailable")
	wantAccessReasons := []string{
		"no_visible_blob",
		"blob_pending",
		"blob_failed",
		"blob_missing",
		"evidence_quarantined",
		"evidence_inconsistent",
		"unsupported_preview",
		"preview_payload_too_large",
	}
	if !reflect.DeepEqual(accessReasons, wantAccessReasons) {
		t.Fatalf("access reason registry got %v want %v", accessReasons, wantAccessReasons)
	}

	invalidObjectStoreReasons := phase5ReasonCodes(registry, "object_store_invalid_request")
	wantInvalidObjectStoreReasons := []string{
		"object_blob_storage_key_malformed",
		"object_blob_storage_key_identity_mismatch",
	}
	if !reflect.DeepEqual(invalidObjectStoreReasons, wantInvalidObjectStoreReasons) {
		t.Fatalf("object_store_invalid_request reason registry got %v want %v", invalidObjectStoreReasons, wantInvalidObjectStoreReasons)
	}
}

func TestPhase5_EvidenceFieldKeyRegistryClosure_U_5_11(t *testing.T) {
	evidenceResource, ok := viewschema.LookupPublicResource("cartulary.view.evidence.v1")
	if !ok {
		t.Fatal("evidence view schema not registered")
	}
	wantEvidenceKeys := []string{
		"evidence.title",
		"evidence.lifecycle_state",
		"evidence.requested_at",
		"evidence.received_at",
		"evidence.storage_ref",
		"evidence.blob_hash",
		"evidence.collector_party_text",
		"evidence.collector_party_id",
		"evidence.source_party_text",
		"evidence.source_party_id",
		"evidence.upload_state",
		"evidence.linked_record_count",
		"evidence.edited_at",
	}
	if got := phase5PublicFieldKeys(evidenceResource); !reflect.DeepEqual(got, wantEvidenceKeys) {
		t.Fatalf("evidence field keys got %v want %v", got, wantEvidenceKeys)
	}
	for _, alias := range []string{"title", "Blob Hash", "blob_hash", "object_blobs.upload_state", "evidence.object_blob_id"} {
		if _, ok := viewschema.LookupField("cartulary.view.evidence.v1", alias); ok {
			t.Fatalf("evidence registry accepted non-canonical field key %q", alias)
		}
	}

	timelineResource, ok := viewschema.LookupPublicResource("cartulary.view.timeline.v1")
	if !ok {
		t.Fatal("timeline view schema not registered")
	}
	gotTimelineEvidenceKeys := []string{}
	for _, field := range timelineResource.Fields {
		switch field.FieldKey {
		case "timeline.evidence_count", "timeline.attached_evidence_ids", "timeline.has_evidence":
			gotTimelineEvidenceKeys = append(gotTimelineEvidenceKeys, field.FieldKey)
		}
	}
	wantTimelineEvidenceKeys := []string{"timeline.evidence_count", "timeline.attached_evidence_ids", "timeline.has_evidence"}
	if !reflect.DeepEqual(gotTimelineEvidenceKeys, wantTimelineEvidenceKeys) {
		t.Fatalf("timeline evidence field keys got %v want %v", gotTimelineEvidenceKeys, wantTimelineEvidenceKeys)
	}
}

func loadPhase5ErrorRegistry(t testing.TB) phase5ErrorRegistry {
	t.Helper()
	artifact, ok := contracts.ErrorArtifactsIndex["contracts/errors/index.json"]
	if !ok {
		t.Fatal("missing generated error registry artifact")
	}
	var registry phase5ErrorRegistry
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode error registry: %v", err)
	}
	return registry
}

func phase5ReasonCodes(registry phase5ErrorRegistry, errorCode string) []string {
	for _, candidate := range registry.ReasonRegistries {
		if candidate.ErrorCode != errorCode {
			continue
		}
		reasons := make([]string, 0, len(candidate.ReasonCodes))
		for _, reason := range candidate.ReasonCodes {
			reasons = append(reasons, reason.Code)
		}
		return reasons
	}
	return nil
}

func phase5PublicFieldKeys(resource viewschema.ViewSchemaResource) []string {
	keys := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		keys = append(keys, field.FieldKey)
	}
	return keys
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
