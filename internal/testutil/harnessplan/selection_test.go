package harnessplan

import (
	"errors"
	"testing"
)

func TestCommandSelection(t *testing.T) {
	inputs, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	selection := func(command, suite string) Selection {
		t.Helper()
		entry, _ := inputs.TaskSurface.target(command)
		resolved, err := resolveInputs(entry, map[string]string{"SUITE": suite, makeInputSources: "SUITE=cli"})
		if err != nil {
			t.Fatal(err)
		}
		chosen, err := inputs.commandSelection(entry, resolved)
		if err != nil {
			t.Fatal(err)
		}
		return chosen
	}
	explicitOnly := func(selection Selection) int {
		count := 0
		for _, id := range selection.Rows {
			if inputs.row(id).AggregatePlacement == "explicit_only" {
				count++
			}
		}
		return count
	}

	t.Run("an explicit selection runs its explicit-only rows", func(t *testing.T) {
		if visual := selection("browser-e2e", "visual"); explicitOnly(visual) == 0 {
			t.Fatalf("the visual suite selects %d rows and none is explicit-only", len(visual.Rows))
		}
	})

	t.Run("any other selection drops explicit-only rows", func(t *testing.T) {
		if integration := selection("test-slice", "backend-integration"); len(integration.Rows) == 0 || explicitOnly(integration) != 0 {
			t.Fatalf("backend integration selects %d rows, %d explicit-only", len(integration.Rows), explicitOnly(integration))
		}
	})

	t.Run("an unknown explicit selection is rejected", func(t *testing.T) {
		model, err := Load(repositoryRoot(t))
		if err != nil {
			t.Fatal(err)
		}
		model.Owner.ExplicitSelections = append(model.Owner.ExplicitSelections, "harness.missing")
		if err := model.validate(); !errors.Is(err, ErrInput) {
			t.Fatalf("validate = %v, want ErrInput", err)
		}
	})
}
