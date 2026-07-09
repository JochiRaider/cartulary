package links

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func RecordRefItemRef(recordID uuid.UUID) string {
	return "record_ref:" + recordID.String()
}

func ParseRecordRefItemRef(itemRef string) (uuid.UUID, error) {
	return uuidFromItemRef(itemRef, "record_ref:")
}

func PartyRefItemRef(partyRecordID uuid.UUID) string {
	return "party_ref:" + partyRecordID.String()
}

func ParsePartyRefItemRef(itemRef string) (uuid.UUID, error) {
	return uuidFromItemRef(itemRef, "party_ref:")
}

func RecordTagItemRef(recordID uuid.UUID, recordTagID uuid.UUID) string {
	return "record_tag:" + recordID.String() + ":" + recordTagID.String()
}

func ParseRecordTagItemRef(itemRef string) (uuid.UUID, uuid.UUID, error) {
	parts := strings.Split(itemRef, ":")
	if len(parts) != 3 || parts[0] != "record_tag" {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	recordID, err := uuid.Parse(parts[1])
	if err != nil || recordID.String() != parts[1] {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	tagID, err := uuid.Parse(parts[2])
	if err != nil || tagID.String() != parts[2] {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	return recordID, tagID, nil
}

func uuidFromItemRef(itemRef string, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	value := strings.TrimPrefix(itemRef, prefix)
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	return parsed, nil
}
