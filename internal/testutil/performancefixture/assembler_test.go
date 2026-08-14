package performancefixture

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type testContribution struct {
	descriptor Descriptor
	calls      *int
	receipt    Receipt
	err        error
}

func (c testContribution) Descriptor() Descriptor { return c.descriptor }

func (c testContribution) Apply(context.Context, *BuildState) (Receipt, error) {
	*c.calls++
	return c.receipt, c.err
}

type testValidator struct {
	result SemanticValidation
}

func (v testValidator) Validate(context.Context, *BuildState) (SemanticValidation, error) {
	return v.result, nil
}

func TestAssemblerRejectsInvalidClosureBeforeMutation(t *testing.T) {
	t.Parallel()
	cases := map[string][]Descriptor{
		"duplicate": {
			testDescriptor("one.v1", nil),
			testDescriptor("one.v1", nil),
		},
		"unknown": {
			testDescriptor("one.v1", []string{"missing.v1"}),
		},
		"cycle": {
			testDescriptor("one.v1", []string{"two.v1"}),
			testDescriptor("two.v1", []string{"one.v1"}),
		},
		"reordered": {
			testDescriptor("two.v1", []string{"one.v1"}),
			testDescriptor("one.v1", nil),
		},
	}
	for name, descriptors := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			_, err := NewAssembler(descriptors, map[string]int{"rows": 1}, testValidator{}, testContribution{calls: &calls})
			if err == nil {
				t.Fatal("expected invalid closure to fail")
			}
			if calls != 0 {
				t.Fatalf("invalid closure mutated state %d times", calls)
			}
		})
	}
}

func TestAssemblerRejectsReceiptAndSemanticMutation(t *testing.T) {
	t.Parallel()
	descriptor := testDescriptor("one.v1", nil)
	calls := 0
	contribution := testContribution{
		descriptor: descriptor,
		calls:      &calls,
		receipt: Receipt{
			ContributionID: descriptor.ContributionID,
			Version:        descriptor.Version,
			OwnerID:        descriptor.OwnerID,
			Counts:         map[string]int{"rows": 2},
		},
	}
	assembler, err := NewAssembler(
		[]Descriptor{descriptor},
		map[string]int{"rows": 1},
		testValidator{result: passingValidation(map[string]int{"rows": 1})},
		contribution,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = assembler.Assemble(context.Background(), testState())
	if err == nil {
		t.Fatal("expected receipt mutation to fail")
	}
	contribution.receipt.Counts["rows"] = 1
	assembler, err = NewAssembler(
		[]Descriptor{descriptor},
		map[string]int{"rows": 1},
		testValidator{result: passingValidation(map[string]int{"rows": 2})},
		contribution,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = assembler.Assemble(context.Background(), testState())
	if err == nil {
		t.Fatal("expected semantic mutation to fail")
	}
}

func TestAssemblerDigestIsDeterministicAndRedacted(t *testing.T) {
	t.Parallel()
	descriptor := testDescriptor("one.v1", nil)
	assemble := func(state *BuildState) Result {
		t.Helper()
		calls := 0
		assembler, err := NewAssembler(
			[]Descriptor{descriptor},
			map[string]int{"rows": 1},
			testValidator{result: passingValidation(map[string]int{"rows": 1})},
			testContribution{
				descriptor: descriptor,
				calls:      &calls,
				receipt: Receipt{
					ContributionID: descriptor.ContributionID,
					Version:        descriptor.Version,
					OwnerID:        descriptor.OwnerID,
					Counts:         map[string]int{"rows": 1},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		result, err := assembler.Assemble(context.Background(), state)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	firstState := testState()
	secondState := testState()
	secondState.RuntimeBundle.BackgroundAccounts[0].Email = "other@example.test"
	secondState.RuntimeBundle.BackgroundAccounts[0].Password = "different-password-material-1234"
	first := assemble(firstState)
	second := assemble(secondState)
	if first.SemanticValidationDigest != second.SemanticValidationDigest {
		t.Fatalf("runtime entropy changed semantic digest: %s != %s", first.SemanticValidationDigest, second.SemanticValidationDigest)
	}
	if !reflect.DeepEqual(first.Receipts, second.Receipts) {
		t.Fatal("runtime entropy changed safe receipts")
	}
	if err := ValidateReceiptRedaction(first, firstState); err != nil {
		t.Fatal(err)
	}
}

func TestAssemblerPropagatesContributionFailure(t *testing.T) {
	t.Parallel()
	descriptor := testDescriptor("one.v1", nil)
	calls := 0
	assembler, err := NewAssembler(
		[]Descriptor{descriptor},
		map[string]int{"rows": 1},
		testValidator{result: passingValidation(map[string]int{"rows": 1})},
		testContribution{descriptor: descriptor, calls: &calls, err: errors.New("injected mutation failure")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := assembler.Assemble(context.Background(), testState()); err == nil {
		t.Fatal("expected injected mutation failure")
	}
}

func testDescriptor(id string, dependencies []string) Descriptor {
	return Descriptor{
		ContributionID: id,
		Version:        id,
		OwnerID:        "module.test",
		Dependencies:   dependencies,
		ExpectedCounts: map[string]int{"rows": 1},
	}
}

func testState() *BuildState {
	return &BuildState{
		SnapshotKey: "a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6",
		Seed:        20260405,
		RuntimeBundle: RuntimeBundle{
			SchemaID:         RuntimeSchemaID,
			FixtureProfileID: LargeGridProfileID,
			SnapshotKey:      "a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6",
			BackgroundAccounts: []BackgroundAccount{{
				Email:    "secret@example.test",
				Password: "secret-password-material-1234",
			}},
		},
		BackgroundUserIDs: []string{"secret-user-id"},
		IncidentID:        "secret-incident-id",
	}
}

func passingValidation(counts map[string]int) SemanticValidation {
	return SemanticValidation{
		Counts:                   counts,
		RelationshipDistribution: true,
		DefaultView:              true,
		Authorization:            true,
		ProjectionReadiness:      true,
		NoActiveSessions:         true,
	}
}
