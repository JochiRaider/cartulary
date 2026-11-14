package entities_test

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestRevisionProviderContributionOwnsEntityHistoryFacet_Unit(t *testing.T) {
	contribution := entities.RevisionProviderContribution()
	if contribution.SourceOwnerModule != revisions.SourceOwnerEntities || len(contribution.Records) != 2 || len(contribution.NonRowTargets) != 3 {
		t.Fatalf("unexpected Entities revision contribution: %#v", contribution)
	}
	wantRecords := map[string]string{
		"host":     "cartulary.revisions.snapshot.host.v1",
		"identity": "cartulary.revisions.snapshot.identity.v1",
	}
	seenRecords := map[string]bool{}
	for _, record := range contribution.Records {
		wantSchema, known := wantRecords[record.RecordType]
		if !known || seenRecords[record.RecordType] || record.SnapshotSchemaID != wantSchema || !slices.Equal(record.HistoryTargetKinds, []string{record.RecordType}) {
			t.Fatalf("unexpected %s record semantics: %#v", record.RecordType, record)
		}
		seenRecords[record.RecordType] = true
		t.Run(record.RecordType+" historical source", func(t *testing.T) { requireEntitySnapshotSemantics(t, record) })
	}

	sourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	entityID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	seenTargets := map[string]bool{}
	for _, target := range contribution.NonRowTargets {
		t.Run(target.TargetKind+" references", func(t *testing.T) { requireRetainedHistoryReferences(t, target) })
		if target.TargetKind == "entity_mention" {
			t.Run("canonical mention history", func(t *testing.T) { requireCanonicalMentionHistory(t, target) })
		}
		if target.TargetKind == "entity_preserved_identifier" {
			t.Run("canonical preserved identifier history", func(t *testing.T) { requireCanonicalPreservedIdentifierHistory(t, target) })
		}
		if target.TargetKind == "entity_alias" {
			t.Run("canonical alias history", func(t *testing.T) { requireCanonicalAliasHistory(t, target) })
		}
		if seenTargets[target.TargetKind] || target.RollbackProvider == nil {
			t.Fatalf("target %q has incomplete semantics", target.TargetKind)
		}
		seenTargets[target.TargetKind] = true
		mutation := revisions.StoredMutation{TargetKind: target.TargetKind, TargetID: uuid.NewString()}
		wantRecordID := entityID
		var wantEntryIDs []uuid.UUID
		switch target.TargetKind {
		case "entity_mention":
			mutation.AfterValue = map[string]any{
				"source_record_id":   sourceID.String(),
				"resolved_record_id": entityID.String(),
			}
			wantRecordID = sourceID
			wantEntryIDs = []uuid.UUID{sourceID}
		case "entity_alias", "entity_preserved_identifier":
			mutation.AfterValue = map[string]any{"record_id": entityID.String()}
		default:
			t.Fatalf("unexpected Entities target %q", target.TargetKind)
		}
		description, err := target.HistoryFacet.DescribeMutation(mutation)
		if err != nil {
			t.Fatalf("describe %s: %v", target.TargetKind, err)
		}
		if !slices.Equal(description.HistoryRecordIDs, []uuid.UUID{wantRecordID}) {
			t.Fatalf("unexpected %s record associations: %v", target.TargetKind, description.HistoryRecordIDs)
		}
		if !slices.Equal(description.HistoryEntryRecordIDs, wantEntryIDs) {
			t.Fatalf("unexpected %s addressability: %v", target.TargetKind, description.HistoryEntryRecordIDs)
		}
	}
}

