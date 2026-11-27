package harnessplan

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestOperationsWaitForTheirProducers(t *testing.T) {
	execution := func(kind string, prerequisites ...string) json.RawMessage {
		data, _ := json.Marshal(map[string]any{"type": kind, "prerequisites": prerequisites})
		return data
	}
	owner := Owner{Operations: map[string]Operation{
		"harness.frontend_install": {Execution: execution("artifact_binding", "$(FRONTEND_INSTALL_RECEIPT)")},
		"harness.bootstrap":        {Execution: execution("installation", "$(FRONTEND_INSTALL_RECEIPT)")},
		"harness.lint":             {Execution: execution("step_command", "$(NODE_BIN)", "$(FRONTEND_INSTALL_RECEIPT)")},
		"harness.wrapper":          {Execution: execution("step_command", "$(NODE_BIN)")},
	}}
	compile := func(units ...*Unit) error {
		c := &compiler{in: &Inputs{Owner: owner}, units: map[string]*Unit{}}
		for _, unit := range units {
			c.units[unit.ID] = unit
		}
		return c.waitForProducers()
	}
	producer := &Unit{ID: "operation:harness.frontend_install"}
	// A user waits for the producer directly or through another unit; an
	// installation and the producer itself make their own artifacts.
	if err := compile(producer, &Unit{ID: "operation:harness.lint", Needs: []string{"operation:harness.frontend_install"}}); err != nil {
		t.Fatal(err)
	}
	if err := compile(producer, &Unit{ID: "operation:harness.wrapper", Needs: []string{"operation:harness.frontend_install"}},
		&Unit{ID: "operation:harness.lint", Needs: []string{"operation:harness.wrapper"}}, &Unit{ID: "operation:harness.bootstrap"}); err != nil {
		t.Fatal(err)
	}
	err := compile(producer, &Unit{ID: "operation:harness.lint"})
	if !errors.Is(err, ErrInput) || !strings.Contains(err.Error(), "operation harness.lint uses $(FRONTEND_INSTALL_RECEIPT) but does not wait for its producer harness.frontend_install") {
		t.Fatalf("an operation that does not wait for its producer: %v", err)
	}
}
