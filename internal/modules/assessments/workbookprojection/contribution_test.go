package workbookprojection

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type assessmentContributionSource struct{ SourceReader }

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Assessment projection contribution unexpectedly constructed")
	}
}

func TestAssessmentProjectionContractOwnsTypedImmutableFacts(t *testing.T) {
	descriptor := Descriptor()
	if descriptor.ProviderID != "assessment" ||
		descriptor.SourceOwnerModule != "assessments" ||
		descriptor.ProjectionStorageOwnerModule != "projections" ||
		len(descriptor.FacadePackages) != 1 ||
		descriptor.FacadePackages[0] != "internal/modules/assessments/workbookprojection" {
		t.Fatalf("unexpected assessment descriptor: %#v", descriptor)
	}

	contribution, err := NewContribution(&assessmentContributionSource{})
	if err != nil {
		t.Fatalf("construct assessment projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	intents := contribution.ProjectionContribution().SurfaceIntents()
	intents[0].FieldKeys[0] = "mutated"
	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/assessments/workbookprojection" {
		t.Fatalf("descriptor mutation escaped contribution: %q", got)
	}
	if got := again.SurfaceIntents()[0].FieldKeys[0]; got != "assessment.subject_ref" {
		t.Fatalf("intent mutation escaped contribution: %q", got)
	}
}

func TestAssessmentProjectionMutationValidation(t *testing.T) {
	recordID := uuid.New()
	score := 85
	input := ProjectionInput{
		RecordID:            recordID,
		IncidentID:          uuid.New(),
		RowVersion:          1,
		SubjectRef:          uuid.New(),
		SubjectType:         "host",
		AssessmentState:     "confirmed",
		ConfidenceScore:     &score,
		ConfidenceBand:      "high",
		Rationale:           "typed source DTO",
		Assessor:            uuid.New(),
		AssessedAt:          time.Now().UTC(),
		SupportingLinkCount: 2,
	}
	if err := (ProjectionMutation{
		Kind:     ProjectionMutationUpsert,
		RecordID: recordID,
		Input:    input,
	}).Validate(); err != nil {
		t.Fatalf("validate canonical upsert: %v", err)
	}
	if err := (ProjectionMutation{
		Kind:     ProjectionMutationDelete,
		RecordID: recordID,
	}).Validate(); err != nil {
		t.Fatalf("validate canonical delete: %v", err)
	}
	input.ConfidenceBand = "medium"
	if err := (ProjectionMutation{
		Kind:     ProjectionMutationUpsert,
		RecordID: recordID,
		Input:    input,
	}).Validate(); err == nil {
		t.Fatal("non-canonical confidence band unexpectedly validated")
	}
}