func requireEntitySnapshotSemantics(t *testing.T, contribution revisions.RecordProviderContribution) {
	t.Helper()
	record := map[string]any{
		"record_id": "11111111-1111-4111-8111-111111111111", "incident_id": "22222222-2222-4222-8222-222222222222",
		"row_version": json.Number("3"), "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
		"created_by_user_id": "33333333-3333-4333-8333-333333333333", "updated_by_user_id": "44444444-4444-4444-8444-444444444444",
	}
	base := maps.Clone(record)
	base["created_at"], base["updated_at"] = "2026-01-01T00:00:00+00:00", "2026-01-02T00:00:00+00:00"
	stateKey := contribution.RecordType + "_state"
	base["display_name"], base["entity_origin"], base[stateKey] = "Entity", "entity_sheet", "canonical"
	base["seed_entity_mention_id"], base["merged_into_record_id"] = nil, nil
	identifiers := []string{"hostname", "aad_device_id", "fqdn"}
	annotations := []string{"location", "os_platform", "business_owner", "criticality", "containment_status"}
	if contribution.RecordType == "identity" {
		identifiers = []string{"upn", "email", "sam_account_name", "aad_object_id", "sid"}
		annotations = []string{"privilege_level", "mfa_state", "reset_status"}
	}
	for _, field := range identifiers {
		base[field] = nil
	}
	for _, field := range annotations {
		base[field] = ""
	}
	validate := contribution.ValidateSnapshotSource
	if validate == nil {
		t.Fatal("source validator absent")
	}
	for _, state := range []string{"stub", "canonical", "merged"} {
		for _, origin := range []string{"entity_sheet", "entity_import", "created_from_mention", "system_upsert"} {
			value := maps.Clone(base)
			value[stateKey], value["entity_origin"] = state, origin
			if state == "merged" {
				value["merged_into_record_id"] = "55555555-5555-4555-8555-555555555555"
			}
			before, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if err := validate(record, value); err != nil {
				t.Fatalf("valid %s/%s rejected: %v", state, origin, err)
			}
			after, err := json.Marshal(value)
			if err != nil || string(after) != string(before) {
				t.Fatal("retained source changed")
			}
		}
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"empty display", func(v map[string]any) { v["display_name"] = "" }},
		{"NUL display", func(v map[string]any) { v["display_name"] = "bad\x00" }},
		{"unknown origin", func(v map[string]any) { v["entity_origin"] = "unknown" }},
		{"unknown state", func(v map[string]any) { v[stateKey] = "active" }},
		{"active target", func(v map[string]any) { v["merged_into_record_id"] = "55555555-5555-4555-8555-555555555555" }},
		{"missing merged target", func(v map[string]any) { v[stateKey] = "merged" }},
		{"self merge", func(v map[string]any) { v[stateKey] = "merged"; v["merged_into_record_id"] = v["record_id"] }},
		{"zero version", func(v map[string]any) { v["row_version"] = json.Number("0") }},
		{"different version", func(v map[string]any) { v["row_version"] = json.Number("4") }},
		{"reversed clocks", func(v map[string]any) { v["updated_at"] = "1999-01-01T00:00:00Z" }},
	}
	for _, field := range []string{"created_at", "updated_at"} {
		cases = append(cases, struct {
			name   string
			mutate func(map[string]any)
		}{"mismatched " + field, func(v map[string]any) { v[field] = "2026-01-01T12:00:00Z" }})
	}
	for _, field := range []string{"created_by_user_id", "updated_by_user_id"} {
		cases = append(cases, struct {
			name   string
			mutate func(map[string]any)
		}{"mismatched " + field, func(v map[string]any) { v[field] = "55555555-5555-4555-8555-555555555555" }})
	}
	for _, field := range identifiers {
		cases = append(cases, struct {
			name   string
			mutate func(map[string]any)
		}{"empty " + field, func(v map[string]any) { v[field] = "" }})
	}
	for _, field := range append(identifiers, annotations...) {
		cases = append(cases, struct {
			name   string
			mutate func(map[string]any)
		}{"NUL " + field, func(v map[string]any) { v[field] = "bad\x00" }})
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			value := maps.Clone(base)
			test.mutate(value)
			before, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if err := validate(record, value); err == nil {
				t.Fatal("invalid historical source accepted")
			}
			after, err := json.Marshal(value)
			if err != nil || string(after) != string(before) {
				t.Fatal("invalid history was repaired")
			}
		})
	}
}

