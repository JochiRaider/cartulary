package valuecodec

import (
	"testing"

	"github.com/google/uuid"
)

func TestDecodeRecordTagMutationValueAcceptsLegacyTagIDAlias(t *testing.T) {
	tagID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	recordID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	value, err := DecodeRecordTagMutationValue(map[string]any{
		"tag_id":              tagID.String(),
		"incident_id":         incidentID.String(),
		"record_id":           recordID.String(),
		"tag_name":            "Urgent",
		"normalized_tag_name": "urgent",
	})
	if err != nil {
		t.Fatalf("decode legacy tag alias: %v", err)
	}
	if value.RecordTagID != tagID || value.IncidentID != incidentID || value.RecordID != recordID {
		t.Fatalf("unexpected identity: %#v", value)
	}
	encoded := value.Map()
	if encoded["record_tag_id"] != tagID.String() {
		t.Fatalf("canonical record_tag_id missing from encoded map: %#v", encoded)
	}
	if _, ok := encoded["tag_id"]; ok {
		t.Fatalf("legacy tag_id alias should be decode-only, got %#v", encoded)
	}
}

func TestDecodeRecordLinkMutationValueAcceptsCompactLegacyShape(t *testing.T) {
	linkID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	incidentID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	srcID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	dstID := uuid.MustParse("77777777-7777-7777-7777-777777777777")

	value, err := DecodeRecordLinkMutationValue(map[string]any{
		"record_link_id": linkID.String(),
		"incident_id":    incidentID.String(),
		"src_record_id":  srcID.String(),
		"dst_record_id":  dstID.String(),
		"link_type":      "references_record",
	})
	if err != nil {
		t.Fatalf("decode compact link value: %v", err)
	}
	if value.RecordLinkID != linkID || value.IncidentID != incidentID || value.SrcRecordID != srcID || value.DstRecordID != dstID {
		t.Fatalf("unexpected identity: %#v", value)
	}
}
