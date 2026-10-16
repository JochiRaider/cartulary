package evidence

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestEvidenceMetadataQueryAdmission(t *testing.T) {
	ids := make([]string, 100)
	want := make([]uuid.UUID, len(ids))
	for i := range ids {
		want[i] = uuid.New()
		ids[i] = want[i].String()
	}
	encode := func(ids []string) string {
		body, _ := json.Marshal(map[string]any{"record_ids": ids})
		return string(body)
	}
	for _, size := range []int{1, 100} {
		got, apiErr := decodeMetadataQuery(strings.NewReader(encode(ids[:size])))
		if apiErr != nil || !slices.Equal(got, want[:size]) {
			t.Fatal("valid ordered selection rejected", apiErr)
		}
	}
	for name, body := range map[string]string{
		"null": "null", "array": "[]", "scalar": "1", "missing": "{}",
		"null identities": `{"record_ids":null}`, "empty": encode(nil),
		"duplicate member":    `{"record_ids":[],"record_ids":["` + ids[0] + `"]}`,
		"trailing value":      encode(ids[:1]) + " {}",
		"unknown member":      `{"record_ids":["` + ids[0] + `"],"unknown":true}`,
		"wrong identity type": `{"record_ids":[1]}`, "invalid identity": encode([]string{"invalid"}),
		"nil identity": encode([]string{uuid.Nil.String()}), "uppercase": encode([]string{"AAAAAAAA-0000-4000-8000-000000000001"}),
		"padded": encode([]string{" " + ids[0]}), "duplicate identity": encode([]string{ids[0], ids[0]}),
		"over limit":      encode(append(slices.Clone(ids), uuid.NewString())),
		"over body bound": encode(ids[:1]) + strings.Repeat(" ", evidenceMetadataRequestMaxBytes),
	} {
		t.Run(name, func(t *testing.T) {
			got, apiErr := decodeMetadataQuery(strings.NewReader(body))
			if got != nil || apiErr == nil || apiErr.Code != "invalid_evidence_metadata_request" || apiErr.Status != 400 {
				t.Fatal("invalid metadata selection admitted", apiErr)
			}
		})
	}
	textKind, key := "text_inline", "private-key"
	blobID := uuid.New()
	available := evidenceAccessRecord{ObjectBlobID: &blobID, BlobMetadataVisible: true, StorageKey: &key,
		EvidenceLifecycleState: "available", UploadState: "available", PreviewKind: &textKind, SizeBytes: 10}
	service := &service{previewMax: 20, textPreviewMax: 10}
	for _, tc := range []struct {
		name, reason string
		change       func(*evidenceAccessRecord)
	}{
		{"eligible", "", func(*evidenceAccessRecord) {}},
		{"unattached", "no_visible_blob", func(a *evidenceAccessRecord) { a.ObjectBlobID = nil }},
		{"missing blob", "blob_missing", func(a *evidenceAccessRecord) { a.BlobMetadataVisible = false }},
		{"pending", "blob_pending", func(a *evidenceAccessRecord) { a.UploadState = "pending" }},
		{"failed", "blob_failed", func(a *evidenceAccessRecord) { a.UploadState = "failed" }},
		{"quarantined", "evidence_quarantined", func(a *evidenceAccessRecord) { a.EvidenceLifecycleState = "quarantined" }},
		{"inconsistent", "evidence_inconsistent", func(a *evidenceAccessRecord) { a.EvidenceLifecycleState = "received" }},
		{"unsupported", "unsupported_preview", func(a *evidenceAccessRecord) { a.PreviewKind = nil }},
		{"text limit", "preview_payload_too_large", func(a *evidenceAccessRecord) { a.SizeBytes = 11 }},
		{"preview limit", "preview_payload_too_large", func(a *evidenceAccessRecord) { a.SizeBytes = 21; kind := "image_inline"; a.PreviewKind = &kind }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			access := available
			tc.change(&access)
			if got := service.previewMetadataBlocker(access); got != tc.reason {
				t.Fatalf("got %q want %q", got, tc.reason)
			}
		})
	}
}
