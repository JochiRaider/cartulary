package workbookprojection

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestTaskDecisionProjectionContractOwnsTwoSemanticSurfaces(t *testing.T) {
	descriptors := Descriptors()
	if len(descriptors) != 2 ||
		descriptors[0].ProviderID != "task_request" ||
		descriptors[1].ProviderID != "decision" {
		t.Fatalf("unexpected Tasks/Decisions descriptors: %#v", descriptors)
	}
	for _, descriptor := range descriptors {
		if descriptor.SourceOwnerModule != "tasksdecisions" ||
			descriptor.ProjectionStorageOwnerModule != "projections" ||
			len(descriptor.FacadePackages) != 1 ||
			descriptor.FacadePackages[0] != "internal/modules/tasksdecisions/workbookprojection" {
			t.Fatalf("unexpected Tasks/Decisions descriptor: %#v", descriptor)
		}
	}
	intents := SurfaceIntents()
	if len(intents) != 2 || len(intents[0].FieldKeys) == 0 || len(intents[1].FieldKeys) == 0 {
		t.Fatalf("incomplete Tasks/Decisions semantic intents: %#v", intents)
	}
}

func TestTaskDecisionProjectionContributionDefensivelyCopiesFactsAndRetainsTypedSources(t *testing.T) {
	contribution, err := NewRuntimeContribution(
		contractSource{},
		contractSource{},
	)
	if err != nil {
		t.Fatalf("construct Tasks/Decisions projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	semantic := contribution.ProjectionContribution().SurfaceIntents()
	semantic[0].FieldKeys[0] = "mutated"

	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/tasksdecisions/workbookprojection" {
		t.Fatalf("descriptor mutation escaped contribution: %q", got)
	}
	if got := again.SurfaceIntents()[0].FieldKeys[0]; got == "mutated" {
		t.Fatal("semantic intent mutation escaped contribution")
	}
	if contribution.TaskRequestSource() == nil || contribution.DecisionSource() == nil {
		t.Fatal("typed sources were not retained")
	}
}

type contractSource struct{}

func (contractSource) LoadTaskRequestProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (TaskRequestProjectionInput, bool, error) {
	return TaskRequestProjectionInput{}, false, nil
}
func (contractSource) ListTaskRequestProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (TaskRequestProjectionInputPage, error) {
	return TaskRequestProjectionInputPage{}, nil
}
func (contractSource) LoadDecisionProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (DecisionProjectionInput, bool, error) {
	return DecisionProjectionInput{}, false, nil
}
func (contractSource) ListDecisionProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (DecisionProjectionInputPage, error) {
	return DecisionProjectionInputPage{}, nil
}
