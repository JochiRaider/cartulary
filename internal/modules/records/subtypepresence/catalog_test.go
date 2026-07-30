package subtypepresence

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type testSource struct {
	types []RecordType
}

func (s *testSource) SupportedRecordTypes() []RecordType {
	if s == nil {
		return nil
	}
	return append([]RecordType(nil), s.types...)
}

func (*testSource) ListSubtypeBindingsTx(context.Context, pgx.Tx, uuid.UUID) ([]Binding, error) {
	return nil, nil
}

func TestCatalogRejectsInvalidContributions(t *testing.T) {
	valid := validContributions()
	var typedNil *testSource
	cases := map[string][]Contribution{
		"missing":        valid[:len(valid)-1],
		"duplicate":      append(append([]Contribution(nil), valid...), valid[0]),
		"unknown_family": append(append([]Contribution(nil), valid...), Contribution{FamilyID: "other", Source: &testSource{types: []RecordType{RecordTypeHost}}}),
		"unknown_type":   replaceContribution(valid, "entities", Contribution{FamilyID: "entities", Source: &testSource{types: []RecordType{RecordType("other")}}}),
		"wrong_family":   replaceContribution(valid, "entities", Contribution{FamilyID: "entities", Source: &testSource{types: []RecordType{RecordTypeParty}}}),
		"duplicate_type": replaceContribution(valid, "entities", Contribution{FamilyID: "entities", Source: &testSource{types: []RecordType{RecordTypeHost, RecordTypeHost, RecordTypeIdentity}}}),
		"typed_nil":      replaceContribution(valid, "entities", Contribution{FamilyID: "entities", Source: typedNil}),
	}
	for name, contributions := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewCatalog(contributions); !errors.Is(err, ErrInvalidCatalog) {
				t.Fatalf("catalog error = %v; want ErrInvalidCatalog", err)
			}
		})
	}
}

func TestCatalogAcceptsClosedTenTypeCoverage(t *testing.T) {
	catalog, err := NewCatalog(validContributions())
	if err != nil {
		t.Fatalf("new valid catalog: %v", err)
	}
	if err := catalog.ValidateContract(); err != nil {
		t.Fatalf("validate catalog: %v", err)
	}
}

func TestValidateBindingsRequiresExactForwardAndReverseCoverage(t *testing.T) {
	incidentID := uuid.New()
	recordID := uuid.New()
	envelopes := []Envelope{{
		RecordID: recordID, IncidentID: incidentID, RecordType: RecordTypeHost,
	}}
	valid := []Binding{{
		RecordID: recordID, IncidentID: incidentID, RecordType: RecordTypeHost,
	}}
	if err := validateBindings(incidentID, envelopes, valid); err != nil {
		t.Fatalf("validate exact bindings: %v", err)
	}
	cases := map[string][]Binding{
		"missing":           nil,
		"duplicate":         append(append([]Binding(nil), valid...), valid[0]),
		"unknown":           {{RecordID: uuid.New(), IncidentID: incidentID, RecordType: RecordTypeHost}},
		"wrong_incident":    {{RecordID: recordID, IncidentID: uuid.New(), RecordType: RecordTypeHost}},
		"incompatible_type": {{RecordID: recordID, IncidentID: incidentID, RecordType: RecordTypeIdentity}},
	}
	for name, bindings := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateBindings(incidentID, envelopes, bindings); !errors.Is(err, ErrIncompleteBindings) {
				t.Fatalf("binding error = %v; want ErrIncompleteBindings", err)
			}
		})
	}
}

func validContributions() []Contribution {
	return []Contribution{
		{FamilyID: "timeline", Source: &testSource{types: []RecordType{RecordTypeTimelineEvent}}},
		{FamilyID: "entities", Source: &testSource{types: []RecordType{RecordTypeHost, RecordTypeIdentity}}},
		{FamilyID: "parties", Source: &testSource{types: []RecordType{RecordTypeParty}}},
		{FamilyID: "indicators", Source: &testSource{types: []RecordType{RecordTypeIndicator}}},
		{FamilyID: "artifacts", Source: &testSource{types: []RecordType{RecordTypeArtifact}}},
		{FamilyID: "tasks_decisions", Source: &testSource{types: []RecordType{RecordTypeTaskRequest, RecordTypeDecision}}},
		{FamilyID: "evidence", Source: &testSource{types: []RecordType{RecordTypeEvidence}}},
		{FamilyID: "assessments", Source: &testSource{types: []RecordType{RecordTypeAssessment}}},
	}
}

func replaceContribution(
	contributions []Contribution,
	familyID string,
	replacement Contribution,
) []Contribution {
	result := append([]Contribution(nil), contributions...)
	for index := range result {
		if result[index].FamilyID == familyID {
			result[index] = replacement
			return result
		}
	}
	return result
}
