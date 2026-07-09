package links

import (
	"testing"

	"github.com/google/uuid"
)

func TestLinksItemRefsAreCanonical(t *testing.T) {
	recordID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	tagID := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	partyID := uuid.MustParse("10000000-0000-4000-8000-000000000003")

	if got := RecordRefItemRef(recordID); got != "record_ref:10000000-0000-4000-8000-000000000001" {
		t.Fatalf("record ref = %q", got)
	}
	if got := PartyRefItemRef(partyID); got != "party_ref:10000000-0000-4000-8000-000000000003" {
		t.Fatalf("party ref = %q", got)
	}
	if got := RecordTagItemRef(recordID, tagID); got != "record_tag:10000000-0000-4000-8000-000000000001:10000000-0000-4000-8000-000000000002" {
		t.Fatalf("record tag ref = %q", got)
	}
}

func TestLinksItemRefParsingIsStrict(t *testing.T) {
	recordID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	tagID := uuid.MustParse("10000000-0000-4000-8000-000000000002")

	if got, err := ParseRecordRefItemRef(RecordRefItemRef(recordID)); err != nil || got != recordID {
		t.Fatalf("parse record ref got %s err %v", got, err)
	}
	if got, err := ParsePartyRefItemRef(PartyRefItemRef(recordID)); err != nil || got != recordID {
		t.Fatalf("parse party ref got %s err %v", got, err)
	}
	if gotRecordID, gotTagID, err := ParseRecordTagItemRef(RecordTagItemRef(recordID, tagID)); err != nil || gotRecordID != recordID || gotTagID != tagID {
		t.Fatalf("parse record tag got record=%s tag=%s err %v", gotRecordID, gotTagID, err)
	}

	invalidRecordRefs := []string{
		"",
		"party_ref:" + recordID.String(),
		"record_ref:" + recordID.String() + ":extra",
		"record_ref:" + recordID.String() + " ",
	}
	for _, itemRef := range invalidRecordRefs {
		if _, err := ParseRecordRefItemRef(itemRef); err == nil {
			t.Fatalf("ParseRecordRefItemRef(%q) succeeded", itemRef)
		}
	}

	invalidPartyRefs := []string{
		"",
		"record_ref:" + recordID.String(),
		"party_ref:" + recordID.String() + ":extra",
		"party_ref:" + recordID.String() + " ",
	}
	for _, itemRef := range invalidPartyRefs {
		if _, err := ParsePartyRefItemRef(itemRef); err == nil {
			t.Fatalf("ParsePartyRefItemRef(%q) succeeded", itemRef)
		}
	}

	invalidTagRefs := []string{
		"",
		"record_tag:" + recordID.String(),
		"record_tag:" + recordID.String() + ":" + tagID.String() + ":extra",
		"record_tag:" + recordID.String() + " :" + tagID.String(),
		"record_tag:client:" + tagID.String(),
	}
	for _, itemRef := range invalidTagRefs {
		if _, _, err := ParseRecordTagItemRef(itemRef); err == nil {
			t.Fatalf("ParseRecordTagItemRef(%q) succeeded", itemRef)
		}
	}
}
