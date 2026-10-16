package reporting

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/links/partyassignment"
	"github.com/google/uuid"
)

func testSubjectAssignmentPermissions(t *testing.T) {
	incident := uuid.MustParse("00000000-0000-0000-0000-000000006001")
	host := uuid.MustParse("00000000-0000-0000-0000-000000006002")
	first := uuid.MustParse("00000000-0000-0000-0000-000000006003")
	second := uuid.MustParse("00000000-0000-0000-0000-000000006004")
	source := fixtureSnapshotSource([]ExportField{
		{SourceRecordID: host.String(), SourceRecordType: "host", SourceFamily: "record_envelope", Path: "/opaque/envelope", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host", "value": "subject metadata"}, DisclosurePartitionRefs: []string{"public"}},
		{SourceRecordID: host.String(), SourceRecordType: "host", SourceFamily: "host", Path: "/opaque/subject", ContentClass: ContentClassSourceEvidence, Value: "private host", DisclosurePartitionRefs: []string{"public"}},
		{SourceRecordID: first.String(), SourceRecordType: "party", SourceFamily: "party", Path: "/opaque/first", ContentClass: ContentClassSourceEvidence, Value: "first private party"},
		{SourceRecordID: second.String(), SourceRecordType: "party", SourceFamily: "party", Path: "/opaque/second", ContentClass: ContentClassSourceEvidence, Value: "second private party"},
		{SourceRecordID: "00000000-0000-0000-0000-000000006009", SourceRecordType: "evidence", SourceFamily: "evidence", Path: "/opaque/citation", ContentClass: ContentClassSourceEvidence, Value: "cites the subject", SupportRefs: []string{"/hosts/" + host.String()}},
	})
	for i, party := range []uuid.UUID{first, second} {
		assignment, err := partyassignment.NewFact(partyassignment.State{
			AssignmentID: uuid.MustParse([]string{"00000000-0000-0000-0000-000000006005", "00000000-0000-0000-0000-000000006006"}[i]),
			IncidentID:   incident, SubjectRecordID: host, SubjectRecordType: "host", PartyID: party,
			AssignmentKind: "subject_party", AssignmentState: "active", Provenance: "manual",
			CreatedByUserID: uuid.MustParse("00000000-0000-0000-0000-000000006007"), CreatedAt: "2026-05-01T00:00:00.000000Z",
		})
		if err != nil {
			t.Fatal(err)
		}
		source.SourceFacts.Assignments = append(source.SourceFacts.Assignments, assignment)
	}
	snapshot, _, err := buildStructuredSnapshotModel(incident.String(), fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", source)
	if err != nil {
		t.Fatal(err)
	}
	fields, err := snapshot.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	refs := fixtureAuthoredSourceRefs(t, snapshot.SnapshotContent, source.Fields)
	before, _ := canonicalJSON(fields)
	partitions := func(fields []ExportField, path string) []string {
		t.Helper()
		for _, field := range fields {
			if field.Path == refs[path] {
				return field.DisclosurePartitionRefs
			}
		}
		t.Fatal("test source missing")
		return nil
	}
	for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal} {
		projected, err := assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, snapshot.SourceFacts, fields, scope)
		if err != nil || !reflect.DeepEqual(partitions(projected, "/opaque/subject"), []string{"public", "party:" + first.String(), "party:" + second.String()}) {
			t.Fatal("active subject assignments did not bind both Parties", scope, err)
		}
		if !reflect.DeepEqual(partitions(projected, "/opaque/envelope/value"), partitions(projected, "/opaque/subject")) {
			t.Fatal("generic envelope bypassed subject permissions")
		}
		if !reflect.DeepEqual(partitions(projected, "/opaque/citation"), partitions(projected, "/opaque/subject")) {
			t.Fatal("citation did not inherit both required Parties")
		}
		if !reflect.DeepEqual(partitions(projected, "/opaque/first"), []string{"party:" + first.String()}) {
			t.Fatal("Party label became public")
		}
		for _, role := range []string{"responsible_party", "observer_party", "none"} {
			facts := snapshot.SourceFacts.Clone()
			if role == "none" {
				facts.Assignments = facts.Assignments[:0]
			} else {
				for i, assignment := range facts.Assignments {
					state := assignment.State
					state.AssignmentKind = role
					facts.Assignments[i], err = partyassignment.NewFact(state)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			projected, err := assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, facts, fields, scope)
			want := "internal_only"
			if scope == ReleaseScopeExternal {
				want = "blocked"
			}
			if err != nil || !reflect.DeepEqual(partitions(projected, "/opaque/subject"), []string{"public", want}) {
				t.Fatal("non-subject relation granted permission", role, scope, err)
			}
			if !reflect.DeepEqual(partitions(projected, "/opaque/envelope/value"), []string{"public", want}) {
				t.Fatal("generic envelope bypassed unassigned subject restrictions")
			}
		}
	}
	after, _ := canonicalJSON(fields)
	if !bytes.Equal(before, after) {
		t.Fatal("disclosure derivation mutated frozen values")
	}
	// Exercise the production redaction entrypoint, not only the assignment
	// helper: an allow rule cannot release an unassigned external subject.
	recipients := []string{"party:" + first.String(), "party:" + second.String()}
	render, sourceDigest, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeExternal, recipients)
	if err != nil {
		t.Fatal(err)
	}
	profile, profileDigest, err := ResolveRedactionProfile(ExternalRedactionProfileID, DefaultRedactionProfileVersion, recipients, snapshot.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(render, snapshot.SourceFacts, profile, profileDigest, sourceDigest, ReleaseScopeExternal, recipients)
	if err != nil || len(result.Model.Fields) != 6 {
		t.Fatal("assigned subject could not be released to every required Party", err)
	}
	foundSubject := false
	for _, entry := range result.Manifest.Entries {
		if entry.Path != *snapshotSourceIndex(snapshot.SnapshotContent)[refs["/opaque/subject"]].SourcePath {
			continue
		}
		foundSubject = true
		if !reflect.DeepEqual(entry.DisclosurePartitionRefs, []string{"public", recipients[0], recipients[1]}) {
			t.Fatal("manifest lost derived permissions")
		}
	}
	if !foundSubject {
		t.Fatal("subject manifest entry missing")
	}
	unassignedFacts := snapshot.SourceFacts.Clone()
	unassignedFacts.Assignments = unassignedFacts.Assignments[:0]
	_, err = RedactExportModel(render, unassignedFacts, profile, profileDigest, sourceDigest, ReleaseScopeExternal, recipients)
	var denied *renderValidationError
	if !errors.As(err, &denied) || denied.FailureCode != "redaction_manifest_invalid" || denied.ReasonCode != "disclosure_partition_unresolved" {
		t.Fatal("production redaction ignored missing subject assignments", err)
	}
	// A valid source-directory flag is not the missing explicit profile grant.
	facts := snapshot.SourceFacts.Clone()
	facts.Parties[0].PublicDirectoryEligible = true
	projected, err := assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, facts, fields, ReleaseScopeExternal)
	if err != nil || !reflect.DeepEqual(partitions(projected, "/opaque/first"), []string{"party:" + first.String()}) {
		t.Fatal("directory eligibility granted disclosure", err)
	}
	// Earlier restrictions survive active assignments.
	for i := range fields {
		if fields[i].Path == refs["/opaque/subject"] {
			fields[i].DisclosurePartitionRefs = []string{"blocked"}
		}
	}
	projected, err = assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, snapshot.SourceFacts, fields, ReleaseScopeExternal)
	if err != nil || !reflect.DeepEqual(partitions(projected, "/opaque/subject"), []string{"party:" + first.String(), "party:" + second.String(), "blocked"}) {
		t.Fatal("assignment erased an earlier block", err)
	}
	if !reflect.DeepEqual(partitions(projected, "/opaque/citation"), []string{"public", "party:" + first.String(), "party:" + second.String(), "blocked"}) {
		t.Fatal("citation erased an earlier block", err)
	}
	facts = snapshot.SourceFacts.Clone()
	state := facts.Assignments[0].State
	state.AssignmentID = uuid.MustParse("00000000-0000-0000-0000-000000006008")
	duplicate, err := partyassignment.NewFact(state)
	if err != nil {
		t.Fatal(err)
	}
	facts.Assignments = append(facts.Assignments, duplicate)
	_, err = assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, facts, fields, ReleaseScopeExternal)
	var failure *renderValidationError
	if !errors.As(err, &failure) || failure.ReasonCode != "duplicate_active_party_assignment" {
		t.Fatal("duplicate source assignment accepted", err)
	}
}
