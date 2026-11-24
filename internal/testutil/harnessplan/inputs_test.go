package harnessplan

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// repositoryRoot walks up from the package directory to the module root.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

func TestPlannerInputs(t *testing.T) {
	root := repositoryRoot(t)
	inputs, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("loads the authored model", func(t *testing.T) {
		if len(inputs.Catalog.Owners) == 0 || len(inputs.Catalog.Rows) == 0 {
			t.Fatal("catalog is empty")
		}
		if len(inputs.Owner.PolicyUnits) == 0 || len(inputs.Owner.Operations) == 0 || len(inputs.Topology.ResourceProfiles) == 0 {
			t.Fatal("owner model or topology is empty")
		}
		runners := map[string]bool{}
		for _, row := range inputs.Catalog.Rows {
			runners[row.Runner] = true
		}
		for _, runner := range []string{RunnerGo, RunnerVitest, RunnerPlaywright, RunnerShell} {
			if !runners[runner] {
				t.Fatalf("no %s rows loaded", runner)
			}
		}
	})

	// Each case breaks one invariant in a copy of the real model.
	mutations := map[string]func(in *Inputs){
		"duplicate row": func(in *Inputs) { in.Catalog.Rows = append(in.Catalog.Rows, in.Catalog.Rows[0]) },
		"unknown resource profile": func(in *Inputs) {
			in.Catalog.Rows[0].ResourceProfileID = "missing"
		},
		"postgres isolation without postgres": func(in *Inputs) {
			for i, row := range in.Catalog.Rows {
				if row.PostgresIsolation == "none" {
					in.Catalog.Rows[i].PostgresIsolation = "exclusive"
					return
				}
			}
		},
		"selector for the wrong runner": func(in *Inputs) {
			for i, row := range in.Catalog.Rows {
				if row.Runner == RunnerGo {
					in.Catalog.Rows[i].Selector.File = "apps/web/src/x.test.ts"
					return
				}
			}
		},
		"policy needs an unknown operation": func(in *Inputs) {
			for id, unit := range in.Owner.PolicyUnits {
				unit.Needs = append(append([]string{}, unit.Needs...), "harness.missing")
				in.Owner.PolicyUnits[id] = unit
				return
			}
		},
		"aggregate root without a policy unit": func(in *Inputs) {
			in.Owner.AggregatePolicyRoots["check"] = append(append([]string{}, in.Owner.AggregatePolicyRoots["check"]...), "harness.missing")
		},
		"shell row naming an unknown operation": func(in *Inputs) {
			for i, row := range in.Catalog.Rows {
				if row.Runner == RunnerShell {
					in.Catalog.Rows[i].Selector.OperationID = "harness.missing"
					return
				}
			}
		},
		"missing capacity policy": func(in *Inputs) { delete(in.Registry.CapacityPolicies, "postgres_lanes") },
		"cacheable policy without a cache profile": func(in *Inputs) {
			unit := in.Owner.PolicyUnits["harness.projection_sql_check"]
			unit.CachePolicy = "content_addressed"
			in.Owner.PolicyUnits["harness.projection_sql_check"] = unit
		},
		"operation bound by two cache profiles": func(in *Inputs) {
			in.Cache.Profiles[1].OperationIDs = append(slices.Clone(in.Cache.Profiles[1].OperationIDs), in.Cache.Profiles[0].OperationIDs[0])
		},
		"populated fixture profile on a Go row": func(in *Inputs) {
			for i, row := range in.Catalog.Rows {
				if row.Runner == RunnerGo && row.EvidenceClass == "measurement" {
					populated := in.Catalog.Rows[slices.IndexFunc(in.Catalog.Rows, func(row Row) bool { return row.FixtureProfileID != "" })]
					in.Catalog.Rows[i].FixtureProfileID = populated.FixtureProfileID
					in.Catalog.Rows[i].VerificationIDs = populated.VerificationIDs
					return
				}
			}
		},
		"fixture profile without a bound verification": func(in *Inputs) {
			i := slices.IndexFunc(in.Catalog.Rows, func(row Row) bool { return row.FixtureProfileID != "" })
			in.Catalog.Rows[i].VerificationIDs = nil
		},
		"partition disagreeing with its row's fixture profile": func(in *Inputs) {
			i := slices.IndexFunc(in.Browser.Partitions, func(partition BrowserPartition) bool { return partition.FixtureProfileID != "" })
			in.Browser.Partitions[i].FixtureProfileID = ""
		},
		"active fixture profile without a builder": func(in *Inputs) {
			in.Fixtures.Builders = in.Fixtures.Builders[1:]
		},
		"fixture builder with an unknown resource profile": func(in *Inputs) {
			in.Fixtures.Builders = slices.Clone(in.Fixtures.Builders)
			in.Fixtures.Builders[0].ResourceProfileID = "missing"
		},
		"unit-aware profile without unit inputs": func(in *Inputs) {
			for i, profile := range in.Cache.Profiles {
				if profile.DependencyStrategy == "unit_aware" {
					in.Cache.Profiles[i].UnitInputs = nil
				}
			}
		},
	}
	for name, mutate := range mutations {
		t.Run("rejects "+name, func(t *testing.T) {
			model, err := Load(root)
			if err != nil {
				t.Fatal(err)
			}
			mutate(model)
			if err := model.validate(); !errors.Is(err, ErrInput) {
				t.Fatalf("validate = %v, want ErrInput", err)
			}
		})
	}
}
