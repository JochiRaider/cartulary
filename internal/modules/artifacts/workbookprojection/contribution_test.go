package workbookprojection

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Artifact projection contribution unexpectedly constructed")
	}
}

func TestArtifactProjectionContractOwnsEightSemanticSurfaces(t *testing.T) {
	contribution, err := NewContribution(artifactSourceStub{})
	if err != nil {
		t.Fatalf("construct Artifact projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	if len(descriptors) != 1 {
		t.Fatalf("Artifact descriptors = %d, want 1", len(descriptors))
	}
	descriptor := descriptors[0]
	if descriptor.ProviderID != "artifact" ||
		descriptor.SourceOwnerModule != "artifacts" ||
		descriptor.ProjectionStorageOwnerModule != "projections" ||
		len(descriptor.ViewSchemaIDs) != 8 ||
		len(descriptor.FacadePackages) != 1 ||
		descriptor.FacadePackages[0] != "internal/modules/artifacts/workbookprojection" {
		t.Fatalf("unexpected Artifact descriptor: %#v", descriptor)
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 8 {
		t.Fatalf("Artifact semantic intents = %d, want 8", len(intents))
	}
	seen := make(map[string]struct{}, len(intents))
	for _, intent := range intents {
		if len(intent.FieldKeys) == 0 || intent.CanonicalSourceFilter == nil ||
			intent.CanonicalSourceFilter.Kind != "artifact_type" ||
			intent.CanonicalSourceFilter.Value == "" {
			t.Fatalf("incomplete Artifact semantic intent: %#v", intent)
		}
		if _, duplicate := seen[intent.ViewSchemaID]; duplicate {
			t.Fatalf("duplicate Artifact semantic intent %q", intent.ViewSchemaID)
		}
		seen[intent.ViewSchemaID] = struct{}{}
	}
}

func TestArtifactProjectionIntentRejectsUnknownViewSchema(t *testing.T) {
	t.Parallel()
	if _, err := surfaceIntent("cartulary.view.unknown.v1"); err == nil {
		t.Fatal("unknown Artifact view schema produced a semantic intent")
	}
}

func TestArtifactProjectionContributionDefensivelyCopiesFacts(t *testing.T) {
	contribution, err := NewContribution(artifactSourceStub{})
	if err != nil {
		t.Fatalf("construct Artifact projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	intents := contribution.ProjectionContribution().SurfaceIntents()
	intents[0].FieldKeys[0] = "mutated"
	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/artifacts/workbookprojection" {
		t.Fatalf("descriptor mutation escaped contribution: %q", got)
	}
	if got := again.SurfaceIntents()[0].FieldKeys[0]; got == "mutated" {
		t.Fatal("semantic intent mutation escaped contribution")
	}
	if contribution.Source() == nil {
		t.Fatal("runtime contribution has no typed source")
	}
}

func TestArtifactProjectionInputClosedVariantMatrix(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	title := "Projection title"
	body := "Projection body"
	day := "2026-08-27"
	actorID := uuid.New()
	envelope, err := NewProjectionEnvelope(
		uuid.New(), uuid.New(), 4, &title, &body, &now, now.Add(time.Minute), now, &actorID, &day, 2,
	)
	if err != nil {
		t.Fatalf("construct projection envelope: %v", err)
	}
	tests := []struct {
		name         string
		artifactType string
		variantType  reflect.Type
		construct    func() (ProjectionInput, error)
	}{
		{name: "note", artifactType: "note", variantType: reflect.TypeOf(NoteVariant{}), construct: func() (ProjectionInput, error) {
			return NewNoteProjectionInput(envelope, NoteVariant{})
		}},
		{name: "communication log", artifactType: "comm_log", variantType: reflect.TypeOf(CommunicationLogVariant{}), construct: func() (ProjectionInput, error) {
			return NewCommunicationLogProjectionInput(envelope, CommunicationLogVariant{
				CommID: "comm-1", CommType: "briefing", Audience: "responders",
				ChannelOrMeeting: "bridge", Summary: "Summary",
			})
		}},
		{name: "handoff", artifactType: "handoff", variantType: reflect.TypeOf(HandoffVariant{}), construct: func() (ProjectionInput, error) {
			return NewHandoffProjectionInput(envelope, HandoffVariant{
				HandoffID: "handoff-1", OutgoingOwnerUserID: actorID, IncomingOwnerUserID: uuid.New(),
				CurrentStateSummary: "Stable", AckState: "pending",
			})
		}},
		{name: "status review", artifactType: "status_review", variantType: reflect.TypeOf(StatusReviewVariant{}), construct: func() (ProjectionInput, error) {
			return NewStatusReviewProjectionInput(envelope, StatusReviewVariant{
				StatusReviewID: "review-1", ReviewOwnerUserID: actorID, CurrentStateSummary: "Recovering",
			})
		}},
		{name: "lesson", artifactType: "lesson", variantType: reflect.TypeOf(LessonVariant{}), construct: func() (ProjectionInput, error) {
			return NewLessonProjectionInput(envelope, LessonVariant{
				LessonID: "lesson-1", Summary: "Capture early", OwnerUserID: actorID, ClosureState: "open",
			})
		}},
		{name: "finding", artifactType: "finding", variantType: reflect.TypeOf(FindingVariant{}), construct: func() (ProjectionInput, error) {
			return NewFindingProjectionInput(envelope, FindingVariant{
				Statement: "Evidence confirms", Kind: "finding", State: "open", OwnerUserID: actorID,
				UpdatedAt: now, ConfidenceBand: "unset",
			})
		}},
		{name: "investigative query", artifactType: "investigative_query", variantType: reflect.TypeOf(InvestigativeQueryVariant{}), construct: func() (ProjectionInput, error) {
			return NewInvestigativeQueryProjectionInput(envelope, InvestigativeQueryVariant{
				QueryID: "query-1", Platform: "edr", Purpose: "scope", QueryText: "search",
				CreatedByUserID: actorID, CreatedAt: now, CreatedDay: day,
			})
		}},
		{name: "forensic keyword", artifactType: "forensic_keyword", variantType: reflect.TypeOf(ForensicKeywordVariant{}), construct: func() (ProjectionInput, error) {
			return NewForensicKeywordProjectionInput(envelope, ForensicKeywordVariant{
				KeywordID: "keyword-1", Pattern: "pattern", Reason: "coverage", MatchMode: "literal",
				CreatedAt: now, CreatedDay: day,
			})
		}},
	}
	inputs := make([]ProjectionInput, 0, len(tests))
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input, err := tc.construct()
			if err != nil {
				t.Fatalf("construct variant: %v", err)
			}
			if input.ArtifactType() != tc.artifactType || reflect.TypeOf(input.Variant()) != tc.variantType {
				t.Fatalf("typed input = %q/%T", input.ArtifactType(), input.Variant())
			}
			inputs = append(inputs, input)
		})
	}
	if len(inputs) != 8 {
		t.Fatalf("valid projection variants = %d, want 8", len(inputs))
	}

	title = "mutated"
	if got := *inputs[0].Envelope().Title(); got != "Projection title" {
		t.Fatalf("envelope mutation escaped: %q", got)
	}
	cursor := inputs[7].Envelope().RecordID()
	page, err := NewProjectionInputPage(inputs, &cursor)
	if err != nil {
		t.Fatalf("construct projection page: %v", err)
	}
	inputs[0] = ProjectionInput{}
	cursor = uuid.Nil
	if got := page.Inputs(); len(got) != 8 || got[0].ArtifactType() != "note" {
		t.Fatalf("page input copy = %#v", got)
	}
	if got := page.NextRecordID(); got == nil || *got == uuid.Nil {
		t.Fatalf("page cursor copy = %v", got)
	}

	if _, err := NewNoteProjectionInput(ProjectionEnvelope{}, NoteVariant{}); err == nil {
		t.Fatal("zero projection envelope was accepted")
	}
	if _, err := NewFindingProjectionInput(envelope, FindingVariant{}); err == nil {
		t.Fatal("incomplete finding variant was accepted")
	}
	if _, err := newProjectionInput(envelope, unknownProjectionVariant{}); err == nil {
		t.Fatal("unknown projection variant was accepted")
	}
	if _, err := NewProjectionInputPage([]ProjectionInput{{}}, nil); err == nil {
		t.Fatal("page accepted a zero projection input")
	}
}

type unknownProjectionVariant struct{}

func (unknownProjectionVariant) projectionVariant() {}

type artifactSourceStub struct{ SourceReader }
