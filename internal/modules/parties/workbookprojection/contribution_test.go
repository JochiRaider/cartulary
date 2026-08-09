package workbookprojection

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type partyProjectionSourceStub struct{}

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Party projection contribution unexpectedly constructed")
	}
}

func (*partyProjectionSourceStub) LoadProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionInput, bool, error) {
	return ProjectionInput{}, false, nil
}

func (*partyProjectionSourceStub) ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error) {
	return ProjectionInputPage{}, nil
}

func TestPartyProjectionContractOwnsSemanticSurface(t *testing.T) {
	descriptor := Descriptor()
	if descriptor.ProviderID != "party" ||
		descriptor.SourceOwnerModule != "parties" ||
		descriptor.ProjectionStorageOwnerModule != "projections" ||
		len(descriptor.ViewSchemaIDs) != 1 ||
		descriptor.ViewSchemaIDs[0] != partyViewSchemaID ||
		len(descriptor.FacadePackages) != 1 ||
		descriptor.FacadePackages[0] != "internal/modules/parties/workbookprojection" {
		t.Fatalf("unexpected Party descriptor: %#v", descriptor)
	}
	intent, err := SurfaceIntent()
	if err != nil {
		t.Fatalf("Party semantic intent: %v", err)
	}
	if intent.ViewSchemaID != partyViewSchemaID || len(intent.FieldKeys) == 0 {
		t.Fatalf("incomplete Party semantic intent: %#v", intent)
	}
}

func TestPartyProjectionIntentRejectsUnknownViewSchema(t *testing.T) {
	t.Parallel()
	if _, err := surfaceIntent("cartulary.view.unknown.v1"); err == nil {
		t.Fatal("unknown Party view schema produced a semantic intent")
	}
}

func TestPartyProjectionContributionDefensivelyCopiesFactsAndRetainsTypedSource(t *testing.T) {
	source := &partyProjectionSourceStub{}
	contribution, err := NewContribution(source)
	if err != nil {
		t.Fatalf("construct Party projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	intents := contribution.ProjectionContribution().SurfaceIntents()
	intents[0].FieldKeys[0] = "mutated"

	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/parties/workbookprojection" {
		t.Fatalf("descriptor mutation escaped contribution: %q", got)
	}
	if got := again.SurfaceIntents()[0].FieldKeys[0]; got == "mutated" {
		t.Fatal("semantic intent mutation escaped contribution")
	}
	if contribution.Source() != source {
		t.Fatal("typed source was not retained")
	}
}
