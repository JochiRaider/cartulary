package subtypepresence

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type testSource struct {
	types    []RecordType
	bindings []Binding
	err      error
}

func (s *testSource) SupportedRecordTypes() []RecordType {
	if s == nil {
		return nil
	}
	return append([]RecordType(nil), s.types...)
}

func (s *testSource) VisitSubtypeBindingsTx(_ context.Context, _ pgx.Tx, _ uuid.UUID, visit func(Binding) error) error {
	if s.err != nil {
		return s.err
	}
	for _, binding := range s.bindings {
		if err := visit(binding); err != nil {
			return err
		}
	}
	return nil
}

// catalogTestTx is a non-nil transaction for sources that never use it.
type catalogTestTx struct{ pgx.Tx }

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
	t.Run("production catalog ValidateTx", checkProductionCatalogValidateTx)
	incidentID := uuid.New()
	recordID := uuid.New()
	envelopes := []Binding{{
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

func validateBindings(incidentID uuid.UUID, envelopes, bindings []Binding) error {
	expected, actual := NewInventory(incidentID), NewInventory(incidentID)
	for _, value := range envelopes {
		if err := expected.Add(value); err != nil {
			return err
		}
	}
	for _, value := range bindings {
		if err := actual.Add(value); err != nil {
			return err
		}
	}
	if !expected.Summary().Equal(actual.Summary()) {
		return ErrIncompleteBindings
	}
	return nil
}

func TestInventoryStreamsExactOrderedBindings(t *testing.T) {
	incidentID, foreignIncidentID := uuid.New(), uuid.New()
	expected, actual := NewInventory(incidentID), NewInventory(incidentID)
	// foreign streams the same record IDs and types under another incident, so
	// the incident is the only difference from expected.
	foreign := NewInventory(foreignIncidentID)
	var last Binding
	for n := uint64(1); n <= 70000; n++ {
		id := uuid.UUID{}
		binary.BigEndian.PutUint64(id[8:], n)
		kind := RecordTypeHost
		if n%2 == 0 {
			kind = RecordTypeIdentity
		}
		last = Binding{RecordID: id, IncidentID: incidentID, RecordType: kind}
		if err := expected.Add(last); err != nil {
			t.Fatal(err)
		}
		if err := foreign.Add(Binding{RecordID: id, IncidentID: foreignIncidentID, RecordType: kind}); err != nil {
			t.Fatal(err)
		}
		if n < 70000 {
			if err := actual.Add(last); err != nil {
				t.Fatal(err)
			}
		}
	}
	frozen := expected.Summary()
	foreignSummary := foreign.Summary()
	for kind := range expectedFamilyByType {
		if frozen.types[kind].count != foreignSummary.types[kind].count {
			t.Fatalf("%s count = %d; foreign count = %d", kind, frozen.types[kind].count, foreignSummary.types[kind].count)
		}
	}
	if frozen.Equal(foreignSummary) || foreignSummary.Equal(frozen) {
		t.Fatal("cross-incident inventory with identical bindings accepted")
	}
	if frozen.Equal(actual.Summary()) {
		t.Fatal("missing final binding was accepted")
	}
	if err := actual.Add(last); err != nil {
		t.Fatal(err)
	}
	if !frozen.Equal(actual.Summary()) {
		t.Fatal("identical streams differ")
	}
	if err := actual.Add(last); !errors.Is(err, ErrIncompleteBindings) {
		t.Fatal("duplicate accepted", err)
	}
	earlier := last
	earlier.RecordID[15]--
	if err := actual.Add(earlier); !errors.Is(err, ErrIncompleteBindings) {
		t.Fatal("unordered binding accepted", err)
	}
	last.RecordID[0] = 1
	if err := expected.Add(last); err != nil {
		t.Fatal(err)
	}
	if frozen.Equal(expected.Summary()) {
		t.Fatal("summary changed with its source or extra binding was ignored")
	}
}

func checkProductionCatalogValidateTx(t *testing.T) {
	ctx := context.Background()
	incidentID := uuid.New()
	host := Binding{RecordID: uuid.MustParse("10000000-0000-4000-8000-000000000001"), IncidentID: incidentID, RecordType: RecordTypeHost}
	party := Binding{RecordID: uuid.MustParse("10000000-0000-4000-8000-000000000002"), IncidentID: incidentID, RecordType: RecordTypeParty}
	build := func(entities, parties []Binding, visitErr error) *Catalog {
		t.Helper()
		contributions := replaceContribution(validContributions(), "entities", Contribution{FamilyID: "entities", Source: &testSource{types: []RecordType{RecordTypeHost, RecordTypeIdentity}, bindings: entities, err: visitErr}})
		contributions = replaceContribution(contributions, "parties", Contribution{FamilyID: "parties", Source: &testSource{types: []RecordType{RecordTypeParty}, bindings: parties}})
		catalog, err := NewCatalog(contributions)
		if err != nil {
			t.Fatalf("NewCatalog: %v", err)
		}
		return catalog
	}
	expectedInventory := NewInventory(incidentID)
	for _, binding := range []Binding{host, party} {
		if err := expectedInventory.Add(binding); err != nil {
			t.Fatal(err)
		}
	}
	expected := expectedInventory.Summary()
	tx := catalogTestTx{}
	if err := build([]Binding{host}, []Binding{party}, nil).ValidateTx(ctx, tx, incidentID, expected); err != nil {
		t.Fatalf("complete bindings rejected: %v", err)
	}
	if err := build(nil, []Binding{party}, nil).ValidateTx(ctx, tx, incidentID, expected); !errors.Is(err, ErrIncompleteBindings) {
		t.Fatalf("missing binding error = %v", err)
	}
	if err := build([]Binding{host, party}, nil, nil).ValidateTx(ctx, tx, incidentID, expected); !errors.Is(err, ErrIncompleteBindings) {
		t.Fatalf("wrong-family binding error = %v", err)
	}
	visitFailure := errors.New("visit failed")
	if err := build(nil, []Binding{party}, visitFailure).ValidateTx(ctx, tx, incidentID, expected); !errors.Is(err, visitFailure) {
		t.Fatalf("visitor error = %v", err)
	}
	catalog := build([]Binding{host}, []Binding{party}, nil)
	for name, call := range map[string]func() error{
		"nil catalog":     func() error { return (*Catalog)(nil).ValidateTx(ctx, tx, incidentID, expected) },
		"nil transaction": func() error { return catalog.ValidateTx(ctx, nil, incidentID, expected) },
		"nil incident":    func() error { return catalog.ValidateTx(ctx, tx, uuid.Nil, expected) },
		"foreign summary": func() error { return catalog.ValidateTx(ctx, tx, incidentID, NewInventory(uuid.New()).Summary()) },
	} {
		if err := call(); !errors.Is(err, ErrInvalidCatalog) {
			t.Fatalf("%s error = %v, want ErrInvalidCatalog", name, err)
		}
	}
}
