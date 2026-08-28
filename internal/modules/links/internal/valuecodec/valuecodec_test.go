package valuecodec

import (
	"maps"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	testLinkID     = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testTagID      = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	testIncidentID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testSourceID   = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	testDestID     = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	testActorID    = uuid.MustParse("66666666-6666-4666-8666-666666666666")
	testCreatedAt  = time.Date(2026, 7, 9, 20, 0, 0, 123, time.UTC)
)

func TestDecodeRecordMutationValuesRejectLegacyShapes(t *testing.T) {
	tests := []struct {
		name   string
		decode func(map[string]any) error
		value  map[string]any
	}{
		{
			name: "record tag alias",
			decode: func(value map[string]any) error {
				_, err := DecodeRecordTagMutationValue(value)
				return err
			},
			value: map[string]any{
				"tag_id": testTagID.String(), "incident_id": testIncidentID.String(),
				"record_id": testSourceID.String(), "tag_name": "Urgent", "normalized_tag_name": "urgent",
			},
		},
		{
			name: "compact record link",
			decode: func(value map[string]any) error {
				_, err := DecodeRecordLinkMutationValue(value)
				return err
			},
			value: map[string]any{
				"record_link_id": testLinkID.String(), "incident_id": testIncidentID.String(),
				"src_record_id": testSourceID.String(), "dst_record_id": testDestID.String(),
				"link_type": "references_record",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.decode(test.value); err == nil {
				t.Fatal("legacy mutation value was accepted")
			}
		})
	}
}

func TestDecodeRecordLinkMutationValueStrictNegativeMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing nullable member", mutate: func(value map[string]any) { delete(value, "field_key") }},
		{name: "unknown member", mutate: func(value map[string]any) { value["comment"] = "legacy" }},
		{name: "mistyped confidence", mutate: func(value map[string]any) { value["confidence"] = "100" }},
		{name: "noncanonical uuid", mutate: func(value map[string]any) { value["record_link_id"] = "11111111-1111-4111-8111-11111111111A" }},
		{name: "noncanonical timestamp", mutate: func(value map[string]any) { value["created_at"] = "2026-07-09T16:00:00.000000123-04:00" }},
		{name: "unknown link type", mutate: func(value map[string]any) { value["link_type"] = "legacy_link" }},
		{name: "unknown provenance", mutate: func(value map[string]any) { value["provenance"] = "legacy" }},
		{name: "manual confidence", mutate: func(value map[string]any) { value["confidence"] = 100 }},
		{name: "partial deletion tuple", mutate: func(value map[string]any) {
			value["deleted_at"] = testCreatedAt.Add(time.Minute).Format(time.RFC3339Nano)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := canonicalLinkValue()
			test.mutate(value)
			if _, err := DecodeRecordLinkMutationValue(value); err == nil {
				t.Fatal("noncanonical link value was accepted")
			}
		})
	}
}

func TestDecodeRecordTagMutationValueStrictNegativeMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing nullable member", mutate: func(value map[string]any) { delete(value, "deleted_at") }},
		{name: "unknown member", mutate: func(value map[string]any) { value["tag_id"] = testTagID.String() }},
		{name: "mistyped nullable member", mutate: func(value map[string]any) { value["deleted_by_user_id"] = 17 }},
		{name: "noncanonical timestamp", mutate: func(value map[string]any) { value["updated_at"] = "2026-07-09T20:00:00.123000000Z" }},
		{name: "partial deletion tuple", mutate: func(value map[string]any) { value["deleted_at"] = testCreatedAt.Format(time.RFC3339Nano) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := canonicalTagValue()
			test.mutate(value)
			if _, err := DecodeRecordTagMutationValue(value); err == nil {
				t.Fatal("noncanonical tag value was accepted")
			}
		})
	}
}

func TestHistoryMutationOperationMatrix(t *testing.T) {
	activeLink := canonicalLinkValue()
	patchedLink := maps.Clone(activeLink)
	patchedLink["provenance"] = "system"
	patchedLink["owner_user_id"] = testDestID.String()
	patchedLink["decided_at"] = testCreatedAt.Add(time.Minute).Format(time.RFC3339Nano)
	deletedLink := maps.Clone(patchedLink)
	deletedLink["deleted_at"] = testCreatedAt.Add(2 * time.Minute).Format(time.RFC3339Nano)
	deletedLink["deleted_by_user_id"] = testActorID.String()

	activeTag := canonicalTagValue()
	patchedTag := maps.Clone(activeTag)
	patchedTag["record_id"] = testDestID.String()
	patchedTag["updated_at"] = testCreatedAt.Add(time.Minute).Format(time.RFC3339Nano)
	deletedTag := maps.Clone(patchedTag)
	deletedTag["updated_at"] = testCreatedAt.Add(2 * time.Minute).Format(time.RFC3339Nano)
	deletedTag["deleted_at"] = deletedTag["updated_at"]
	deletedTag["deleted_by_user_id"] = testActorID.String()

	tests := []struct {
		name      string
		kind      string
		targetID  string
		operation string
		before    map[string]any
		after     map[string]any
		wantError bool
	}{
		{name: "link create", kind: "record_link", targetID: testLinkID.String(), operation: "create", after: activeLink},
		{name: "link patch", kind: "record_link", targetID: testLinkID.String(), operation: "patch", before: activeLink, after: patchedLink},
		{name: "link delete", kind: "record_link", targetID: testLinkID.String(), operation: "delete", before: patchedLink, after: deletedLink},
		{name: "link rollback create", kind: "record_link", targetID: testLinkID.String(), operation: "rollback", before: patchedLink, after: deletedLink},
		{name: "link rollback delete", kind: "record_link", targetID: testLinkID.String(), operation: "rollback", before: deletedLink, after: patchedLink},
		{name: "link no-op patch", kind: "record_link", targetID: testLinkID.String(), operation: "patch", before: activeLink, after: maps.Clone(activeLink), wantError: true},
		{name: "link missing side", kind: "record_link", targetID: testLinkID.String(), operation: "patch", after: patchedLink, wantError: true},
		{name: "tag create", kind: "record_tag", targetID: recordTagTargetID(testSourceID, testTagID), operation: "create", after: activeTag},
		{name: "tag patch", kind: "record_tag", targetID: recordTagTargetID(testSourceID, testTagID), operation: "patch", before: activeTag, after: patchedTag},
		{name: "tag delete", kind: "record_tag", targetID: recordTagTargetID(testDestID, testTagID), operation: "delete", before: patchedTag, after: deletedTag},
		{name: "tag rollback", kind: "record_tag", targetID: recordTagTargetID(testDestID, testTagID), operation: "rollback", before: deletedTag, after: patchedTag},
		{name: "tag bare uuid", kind: "record_tag", targetID: testTagID.String(), operation: "patch", before: activeTag, after: patchedTag, wantError: true},
		{name: "tag after-addressed patch", kind: "record_tag", targetID: recordTagTargetID(testDestID, testTagID), operation: "patch", before: activeTag, after: patchedTag, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateHistoryMutation(test.kind, test.targetID, test.operation, test.before, test.after)
			if (err != nil) != test.wantError {
				t.Fatalf("validation error = %v, wantError=%t", err, test.wantError)
			}
		})
	}
}

