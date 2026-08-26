package tasksdecisions

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
)

func TestReportingContributionCharacterization_Unit(t *testing.T) {
	if _, err := NewReportingContribution(nil); err == nil || err.Error() != "compose Tasks/Decisions reporting provider: projection reader is required" {
		t.Fatalf("nil Reporting reader error = %v", err)
	}

	taskID := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	decisionID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	reader := &reportingCharacterizationReader{
		tasks: []taskprojection.TaskDerivedFact{{
			RecordID: taskID,
			Value:    map[string]any{"task.status": "open"},
		}},
		decisions: []taskprojection.DecisionDerivedFact{{
			RecordID: decisionID,
			Value:    map[string]any{"decision.status": "approved"},
		}},
	}
	contribution, err := NewReportingContribution(reader)
	if err != nil {
		t.Fatalf("construct Reporting contribution: %v", err)
	}
	if got := contribution.ProviderKey(); got != "tasksdecisions" {
		t.Fatalf("Reporting provider key = %q", got)
	}
	supportRefs := map[string][]string{
		taskID.String():     {"evidence:task"},
		decisionID.String(): {"evidence:decision"},
	}
	output, err := contribution.CollectFactsTx(t.Context(), nil, uuid.New(), supportRefs)
	if err != nil {
		t.Fatalf("collect Reporting facts: %v", err)
	}
	if got, want := reader.calls, []string{"tasks", "decisions"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Reporting reader order: got %#v want %#v", got, want)
	}
	if output.ProviderKey != "tasksdecisions" || len(output.FieldFacts) != 2 {
		t.Fatalf("Reporting output = %#v", output)
	}
	if got, want := output.FieldFacts[0].Path, "/decisions/"+decisionID.String(); got != want {
		t.Fatalf("first Reporting fact path = %q, want %q", got, want)
	}
	if got, want := output.FieldFacts[1].Path, "/task_requests/"+taskID.String(); got != want {
		t.Fatalf("second Reporting fact path = %q, want %q", got, want)
	}
	if got, want := output.FieldFacts[0].SupportRefs, []string{"evidence:decision"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("decision support refs: got %#v want %#v", got, want)
	}
	if got, want := output.FieldFacts[1].SupportRefs, []string{"evidence:task"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("task support refs: got %#v want %#v", got, want)
	}
	supportRefs[decisionID.String()][0] = "mutated"
	if got := output.FieldFacts[0].SupportRefs[0]; got != "evidence:decision" {
		t.Fatalf("Reporting output retained caller-owned support refs: %q", got)
	}

	t.Run("task failure stops decision collection", func(t *testing.T) {
		sentinel := errors.New("task facts unavailable")
		reader := &reportingCharacterizationReader{taskErr: sentinel}
		contribution, err := NewReportingContribution(reader)
		if err != nil {
			t.Fatalf("construct task-failing contribution: %v", err)
		}
		if _, err := contribution.CollectFactsTx(t.Context(), nil, uuid.New(), nil); !errors.Is(err, sentinel) {
			t.Fatalf("task Reporting error = %v, want %v", err, sentinel)
		}
		if got, want := reader.calls, []string{"tasks"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("calls after task failure: got %#v want %#v", got, want)
		}
	})

	t.Run("decision failure follows successful task collection", func(t *testing.T) {
		sentinel := errors.New("decision facts unavailable")
		reader := &reportingCharacterizationReader{decisionErr: sentinel}
		contribution, err := NewReportingContribution(reader)
		if err != nil {
			t.Fatalf("construct decision-failing contribution: %v", err)
		}
		if _, err := contribution.CollectFactsTx(t.Context(), nil, uuid.New(), nil); !errors.Is(err, sentinel) {
			t.Fatalf("decision Reporting error = %v, want %v", err, sentinel)
		}
		if got, want := reader.calls, []string{"tasks", "decisions"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("calls after decision failure: got %#v want %#v", got, want)
		}
	})
}

type reportingCharacterizationReader struct {
	tasks       []taskprojection.TaskDerivedFact
	decisions   []taskprojection.DecisionDerivedFact
	taskErr     error
	decisionErr error
	calls       []string
}

func (reader *reportingCharacterizationReader) CollectTaskDerivedFactsTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
) ([]taskprojection.TaskDerivedFact, error) {
	reader.calls = append(reader.calls, "tasks")
	return reader.tasks, reader.taskErr
}

func (reader *reportingCharacterizationReader) CollectDecisionDerivedFactsTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
) ([]taskprojection.DecisionDerivedFact, error) {
	reader.calls = append(reader.calls, "decisions")
	return reader.decisions, reader.decisionErr
}

var _ taskprojection.ReportingReader = (*reportingCharacterizationReader)(nil)
