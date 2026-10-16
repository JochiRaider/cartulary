package exportprovider

import (
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/links/partyassignment"
)

const SourceFactsSchemaID = "cartulary.reporting_source_facts.v2"

// A tombstone proves absence from releasable source content without retaining
// deleted values, deletion actors, storage locations or wall-clock guesses.
type DeletedRecordFact struct {
	RecordID   string `json:"record_id"`
	RecordType string `json:"record_type"`
}

// PartyFact contains source-owned disclosure inputs, never a decision derived
// from a rendered label or a recipient request.
type PartyFact struct {
	PartyID                 string `json:"party_id"`
	PartitionSegment        string `json:"party_partition_segment"`
	PublicDirectoryEligible bool   `json:"public_directory_eligible"`
}

// SourceFacts are retained with the snapshot and private admitted render input.
// They must never be copied into a distributable post-redaction model.
type SourceFacts struct {
	SchemaID       string                 `json:"schema_id"`
	Parties        []PartyFact            `json:"parties"`
	Assignments    []partyassignment.Fact `json:"assignments"`
	DeletedRecords []DeletedRecordFact    `json:"deleted_records"`
}

func EmptySourceFacts() SourceFacts {
	return SourceFacts{SchemaID: SourceFactsSchemaID, Parties: []PartyFact{}, Assignments: []partyassignment.Fact{}, DeletedRecords: []DeletedRecordFact{}}
}

func (s SourceFacts) Validate() error {
	invalid := errors.New("invalid reporting source facts")
	if s.SchemaID != SourceFactsSchemaID || s.Parties == nil || s.Assignments == nil || s.DeletedRecords == nil {
		return invalid
	}
	previous := ""
	for _, record := range s.DeletedRecords {
		if record.RecordID == "" || record.RecordID <= previous || !utf8.ValidString(record.RecordID) || utf8.RuneCountInString(record.RecordID) > 128 || strings.TrimSpace(record.RecordID) != record.RecordID {
			return invalid
		}
		for _, r := range record.RecordID {
			if r < 0x20 || r >= 0x7f && r <= 0x9f || strings.ContainsRune("/\\#", r) {
				return invalid
			}
		}
		switch record.RecordType {
		case "timeline_event", "host", "identity", "party", "indicator", "artifact", "task_request", "decision", "evidence", "assessment":
		default:
			return invalid
		}
		previous = record.RecordID
	}
	previous = ""
	for _, party := range s.Parties {
		if party.PartyID == "" || party.PartitionSegment != party.PartyID || party.PartyID <= previous || !utf8.ValidString(party.PartitionSegment) || utf8.RuneCountInString(party.PartitionSegment) > 128 {
			return invalid
		}
		for _, r := range party.PartitionSegment {
			if unicode.IsSpace(r) || unicode.IsControl(r) || strings.ContainsRune(":/\\#", r) {
				return invalid
			}
		}
		previous = party.PartyID
	}
	previous = ""
	seen := map[[3]string]bool{}
	for _, assignment := range s.Assignments {
		id := assignment.AssignmentID.String()
		key := [3]string{assignment.SubjectRecordID.String(), assignment.PartyID.String(), assignment.AssignmentKind}
		if assignment.Validate() != nil || id <= previous || seen[key] {
			return invalid
		}
		previous, seen[key] = id, true
	}
	return nil
}

func (s SourceFacts) Clone() SourceFacts {
	result := EmptySourceFacts()
	result.SchemaID = s.SchemaID
	if s.DeletedRecords == nil {
		result.DeletedRecords = nil
	} else {
		result.DeletedRecords = append(result.DeletedRecords, s.DeletedRecords...)
	}
	if s.Parties == nil {
		result.Parties = nil
	} else {
		result.Parties = append(result.Parties, s.Parties...)
	}
	if s.Assignments == nil {
		result.Assignments = nil
	} else {
		result.Assignments = append(result.Assignments, s.Assignments...)
	}
	// Valid current assignments have no pointer-bearing optional values.
	return result
}

func (s *SourceFacts) Sort() {
	sort.Slice(s.DeletedRecords, func(i, j int) bool { return s.DeletedRecords[i].RecordID < s.DeletedRecords[j].RecordID })
	sort.Slice(s.Parties, func(i, j int) bool { return s.Parties[i].PartyID < s.Parties[j].PartyID })
	sort.Slice(s.Assignments, func(i, j int) bool {
		return s.Assignments[i].AssignmentID.String() < s.Assignments[j].AssignmentID.String()
	})
}