// This checks only reference projection, independently of full value admission.
func requireRetainedHistoryReferences(t *testing.T, target revisions.NonRowProviderContribution) {
	t.Helper()
	if target.HistoryReferences == nil {
		t.Fatal("reference projector missing")
	}
	firstIncident := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	secondIncident := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	firstRecord := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	secondRecord := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	required, nullable, explicitIncident := "", "", true
	switch target.TargetKind {
	case "entity_mention":
		required, nullable, explicitIncident = "source_record_id", "resolved_record_id", false
	case "entity_alias", "entity_preserved_identifier":
		required = "record_id"
	default:
		t.Fatalf("unexpected target %s", target.TargetKind)
	}
	makeMutation := func() revisions.StoredMutation {
		before := map[string]any{required: firstRecord.String()}
		after := map[string]any{required: secondRecord.String()}
		if explicitIncident {
			before["incident_id"], after["incident_id"] = firstIncident.String(), secondIncident.String()
		}
		if nullable != "" {
			before[nullable], after[nullable] = secondRecord.String(), firstRecord.String()
		}
		return revisions.StoredMutation{TargetKind: target.TargetKind, BeforeValue: before, AfterValue: after}
	}
	scopeAfter := secondIncident
	if !explicitIncident {
		scopeAfter = firstIncident
	}
	want := []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}}
	if nullable != "" {
		want = append(want, revisions.HistoryRecordReference{IncidentID: firstIncident, RecordID: secondRecord})
	}
	want = append(want, revisions.HistoryRecordReference{IncidentID: scopeAfter, RecordID: secondRecord})
	if nullable != "" {
		want = append(want, revisions.HistoryRecordReference{IncidentID: scopeAfter, RecordID: firstRecord})
	}
	input := makeMutation()
	unchanged := func(t *testing.T, before []byte, input revisions.StoredMutation) {
		t.Helper()
		after, err := json.Marshal(input)
		if err != nil || string(before) != string(after) {
			t.Fatal("reference projection changed retained values")
		}
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := target.HistoryReferences(input, firstIncident)
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("references=%v want=%v error=%v", got, want, err)
	}
	unchanged(t, bytes, input)
	if nullable != "" {
		input.BeforeValue[nullable], input.AfterValue[nullable] = nil, nil
		got, err := target.HistoryReferences(input, firstIncident)
		expected := []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}, {IncidentID: scopeAfter, RecordID: secondRecord}}
		if err != nil || !slices.Equal(got, expected) {
			t.Fatalf("nullable references=%v error=%v", got, err)
		}
	}
	fields := []string{required}
	if explicitIncident {
		fields = append(fields, "incident_id")
	}
	if nullable != "" {
		fields = append(fields, nullable)
	}
	for _, side := range []string{"before", "after"} {
		for _, field := range fields {
			for _, defect := range []string{"missing", "null", "type", "zero", "upper", "compact", "whitespace", "empty"} {
				if field == nullable && defect == "null" {
					continue
				}
				t.Run(side+"/"+field+"/"+defect, func(t *testing.T) {
					input := makeMutation()
					value := input.BeforeValue
					if side == "after" {
						value = input.AfterValue
					}
					switch defect {
					case "missing":
						delete(value, field)
					case "null":
						value[field] = nil
					case "type":
						value[field] = 7
					case "zero":
						value[field] = uuid.Nil.String()
					case "upper":
						value[field] = strings.ToUpper(firstRecord.String())
					case "compact":
						value[field] = strings.ReplaceAll(firstRecord.String(), "-", "")
					case "whitespace":
						value[field] = " " + firstRecord.String()
					case "empty":
						value[field] = ""
					}
					bytes, err := json.Marshal(input)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := target.HistoryReferences(input, firstIncident); err == nil {
						t.Fatal("invalid reference accepted")
					}
					unchanged(t, bytes, input)
				})
			}
		}
	}
	if got, err := target.HistoryReferences(revisions.StoredMutation{TargetKind: target.TargetKind}, firstIncident); err != nil || len(got) != 0 {
		t.Fatalf("null sides=%v error=%v", got, err)
	}
}
