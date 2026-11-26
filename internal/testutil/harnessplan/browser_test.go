package harnessplan

import (
	"slices"
	"strings"
	"testing"
)

func TestBrowserPacking(t *testing.T) {
	inputs, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	row := func(id string) *Row { return &inputs.Catalog.Rows[inputs.rowIndex[id]] }
	ids := func(bins []browserBin) [][]string {
		out := make([][]string, len(bins))
		for i, bin := range bins {
			for _, r := range bin.rows {
				out[i] = append(out[i], r.RowID)
			}
			slices.Sort(out[i])
		}
		return out
	}
	policy := inputs.Browser.InvocationPolicy
	lifecycle := inputs.Owner.LifecycleEstimatesMS

	t.Run("every partition packs each row once into cohorts that fit their envelope", func(t *testing.T) {
		for _, lanes := range []int{1, 4, 6} {
			c := &compiler{in: inputs, options: Options{CPUTokens: 19, GoEnvironment: testGoEnvironment, BrowserLanes: lanes}}
			for i := range inputs.Browser.Partitions {
				partition := &inputs.Browser.Partitions[i]
				rows := make([]*Row, len(partition.SelectedRowIDs))
				for j, id := range partition.SelectedRowIDs {
					rows[j] = row(id)
				}
				if len(rows) == 0 {
					continue
				}
				bins := c.packBrowser(partition, rows)
				reversed := slices.Clone(rows)
				slices.Reverse(reversed)
				if !slices.EqualFunc(ids(bins), ids(c.packBrowser(partition, reversed)), slices.Equal) {
					t.Fatalf("%s: packing depends on row order", partition.Name)
				}
				if len(bins) > len(rows) {
					t.Fatalf("%s: %d cohorts for %d rows", partition.Name, len(bins), len(rows))
				}
				overhead := policy.StartupMS + policy.CleanupMS
				envelope := c.timeout(partition.ResourceProfileID) - max(lifecycle["browser_allocation_acquire"], lifecycle["browser_reset"]) -
					lifecycle["browser_allocation_release"] - overhead
				seen := map[string]bool{}
				for _, bin := range bins {
					if len(bin.rows) == 0 {
						t.Fatalf("%s: empty cohort", partition.Name)
					}
					if len(bin.rows) > 1 && bin.work-overhead > envelope {
						t.Fatalf("%s: a %d-row cohort of %d ms exceeds its %d ms envelope", partition.Name, len(bin.rows), bin.work-overhead, envelope)
					}
					if len(bin.rows) > 1 && (partition.StateScope != "test_owned" || partition.FixtureProfileID != "") {
						t.Fatalf("%s: a row outside test-owned state shares a cohort", partition.Name)
					}
					for _, r := range bin.rows {
						if seen[r.RowID] {
							t.Fatalf("%s: row %s packed twice", partition.Name, r.RowID)
						}
						seen[r.RowID] = true
					}
				}
				if len(seen) != len(rows) {
					t.Fatalf("%s: %d of %d rows packed", partition.Name, len(seen), len(rows))
				}
			}
		}
	})

	t.Run("ordinary cohorts are unchained, claim one browser lane and hold their own stack", func(t *testing.T) {
		var ordinary []*Row
		for i := range inputs.Browser.Partitions {
			partition := &inputs.Browser.Partitions[i]
			if partition.ResourceProfileID == browserOrdinary && partition.FixtureProfileID == "" {
				for _, id := range partition.SelectedRowIDs {
					ordinary = append(ordinary, row(id))
				}
			}
		}
		for _, lanes := range []int{1, 4, 6} {
			c := &compiler{in: inputs, options: Options{CPUTokens: 19, GoEnvironment: testGoEnvironment, BrowserLanes: lanes},
				browser: map[string]*browserCohort{}, assigned: map[string]string{}, units: map[string]*Unit{}, targets: map[string][]string{},
				rows: map[string]*Row{}, shellRows: map[string][]*Row{}, selected: map[string]bool{}, planned: map[string]plannedUnit{}}
			if err := c.planBrowser(ordinary); err != nil {
				t.Fatal(err)
			}
			stacks := map[string]string{}
			for id, cohort := range c.browser {
				if _, err := c.browserUnit(id, nil); err != nil {
					t.Fatal(err)
				}
				unit := c.units[id]
				if unit.Claims[BrowserLaneResource] != 1 || len(unit.ExclusiveLocks) != 0 {
					t.Fatalf("%s: lane claim %d, exclusive locks %v", id, unit.Claims[BrowserLaneResource], unit.ExclusiveLocks)
				}
				for _, need := range unit.Needs {
					if strings.HasPrefix(need, "browser_group:") {
						t.Fatalf("%s waits for browser group %s", id, need)
					}
				}
				if other, ok := stacks[cohort.stack]; ok {
					t.Fatalf("%s and %s share stack %s", id, other, cohort.stack)
				}
				stacks[cohort.stack] = id
				for _, resident := range unit.Residents {
					if resident.Key != "object_store:shared" && resident.Scope != ScopeUnit {
						t.Fatalf("%s: resident %s has scope %s", id, resident.Key, resident.Scope)
					}
				}
			}
		}
	})
}