func TestDecodedMutationValuesRetainEveryCanonicalField(t *testing.T) {
	link, err := DecodeRecordLinkMutationValue(canonicalLinkValue())
	if err != nil {
		t.Fatalf("decode link restore plan: %v", err)
	}
	if link.RecordLinkID != testLinkID || link.OwnerUserID != testActorID || link.CreatedByUserID != testActorID ||
		!link.DecidedAt.Equal(testCreatedAt) || !link.CreatedAt.Equal(testCreatedAt) || link.Provenance != "manual" ||
		link.FieldKey != nil || link.Confidence != nil || link.DeletedAt != nil || link.DeletedByUserID != nil {
		t.Fatalf("decoded link value lost canonical state: %#v", link)
	}

	tag, err := DecodeRecordTagMutationValue(canonicalTagValue())
	if err != nil {
		t.Fatalf("decode tag restore plan: %v", err)
	}
	if tag.RecordTagID != testTagID || tag.RecordID != testSourceID || tag.CreatedByUserID != testActorID ||
		tag.TagName != "Urgent" || tag.NormalizedTagName != "urgent" || !tag.CreatedAt.Equal(testCreatedAt) ||
		!tag.UpdatedAt.Equal(testCreatedAt) || tag.DeletedAt != nil || tag.DeletedByUserID != nil {
		t.Fatalf("decoded tag value lost canonical state: %#v", tag)
	}
}

func TestBuildRecordMutationValuesEmitCanonicalFreshMaps(t *testing.T) {
	link := BuildRecordLinkMutationValue(RecordLinkMutationInput{
		RecordLinkID: testLinkID, IncidentID: testIncidentID, SrcRecordID: testSourceID,
		DstRecordID: testDestID, LinkType: "references_record", Provenance: "manual",
		OwnerUserID: testActorID, CreatedByUserID: testActorID, DecidedAt: testCreatedAt, CreatedAt: testCreatedAt,
	})
	first := link.Map()
	first["record_link_id"] = "mutated"
	second := link.Map()
	if second["record_link_id"] != testLinkID.String() || len(second) != len(recordLinkMembers) {
		t.Fatalf("link map was not canonical and fresh: %#v", second)
	}
	if _, err := DecodeRecordLinkMutationValue(second); err != nil {
		t.Fatalf("decode built link value: %v", err)
	}

	tag := BuildRecordTagMutationValue(RecordTagMutationInput{
		RecordTagID: testTagID, IncidentID: testIncidentID, RecordID: testSourceID,
		TagName: "Urgent", NormalizedTagName: "urgent", CreatedByUserID: testActorID,
		CreatedAt: testCreatedAt, UpdatedAt: testCreatedAt,
	})
	firstTag := tag.Map()
	firstTag["tag_name"] = "mutated"
	secondTag := tag.Map()
	if secondTag["tag_name"] != "Urgent" || len(secondTag) != len(recordTagMembers) {
		t.Fatalf("tag map was not canonical and fresh: %#v", secondTag)
	}
	if _, err := DecodeRecordTagMutationValue(secondTag); err != nil {
		t.Fatalf("decode built tag value: %v", err)
	}
}

func canonicalLinkValue() map[string]any {
	return BuildRecordLinkMutationValue(RecordLinkMutationInput{
		RecordLinkID: testLinkID, IncidentID: testIncidentID, SrcRecordID: testSourceID,
		DstRecordID: testDestID, LinkType: "references_record", Provenance: "manual",
		OwnerUserID: testActorID, CreatedByUserID: testActorID, DecidedAt: testCreatedAt, CreatedAt: testCreatedAt,
	}).Map()
}

func canonicalTagValue() map[string]any {
	return BuildRecordTagMutationValue(RecordTagMutationInput{
		RecordTagID: testTagID, IncidentID: testIncidentID, RecordID: testSourceID,
		TagName: "Urgent", NormalizedTagName: "urgent", CreatedByUserID: testActorID,
		CreatedAt: testCreatedAt, UpdatedAt: testCreatedAt,
	}).Map()
}
