package valuecodec

import (
	"testing"
	"time"

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

func TestDecodeRecordLinkRestorePlanAppliesLegacyDefaults(t *testing.T) {
	linkID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	incidentID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	srcID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	dstID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	actorID := uuid.MustParse("88888888-8888-8888-8888-888888888888")

	plan, err := DecodeRecordLinkRestorePlan(map[string]any{
		"record_link_id": linkID.String(),
		"incident_id":    incidentID.String(),
		"src_record_id":  srcID.String(),
		"dst_record_id":  dstID.String(),
		"link_type":      "references_record",
		"owner_user_id":  "not-a-uuid",
	}, actorID)
	if err != nil {
		t.Fatalf("decode restore plan: %v", err)
	}
	if plan.Identity.RecordLinkID != linkID || plan.Identity.IncidentID != incidentID || plan.Identity.SrcRecordID != srcID || plan.Identity.DstRecordID != dstID {
		t.Fatalf("unexpected identity: %#v", plan.Identity)
	}
	if plan.FieldKeyValue() != nil || plan.Provenance != "rollback" || plan.Confidence != nil || plan.DecidedAt != nil || plan.OwnerUserID != actorID {
		t.Fatalf("unexpected defaulted restore plan: %#v", plan)
	}
}

func TestDecodeRecordTagRestorePlanUsesCanonicalTagIdentity(t *testing.T) {
	tagID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	recordID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	plan, err := DecodeRecordTagRestorePlan(map[string]any{
		"tag_id":              tagID.String(),
		"incident_id":         incidentID.String(),
		"record_id":           recordID.String(),
		"tag_name":            "Urgent",
		"normalized_tag_name": "urgent",
	})
	if err != nil {
		t.Fatalf("decode tag restore plan: %v", err)
	}
	if plan.Identity.RecordTagID != tagID || plan.Identity.IncidentID != incidentID || plan.Identity.RecordID != recordID {
		t.Fatalf("unexpected identity: %#v", plan.Identity)
	}
	if plan.TagName != "Urgent" || plan.NormalizedTagName != "urgent" {
		t.Fatalf("unexpected tag text: %#v", plan)
	}
}

func TestBuildRecordMutationValuesEmitCanonicalMaps(t *testing.T) {
	linkID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	tagID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	srcID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	dstID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	actorID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	fieldKey := "timeline.attached_evidence_ids"
	confidence := 80
	createdAt := time.Date(2026, 7, 9, 20, 0, 0, 123, time.UTC)
	deletedAt := createdAt.Add(time.Minute)

	linkValue := BuildRecordLinkMutationValue(RecordLinkMutationInput{
		RecordLinkID:    linkID,
		IncidentID:      incidentID,
		SrcRecordID:     srcID,
		DstRecordID:     dstID,
		LinkType:        "attached_evidence",
		FieldKey:        &fieldKey,
		Provenance:      "manual",
		Confidence:      &confidence,
		OwnerUserID:     actorID,
		CreatedByUserID: actorID,
		DecidedAt:       createdAt,
		CreatedAt:       createdAt,
		DeletedAt:       &deletedAt,
		DeletedByUserID: &actorID,
	}).Map()
	if linkValue["record_link_id"] != linkID.String() || linkValue["field_key"] != fieldKey || linkValue["confidence"] != &confidence || linkValue["deleted_by_user_id"] != actorID.String() {
		t.Fatalf("unexpected link value: %#v", linkValue)
	}
	if _, err := DecodeRecordLinkMutationValue(linkValue); err != nil {
		t.Fatalf("built link value should decode: %v", err)
	}

	tagValue := BuildRecordTagMutationValue(RecordTagMutationInput{
		RecordTagID:       tagID,
		IncidentID:        incidentID,
		RecordID:          srcID,
		TagName:           "Urgent",
		NormalizedTagName: "urgent",
		CreatedByUserID:   actorID,
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
		DeletedAt:         &deletedAt,
		DeletedByUserID:   &actorID,
	}).Map()
	if tagValue["record_tag_id"] != tagID.String() || tagValue["tag_name"] != "Urgent" || tagValue["deleted_at"] != deletedAt.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected tag value: %#v", tagValue)
	}
	if _, ok := tagValue["tag_id"]; ok {
		t.Fatalf("canonical built tag value should not include legacy tag_id: %#v", tagValue)
	}
	if _, err := DecodeRecordTagMutationValue(tagValue); err != nil {
		t.Fatalf("built tag value should decode: %v", err)
	}
}
